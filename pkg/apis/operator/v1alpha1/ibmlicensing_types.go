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
	extensionsv1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type IBMLicensingIngressOptions struct {
	// Path after host where API will be available f.e. https://<hostname>:<port>/ibm-licensing-service-instance
	Path *string `json:"path,omitempty"`
	// Additional annotations that should include f.e. ingress class if using not default ingress controller
	Annotations map[string]string `json:"annotations,omitempty"`
	// TLS Options to enable secure connection
	TLS []extensionsv1.IngressTLS `json:"tls,omitempty"`
	// If you use non-default host include it here
	Host *string `json:"host,omitempty"`
}

type IBMLicensingRouteOptions struct {
	TLS *routev1.TLSConfig `json:"tls,omitempty"`
}

// IBMLicensingSpec defines the desired state of IBMLicensing
type IBMLicensingSpec struct {
	// deprecated, define image in operator.yaml
	ImageRegistry string `json:"imageRegistry,omitempty"`
	// deprecated, define image in operator.yaml
	ImageName string `json:"imageName,omitempty"`
	// deprecated, define image in operator.yaml
	ImageTagPostfix string `json:"imageTagPostfix,omitempty"`
	// Secret name used to store application token, either one that exists, or one that will be created, for now only one value possible
	// +kubebuilder:validation:Enum=ibm-licensing-token
	APISecretToken string `json:"apiSecretToken,omitempty"`
	// Where should data be collected, options: metering, datacollector
	// +kubebuilder:validation:Enum=metering;datacollector
	Datasource string `json:"datasource"`
	// Enables https access at pod level, httpsCertsSource needed if true
	HTTPSEnable bool `json:"httpsEnable"`
	// options: self-signed or custom
	// +kubebuilder:validation:Enum=self-signed;custom
	HTTPSCertsSource string `json:"httpsCertsSource,omitempty"`
	// Should application pod show additional information, options: DEBUG, INFO
	// +kubebuilder:validation:Enum=DEBUG;INFO
	LogLevel string `json:"logLevel,omitempty"`
	// Existing or to be created namespace where application will start. In case metering data collection is used,
	// should be the same namespace as metering components
	InstanceNamespace string `json:"instanceNamespace"`
	// If default SCC user ID fails, you can set runAsUser option to fix that
	SecurityContext *IBMLicensingSecurityContext `json:"securityContext,omitempty"`
	// Array of pull secrets which should include existing at InstanceNamespace secret to allow pulling IBM Licensing image
	// +listType=set
	ImagePullSecrets []string `json:"imagePullSecrets,omitempty"`
	// Should Route be created to expose IBM Licensing Service API? (only on OpenShift cluster)
	RouteEnabled *bool `json:"routeEnabled,omitempty"`
	// Should Ingress be created to expose IBM Licensing Service API?
	IngressEnabled *bool `json:"ingressEnabled,omitempty"`
	// If ingress is enabled, you can set its parameters
	IngressOptions *IBMLicensingIngressOptions `json:"ingressOptions,omitempty"`
	// If route is enabled, you can set its parameters
	RouteOptions *IBMLicensingRouteOptions `json:"routeOptions,omitempty"`
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
	LicensingPods []corev1.PodStatus `json:"licensingPods"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IBMLicensing is the Schema for the ibmlicensings API
// +kubebuilder:printcolumn:name="Pod Phase",type=string,JSONPath=`.status..phase`
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
