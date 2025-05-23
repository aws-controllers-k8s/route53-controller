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

package record_set

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
	"github.com/aws/aws-sdk-go-v2/aws"
	svcsdk "github.com/aws/aws-sdk-go-v2/service/route53"
	svcsdktypes "github.com/aws/aws-sdk-go-v2/service/route53/types"
	smithy "github.com/aws/smithy-go"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	svcapitypes "github.com/aws-controllers-k8s/route53-controller/apis/v1alpha1"
)

// Hack to avoid import errors during build...
var (
	_ = &metav1.Time{}
	_ = strings.ToLower("")
	_ = &svcsdk.Client{}
	_ = &svcapitypes.RecordSet{}
	_ = ackv1alpha1.AWSAccountID("")
	_ = &ackerr.NotFound
	_ = &ackcondition.NotManagedMessage
	_ = &reflect.Value{}
	_ = fmt.Sprintf("")
	_ = &ackrequeue.NoRequeue{}
	_ = &aws.Config{}
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
	if rm.requiredFieldsMissingFromReadManyInput(r) {
		return nil, ackerr.NotFound
	}

	input, err := rm.newListRequestPayload(r)
	if err != nil {
		return nil, err
	}

	// Retrieve the domain name of the hosted zone through the ID.
	domain, err := rm.getHostedZoneDomain(ctx, r)
	if err != nil {
		return nil, err
	}

	// Return the combined value of the user specified subdomain and the hosted zone domain.
	dnsName := rm.getDNSName(r, domain)

	// Setting the starting point to the following values reduces the number of irrelevant
	// records that are returned.
	input.StartRecordName = &dnsName
	if r.ko.Spec.RecordType != nil {
		input.StartRecordType = svcsdktypes.RRType(*r.ko.Spec.RecordType)
	}
	if r.ko.Spec.SetIdentifier != nil {
		input.StartRecordIdentifier = r.ko.Spec.SetIdentifier
	}

	var resp *svcsdk.ListResourceRecordSetsOutput
	resp, err = rm.sdkapi.ListResourceRecordSets(ctx, input)
	rm.metrics.RecordAPICall("READ_MANY", "ListResourceRecordSets", err)
	if err != nil {
		var awsErr smithy.APIError
		if errors.As(err, &awsErr) && awsErr.ErrorCode() == "NoSuchHostedZone" {
			return nil, ackerr.NotFound
		}
		return nil, err
	}

	// Merge in the information we read from the API call above to the copy of
	// the original Kubernetes object we passed to the function
	ko := r.ko.DeepCopy()

	// ListResourceRecordSets does not result in an exact match of relevant records as
	// it just consumes starting values for HostedZoneID, Name, RecordType, and SetIdentifier
	// from an alphabetically sorted list. As an example, if we are filtering for 'A' records,
	// ListResourceRecordSets could still return 'CNAME' records.
	var recordSets []svcsdktypes.ResourceRecordSet
	for _, elem := range resp.ResourceRecordSets {
		if elem.Name != nil {
			// ListResourceRecordSets returns the full DNS name, so we need to reconstruct
			// the output to compare with the user specified subdomain. If a '*' value is
			// in the subdomain, ListResourceRecordSets returns it as an encoded value, so
			// this needs to be decoded before our comparison.
			subdomain := strings.TrimSuffix(*elem.Name, domain)
			subdomain = decodeRecordName(subdomain)

			// If user supplied no subdomain, we know that records with subdomains cannot
			// be a match and vice versa.
			if (r.ko.Spec.Name == nil && subdomain != "") || (r.ko.Spec.Name != nil && subdomain == "") {
				continue
			}

			// For cases where the user supplied a value to Spec.Name, irrelevant records
			// from ListResourceRecordSets will be further filtered out at a later point in
			// sdkFind. For now, parse out the "." at the end of the returned subdomain.
			if subdomain != "" {
				subdomain = subdomain[:len(subdomain)-1]
				elem.Name = &subdomain
			} else {
				elem.Name = nil
			}
		}

		// Similar to above, remove the "." at the end and decode the "*" value as necessary.
		if elem.AliasTarget != nil && ko.Spec.AliasTarget != nil {
			if elem.AliasTarget.DNSName != nil && ko.Spec.AliasTarget.DNSName != nil {
				dnsName = *elem.AliasTarget.DNSName
				decodedName := decodeRecordName(dnsName[:len(dnsName)-1])
				elem.AliasTarget.DNSName = &decodedName
			}
		}

		// RecordTypes are required, so discard records that don't have them.
		if elem.Type == "" || (string(elem.Type) != *ko.Spec.RecordType) {
			continue
		}

		recordSets = append(recordSets, elem)
	}
	resp.ResourceRecordSets = recordSets

	found := false
	for _, elem := range resp.ResourceRecordSets {
		if elem.AliasTarget != nil {
			f0 := &svcapitypes.AliasTarget{}
			if elem.AliasTarget.DNSName != nil {
				f0.DNSName = elem.AliasTarget.DNSName
			}
			f0.EvaluateTargetHealth = &elem.AliasTarget.EvaluateTargetHealth
			if elem.AliasTarget.HostedZoneId != nil {
				f0.HostedZoneID = elem.AliasTarget.HostedZoneId
			}
			ko.Spec.AliasTarget = f0
		} else {
			ko.Spec.AliasTarget = nil
		}
		if elem.CidrRoutingConfig != nil {
			f1 := &svcapitypes.CIDRRoutingConfig{}
			if elem.CidrRoutingConfig.CollectionId != nil {
				f1.CollectionID = elem.CidrRoutingConfig.CollectionId
			}
			if elem.CidrRoutingConfig.LocationName != nil {
				f1.LocationName = elem.CidrRoutingConfig.LocationName
			}
			ko.Spec.CIDRRoutingConfig = f1
		} else {
			ko.Spec.CIDRRoutingConfig = nil
		}
		if elem.Failover != "" {
			ko.Spec.Failover = aws.String(string(elem.Failover))
		} else {
			ko.Spec.Failover = nil
		}
		if elem.GeoLocation != nil {
			f3 := &svcapitypes.GeoLocation{}
			if elem.GeoLocation.ContinentCode != nil {
				f3.ContinentCode = elem.GeoLocation.ContinentCode
			}
			if elem.GeoLocation.CountryCode != nil {
				f3.CountryCode = elem.GeoLocation.CountryCode
			}
			if elem.GeoLocation.SubdivisionCode != nil {
				f3.SubdivisionCode = elem.GeoLocation.SubdivisionCode
			}
			ko.Spec.GeoLocation = f3
		} else {
			ko.Spec.GeoLocation = nil
		}
		if elem.HealthCheckId != nil {
			ko.Spec.HealthCheckID = elem.HealthCheckId
		} else {
			ko.Spec.HealthCheckID = nil
		}
		if elem.MultiValueAnswer != nil {
			ko.Spec.MultiValueAnswer = elem.MultiValueAnswer
		} else {
			ko.Spec.MultiValueAnswer = nil
		}
		if elem.Name != nil {
			if ko.Spec.Name != nil {
				if *elem.Name != *ko.Spec.Name {
					continue
				}
			}
			ko.Spec.Name = elem.Name
		} else {
			ko.Spec.Name = nil
		}
		if elem.Region != "" {
			ko.Spec.Region = aws.String(string(elem.Region))
		} else {
			ko.Spec.Region = nil
		}
		if elem.ResourceRecords != nil {
			f8 := []*svcapitypes.ResourceRecord{}
			for _, f8iter := range elem.ResourceRecords {
				f8elem := &svcapitypes.ResourceRecord{}
				if f8iter.Value != nil {
					f8elem.Value = f8iter.Value
				}
				f8 = append(f8, f8elem)
			}
			ko.Spec.ResourceRecords = f8
		} else {
			ko.Spec.ResourceRecords = nil
		}
		if elem.SetIdentifier != nil {
			if ko.Spec.SetIdentifier != nil {
				if *elem.SetIdentifier != *ko.Spec.SetIdentifier {
					continue
				}
			}
			ko.Spec.SetIdentifier = elem.SetIdentifier
		} else {
			ko.Spec.SetIdentifier = nil
		}
		if elem.TTL != nil {
			ko.Spec.TTL = elem.TTL
		} else {
			ko.Spec.TTL = nil
		}
		if elem.Weight != nil {
			ko.Spec.Weight = elem.Weight
		} else {
			ko.Spec.Weight = nil
		}
		found = true
		break
	}
	if !found {
		return nil, ackerr.NotFound
	}

	rm.setStatusDefaults(ko)

	// Status represents whether record changes have been fully propagated to all
	// Route 53 authoritative DNS servers. The current status for the propagation
	// should be updated if it's not already INSYNC.
	err = rm.syncStatus(ctx, ko)
	if err != nil {
		return nil, err
	}
	if ko.Status.Status == nil || svcsdktypes.ChangeStatus(*ko.Status.Status) != svcsdktypes.ChangeStatusInsync {
		ackcondition.SetSynced(&resource{ko}, corev1.ConditionFalse, nil, nil)
	}

	return &resource{ko}, nil
}

