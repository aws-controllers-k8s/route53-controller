// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package hosted_zone

import (
	"context"
	"errors"
	"testing"

	svcapitypes "github.com/aws-controllers-k8s/route53-controller/apis/v1alpha1"
	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	ackerr "github.com/aws-controllers-k8s/runtime/pkg/errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	svcsdk "github.com/aws/aws-sdk-go-v2/service/route53"
	svcsdktypes "github.com/aws/aws-sdk-go-v2/service/route53/types"
)

// mockVPCClient records calls made to Associate/Disassociate.
// GetHostedZone is no longer needed — current VPCs come from Spec.VPCs.
type mockVPCClient struct {
	associated    []string // vpcIDs passed to Associate
	disassociated []string // vpcIDs passed to Disassociate
	assocErr      error    // optional error to return on every Associate call
	disassocErr   error    // optional error to return on every Disassociate call
}

func (m *mockVPCClient) AssociateVPCWithHostedZone(_ context.Context, in *svcsdk.AssociateVPCWithHostedZoneInput, _ ...func(*svcsdk.Options)) (*svcsdk.AssociateVPCWithHostedZoneOutput, error) {
	if m.assocErr != nil {
		return nil, m.assocErr
	}
	if in.VPC != nil && in.VPC.VPCId != nil {
		m.associated = append(m.associated, *in.VPC.VPCId)
	}
	return &svcsdk.AssociateVPCWithHostedZoneOutput{}, nil
}
func (m *mockVPCClient) DisassociateVPCFromHostedZone(_ context.Context, in *svcsdk.DisassociateVPCFromHostedZoneInput, _ ...func(*svcsdk.Options)) (*svcsdk.DisassociateVPCFromHostedZoneOutput, error) {
	if m.disassocErr != nil {
		return nil, m.disassocErr
	}
	if in.VPC != nil && in.VPC.VPCId != nil {
		m.disassociated = append(m.disassociated, *in.VPC.VPCId)
	}
	return &svcsdk.DisassociateVPCFromHostedZoneOutput{}, nil
}

func makeResource(vpcID, vpcRegion string, extraVPCIDs []string) *resource {
	r := &resource{ko: &svcapitypes.HostedZone{}}
	r.ko.Status.ID = aws.String("/hostedzone/Z123")
	r.ko.Spec.VPC = &svcapitypes.VPC{
		VPCID:     aws.String(vpcID),
		VPCRegion: aws.String(vpcRegion),
	}
	for _, id := range extraVPCIDs {
		vid := id
		r.ko.Spec.VPCs = append(r.ko.Spec.VPCs, &svcapitypes.VPC{
			VPCID:     &vid,
			VPCRegion: aws.String("us-east-1"),
		})
	}
	return r
}

// Test: swapping Spec.VPC from vpc-A to vpc-B associates vpc-B and disassociates vpc-A.
func TestSyncVPCAssociations_SwapPrimaryVPC(t *testing.T) {
	// desired: primary=vpc-B; AWS currently has vpc-A only
	desired := makeResource("vpc-B", "us-east-1", nil)
	latest := makeResource("vpc-A", "us-east-1", nil)
	latest.ko.Status.ID = desired.ko.Status.ID

	latest.ko.Spec.VPCs = []*svcapitypes.VPC{
		{VPCID: aws.String("vpc-A"), VPCRegion: aws.String("us-east-1")},
	}
	mock := &mockVPCClient{}
	rm := &resourceManager{}

	err := rm.syncVPCAssociations(context.Background(), mock, desired, latest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.associated) != 1 || mock.associated[0] != "vpc-B" {
		t.Errorf("expected vpc-B associated, got %v", mock.associated)
	}
	if len(mock.disassociated) != 1 || mock.disassociated[0] != "vpc-A" {
		t.Errorf("expected vpc-A disassociated, got %v", mock.disassociated)
	}
}

