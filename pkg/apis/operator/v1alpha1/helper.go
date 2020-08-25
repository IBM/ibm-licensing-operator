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
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/go-logr/logr"
	routev1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	client_reader "sigs.k8s.io/controller-runtime/pkg/client"
)

const defaultQuayRegistry = "quay.io/opencloudio"

const defaultLicensingImageName = "ibm-licensing"
const defaultLicensingImageTagPostfix = "sha256:5d92000b11499b6505967ade7e1ad093089704371834923b1d67f8bc950599a2"

const defaultReporterImageName = "ibm-license-service-reporter"
const defaultReporterImageTagPostfix = "sha256:ec8edc42a291bbf3887e371b29ada2432bfd910e3113648594c4951721dfed63"

const defaultReporterUIImageName = "ibm-license-service-reporter-ui"
const defaultReporterUIImageTagPostfix = "sha256:7696f035fe4f3a9405efeeab8bef44a2aa9f5f90c15329d28c2238356c2b300f"

const defaultDatabaseImageName = "ibm-postgresql"
const defaultDatabaseImageTagPostfix = "sha256:5e68245c21a7252afcca65f82faabdc7551429b844896f8499a52bf092c49caf"

var cpu200m = resource.NewMilliQuantity(200, resource.DecimalSI)
var cpu300m = resource.NewMilliQuantity(300, resource.DecimalSI)
var memory256Mi = resource.NewQuantity(256*1024*1024, resource.BinarySI)
var memory300Mi = resource.NewQuantity(256*1024*1024, resource.BinarySI)
var cpu500m = resource.NewMilliQuantity(500, resource.DecimalSI)
var memory512Mi = resource.NewQuantity(512*1024*1024, resource.BinarySI)
var size1Gi = resource.NewQuantity(1024*1024*1024, resource.BinarySI)

type Container struct {
	// IBM Licensing Service docker Image Registry, will override default value and disable OPERAND_LICENSING_IMAGE env value in operator deployment
	ImageRegistry string `json:"imageRegistry,omitempty"`
	// IBM Licensing Service docker Image Name, will override default value and disable OPERAND_LICENSING_IMAGE env value in operator deployment
	ImageName string `json:"imageName,omitempty"`
	// IBM Licensing Service docker Image Tag or Digest, will override default value and disable OPERAND_LICENSING_IMAGE env value in operator deployment
	ImageTagPostfix string `json:"imageTagPostfix,omitempty"`

	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// +kubebuilder:validation:Enum=Always;IfNotPresent;Never
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
}

type IBMLicenseServiceRouteOptions struct {
	TLS *routev1.TLSConfig `json:"tls,omitempty"`
}

type IBMLicenseServiceBaseSpec struct {
	// Should application pod show additional information, options: DEBUG, INFO
	// +kubebuilder:validation:Enum=DEBUG;INFO
	LogLevel string `json:"logLevel,omitempty"`
	// Secret name used to store application token, either one that exists, or one that will be created
	APISecretToken string `json:"apiSecretToken,omitempty"`
	// Array of pull secrets which should include existing at InstanceNamespace secret to allow pulling IBM Licensing image
	ImagePullSecrets []string `json:"imagePullSecrets,omitempty"`
	// options: self-signed or custom
	// +kubebuilder:validation:Enum=self-signed;custom
	HTTPSCertsSource string `json:"httpsCertsSource,omitempty"`
	// Route parameters
	RouteOptions *IBMLicenseServiceRouteOptions `json:"routeOptions,omitempty"`
	// Version
	Version string `json:"version,omitempty"`
}

func (spec *IBMLicensingSpec) IsMetering() bool {
	return spec.Datasource == "metering"
}

func (spec *IBMLicensingSpec) IsDebug() bool {
	return spec.LogLevel == "DEBUG"
}

