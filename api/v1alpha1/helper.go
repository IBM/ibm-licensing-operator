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
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/IBM/ibm-licensing-operator/api/v1alpha1/features"

	"github.com/go-logr/logr"
	routev1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	client_reader "sigs.k8s.io/controller-runtime/pkg/client"
)

const localReporterURL = "https://ibm-license-service-reporter:8080"
const defaultLicensingTokenSecretName = "ibm-licensing-token"         //#nosec
const defaultReporterTokenSecretName = "ibm-licensing-reporter-token" //#nosec
const OperandLicensingImageEnvVar = "IBM_LICENSING_IMAGE"
const OperandUsageImageEnvVar = "IBM_LICENSING_USAGE_IMAGE"
const OperandReporterDatabaseImageEnvVar = "IBM_POSTGRESQL_IMAGE"
const OperandReporterUIImageEnvVar = "IBM_LICENSE_SERVICE_REPORTER_UI_IMAGE"
const OperandReporterReceiverImageEnvVar = "IBM_LICENSE_SERVICE_REPORTER_IMAGE"

var cpu50m = resource.NewMilliQuantity(50, resource.DecimalSI)
var cpu100m = resource.NewMilliQuantity(100, resource.DecimalSI)
var memory64Mi = resource.NewQuantity(64*1024*1024, resource.BinarySI)
var memory128Mi = resource.NewQuantity(128*1024*1024, resource.BinarySI)

var cpu200m = resource.NewMilliQuantity(200, resource.DecimalSI)
var cpu300m = resource.NewMilliQuantity(300, resource.DecimalSI)
var memory256Mi = resource.NewQuantity(256*1024*1024, resource.BinarySI)
var memory300Mi = resource.NewQuantity(256*1024*1024, resource.BinarySI)
var memory384Mi = resource.NewQuantity(384*1024*1024, resource.BinarySI)
var cpu500m = resource.NewMilliQuantity(500, resource.DecimalSI)
var memory1Gi = resource.NewQuantity(1024*1024*1024, resource.BinarySI)
var size1Gi = resource.NewQuantity(1024*1024*1024, resource.BinarySI)

type Container struct {
	// IBM Licensing Service docker Image Registry, will override default value and disable IBM_LICENSING_IMAGE env value in operator deployment
	ImageRegistry string `json:"imageRegistry,omitempty"`
	// IBM Licensing Service docker Image Name, will override default value and disable IBM_LICENSING_IMAGE env value in operator deployment
	ImageName string `json:"imageName,omitempty"`
	// IBM Licensing Service docker Image Tag or Digest, will override default value and disable IBM_LICENSING_IMAGE env value in operator deployment
	ImageTagPostfix string `json:"imageTagPostfix,omitempty"`

	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// +kubebuilder:validation:Enum=Always;IfNotPresent;Never
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
}

type IBMLicenseServiceRouteOptions struct {
	TLS *routev1.TLSConfig `json:"tls,omitempty"`
}

// HTTPSCertsSource describes how certificate is set in available APIs
type HTTPSCertsSource string

const (
	// OcpCertsSource means application will use cert manager
	OcpCertsSource HTTPSCertsSource = "ocp"
	// SelfSignedCertsSource means application will create certificate by itself and use it
	SelfSignedCertsSource HTTPSCertsSource = "self-signed"
	// CustomCertsSource means application will use certificate created by user
	CustomCertsSource HTTPSCertsSource = "custom"

	// Option for operand HTTPS_CERTS_SOURCE
	// ExternalCertsSource means operand will use certificate from a volume mounted to a container
	ExternalCertsSource = "external"
)

type IBMLicenseServiceBaseSpec struct {
	// Should application pod show additional information, options: DEBUG, INFO, VERBOSE
	// +kubebuilder:validation:Enum=DEBUG;INFO;VERBOSE
	LogLevel string `json:"logLevel,omitempty"`
	// Secret name used to store application token, either one that exists, or one that will be created
	APISecretToken string `json:"apiSecretToken,omitempty"`
	// Array of pull secrets which should include existing at InstanceNamespace secret to allow pulling IBM Licensing image
	ImagePullSecrets []string `json:"imagePullSecrets,omitempty"`
	// options: self-signed or custom
	// +kubebuilder:validation:Enum=self-signed;custom;ocp
	HTTPSCertsSource HTTPSCertsSource `json:"httpsCertsSource,omitempty"`
	// Route parameters
	RouteOptions *IBMLicenseServiceRouteOptions `json:"routeOptions,omitempty"`
	// Version
	Version string `json:"version,omitempty"`
}

