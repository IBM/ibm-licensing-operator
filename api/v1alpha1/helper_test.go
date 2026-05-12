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
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/utils/ptr"
)

func TestIsNodeCpuCappingEnabledNil(t *testing.T) {
	spec := &IBMLicensingSpec{}
	assert.True(t, spec.IsNodeCpuCappingEnabled(),
		"NodeCpuCappingEnabled unset should default to true.")
}

func TestIsNodeCpuCappingEnabledExplicitTrue(t *testing.T) {
	spec := &IBMLicensingSpec{NodeCpuCappingEnabled: ptr.To(true)}
	assert.True(t, spec.IsNodeCpuCappingEnabled(),
		"NodeCpuCappingEnabled=true should return true.")
}

func TestIsNodeCpuCappingEnabledExplicitFalse(t *testing.T) {
	spec := &IBMLicensingSpec{NodeCpuCappingEnabled: ptr.To(false)}
	assert.False(t, spec.IsNodeCpuCappingEnabled(),
		"NodeCpuCappingEnabled=false should return false.")
}
