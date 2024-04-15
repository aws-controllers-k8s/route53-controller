package health_check

import (
	"context"
	"fmt"
	"time"

	svcapitypes "github.com/aws-controllers-k8s/route53-controller/apis/v1alpha1"
	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	ackrtlog "github.com/aws-controllers-k8s/runtime/pkg/runtime/log"
	svcsdk "github.com/aws/aws-sdk-go/service/route53"
)

// getCallerReference will generate a CallerReference for a given health check
// using the current timestamp, so that it produces a unique value
func getCallerReference() string {
	return fmt.Sprintf("%d", time.Now().UnixMilli())
}

func (rm *resourceManager) customUpdateHealthCheck(
	ctx context.Context,
	desired *resource,
	latest *resource,
	delta *ackcompare.Delta,
) (updated *resource, err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.customUpdateHealthCheck")
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

	resourceId := latest.ko.Status.ID

	desiredTags := ToACKTags(desired.ko.Spec.Tags)
	latestTags := ToACKTags(latest.ko.Spec.Tags)

	added, _, removed := ackcompare.GetTagsDifference(latestTags, desiredTags)

	toAdd := FromACKTags(added)

	var toDeleteTagKeys []*string
	for k, _ := range removed {
		toDeleteTagKeys = append(toDeleteTagKeys, &k)
	}

	resourceType := "HealthCheck"
	if len(toDeleteTagKeys) > 0 {
		rlog.Debug("removing tags from HealthCheck resource", "tags", toDeleteTagKeys)
		_, err = rm.sdkapi.ChangeTagsForResource(
			&svcsdk.ChangeTagsForResourceInput{
				ResourceId:    resourceId,
				RemoveTagKeys: toDeleteTagKeys,
				ResourceType:  &resourceType,
			},
		)
		rm.metrics.RecordAPICall("UPDATE", "DeleteTags", err)
		if err != nil {
			return err
		}

	}

	if len(toAdd) > 0 {
		rlog.Debug("adding tags to HealthCheck resource", "tags", toAdd)
		_, err = rm.sdkapi.ChangeTagsForResource(
			&svcsdk.ChangeTagsForResourceInput{
				ResourceId:   resourceId,
				AddTags:      rm.sdkTags(toAdd),
				ResourceType: &resourceType,
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
) (sdktags []*svcsdk.Tag) {

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