func (spec *IBMLicensingSpec) IsMetering() bool {
	return spec.Datasource == "metering"
}

func (spec *IBMLicensingSpec) GetDefaultReporterTokenName() string {
	return defaultReporterTokenSecretName
}

func (spec *IBMLicenseServiceBaseSpec) IsDebug() bool {
	return spec.LogLevel == "DEBUG"
}

func (spec *IBMLicenseServiceBaseSpec) IsVerbose() bool {
	return spec.LogLevel == "VERBOSE"
}

func (spec *IBMLicensingSpec) FillDefaultValues(reqLogger logr.Logger, isOCP4CertManager bool, isRouteEnabled bool, rhmpEnabled bool,
	isAlertingEnabledByDefault bool, operatorNamespace string) error {
	if spec.InstanceNamespace == "" {
		spec.InstanceNamespace = operatorNamespace
	}
	spec.Container.setImagePullPolicyIfNotSet()
	if spec.HTTPSCertsSource == "" {
		if isOCP4CertManager {
			spec.HTTPSCertsSource = OcpCertsSource
		} else {
			spec.HTTPSCertsSource = SelfSignedCertsSource
		}
	}
	if spec.RouteEnabled == nil {
		spec.RouteEnabled = &isRouteEnabled
	}
	isNotOnOpenshiftCluster := !isRouteEnabled
	if spec.IngressEnabled == nil {
		spec.IngressEnabled = &isNotOnOpenshiftCluster
	}
	if spec.RHMPEnabled == nil {
		spec.RHMPEnabled = &rhmpEnabled
		if rhmpEnabled {
			reqLogger.Info("RHMP reporting enabled automatically")
		} else {
			reqLogger.Info("RHMP wasn't detected")
		}
	}
	if isAlertingEnabledByDefault {
		if spec.Features == nil {
			spec.Features = &Features{}
		}
		if spec.Features.Alerting == nil {
			spec.Features.Alerting = &features.Alerting{}
		}
		if spec.Features.Alerting.Enabled == nil {
			trueVal := true
			spec.Features.Alerting.Enabled = &trueVal
		}
	}
	if spec.APISecretToken == "" {
		spec.APISecretToken = defaultLicensingTokenSecretName
	}

	spec.Container.initResourcesIfNil()
	spec.Container.setResourceLimitMemoryIfNotSet(*memory1Gi)
	spec.Container.setResourceRequestMemoryIfNotSet(*memory256Mi)
	spec.Container.setResourceLimitCPUIfNotSet(*cpu500m)
	spec.Container.setResourceRequestCPUIfNotSet(*cpu200m)

	if err := spec.setContainer(OperandLicensingImageEnvVar); err != nil {
		return err
	}

	if spec.UsageEnabled {
		spec.UsageContainer.setImagePullPolicyIfNotSet()
		spec.UsageContainer.initResourcesIfNil()
		spec.UsageContainer.setResourceLimitMemoryIfNotSet(*memory128Mi)
		spec.UsageContainer.setResourceRequestMemoryIfNotSet(*memory64Mi)
		spec.UsageContainer.setResourceLimitCPUIfNotSet(*cpu100m)
		spec.UsageContainer.setResourceRequestCPUIfNotSet(*cpu50m)
		if err := spec.UsageContainer.setContainer(OperandUsageImageEnvVar); err != nil {
			return err
		}
	}

	return nil
}

func (spec *IBMLicensingSpec) IsRouteEnabled() bool {
	return spec.RouteEnabled != nil && *spec.RouteEnabled
}

func (spec *IBMLicensingSpec) IsIngressEnabled() bool {
	return spec.IngressEnabled != nil && *spec.IngressEnabled
}

func (spec *IBMLicensingSpec) IsRHMPEnabled() bool {
	return spec.RHMPEnabled != nil && *spec.RHMPEnabled
}

func (spec *IBMLicensingSpec) IsPrometheusServiceNeeded() bool {
	return spec.IsRHMPEnabled() || spec.IsAlertingEnabled()
}

func (spec *IBMLicensingSpec) IsChargebackEnabled() bool {
	if spec.IsRHMPEnabled() {
		return true
	}
	return spec.ChargebackEnabled != nil && *spec.ChargebackEnabled
}