// requiredFieldsMissingFromReadManyInput returns true if there are any fields
// for the ReadMany Input shape that are required but not present in the
// resource's Spec or Status
func (rm *resourceManager) requiredFieldsMissingFromReadManyInput(
	r *resource,
) bool {
	return false
}

// newListRequestPayload returns SDK-specific struct for the HTTP request
// payload of the List API call for the resource
func (rm *resourceManager) newListRequestPayload(
	r *resource,
) (*svcsdk.ListResourceRecordSetsInput, error) {
	res := &svcsdk.ListResourceRecordSetsInput{}

	if r.ko.Spec.HostedZoneID != nil {
		res.HostedZoneId = r.ko.Spec.HostedZoneID
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

	action := svcsdktypes.ChangeActionCreate
	recordSet, err := rm.newResourceRecordSet(ctx, desired)
	if err != nil {
		return nil, err
	}
	changeBatch := rm.newChangeBatch(action, recordSet)
	input.ChangeBatch = changeBatch

	var resp *svcsdk.ChangeResourceRecordSetsOutput
	_ = resp
	resp, err = rm.sdkapi.ChangeResourceRecordSets(ctx, input)
	rm.metrics.RecordAPICall("CREATE", "ChangeResourceRecordSets", err)
	if err != nil {
		return nil, err
	}
	// Merge in the information we read from the API call above to the copy of
	// the original Kubernetes object we passed to the function
	ko := desired.ko.DeepCopy()

	if resp.ChangeInfo.Id != nil {
		ko.Status.ID = resp.ChangeInfo.Id
	} else {
		ko.Status.ID = nil
	}
	if resp.ChangeInfo.Status != "" {
		ko.Status.Status = aws.String(string(resp.ChangeInfo.Status))
	} else {
		ko.Status.Status = nil
	}
	if resp.ChangeInfo.SubmittedAt != nil {
		ko.Status.SubmittedAt = &metav1.Time{*resp.ChangeInfo.SubmittedAt}
	} else {
		ko.Status.SubmittedAt = nil
	}

	rm.setStatusDefaults(ko)
	return &resource{ko}, nil
}

// newCreateRequestPayload returns an SDK-specific struct for the HTTP request
// payload of the Create API call for the resource
func (rm *resourceManager) newCreateRequestPayload(
	ctx context.Context,
	r *resource,
) (*svcsdk.ChangeResourceRecordSetsInput, error) {
	res := &svcsdk.ChangeResourceRecordSetsInput{}

	if r.ko.Spec.HostedZoneID != nil {
		res.HostedZoneId = r.ko.Spec.HostedZoneID
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
	return rm.customUpdateRecordSet(ctx, desired, latest, delta)
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

	action := svcsdktypes.ChangeActionDelete
	recordSet, err := rm.newResourceRecordSet(ctx, r)
	if err != nil {
		return nil, err
	}
	changeBatch := rm.newChangeBatch(action, recordSet)
	input.ChangeBatch = changeBatch

	var resp *svcsdk.ChangeResourceRecordSetsOutput
	_ = resp
	resp, err = rm.sdkapi.ChangeResourceRecordSets(ctx, input)
	rm.metrics.RecordAPICall("DELETE", "ChangeResourceRecordSets", err)
	return nil, err
}

// newDeleteRequestPayload returns an SDK-specific struct for the HTTP request
// payload of the Delete API call for the resource
func (rm *resourceManager) newDeleteRequestPayload(
	r *resource,
) (*svcsdk.ChangeResourceRecordSetsInput, error) {
	res := &svcsdk.ChangeResourceRecordSetsInput{}

	if r.ko.Spec.HostedZoneID != nil {
		res.HostedZoneId = r.ko.Spec.HostedZoneID
	}

	return res, nil
}

// setStatusDefaults sets default properties into supplied custom resource
func (rm *resourceManager) setStatusDefaults(
	ko *svcapitypes.RecordSet,
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

	var terminalErr smithy.APIError
	if !errors.As(err, &terminalErr) {
		return false
	}
	switch terminalErr.ErrorCode() {
	case "InvalidChangeBatch",
		"InvalidInput",
		"NoSuchHostedZone",
		"NoSuchHealthCheck":
		return true
	default:
		return false
	}
}
