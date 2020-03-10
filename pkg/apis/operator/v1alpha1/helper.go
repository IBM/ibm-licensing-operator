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

func (spec *IBMLicensingSpec) IsMetering() bool {
	return spec.Datasource == "metering"
}

func (spec *IBMLicensingSpec) IsDebug() bool {
	return spec.LogLevel == "DEBUG"
}

func (spec *IBMLicensingSpec) GetFullImage() string {
	return spec.ImageRegistry + "/" + spec.ImageName + ":" + spec.ImageTagPostfix
}

func (spec *IBMLicensingSpec) FillDefaultValues(isOpenshiftCluster bool) {
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
	if spec.ImageRegistry == "" {
		spec.ImageRegistry = "quay.io/opencloudio"
	}
	if spec.ImageName == "" {
		spec.ImageName = "ibm-licensing"
	}
	if spec.ImageTagPostfix == "" {
		spec.ImageTagPostfix = "1.0.0"
	}
	if spec.APISecretToken == "" {
		spec.APISecretToken = "ibm-licensing-token"
	}
}

func (spec *IBMLicensingSpec) IsRouteEnabled() bool {
	return spec.RouteEnabled != nil && *spec.RouteEnabled
}

func (spec *IBMLicensingSpec) IsIngressEnabled() bool {
	return spec.IngressEnabled != nil && *spec.IngressEnabled
}
