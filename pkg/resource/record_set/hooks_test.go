package record_set

import (
	"testing"
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
