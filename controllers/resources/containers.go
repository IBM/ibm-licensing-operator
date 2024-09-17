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
	corev1 "k8s.io/api/core/v1"

	operatorv1alpha1 "github.com/IBM/ibm-licensing-operator/api/v1alpha1"
)

func GetSecurityContext() *corev1.SecurityContext {
	procMount := corev1.DefaultProcMount
	securityContext := &corev1.SecurityContext{
		AllowPrivilegeEscalation: &FalseVar,
		Privileged:               &FalseVar,
		ReadOnlyRootFilesystem:   &TrueVar,
		RunAsNonRoot:             &TrueVar,
		Capabilities: &corev1.Capabilities{
			Drop: []corev1.Capability{
				"ALL",
			},
		},
		SeccompProfile: &corev1.SeccompProfile{
			Type: corev1.SeccompProfileTypeRuntimeDefault,
		},
		ProcMount: &procMount,
	}
	return securityContext
}

func GetReadinessProbe(probeHandler corev1.ProbeHandler) *corev1.Probe {
	return &corev1.Probe{
		ProbeHandler:        probeHandler,
		InitialDelaySeconds: 60,
		TimeoutSeconds:      10,
		PeriodSeconds:       60,
	}
}

func GetLivenessProbe(probeHandler corev1.ProbeHandler) *corev1.Probe {
	return &corev1.Probe{
		ProbeHandler:        probeHandler,
		InitialDelaySeconds: 120,
		TimeoutSeconds:      10,
		PeriodSeconds:       300,
	}
}

func GetContainerBase(container operatorv1alpha1.Container) corev1.Container {
	return corev1.Container{
		Image:           container.GetFullImage(),
		ImagePullPolicy: container.ImagePullPolicy,
		SecurityContext: GetSecurityContext(),
		Resources:       container.Resources,
	}
}
