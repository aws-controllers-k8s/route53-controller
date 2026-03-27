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
	"fmt"
	"strings"
	"time"

	svcapitypes "github.com/aws-controllers-k8s/route53-controller/apis/v1alpha1"
	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	ackerr "github.com/aws-controllers-k8s/runtime/pkg/errors"
	ackrtlog "github.com/aws-controllers-k8s/runtime/pkg/runtime/log"
	"github.com/aws/aws-sdk-go-v2/aws"
	svcsdk "github.com/aws/aws-sdk-go-v2/service/route53"
	svcsdktypes "github.com/aws/aws-sdk-go-v2/service/route53/types"
)

// vpcAssociationClient is the subset of the Route53 API used by syncVPCAssociations.
// *svcsdk.Client satisfies this interface; it is defined separately to allow unit testing.
type vpcAssociationClient interface {
	AssociateVPCWithHostedZone(ctx context.Context, input *svcsdk.AssociateVPCWithHostedZoneInput, opts ...func(*svcsdk.Options)) (*svcsdk.AssociateVPCWithHostedZoneOutput, error)
	DisassociateVPCFromHostedZone(ctx context.Context, input *svcsdk.DisassociateVPCFromHostedZoneInput, opts ...func(*svcsdk.Options)) (*svcsdk.DisassociateVPCFromHostedZoneOutput, error)
}

// getCallerReference will generate a CallerReference for a given hosted zone
// using the name of the zone and the current timestamp, so that it produces a
// unique value
func getCallerReference(zone *svcapitypes.HostedZone) string {
	return fmt.Sprintf("%s-%d", *zone.Spec.Name, time.Now().UnixMilli())
}

func (rm *resourceManager) customUpdateHostedZone(
	ctx context.Context,
	desired *resource,
	latest *resource,
	delta *ackcompare.Delta,
) (updated *resource, err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.customUpdateHostedZone")
	defer exit(err)

	if err := validateVPCFields(desired.ko.Spec); err != nil {
		return nil, err
	}

	// Default `updated` to `desired` because it is likely
	// EC2 `modify` APIs do NOT return output, only errors.
	// If the `modify` calls (i.e. `sync`) do NOT return
	// an error, then the update was successful and desired.Spec
	// (now updated.Spec) reflects the latest resource state.
	updated = rm.concreteResource(desired.DeepCopy())

	if delta.DifferentAt("Spec.Tags") {
		if err := rm.syncTags(ctx, desired, latest); err != nil {
			return nil, err
		}
	}

	if delta.DifferentAt("Spec.VPCs") || delta.DifferentAt("Spec.VPC") {
		if err := rm.syncVPCAssociations(ctx, rm.sdkapi, desired, latest); err != nil {
			return nil, err
		}
		// Reflect the new association state in status immediately rather than
		// waiting for the next sdkFind call.
		// if len(desired.ko.Spec.VPCs) > 0 {
		// 	updated.ko.Status.AssociatedVPCs = desired.ko.Spec.VPCs
		// } else if desired.ko.Spec.VPC != nil {
		// 	updated.ko.Status.AssociatedVPCs = []*svcapitypes.VPC{desired.ko.Spec.VPC}
		// }
	}

	return updated, nil
}

// syncTags used to keep tags in sync by calling Create and Delete API's
func (rm *resourceManager) syncTags(
	ctx context.Context,
	desired *resource,
	latest *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.syncTags")
	defer func(err error) {
		exit(err)
	}(err)

	//
	resourceId := aws.String(strings.TrimPrefix(*latest.ko.Status.ID, "/hostedzone/"))

	desiredTags, _ := convertToOrderedACKTags(desired.ko.Spec.Tags)
	latestTags, _ := convertToOrderedACKTags(latest.ko.Spec.Tags)

	added, _, removed := ackcompare.GetTagsDifference(latestTags, desiredTags)

	toAdd := fromACKTags(added, nil)

	var toDeleteTagKeys []*string
	for k := range removed {
		toDeleteTagKeys = append(toDeleteTagKeys, &k)
	}

	resourceType := svcsdktypes.TagResourceTypeHostedzone
	if len(toDeleteTagKeys) > 0 {
		rlog.Debug("removing tags from HostedZone resource", "tags", toDeleteTagKeys)
		_, err = rm.sdkapi.ChangeTagsForResource(
			ctx,
			&svcsdk.ChangeTagsForResourceInput{
				ResourceId:    resourceId,
				RemoveTagKeys: aws.ToStringSlice(toDeleteTagKeys),
				ResourceType:  resourceType,
			},
		)
		rm.metrics.RecordAPICall("UPDATE", "DeleteTags", err)
		if err != nil {
			return err
		}

	}

	if len(toAdd) > 0 {
		rlog.Debug("adding tags to HostedZone resource", "tags", toAdd)
		_, err = rm.sdkapi.ChangeTagsForResource(
			ctx,
			&svcsdk.ChangeTagsForResourceInput{
				ResourceId:   resourceId,
				AddTags:      rm.sdkTags(toAdd),
				ResourceType: resourceType,
			},
		)
		rm.metrics.RecordAPICall("UPDATE", "CreateTags", err)
		if err != nil {
			return err
		}
	}

	return nil
}

