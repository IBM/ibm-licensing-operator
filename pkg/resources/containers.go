//
// Copyright 2020 IBM Corporation
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
	"strconv"

	operatorv1alpha1 "github.com/ibm/ibm-licensing-operator/pkg/apis/operator/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func getLicensingSecurityContext(spec operatorv1alpha1.IBMLicensingSpec) *corev1.SecurityContext {
	procMount := corev1.DefaultProcMount
	securityContext := &corev1.SecurityContext{
		AllowPrivilegeEscalation: &FalseVar,
		Privileged:               &FalseVar,
		ReadOnlyRootFilesystem:   &FalseVar,
		RunAsNonRoot:             &TrueVar,
		Capabilities: &corev1.Capabilities{
			Drop: []corev1.Capability{
				"ALL",
			},
		},
		ProcMount: &procMount,
	}
	if spec.SecurityContext.RunAsUser != 0 {
		securityContext.RunAsUser = &spec.SecurityContext.RunAsUser
	}
	return securityContext
}

func getLicensingEnvironmentVariables(spec operatorv1alpha1.IBMLicensingSpec) []corev1.EnvVar {
	var httpsEnableString = strconv.FormatBool(spec.HTTPSEnable)
	var environmentVariables = []corev1.EnvVar{
		{
			Name:  "NAMESPACE",
			Value: spec.InstanceNamespace,
		},
		{
			Name:  "DATASOURCE",
			Value: spec.Datasource,
		},
		{
			Name:  "HTTPS_ENABLE",
			Value: httpsEnableString,
		},
	}
	if spec.IsDebug() {
		environmentVariables = append(environmentVariables, []corev1.EnvVar{
			{
				Name:  "logging.level.com.ibm",
				Value: "DEBUG",
			},
		}...)
	}
	if spec.HTTPSEnable {
		environmentVariables = append(environmentVariables, []corev1.EnvVar{
			{
				Name:  "HTTPS_CERTS_SOURCE",
				Value: spec.HTTPSCertsSource,
			},
		}...)
	}
	if spec.IsMetering() {
		environmentVariables = append(environmentVariables, []corev1.EnvVar{
			{
				Name:  "METERING_URL",
				Value: "https://metering-server:4002/api/v1/metricData",
			},
		}...)
	}
	return environmentVariables
}

func getProbeScheme(spec operatorv1alpha1.IBMLicensingSpec) corev1.URIScheme {
	if spec.HTTPSEnable {
		return "HTTPS"
	}
	return ""
}

func getProbeHandler(spec operatorv1alpha1.IBMLicensingSpec) corev1.Handler {
	var probeScheme = getProbeScheme(spec)
	return corev1.Handler{
		HTTPGet: &corev1.HTTPGetAction{
			Path: "/",
			Port: intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: licensingContainerPort,
			},
			Scheme: probeScheme,
		},
	}
}

func GetLicensingContainer(spec operatorv1alpha1.IBMLicensingSpec) corev1.Container {
	var probeHandler = getProbeHandler(spec)
	return corev1.Container{
		Image:           spec.GetFullImage(),
		Name:            "license-service",
		ImagePullPolicy: corev1.PullAlways,
		VolumeMounts:    getLicensingVolumeMounts(spec),
		Env:             getLicensingEnvironmentVariables(spec),
		Ports: []corev1.ContainerPort{
			{
				ContainerPort: licensingContainerPort,
				Protocol:      corev1.ProtocolTCP,
			},
		},
		LivenessProbe: &corev1.Probe{
			Handler:             probeHandler,
			InitialDelaySeconds: 120,
			TimeoutSeconds:      10,
			PeriodSeconds:       300,
			// SuccessThreshold:    1,
			// FailureThreshold:    3,
		},
		ReadinessProbe: &corev1.Probe{
			Handler:             probeHandler,
			InitialDelaySeconds: 60,
			TimeoutSeconds:      10,
			PeriodSeconds:       60,
			// SuccessThreshold:    1,
			// FailureThreshold:    3,
		},
		Resources: corev1.ResourceRequirements{
			Limits: map[corev1.ResourceName]resource.Quantity{
				corev1.ResourceCPU:    *cpu500m,
				corev1.ResourceMemory: *memory512Mi},
			Requests: map[corev1.ResourceName]resource.Quantity{
				corev1.ResourceCPU:    *cpu200m,
				corev1.ResourceMemory: *memory256Mi},
		},
		SecurityContext: getLicensingSecurityContext(spec),
	}
}
