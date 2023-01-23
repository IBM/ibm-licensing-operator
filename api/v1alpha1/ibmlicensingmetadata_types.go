//
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
//

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type IBMLicensingMetadataCondition struct {
	// List of annotations used for matching pod
	Annotation map[string]string `json:"annotation"`
}

// IBMLicensingMetadataSpec defines the desired state of IBMLicensingMetadata
// +kubebuilder:pruning:PreserveUnknownFields
type IBMLicensingMetadataSpec struct {
	Condition IBMLicensingMetadataCondition `json:"condition"`
	// List of annotations that matched pod would be extended
	Extend map[string]string `json:"extend"`
}

// IBMLicensingMetadataStatus defines the observed state of IBMLicensingMetadata
type IBMLicensingMetadataStatus struct {
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// IBMLicensingMetadata is the schema for IBM License Service. Thanks to IBMLicensingMetadata, you can track
// the license usage of Virtual Processor Core (VPC) metric by Pods that are managed by automation,
// for which licensing annotations are not defined at the time of deployment, such as the open source OpenLiberty.
// / For more information, see documentation: https://ibm.biz/icpfs39install.
// License: By installing this product you accept the license terms. For more information about the license,
// see https://ibm.biz/icpfs39license.
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=ibmlicensingmetadatas,scope=Namespaced
// +kubebuilder:deprecatedversion
type IBMLicensingMetadata struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IBMLicensingMetadataSpec   `json:"spec,omitempty"`
	Status IBMLicensingMetadataStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IBMLicensingMetadataList contains a list of IBMLicensingMetadata
type IBMLicensingMetadataList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IBMLicensingMetadata `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IBMLicensingMetadata{}, &IBMLicensingMetadataList{})
}