func (spec *IBMLicenseServiceReporterSpec) FillDefaultValues(reqLogger logr.Logger, r client_reader.Reader) error {
	if err := spec.DatabaseContainer.setContainer(OperandReporterDatabaseImageEnvVar); err != nil {
		return err
	}
	if err := spec.ReporterUIContainer.setContainer(OperandReporterUIImageEnvVar); err != nil {
		return err
	}
	if err := spec.ReceiverContainer.setContainer(OperandReporterReceiverImageEnvVar); err != nil {
		return err
	}

	spec.DatabaseContainer.initResourcesIfNil()
	spec.DatabaseContainer.setImagePullPolicyIfNotSet()
	spec.DatabaseContainer.setResourceLimitMemoryIfNotSet(*memory300Mi)
	spec.DatabaseContainer.setResourceRequestMemoryIfNotSet(*memory256Mi)
	spec.DatabaseContainer.setResourceLimitCPUIfNotSet(*cpu300m)
	spec.DatabaseContainer.setResourceRequestCPUIfNotSet(*cpu200m)

	spec.ReceiverContainer.initResourcesIfNil()
	spec.ReceiverContainer.setImagePullPolicyIfNotSet()
	spec.ReceiverContainer.setResourceLimitMemoryIfNotSet(*memory384Mi)
	spec.ReceiverContainer.setResourceRequestMemoryIfNotSet(*memory256Mi)
	spec.ReceiverContainer.setResourceLimitCPUIfNotSet(*cpu300m)
	spec.ReceiverContainer.setResourceRequestCPUIfNotSet(*cpu200m)

	spec.ReporterUIContainer.initResourcesIfNil()
	spec.ReporterUIContainer.setImagePullPolicyIfNotSet()
	spec.ReporterUIContainer.setResourceLimitMemoryIfNotSet(*memory300Mi)
	spec.ReporterUIContainer.setResourceRequestMemoryIfNotSet(*memory256Mi)
	spec.ReporterUIContainer.setResourceLimitCPUIfNotSet(*cpu300m)
	spec.ReporterUIContainer.setResourceRequestCPUIfNotSet(*cpu200m)

	if spec.Capacity.IsZero() {
		spec.Capacity = *size1Gi
	}

	if spec.APISecretToken == "" {
		spec.APISecretToken = defaultReporterTokenSecretName
	}
	if spec.HTTPSCertsSource == "" {
		spec.HTTPSCertsSource = OcpCertsSource
	}
	if spec.StorageClass == "" {
		storageClass, err := getStorageClass(reqLogger, r)
		if err != nil {
			reqLogger.Error(err, "Failed to get StorageCLass for IBM License Service Reporter")
			return err
		}
		spec.StorageClass = storageClass
	}
	return nil

}

func getStorageClass(reqLogger logr.Logger, r client_reader.Reader) (string, error) {
	var defaultSC []string

	scList := &storagev1.StorageClassList{}
	reqLogger.Info("getStorageClass")
	err := r.List(context.TODO(), scList)
	if err != nil {
		return "", err
	}
	if len(scList.Items) == 0 {
		return "", fmt.Errorf("could not find storage class in the cluster")
	}

	for _, sc := range scList.Items {
		if sc.Provisioner == "kubernetes.io/no-provisioner" {
			continue
		}
		if sc.ObjectMeta.GetAnnotations()["storageclass.kubernetes.io/is-default-class"] == "true" {
			defaultSC = append(defaultSC, sc.GetName())
			continue
		}
	}

	if len(defaultSC) != 0 {
		reqLogger.Info("StorageClass configuration", "Name", defaultSC[0])
		return defaultSC[0], nil
	}

	return "", fmt.Errorf("could not find dynamic provisioner default storage class in the cluster")
}

func (container *Container) initResourcesIfNil() {
	if container.Resources.Limits == nil {
		container.Resources.Limits = corev1.ResourceList{}
	}
	if container.Resources.Requests == nil {
		container.Resources.Requests = corev1.ResourceList{}
	}
}

func (container *Container) setResourceLimitCPUIfNotSet(value resource.Quantity) {
	if container.Resources.Limits.Cpu().IsZero() {
		container.Resources.Limits[corev1.ResourceCPU] = value
	}
}

func (container *Container) setResourceRequestCPUIfNotSet(value resource.Quantity) {
	if container.Resources.Requests.Cpu().IsZero() {
		container.Resources.Requests[corev1.ResourceCPU] = value
	}
}

func (container *Container) setResourceLimitMemoryIfNotSet(value resource.Quantity) {
	if container.Resources.Limits.Memory().IsZero() {
		container.Resources.Limits[corev1.ResourceMemory] = value
	}
}

