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

package reporter

import (
	"strconv"

	operatorv1alpha1 "github.com/ibm/ibm-licensing-operator/pkg/apis/operator/v1alpha1"
	res "github.com/ibm/ibm-licensing-operator/pkg/resources"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func getReporterUIEnvironmentVariables() []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name: "WLP_CLIENT_ID",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "platform-oidc-credentials",
					},
					Key: "WLP_CLIENT_ID",
				},
			},
		},
		{
			Name: "WLP_CLIENT_SECRET",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "platform-oidc-credentials",
					},
					Key: "WLP_CLIENT_SECRET",
				},
			},
		},
		{
			Name:  "HTTP_PORT",
			Value: strconv.Itoa(UIPort),
		},
		{
			Name:  "cfcRouterUrl",
			Value: "https://icp-management-ingress",
		},
		{
			Name:  "PLATFORM_IDENTITY_PROVIDER_URL",
			Value: "https://icp-management-ingress/idprovider",
		},
	}

}

func getReporterUIProbeHandler() corev1.Handler {
	return corev1.Handler{
		HTTPGet: &corev1.HTTPGetAction{
			Path: "/license-advisor/version.txt",
			Port: intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: reporterUIServicePort.IntVal,
			},
			Scheme: "HTTP",
		},
	}
}

func GetReporterUIContainer(instance *operatorv1alpha1.IBMLicenseServiceReporter) corev1.Container {
	container := res.GetContainerBase(instance.Spec.ReporterUIContainer)
	container.ImagePullPolicy = corev1.PullAlways
	container.Env = getReporterUIEnvironmentVariables()
	container.Resources = instance.Spec.ReporterUIContainer.Resources
	container.Name = UIContainerName
	container.Ports = []corev1.ContainerPort{
		{
			ContainerPort: UIPort,
			Protocol:      corev1.ProtocolTCP,
		},
	}
	container.LivenessProbe = res.GetLivenessProbe(getReporterUIProbeHandler())
	container.ReadinessProbe = res.GetReadinessProbe(getReporterUIProbeHandler())
	return container
}
