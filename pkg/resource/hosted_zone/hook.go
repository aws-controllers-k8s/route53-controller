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

	desiredTags := ToACKTags(desired.ko.Spec.Tags)
	latestTags := ToACKTags(latest.ko.Spec.Tags)

	added, _, removed := ackcompare.GetTagsDifference(latestTags, desiredTags)

	toAdd := FromACKTags(added)

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
		desiredTags := ToACKTags(a.ko.Spec.Tags)
		latestTags := ToACKTags(b.ko.Spec.Tags)

		added, _, removed := ackcompare.GetTagsDifference(latestTags, desiredTags)

		toAdd := FromACKTags(added)
		toDelete := FromACKTags(removed)

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