func (container *Container) setResourceRequestMemoryIfNotSet(value resource.Quantity) {
	if container.Resources.Requests.Memory().IsZero() {
		container.Resources.Requests[corev1.ResourceMemory] = value
	}
}

func (container *Container) GetFullImage() string {
	// If there is ":" in image tag then we use "@" for digest as only digest can have it
	if strings.ContainsAny(container.ImageTagPostfix, ":") {
		return container.ImageRegistry + "/" + container.ImageName + "@" + container.ImageTagPostfix
	}
	return container.ImageRegistry + "/" + container.ImageName + ":" + container.ImageTagPostfix
}

// getImageParametersFromEnv get image info from full image reference
func (container *Container) getImageParametersFromEnv(envVariableName string) error {
	fullImageName := os.Getenv(envVariableName)
	// First get imageName, to do that we need to split FullImage like path
	imagePathSplitted := strings.Split(fullImageName, "/")
	if len(imagePathSplitted) < 2 {
		text := fmt.Sprintf("ENV variable: %s should have registry and image separated with \"/\" symbol", envVariableName)
		return errors.New(text)
	}
	imageWithTag := imagePathSplitted[len(imagePathSplitted)-1]
	var imageWithTagSplitted []string
	// Check if digest and split into Image Name and TagPostfix
	if strings.Contains(imageWithTag, "@") {
		imageWithTagSplitted = strings.Split(imageWithTag, "@")
		if len(imageWithTagSplitted) != 2 {
			text := fmt.Sprintf("ENV variable: %s in operator deployment should have digest and image name separated by only one \"@\" symbol", envVariableName)
			return errors.New(text)
		}
	} else {
		imageWithTagSplitted = strings.Split(imageWithTag, ":")
		if len(imageWithTagSplitted) != 2 {
			text := fmt.Sprintf("ENV variable: %s in operator deployment should have image tag and image name separated by only one \":\" symbol", envVariableName)
			return errors.New(text)
		}
	}
	container.ImageTagPostfix = imageWithTagSplitted[1]
	container.ImageName = imageWithTagSplitted[0]
	container.ImageRegistry = strings.Join(imagePathSplitted[:len(imagePathSplitted)-1], "/")
	return nil
}

func (container *Container) setContainer(envVar string) error {
	temp := Container{}
	if err := temp.getImageParametersFromEnv(envVar); err != nil {
		return err
	}
	// If CR has at least one override, make sure all parts of the image are filled at least with default values c ENV
	if container.ImageName == "" {
		container.ImageName = temp.ImageName
	}
	if container.ImageRegistry == "" {
		container.ImageRegistry = temp.ImageRegistry
	}
	if container.ImageTagPostfix == "" {
		container.ImageTagPostfix = temp.ImageTagPostfix
	}
	return nil
}

//nolint:revive
func CheckOperandEnvVar() error {
	c := Container{}
	if err := c.getImageParametersFromEnv(OperandLicensingImageEnvVar); err != nil {
		return err
	}
	if err := c.getImageParametersFromEnv(OperandReporterDatabaseImageEnvVar); err != nil {
		return err
	}
	if err := c.getImageParametersFromEnv(OperandReporterUIImageEnvVar); err != nil {
		return err
	}
	if err := c.getImageParametersFromEnv(OperandReporterReceiverImageEnvVar); err != nil {
		return err
	}

	return nil
}

func (container *Container) setImagePullPolicyIfNotSet() {
	if container.ImagePullPolicy == "" {
		container.ImagePullPolicy = corev1.PullIfNotPresent
	}
}

func (spec *IBMLicensingSpec) SetDefaultSenderParameters() bool {

	//returns true if any changes were made
	changed := false

	if spec.Sender == nil {
		spec.Sender = &IBMLicensingSenderSpec{}
	}

	if spec.Sender.ReporterURL == "" {
		spec.Sender.ReporterURL = localReporterURL
		changed = true
	}

	if spec.Sender.ReporterSecretToken == "" {
		spec.Sender.ReporterSecretToken = defaultReporterTokenSecretName
		changed = true
	}

	return changed
}

func (spec *IBMLicensingSpec) RemoveDefaultSenderParameters() bool {

	//returns true if any changes were made

	//checking only token because secret removes automatically and may cause k8s config error
	//if checking also url and not removing on only url change
	if spec.Sender != nil && spec.Sender.ReporterSecretToken == defaultReporterTokenSecretName {
		spec.Sender = nil
		return true
	}

	return false
}
