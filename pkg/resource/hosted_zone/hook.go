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
	GetHostedZone(ctx context.Context, input *svcsdk.GetHostedZoneInput, opts ...func(*svcsdk.Options)) (*svcsdk.GetHostedZoneOutput, error)
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

	if delta.DifferentAt("Spec.VPC") || delta.DifferentAt("Spec.AdditionalVPCs") {
		if err := rm.syncVPCAssociations(ctx, rm.sdkapi, desired, latest); err != nil {
			return nil, err
		}
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

// populateAdditionalVPCs fills ko.Spec.AdditionalVPCs from the AWS VPC list,
// excluding the primary VPC (Spec.VPC) to preserve the no-duplication invariant.
func populateAdditionalVPCs(ko *svcapitypes.HostedZone, vpcs []svcsdktypes.VPC) {
	primaryVPCID := ""
	if ko.Spec.VPC != nil && ko.Spec.VPC.VPCID != nil {
		primaryVPCID = *ko.Spec.VPC.VPCID
	}
	var additionalVPCs []*svcapitypes.VPC
	for _, v := range vpcs {
		if v.VPCId == nil || *v.VPCId == primaryVPCID {
			continue
		}
		region := string(v.VPCRegion)
		additionalVPCs = append(additionalVPCs, &svcapitypes.VPC{
			VPCID:     v.VPCId,
			VPCRegion: &region,
		})
	}
	ko.Spec.AdditionalVPCs = additionalVPCs
}

// compareAdditionalVPCs is a custom comparison function for comparing lists of
// VPC structs where the order of the structs in the list is not important.
func compareAdditionalVPCs(
	delta *ackcompare.Delta,
	a *resource,
	b *resource,
) {
	aVPCs := a.ko.Spec.AdditionalVPCs
	bVPCs := b.ko.Spec.AdditionalVPCs
	if len(aVPCs) != len(bVPCs) {
		delta.Add("Spec.AdditionalVPCs", aVPCs, bVPCs)
		return
	}
	if len(aVPCs) == 0 {
		return
	}
	// If any entry in a has a nil VPCID or VPCRegion, we cannot build a
	// reliable set — treat it as a difference.
	for _, v := range aVPCs {
		if v.VPCID == nil || v.VPCRegion == nil {
			delta.Add("Spec.AdditionalVPCs", aVPCs, bVPCs)
			return
		}
	}
	setA := additionalVPCsToSet(aVPCs)
	for _, v := range bVPCs {
		if v.VPCID == nil || v.VPCRegion == nil {
			delta.Add("Spec.AdditionalVPCs", aVPCs, bVPCs)
			return
		}
		if region, ok := setA[*v.VPCID]; !ok || region != *v.VPCRegion {
			delta.Add("Spec.AdditionalVPCs", aVPCs, bVPCs)
			return
		}
	}
}

// additionalVPCsToSet converts a slice of VPCs to a map of VPCID → VPCRegion.
func additionalVPCsToSet(vpcs []*svcapitypes.VPC) map[string]string {
	set := make(map[string]string, len(vpcs))
	for _, v := range vpcs {
		if v.VPCID != nil && v.VPCRegion != nil {
			set[*v.VPCID] = *v.VPCRegion
		}
	}
	return set
}

// syncVPCAssociations reconciles VPC associations for a private hosted zone.
// It treats the union of Spec.VPC and Spec.AdditionalVPCs as the full desired
// set. All VPCs returned by GetHostedZone are the current set. VPCs in desired
// but not current are associated; VPCs in current but not desired are
// disassociated. LastVPCAssociation errors are surfaced as terminal conditions.
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

	// Fetch current VPC associations from AWS.
	resp, err := client.GetHostedZone(ctx, &svcsdk.GetHostedZoneInput{Id: hostedZoneID})
	if rm.metrics != nil {
		rm.metrics.RecordAPICall("READ_ONE", "GetHostedZone", err)
	}
	if err != nil {
		return err
	}

	// Build full desired set: Spec.VPC (if non-nil) + all Spec.AdditionalVPCs.
	desiredSet := make(map[string]*svcapitypes.VPC)
	if desired.ko.Spec.VPC != nil && desired.ko.Spec.VPC.VPCID != nil {
		desiredSet[*desired.ko.Spec.VPC.VPCID] = desired.ko.Spec.VPC
	}
	for _, v := range desired.ko.Spec.AdditionalVPCs {
		if v.VPCID == nil || v.VPCRegion == nil {
			rlog.Debug("skipping AdditionalVPCs entry with nil VPCID or VPCRegion")
			continue
		}
		desiredSet[*v.VPCID] = v
	}

	// Build current set from ALL VPCs returned by AWS — no exclusions.
	current := make(map[string]svcsdktypes.VPCRegion)
	for _, v := range resp.VPCs {
		if v.VPCId != nil {
			current[*v.VPCId] = v.VPCRegion
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