// sdkTags converts *svcapitypes.Tag array to a *svcsdk.Tag array
func (rm *resourceManager) sdkTags(
	tags []*svcapitypes.Tag,
) (sdktags []svcsdktypes.Tag) {

	for _, i := range tags {
		sdktag := rm.newTag(*i)
		sdktags = append(sdktags, sdktag)
	}

	return sdktags
}

// compareTags is a custom comparison function for comparing lists of Tag
// structs where the order of the structs in the list is not important.
func compareTags(
	delta *ackcompare.Delta,
	a *resource,
	b *resource,
) {
	if len(a.ko.Spec.Tags) != len(b.ko.Spec.Tags) {
		delta.Add("Spec.Tags", a.ko.Spec.Tags, b.ko.Spec.Tags)
	} else if len(a.ko.Spec.Tags) > 0 {
		desiredTags, _ := convertToOrderedACKTags(a.ko.Spec.Tags)
		latestTags, _ := convertToOrderedACKTags(b.ko.Spec.Tags)

		added, _, removed := ackcompare.GetTagsDifference(latestTags, desiredTags)

		toAdd := fromACKTags(added, nil)
		toDelete := fromACKTags(removed, nil)

		if len(toAdd) != 0 || len(toDelete) != 0 {
			delta.Add("Spec.Tags", a.ko.Spec.Tags, b.ko.Spec.Tags)
		}
	}
}

func (rm *resourceManager) setResourceAdditionalFields(
	ctx context.Context,
	ko *svcapitypes.HostedZone,
) (err error) {
	if ko.Status.ID != nil {
		// Get the tags for hosted_zone resource
		tags_input := svcsdk.ListTagsForResourceInput{}
		tags_input.ResourceId = aws.String(strings.TrimPrefix(*ko.Status.ID, "/hostedzone/"))
		tags_input.ResourceType = svcsdktypes.TagResourceTypeHostedzone
		var tags_resp *svcsdk.ListTagsForResourceOutput
		tags_resp, err = rm.sdkapi.ListTagsForResource(ctx, &tags_input)
		rm.metrics.RecordAPICall("READ_ONE", "ListTagsForResource", err)
		if err != nil {
			var notFound *svcsdktypes.NoSuchHostedZone
			if errors.As(err, &notFound) {
				return ackerr.NotFound
			}
			return err
		}

		tags := FromRoute53Tags(tags_resp.ResourceTagSet.Tags)

		ko.Spec.Tags = tags
	} else {
		ko.Spec.Tags = []*svcapitypes.Tag{}
	}
	return nil
}

// validateVPCFields checks for invalid VPC field combinations and returns a
// terminal error on conflict. Rules:
//   - spec.vpc and spec.vpcs are mutually exclusive.
//   - If hostedZoneConfig.privateZone is explicitly false, VPC fields are rejected.
//   - VPC presence alone (without privateZone: true) is accepted as implicit private
//     zone intent; Route 53 infers zone type from the VPC parameter.
//   - Each entry in spec.vpcs must have non-nil vpcID and vpcRegion.
//
// Only call on create and update paths.
func validateVPCFields(spec svcapitypes.HostedZoneSpec) error {
	if spec.VPC != nil && len(spec.VPCs) > 0 {
		return ackerr.NewTerminalError(fmt.Errorf(
			"spec.vpc and spec.vpcs are mutually exclusive; use spec.vpcs to manage all VPC associations",
		))
	}
	explicitlyPublic := spec.HostedZoneConfig != nil &&
		spec.HostedZoneConfig.PrivateZone != nil &&
		!*spec.HostedZoneConfig.PrivateZone
	if explicitlyPublic && (spec.VPC != nil || len(spec.VPCs) > 0) {
		return ackerr.NewTerminalError(fmt.Errorf(
			"spec.vpc and spec.vpcs are only valid for private hosted zones; remove hostedZoneConfig.privateZone: false, or remove VPC fields",
		))
	}
	for i, v := range spec.VPCs {
		if v == nil {
			return ackerr.NewTerminalError(fmt.Errorf("spec.vpcs[%d] must not be nil", i))
		}
		if v.VPCID == nil || *v.VPCID == "" {
			return ackerr.NewTerminalError(fmt.Errorf("spec.vpcs[%d].vpcID is required", i))
		}
		if v.VPCRegion == nil || *v.VPCRegion == "" {
			return ackerr.NewTerminalError(fmt.Errorf("spec.vpcs[%d].vpcRegion is required", i))
		}
	}
	return nil
}

