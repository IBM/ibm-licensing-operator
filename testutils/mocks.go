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

package testutils

import (
	corev1 "k8s.io/api/core/v1"
)

var Envs1 = []corev1.EnvVar{
	{Name: "env1", Value: "val1"},
	{Name: "env2", Value: "env2"},
}

var Envs2 = []corev1.EnvVar{
	{Name: "env1", Value: "val1"},
	{Name: "env2", Value: "env2"},
	{Name: "env3", Value: "env3"},
}

var Envs3 = []corev1.EnvVar{
	{Name: "env", Value: "val"},
}

var Envs5 = []corev1.EnvVar{
	{Name: "env1", Value: "val3"},
	{Name: "env2", Value: "env2"},
}

var Envs6 = []corev1.EnvVar{
	{Name: "env1", Value: "val1"},
	{Name: "env2", Value: "env2", ValueFrom: &corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "secret-a"}}}},
}

var Envs7 = []corev1.EnvVar{
	{Name: "env1", Value: "val1"},
	{Name: "env2", Value: "env2", ValueFrom: &corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "secret-b"}}}},
}
