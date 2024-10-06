// Copyright 2024
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ReleaseSpec defines the desired state of Release
type ReleaseSpec struct {
	// Version of the HMC Release in the semver format.
	Version string `json:"version"`
	// HMC references the HMC template.
	HMC CoreProviderTemplate `json:"hmc"`
	// CAPI references the Cluster API template.
	CAPI CoreProviderTemplate `json:"capi"`
	// Providers contains a list of Providers associated with the Release.
	Providers []NamedProviderTemplate `json:"providers,omitempty"`
}

type CoreProviderTemplate struct {
	// Template references the Template associated with the provider.
	Template string `json:"template"`
}

type NamedProviderTemplate struct {
	CoreProviderTemplate `json:",inline"`
	// Name of the provider.
	Name string `json:"name"`
}

func (in *Release) ProviderTemplate(name string) string {
	for _, p := range in.Spec.Providers {
		if p.Name == name {
			return p.Template
		}
	}
	return ""
}

// ReleaseStatus defines the observed state of Release
type ReleaseStatus struct {
	// Templates indicates the status of templates associated with the Release.
	Templates ComponentStatus `json:"templates,omitempty"`
	// Conditions contains details for the current state of the Release
	Conditions []metav1.Condition `json:"conditions,omitempty"`
	// Ready indicates whether HMC is ready to be upgraded to this Release.
	Ready bool `json:"ready,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster

// Release is the Schema for the releases API
type Release struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ReleaseSpec   `json:"spec,omitempty"`
	Status ReleaseStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ReleaseList contains a list of Release
type ReleaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Release `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Release{}, &ReleaseList{})
}