func (container *Container) GetFullImage() string {
	// If there is ":" in image tag then we use "@" for digest as only digest can have it
	if strings.ContainsAny(container.ImageTagPostfix, ":") {
		return container.ImageRegistry + "/" + container.ImageName + "@" + container.ImageTagPostfix
	}
	return container.ImageRegistry + "/" + container.ImageName + ":" + container.ImageTagPostfix
}

// isImageEmpty returns true when any part of image name is not defined
func (container *Container) isImageEmpty() bool {
	return container.ImageRegistry == "" && container.ImageName == "" && container.ImageTagPostfix == ""
}

// setImageParametersFromEnv set container image info from full image reference
func (container *Container) setImageParametersFromEnv(fullImageName string) error {
	// First get imageName, to do that we need to split FullImage like path
	imagePathSplitted := strings.Split(fullImageName, "/")
	if len(imagePathSplitted) < 2 {
		return errors.New("your image ENV variable in operator deployment should have registry and image separated with \"/\" symbol")
	}
	imageWithTag := imagePathSplitted[len(imagePathSplitted)-1]
	var imageWithTagSplitted []string
	// Check if digest and split into Image Name and TagPostfix
	if strings.Contains(imageWithTag, "@") {
		imageWithTagSplitted = strings.Split(imageWithTag, "@")
		if len(imageWithTagSplitted) != 2 {
			return errors.New("your image ENV variable in operator deployment should have digest and image name separated by only one \"@\" symbol")
		}
	} else {
		imageWithTagSplitted = strings.Split(imageWithTag, ":")
		if len(imageWithTagSplitted) != 2 {
			return errors.New("your image ENV variable in operator deployment should have image tag and image name separated by only one \":\" symbol")
		}
	}
	container.ImageTagPostfix = imageWithTagSplitted[1]
	container.ImageName = imageWithTagSplitted[0]
	container.ImageRegistry = strings.Join(imagePathSplitted[:len(imagePathSplitted)-1], "/")
	return nil
}

func (container *Container) setImagePullPolicyIfNotSet() {
	if container.ImagePullPolicy == "" {
		container.ImagePullPolicy = corev1.PullIfNotPresent
	}
}

