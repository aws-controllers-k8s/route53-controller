package record_set

import (
	"context"
	"testing"

	svcapitypes "github.com/aws-controllers-k8s/route53-controller/apis/v1alpha1"
	"github.com/aws/aws-sdk-go-v2/aws"
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

// callsGetChange reports whether syncStatus proceeded past its short-circuit
// checks to the GetChange API call. rm is constructed with a nil sdkapi, so
// reaching the API call panics on the nil pointer dereference; we recover that
// panic and treat it as "the API path was taken". A clean return means
// syncStatus short-circuited without polling GetChange.
func callsGetChange(ko *svcapitypes.RecordSet) (called bool) {
	rm := &resourceManager{}
	defer func() {
		if recover() != nil {
			called = true
		}
	}()
	_ = rm.syncStatus(context.Background(), ko)
	return false
}

func Test_syncStatus_skipsPollWhenInsync(t *testing.T) {
	tests := []struct {
		testName        string
		id              *string
		status          *string
		wantCallsChange bool
	}{
		{
			testName:        "no change ID skips poll",
			id:              nil,
			status:          aws.String("INSYNC"),
			wantCallsChange: false,
		},
		{
			testName:        "INSYNC change is terminal and skips poll",
			id:              aws.String("/change/C1234567890"),
			status:          aws.String("INSYNC"),
			wantCallsChange: false,
		},
		{
			testName:        "PENDING change is re-polled",
			id:              aws.String("/change/C1234567890"),
			status:          aws.String("PENDING"),
			wantCallsChange: true,
		},
		{
			testName:        "nil status with change ID is re-polled",
			id:              aws.String("/change/C1234567890"),
			status:          nil,
			wantCallsChange: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			ko := &svcapitypes.RecordSet{}
			ko.Status.ID = tt.id
			ko.Status.Status = tt.status

			if got := callsGetChange(ko); got != tt.wantCallsChange {
				t.Errorf("syncStatus polled GetChange = %v, want %v", got, tt.wantCallsChange)
			}
		})
	}
}
