package health_check

import (
	"context"
	"fmt"
	"strings"
	"time"

	svcapitypes "github.com/aws-controllers-k8s/route53-controller/apis/v1alpha1"
	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	ackerr "github.com/aws-controllers-k8s/runtime/pkg/errors"
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
	exit := rlog.Trace("rm.sdkUpdate")
	defer func() {
		exit(err)
	}()

	// Do not proceed with update if an immutable field was updated
	if immutableFieldChanges := rm.getImmutableFieldChanges(delta); len(immutableFieldChanges) > 0 {
		msg := fmt.Sprintf("Immutable Spec fields have been modified: %s", strings.Join(immutableFieldChanges, ","))
		return nil, ackerr.NewTerminalError(fmt.Errorf(msg))
	}

	// Merge in the information we read from the API call above to the copy of
	// the original Kubernetes object we passed to the function
	ko := desired.ko.DeepCopy()

	input := &svcsdk.UpdateHealthCheckInput{}

	if desired.ko.Status.ID != nil {
		input.SetHealthCheckId(*desired.ko.Status.ID)
	}
	if desired.ko.Status.HealthCheckVersion != nil {
		input.SetHealthCheckVersion(*desired.ko.Status.HealthCheckVersion)
	}

	if desired.ko.Spec.HealthCheckConfig != nil {
		if desired.ko.Spec.HealthCheckConfig.AlarmIdentifier != nil {
			alarm_identifier := &svcsdk.AlarmIdentifier{}
			if desired.ko.Spec.HealthCheckConfig.AlarmIdentifier.Name != nil {
				alarm_identifier.SetName(*desired.ko.Spec.HealthCheckConfig.AlarmIdentifier.Name)
			}
			if desired.ko.Spec.HealthCheckConfig.AlarmIdentifier.Region != nil {
				alarm_identifier.SetRegion(*desired.ko.Spec.HealthCheckConfig.AlarmIdentifier.Region)
			}
			input.SetAlarmIdentifier(alarm_identifier)
		}
		if desired.ko.Spec.HealthCheckConfig.ChildHealthChecks != nil {
			child_health_checks := []*string{}
			for _, item := range desired.ko.Spec.HealthCheckConfig.ChildHealthChecks {
				child_health_check := *item
				child_health_checks = append(child_health_checks, &child_health_check)
			}
			input.SetChildHealthChecks(child_health_checks)
		}
		if desired.ko.Spec.HealthCheckConfig.Disabled != nil {
			input.SetDisabled(*desired.ko.Spec.HealthCheckConfig.Disabled)
		}
		if desired.ko.Spec.HealthCheckConfig.EnableSNI != nil {
			input.SetEnableSNI(*desired.ko.Spec.HealthCheckConfig.EnableSNI)
		}
		if desired.ko.Spec.HealthCheckConfig.FailureThreshold != nil {
			input.SetFailureThreshold(*desired.ko.Spec.HealthCheckConfig.FailureThreshold)
		}
		if desired.ko.Spec.HealthCheckConfig.FullyQualifiedDomainName != nil {
			input.SetFullyQualifiedDomainName(*desired.ko.Spec.HealthCheckConfig.FullyQualifiedDomainName)
		}
		if desired.ko.Spec.HealthCheckConfig.HealthThreshold != nil {
			input.SetHealthThreshold(*desired.ko.Spec.HealthCheckConfig.HealthThreshold)
		}
		if desired.ko.Spec.HealthCheckConfig.IPAddress != nil {
			input.SetIPAddress(*desired.ko.Spec.HealthCheckConfig.IPAddress)
		}
		if desired.ko.Spec.HealthCheckConfig.InsufficientDataHealthStatus != nil {
			input.SetInsufficientDataHealthStatus(*desired.ko.Spec.HealthCheckConfig.InsufficientDataHealthStatus)
		}
		if desired.ko.Spec.HealthCheckConfig.Inverted != nil {
			input.SetInverted(*desired.ko.Spec.HealthCheckConfig.Inverted)
		}
		if desired.ko.Spec.HealthCheckConfig.Port != nil {
			input.SetPort(*desired.ko.Spec.HealthCheckConfig.Port)
		}
		if desired.ko.Spec.HealthCheckConfig.Regions != nil {
			regions := []*string{}
			for _, item := range desired.ko.Spec.HealthCheckConfig.Regions {
				region := *item
				regions = append(regions, &region)
			}
			input.SetRegions(regions)
		}
		if desired.ko.Spec.HealthCheckConfig.ResourcePath != nil {
			input.SetResourcePath(*desired.ko.Spec.HealthCheckConfig.ResourcePath)
		}
		if desired.ko.Spec.HealthCheckConfig.SearchString != nil {
			input.SetSearchString(*desired.ko.Spec.HealthCheckConfig.SearchString)
		}
	}

	var resp *svcsdk.UpdateHealthCheckOutput
	_ = resp
	resp, err = rm.sdkapi.UpdateHealthCheckWithContext(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "UpdateHealthCheck", err)
	if err != nil {
		return nil, err
	}

	if resp.HealthCheck.CallerReference != nil {
		ko.Status.CallerReference = resp.HealthCheck.CallerReference
	} else {
		ko.Status.CallerReference = nil
	}
	if resp.HealthCheck.CloudWatchAlarmConfiguration != nil {
		f1 := &svcapitypes.CloudWatchAlarmConfiguration{}
		if resp.HealthCheck.CloudWatchAlarmConfiguration.ComparisonOperator != nil {
			f1.ComparisonOperator = resp.HealthCheck.CloudWatchAlarmConfiguration.ComparisonOperator
		}
		if resp.HealthCheck.CloudWatchAlarmConfiguration.Dimensions != nil {
			f1f1 := []*svcapitypes.Dimension{}
			for _, f1f1iter := range resp.HealthCheck.CloudWatchAlarmConfiguration.Dimensions {
				f1f1elem := &svcapitypes.Dimension{}
				if f1f1iter.Name != nil {
					f1f1elem.Name = f1f1iter.Name
				}
				if f1f1iter.Value != nil {
					f1f1elem.Value = f1f1iter.Value
				}
				f1f1 = append(f1f1, f1f1elem)
			}
			f1.Dimensions = f1f1
		}
		if resp.HealthCheck.CloudWatchAlarmConfiguration.EvaluationPeriods != nil {
			f1.EvaluationPeriods = resp.HealthCheck.CloudWatchAlarmConfiguration.EvaluationPeriods
		}
		if resp.HealthCheck.CloudWatchAlarmConfiguration.MetricName != nil {
			f1.MetricName = resp.HealthCheck.CloudWatchAlarmConfiguration.MetricName
		}
		if resp.HealthCheck.CloudWatchAlarmConfiguration.Namespace != nil {
			f1.Namespace = resp.HealthCheck.CloudWatchAlarmConfiguration.Namespace
		}
		if resp.HealthCheck.CloudWatchAlarmConfiguration.Period != nil {
			f1.Period = resp.HealthCheck.CloudWatchAlarmConfiguration.Period
		}
		if resp.HealthCheck.CloudWatchAlarmConfiguration.Statistic != nil {
			f1.Statistic = resp.HealthCheck.CloudWatchAlarmConfiguration.Statistic
		}
		if resp.HealthCheck.CloudWatchAlarmConfiguration.Threshold != nil {
			f1.Threshold = resp.HealthCheck.CloudWatchAlarmConfiguration.Threshold
		}
		ko.Status.CloudWatchAlarmConfiguration = f1
	} else {
		ko.Status.CloudWatchAlarmConfiguration = nil
	}
	if resp.HealthCheck.HealthCheckConfig != nil {
		f2 := &svcapitypes.HealthCheckConfig{}
		if resp.HealthCheck.HealthCheckConfig.AlarmIdentifier != nil {
			f2f0 := &svcapitypes.AlarmIdentifier{}
			if resp.HealthCheck.HealthCheckConfig.AlarmIdentifier.Name != nil {
				f2f0.Name = resp.HealthCheck.HealthCheckConfig.AlarmIdentifier.Name
			}
			if resp.HealthCheck.HealthCheckConfig.AlarmIdentifier.Region != nil {
				f2f0.Region = resp.HealthCheck.HealthCheckConfig.AlarmIdentifier.Region
			}
			f2.AlarmIdentifier = f2f0
		}
		if resp.HealthCheck.HealthCheckConfig.ChildHealthChecks != nil {
			f2f1 := []*string{}
			for _, f2f1iter := range resp.HealthCheck.HealthCheckConfig.ChildHealthChecks {
				var f2f1elem string
				f2f1elem = *f2f1iter
				f2f1 = append(f2f1, &f2f1elem)
			}
			f2.ChildHealthChecks = f2f1
		}
		if resp.HealthCheck.HealthCheckConfig.Disabled != nil {
			f2.Disabled = resp.HealthCheck.HealthCheckConfig.Disabled
		}
		if resp.HealthCheck.HealthCheckConfig.EnableSNI != nil {
			f2.EnableSNI = resp.HealthCheck.HealthCheckConfig.EnableSNI
		}
		if resp.HealthCheck.HealthCheckConfig.FailureThreshold != nil {
			f2.FailureThreshold = resp.HealthCheck.HealthCheckConfig.FailureThreshold
		}
		if resp.HealthCheck.HealthCheckConfig.FullyQualifiedDomainName != nil {
			f2.FullyQualifiedDomainName = resp.HealthCheck.HealthCheckConfig.FullyQualifiedDomainName
		}
		if resp.HealthCheck.HealthCheckConfig.HealthThreshold != nil {
			f2.HealthThreshold = resp.HealthCheck.HealthCheckConfig.HealthThreshold
		}
		if resp.HealthCheck.HealthCheckConfig.IPAddress != nil {
			f2.IPAddress = resp.HealthCheck.HealthCheckConfig.IPAddress
		}
		if resp.HealthCheck.HealthCheckConfig.InsufficientDataHealthStatus != nil {
			f2.InsufficientDataHealthStatus = resp.HealthCheck.HealthCheckConfig.InsufficientDataHealthStatus
		}
		if resp.HealthCheck.HealthCheckConfig.Inverted != nil {
			f2.Inverted = resp.HealthCheck.HealthCheckConfig.Inverted
		}
		if resp.HealthCheck.HealthCheckConfig.MeasureLatency != nil {
			f2.MeasureLatency = resp.HealthCheck.HealthCheckConfig.MeasureLatency
		}
		if resp.HealthCheck.HealthCheckConfig.Port != nil {
			f2.Port = resp.HealthCheck.HealthCheckConfig.Port
		}
		if resp.HealthCheck.HealthCheckConfig.Regions != nil {
			f2f12 := []*string{}
			for _, f2f12iter := range resp.HealthCheck.HealthCheckConfig.Regions {
				var f2f12elem string
				f2f12elem = *f2f12iter
				f2f12 = append(f2f12, &f2f12elem)
			}
			f2.Regions = f2f12
		}
		if resp.HealthCheck.HealthCheckConfig.RequestInterval != nil {
			f2.RequestInterval = resp.HealthCheck.HealthCheckConfig.RequestInterval
		}
		if resp.HealthCheck.HealthCheckConfig.ResourcePath != nil {
			f2.ResourcePath = resp.HealthCheck.HealthCheckConfig.ResourcePath
		}
		if resp.HealthCheck.HealthCheckConfig.RoutingControlArn != nil {
			f2.RoutingControlARN = resp.HealthCheck.HealthCheckConfig.RoutingControlArn
		}
		if resp.HealthCheck.HealthCheckConfig.SearchString != nil {
			f2.SearchString = resp.HealthCheck.HealthCheckConfig.SearchString
		}
		if resp.HealthCheck.HealthCheckConfig.Type != nil {
			f2.Type = resp.HealthCheck.HealthCheckConfig.Type
		}
		ko.Spec.HealthCheckConfig = f2
	} else {
		ko.Spec.HealthCheckConfig = nil
	}
	if resp.HealthCheck.HealthCheckVersion != nil {
		ko.Status.HealthCheckVersion = resp.HealthCheck.HealthCheckVersion
	} else {
		ko.Status.HealthCheckVersion = nil
	}
	if resp.HealthCheck.Id != nil {
		ko.Status.ID = resp.HealthCheck.Id
	} else {
		ko.Status.ID = nil
	}
	if resp.HealthCheck.LinkedService != nil {
		f5 := &svcapitypes.LinkedService{}
		if resp.HealthCheck.LinkedService.Description != nil {
			f5.Description = resp.HealthCheck.LinkedService.Description
		}
		if resp.HealthCheck.LinkedService.ServicePrincipal != nil {
			f5.ServicePrincipal = resp.HealthCheck.LinkedService.ServicePrincipal
		}
		ko.Status.LinkedService = f5
	} else {
		ko.Status.LinkedService = nil
	}

	rm.setStatusDefaults(ko)
	updated = &resource{ko}

	if err := rm.syncTags(ctx, desired, updated); err != nil {
		return nil, err
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

	resourceType := "healthcheck"
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
