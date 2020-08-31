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

package service

import (
	"strconv"

	operatorv1alpha1 "github.com/ibm/ibm-licensing-operator/pkg/apis/operator/v1alpha1"
	res "github.com/ibm/ibm-licensing-operator/pkg/resources"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func getLicensingEnvironmentVariables(instance *operatorv1alpha1.IBMLicenseService) []corev1.EnvVar {
	var httpsEnableString = strconv.FormatBool(instance.Spec.HTTPSEnable)
	var environmentVariables = []corev1.EnvVar{
		{
			Name:  "NAMESPACE",
			Value: instance.GetNamespace(),
		},
		{
			Name:  "DATASOURCE",
			Value: instance.Spec.Datasource,
		},
		{
			Name:  "HTTPS_ENABLE",
			Value: httpsEnableString,
		},
	}
	if instance.Spec.IsDebug() {
		environmentVariables = append(environmentVariables, corev1.EnvVar{
			Name:  "logging.level.com.ibm",
			Value: "DEBUG",
		})
	}
	if instance.Spec.HTTPSEnable {
		environmentVariables = append(environmentVariables, corev1.EnvVar{
			Name:  "HTTPS_CERTS_SOURCE",
			Value: instance.Spec.HTTPSCertsSource,
		})
	}
	if instance.Spec.IsMetering() {
		environmentVariables = append(environmentVariables, corev1.EnvVar{
			Name:  "METERING_URL",
			Value: "https://metering-server:4002/api/v1/metricData",
		})
	}
	if instance.Spec.Sender != nil {
		environmentVariables = append(environmentVariables, []corev1.EnvVar{
			{
				Name:  "CLUSTER_ID",
				Value: instance.Spec.Sender.ClusterID,
			},
			{
				Name: "HUB_TOKEN",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: instance.Spec.Sender.ReporterSecretToken,
						},
						Key: ReporterSecretTokenKeyName,
					},
				},
			},
			{
				Name:  "HUB_URL",
				Value: instance.Spec.Sender.ReporterURL,
			},
			{
				Name:  "CLUSTER_NAME",
				Value: instance.Spec.Sender.ClusterName,
			},
		}...)
	}
	return environmentVariables
}

func getProbeScheme(spec operatorv1alpha1.IBMLicenseServiceSpec) corev1.URIScheme {
	if spec.HTTPSEnable {
		return "HTTPS"
	}
	return ""
}

func getProbeHandler(spec operatorv1alpha1.IBMLicenseServiceSpec) corev1.Handler {
	var probeScheme = getProbeScheme(spec)
	return corev1.Handler{
		HTTPGet: &corev1.HTTPGetAction{
			Path: "/",
			Port: intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: licensingServicePort.IntVal,
			},
			Scheme: probeScheme,
		},
	}
}

func getMeteringSecretCheckScript() string {
	script := `while true; do
  echo "$(date): Checking for metering secret"
  ls /opt/metering/certs/* && break
  echo "$(date): Required metering secret not found ... try again in 30s"
  sleep 30
done
echo "$(date): All required secrets exist"
`
	return script
}

func GetLicensingInitContainers(instance *operatorv1alpha1.IBMLicenseService) []corev1.Container {
	if !instance.Spec.IsMetering() {
		return nil
	}
	baseContainer := getLicensingContainerBase(instance)
	meteringSecretCheckContainer := corev1.Container{}
	baseContainer.DeepCopyInto(&meteringSecretCheckContainer)
	meteringSecretCheckContainer.Name = "metering-check-secret"
	meteringSecretCheckContainer.Command = []string{
		"sh",
		"-c",
		getMeteringSecretCheckScript(),
	}
	return []corev1.Container{
		meteringSecretCheckContainer,
	}
}

func getLicensingContainerBase(instance *operatorv1alpha1.IBMLicenseService) corev1.Container {
	container := res.GetContainerBase(instance.Spec.Container)
	if instance.Spec.SecurityContext != nil {
		container.SecurityContext.RunAsUser = &instance.Spec.SecurityContext.RunAsUser
	}
	container.VolumeMounts = getLicensingVolumeMounts(instance.Spec)
	container.Env = getLicensingEnvironmentVariables(instance)
	container.Ports = []corev1.ContainerPort{
		{
			ContainerPort: licensingServicePort.IntVal,
			Protocol:      corev1.ProtocolTCP,
		},
	}
	return container
}

func GetLicensingContainer(instance *operatorv1alpha1.IBMLicenseService) corev1.Container {
	container := getLicensingContainerBase(instance)
	probeHandler := getProbeHandler(instance.Spec)
	container.Name = "license-service"
	container.LivenessProbe = res.GetLivenessProbe(probeHandler)
	container.ReadinessProbe = res.GetReadinessProbe(probeHandler)
	return container
}
