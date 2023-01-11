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
	"github.com/IBM/ibm-licensing-operator/api/v1alpha1/features"
)

type Features struct {
	// Configure if you have HyperThreading (HT) or Symmetrical Multi-Threading (SMT) enabled
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Hyper Threading",xDescriptors="urn:alm:descriptor:com.tectonic.ui:hidden"
	// +optional
	HyperThreading *features.HyperThreading `json:"hyperThreading,omitempty"`

	// Authorization settings.
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Auth",xDescriptors="urn:alm:descriptor:com.tectonic.ui:hidden"
	// +optional
	Auth *features.Auth `json:"auth,omitempty"`

	// Change prometheus query source settings.
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Prometheus query source settings",xDescriptors="urn:alm:descriptor:com.tectonic.ui:hidden"
	// +optional
	PrometheusQuerySource *features.PrometheusQuerySource `json:"prometheusQuerySource,omitempty"`

	// Change alerting settings.
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Alerting settings",xDescriptors="urn:alm:descriptor:com.tectonic.ui:hidden"
	// +optional
	Alerting *features.Alerting `json:"alerting,omitempty"`

	// Special terms, must be granted by IBM Pricing.
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Namespace scope enabled",xDescriptors="urn:alm:descriptor:com.tectonic.ui:hidden"
	// +optional
	NamespaceScopeEnabled *bool `json:"nssEnabled,omitempty"`
}

func (spec *IBMLicensingSpec) HaveFeatures() bool {
	return spec.Features != nil
}

func (spec *IBMLicensingSpec) IsNamespaceScopeEnabled() bool {
	return spec.HaveFeatures() && spec.Features.NamespaceScopeEnabled != nil && *spec.Features.NamespaceScopeEnabled
}

func (spec *IBMLicensingSpec) IsHyperThreadingEnabled() bool {
	return spec.HaveFeatures() && spec.Features.HyperThreading != nil
}

func (spec *IBMLicensingSpec) IsURLBasedAuthEnabled() bool {
	if spec.HaveFeatures() && spec.Features.Auth != nil && !spec.Features.Auth.URLBasedEnabled {
		return false
	}
	return true
}

func (spec *IBMLicensingSpec) IsAlertingEnabled() bool {
	// this is also set by default when filling default values during reconciliation
	// return true only and only if the value is set to true
	if spec.HaveFeatures() && spec.Features.Alerting != nil && spec.Features.Alerting.Enabled != nil {
		return *spec.Features.Alerting.Enabled
	}
	return false
}

func (spec *IBMLicensingSpec) IsPrometheusQuerySourceEnabled() bool {
	// return false only and only if the value is set to false
	if spec.HaveFeatures() && spec.Features.PrometheusQuerySource != nil &&
		spec.Features.PrometheusQuerySource.Enabled != nil &&
		!*spec.Features.PrometheusQuerySource.Enabled {
		return false
	}
	return true
}

func (spec *IBMLicensingSpec) GetPrometheusQuerySourceURL() string {
	if spec.HaveFeatures() && spec.Features.PrometheusQuerySource != nil {
		return spec.Features.PrometheusQuerySource.URL
	}
	return ""
}

func (spec *IBMLicensingSpec) GetHyperThreadingThreadsPerCoreOrNil() *int {
	if !spec.IsHyperThreadingEnabled() {
		return nil
	}
	threadsPerCore := spec.Features.HyperThreading.ThreadsPerCore
	// threadsPerCore works as a multiplier so when it is 1 we ignore it
	if threadsPerCore == 1 {
		return nil
	}
	return &threadsPerCore
}