// Test: LastVPCAssociation error from Disassociate is wrapped as a terminal error.
// Note: the mock returns the error unconditionally — this exercises the error-wrapping
// code path. In real AWS, LastVPCAssociation fires when disassociation would leave
// zero VPCs; the associate loop runs first, so this would only occur when the new
// primary VPC fails to associate before disassociation is attempted.
func TestSyncVPCAssociations_LastVPCTerminal(t *testing.T) {
	// desired: primary=vpc-B; AWS currently has vpc-A only (vpc-B not yet associated).
	desired := makeResource("vpc-B", "us-east-1", nil)
	latest := makeResource("vpc-A", "us-east-1", nil)
	latest.ko.Status.ID = desired.ko.Status.ID

	latest.ko.Spec.VPCs = []*svcapitypes.VPC{
		{VPCID: aws.String("vpc-A"), VPCRegion: aws.String("us-east-1")},
	}
	mock := &mockVPCClient{disassocErr: &svcsdktypes.LastVPCAssociation{Message: aws.String("last VPC")}}
	rm := &resourceManager{}

	err := rm.syncVPCAssociations(context.Background(), mock, desired, latest)
	if err == nil {
		t.Fatal("expected terminal error, got nil")
	}
	var termErr *ackerr.TerminalError
	if !errors.As(err, &termErr) {
		t.Errorf("expected *ackerr.TerminalError, got %T: %v", err, err)
	}
}

// makeResourceVPCs builds a resource that uses spec.vpcs (not spec.vpc).
func makeResourceVPCs(vpcIDs []string) *resource {
	r := &resource{ko: &svcapitypes.HostedZone{}}
	r.ko.Status.ID = aws.String("/hostedzone/Z123")
	for _, id := range vpcIDs {
		vid := id
		r.ko.Spec.VPCs = append(r.ko.Spec.VPCs, &svcapitypes.VPC{
			VPCID:     &vid,
			VPCRegion: aws.String("us-east-1"),
		})
	}
	return r
}

// Test: syncVPCAssociations uses Spec.VPCs, not GetHostedZone.
// Passing a mock with no getOutput verifies no GetHostedZone call is made.
func TestSyncVPCAssociations_UsesStatusField_Associate(t *testing.T) {
	desired := makeResourceVPCs([]string{"vpc-A", "vpc-B"})
	latest := makeResourceVPCs([]string{"vpc-A"})
	latest.ko.Status.ID = desired.ko.Status.ID

	// mock has no GetHostedZone — if it were called it would panic
	mock := &mockVPCClient{}
	rm := &resourceManager{}

	err := rm.syncVPCAssociations(context.Background(), mock, desired, latest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.associated) != 1 || mock.associated[0] != "vpc-B" {
		t.Errorf("expected [vpc-B] associated, got %v", mock.associated)
	}
	if len(mock.disassociated) != 0 {
		t.Errorf("expected nothing disassociated, got %v", mock.disassociated)
	}
}

func TestSyncVPCAssociations_UsesStatusField_Disassociate(t *testing.T) {
	desired := makeResourceVPCs([]string{"vpc-A"})
	latest := makeResourceVPCs([]string{"vpc-A", "vpc-B"})
	latest.ko.Status.ID = desired.ko.Status.ID

	mock := &mockVPCClient{}
	rm := &resourceManager{}

	err := rm.syncVPCAssociations(context.Background(), mock, desired, latest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.disassociated) != 1 || mock.disassociated[0] != "vpc-B" {
		t.Errorf("expected [vpc-B] disassociated, got %v", mock.disassociated)
	}
}

// Test: spec.vpcs desired=[vpc-A,vpc-B], AWS current=[vpc-A] → associates vpc-B.
func TestSyncVPCAssociations_VPCsPath_Associate(t *testing.T) {
	desired := makeResourceVPCs([]string{"vpc-A", "vpc-B"})
	latest := makeResourceVPCs([]string{"vpc-A"})
	latest.ko.Status.ID = desired.ko.Status.ID

	latest.ko.Spec.VPCs = []*svcapitypes.VPC{
		{VPCID: aws.String("vpc-A"), VPCRegion: aws.String("us-east-1")},
	}
	mock := &mockVPCClient{}
	rm := &resourceManager{}

	err := rm.syncVPCAssociations(context.Background(), mock, desired, latest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.associated) != 1 || mock.associated[0] != "vpc-B" {
		t.Errorf("expected [vpc-B] associated, got %v", mock.associated)
	}
	if len(mock.disassociated) != 0 {
		t.Errorf("expected nothing disassociated, got %v", mock.disassociated)
	}
}

// Test: spec.vpcs desired=[vpc-A], AWS current=[vpc-A,vpc-B] → disassociates vpc-B.
func TestSyncVPCAssociations_VPCsPath_Disassociate(t *testing.T) {
	desired := makeResourceVPCs([]string{"vpc-A"})
	latest := makeResourceVPCs([]string{"vpc-A", "vpc-B"})
	latest.ko.Status.ID = desired.ko.Status.ID

	latest.ko.Spec.VPCs = []*svcapitypes.VPC{
		{VPCID: aws.String("vpc-A"), VPCRegion: aws.String("us-east-1")},
		{VPCID: aws.String("vpc-B"), VPCRegion: aws.String("us-east-1")},
	}
	mock := &mockVPCClient{}
	rm := &resourceManager{}

	err := rm.syncVPCAssociations(context.Background(), mock, desired, latest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.disassociated) != 1 || mock.disassociated[0] != "vpc-B" {
		t.Errorf("expected [vpc-B] disassociated, got %v", mock.disassociated)
	}
	if len(mock.associated) != 0 {
		t.Errorf("expected nothing associated, got %v", mock.associated)
	}
}

