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

package reporter

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	operatorv1alpha1 "github.com/IBM/ibm-licensing-operator/api/v1alpha1"
	"github.com/IBM/ibm-licensing-operator/controllers/resources"
)

func GetLicenseReporterInitContainers(instance *operatorv1alpha1.IBMLicenseServiceReporter) []corev1.Container {
	containers := []corev1.Container{}
	if resources.IsServiceCAAPI && instance.Spec.HTTPSCertsSource == operatorv1alpha1.OcpCertsSource {
		baseContainer := GetReceiverContainer(instance)
		baseContainer.LivenessProbe = nil
		baseContainer.ReadinessProbe = nil
		ocpSecretCheckContainer := corev1.Container{}
		baseContainer.DeepCopyInto(&ocpSecretCheckContainer)
		ocpSecretCheckContainer.Name = resources.OcpCheckString
		ocpSecretCheckContainer.Command = []string{
			"sh",
			"-c",
			resources.GetOCPSecretCheckScript(),
		}
		containers = append(containers, ocpSecretCheckContainer)
	}
	return containers
}

func getReceiverProbeHandler() corev1.ProbeHandler {
	return corev1.ProbeHandler{
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
	container := resources.GetContainerBase(instance.Spec.ReceiverContainer)
	container.Env = getReciverEnvVariables(instance.Spec)
	container.VolumeMounts = getReceiverVolumeMounts()
	container.Name = ReceiverContainerName
	container.Ports = []corev1.ContainerPort{
		{
			ContainerPort: ReceiverPort,
			Protocol:      corev1.ProtocolTCP,
		},
	}
	container.LivenessProbe = resources.GetLivenessProbe(getReceiverProbeHandler())
	container.ReadinessProbe = resources.GetReadinessProbe(getReceiverProbeHandler())
	return container
}