// compareVPCs is a custom comparison function that diffs desired.Spec.VPCs
// against latest.Status.AssociatedVPCs (the AWS actual state). Order is not
// significant. Uses "Spec.VPCs" as the delta key — this must match what
// customUpdateHostedZone checks.
func compareVPCs(
	delta *ackcompare.Delta,
	a *resource,
	b *resource,
) {
	// On the legacy spec.vpc path, Spec.VPCs is nil — skip this comparison.
	// The Spec.VPC field is handled by compareVPC below.
	aVPCs := a.ko.Spec.VPCs
	if aVPCs == nil {
		return
	}
	bVPCs := b.ko.Spec.VPCs
	if len(aVPCs) != len(bVPCs) {
		delta.Add("Spec.VPCs", aVPCs, bVPCs)
		return
	}
	if len(aVPCs) == 0 {
		return
	}
	for _, v := range aVPCs {
		if v == nil || v.VPCID == nil || v.VPCRegion == nil {
			delta.Add("Spec.VPCs", aVPCs, bVPCs)
			return
		}
	}
	setA := vpcListToSet(aVPCs)
	for _, v := range bVPCs {
		if region, ok := setA[*v.VPCID]; !ok || region != *v.VPCRegion {
			delta.Add("Spec.VPCs", aVPCs, bVPCs)
			return
		}
	}
}

// compareVPC is a custom comparison function for the legacy spec.vpc path.
// It compares desired.Spec.VPC against the VPCs actually associated in AWS
// (latest.Status.AssociatedVPCs). This is necessary because sdkFind starts
// with ko = r.ko.DeepCopy(), so latest.Spec.VPC is always a copy of
// desired.Spec.VPC — the generated delta code for Spec.VPC never fires.
// Uses "Spec.VPC" as the delta key — must match what customUpdateHostedZone checks.
func compareVPC(
	delta *ackcompare.Delta,
	a *resource,
	b *resource,
) {
	// Only applies on the spec.vpc (legacy) path.
	// spec.vpcs path is handled by compareVPCs.
	if a.ko.Spec.VPC == nil {
		return
	}
	bVPCs := b.ko.Spec.VPCs
	if len(bVPCs) != 1 {
		delta.Add("Spec.VPC", a.ko.Spec.VPC, bVPCs)
		return
	}
	v := bVPCs[0]
	if v.VPCID == nil || v.VPCRegion == nil {
		delta.Add("Spec.VPC", a.ko.Spec.VPC, bVPCs)
		return
	}
	if a.ko.Spec.VPC.VPCID == nil || a.ko.Spec.VPC.VPCRegion == nil {
		delta.Add("Spec.VPC", a.ko.Spec.VPC, bVPCs)
		return
	}
	if *v.VPCID != *a.ko.Spec.VPC.VPCID || *v.VPCRegion != *a.ko.Spec.VPC.VPCRegion {
		delta.Add("Spec.VPC", a.ko.Spec.VPC, bVPCs)
	}
}

// vpcListToSet converts a slice of VPCs to a map of VPCID → VPCRegion.
func vpcListToSet(vpcs []*svcapitypes.VPC) map[string]string {
	set := make(map[string]string, len(vpcs))
	for _, v := range vpcs {
		if v.VPCID != nil && v.VPCRegion != nil {
			set[*v.VPCID] = *v.VPCRegion
		}
	}
	return set
}

// shouldRunVPCPreCleanup reports whether the delete path must disassociate
// extra VPCs before deleting the hosted zone. Route53 rejects DeleteHostedZone
// when more than one VPC is associated. The guard uses Status.AssociatedVPCs
// (authoritative AWS state) rather than Spec.VPCs so that a failed prior sync
// that left more VPCs associated than the spec reflects is still cleaned up.
// It applies on both the spec.vpcs path and the legacy spec.vpc path.
func shouldRunVPCPreCleanup(r *resource) bool {
	return len(r.ko.Spec.VPCs) > 1
}