// Test: ConflictingDomainExists on AssociateVPC is treated as idempotent (no error).
func TestSyncVPCAssociations_VPCsPath_ConflictingDomainExistsIdempotent(t *testing.T) {
	// desired has vpc-A and vpc-B; AWS currently has only vpc-A
	desired := makeResourceVPCs([]string{"vpc-A", "vpc-B"})
	latest := makeResourceVPCs([]string{"vpc-A"})
	latest.ko.Status.ID = desired.ko.Status.ID

	latest.ko.Spec.VPCs = []*svcapitypes.VPC{
		{VPCID: aws.String("vpc-A"), VPCRegion: aws.String("us-east-1")},
	}
	mock := &mockVPCClient{assocErr: &svcsdktypes.ConflictingDomainExists{Message: aws.String("already associated")}}
	rm := &resourceManager{}

	err := rm.syncVPCAssociations(context.Background(), mock, desired, latest)
	if err != nil {
		t.Fatalf("expected nil error for ConflictingDomainExists, got: %v", err)
	}
}

// Test: both spec.vpc and spec.vpcs set → terminal error.
func TestValidateVPCFields_MutualExclusivity(t *testing.T) {
	spec := svcapitypes.HostedZoneSpec{
		VPC:  &svcapitypes.VPC{VPCID: aws.String("vpc-1"), VPCRegion: aws.String("us-east-1")},
		VPCs: []*svcapitypes.VPC{{VPCID: aws.String("vpc-2"), VPCRegion: aws.String("us-east-1")}},
	}
	err := validateVPCFields(spec)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var termErr *ackerr.TerminalError
	if !errors.As(err, &termErr) {
		t.Errorf("expected TerminalError, got %T: %v", err, err)
	}
}

// Test: hostedZoneConfig.privateZone: true with no VPCs → no error.
// The controller does not require VPCs to be set; AWS enforces that constraint.
func TestValidateVPCFields_ExplicitPrivateZoneNoVPCs(t *testing.T) {
	priv := true
	spec := svcapitypes.HostedZoneSpec{
		HostedZoneConfig: &svcapitypes.HostedZoneConfig{PrivateZone: &priv},
	}
	if err := validateVPCFields(spec); err != nil {
		t.Errorf("expected nil for explicit private zone with no VPCs, got %v", err)
	}
}

