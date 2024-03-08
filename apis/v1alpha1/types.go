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

package v1alpha1

import (
	ackv1alpha1 "github.com/aws-controllers-k8s/runtime/apis/core/v1alpha1"
	"github.com/aws/aws-sdk-go/aws"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Hack to avoid import errors during build...
var (
	_ = &metav1.Time{}
	_ = &aws.JSONValue{}
	_ = ackv1alpha1.AWSAccountID("")
)

// A complex type that identifies the CloudWatch alarm that you want Amazon
// Route 53 health checkers to use to determine whether the specified health
// check is healthy.
type AlarmIdentifier struct {
	Name   *string `json:"name,omitempty"`
	Region *string `json:"region,omitempty"`
}

// Alias resource record sets only: Information about the Amazon Web Services
// resource, such as a CloudFront distribution or an Amazon S3 bucket, that
// you want to route traffic to.
//
// When creating resource record sets for a private hosted zone, note the following:
//
//   - For information about creating failover resource record sets in a private
//     hosted zone, see Configuring Failover in a Private Hosted Zone (https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/dns-failover-private-hosted-zones.html).
type AliasTarget struct {
	DNSName              *string `json:"dnsName,omitempty"`
	EvaluateTargetHealth *bool   `json:"evaluateTargetHealth,omitempty"`
	HostedZoneID         *string `json:"hostedZoneID,omitempty"`
}

// A complex type that identifies a CIDR collection.
type CIDRCollection struct {
	ID *string `json:"id,omitempty"`
}

// The object that is specified in resource record set object when you are linking
// a resource record set to a CIDR location.
//
// A LocationName with an asterisk “*” can be used to create a default CIDR
// record. CollectionId is still required for default record.
type CIDRRoutingConfig struct {
	CollectionID *string `json:"collectionID,omitempty"`
	LocationName *string `json:"locationName,omitempty"`
}

// The information for each resource record set that you want to change.
type Change struct {
	Action *string `json:"action,omitempty"`
	// Information about the resource record set to create or delete.
	ResourceRecordSet *ResourceRecordSet `json:"resourceRecordSet,omitempty"`
}

// The information for a change request.
type ChangeBatch struct {
	Changes []*Change `json:"changes,omitempty"`
	Comment *string   `json:"comment,omitempty"`
}

// A complex type that describes change information about changes made to your
// hosted zone.
type ChangeInfo struct {
	ID          *string      `json:"id,omitempty"`
	Status      *string      `json:"status,omitempty"`
	SubmittedAt *metav1.Time `json:"submittedAt,omitempty"`
}

// A complex type that contains information about the CloudWatch alarm that
// Amazon Route 53 is monitoring for this health check.
type CloudWatchAlarmConfiguration struct {
	ComparisonOperator *string      `json:"comparisonOperator,omitempty"`
	Dimensions         []*Dimension `json:"dimensions,omitempty"`
	EvaluationPeriods  *int64       `json:"evaluationPeriods,omitempty"`
	MetricName         *string      `json:"metricName,omitempty"`
	Namespace          *string      `json:"namespace,omitempty"`
	Period             *int64       `json:"period,omitempty"`
	Statistic          *string      `json:"statistic,omitempty"`
	Threshold          *float64     `json:"threshold,omitempty"`
}

// A complex type that is an entry in an CidrCollection (https://docs.aws.amazon.com/Route53/latest/APIReference/API_CidrCollection.html)
// array.
type CollectionSummary struct {
	ID *string `json:"id,omitempty"`
}

// A complex type that lists the name servers in a delegation set, as well as
// the CallerReference and the ID for the delegation set.
type DelegationSet struct {
	CallerReference *string   `json:"callerReference,omitempty"`
	ID              *string   `json:"id,omitempty"`
	NameServers     []*string `json:"nameServers,omitempty"`
}

// For the metric that the CloudWatch alarm is associated with, a complex type
// that contains information about one dimension.
type Dimension struct {
	Name  *string `json:"name,omitempty"`
	Value *string `json:"value,omitempty"`
}

// A complex type that contains information about a geographic location.
type GeoLocation struct {
	ContinentCode   *string `json:"continentCode,omitempty"`
	CountryCode     *string `json:"countryCode,omitempty"`
	SubdivisionCode *string `json:"subdivisionCode,omitempty"`
}

// A complex type that contains the codes and full continent, country, and subdivision
// names for the specified geolocation code.
type GeoLocationDetails struct {
	ContinentCode   *string `json:"continentCode,omitempty"`
	CountryCode     *string `json:"countryCode,omitempty"`
	SubdivisionCode *string `json:"subdivisionCode,omitempty"`
}

// A complex type that contains information about the health check.
type HealthCheckConfig struct {
	// A complex type that identifies the CloudWatch alarm that you want Amazon
	// Route 53 health checkers to use to determine whether the specified health
	// check is healthy.
	AlarmIdentifier              *AlarmIdentifier `json:"alarmIdentifier,omitempty"`
	ChildHealthChecks            []*string        `json:"childHealthChecks,omitempty"`
	Disabled                     *bool            `json:"disabled,omitempty"`
	EnableSNI                    *bool            `json:"enableSNI,omitempty"`
	FailureThreshold             *int64           `json:"failureThreshold,omitempty"`
	FullyQualifiedDomainName     *string          `json:"fullyQualifiedDomainName,omitempty"`
	HealthThreshold              *int64           `json:"healthThreshold,omitempty"`
	IPAddress                    *string          `json:"iPAddress,omitempty"`
	InsufficientDataHealthStatus *string          `json:"insufficientDataHealthStatus,omitempty"`
	Inverted                     *bool            `json:"inverted,omitempty"`
	MeasureLatency               *bool            `json:"measureLatency,omitempty"`
	Port                         *int64           `json:"port,omitempty"`
	Regions                      []*string        `json:"regions,omitempty"`
	RequestInterval              *int64           `json:"requestInterval,omitempty"`
	ResourcePath                 *string          `json:"resourcePath,omitempty"`
	RoutingControlARN            *string          `json:"routingControlARN,omitempty"`
	SearchString                 *string          `json:"searchString,omitempty"`
	Type                         *string          `json:"type_,omitempty"`
}

// A complex type that contains the last failure reason as reported by one Amazon
// Route 53 health checker.
type HealthCheckObservation struct {
	IPAddress *string `json:"iPAddress,omitempty"`
	Region    *string `json:"region,omitempty"`
}

// A complex type that contains information about one health check that is associated
// with the current Amazon Web Services account.
type HealthCheck_SDK struct {
	CallerReference *string `json:"callerReference,omitempty"`
	// A complex type that contains information about the CloudWatch alarm that
	// Amazon Route 53 is monitoring for this health check.
	CloudWatchAlarmConfiguration *CloudWatchAlarmConfiguration `json:"cloudWatchAlarmConfiguration,omitempty"`
	// A complex type that contains information about the health check.
	HealthCheckConfig  *HealthCheckConfig `json:"healthCheckConfig,omitempty"`
	HealthCheckVersion *int64             `json:"healthCheckVersion,omitempty"`
	ID                 *string            `json:"id,omitempty"`
	// If a health check or hosted zone was created by another service, LinkedService
	// is a complex type that describes the service that created the resource. When
	// a resource is created by another service, you can't edit or delete it using
	// Amazon Route 53.
	LinkedService *LinkedService `json:"linkedService,omitempty"`
}

// A complex type that contains an optional comment about your hosted zone.
// If you don't want to specify a comment, omit both the HostedZoneConfig and
// Comment elements.
type HostedZoneConfig struct {
	Comment     *string `json:"comment,omitempty"`
	PrivateZone *bool   `json:"privateZone,omitempty"`
}

// In the response to a ListHostedZonesByVPC request, the HostedZoneSummaries
// element contains one HostedZoneSummary element for each hosted zone that
// the specified Amazon VPC is associated with. Each HostedZoneSummary element
// contains the hosted zone name and ID, and information about who owns the
// hosted zone.
type HostedZoneSummary struct {
	HostedZoneID *string `json:"hostedZoneID,omitempty"`
	Name         *string `json:"name,omitempty"`
}

// A complex type that contains general information about the hosted zone.
type HostedZone_SDK struct {
	CallerReference *string `json:"callerReference,omitempty"`
	// A complex type that contains an optional comment about your hosted zone.
	// If you don't want to specify a comment, omit both the HostedZoneConfig and
	// Comment elements.
	Config *HostedZoneConfig `json:"config,omitempty"`
	ID     *string           `json:"id,omitempty"`
	// If a health check or hosted zone was created by another service, LinkedService
	// is a complex type that describes the service that created the resource. When
	// a resource is created by another service, you can't edit or delete it using
	// Amazon Route 53.
	LinkedService          *LinkedService `json:"linkedService,omitempty"`
	Name                   *string        `json:"name,omitempty"`
	ResourceRecordSetCount *int64         `json:"resourceRecordSetCount,omitempty"`
}

// A key-signing key (KSK) is a complex type that represents a public/private
// key pair. The private key is used to generate a digital signature for the
// zone signing key (ZSK). The public key is stored in the DNS and is used to
// authenticate the ZSK. A KSK is always associated with a hosted zone; it cannot
// exist by itself.
type KeySigningKey struct {
	CreatedDate      *metav1.Time `json:"createdDate,omitempty"`
	LastModifiedDate *metav1.Time `json:"lastModifiedDate,omitempty"`
}

// If a health check or hosted zone was created by another service, LinkedService
// is a complex type that describes the service that created the resource. When
// a resource is created by another service, you can't edit or delete it using
// Amazon Route 53.
type LinkedService struct {
	Description      *string `json:"description,omitempty"`
	ServicePrincipal *string `json:"servicePrincipal,omitempty"`
}

// A complex type that contains information about the CIDR location.
type LocationSummary struct {
	LocationName *string `json:"locationName,omitempty"`
}

// A complex type that contains information about a configuration for DNS query
// logging.
type QueryLoggingConfig struct {
	HostedZoneID *string `json:"hostedZoneID,omitempty"`
}

// Information specific to the resource record.
//
// If you're creating an alias resource record set, omit ResourceRecord.
type ResourceRecord struct {
	Value *string `json:"value,omitempty"`
}

// Information about the resource record set to create or delete.
type ResourceRecordSet struct {
	// Alias resource record sets only: Information about the Amazon Web Services
	// resource, such as a CloudFront distribution or an Amazon S3 bucket, that
	// you want to route traffic to.
	//
	// When creating resource record sets for a private hosted zone, note the following:
	//
	//    * For information about creating failover resource record sets in a private
	//    hosted zone, see Configuring Failover in a Private Hosted Zone (https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/dns-failover-private-hosted-zones.html).
	AliasTarget *AliasTarget `json:"aliasTarget,omitempty"`
	// The object that is specified in resource record set object when you are linking
	// a resource record set to a CIDR location.
	//
	// A LocationName with an asterisk “*” can be used to create a default CIDR
	// record. CollectionId is still required for default record.
	CIDRRoutingConfig *CIDRRoutingConfig `json:"cidrRoutingConfig,omitempty"`
	Failover          *string            `json:"failover,omitempty"`
	// A complex type that contains information about a geographic location.
	GeoLocation             *GeoLocation      `json:"geoLocation,omitempty"`
	HealthCheckID           *string           `json:"healthCheckID,omitempty"`
	MultiValueAnswer        *bool             `json:"multiValueAnswer,omitempty"`
	Name                    *string           `json:"name,omitempty"`
	Region                  *string           `json:"region,omitempty"`
	ResourceRecords         []*ResourceRecord `json:"resourceRecords,omitempty"`
	SetIdentifier           *string           `json:"setIdentifier,omitempty"`
	TTL                     *int64            `json:"ttl,omitempty"`
	TrafficPolicyInstanceID *string           `json:"trafficPolicyInstanceID,omitempty"`
	Type                    *string           `json:"type_,omitempty"`
	Weight                  *int64            `json:"weight,omitempty"`
}

// A complex type containing a resource and its associated tags.
type ResourceTagSet struct {
	Tags []*Tag `json:"tags,omitempty"`
}

// A complex type that contains the status that one Amazon Route 53 health checker
// reports and the time of the health check.
type StatusReport struct {
	CheckedTime *metav1.Time `json:"checkedTime,omitempty"`
}

// A complex type that contains information about a tag that you want to add
// or edit for the specified health check or hosted zone.
type Tag struct {
	Key   *string `json:"key,omitempty"`
	Value *string `json:"value,omitempty"`
}

// A complex type that contains settings for a traffic policy.
type TrafficPolicy struct {
	Type *string `json:"type_,omitempty"`
}

// A complex type that contains settings for the new traffic policy instance.
type TrafficPolicyInstance struct {
	HostedZoneID      *string `json:"hostedZoneID,omitempty"`
	ID                *string `json:"id,omitempty"`
	Name              *string `json:"name,omitempty"`
	TTL               *int64  `json:"ttl,omitempty"`
	TrafficPolicyType *string `json:"trafficPolicyType,omitempty"`
}

// A complex type that contains information about the latest version of one
// traffic policy that is associated with the current Amazon Web Services account.
type TrafficPolicySummary struct {
	Type *string `json:"type_,omitempty"`
}

// (Private hosted zones only) A complex type that contains information about
// an Amazon VPC.
//
// If you associate a private hosted zone with an Amazon VPC when you make a
// CreateHostedZone (https://docs.aws.amazon.com/Route53/latest/APIReference/API_CreateHostedZone.html)
// request, the following parameters are also required.
type VPC struct {
	// (Private hosted zones only) The ID of an Amazon VPC.
	VPCID     *string `json:"vpcID,omitempty"`
	VPCRegion *string `json:"vpcRegion,omitempty"`
}
