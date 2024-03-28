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

package resources

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/IBM/ibm-licensing-operator/testutils"
	"github.com/stretchr/testify/assert"
)

func TestEqualsEnvVars(t *testing.T) {
	type args struct {
		envVars1 []corev1.EnvVar
		envVars2 []corev1.EnvVar
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"Equal env vars - empty", args{envVars1: []corev1.EnvVar{}, envVars2: []corev1.EnvVar{}}, true},
		{"Equal env vars", args{envVars1: testutils.Envs1, envVars2: testutils.Envs1}, true},
		{"Equal vars - with value refs", args{envVars1: testutils.Envs6, envVars2: testutils.Envs6}, true},
		{"Not equal env vars - different lengths", args{envVars1: testutils.Envs1, envVars2: testutils.Envs2}, false},
		{"Not equal env vars - different values", args{envVars1: testutils.Envs1, envVars2: testutils.Envs5}, false},
		{"Not equal env vars", args{envVars1: testutils.Envs1, envVars2: testutils.Envs3}, false},
		{"Not equal env vars - missing value ref", args{envVars1: testutils.Envs1, envVars2: testutils.Envs6}, false},
		{"Not equal env vars - different value ref", args{envVars1: testutils.Envs6, envVars2: testutils.Envs7}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := equalEnvVars(tt.args.envVars1, tt.args.envVars2); got != tt.want {
				t.Errorf("equalEnvVars() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEqualResources(t *testing.T) {

	const hugePages2Mi = corev1.ResourceHugePagesPrefix + "2Mi"

	tests := []struct {
		name     string
		expected corev1.ResourceList
		actual   corev1.ResourceList
		equal    bool
	}{
		{
			name:  "both resource lists are empty",
			equal: true,
		},
		{
			name:     "equal cpu limits",
			expected: corev1.ResourceList{corev1.ResourceLimitsCPU: resource.MustParse("100m")},
			actual:   corev1.ResourceList{corev1.ResourceLimitsCPU: resource.MustParse("100m")},
			equal:    true,
		},
		{
			name:     "not equal cpu limits",
			expected: corev1.ResourceList{corev1.ResourceLimitsCPU: resource.MustParse("100m")},
			actual:   corev1.ResourceList{corev1.ResourceLimitsCPU: resource.MustParse("200m")},
			equal:    false,
		},
		{
			name:     "not equal cpu limits - missing actual",
			expected: corev1.ResourceList{corev1.ResourceLimitsCPU: resource.MustParse("100m")},
			actual:   nil,
			equal:    false,
		},
		{
			name:     "equal memory requests",
			expected: corev1.ResourceList{corev1.ResourceRequestsMemory: resource.MustParse("100Mi")},
			actual:   corev1.ResourceList{corev1.ResourceRequestsMemory: resource.MustParse("100Mi")},
			equal:    true,
		},
		{
			name:     "not equal memory requests",
			expected: corev1.ResourceList{corev1.ResourceRequestsMemory: resource.MustParse("100Mi")},
			actual:   corev1.ResourceList{corev1.ResourceRequestsMemory: resource.MustParse("200Mi")},
			equal:    false,
		},
		{
			name:     "not equal memory requests - missing expected",
			expected: nil,
			actual:   corev1.ResourceList{corev1.ResourceRequestsMemory: resource.MustParse("200Mi")},
			equal:    false,
		},
		{
			name:     "equal hugepages",
			expected: corev1.ResourceList{hugePages2Mi: resource.MustParse("100Mi")},
			actual:   corev1.ResourceList{hugePages2Mi: resource.MustParse("100Mi")},
			equal:    true,
		},
		{
			name:     "not equal hugepages - missing actual",
			expected: corev1.ResourceList{hugePages2Mi: resource.MustParse("100Mi")},
			actual:   nil,
			equal:    false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			equal := equalResources(test.expected, test.actual)
			assert.Equal(t, test.equal, equal)
		})
	}
}
