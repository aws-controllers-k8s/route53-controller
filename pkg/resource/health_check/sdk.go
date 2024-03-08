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

// Code generated by ack-generate. DO NOT EDIT.

package health_check

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	ackv1alpha1 "github.com/aws-controllers-k8s/runtime/apis/core/v1alpha1"
	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	ackcondition "github.com/aws-controllers-k8s/runtime/pkg/condition"
	ackerr "github.com/aws-controllers-k8s/runtime/pkg/errors"
	ackrequeue "github.com/aws-controllers-k8s/runtime/pkg/requeue"
	ackrtlog "github.com/aws-controllers-k8s/runtime/pkg/runtime/log"
	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/route53"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	svcapitypes "github.com/aws-controllers-k8s/route53-controller/apis/v1alpha1"
)

// Hack to avoid import errors during build...
var (
	_ = &metav1.Time{}
	_ = strings.ToLower("")
	_ = &aws.JSONValue{}
	_ = &svcsdk.Route53{}
	_ = &svcapitypes.HealthCheck{}
	_ = ackv1alpha1.AWSAccountID("")
	_ = &ackerr.NotFound
	_ = &ackcondition.NotManagedMessage
	_ = &reflect.Value{}
	_ = fmt.Sprintf("")
	_ = &ackrequeue.NoRequeue{}
)

// sdkFind returns SDK-specific information about a supplied resource
func (rm *resourceManager) sdkFind(
	ctx context.Context,
	r *resource,
) (latest *resource, err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.sdkFind")
	defer func() {
		exit(err)
	}()
	// If any required fields in the input shape are missing, AWS resource is
	// not created yet. Return NotFound here to indicate to callers that the
	// resource isn't yet created.
	if rm.requiredFieldsMissingFromReadOneInput(r) {
		return nil, ackerr.NotFound
	}

	input, err := rm.newDescribeRequestPayload(r)
	if err != nil {
		return nil, err
	}

	var resp *svcsdk.GetHealthCheckOutput
	resp, err = rm.sdkapi.GetHealthCheckWithContext(ctx, input)
	rm.metrics.RecordAPICall("READ_ONE", "GetHealthCheck", err)
	if err != nil {
		if reqErr, ok := ackerr.AWSRequestFailure(err); ok && reqErr.StatusCode() == 404 {
			return nil, ackerr.NotFound
		}
		if awsErr, ok := ackerr.AWSError(err); ok && awsErr.Code() == "NoSuchHealthCheck" {
			return nil, ackerr.NotFound
		}
		return nil, err
	}

	// Merge in the information we read from the API call above to the copy of
	// the original Kubernetes object we passed to the function
	ko := r.ko.DeepCopy()

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
	return &resource{ko}, nil
}

// requiredFieldsMissingFromReadOneInput returns true if there are any fields
// for the ReadOne Input shape that are required but not present in the
// resource's Spec or Status
func (rm *resourceManager) requiredFieldsMissingFromReadOneInput(
	r *resource,
) bool {
	return r.ko.Status.ID == nil

}

// newDescribeRequestPayload returns SDK-specific struct for the HTTP request
// payload of the Describe API call for the resource
func (rm *resourceManager) newDescribeRequestPayload(
	r *resource,
) (*svcsdk.GetHealthCheckInput, error) {
	res := &svcsdk.GetHealthCheckInput{}

	if r.ko.Status.ID != nil {
		res.SetHealthCheckId(*r.ko.Status.ID)
	}

	return res, nil
}

// sdkCreate creates the supplied resource in the backend AWS service API and
// returns a copy of the resource with resource fields (in both Spec and
// Status) filled in with values from the CREATE API operation's Output shape.
func (rm *resourceManager) sdkCreate(
	ctx context.Context,
	desired *resource,
) (created *resource, err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.sdkCreate")
	defer func() {
		exit(err)
	}()
	input, err := rm.newCreateRequestPayload(ctx, desired)
	if err != nil {
		return nil, err
	}

	var resp *svcsdk.CreateHealthCheckOutput
	_ = resp
	resp, err = rm.sdkapi.CreateHealthCheckWithContext(ctx, input)
	rm.metrics.RecordAPICall("CREATE", "CreateHealthCheck", err)
	if err != nil {
		return nil, err
	}
	// Merge in the information we read from the API call above to the copy of
	// the original Kubernetes object we passed to the function
	ko := desired.ko.DeepCopy()

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
	return &resource{ko}, nil
}

