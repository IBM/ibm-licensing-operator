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
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	client_reader "sigs.k8s.io/controller-runtime/pkg/client"
)

const defaultImageRegistry = "quay.io/opencloudio"
const defaultLicensingImageName = "ibm-licensing"
const defaultLicensingImageTagPostfix = "1.2.0"

const defaultReceiverImageName = "bm-license-advisor-receiver"
const defaultDatabaseImageName = "ibm-license-advisor-db"
const defaultReceiverImageRegistry = "quay.io/opencloudio"
const defaultDatabaseImageRegistry = "quay.io/opencloudio"
const defaultReceiverImageTagPostfix = "1.2.0"
const defaultDatabaseImageTagPostfix = "1.2.0"

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

	// Resources and limits for container
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

type IBMLicenseServiceBaseSpec struct {
	// Should application pod show additional information, options: DEBUG, INFO
	// +kubebuilder:validation:Enum=DEBUG;INFO
	LogLevel string `json:"logLevel,omitempty"`
	// Secret name used to store application token, either one that exists, or one that will be created, for now only one value possible
	// +kubebuilder:validation:Enum=ibm-licensing-token
	APISecretToken string `json:"apiSecretToken,omitempty"`
	// Array of pull secrets which should include existing at InstanceNamespace secret to allow pulling IBM Licensing image
	ImagePullSecrets []string `json:"imagePullSecrets,omitempty"`
	// IBM License Service Pod pull policy, default: IfNotPresent
	// +kubebuilder:validation:Enum=Always;IfNotPresent;Never
	ImagePullPolicy string `json:"imagePullPolicy,omitempty"`
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

func (spec *IBMLicensingSpec) GetFullImage() string {
	// If there is ":" in image tag then we use "@" for digest as only digest can have it
	if strings.ContainsAny(spec.ImageTagPostfix, ":") {
		return spec.ImageRegistry + "/" + spec.ImageName + "@" + spec.ImageTagPostfix
	}
	return spec.ImageRegistry + "/" + spec.ImageName + ":" + spec.ImageTagPostfix
}

// IsImageEmpty returns true when any part of image name is not defined
func (spec *IBMLicensingSpec) IsImageEmpty() bool {
	return spec.ImageRegistry == "" && spec.ImageName == "" && spec.ImageTagPostfix == ""
}

// setImageParametersFromEnv set container image info from full image reference
func (spec *IBMLicensingSpec) setImageParametersFromEnv(fullImageName string) error {
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
	spec.ImageTagPostfix = imageWithTagSplitted[1]
	spec.ImageName = imageWithTagSplitted[0]
	spec.ImageRegistry = strings.Join(imagePathSplitted[:len(imagePathSplitted)-1], "/")
	return nil
}

func (spec *IBMLicensingSpec) FillDefaultValues(isOpenshiftCluster bool) error {
	if spec.ImagePullPolicy == "" {
		spec.ImagePullPolicy = "IfNotPresent"
	}
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

	if spec.Resources.Limits.Cpu().IsZero() || spec.Resources.Requests.Cpu().IsZero() ||
		spec.Resources.Limits.Memory().IsZero() || spec.Resources.Requests.Memory().IsZero() {
		spec.Resources = corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    *cpu500m,
				corev1.ResourceMemory: *memory512Mi,
			},
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    *cpu200m,
				corev1.ResourceMemory: *memory256Mi,
			},
		}
	}

	licensingFullImageFromEnv := os.Getenv("OPERAND_LICENSING_IMAGE")

	// Check if operator image variable is set and CR has no overrides
	if licensingFullImageFromEnv != "" && spec.IsImageEmpty() {
		err := spec.setImageParametersFromEnv(licensingFullImageFromEnv)
		if err != nil {
			return err
		}
	} else {
		// If CR has at least one override, make sure all parts of the image are filled at least with default values
		if spec.ImageRegistry == "" {
			spec.ImageRegistry = defaultImageRegistry
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
	if spec.DatabaseContainer.ImageName == "" {
		spec.DatabaseContainer.ImageName = defaultDatabaseImageName
	}
	if spec.DatabaseContainer.ImageRegistry == "" {
		spec.DatabaseContainer.ImageRegistry = defaultDatabaseImageRegistry
	}
	if spec.DatabaseContainer.ImageTagPostfix == "" {
		spec.DatabaseContainer.ImageTagPostfix = defaultDatabaseImageTagPostfix
	}

	if spec.ReceiverContainer.ImageName == "" {
		spec.ReceiverContainer.ImageName = defaultReceiverImageName
	}
	if spec.ReceiverContainer.ImageRegistry == "" {
		spec.ReceiverContainer.ImageRegistry = defaultReceiverImageRegistry
	}
	if spec.ReceiverContainer.ImageTagPostfix == "" {
		spec.ReceiverContainer.ImageTagPostfix = defaultReceiverImageTagPostfix
	}

	if spec.DatabaseContainer.Resources.Limits.Cpu().IsZero() || spec.DatabaseContainer.Resources.Requests.Cpu().IsZero() ||
		spec.DatabaseContainer.Resources.Limits.Memory().IsZero() || spec.DatabaseContainer.Resources.Requests.Memory().IsZero() {
		spec.DatabaseContainer.Resources = corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    *cpu300m,
				corev1.ResourceMemory: *memory300Mi,
			},
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    *cpu200m,
				corev1.ResourceMemory: *memory256Mi,
			},
		}
	}

	if spec.ReceiverContainer.Resources.Limits.Cpu().IsZero() || spec.ReceiverContainer.Resources.Requests.Cpu().IsZero() ||
		spec.ReceiverContainer.Resources.Limits.Memory().IsZero() || spec.ReceiverContainer.Resources.Requests.Memory().IsZero() {
		spec.ReceiverContainer.Resources = corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    *cpu300m,
				corev1.ResourceMemory: *memory300Mi,
			},
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    *cpu200m,
				corev1.ResourceMemory: *memory256Mi,
			},
		}
	}

	if spec.Capacity.IsZero() {
		spec.Capacity = *size1Gi
	}

	if spec.APISecretToken == "" {
		spec.APISecretToken = "ibm-licensing-hub-token"
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