func (spec *IBMLicensingSpec) FillDefaultValues(isOpenshiftCluster bool) error {
	spec.Container.setImagePullPolicyIfNotSet()
	if spec.HTTPSCertsSource == "" {
		spec.HTTPSCertsSource = "self-signed"
	}
	if spec.RouteEnabled == nil {
		spec.RouteEnabled = &isOpenshiftCluster
	}
	isNotOnOpenshiftCluster := !isOpenshiftCluster
	if spec.IngressEnabled == nil {
		spec.IngressEnabled = &isNotOnOpenshiftCluster
	}
	if spec.APISecretToken == "" {
		spec.APISecretToken = "ibm-licensing-token"
	}

	spec.Container.initResourcesIfNil()
	spec.Container.setResourceLimitMemoryIfNotSet(*memory512Mi)
	spec.Container.setResourceRequestMemoryIfNotSet(*memory256Mi)
	spec.Container.setResourceLimitCPUIfNotSet(*cpu500m)
	spec.Container.setResourceRequestCPUIfNotSet(*cpu200m)

	licensingFullImageFromEnv := os.Getenv("OPERAND_LICENSING_IMAGE")

	// Check if operator image variable is set and CR has no overrides
	if licensingFullImageFromEnv != "" && spec.isImageEmpty() {
		err := spec.setImageParametersFromEnv(licensingFullImageFromEnv)
		if err != nil {
			return err
		}
	} else {
		// If CR has at least one override, make sure all parts of the image are filled at least with default values
		if spec.ImageRegistry == "" {
			spec.ImageRegistry = defaultQuayRegistry
		}
		if spec.ImageName == "" {
			spec.ImageName = defaultLicensingImageName
		}
		if spec.ImageTagPostfix == "" {
			spec.ImageTagPostfix = defaultLicensingImageTagPostfix
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

func (spec *IBMLicenseServiceReporterSpec) FillDefaultValues(reqLogger logr.Logger, r client_reader.Reader) error {
	databaseFullImageFromEnv := os.Getenv("OPERAND_REPORTER_DATABASE_IMAGE")
	// Check if operator image variable is set and CR has no overrides
	if databaseFullImageFromEnv != "" && spec.DatabaseContainer.isImageEmpty() {
		err := spec.DatabaseContainer.setImageParametersFromEnv(databaseFullImageFromEnv)
		if err != nil {
			return err
		}
	} else {
		// If CR has at least one override, make sure all parts of the image are filled at least with default values
		if spec.DatabaseContainer.ImageName == "" {
			spec.DatabaseContainer.ImageName = defaultDatabaseImageName
		}
		if spec.DatabaseContainer.ImageRegistry == "" {
			spec.DatabaseContainer.ImageRegistry = defaultQuayRegistry
		}
		if spec.DatabaseContainer.ImageTagPostfix == "" {
			spec.DatabaseContainer.ImageTagPostfix = defaultDatabaseImageTagPostfix
		}
	}

	receiverFullImageFromEnv := os.Getenv("OPERAND_REPORTER_RECEIVER_IMAGE")
	// Check if operator image variable is set and CR has no overrides
	if receiverFullImageFromEnv != "" && spec.ReceiverContainer.isImageEmpty() {
		err := spec.ReceiverContainer.setImageParametersFromEnv(receiverFullImageFromEnv)
		if err != nil {
			return err
		}
	} else {
		// If CR has at least one override, make sure all parts of the image are filled at least with default values
		if spec.ReceiverContainer.ImageName == "" {
			spec.ReceiverContainer.ImageName = defaultReporterImageName
		}
		if spec.ReceiverContainer.ImageRegistry == "" {
			spec.ReceiverContainer.ImageRegistry = defaultQuayRegistry
		}
		if spec.ReceiverContainer.ImageTagPostfix == "" {
			spec.ReceiverContainer.ImageTagPostfix = defaultReporterImageTagPostfix
		}
	}

	uiFullImageFromEnv := os.Getenv("OPERAND_REPORTER_UI_IMAGE")
	// Check if operator image variable is set and CR has no overrides
	if uiFullImageFromEnv != "" && spec.ReporterUIContainer.isImageEmpty() {
		err := spec.ReporterUIContainer.setImageParametersFromEnv(uiFullImageFromEnv)
		if err != nil {
			return err
		}
	} else {
		if spec.ReporterUIContainer.ImageName == "" {
			spec.ReporterUIContainer.ImageName = defaultReporterUIImageName
		}
		if spec.ReporterUIContainer.ImageRegistry == "" {
			spec.ReporterUIContainer.ImageRegistry = defaultQuayRegistry
		}
		if spec.ReporterUIContainer.ImageTagPostfix == "" {
			spec.ReporterUIContainer.ImageTagPostfix = defaultReporterUIImageTagPostfix
		}
	}

	spec.DatabaseContainer.initResourcesIfNil()
	spec.DatabaseContainer.setImagePullPolicyIfNotSet()
	spec.DatabaseContainer.setResourceLimitMemoryIfNotSet(*memory300Mi)
	spec.DatabaseContainer.setResourceRequestMemoryIfNotSet(*memory256Mi)
	spec.DatabaseContainer.setResourceLimitCPUIfNotSet(*cpu300m)
	spec.DatabaseContainer.setResourceRequestCPUIfNotSet(*cpu200m)

	spec.ReceiverContainer.initResourcesIfNil()
	spec.ReceiverContainer.setImagePullPolicyIfNotSet()
	spec.ReceiverContainer.setResourceLimitMemoryIfNotSet(*memory300Mi)
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
		spec.APISecretToken = "ibm-licensing-reporter-token"
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