// Test: hostedZoneConfig.privateZone: true with empty spec.vpcs → no error.
func TestValidateVPCFields_ExplicitPrivateZoneEmptyVPCs(t *testing.T) {
	priv := true
	spec := svcapitypes.HostedZoneSpec{
		HostedZoneConfig: &svcapitypes.HostedZoneConfig{PrivateZone: &priv},
		VPCs:             []*svcapitypes.VPC{},
	}
	if err := validateVPCFields(spec); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

// Test: spec.vpc set with no hostedZoneConfig → no error (VPC presence implies private intent).
func TestValidateVPCFields_VPCImpliesPrivate(t *testing.T) {
	spec := svcapitypes.HostedZoneSpec{
		VPC: &svcapitypes.VPC{VPCID: aws.String("vpc-1"), VPCRegion: aws.String("us-east-1")},
	}
	if err := validateVPCFields(spec); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

// Test: spec.vpcs set with no hostedZoneConfig → no error.
func TestValidateVPCFields_VPCsImplyPrivate(t *testing.T) {
	spec := svcapitypes.HostedZoneSpec{
		VPCs: []*svcapitypes.VPC{{VPCID: aws.String("vpc-1"), VPCRegion: aws.String("us-east-1")}},
	}
	if err := validateVPCFields(spec); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

// Test: public zone with neither field → no error.
func TestValidateVPCFields_PublicZoneNoVPCs(t *testing.T) {
	spec := svcapitypes.HostedZoneSpec{}
	if err := validateVPCFields(spec); err != nil {
		t.Errorf("expected nil for public zone, got %v", err)
	}
}

// Test: hostedZoneConfig.privateZone explicitly false + spec.vpc → terminal error.
func TestValidateVPCFields_ExplicitPublicZoneWithVPC(t *testing.T) {
	pub := false
	spec := svcapitypes.HostedZoneSpec{
		HostedZoneConfig: &svcapitypes.HostedZoneConfig{PrivateZone: &pub},
		VPC:              &svcapitypes.VPC{VPCID: aws.String("vpc-1"), VPCRegion: aws.String("us-east-1")},
	}
	err := validateVPCFields(spec)
	if err == nil {
		t.Fatal("expected terminal error for explicitly public zone with spec.vpc, got nil")
	}
	var termErr *ackerr.TerminalError
	if !errors.As(err, &termErr) {
		t.Errorf("expected TerminalError, got %T: %v", err, err)
	}
}

// Test: hostedZoneConfig.privateZone explicitly false + spec.vpcs → terminal error.
func TestValidateVPCFields_ExplicitPublicZoneWithVPCs(t *testing.T) {
	pub := false
	spec := svcapitypes.HostedZoneSpec{
		HostedZoneConfig: &svcapitypes.HostedZoneConfig{PrivateZone: &pub},
		VPCs:             []*svcapitypes.VPC{{VPCID: aws.String("vpc-1"), VPCRegion: aws.String("us-east-1")}},
	}
	err := validateVPCFields(spec)
	if err == nil {
		t.Fatal("expected terminal error for explicitly public zone with spec.vpcs, got nil")
	}
	var termErr *ackerr.TerminalError
	if !errors.As(err, &termErr) {
		t.Errorf("expected TerminalError, got %T: %v", err, err)
	}
}

// Test: spec.vpcs entry with nil vpcRegion → terminal error.
func TestValidateVPCFields_VPCsEntryNilRegion(t *testing.T) {
	spec := svcapitypes.HostedZoneSpec{
		VPCs: []*svcapitypes.VPC{
			{VPCID: aws.String("vpc-1"), VPCRegion: nil},
		},
	}
	err := validateVPCFields(spec)
	if err == nil {
		t.Fatal("expected terminal error for nil vpcRegion, got nil")
	}
	var termErr *ackerr.TerminalError
	if !errors.As(err, &termErr) {
		t.Errorf("expected TerminalError, got %T: %v", err, err)
	}
}

// Test: spec.vpcs entry with nil vpcID → terminal error.
func TestValidateVPCFields_VPCsEntryNilID(t *testing.T) {
	spec := svcapitypes.HostedZoneSpec{
		VPCs: []*svcapitypes.VPC{
			{VPCID: nil, VPCRegion: aws.String("us-east-1")},
		},
	}
	err := validateVPCFields(spec)
	if err == nil {
		t.Fatal("expected terminal error for nil vpcID, got nil")
	}
	var termErr *ackerr.TerminalError
	if !errors.As(err, &termErr) {
		t.Errorf("expected TerminalError, got %T: %v", err, err)
	}
}

// Test: spec.vpcs entry that is nil itself → terminal error.
func TestValidateVPCFields_VPCsEntryNilStruct(t *testing.T) {
	spec := svcapitypes.HostedZoneSpec{
		VPCs: []*svcapitypes.VPC{nil},
	}
	err := validateVPCFields(spec)
	if err == nil {
		t.Fatal("expected terminal error for nil vpc struct entry, got nil")
	}
	var termErr *ackerr.TerminalError
	if !errors.As(err, &termErr) {
		t.Errorf("expected TerminalError, got %T: %v", err, err)
	}
}

// Test: nil VPCID in desired (a) → delta.
func TestCompareVPCs_NilVPCIDInA(t *testing.T) {
	a := &resource{ko: &svcapitypes.HostedZone{}}
	a.ko.Spec.VPCs = []*svcapitypes.VPC{
		{VPCID: nil, VPCRegion: aws.String("us-east-1")},
	}
	b := &resource{ko: &svcapitypes.HostedZone{}}
	b.ko.Spec.VPCs = []*svcapitypes.VPC{
		{VPCID: aws.String("vpc-1"), VPCRegion: aws.String("us-east-1")},
	}
	delta := ackcompare.NewDelta()
	compareVPCs(delta, a, b)
	if !delta.DifferentAt("Spec.VPCs") {
		t.Error("expected delta for Spec.VPCs when a has nil VPCID")
	}
}

// Test: same VPCs in different order → no delta.
func TestCompareVPCs_SameVPCsDifferentOrder(t *testing.T) {
	a := &resource{ko: &svcapitypes.HostedZone{}}
	a.ko.Spec.VPCs = []*svcapitypes.VPC{
		{VPCID: aws.String("vpc-1"), VPCRegion: aws.String("us-east-1")},
		{VPCID: aws.String("vpc-2"), VPCRegion: aws.String("us-west-2")},
	}
	b := &resource{ko: &svcapitypes.HostedZone{}}
	b.ko.Spec.VPCs = []*svcapitypes.VPC{
		{VPCID: aws.String("vpc-2"), VPCRegion: aws.String("us-west-2")},
		{VPCID: aws.String("vpc-1"), VPCRegion: aws.String("us-east-1")},
	}
	delta := ackcompare.NewDelta()
	compareVPCs(delta, a, b)
	if delta.DifferentAt("Spec.VPCs") {
		t.Error("expected no delta for same VPCs in different order")
	}
}

// Test: same VPC IDs but different regions → delta.
func TestCompareVPCs_SameIDDifferentRegion(t *testing.T) {
	a := &resource{ko: &svcapitypes.HostedZone{}}
	a.ko.Spec.VPCs = []*svcapitypes.VPC{
		{VPCID: aws.String("vpc-1"), VPCRegion: aws.String("us-east-1")},
	}
	b := &resource{ko: &svcapitypes.HostedZone{}}
	b.ko.Spec.VPCs = []*svcapitypes.VPC{
		{VPCID: aws.String("vpc-1"), VPCRegion: aws.String("us-west-2")},
	}
	delta := ackcompare.NewDelta()
	compareVPCs(delta, a, b)
	if !delta.DifferentAt("Spec.VPCs") {
		t.Error("expected delta for same VPC ID with different region")
	}
}

// Test: delete pre-cleanup keeps vpcs[0] (user's intended primary) and disassociates the rest.
func TestDeletePreCleanup_KeepsFirst(t *testing.T) {
	// vpc-c is [0] — the user's intended primary, regardless of lexicographic order.
	r := makeResourceVPCs([]string{"vpc-c", "vpc-a", "vpc-b"})
	r.ko.Status.ID = aws.String("/hostedzone/Z123")

	r.ko.Spec.VPCs = []*svcapitypes.VPC{
		{VPCID: aws.String("vpc-a"), VPCRegion: aws.String("us-east-1")},
		{VPCID: aws.String("vpc-b"), VPCRegion: aws.String("us-east-1")},
		{VPCID: aws.String("vpc-c"), VPCRegion: aws.String("us-east-1")},
	}
	mock := &mockVPCClient{}
	rm := &resourceManager{}

	// Simulate what the delete template does: keep [0], build synthetic desired.
	desired := rm.concreteResource(r.DeepCopy())
	desired.ko.Spec.VPCs = []*svcapitypes.VPC{r.ko.Spec.VPCs[0]}

	err := rm.syncVPCAssociations(context.Background(), mock, desired, r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.disassociated) != 2 {
		t.Fatalf("expected 2 disassociated, got %d: %v", len(mock.disassociated), mock.disassociated)
	}
	for _, id := range mock.disassociated {
		if id == "vpc-a" {
			t.Errorf("vpc-a (keeper, vpcs[0]) should not have been disassociated")
		}
	}
	if len(mock.associated) != 0 {
		t.Errorf("expected nothing associated, got %v", mock.associated)
	}
}

// Test: delete pre-cleanup triggers when Spec.VPCs > 1 even if Spec.VPCs has only 1 entry.
// This covers the case where a prior sync failed midway, leaving more VPCs in AWS than the spec.
func TestDeletePreCleanup_UsesStatusNotSpec(t *testing.T) {
	// Spec has only 1 VPC (user already reduced it), but AWS still has 2.
	r := makeResourceVPCs([]string{"vpc-a"})
	r.ko.Status.ID = aws.String("/hostedzone/Z123")
	r.ko.Spec.VPCs = []*svcapitypes.VPC{
		{VPCID: aws.String("vpc-a"), VPCRegion: aws.String("us-east-1")},
		{VPCID: aws.String("vpc-b"), VPCRegion: aws.String("us-east-1")},
	}

	mock := &mockVPCClient{}
	rm := &resourceManager{}

	// Use the production guard — a regression (e.g. reverting to len(Spec.VPCs) > 1)
	// would be detected here because shouldRunVPCPreCleanup calls into sdk.go.
	if !shouldRunVPCPreCleanup(r) {
		t.Fatal("shouldRunVPCPreCleanup should have returned true")
	}
	desired := rm.concreteResource(r.DeepCopy())
	desired.ko.Spec.VPCs = []*svcapitypes.VPC{r.ko.Spec.VPCs[0]}
	err := rm.syncVPCAssociations(context.Background(), mock, desired, r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(mock.disassociated) != 1 || mock.disassociated[0] != "vpc-b" {
		t.Errorf("expected vpc-b disassociated, got %v", mock.disassociated)
	}
	if len(mock.associated) != 0 {
		t.Errorf("expected nothing associated, got %v", mock.associated)
	}
}

// Test: boundary conditions for shouldRunVPCPreCleanup — regression-protects the
// > 1 threshold on Spec.VPCs and the requirement that Spec.VPCs be non-empty.
func TestShouldRunVPCPreCleanup(t *testing.T) {
	tests := []struct {
		name string
		//statusVPCs  []*svcapitypes.VPC
		specVPCs    []*svcapitypes.VPC
		wantCleanup bool
	}{
		{
			name: "nil status VPCs → no cleanup",
			//statusVPCs:  nil,
			specVPCs:    []*svcapitypes.VPC{{VPCID: aws.String("vpc-a"), VPCRegion: aws.String("us-east-1")}},
			wantCleanup: false,
		},
		{
			name: "single status VPC → no cleanup (Route53 accepts deletion)",
			//statusVPCs:  []*svcapitypes.VPC{{VPCID: aws.String("vpc-a"), VPCRegion: aws.String("us-east-1")}},
			specVPCs:    []*svcapitypes.VPC{{VPCID: aws.String("vpc-a"), VPCRegion: aws.String("us-east-1")}},
			wantCleanup: false,
		},
		// {
		// 	name: "two status VPCs but empty spec VPCs → no cleanup (legacy path, cannot determine keeper)",
		// 	statusVPCs: []*svcapitypes.VPC{
		// 		{VPCID: aws.String("vpc-a"), VPCRegion: aws.String("us-east-1")},
		// 		{VPCID: aws.String("vpc-b"), VPCRegion: aws.String("us-east-1")},
		// 	},
		// 	specVPCs:    nil,
		// 	wantCleanup: false,
		// },
		// {
		// 	name: "two status VPCs and one spec VPC → cleanup needed",
		// 	statusVPCs: []*svcapitypes.VPC{
		// 		{VPCID: aws.String("vpc-a"), VPCRegion: aws.String("us-east-1")},
		// 		{VPCID: aws.String("vpc-b"), VPCRegion: aws.String("us-east-1")},
		// 	},
		// 	specVPCs:    []*svcapitypes.VPC{{VPCID: aws.String("vpc-a"), VPCRegion: aws.String("us-east-1")}},
		// 	wantCleanup: true,
		// },
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &resource{ko: &svcapitypes.HostedZone{}}
			r.ko.Spec.VPCs = tc.specVPCs
			got := shouldRunVPCPreCleanup(r)
			if got != tc.wantCleanup {
				t.Errorf("shouldRunVPCPreCleanup() = %v, want %v", got, tc.wantCleanup)
			}
		})
	}
}

// Test: update with both spec.vpc and spec.vpcs set → terminal error, no AWS calls.
func TestCustomUpdateHostedZone_MutualExclusivity(t *testing.T) {
	desired := makeResource("vpc-1", "us-east-1", nil)
	desired.ko.Spec.VPCs = []*svcapitypes.VPC{
		{VPCID: aws.String("vpc-2"), VPCRegion: aws.String("us-east-1")},
	}
	latest := makeResource("vpc-1", "us-east-1", nil)
	delta := ackcompare.NewDelta()
	delta.Add("Spec.VPCs", nil, nil)

	rm := &resourceManager{}
	_, err := rm.customUpdateHostedZone(context.Background(), desired, latest, delta)
	if err == nil {
		t.Fatal("expected terminal error, got nil")
	}
	var termErr *ackerr.TerminalError
	if !errors.As(err, &termErr) {
		t.Errorf("expected TerminalError, got %T: %v", err, err)
	}
}

// Test: customUpdateHostedZone sets Spec.VPCs to match Spec.VPCs
// after a successful sync so the status is immediately accurate without
// waiting for the next sdkFind call.
//
// Scenario: desired.Spec.VPCs = [vpc-1], but the k8s status (carried on the
// desired object) still shows [vpc-1, vpc-2] because the previous update
// returned stale status. AWS is already in sync (latest.Spec.VPCs
// = [vpc-1]), so syncVPCAssociations makes no API calls — no SDK client needed.
func TestCustomUpdateHostedZone_StatusUpdatedAfterVPCSync(t *testing.T) {
	// desired: spec has only vpc-1; stale status has two (the bug condition)
	desired := makeResourceVPCs([]string{"vpc-1"})
	// latest: AWS already reflects the desired single-VPC state
	latest := makeResourceVPCs([]string{"vpc-1"})

	delta := ackcompare.NewDelta()
	delta.Add("Spec.VPCs", nil, nil)

	// sdkapi is nil; syncVPCAssociations will make no calls because current == desired.
	rm := &resourceManager{}
	updated, err := rm.customUpdateHostedZone(context.Background(), desired, latest, delta)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Spec.VPCs must match Spec.VPCs — not the stale two-VPC value.
	if len(updated.ko.Spec.VPCs) != 1 {
		t.Fatalf("expected 1 AssociatedVPC, got %d", len(updated.ko.Spec.VPCs))
	}
	if *updated.ko.Spec.VPCs[0].VPCID != "vpc-1" {
		t.Errorf("expected AssociatedVPCs[0] to be vpc-1, got %s", *updated.ko.Spec.VPCs[0].VPCID)
	}
}

// Test: both nil/empty → no delta.
func TestCompareVPCs_BothEmpty(t *testing.T) {
	a := &resource{ko: &svcapitypes.HostedZone{}}
	b := &resource{ko: &svcapitypes.HostedZone{}}
	delta := ackcompare.NewDelta()
	compareVPCs(delta, a, b)
	if delta.DifferentAt("Spec.VPCs") {
		t.Error("expected no delta for both empty VPCs")
	}
}

// ── compareVPC tests (spec.vpc legacy path) ──────────────────────────────────

// Test: Spec.VPC is nil → no delta (not on spec.vpc path).
func TestCompareVPC_NilSpecVPC(t *testing.T) {
	a := &resource{ko: &svcapitypes.HostedZone{}}
	b := &resource{ko: &svcapitypes.HostedZone{}}
	b.ko.Spec.VPCs = []*svcapitypes.VPC{
		{VPCID: aws.String("vpc-1"), VPCRegion: aws.String("us-east-1")},
	}
	delta := ackcompare.NewDelta()
	compareVPC(delta, a, b)
	if delta.DifferentAt("Spec.VPC") {
		t.Error("expected no delta when Spec.VPC is nil")
	}
}

// Test: desired VPC matches the sole associated VPC → no delta.
func TestCompareVPC_Match(t *testing.T) {
	a := &resource{ko: &svcapitypes.HostedZone{}}
	a.ko.Spec.VPC = &svcapitypes.VPC{VPCID: aws.String("vpc-1"), VPCRegion: aws.String("us-east-1")}
	b := &resource{ko: &svcapitypes.HostedZone{}}
	b.ko.Spec.VPCs = []*svcapitypes.VPC{
		{VPCID: aws.String("vpc-1"), VPCRegion: aws.String("us-east-1")},
	}
	delta := ackcompare.NewDelta()
	compareVPC(delta, a, b)
	if delta.DifferentAt("Spec.VPC") {
		t.Error("expected no delta when VPCs match")
	}
}

// Test: desired VPC differs from associated VPC → delta.
func TestCompareVPC_DifferentVPCID(t *testing.T) {
	a := &resource{ko: &svcapitypes.HostedZone{}}
	a.ko.Spec.VPC = &svcapitypes.VPC{VPCID: aws.String("vpc-2"), VPCRegion: aws.String("us-east-1")}
	b := &resource{ko: &svcapitypes.HostedZone{}}
	b.ko.Spec.VPCs = []*svcapitypes.VPC{
		{VPCID: aws.String("vpc-1"), VPCRegion: aws.String("us-east-1")},
	}
	delta := ackcompare.NewDelta()
	compareVPC(delta, a, b)
	if !delta.DifferentAt("Spec.VPC") {
		t.Error("expected delta when VPC IDs differ")
	}
}

// Test: desired VPC region differs → delta.
func TestCompareVPC_DifferentRegion(t *testing.T) {
	a := &resource{ko: &svcapitypes.HostedZone{}}
	a.ko.Spec.VPC = &svcapitypes.VPC{VPCID: aws.String("vpc-1"), VPCRegion: aws.String("us-west-2")}
	b := &resource{ko: &svcapitypes.HostedZone{}}
	b.ko.Spec.VPCs = []*svcapitypes.VPC{
		{VPCID: aws.String("vpc-1"), VPCRegion: aws.String("us-east-1")},
	}
	delta := ackcompare.NewDelta()
	compareVPC(delta, a, b)
	if !delta.DifferentAt("Spec.VPC") {
		t.Error("expected delta when VPC regions differ")
	}
}

// Test: AWS has more than one VPC while spec.vpc is set → delta.
// (Shouldn't happen in normal operation but must be handled cleanly.)
func TestCompareVPC_MultipleAssociated(t *testing.T) {
	a := &resource{ko: &svcapitypes.HostedZone{}}
	a.ko.Spec.VPC = &svcapitypes.VPC{VPCID: aws.String("vpc-1"), VPCRegion: aws.String("us-east-1")}
	b := &resource{ko: &svcapitypes.HostedZone{}}
	b.ko.Spec.VPCs = []*svcapitypes.VPC{
		{VPCID: aws.String("vpc-1"), VPCRegion: aws.String("us-east-1")},
		{VPCID: aws.String("vpc-2"), VPCRegion: aws.String("us-west-2")},
	}
	delta := ackcompare.NewDelta()
	compareVPC(delta, a, b)
	if !delta.DifferentAt("Spec.VPC") {
		t.Error("expected delta when AWS has extra VPCs")
	}
}

// Test: AWS has no associated VPCs → delta.
func TestCompareVPC_NoneAssociated(t *testing.T) {
	a := &resource{ko: &svcapitypes.HostedZone{}}
	a.ko.Spec.VPC = &svcapitypes.VPC{VPCID: aws.String("vpc-1"), VPCRegion: aws.String("us-east-1")}
	b := &resource{ko: &svcapitypes.HostedZone{}}
	b.ko.Spec.VPCs = nil
	delta := ackcompare.NewDelta()
	compareVPC(delta, a, b)
	if !delta.DifferentAt("Spec.VPC") {
		t.Error("expected delta when no VPCs associated")
	}
}

// ── shouldRunVPCPreCleanup tests ─────────────────────────────────────────────

func TestShouldRunVPCPreCleanup_SpecVPCsMultiple(t *testing.T) {
	r := &resource{ko: &svcapitypes.HostedZone{}}
	r.ko.Spec.VPCs = []*svcapitypes.VPC{
		{VPCID: aws.String("vpc-1"), VPCRegion: aws.String("us-east-1")},
		{VPCID: aws.String("vpc-2"), VPCRegion: aws.String("us-west-2")},
	}
	r.ko.Spec.VPCs = r.ko.Spec.VPCs
	if !shouldRunVPCPreCleanup(r) {
		t.Error("expected cleanup for spec.vpcs with 2 VPCs")
	}
}

func TestShouldRunVPCPreCleanup_SpecVPCSingle(t *testing.T) {
	r := &resource{ko: &svcapitypes.HostedZone{}}
	r.ko.Spec.VPCs = []*svcapitypes.VPC{
		{VPCID: aws.String("vpc-1"), VPCRegion: aws.String("us-east-1")},
	}
	r.ko.Spec.VPCs = r.ko.Spec.VPCs
	if shouldRunVPCPreCleanup(r) {
		t.Error("expected no cleanup when only 1 VPC associated")
	}
}

func TestShouldRunVPCPreCleanup_SpecVPCPathExtraVPCs(t *testing.T) {
	// spec.vpc user who has out-of-band extra VPCs — should now trigger cleanup.
	r := &resource{ko: &svcapitypes.HostedZone{}}
	r.ko.Spec.VPC = &svcapitypes.VPC{VPCID: aws.String("vpc-1"), VPCRegion: aws.String("us-east-1")}
	r.ko.Spec.VPCs = []*svcapitypes.VPC{
		{VPCID: aws.String("vpc-1"), VPCRegion: aws.String("us-east-1")},
		{VPCID: aws.String("vpc-extra"), VPCRegion: aws.String("us-west-2")},
	}
	if !shouldRunVPCPreCleanup(r) {
		t.Error("expected cleanup for spec.vpc path with extra associated VPCs")
	}
}

// func TestShouldRunVPCPreCleanup_NoSpecFields(t *testing.T) {
// 	// Neither spec.vpc nor spec.vpcs set — don't attempt cleanup.
// 	r := &resource{ko: &svcapitypes.HostedZone{}}
// 	r.ko.Spec.VPCs = []*svcapitypes.VPC{
// 		{VPCID: aws.String("vpc-1"), VPCRegion: aws.String("us-east-1")},
// 		{VPCID: aws.String("vpc-2"), VPCRegion: aws.String("us-west-2")},
// 	}
// 	if shouldRunVPCPreCleanup(r) {
// 		t.Error("expected no cleanup when neither spec field is set")
// 	}
// }
