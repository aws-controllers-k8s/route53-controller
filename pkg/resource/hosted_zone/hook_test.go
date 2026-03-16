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

// Test: populateAdditionalVPCs excludes the primary VPC (Spec.VPC) from AdditionalVPCs.
func TestPopulateAdditionalVPCs_ExcludesPrimaryVPC(t *testing.T) {
	ko := &svcapitypes.HostedZone{}
	ko.Spec.VPC = &svcapitypes.VPC{VPCID: aws.String("vpc-primary"), VPCRegion: aws.String("us-east-1")}

	awsVPCs := []svcsdktypes.VPC{
		{VPCId: aws.String("vpc-primary"), VPCRegion: svcsdktypes.VPCRegionUsEast1},
		{VPCId: aws.String("vpc-additional"), VPCRegion: svcsdktypes.VPCRegionUsEast1},
	}

	populateAdditionalVPCs(ko, awsVPCs)

	if len(ko.Spec.AdditionalVPCs) != 1 {
		t.Fatalf("expected 1 additional VPC, got %d", len(ko.Spec.AdditionalVPCs))
	}
	if *ko.Spec.AdditionalVPCs[0].VPCID != "vpc-additional" {
		t.Errorf("expected vpc-additional, got %s", *ko.Spec.AdditionalVPCs[0].VPCID)
	}
}

// mockVPCClient records calls made to Associate/Disassociate.
type mockVPCClient struct {
	getOutput     *svcsdk.GetHostedZoneOutput
	getErr        error
	associated    []string // vpcIDs passed to Associate
	disassociated []string // vpcIDs passed to Disassociate
	assocErr      error    // optional error to return on every Associate call
	disassocErr   error    // optional error to return on every Disassociate call
}

