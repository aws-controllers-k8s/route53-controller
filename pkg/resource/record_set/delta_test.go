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

package record_set

import (
	"testing"

	svcapitypes "github.com/aws-controllers-k8s/route53-controller/apis/v1alpha1"
	"github.com/aws/aws-sdk-go-v2/aws"
)

func recordSetWithAliasTarget(aliasTarget *svcapitypes.AliasTarget) *resource {
	return &resource{
		ko: &svcapitypes.RecordSet{
			Spec: svcapitypes.RecordSetSpec{
				Name:         aws.String("alias"),
				HostedZoneID: aws.String("Z00000000000000000000"),
				RecordType:   aws.String("A"),
				AliasTarget:  aliasTarget,
			},
		},
	}
}

func TestNewResourceDelta_AliasTarget(t *testing.T) {
	tests := []struct {
		testName string
		desired  *svcapitypes.AliasTarget
		latest   *svcapitypes.AliasTarget
		wantDiff bool
	}{
		{
			testName: "identical alias targets",
			desired: &svcapitypes.AliasTarget{
				DNSName:              aws.String("my-lb-123.us-east-1.elb.amazonaws.com"),
				HostedZoneID:         aws.String("Z35SXDOTRQ7X7K"),
				EvaluateTargetHealth: aws.Bool(false),
			},
			latest: &svcapitypes.AliasTarget{
				DNSName:              aws.String("my-lb-123.us-east-1.elb.amazonaws.com"),
				HostedZoneID:         aws.String("Z35SXDOTRQ7X7K"),
				EvaluateTargetHealth: aws.Bool(false),
			},
			wantDiff: false,
		},
		{
			testName: "DNSName trailing dot only on the AWS side",
			desired: &svcapitypes.AliasTarget{
				DNSName:              aws.String("my-lb-123.us-east-1.elb.amazonaws.com"),
				HostedZoneID:         aws.String("Z35SXDOTRQ7X7K"),
				EvaluateTargetHealth: aws.Bool(false),
			},
			latest: &svcapitypes.AliasTarget{
				DNSName:              aws.String("my-lb-123.us-east-1.elb.amazonaws.com."),
				HostedZoneID:         aws.String("Z35SXDOTRQ7X7K"),
				EvaluateTargetHealth: aws.Bool(false),
			},
			wantDiff: false,
		},
		{
			testName: "DNSName trailing dot only on the spec side",
			desired: &svcapitypes.AliasTarget{
				DNSName:              aws.String("my-lb-123.us-east-1.elb.amazonaws.com."),
				HostedZoneID:         aws.String("Z35SXDOTRQ7X7K"),
				EvaluateTargetHealth: aws.Bool(false),
			},
			latest: &svcapitypes.AliasTarget{
				DNSName:              aws.String("my-lb-123.us-east-1.elb.amazonaws.com"),
				HostedZoneID:         aws.String("Z35SXDOTRQ7X7K"),
				EvaluateTargetHealth: aws.Bool(false),
			},
			wantDiff: false,
		},
		{
			testName: "DNSName case difference",
			desired: &svcapitypes.AliasTarget{
				DNSName:              aws.String("My-LB-123.us-east-1.elb.amazonaws.com"),
				HostedZoneID:         aws.String("Z35SXDOTRQ7X7K"),
				EvaluateTargetHealth: aws.Bool(false),
			},
			latest: &svcapitypes.AliasTarget{
				DNSName:              aws.String("my-lb-123.us-east-1.elb.amazonaws.com."),
				HostedZoneID:         aws.String("Z35SXDOTRQ7X7K"),
				EvaluateTargetHealth: aws.Bool(false),
			},
			wantDiff: false,
		},
		{
			testName: "HostedZoneID prefixed only on the AWS side",
			desired: &svcapitypes.AliasTarget{
				DNSName:              aws.String("my-lb-123.us-east-1.elb.amazonaws.com"),
				HostedZoneID:         aws.String("Z35SXDOTRQ7X7K"),
				EvaluateTargetHealth: aws.Bool(false),
			},
			latest: &svcapitypes.AliasTarget{
				DNSName:              aws.String("my-lb-123.us-east-1.elb.amazonaws.com"),
				HostedZoneID:         aws.String("/hostedzone/Z35SXDOTRQ7X7K"),
				EvaluateTargetHealth: aws.Bool(false),
			},
			wantDiff: false,
		},
		{
			testName: "HostedZoneID prefixed only on the spec side",
			desired: &svcapitypes.AliasTarget{
				DNSName:              aws.String("my-lb-123.us-east-1.elb.amazonaws.com"),
				HostedZoneID:         aws.String("/hostedzone/Z35SXDOTRQ7X7K"),
				EvaluateTargetHealth: aws.Bool(false),
			},
			latest: &svcapitypes.AliasTarget{
				DNSName:              aws.String("my-lb-123.us-east-1.elb.amazonaws.com"),
				HostedZoneID:         aws.String("Z35SXDOTRQ7X7K"),
				EvaluateTargetHealth: aws.Bool(false),
			},
			wantDiff: false,
		},
		{
			testName: "trailing dot and hostedzone prefix together",
			desired: &svcapitypes.AliasTarget{
				DNSName:              aws.String("my-lb-123.us-east-1.elb.amazonaws.com"),
				HostedZoneID:         aws.String("Z35SXDOTRQ7X7K"),
				EvaluateTargetHealth: aws.Bool(false),
			},
			latest: &svcapitypes.AliasTarget{
				DNSName:              aws.String("my-lb-123.us-east-1.elb.amazonaws.com."),
				HostedZoneID:         aws.String("/hostedzone/Z35SXDOTRQ7X7K"),
				EvaluateTargetHealth: aws.Bool(false),
			},
			wantDiff: false,
		},
		{
			testName: "EvaluateTargetHealth unset in spec and false from AWS",
			desired: &svcapitypes.AliasTarget{
				DNSName:      aws.String("my-lb-123.us-east-1.elb.amazonaws.com"),
				HostedZoneID: aws.String("Z35SXDOTRQ7X7K"),
			},
			latest: &svcapitypes.AliasTarget{
				DNSName:              aws.String("my-lb-123.us-east-1.elb.amazonaws.com."),
				HostedZoneID:         aws.String("Z35SXDOTRQ7X7K"),
				EvaluateTargetHealth: aws.Bool(false),
			},
			wantDiff: false,
		},
		{
			testName: "no alias target on either side",
			desired:  nil,
			latest:   nil,
			wantDiff: false,
		},
		{
			testName: "alias target added to the spec",
			desired: &svcapitypes.AliasTarget{
				DNSName:              aws.String("my-lb-123.us-east-1.elb.amazonaws.com"),
				HostedZoneID:         aws.String("Z35SXDOTRQ7X7K"),
				EvaluateTargetHealth: aws.Bool(false),
			},
			latest:   nil,
			wantDiff: true,
		},
		{
			testName: "different DNSName",
			desired: &svcapitypes.AliasTarget{
				DNSName:              aws.String("other-lb-456.us-east-1.elb.amazonaws.com"),
				HostedZoneID:         aws.String("Z35SXDOTRQ7X7K"),
				EvaluateTargetHealth: aws.Bool(false),
			},
			latest: &svcapitypes.AliasTarget{
				DNSName:              aws.String("my-lb-123.us-east-1.elb.amazonaws.com."),
				HostedZoneID:         aws.String("Z35SXDOTRQ7X7K"),
				EvaluateTargetHealth: aws.Bool(false),
			},
			wantDiff: true,
		},
		{
			testName: "different HostedZoneID",
			desired: &svcapitypes.AliasTarget{
				DNSName:              aws.String("my-lb-123.us-east-1.elb.amazonaws.com"),
				HostedZoneID:         aws.String("Z2FDTNDATAQYW2"),
				EvaluateTargetHealth: aws.Bool(false),
			},
			latest: &svcapitypes.AliasTarget{
				DNSName:              aws.String("my-lb-123.us-east-1.elb.amazonaws.com"),
				HostedZoneID:         aws.String("/hostedzone/Z35SXDOTRQ7X7K"),
				EvaluateTargetHealth: aws.Bool(false),
			},
			wantDiff: true,
		},
		{
			testName: "different EvaluateTargetHealth",
			desired: &svcapitypes.AliasTarget{
				DNSName:              aws.String("my-lb-123.us-east-1.elb.amazonaws.com"),
				HostedZoneID:         aws.String("Z35SXDOTRQ7X7K"),
				EvaluateTargetHealth: aws.Bool(true),
			},
			latest: &svcapitypes.AliasTarget{
				DNSName:              aws.String("my-lb-123.us-east-1.elb.amazonaws.com."),
				HostedZoneID:         aws.String("Z35SXDOTRQ7X7K"),
				EvaluateTargetHealth: aws.Bool(false),
			},
			wantDiff: true,
		},
		{
			testName: "DNSName unset in spec but reported by AWS",
			desired: &svcapitypes.AliasTarget{
				HostedZoneID:         aws.String("Z35SXDOTRQ7X7K"),
				EvaluateTargetHealth: aws.Bool(false),
			},
			latest: &svcapitypes.AliasTarget{
				DNSName:              aws.String("my-lb-123.us-east-1.elb.amazonaws.com."),
				HostedZoneID:         aws.String("Z35SXDOTRQ7X7K"),
				EvaluateTargetHealth: aws.Bool(false),
			},
			wantDiff: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			a := recordSetWithAliasTarget(tt.desired)
			b := recordSetWithAliasTarget(tt.latest)
			aBefore := a.ko.DeepCopy()
			bBefore := b.ko.DeepCopy()

			delta := newResourceDelta(a, b)

			gotDiff := delta.DifferentAt("Spec.AliasTarget")
			if gotDiff != tt.wantDiff {
				t.Errorf("DifferentAt(Spec.AliasTarget) = %v, want %v (differences: %v)",
					gotDiff, tt.wantDiff, delta.Differences)
			}

			if !equalAliasTarget(aBefore.Spec.AliasTarget, a.ko.Spec.AliasTarget) {
				t.Errorf("newResourceDelta mutated the desired resource: before %+v, after %+v",
					aBefore.Spec.AliasTarget, a.ko.Spec.AliasTarget)
			}
			if !equalAliasTarget(bBefore.Spec.AliasTarget, b.ko.Spec.AliasTarget) {
				t.Errorf("newResourceDelta mutated the latest resource: before %+v, after %+v",
					bBefore.Spec.AliasTarget, b.ko.Spec.AliasTarget)
			}
		})
	}
}

