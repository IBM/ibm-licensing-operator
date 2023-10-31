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

	"github.com/IBM/ibm-licensing-operator/testutils"
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

func TestEqualVolumes(t *testing.T) {
	type args struct {
		volumes1 []corev1.Volume
		volumes2 []corev1.Volume
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"Equal volumes - empty", args{volumes1: []corev1.Volume{}, volumes2: []corev1.Volume{}}, true},
		{"Equal volumes", args{volumes1: testutils.Volumes1, volumes2: testutils.Volumes1}, true},
		{"Equal volumes - different order", args{volumes1: testutils.Volumes1, volumes2: testutils.Volumes1DiffOrder}, true},
		{"Not equal volumes - different volume", args{volumes1: testutils.Volumes1, volumes2: testutils.Volumes3}, false},
		{"Not equal volumes - missing volume", args{volumes1: testutils.Volumes1, volumes2: testutils.Volumes3}, false},
		{"Not equal volumes - additional volume", args{volumes1: testutils.Volumes1, volumes2: testutils.Volumes1AdditionalVolume}, false},
		{"Not equal volumes - different emptyDir size limit", args{volumes1: testutils.Volumes1, volumes2: testutils.Volumes1DiffEmptyDirSize}, false},
		{"Not equal volumes - deep diff", args{volumes1: testutils.Volumes1, volumes2: testutils.Volumes4}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := equalVolumes(tt.args.volumes1, tt.args.volumes2); got != tt.want {
				t.Errorf("equalVolumes() = %v, want %v", got, tt.want)
			}
		})
	}
}
