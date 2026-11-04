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

	"github.com/go-logr/logr"
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

func TestShouldUpdateDeploymentTolerationsAndNodeSelector(t *testing.T) {
	ibmDefaults := []corev1.Toleration{
		{Key: "dedicated", Operator: corev1.TolerationOpExists, Effect: corev1.TaintEffectNoSchedule},
		{Key: "CriticalAddonsOnly", Operator: corev1.TolerationOpExists},
	}
	userTolerations := []corev1.Toleration{
		{Key: "ils-only", Operator: corev1.TolerationOpExists, Effect: corev1.TaintEffectNoSchedule},
	}
	changedTolerations := []corev1.Toleration{
		{Key: "ils-only", Operator: corev1.TolerationOpEqual, Value: "true", Effect: corev1.TaintEffectNoSchedule},
	}

	tests := []struct {
		name             string
		foundTolerations []corev1.Toleration
		wantTolerations  []corev1.Toleration
		foundSelector    map[string]string
		wantSelector     map[string]string
		wantUpdate       bool
	}{
		{"nothing changed", ibmDefaults, ibmDefaults, nil, nil, false},
		{"same user tolerations", userTolerations, userTolerations, nil, nil, false},
		{"tolerations added", ibmDefaults, userTolerations, nil, nil, true},
		{"tolerations changed", userTolerations, changedTolerations, nil, nil, true},
		{"tolerations removed, back to the defaults", userTolerations, ibmDefaults, nil, nil, true},
		{"same node selector", nil, nil, map[string]string{"zone": "a"}, map[string]string{"zone": "a"}, false},
		{"node selector added", nil, nil, nil, map[string]string{"zone": "a"}, true},
		{"node selector changed", nil, nil, map[string]string{"zone": "a"}, map[string]string{"zone": "b"}, true},
		{"node selector removed", nil, nil, map[string]string{"zone": "a"}, nil, true},
		{"empty and nil node selectors are equal", nil, nil, map[string]string{}, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := logr.Discard()
			found := &corev1.PodTemplateSpec{Spec: corev1.PodSpec{
				Tolerations:  tt.foundTolerations,
				NodeSelector: tt.foundSelector,
			}}
			expected := &corev1.PodTemplateSpec{Spec: corev1.PodSpec{
				Tolerations:  tt.wantTolerations,
				NodeSelector: tt.wantSelector,
			}}

			if got := ShouldUpdateDeployment(&logger, expected, found); got != tt.wantUpdate {
				t.Errorf("ShouldUpdateDeployment() = %v, want %v", got, tt.wantUpdate)
			}
		})
	}
}
