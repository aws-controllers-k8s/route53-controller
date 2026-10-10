package record_set

import (
	"testing"

	svcapitypes "github.com/aws-controllers-k8s/route53-controller/apis/v1alpha1"
)

func Test_PopulateResourceFromAnnotation(t *testing.T) {
	tests := []struct {
		name         string
		fields       map[string]string
		wantID       *string
		wantErr      bool
		wantTerminal bool
	}{
		{
			name: "id present — sets Status.ID",
			fields: map[string]string{
				"id":           "C1234ABCDE",
				"hostedZoneID": "Z1234",
			},
			wantID: strPtr("C1234ABCDE"),
		},
		{
			name: "id absent — adoption of pre-existing record succeeds",
			fields: map[string]string{
				"hostedZoneID": "Z1234",
			},
			wantID: nil,
		},
		{
			name: "id empty string — treated as absent",
			fields: map[string]string{
				"id":           "",
				"hostedZoneID": "Z1234",
			},
			wantID: nil,
		},
		{
			name: "hostedZoneID absent — terminal error",
			fields: map[string]string{
				"id": "C1234ABCDE",
			},
			wantErr:      true,
			wantTerminal: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &resource{ko: &svcapitypes.RecordSet{}}
			err := r.PopulateResourceFromAnnotation(tt.fields)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantID == nil {
				if r.ko.Status.ID != nil {
					t.Errorf("expected Status.ID nil, got %q", *r.ko.Status.ID)
				}
			} else {
				if r.ko.Status.ID == nil {
					t.Errorf("expected Status.ID %q, got nil", *tt.wantID)
				} else if *r.ko.Status.ID != *tt.wantID {
					t.Errorf("Status.ID = %q, want %q", *r.ko.Status.ID, *tt.wantID)
				}
			}
		})
	}
}

func strPtr(s string) *string { return &s }
