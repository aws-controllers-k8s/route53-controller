package record_set

import (
	"context"
	"errors"
	"strings"

	svcapitypes "github.com/aws-controllers-k8s/route53-controller/apis/v1alpha1"
	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	ackcondition "github.com/aws-controllers-k8s/runtime/pkg/condition"
	ackerr "github.com/aws-controllers-k8s/runtime/pkg/errors"
	ackrtlog "github.com/aws-controllers-k8s/runtime/pkg/runtime/log"
	"github.com/aws/aws-sdk-go-v2/aws"
	svcsdk "github.com/aws/aws-sdk-go-v2/service/route53"
	svcsdktypes "github.com/aws/aws-sdk-go-v2/service/route53/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// newResourceRecords returns a slice of ResourceRecord pointer objects
// with values set by the resource's corresponding spec field.
func (rm *resourceManager) newResourceRecords(
	r *resource,
) []svcsdktypes.ResourceRecord {
	if r.ko.Spec.ResourceRecords == nil {
		return nil
	}

	res := make([]svcsdktypes.ResourceRecord, len(r.ko.Spec.ResourceRecords))
	for i, rr := range r.ko.Spec.ResourceRecords {
		value := *rr.Value
		res[i] = svcsdktypes.ResourceRecord{
			Value: &value,
		}
	}
	return res
}

// newAliasTarget returns a pointer to an AliasTarget object
// with values set by the resource's corresponding spec field.
func (rm *resourceManager) newAliasTarget(
	r *resource,
) *svcsdktypes.AliasTarget {
	if r.ko.Spec.AliasTarget == nil {
		return nil
	}

	res := &svcsdktypes.AliasTarget{}
	if r.ko.Spec.AliasTarget.DNSName != nil {
		res.DNSName = r.ko.Spec.AliasTarget.DNSName
	}
	if r.ko.Spec.AliasTarget.EvaluateTargetHealth != nil {
		res.EvaluateTargetHealth = *r.ko.Spec.AliasTarget.EvaluateTargetHealth
	}
	if r.ko.Spec.AliasTarget.HostedZoneID != nil {
		res.HostedZoneId = r.ko.Spec.AliasTarget.HostedZoneID
	}
	return res
}

// newCIDRRoutingConfig returns a pointer to a CIDRRoutingConfig object
// with values set by the resource's corresponding spec field.
func (rm *resourceManager) newCIDRRoutingConfig(
	r *resource,
) *svcsdktypes.CidrRoutingConfig {
	if r.ko.Spec.CIDRRoutingConfig == nil {
		return nil
	}

	res := &svcsdktypes.CidrRoutingConfig{}
	if r.ko.Spec.CIDRRoutingConfig.CollectionID != nil {
		res.CollectionId = r.ko.Spec.CIDRRoutingConfig.CollectionID
	}
	if r.ko.Spec.CIDRRoutingConfig.LocationName != nil {
		res.LocationName = r.ko.Spec.CIDRRoutingConfig.LocationName
	}
	return res
}

// newGeoLocation returns a pointer to a GeoLocation object
// with values set by the resource's corresponding spec field.
func (rm *resourceManager) newGeoLocation(
	r *resource,
) *svcsdktypes.GeoLocation {
	if r.ko.Spec.GeoLocation == nil {
		return nil
	}

	res := &svcsdktypes.GeoLocation{}
	if r.ko.Spec.GeoLocation.ContinentCode != nil {
		res.ContinentCode = r.ko.Spec.GeoLocation.ContinentCode
	}
	if r.ko.Spec.GeoLocation.CountryCode != nil {
		res.CountryCode = r.ko.Spec.GeoLocation.CountryCode
	}
	if r.ko.Spec.GeoLocation.SubdivisionCode != nil {
		res.SubdivisionCode = r.ko.Spec.GeoLocation.SubdivisionCode
	}
	return res
}

// newResourceRecordSet returns a pointer to a ResourceRecordSet object
// with each field set by the resource's corresponding spec field.
func (rm *resourceManager) newResourceRecordSet(
	ctx context.Context,
	r *resource,
) (*svcsdktypes.ResourceRecordSet, error) {
	res := &svcsdktypes.ResourceRecordSet{}

	domain, err := rm.getHostedZoneDomain(ctx, r)
	if err != nil {
		return nil, err
	}
	dnsName := rm.getDNSName(r, domain)

	// Set required fields for the ChangeResourceRecordSets API
	res.Name = &dnsName
	res.Type = svcsdktypes.RRType(*r.ko.Spec.RecordType)

	// Set optional fields
	if r.ko.Spec.Failover != nil {
		res.Failover = svcsdktypes.ResourceRecordSetFailover(*r.ko.Spec.Failover)
	}
	if r.ko.Spec.HealthCheckID != nil {
		res.HealthCheckId = r.ko.Spec.HealthCheckID
	}
	if r.ko.Spec.MultiValueAnswer != nil {
		res.MultiValueAnswer = r.ko.Spec.MultiValueAnswer
	}
	if r.ko.Spec.Region != nil {
		res.Region = svcsdktypes.ResourceRecordSetRegion(*r.ko.Spec.Region)
	}
	if r.ko.Spec.SetIdentifier != nil {
		res.SetIdentifier = r.ko.Spec.SetIdentifier
	}
	if r.ko.Spec.TTL != nil {
		res.TTL = r.ko.Spec.TTL
	}
	if r.ko.Spec.Weight != nil {
		res.Weight = r.ko.Spec.Weight
	}

	// Set resource records if available
	resourceRecords := rm.newResourceRecords(r)
	res.ResourceRecords = resourceRecords

	// Set alias target if available
	aliasTarget := rm.newAliasTarget(r)
	res.AliasTarget = aliasTarget

	// Set CIDR routing config if available
	cidrRoutingConfig := rm.newCIDRRoutingConfig(r)
	res.CidrRoutingConfig = cidrRoutingConfig

	// Set geolocation if available
	geoLocation := rm.newGeoLocation(r)
	res.GeoLocation = geoLocation

	return res, nil
}

// newChangeBatch returns a pointer to a ChangeBatch object
// with each field set by the resource's corresponding spec field.
func (rm *resourceManager) newChangeBatch(
	action svcsdktypes.ChangeAction,
	recordSet *svcsdktypes.ResourceRecordSet,
) *svcsdktypes.ChangeBatch {
	change := svcsdktypes.Change{}
	change.Action = svcsdktypes.ChangeAction(action)
	change.ResourceRecordSet = recordSet
	return &svcsdktypes.ChangeBatch{
		Changes: []svcsdktypes.Change{change},
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

	// Merge in the information we read from the API call above to the copy of
	// the original Kubernetes object we passed to the function
	ko := desired.ko.DeepCopy()

	input := &svcsdk.ChangeResourceRecordSetsInput{}
	input.HostedZoneId = desired.ko.Spec.HostedZoneID

	action := svcsdktypes.ChangeActionUpsert
	recordSet, err := rm.newResourceRecordSet(ctx, desired)
	if err != nil {
		return nil, err
	}
	changeBatch := rm.newChangeBatch(action, recordSet)
	input.ChangeBatch = changeBatch

	var resp *svcsdk.ChangeResourceRecordSetsOutput
	resp, err = rm.sdkapi.ChangeResourceRecordSets(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "ChangeResourceRecordSets", err)

	// The previous change batch is no longer representative of the newly applied change.
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

	// Ensure that the status eventually becomes INSYNC after an update has been detected
	err = rm.syncStatus(ctx, ko)
	if err != nil {
		return nil, err
	}
	if ko.Status.Status == nil || svcsdktypes.ChangeStatus(*ko.Status.Status) == svcsdktypes.ChangeStatusPending {
		ackcondition.SetSynced(&resource{ko}, corev1.ConditionFalse, nil, nil)
	}

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

	// It is possible to hit this condition if the previous change batch was
	// invalid (e.g. bad parameter). In such cases, a new change ID will be
	// assigned after going through a successful update.
	if ko.Status.ID == nil {
		ko.Status.Status = nil
		return nil
	}

	changeInput := &svcsdk.GetChangeInput{}
	changeInput.Id = ko.Status.ID

	resp, err := rm.sdkapi.GetChange(ctx, changeInput)
	rm.metrics.RecordAPICall("READ_ONE", "GetChange", err)
	if err != nil {
		return err
	}

	status := string(resp.ChangeInfo.Status)
	ko.Status.Status = &status
	return nil
}

// getHostedZoneDomain gets the domain name of the hosted zone.
func (rm *resourceManager) getHostedZoneDomain(
	ctx context.Context,
	r *resource,
) (string, error) {
	var err error

	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.getHostedZoneDomain")
	defer func() {
		exit(err)
	}()

	input := &svcsdk.GetHostedZoneInput{}
	if r.ko.Spec.HostedZoneID != nil {
		input.Id = r.ko.Spec.HostedZoneID
	}

	resp, err := rm.sdkapi.GetHostedZone(ctx, input)
	rm.metrics.RecordAPICall("READ_ONE", "GetHostedZone", err)
	if err != nil {
		var notFound *svcsdktypes.NoSuchHostedZone
		if errors.As(err, &notFound) {
			return "", ackerr.NotFound
		}
		return "", err
	}
	return *resp.HostedZone.Name, nil
}

// getDNSName returns the appended value of the user supplied subdomain and the
// domain of the hosted zone. If a subdomain is not supplied, the full DNS name
// will just equate to the hosted zone domain name.
func (rm *resourceManager) getDNSName(
	r *resource,
	domain string,
) (dnsName string) {
	if r.ko.Spec.Name != nil {
		dnsName = *r.ko.Spec.Name
	}

	if strings.HasSuffix(dnsName, ".") {
		return dnsName
	} else if len(dnsName) > 0 {
		dnsName += "."
	}
	dnsName += domain
	return dnsName
}

// decodeRecordName decodes special characters from the DNSName of a record set.
// ListResourceRecordSets returns the DNS names with an encoded value for "*",
// so the DNSName needs to be decoded before comparing with our spec values.
func decodeRecordName(name string) string {
	if strings.Contains(name, "\\052") {
		return strings.Replace(name, "\\052", "*", -1)
	}
	return name
}
