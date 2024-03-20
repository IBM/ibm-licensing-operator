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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IBMLicenseServiceReporterSpec defines the desired state of IBMLicenseServiceReporter
// +kubebuilder:pruning:PreserveUnknownFields
type IBMLicenseServiceReporterSpec struct {
	// Environment variable setting
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Environment variable setting",xDescriptors="urn:alm:descriptor:com.tectonic.ui:hidden"
	// +optional
	EnvVariable map[string]string `json:"envVariable,omitempty"`
	// Receiver Settings
	ReceiverContainer Container `json:"receiverContainer,omitempty"`
	// Receiver Settings
	ReporterUIContainer Container `json:"reporterUIContainer,omitempty"`
	// Database Settings
	DatabaseContainer Container `json:"databaseContainer,omitempty"`
	// Common Parameters for operator
	IBMLicenseServiceBaseSpec `json:",inline"`
	// Storage class used by database to provide persistency
	StorageClass string `json:"storageClass,omitempty"`
	// Persistent Volume Claim Capacity
	Capacity resource.Quantity `json:"capacity,omitempty" protobuf:"bytes,2,opt,name=capacity"`
}

// IBMLicenseServiceReporterStatus defines the observed state of IBMLicenseServiceReporter
type IBMLicenseServiceReporterStatus struct {
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	// +optional
	LicensingReporterPods []corev1.PodStatus `json:"LicensingReporterPods,omitempty"`

	// Compatibility with LicenseReporter v4.x
	// +optinal
	LicenseServiceReporterPods []corev1.PodStatus `json:"LicenseServiceReporterPods,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IBMLicenseServiceReporter is the Schema for the ibmlicenseservicereporters API.
// Documentation For additional details regarding install parameters check: https://ibm.biz/icpfs39install.
// License By installing this product you accept the license terms https://ibm.biz/icpfs39license.
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=ibmlicenseservicereporters,scope=Namespaced
type IBMLicenseServiceReporter struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IBMLicenseServiceReporterSpec   `json:"spec,omitempty"`
	Status IBMLicenseServiceReporterStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IBMLicenseServiceReporterList contains a list of IBMLicenseServiceReporter
type IBMLicenseServiceReporterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IBMLicenseServiceReporter `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IBMLicenseServiceReporter{}, &IBMLicenseServiceReporterList{})
}
