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
	"errors"
	"os"
	"strings"
)

const defaultImageRegistry = "quay.io/opencloudio"
const defaultLicensingImageName = "ibm-licensing"
const defaultLicensingImageTagPostfix = "1.2.0"

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