// newCreateRequestPayload returns an SDK-specific struct for the HTTP request
// payload of the Create API call for the resource
func (rm *resourceManager) newCreateRequestPayload(
	ctx context.Context,
	r *resource,
) (*svcsdk.CreateHealthCheckInput, error) {
	res := &svcsdk.CreateHealthCheckInput{}

	if r.ko.Spec.HealthCheckConfig != nil {
		f0 := &svcsdk.HealthCheckConfig{}
		if r.ko.Spec.HealthCheckConfig.AlarmIdentifier != nil {
			f0f0 := &svcsdk.AlarmIdentifier{}
			if r.ko.Spec.HealthCheckConfig.AlarmIdentifier.Name != nil {
				f0f0.SetName(*r.ko.Spec.HealthCheckConfig.AlarmIdentifier.Name)
			}
			if r.ko.Spec.HealthCheckConfig.AlarmIdentifier.Region != nil {
				f0f0.SetRegion(*r.ko.Spec.HealthCheckConfig.AlarmIdentifier.Region)
			}
			f0.SetAlarmIdentifier(f0f0)
		}
		if r.ko.Spec.HealthCheckConfig.ChildHealthChecks != nil {
			f0f1 := []*string{}
			for _, f0f1iter := range r.ko.Spec.HealthCheckConfig.ChildHealthChecks {
				var f0f1elem string
				f0f1elem = *f0f1iter
				f0f1 = append(f0f1, &f0f1elem)
			}
			f0.SetChildHealthChecks(f0f1)
		}
		if r.ko.Spec.HealthCheckConfig.Disabled != nil {
			f0.SetDisabled(*r.ko.Spec.HealthCheckConfig.Disabled)
		}
		if r.ko.Spec.HealthCheckConfig.EnableSNI != nil {
			f0.SetEnableSNI(*r.ko.Spec.HealthCheckConfig.EnableSNI)
		}
		if r.ko.Spec.HealthCheckConfig.FailureThreshold != nil {
			f0.SetFailureThreshold(*r.ko.Spec.HealthCheckConfig.FailureThreshold)
		}
		if r.ko.Spec.HealthCheckConfig.FullyQualifiedDomainName != nil {
			f0.SetFullyQualifiedDomainName(*r.ko.Spec.HealthCheckConfig.FullyQualifiedDomainName)
		}
		if r.ko.Spec.HealthCheckConfig.HealthThreshold != nil {
			f0.SetHealthThreshold(*r.ko.Spec.HealthCheckConfig.HealthThreshold)
		}
		if r.ko.Spec.HealthCheckConfig.IPAddress != nil {
			f0.SetIPAddress(*r.ko.Spec.HealthCheckConfig.IPAddress)
		}
		if r.ko.Spec.HealthCheckConfig.InsufficientDataHealthStatus != nil {
			f0.SetInsufficientDataHealthStatus(*r.ko.Spec.HealthCheckConfig.InsufficientDataHealthStatus)
		}
		if r.ko.Spec.HealthCheckConfig.Inverted != nil {
			f0.SetInverted(*r.ko.Spec.HealthCheckConfig.Inverted)
		}
		if r.ko.Spec.HealthCheckConfig.MeasureLatency != nil {
			f0.SetMeasureLatency(*r.ko.Spec.HealthCheckConfig.MeasureLatency)
		}
		if r.ko.Spec.HealthCheckConfig.Port != nil {
			f0.SetPort(*r.ko.Spec.HealthCheckConfig.Port)
		}
		if r.ko.Spec.HealthCheckConfig.Regions != nil {
			f0f12 := []*string{}
			for _, f0f12iter := range r.ko.Spec.HealthCheckConfig.Regions {
				var f0f12elem string
				f0f12elem = *f0f12iter
				f0f12 = append(f0f12, &f0f12elem)
			}
			f0.SetRegions(f0f12)
		}
		if r.ko.Spec.HealthCheckConfig.RequestInterval != nil {
			f0.SetRequestInterval(*r.ko.Spec.HealthCheckConfig.RequestInterval)
		}
		if r.ko.Spec.HealthCheckConfig.ResourcePath != nil {
			f0.SetResourcePath(*r.ko.Spec.HealthCheckConfig.ResourcePath)
		}
		if r.ko.Spec.HealthCheckConfig.RoutingControlARN != nil {
			f0.SetRoutingControlArn(*r.ko.Spec.HealthCheckConfig.RoutingControlARN)
		}
		if r.ko.Spec.HealthCheckConfig.SearchString != nil {
			f0.SetSearchString(*r.ko.Spec.HealthCheckConfig.SearchString)
		}
		if r.ko.Spec.HealthCheckConfig.Type != nil {
			f0.SetType(*r.ko.Spec.HealthCheckConfig.Type)
		}
		res.SetHealthCheckConfig(f0)
	}

	return res, nil
}

// sdkUpdate patches the supplied resource in the backend AWS service API and
// returns a new resource with updated fields.
func (rm *resourceManager) sdkUpdate(
	ctx context.Context,
	desired *resource,
	latest *resource,
	delta *ackcompare.Delta,
) (*resource, error) {
	return rm.customUpdateHealthCheck(ctx, desired, latest, delta)
}

// sdkDelete deletes the supplied resource in the backend AWS service API
func (rm *resourceManager) sdkDelete(
	ctx context.Context,
	r *resource,
) (latest *resource, err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.sdkDelete")
	defer func() {
		exit(err)
	}()
	input, err := rm.newDeleteRequestPayload(r)
	if err != nil {
		return nil, err
	}
	var resp *svcsdk.DeleteHealthCheckOutput
	_ = resp
	resp, err = rm.sdkapi.DeleteHealthCheckWithContext(ctx, input)
	rm.metrics.RecordAPICall("DELETE", "DeleteHealthCheck", err)
	return nil, err
}