func equalAliasTarget(a *svcapitypes.AliasTarget, b *svcapitypes.AliasTarget) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return aws.ToString(a.DNSName) == aws.ToString(b.DNSName) &&
		aws.ToString(a.HostedZoneID) == aws.ToString(b.HostedZoneID) &&
		aws.ToBool(a.EvaluateTargetHealth) == aws.ToBool(b.EvaluateTargetHealth)
}

func Test_normalizeAliasDNSName(t *testing.T) {
	tests := []struct {
		testName string
		in       string
		want     string
	}{
		{"trailing dot removed", "my-lb.us-east-1.elb.amazonaws.com.", "my-lb.us-east-1.elb.amazonaws.com"},
		{"no trailing dot unchanged", "my-lb.us-east-1.elb.amazonaws.com", "my-lb.us-east-1.elb.amazonaws.com"},
		{"lower cased", "My-LB.US-East-1.ELB.amazonaws.com.", "my-lb.us-east-1.elb.amazonaws.com"},
		{"empty string", "", ""},
		{"lone dot", ".", ""},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			if got := normalizeAliasDNSName(tt.in); got != tt.want {
				t.Errorf("normalizeAliasDNSName(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func Test_normalizeAliasHostedZoneID(t *testing.T) {
	tests := []struct {
		testName string
		in       string
		want     string
	}{
		{"prefix removed", "/hostedzone/Z35SXDOTRQ7X7K", "Z35SXDOTRQ7X7K"},
		{"bare id unchanged", "Z35SXDOTRQ7X7K", "Z35SXDOTRQ7X7K"},
		{"empty string", "", ""},
		{"prefix only", "/hostedzone/", ""},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			if got := normalizeAliasHostedZoneID(tt.in); got != tt.want {
				t.Errorf("normalizeAliasHostedZoneID(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
