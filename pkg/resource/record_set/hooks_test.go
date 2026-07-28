package record_set

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"

	svcapitypes "github.com/aws-controllers-k8s/route53-controller/apis/v1alpha1"
)

func Test_getDNSName(t *testing.T) {
	rm := &resourceManager{}

	tests := []struct {
		testName   string
		recordName string
		domain     string
		want       string
	}{
		{
			testName:   "nil name returns hosted zone domain",
			recordName: "",
			domain:     "example.com.",
			want:       "example.com.",
		},
		{
			testName:   "relative subdomain is appended to domain",
			recordName: "www",
			domain:     "example.com.",
			want:       "www.example.com.",
		},
		{
			testName:   "fqdn name returned as-is",
			recordName: "absolute.example.com.",
			domain:     "example.com.",
			want:       "absolute.example.com.",
		},
		{
			testName:   "wildcard subdomain is appended to domain",
			recordName: "*.test",
			domain:     "example.com.",
			want:       "*.test.example.com.",
		},
		{
			testName:   "wildcard fqdn returned as-is",
			recordName: "*.example.com.",
			domain:     "example.com.",
			want:       "*.example.com.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := rm.getDNSName(tt.recordName, tt.domain)
			if got != tt.want {
				t.Errorf("getDNSName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func Test_customPreCompare(t *testing.T) {
	tests := []struct {
		testName     string
		aAliasTarget *svcapitypes.AliasTarget
		bAliasTarget *svcapitypes.AliasTarget
		wantDeltaLen int
	}{
		{
			testName: "DNSName with trailing dot on AWS side normalizes to no dot, no delta",
			aAliasTarget: &svcapitypes.AliasTarget{
				DNSName:              aws.String("my-lb.elb.amazonaws.com"),
				EvaluateTargetHealth: aws.Bool(false),
				HostedZoneID:         aws.String("Z35SXDOTRQ7X7K"),
			},
			bAliasTarget: &svcapitypes.AliasTarget{
				DNSName:              aws.String("my-lb.elb.amazonaws.com."),
				EvaluateTargetHealth: aws.Bool(false),
				HostedZoneID:         aws.String("Z35SXDOTRQ7X7K"),
			},
			wantDeltaLen: 0,
		},
		{
			testName: "DNSName without trailing dot on both sides, no delta",
			aAliasTarget: &svcapitypes.AliasTarget{
				DNSName:              aws.String("my-lb.elb.amazonaws.com"),
				EvaluateTargetHealth: aws.Bool(true),
				HostedZoneID:         aws.String("Z35SXDOTRQ7X7K"),
			},
			bAliasTarget: &svcapitypes.AliasTarget{
				DNSName:              aws.String("my-lb.elb.amazonaws.com"),
				EvaluateTargetHealth: aws.Bool(true),
				HostedZoneID:         aws.String("Z35SXDOTRQ7X7K"),
			},
			wantDeltaLen: 0,
		},
		{
			testName: "HostedZoneID with /hostedzone/ prefix normalizes, no delta",
			aAliasTarget: &svcapitypes.AliasTarget{
				DNSName:              aws.String("my-lb.elb.amazonaws.com"),
				EvaluateTargetHealth: aws.Bool(false),
				HostedZoneID:         aws.String("Z35SXDOTRQ7X7K"),
			},
			bAliasTarget: &svcapitypes.AliasTarget{
				DNSName:              aws.String("my-lb.elb.amazonaws.com"),
				EvaluateTargetHealth: aws.Bool(false),
				HostedZoneID:         aws.String("/hostedzone/Z35SXDOTRQ7X7K"),
			},
			wantDeltaLen: 0,
		},
		{
			testName: "HostedZoneID without prefix on both sides, no delta",
			aAliasTarget: &svcapitypes.AliasTarget{
				DNSName:              aws.String("my-lb.elb.amazonaws.com"),
				EvaluateTargetHealth: aws.Bool(false),
				HostedZoneID:         aws.String("Z35SXDOTRQ7X7K"),
			},
			bAliasTarget: &svcapitypes.AliasTarget{
				DNSName:              aws.String("my-lb.elb.amazonaws.com"),
				EvaluateTargetHealth: aws.Bool(false),
				HostedZoneID:         aws.String("Z35SXDOTRQ7X7K"),
			},
			wantDeltaLen: 0,
		},
		{
			testName: "Both DNSName trailing dot and HostedZoneID prefix need normalization, no delta",
			aAliasTarget: &svcapitypes.AliasTarget{
				DNSName:              aws.String("my-lb.elb.amazonaws.com"),
				EvaluateTargetHealth: aws.Bool(false),
				HostedZoneID:         aws.String("Z35SXDOTRQ7X7K"),
			},
			bAliasTarget: &svcapitypes.AliasTarget{
				DNSName:              aws.String("my-lb.elb.amazonaws.com."),
				EvaluateTargetHealth: aws.Bool(false),
				HostedZoneID:         aws.String("/hostedzone/Z35SXDOTRQ7X7K"),
			},
			wantDeltaLen: 0,
		},
		{
			testName:     "Nil AliasTarget on both sides, no panic, no delta",
			aAliasTarget: nil,
			bAliasTarget: nil,
			wantDeltaLen: 0,
		},
		{
			testName: "Genuinely different DNSName (not just formatting), delta IS present after normalization",
			aAliasTarget: &svcapitypes.AliasTarget{
				DNSName:              aws.String("different-lb.elb.amazonaws.com"),
				EvaluateTargetHealth: aws.Bool(false),
				HostedZoneID:         aws.String("Z35SXDOTRQ7X7K"),
			},
			bAliasTarget: &svcapitypes.AliasTarget{
				DNSName:              aws.String("my-lb.elb.amazonaws.com."),
				EvaluateTargetHealth: aws.Bool(false),
				HostedZoneID:         aws.String("Z35SXDOTRQ7X7K"),
			},
			wantDeltaLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			a := &resource{ko: &svcapitypes.RecordSet{}}
			b := &resource{ko: &svcapitypes.RecordSet{}}
			a.ko.Spec.AliasTarget = tt.aAliasTarget
			b.ko.Spec.AliasTarget = tt.bAliasTarget

			// newResourceDelta internally calls customPreCompare before comparing fields
			delta := newResourceDelta(a, b)

			// Count only differences in AliasTarget fields
			aliasTargetDiffs := 0
			for _, diff := range delta.Differences {
				if diff.Path.Contains("Spec.AliasTarget") {
					aliasTargetDiffs++
				}
			}

			if aliasTargetDiffs != tt.wantDeltaLen {
				t.Errorf("Test %q: got %d AliasTarget differences, want %d. Differences: %v",
					tt.testName, aliasTargetDiffs, tt.wantDeltaLen, delta.Differences)
			}
		})
	}
}
