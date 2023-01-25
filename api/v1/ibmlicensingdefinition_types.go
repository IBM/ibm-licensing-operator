// Copyright 2023 IBM Corporation
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
package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type IBMLicensingDefinitionConditionMetadata struct {
	// List of annotations used for matching pod
	Annotations map[string]string `json:"annotations,omitempty"`
	// List of labels used for matching pod
	Labels map[string]string `json:"labels,omitempty"`
}

type IBMLicensingDefinitionCondition struct {
	Metadata IBMLicensingDefinitionConditionMetadata `json:"metadata,omitempty"`
}

// IBMLicensingDefinitionSpec defines the desired state of IBMLicensingDefinition
// +kubebuilder:pruning:PreserveUnknownFields
type IBMLicensingDefinitionSpec struct {
	// Action of Custom Resource
	// +kubebuilder:validation:Enum=modifyOriginal;cloneModify
	Action string `json:"action"`

	// Condition used to match pods
	Condition IBMLicensingDefinitionCondition `json:"condition"`

	// Scope of Custom Resource
	// +kubebuilder:validation:Enum=cluster
	Scope string `json:"scope"`

	// List of annotations that matched pod would be extended
	Set map[string]string `json:"set"`
}

// IBMLicensingDefinitionStatus defines the observed state of IBMLicensingDefinition
type IBMLicensingDefinitionStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// IBMLicensingDefinition is the schema for IBM License Service.
// +operator-sdk:csv:customresourcedefinitions:displayName="IBM Licensing Definition"
// +kubebuilder:resource:path=ibmlicensingdefinitions,scope=Namespaced
// +kubebuilder:subresource:status
type IBMLicensingDefinition struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IBMLicensingDefinitionSpec   `json:"spec,omitempty"`
	Status IBMLicensingDefinitionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// IBMLicensingDefinitionList contains a list of IBMLicensingDefinition
type IBMLicensingDefinitionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IBMLicensingDefinition `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IBMLicensingDefinition{}, &IBMLicensingDefinitionList{})
}