func (m *mockVPCClient) GetHostedZone(_ context.Context, _ *svcsdk.GetHostedZoneInput, _ ...func(*svcsdk.Options)) (*svcsdk.GetHostedZoneOutput, error) {
	return m.getOutput, m.getErr
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

func makeResource(vpcID, vpcRegion string, additionalVPCIDs []string) *resource {
	r := &resource{ko: &svcapitypes.HostedZone{}}
	r.ko.Status.ID = aws.String("/hostedzone/Z123")
	r.ko.Spec.VPC = &svcapitypes.VPC{
		VPCID:     aws.String(vpcID),
		VPCRegion: aws.String(vpcRegion),
	}
	for _, id := range additionalVPCIDs {
		vid := id
		r.ko.Spec.AdditionalVPCs = append(r.ko.Spec.AdditionalVPCs, &svcapitypes.VPC{
			VPCID:     &vid,
			VPCRegion: aws.String("us-east-1"),
		})
	}
	return r
}

func makeGetOutput(vpcIDs ...string) *svcsdk.GetHostedZoneOutput {
	out := &svcsdk.GetHostedZoneOutput{}
	for _, id := range vpcIDs {
		vpcID := id
		out.VPCs = append(out.VPCs, svcsdktypes.VPC{
			VPCId:     &vpcID,
			VPCRegion: svcsdktypes.VPCRegionUsEast1,
		})
	}
	return out
}

// Test: swapping Spec.VPC from vpc-A to vpc-B associates vpc-B and disassociates vpc-A.
func TestSyncVPCAssociations_SwapPrimaryVPC(t *testing.T) {
	// desired: primary=vpc-B; AWS currently has vpc-A only
	desired := makeResource("vpc-B", "us-east-1", nil)
	latest := makeResource("vpc-A", "us-east-1", nil)
	latest.ko.Status.ID = desired.ko.Status.ID

	mock := &mockVPCClient{getOutput: makeGetOutput("vpc-A")}
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

	mock := &mockVPCClient{
		getOutput:   makeGetOutput("vpc-A"),
		disassocErr: &svcsdktypes.LastVPCAssociation{Message: aws.String("last VPC")},
	}
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

// Test: compareAdditionalVPCs detects a nil VPCID in desired list as a delta.
func TestCompareAdditionalVPCs_NilVPCIDInA(t *testing.T) {
	a := &resource{ko: &svcapitypes.HostedZone{}}
	a.ko.Spec.AdditionalVPCs = []*svcapitypes.VPC{
		{VPCID: nil, VPCRegion: aws.String("us-east-1")},
	}
	b := &resource{ko: &svcapitypes.HostedZone{}}
	b.ko.Spec.AdditionalVPCs = []*svcapitypes.VPC{
		{VPCID: aws.String("vpc-1"), VPCRegion: aws.String("us-east-1")},
	}

	delta := ackcompare.NewDelta()
	compareAdditionalVPCs(delta, a, b)

	if !delta.DifferentAt("Spec.AdditionalVPCs") {
		t.Error("expected delta for Spec.AdditionalVPCs when a has nil VPCID, got none")
	}
}

// Test: same VPCs in different order — no delta expected.
func TestCompareAdditionalVPCs_SameVPCsDifferentOrder(t *testing.T) {
	a := &resource{ko: &svcapitypes.HostedZone{}}
	a.ko.Spec.AdditionalVPCs = []*svcapitypes.VPC{
		{VPCID: aws.String("vpc-1"), VPCRegion: aws.String("us-east-1")},
		{VPCID: aws.String("vpc-2"), VPCRegion: aws.String("us-west-2")},
	}
	b := &resource{ko: &svcapitypes.HostedZone{}}
	b.ko.Spec.AdditionalVPCs = []*svcapitypes.VPC{
		{VPCID: aws.String("vpc-2"), VPCRegion: aws.String("us-west-2")},
		{VPCID: aws.String("vpc-1"), VPCRegion: aws.String("us-east-1")},
	}
	delta := ackcompare.NewDelta()
	compareAdditionalVPCs(delta, a, b)
	if delta.DifferentAt("Spec.AdditionalVPCs") {
		t.Error("expected no delta for same VPCs in different order")
	}
}

// Test: same VPC IDs but different regions — delta expected.
func TestCompareAdditionalVPCs_SameIDDifferentRegion(t *testing.T) {
	a := &resource{ko: &svcapitypes.HostedZone{}}
	a.ko.Spec.AdditionalVPCs = []*svcapitypes.VPC{
		{VPCID: aws.String("vpc-1"), VPCRegion: aws.String("us-east-1")},
	}
	b := &resource{ko: &svcapitypes.HostedZone{}}
	b.ko.Spec.AdditionalVPCs = []*svcapitypes.VPC{
		{VPCID: aws.String("vpc-1"), VPCRegion: aws.String("us-west-2")},
	}
	delta := ackcompare.NewDelta()
	compareAdditionalVPCs(delta, a, b)
	if !delta.DifferentAt("Spec.AdditionalVPCs") {
		t.Error("expected delta for same VPC ID with different region")
	}
}

// Test: both nil/empty lists — no delta expected.
func TestCompareAdditionalVPCs_BothEmpty(t *testing.T) {
	a := &resource{ko: &svcapitypes.HostedZone{}}
	b := &resource{ko: &svcapitypes.HostedZone{}}
	delta := ackcompare.NewDelta()
	compareAdditionalVPCs(delta, a, b)
	if delta.DifferentAt("Spec.AdditionalVPCs") {
		t.Error("expected no delta for both empty AdditionalVPCs")
	}
}

// Test: nil VPCID in b — delta expected (already handled by existing b-iteration check).
func TestCompareAdditionalVPCs_NilVPCIDInB(t *testing.T) {
	a := &resource{ko: &svcapitypes.HostedZone{}}
	a.ko.Spec.AdditionalVPCs = []*svcapitypes.VPC{
		{VPCID: aws.String("vpc-1"), VPCRegion: aws.String("us-east-1")},
	}
	b := &resource{ko: &svcapitypes.HostedZone{}}
	b.ko.Spec.AdditionalVPCs = []*svcapitypes.VPC{
		{VPCID: nil, VPCRegion: aws.String("us-east-1")},
	}
	delta := ackcompare.NewDelta()
	compareAdditionalVPCs(delta, a, b)
	if !delta.DifferentAt("Spec.AdditionalVPCs") {
		t.Error("expected delta when b has nil VPCID")
	}
}

// Test: GetHostedZone error is propagated as-is.
func TestSyncVPCAssociations_GetHostedZoneError(t *testing.T) {
	desired := makeResource("vpc-A", "us-east-1", nil)
	latest := makeResource("vpc-A", "us-east-1", nil)

	mock := &mockVPCClient{getErr: errors.New("iam permission denied")}
	rm := &resourceManager{}

	err := rm.syncVPCAssociations(context.Background(), mock, desired, latest)
	if err == nil {
		t.Fatal("expected error from GetHostedZone, got nil")
	}
	if err.Error() != "iam permission denied" {
		t.Errorf("expected 'iam permission denied', got %q", err.Error())
	}
}

// Test: ConflictingDomainExists on AssociateVPC is treated as idempotent (no error returned).
func TestSyncVPCAssociations_ConflictingDomainExistsIdempotent(t *testing.T) {
	// desired has vpc-B as additional; AWS currently has only vpc-A
	desired := makeResource("vpc-A", "us-east-1", []string{"vpc-B"})
	latest := makeResource("vpc-A", "us-east-1", nil)
	latest.ko.Status.ID = desired.ko.Status.ID

	mock := &mockVPCClient{
		getOutput: makeGetOutput("vpc-A"),
		assocErr:  &svcsdktypes.ConflictingDomainExists{Message: aws.String("already associated")},
	}
	rm := &resourceManager{}

	err := rm.syncVPCAssociations(context.Background(), mock, desired, latest)
	if err != nil {
		t.Fatalf("expected nil error for ConflictingDomainExists, got: %v", err)
	}
}
