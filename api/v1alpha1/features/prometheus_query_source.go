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

package features

// +k8s:deepcopy-gen=true
type PrometheusQuerySource struct {
	// Should this function be enabled (by default it is).
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Enabled",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// What url to use for prometheus API (by default use OCP Thanos Querier).
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="URL",xDescriptors="urn:alm:descriptor:com.tectonic.ui:hidden"
	// +optional
	URL string `json:"url,omitempty"`
}
