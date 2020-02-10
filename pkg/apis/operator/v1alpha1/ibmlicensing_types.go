//
// Copyright 2020 IBM Corporation
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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// IBMLicensingSpec defines the desired state of IBMLicensing
type IBMLicensingSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	ImageRegistry   string `json:"imageRegistry"`
	ImageTagPostfix string `json:"imageTagPostfix"`
	APISecretToken  string `json:"apiSecretToken"`
	// ?TODO: maybe change to enum in future:
	Datasource  string `json:"datasource"`
	HTTPSEnable bool   `json:"httpsEnable"`
	// ?TODO: maybe change to enum in future:
	HTTPSCertsSource string `json:"httpsCertsSource,omitempty"`
	// ?TODO: maybe change to enum in future:
	LogLevel         string                      `json:"logLevel,omitempty"`
	APINamespace     string                      `json:"apiNamespace"`
	SecurityContext  IBMLicensingSecurityContext `json:"securityContext,omitempty"`
	ImagePullSecrets []string                    `json:"imagePullSecrets,omitempty"`
}

type IBMLicensingSecurityContext struct {
	RunAsUser int64 `json:"runAsUser"`
}

// IBMLicensingStatus defines the observed state of IBMLicensing
type IBMLicensingStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// LicensingPods are the names of the licensing pods
	// +listType=set
	LicensingPods []string `json:"licensingPods"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IBMLicensing is the Schema for the ibmlicensings API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=ibmlicensings,scope=Cluster
type IBMLicensing struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IBMLicensingSpec   `json:"spec,omitempty"`
	Status IBMLicensingStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IBMLicensingList contains a list of IBMLicensing
type IBMLicensingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IBMLicensing `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IBMLicensing{}, &IBMLicensingList{})
}
