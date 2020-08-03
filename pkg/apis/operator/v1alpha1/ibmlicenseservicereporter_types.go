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
	routev1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type IBMLicenseServiceRouteOptions struct {
	TLS *routev1.TLSConfig `json:"tls,omitempty"`
}

// License properties
type License struct {
	// Accept is an opt-in license acceptance required to deploy resources
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="License Acceptance"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:booleanSwitch"
	Accept bool `json:"accept"`
}

type Container struct {
	// Docker Image
	Image string `json:"image,omitempty"`
	// Resources and limits for container
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

// IBMLicenseServiceReporterSpec defines the desired state of IBMLicenseServiceReporter
type IBMLicenseServiceReporterSpec struct {

	// Opt-in license acceptance is required to create resources
	License License `json:"license"`

	// Receiver Settings
	ReceiverContainer Container `json:"receiverContainer,omitempty"`

	// Database Settings
	DatabaseContainer Container `json:"databaseContainer,omitempty"`

	// Storage class used by database to provide persistency
	StorageClass string `json:"storageClass,omitempty"`

	// Persistent Volume Claim Capacity
	Capacity string `json:"capacity,omitempty"`

	// IBM License Service Reporter Pod pull policy, default: IfNotPresent
	// +kubebuilder:validation:Enum=Always;IfNotPresent;Never
	ImagePullPolicy string `json:"imagePullPolicy,omitempty"`

	// Array of pull secrets which should include existing at InstanceNamespace secret to allow pulling IBM Licensing Reporter images
	ImagePullSecrets []string `json:"imagePullSecrets,omitempty"`

	// Secret name used to store application token, either one that exists, or one that will be created
	APISecretToken string `json:"apiSecretToken,omitempty"`

	// options: self-signed or custom
	// +kubebuilder:validation:Enum=self-signed;custom
	HTTPSCertsSource string `json:"httpsCertsSource,omitempty"`

	// Route parameters
	RouteOptions *IBMLicenseServiceRouteOptions `json:"routeOptions,omitempty"`

	// Version
	Version string `json:"version,omitempty"`
}

// IBMLicenseServiceReporterStatus defines the observed state of IBMLicenseServiceReporter
type IBMLicenseServiceReporterStatus struct {
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	LicensingReporterPods []corev1.PodStatus `json:"LicensingReporterPods"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IBMLicenseServiceReporter is the Schema for the ibmlicenseservicereporters API
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
