package record_set

import (
	"context"
	"fmt"
	"strings"

	svcapitypes "github.com/aws-controllers-k8s/route53-controller/apis/v1alpha1"
	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	ackerr "github.com/aws-controllers-k8s/runtime/pkg/errors"
	ackrtlog "github.com/aws-controllers-k8s/runtime/pkg/runtime/log"
	svcsdk "github.com/aws/aws-sdk-go/service/route53"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// newResourceRecords returns a slice of ResourceRecord pointer objects
// with values set by the resource's corresponding spec field.
func (rm *resourceManager) newResourceRecords(
	r *resource,
) []*svcsdk.ResourceRecord {
	if r.ko.Spec.ResourceRecords == nil {
		return nil
	}

	res := make([]*svcsdk.ResourceRecord, len(r.ko.Spec.ResourceRecords))
	for i, rr := range r.ko.Spec.ResourceRecords {
		value := *rr.Value
		res[i] = &svcsdk.ResourceRecord{
			Value: &value,
		}
	}
	return res
}

// newAliasTarget returns a pointer to an AliasTarget object
// with values set by the resource's corresponding spec field.
func (rm *resourceManager) newAliasTarget(
	r *resource,
) *svcsdk.AliasTarget {
	if r.ko.Spec.AliasTarget == nil {
		return nil
	}

	res := &svcsdk.AliasTarget{}
	if r.ko.Spec.AliasTarget.DNSName != nil {
		res.SetDNSName(*r.ko.Spec.AliasTarget.DNSName)
	}
	if r.ko.Spec.AliasTarget.EvaluateTargetHealth != nil {
		res.SetEvaluateTargetHealth(*r.ko.Spec.AliasTarget.EvaluateTargetHealth)
	}
	if r.ko.Spec.AliasTarget.HostedZoneID != nil {
		res.SetHostedZoneId(*r.ko.Spec.AliasTarget.HostedZoneID)
	}
	return res
}

// newCIDRRoutingConfig returns a pointer to a CIDRRoutingConfig object
// with values set by the resource's corresponding spec field.
func (rm *resourceManager) newCIDRRoutingConfig(
	r *resource,
) *svcsdk.CidrRoutingConfig {
	if r.ko.Spec.CIDRRoutingConfig == nil {
		return nil
	}

	res := &svcsdk.CidrRoutingConfig{}
	if r.ko.Spec.CIDRRoutingConfig.CollectionID != nil {
		res.SetCollectionId(*r.ko.Spec.CIDRRoutingConfig.CollectionID)
	}
	if r.ko.Spec.CIDRRoutingConfig.LocationName != nil {
		res.SetLocationName(*r.ko.Spec.CIDRRoutingConfig.LocationName)
	}
	return res
}

// newGeoLocation returns a pointer to a GeoLocation object
// with values set by the resource's corresponding spec field.
func (rm *resourceManager) newGeoLocation(
	r *resource,
) *svcsdk.GeoLocation {
	if r.ko.Spec.GeoLocation == nil {
		return nil
	}

	res := &svcsdk.GeoLocation{}
	if r.ko.Spec.GeoLocation.ContinentCode != nil {
		res.SetContinentCode(*r.ko.Spec.GeoLocation.ContinentCode)
	}
	if r.ko.Spec.GeoLocation.CountryCode != nil {
		res.SetCountryCode(*r.ko.Spec.GeoLocation.CountryCode)
	}
	if r.ko.Spec.GeoLocation.SubdivisionCode != nil {
		res.SetSubdivisionCode(*r.ko.Spec.GeoLocation.SubdivisionCode)
	}
	return res
}

// newResourceRecordSet returns a pointer to a ResourceRecordSet object
// with each field set by the resource's corresponding spec field.
func (rm *resourceManager) newResourceRecordSet(
	r *resource,
) *svcsdk.ResourceRecordSet {
	res := &svcsdk.ResourceRecordSet{}

	// Set required fields
	res.SetName(*r.ko.Spec.Name)
	res.SetType(*r.ko.Spec.RecordType)

	// Set optional fields
	if r.ko.Spec.Failover != nil {
		res.SetFailover(*r.ko.Spec.Failover)
	}
	if r.ko.Spec.HealthCheckID != nil {
		res.SetHealthCheckId(*r.ko.Spec.HealthCheckID)
	}
	if r.ko.Spec.MultiValueAnswer != nil {
		res.SetMultiValueAnswer(*r.ko.Spec.MultiValueAnswer)
	}
	if r.ko.Spec.Region != nil {
		res.SetRegion(*r.ko.Spec.Region)
	}
	if r.ko.Spec.SetIdentifier != nil {
		res.SetSetIdentifier(*r.ko.Spec.SetIdentifier)
	}
	if r.ko.Spec.TTL != nil {
		res.SetTTL(*r.ko.Spec.TTL)
	}
	if r.ko.Spec.Weight != nil {
		res.SetWeight(*r.ko.Spec.Weight)
	}

	// Set resource records if available
	resourceRecords := rm.newResourceRecords(r)
	res.SetResourceRecords(resourceRecords)

	// Set alias target if available
	aliasTarget := rm.newAliasTarget(r)
	res.SetAliasTarget(aliasTarget)

	// Set CIDR routing config if available
	cidrRoutingConfig := rm.newCIDRRoutingConfig(r)
	res.SetCidrRoutingConfig(cidrRoutingConfig)

	// Set geolocation if available
	geoLocation := rm.newGeoLocation(r)
	res.SetGeoLocation(geoLocation)

	return res
}

// newChangeBatch returns a pointer to a ChangeBatch object
// with each field set by the resource's corresponding spec field.
func (rm *resourceManager) newChangeBatch(
	action string,
	recordSet *svcsdk.ResourceRecordSet,
) *svcsdk.ChangeBatch {
	change := &svcsdk.Change{}
	change.SetAction(action)
	change.SetResourceRecordSet(recordSet)
	return &svcsdk.ChangeBatch{
		Changes: []*svcsdk.Change{change},
	}
}

// customUpdateRecordSet is the custom implementation for
// RecordSet resource's update operation.
func (rm *resourceManager) customUpdateRecordSet(
	ctx context.Context,
	desired *resource,
	latest *resource,
	delta *ackcompare.Delta,
) (updated *resource, err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.customUpdateRecordSet")
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

	input := &svcsdk.ChangeResourceRecordSetsInput{}
	input.SetHostedZoneId(*desired.ko.Spec.HostedZoneID)

	action := svcsdk.ChangeActionUpsert
	recordSet := rm.newResourceRecordSet(desired)
	changeBatch := rm.newChangeBatch(action, recordSet)
	input.SetChangeBatch(changeBatch)

	var resp *svcsdk.ChangeResourceRecordSetsOutput
	_ = resp
	resp, err = rm.sdkapi.ChangeResourceRecordSetsWithContext(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "ChangeResourceRecordSets", err)

	// Reset previous PENDING or INSYNC status to represent that the
	// updated record set is now in an invalid state
	if err != nil {
		ko.Status.ID = nil
		ko.Status.Status = nil
		ko.Status.SubmittedAt = nil
		return &resource{ko}, err
	}

	if resp.ChangeInfo.Id != nil {
		ko.Status.ID = resp.ChangeInfo.Id
	} else {
		ko.Status.ID = nil
	}
	if resp.ChangeInfo.Status != nil {
		ko.Status.Status = resp.ChangeInfo.Status
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

// syncStatus will sync the state of record sets. PENDING indicates that the
// request has not yet been applied to all Route53 DNS servers and INSYNC
// represents that the request has been fully propagated to all DNS servers.
func (rm *resourceManager) syncStatus(
	ctx context.Context,
	ko *svcapitypes.RecordSet,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.syncStatus")
	defer func() {
		exit(err)
	}()

	changeInput := &svcsdk.GetChangeInput{}
	if ko.Status.ID != nil {
		changeInput.SetId(*ko.Status.ID)
	}
	resp, err := rm.sdkapi.GetChangeWithContext(ctx, changeInput)
	rm.metrics.RecordAPICall("GET", "GetChange", err)
	if err != nil {
		return err
	}

	status := *resp.ChangeInfo.Status
	ko.Status.Status = &status
	return nil
}