// syncVPCAssociations reconciles VPC associations for a private hosted zone.
// When spec.vpcs is set (len > 0), the full desired set is Spec.VPCs.
// Otherwise (legacy path), the desired set is {Spec.VPC} if non-nil.
// The current set is read from latest.ko.Status.AssociatedVPCs, which is
// populated by sdkFind (read path) and seeded from the create response
// (create path). LastVPCAssociation errors are surfaced as terminal conditions.
func (rm *resourceManager) syncVPCAssociations(
	ctx context.Context,
	client vpcAssociationClient,
	desired *resource,
	latest *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.syncVPCAssociations")
	defer func() {
		exit(err)
	}()

	hostedZoneID := latest.ko.Status.ID

	// Build full desired set from whichever field is active.
	// spec.vpcs path: len(Spec.VPCs) > 0 — use the full VPCs list.
	// spec.vpc path (legacy): use Spec.VPC only.
	desiredSet := make(map[string]*svcapitypes.VPC)
	if len(desired.ko.Spec.VPCs) > 0 {
		for _, v := range desired.ko.Spec.VPCs {
			if v.VPCID == nil || v.VPCRegion == nil {
				rlog.Debug("skipping Spec.VPCs entry with nil VPCID or VPCRegion")
				continue
			}
			desiredSet[*v.VPCID] = v
		}
	} else if desired.ko.Spec.VPC != nil && desired.ko.Spec.VPC.VPCID != nil {
		desiredSet[*desired.ko.Spec.VPC.VPCID] = desired.ko.Spec.VPC
	}

	// Build current set from Status.AssociatedVPCs, which is populated by
	// sdkFind (read path) and seeded from the create response (create path).
	// This avoids an extra GetHostedZone call on every reconcile.
	current := make(map[string]svcsdktypes.VPCRegion)
	for _, v := range latest.ko.Spec.VPCs {
		if v.VPCID != nil && v.VPCRegion != nil {
			current[*v.VPCID] = svcsdktypes.VPCRegion(*v.VPCRegion)
		}
	}

	// Associate VPCs that are desired but not yet associated.
	for id, v := range desiredSet {
		if _, ok := current[id]; ok {
			continue // already associated
		}
		rlog.Debug("associating VPC with hosted zone", "vpcID", id)
		region := svcsdktypes.VPCRegion(*v.VPCRegion)
		_, err = client.AssociateVPCWithHostedZone(ctx, &svcsdk.AssociateVPCWithHostedZoneInput{
			HostedZoneId: hostedZoneID,
			VPC: &svcsdktypes.VPC{
				VPCId:     v.VPCID,
				VPCRegion: region,
			},
		})
		if rm.metrics != nil {
			rm.metrics.RecordAPICall("UPDATE", "AssociateVPCWithHostedZone", err)
		}
		if err != nil {
			var conflict *svcsdktypes.ConflictingDomainExists
			if errors.As(err, &conflict) {
				continue // already associated — idempotent
			}
			return err
		}
	}

	// Disassociate VPCs that are associated but no longer desired.
	for id, region := range current {
		if _, ok := desiredSet[id]; ok {
			continue
		}
		rlog.Debug("disassociating VPC from hosted zone", "vpcID", id)
		vpcID := id
		_, err = client.DisassociateVPCFromHostedZone(ctx, &svcsdk.DisassociateVPCFromHostedZoneInput{
			HostedZoneId: hostedZoneID,
			VPC: &svcsdktypes.VPC{
				VPCId:     &vpcID,
				VPCRegion: region,
			},
		})
		if rm.metrics != nil {
			rm.metrics.RecordAPICall("UPDATE", "DisassociateVPCFromHostedZone", err)
		}
		if err != nil {
			var notFound *svcsdktypes.VPCAssociationNotFound
			if errors.As(err, &notFound) {
				continue // already disassociated — idempotent
			}
			var lastVPC *svcsdktypes.LastVPCAssociation
			if errors.As(err, &lastVPC) {
				// Route53 rejects disassociation of the last VPC. This is a
				// permanent user error — the desired state is invalid. Surface
				// it as a terminal condition so the operator is notified.
				return ackerr.NewTerminalError(err)
			}
			return err
		}
	}

	return nil
}
