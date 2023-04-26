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

type License struct {
	// Accept license terms of usage: https://ibm.biz/icpfs39license.
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="License acceptance",xDescriptors="urn:alm:descriptor:com.tectonic.ui:checkbox"
	Accept *bool `json:"accept,omitempty"`
}

func (spec *IBMLicensingSpec) IsLicenseAccepted() bool {
	return spec.License != nil && spec.License.Accept != nil && *spec.License.Accept
}
