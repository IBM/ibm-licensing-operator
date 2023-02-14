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
	routev1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type IBMLicensingIngressOptions struct {

	// Path after host where API will be available f.e. https://<hostname>:<port>/ibm-licensing-service-instance
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Path",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	// +optional
	Path *string `json:"path,omitempty"`

	// Additional annotations that should include f.e. ingress class if using not default ingress controller
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Annotations",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// TLS Options to enable secure connection
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="TLS",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	// +optional
	TLS []networkingv1.IngressTLS `json:"tls,omitempty"`

	// If you use non-default host include it here
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Host",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	// +optional
	Host *string `json:"host,omitempty"`
}

type IBMLicensingRouteOptions struct {

	// TLS Config
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="TLS Config",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	// +optional
	TLS *routev1.TLSConfig `json:"tls,omitempty"`
}

// IBMLicensingSpec defines the desired state of IBMLicensing
// +kubebuilder:pruning:PreserveUnknownFields
type IBMLicensingSpec struct {

	// Environment variable setting
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Environment variable setting",xDescriptors="urn:alm:descriptor:com.tectonic.ui:hidden"
	// +optional
	EnvVariable map[string]string `json:"envVariable,omitempty"`

	// Container Settings
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Container Settings",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	// +optional
	Container `json:",inline"`

	// Common Parameters for Operator
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Common Parameters",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	// +optional
	IBMLicenseServiceBaseSpec `json:",inline"`

	// Where should data be collected, options: metering, datacollector
	// +kubebuilder:validation:Enum=metering;datacollector
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Datasource",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	Datasource string `json:"datasource"`

	// Enables https access at pod level, httpsCertsSource needed if true
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="HTTPS Enable",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	HTTPSEnable bool `json:"httpsEnable"`

	// Existing or to be created namespace where application will start. In case metering data collection is used,
	// should be the same namespace as metering components
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Instance Namespace",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	InstanceNamespace string `json:"instanceNamespace,omitempty"`

	// If default SCC user ID fails, you can set runAsUser option to fix that
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Security Context",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	// +optional
	SecurityContext *IBMLicensingSecurityContext `json:"securityContext,omitempty"`

	// Should Route be created to expose IBM Licensing Service API? (only on OpenShift cluster)
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Route Enabled",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	// +optional
	RouteEnabled *bool `json:"routeEnabled,omitempty"`

	// Is Red Hat Marketplace enabled
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="RHMP Enabled",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	// +optional
	RHMPEnabled *bool `json:"rhmpEnabled,omitempty"`

	// Should collect usage based metrics?
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Usage Enabled",xDescriptors="urn:alm:descriptor:com.tectonic.ui:hidden"
	// +optional
	UsageEnabled bool `json:"usageEnabled,omitempty"`

	// Usage Container Settings
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Usage Container Settings",xDescriptors="urn:alm:descriptor:com.tectonic.ui:hidden"
	// +optional
	UsageContainer Container `json:"usageContainer,omitempty"`

	// Consider updating to enable chargeback feature
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Chargeback Enabled",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	// +optional
	ChargebackEnabled *bool `json:"chargebackEnabled,omitempty"`

	// Chargeback data retention period in days. Default value is 62 days.
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Chargeback Retention Period in days",xDescriptors="urn:alm:descriptor:com.tectonic.ui:number"
	// +optional
	ChargebackRetentionPeriod *int `json:"chargebackRetentionPeriod,omitempty"`

	// Should Ingress be created to expose IBM Licensing Service API?
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Ingress Enabled",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	// +optional
	IngressEnabled *bool `json:"ingressEnabled,omitempty"`

	// If ingress is enabled, you can set its parameters
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Ingress Options",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	// +optional
	IngressOptions *IBMLicensingIngressOptions `json:"ingressOptions,omitempty"`

	// Sender configuration, set if you have multi-cluster environment from which you collect data
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Sender",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	// +optional
	Sender *IBMLicensingSenderSpec `json:"sender,omitempty"`

	// Set additional features under this field
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Features"
	// +optional
	Features *Features `json:"features,omitempty"`
}

type IBMLicensingSenderSpec struct {

	// URL for License Service Reporter receiver that collects and aggregate multi cluster licensing data.
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Reporter URL",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	// +optional
	ReporterURL string `json:"reporterURL,omitempty"`

	// License Service Reporter authentication token, provided by secret that you need to create in instance namespace
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Reporter Secret Token",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	// +optional
	ReporterSecretToken string `json:"reporterSecretToken,omitempty"`

	// What is the name of this reporting cluster in multi-cluster system. If not provided, CLUSTER_ID will be used as CLUSTER_NAME at Operand level
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Cluster Name",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	// +optional
	ClusterName string `json:"clusterName,omitempty"`

	// Unique ID of reporting cluster
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Cluster ID",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	// +optional
	ClusterID string `json:"clusterID,omitempty"`
}

type IBMLicensingSecurityContext struct {
	RunAsUser int64 `json:"runAsUser"`
}

// IBMLicensingStatus defines the observed state of IBMLicensing
type IBMLicensingStatus struct {
	// State field that defines status of the IBMLicensing
	State string `json:"state,omitempty"`
	// The status of IBM License Service Pods.
	LicensingPods []corev1.PodStatus         `json:"licensingPods,omitempty"`
	Features      IBMLicensingFeaturesStatus `json:"features,omitempty"`
}

type IBMLicensingFeaturesStatus struct {
	RHMPEnabled *bool `json:"rhmpEnabled,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IBM License Service is the Schema for the ibmlicensings API.
// Documentation For additional details regarding install parameters check: https://ibm.biz/icpfs39install.
// License By installing this product you accept the license terms https://ibm.biz/icpfs39license.
// +kubebuilder:printcolumn:name="Pod Phase",type=string,JSONPath=`.status..phase`
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=ibmlicensings,scope=Cluster
// +operator-sdk:csv:customresourcedefinitions:displayName="IBM License Service"
// +operator-sdk:csv:customresourcedefinitions:resources={{Service,v1,},{Pod,v1,}}
// +operator-sdk:csv:customresourcedefinitions:resources={{Deployment,v1,},{Secret,v1,}}
// +operator-sdk:csv:customresourcedefinitions:resources={{Route,v1,},{ServiceAccount,v1,}}
// +operator-sdk:csv:customresourcedefinitions:resources={{Role,v1,},{RoleBinding,v1,}}
// +operator-sdk:csv:customresourcedefinitions:resources={{ReplicaSets,v1,},{Ingresses,v1beta1,}}
// +operator-sdk:csv:customresourcedefinitions:resources={{status,v1alpha1,},{configmaps,v1,}}
// +operator-sdk:csv:customresourcedefinitions:resources={{ibmlicensings,v1alpha1,},{ibmlicenseservicereporters,v1alpha1},{ibmlicensingmetadatas,v1alpha1}}
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
