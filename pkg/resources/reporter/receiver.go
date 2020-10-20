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
	operatorv1alpha1 "github.com/ibm/ibm-licensing-operator/pkg/apis/operator/v1alpha1"
	res "github.com/ibm/ibm-licensing-operator/pkg/resources"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func GetLicenseReporterInitContainers(instance *operatorv1alpha1.IBMLicenseServiceReporter) []corev1.Container {
	containers := []corev1.Container{}
	if res.IsOCPCertManagerAPI() && instance.Spec.HTTPSCertsSource == operatorv1alpha1.OcpCertsSource {
		baseContainer := GetReceiverContainer(instance)
		baseContainer.LivenessProbe = nil
		baseContainer.ReadinessProbe = nil
		ocpSecretCheckContainer := corev1.Container{}
		baseContainer.DeepCopyInto(&ocpSecretCheckContainer)
		ocpSecretCheckContainer.Name = res.OcpCheckString
		ocpSecretCheckContainer.Command = []string{
			"sh",
			"-c",
			res.GetOCPSecretCheckScript(),
		}
		containers = append(containers, ocpSecretCheckContainer)
	}
	return containers
}

func getReceiverProbeHandler() corev1.Handler {
	return corev1.Handler{
		HTTPGet: &corev1.HTTPGetAction{
			Path: "/",
			Port: intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: receiverServicePort.IntVal,
			},
			Scheme: "HTTPS",
		},
	}
}

func GetReceiverContainer(instance *operatorv1alpha1.IBMLicenseServiceReporter) corev1.Container {
	container := res.GetContainerBase(instance.Spec.ReceiverContainer)
	container.Env = []corev1.EnvVar{
		{
			Name:  "HTTPS_CERTS_SOURCE",
			Value: string(instance.Spec.HTTPSCertsSource),
		},
	}
	container.EnvFrom = getDatabaseEnvFromSourceVariables()
	container.VolumeMounts = getVolumeMounts(instance.Spec)
	container.Name = ReceiverContainerName
	container.Ports = []corev1.ContainerPort{
		{
			ContainerPort: ReceiverPort,
			Protocol:      corev1.ProtocolTCP,
		},
	}
	container.LivenessProbe = res.GetLivenessProbe(getReceiverProbeHandler())
	container.ReadinessProbe = res.GetReadinessProbe(getReceiverProbeHandler())
	return container
}
