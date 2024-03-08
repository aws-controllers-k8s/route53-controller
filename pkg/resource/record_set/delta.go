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
	"bytes"
	"reflect"

	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	acktags "github.com/aws-controllers-k8s/runtime/pkg/tags"
)

// Hack to avoid import errors during build...
var (
	_ = &bytes.Buffer{}
	_ = &reflect.Method{}
	_ = &acktags.Tags{}
)

// newResourceDelta returns a new `ackcompare.Delta` used to compare two
// resources
func newResourceDelta(
	a *resource,
	b *resource,
) *ackcompare.Delta {
	delta := ackcompare.NewDelta()
	if (a == nil && b != nil) ||
		(a != nil && b == nil) {
		delta.Add("", a, b)
		return delta
	}

	if ackcompare.HasNilDifference(a.ko.Spec.AliasTarget, b.ko.Spec.AliasTarget) {
		delta.Add("Spec.AliasTarget", a.ko.Spec.AliasTarget, b.ko.Spec.AliasTarget)
	} else if a.ko.Spec.AliasTarget != nil && b.ko.Spec.AliasTarget != nil {
		if ackcompare.HasNilDifference(a.ko.Spec.AliasTarget.DNSName, b.ko.Spec.AliasTarget.DNSName) {
			delta.Add("Spec.AliasTarget.DNSName", a.ko.Spec.AliasTarget.DNSName, b.ko.Spec.AliasTarget.DNSName)
		} else if a.ko.Spec.AliasTarget.DNSName != nil && b.ko.Spec.AliasTarget.DNSName != nil {
			if *a.ko.Spec.AliasTarget.DNSName != *b.ko.Spec.AliasTarget.DNSName {
				delta.Add("Spec.AliasTarget.DNSName", a.ko.Spec.AliasTarget.DNSName, b.ko.Spec.AliasTarget.DNSName)
			}
		}
		if ackcompare.HasNilDifference(a.ko.Spec.AliasTarget.EvaluateTargetHealth, b.ko.Spec.AliasTarget.EvaluateTargetHealth) {
			delta.Add("Spec.AliasTarget.EvaluateTargetHealth", a.ko.Spec.AliasTarget.EvaluateTargetHealth, b.ko.Spec.AliasTarget.EvaluateTargetHealth)
		} else if a.ko.Spec.AliasTarget.EvaluateTargetHealth != nil && b.ko.Spec.AliasTarget.EvaluateTargetHealth != nil {
			if *a.ko.Spec.AliasTarget.EvaluateTargetHealth != *b.ko.Spec.AliasTarget.EvaluateTargetHealth {
				delta.Add("Spec.AliasTarget.EvaluateTargetHealth", a.ko.Spec.AliasTarget.EvaluateTargetHealth, b.ko.Spec.AliasTarget.EvaluateTargetHealth)
			}
		}
		if ackcompare.HasNilDifference(a.ko.Spec.AliasTarget.HostedZoneID, b.ko.Spec.AliasTarget.HostedZoneID) {
			delta.Add("Spec.AliasTarget.HostedZoneID", a.ko.Spec.AliasTarget.HostedZoneID, b.ko.Spec.AliasTarget.HostedZoneID)
		} else if a.ko.Spec.AliasTarget.HostedZoneID != nil && b.ko.Spec.AliasTarget.HostedZoneID != nil {
			if *a.ko.Spec.AliasTarget.HostedZoneID != *b.ko.Spec.AliasTarget.HostedZoneID {
				delta.Add("Spec.AliasTarget.HostedZoneID", a.ko.Spec.AliasTarget.HostedZoneID, b.ko.Spec.AliasTarget.HostedZoneID)
			}
		}
	}
	if ackcompare.HasNilDifference(a.ko.Spec.ChangeBatch, b.ko.Spec.ChangeBatch) {
		delta.Add("Spec.ChangeBatch", a.ko.Spec.ChangeBatch, b.ko.Spec.ChangeBatch)
	} else if a.ko.Spec.ChangeBatch != nil && b.ko.Spec.ChangeBatch != nil {
		if len(a.ko.Spec.ChangeBatch.Changes) != len(b.ko.Spec.ChangeBatch.Changes) {
			delta.Add("Spec.ChangeBatch.Changes", a.ko.Spec.ChangeBatch.Changes, b.ko.Spec.ChangeBatch.Changes)
		} else if len(a.ko.Spec.ChangeBatch.Changes) > 0 {
			if !reflect.DeepEqual(a.ko.Spec.ChangeBatch.Changes, b.ko.Spec.ChangeBatch.Changes) {
				delta.Add("Spec.ChangeBatch.Changes", a.ko.Spec.ChangeBatch.Changes, b.ko.Spec.ChangeBatch.Changes)
			}
		}
		if ackcompare.HasNilDifference(a.ko.Spec.ChangeBatch.Comment, b.ko.Spec.ChangeBatch.Comment) {
			delta.Add("Spec.ChangeBatch.Comment", a.ko.Spec.ChangeBatch.Comment, b.ko.Spec.ChangeBatch.Comment)
		} else if a.ko.Spec.ChangeBatch.Comment != nil && b.ko.Spec.ChangeBatch.Comment != nil {
			if *a.ko.Spec.ChangeBatch.Comment != *b.ko.Spec.ChangeBatch.Comment {
				delta.Add("Spec.ChangeBatch.Comment", a.ko.Spec.ChangeBatch.Comment, b.ko.Spec.ChangeBatch.Comment)
			}
		}
	}
	if ackcompare.HasNilDifference(a.ko.Spec.CIDRRoutingConfig, b.ko.Spec.CIDRRoutingConfig) {
		delta.Add("Spec.CIDRRoutingConfig", a.ko.Spec.CIDRRoutingConfig, b.ko.Spec.CIDRRoutingConfig)
	} else if a.ko.Spec.CIDRRoutingConfig != nil && b.ko.Spec.CIDRRoutingConfig != nil {
		if ackcompare.HasNilDifference(a.ko.Spec.CIDRRoutingConfig.CollectionID, b.ko.Spec.CIDRRoutingConfig.CollectionID) {
			delta.Add("Spec.CIDRRoutingConfig.CollectionID", a.ko.Spec.CIDRRoutingConfig.CollectionID, b.ko.Spec.CIDRRoutingConfig.CollectionID)
		} else if a.ko.Spec.CIDRRoutingConfig.CollectionID != nil && b.ko.Spec.CIDRRoutingConfig.CollectionID != nil {
			if *a.ko.Spec.CIDRRoutingConfig.CollectionID != *b.ko.Spec.CIDRRoutingConfig.CollectionID {
				delta.Add("Spec.CIDRRoutingConfig.CollectionID", a.ko.Spec.CIDRRoutingConfig.CollectionID, b.ko.Spec.CIDRRoutingConfig.CollectionID)
			}
		}
		if ackcompare.HasNilDifference(a.ko.Spec.CIDRRoutingConfig.LocationName, b.ko.Spec.CIDRRoutingConfig.LocationName) {
			delta.Add("Spec.CIDRRoutingConfig.LocationName", a.ko.Spec.CIDRRoutingConfig.LocationName, b.ko.Spec.CIDRRoutingConfig.LocationName)
		} else if a.ko.Spec.CIDRRoutingConfig.LocationName != nil && b.ko.Spec.CIDRRoutingConfig.LocationName != nil {
			if *a.ko.Spec.CIDRRoutingConfig.LocationName != *b.ko.Spec.CIDRRoutingConfig.LocationName {
				delta.Add("Spec.CIDRRoutingConfig.LocationName", a.ko.Spec.CIDRRoutingConfig.LocationName, b.ko.Spec.CIDRRoutingConfig.LocationName)
			}
		}
	}
	if ackcompare.HasNilDifference(a.ko.Spec.Failover, b.ko.Spec.Failover) {
		delta.Add("Spec.Failover", a.ko.Spec.Failover, b.ko.Spec.Failover)
	} else if a.ko.Spec.Failover != nil && b.ko.Spec.Failover != nil {
		if *a.ko.Spec.Failover != *b.ko.Spec.Failover {
			delta.Add("Spec.Failover", a.ko.Spec.Failover, b.ko.Spec.Failover)
		}
	}
	if ackcompare.HasNilDifference(a.ko.Spec.GeoLocation, b.ko.Spec.GeoLocation) {
		delta.Add("Spec.GeoLocation", a.ko.Spec.GeoLocation, b.ko.Spec.GeoLocation)
	} else if a.ko.Spec.GeoLocation != nil && b.ko.Spec.GeoLocation != nil {
		if ackcompare.HasNilDifference(a.ko.Spec.GeoLocation.ContinentCode, b.ko.Spec.GeoLocation.ContinentCode) {
			delta.Add("Spec.GeoLocation.ContinentCode", a.ko.Spec.GeoLocation.ContinentCode, b.ko.Spec.GeoLocation.ContinentCode)
		} else if a.ko.Spec.GeoLocation.ContinentCode != nil && b.ko.Spec.GeoLocation.ContinentCode != nil {
			if *a.ko.Spec.GeoLocation.ContinentCode != *b.ko.Spec.GeoLocation.ContinentCode {
				delta.Add("Spec.GeoLocation.ContinentCode", a.ko.Spec.GeoLocation.ContinentCode, b.ko.Spec.GeoLocation.ContinentCode)
			}
		}
		if ackcompare.HasNilDifference(a.ko.Spec.GeoLocation.CountryCode, b.ko.Spec.GeoLocation.CountryCode) {
			delta.Add("Spec.GeoLocation.CountryCode", a.ko.Spec.GeoLocation.CountryCode, b.ko.Spec.GeoLocation.CountryCode)
		} else if a.ko.Spec.GeoLocation.CountryCode != nil && b.ko.Spec.GeoLocation.CountryCode != nil {
			if *a.ko.Spec.GeoLocation.CountryCode != *b.ko.Spec.GeoLocation.CountryCode {
				delta.Add("Spec.GeoLocation.CountryCode", a.ko.Spec.GeoLocation.CountryCode, b.ko.Spec.GeoLocation.CountryCode)
			}
		}
		if ackcompare.HasNilDifference(a.ko.Spec.GeoLocation.SubdivisionCode, b.ko.Spec.GeoLocation.SubdivisionCode) {
			delta.Add("Spec.GeoLocation.SubdivisionCode", a.ko.Spec.GeoLocation.SubdivisionCode, b.ko.Spec.GeoLocation.SubdivisionCode)
		} else if a.ko.Spec.GeoLocation.SubdivisionCode != nil && b.ko.Spec.GeoLocation.SubdivisionCode != nil {
			if *a.ko.Spec.GeoLocation.SubdivisionCode != *b.ko.Spec.GeoLocation.SubdivisionCode {
				delta.Add("Spec.GeoLocation.SubdivisionCode", a.ko.Spec.GeoLocation.SubdivisionCode, b.ko.Spec.GeoLocation.SubdivisionCode)
			}
		}
	}
	if ackcompare.HasNilDifference(a.ko.Spec.HostedZoneID, b.ko.Spec.HostedZoneID) {
		delta.Add("Spec.HostedZoneID", a.ko.Spec.HostedZoneID, b.ko.Spec.HostedZoneID)
	} else if a.ko.Spec.HostedZoneID != nil && b.ko.Spec.HostedZoneID != nil {
		if *a.ko.Spec.HostedZoneID != *b.ko.Spec.HostedZoneID {
			delta.Add("Spec.HostedZoneID", a.ko.Spec.HostedZoneID, b.ko.Spec.HostedZoneID)
		}
	}
	if !reflect.DeepEqual(a.ko.Spec.HostedZoneRef, b.ko.Spec.HostedZoneRef) {
		delta.Add("Spec.HostedZoneRef", a.ko.Spec.HostedZoneRef, b.ko.Spec.HostedZoneRef)
	}
	if ackcompare.HasNilDifference(a.ko.Spec.MultiValueAnswer, b.ko.Spec.MultiValueAnswer) {
		delta.Add("Spec.MultiValueAnswer", a.ko.Spec.MultiValueAnswer, b.ko.Spec.MultiValueAnswer)
	} else if a.ko.Spec.MultiValueAnswer != nil && b.ko.Spec.MultiValueAnswer != nil {
		if *a.ko.Spec.MultiValueAnswer != *b.ko.Spec.MultiValueAnswer {
			delta.Add("Spec.MultiValueAnswer", a.ko.Spec.MultiValueAnswer, b.ko.Spec.MultiValueAnswer)
		}
	}
	if ackcompare.HasNilDifference(a.ko.Spec.Name, b.ko.Spec.Name) {
		delta.Add("Spec.Name", a.ko.Spec.Name, b.ko.Spec.Name)
	} else if a.ko.Spec.Name != nil && b.ko.Spec.Name != nil {
		if *a.ko.Spec.Name != *b.ko.Spec.Name {
			delta.Add("Spec.Name", a.ko.Spec.Name, b.ko.Spec.Name)
		}
	}
	if ackcompare.HasNilDifference(a.ko.Spec.RecordType, b.ko.Spec.RecordType) {
		delta.Add("Spec.RecordType", a.ko.Spec.RecordType, b.ko.Spec.RecordType)
	} else if a.ko.Spec.RecordType != nil && b.ko.Spec.RecordType != nil {
		if *a.ko.Spec.RecordType != *b.ko.Spec.RecordType {
			delta.Add("Spec.RecordType", a.ko.Spec.RecordType, b.ko.Spec.RecordType)
		}
	}
	if ackcompare.HasNilDifference(a.ko.Spec.Region, b.ko.Spec.Region) {
		delta.Add("Spec.Region", a.ko.Spec.Region, b.ko.Spec.Region)
	} else if a.ko.Spec.Region != nil && b.ko.Spec.Region != nil {
		if *a.ko.Spec.Region != *b.ko.Spec.Region {
			delta.Add("Spec.Region", a.ko.Spec.Region, b.ko.Spec.Region)
		}
	}
	if len(a.ko.Spec.ResourceRecords) != len(b.ko.Spec.ResourceRecords) {
		delta.Add("Spec.ResourceRecords", a.ko.Spec.ResourceRecords, b.ko.Spec.ResourceRecords)
	} else if len(a.ko.Spec.ResourceRecords) > 0 {
		if !reflect.DeepEqual(a.ko.Spec.ResourceRecords, b.ko.Spec.ResourceRecords) {
			delta.Add("Spec.ResourceRecords", a.ko.Spec.ResourceRecords, b.ko.Spec.ResourceRecords)
		}
	}
	if ackcompare.HasNilDifference(a.ko.Spec.SetIdentifier, b.ko.Spec.SetIdentifier) {
		delta.Add("Spec.SetIdentifier", a.ko.Spec.SetIdentifier, b.ko.Spec.SetIdentifier)
	} else if a.ko.Spec.SetIdentifier != nil && b.ko.Spec.SetIdentifier != nil {
		if *a.ko.Spec.SetIdentifier != *b.ko.Spec.SetIdentifier {
			delta.Add("Spec.SetIdentifier", a.ko.Spec.SetIdentifier, b.ko.Spec.SetIdentifier)
		}
	}
	if ackcompare.HasNilDifference(a.ko.Spec.TTL, b.ko.Spec.TTL) {
		delta.Add("Spec.TTL", a.ko.Spec.TTL, b.ko.Spec.TTL)
	} else if a.ko.Spec.TTL != nil && b.ko.Spec.TTL != nil {
		if *a.ko.Spec.TTL != *b.ko.Spec.TTL {
			delta.Add("Spec.TTL", a.ko.Spec.TTL, b.ko.Spec.TTL)
		}
	}
	if ackcompare.HasNilDifference(a.ko.Spec.Weight, b.ko.Spec.Weight) {
		delta.Add("Spec.Weight", a.ko.Spec.Weight, b.ko.Spec.Weight)
	} else if a.ko.Spec.Weight != nil && b.ko.Spec.Weight != nil {
		if *a.ko.Spec.Weight != *b.ko.Spec.Weight {
			delta.Add("Spec.Weight", a.ko.Spec.Weight, b.ko.Spec.Weight)
		}
	}

	return delta
}
