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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HealthCheckSpec defines the desired state of HealthCheck.
//
// A complex type that contains information about one health check that is associated
// with the current Amazon Web Services account.
type HealthCheckSpec struct {

	// A complex type that contains settings for a new health check.
	// +kubebuilder:validation:Required
	HealthCheckConfig *HealthCheckConfig `json:"healthCheckConfig"`
	// A complex type that contains a list of the tags that you want to add to the
	// specified health check or hosted zone and/or the tags that you want to edit
	// Value for.
	//
	// You can add a maximum of 10 tags to a health check or a hosted zone.
	Tags []*Tag `json:"tags,omitempty"`
}

// HealthCheckStatus defines the observed state of HealthCheck
type HealthCheckStatus struct {
	// All CRs managed by ACK have a common `Status.ACKResourceMetadata` member
	// that is used to contain resource sync state, account ownership,
	// constructed ARN for the resource
	// +kubebuilder:validation:Optional
	ACKResourceMetadata *ackv1alpha1.ResourceMetadata `json:"ackResourceMetadata"`
	// All CRs managed by ACK have a common `Status.Conditions` member that
	// contains a collection of `ackv1alpha1.Condition` objects that describe
	// the various terminal states of the CR and its backend AWS service API
	// resource
	// +kubebuilder:validation:Optional
	Conditions []*ackv1alpha1.Condition `json:"conditions"`
	// A unique string that you specified when you created the health check.
	// +kubebuilder:validation:Optional
	CallerReference *string `json:"callerReference,omitempty"`
	// A complex type that contains information about the CloudWatch alarm that
	// Amazon Route 53 is monitoring for this health check.
	// +kubebuilder:validation:Optional
	CloudWatchAlarmConfiguration *CloudWatchAlarmConfiguration `json:"cloudWatchAlarmConfiguration,omitempty"`
	// The version of the health check. You can optionally pass this value in a
	// call to UpdateHealthCheck to prevent overwriting another change to the health
	// check.
	// +kubebuilder:validation:Optional
	HealthCheckVersion *int64 `json:"healthCheckVersion,omitempty"`
	// The identifier that Amazon Route 53 assigned to the health check when you
	// created it. When you add or update a resource record set, you use this value
	// to specify which health check to use. The value can be up to 64 characters
	// long.
	// +kubebuilder:validation:Optional
	ID *string `json:"id,omitempty"`
	// If the health check was created by another service, the service that created
	// the health check. When a health check is created by another service, you
	// can't edit or delete it using Amazon Route 53.
	// +kubebuilder:validation:Optional
	LinkedService *LinkedService `json:"linkedService,omitempty"`
}

// HealthCheck is the Schema for the HealthChecks API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type HealthCheck struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              HealthCheckSpec   `json:"spec,omitempty"`
	Status            HealthCheckStatus `json:"status,omitempty"`
}

// HealthCheckList contains a list of HealthCheck
// +kubebuilder:object:root=true
type HealthCheckList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HealthCheck `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HealthCheck{}, &HealthCheckList{})
}