// newDeleteRequestPayload returns an SDK-specific struct for the HTTP request
// payload of the Delete API call for the resource
func (rm *resourceManager) newDeleteRequestPayload(
	r *resource,
) (*svcsdk.DeleteHealthCheckInput, error) {
	res := &svcsdk.DeleteHealthCheckInput{}

	return res, nil
}

// setStatusDefaults sets default properties into supplied custom resource
func (rm *resourceManager) setStatusDefaults(
	ko *svcapitypes.HealthCheck,
) {
	if ko.Status.ACKResourceMetadata == nil {
		ko.Status.ACKResourceMetadata = &ackv1alpha1.ResourceMetadata{}
	}
	if ko.Status.ACKResourceMetadata.Region == nil {
		ko.Status.ACKResourceMetadata.Region = &rm.awsRegion
	}
	if ko.Status.ACKResourceMetadata.OwnerAccountID == nil {
		ko.Status.ACKResourceMetadata.OwnerAccountID = &rm.awsAccountID
	}
	if ko.Status.Conditions == nil {
		ko.Status.Conditions = []*ackv1alpha1.Condition{}
	}
}

// updateConditions returns updated resource, true; if conditions were updated
// else it returns nil, false
func (rm *resourceManager) updateConditions(
	r *resource,
	onSuccess bool,
	err error,
) (*resource, bool) {
	ko := r.ko.DeepCopy()
	rm.setStatusDefaults(ko)

	// Terminal condition
	var terminalCondition *ackv1alpha1.Condition = nil
	var recoverableCondition *ackv1alpha1.Condition = nil
	var syncCondition *ackv1alpha1.Condition = nil
	for _, condition := range ko.Status.Conditions {
		if condition.Type == ackv1alpha1.ConditionTypeTerminal {
			terminalCondition = condition
		}
		if condition.Type == ackv1alpha1.ConditionTypeRecoverable {
			recoverableCondition = condition
		}
		if condition.Type == ackv1alpha1.ConditionTypeResourceSynced {
			syncCondition = condition
		}
	}
	var termError *ackerr.TerminalError
	if rm.terminalAWSError(err) || err == ackerr.SecretTypeNotSupported || err == ackerr.SecretNotFound || errors.As(err, &termError) {
		if terminalCondition == nil {
			terminalCondition = &ackv1alpha1.Condition{
				Type: ackv1alpha1.ConditionTypeTerminal,
			}
			ko.Status.Conditions = append(ko.Status.Conditions, terminalCondition)
		}
		var errorMessage = ""
		if err == ackerr.SecretTypeNotSupported || err == ackerr.SecretNotFound || errors.As(err, &termError) {
			errorMessage = err.Error()
		} else {
			awsErr, _ := ackerr.AWSError(err)
			errorMessage = awsErr.Error()
		}
		terminalCondition.Status = corev1.ConditionTrue
		terminalCondition.Message = &errorMessage
	} else {
		// Clear the terminal condition if no longer present
		if terminalCondition != nil {
			terminalCondition.Status = corev1.ConditionFalse
			terminalCondition.Message = nil
		}
		// Handling Recoverable Conditions
		if err != nil {
			if recoverableCondition == nil {
				// Add a new Condition containing a non-terminal error
				recoverableCondition = &ackv1alpha1.Condition{
					Type: ackv1alpha1.ConditionTypeRecoverable,
				}
				ko.Status.Conditions = append(ko.Status.Conditions, recoverableCondition)
			}
			recoverableCondition.Status = corev1.ConditionTrue
			awsErr, _ := ackerr.AWSError(err)
			errorMessage := err.Error()
			if awsErr != nil {
				errorMessage = awsErr.Error()
			}
			recoverableCondition.Message = &errorMessage
		} else if recoverableCondition != nil {
			recoverableCondition.Status = corev1.ConditionFalse
			recoverableCondition.Message = nil
		}
	}
	// Required to avoid the "declared but not used" error in the default case
	_ = syncCondition
	if terminalCondition != nil || recoverableCondition != nil || syncCondition != nil {
		return &resource{ko}, true // updated
	}
	return nil, false // not updated
}

// terminalAWSError returns awserr, true; if the supplied error is an aws Error type
// and if the exception indicates that it is a Terminal exception
// 'Terminal' exception are specified in generator configuration
func (rm *resourceManager) terminalAWSError(err error) bool {
	if err == nil {
		return false
	}
	awsErr, ok := ackerr.AWSError(err)
	if !ok {
		return false
	}
	switch awsErr.Code() {
	case "TooManyHealthChecks",
		"HealthCheckAlreadyExists",
		"InvalidInput",
		"HealthCheckInUse":
		return true
	default:
		return false
	}
}
