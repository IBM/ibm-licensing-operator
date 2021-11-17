//
// Copyright 2021 IBM Corporation
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
	"github.com/ibm/ibm-licensing-operator/api/v1alpha1/features"
)

type Features struct {
	// Configure if you have HyperThreading (HT) or Symmetrical Multi-Threading (SMT) enabled
	// +optional
	HyperThreading *features.HyperThreading `json:"hyperThreading,omitempty"`

	//Enable legacy Auth
	//+optional
	LegacyAuthEnabled *bool `json:"legacyAuthEnabled,omitempty"`
}

func (spec *IBMLicensingSpec) HaveFeatures() bool {
	return spec.Features != nil
}

func (spec *IBMLicensingSpec) IsHyperThreadingEnabled() bool {
	return spec.HaveFeatures() && spec.Features.HyperThreading != nil
}

func (spec *IBMLicensingSpec) IsLegacyAuthEnabled() bool {
	if spec.HaveFeatures() && spec.Features.LegacyAuthEnabled != nil && !*spec.Features.LegacyAuthEnabled {
		return false
	} else {
		return true
	}
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
