//
// Copyright 2022 IBM Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IBMLicensingQuerySourceSpec defines the desired state of IBMLicensingQuerySource
type IBMLicensingQuerySourceSpec struct {
	// Policy on how should newly queried data be aggregated to previous data (defaults to MAX)
	// +kubebuilder:validation:Enum=MAX;ADD
	AggregationPolicy string `json:"aggregationPolicy,omitempty"`

	// What query should be send to prometheuses to get the licensing usage
	Query string `json:"query"`

	// Product and cloudpak annotations mapping the query to licensing usage
	Annotations map[string]string `json:"annotations"`
}

// IBMLicensingQuerySourceStatus defines the observed state of IBMLicensingQuerySource
type IBMLicensingQuerySourceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// IBMLicensingQuerySource is the schema for IBM License Service.
// +operator-sdk:csv:customresourcedefinitions:displayName="IBM Licensing Query Source"
// +kubebuilder:resource:path=ibmlicensingquerysources,scope=Namespaced
// +kubebuilder:subresource:status
type IBMLicensingQuerySource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IBMLicensingQuerySourceSpec   `json:"spec,omitempty"`
	Status IBMLicensingQuerySourceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// IBMLicensingQuerySourceList contains a list of IBMLicensingQuerySource
type IBMLicensingQuerySourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IBMLicensingQuerySource `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IBMLicensingQuerySource{}, &IBMLicensingQuerySourceList{})
}
