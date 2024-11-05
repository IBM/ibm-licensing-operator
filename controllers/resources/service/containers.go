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

package service

import (
	"strconv"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	operatorv1alpha1 "github.com/IBM/ibm-licensing-operator/api/v1alpha1"
	"github.com/IBM/ibm-licensing-operator/controllers/resources"
)

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
		{
			Name:  "ENABLE_INSTANA_METRIC_COLLECTION",
			Value: strconv.FormatBool(spec.EnableInstanaMetricCollection),
		},
	}
	if spec.IsDebug() {
		environmentVariables = append(environmentVariables, corev1.EnvVar{
			Name:  "logging.level.com.ibm",
			Value: "DEBUG",
		})
	}
	if spec.IsDebug() || spec.IsVerbose() {
		environmentVariables = append(environmentVariables, corev1.EnvVar{
			Name:  "SPRING_PROFILES_ACTIVE",
			Value: "verbose",
		})
	}
	if spec.HTTPSEnable {
		environmentVariables = append(environmentVariables, corev1.EnvVar{
			Name:  "HTTPS_CERTS_SOURCE",
			Value: string(operatorv1alpha1.ExternalCertsSource),
		})
	}
	if spec.IsMetering() {
		environmentVariables = append(environmentVariables, corev1.EnvVar{
			Name:  "METERING_URL",
			Value: "https://metering-server:4002/api/v1/metricData",
		})
	}
	if spec.IsRHMPEnabled() {
		environmentVariables = append(environmentVariables, corev1.EnvVar{
			Name:  "enable.metrics",
			Value: "true",
		})
	}
	if spec.IsAlertingEnabled() {
		environmentVariables = append(environmentVariables, corev1.EnvVar{
			Name:  "enable.alerting",
			Value: "true",
		})
	}
	if spec.IsChargebackEnabled() {
		environmentVariables = append(environmentVariables, corev1.EnvVar{
			Name:  "ENABLE_CHARGEBACK",
			Value: "true",
		})
	}
	htThreadsPerCores := spec.GetHyperThreadingThreadsPerCoreOrNil()
	if htThreadsPerCores != nil {
		environmentVariables = append(environmentVariables, corev1.EnvVar{
			Name:  "HYPER_THREADING_THREADS_PER_CORE",
			Value: strconv.Itoa(*htThreadsPerCores),
		})
	}
	if spec.IsNamespaceScopeEnabled() {
		environmentVariables = append(environmentVariables, corev1.EnvVar{
			Name:  "NAMESPACE_SCOPE_ENABLED",
			Value: "true",
		})
		if spec.IsCustomNamespaceScopeConfigMap() {
			customNsConfigMapName := spec.GetCustomNamespaceScopeConfigMap()
			environmentVariables = append(environmentVariables, corev1.EnvVar{
				Name: "WATCH_NAMESPACE",
				ValueFrom: &corev1.EnvVarSource{
					ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
						Key:                  "namespaces",
						LocalObjectReference: corev1.LocalObjectReference{Name: customNsConfigMapName},
					},
				},
			})
		} else {
			// It's not possible for error to occur here so we can ignore it.
			// Should an error occur, it would already fail in main.go and would not reach this code.
			watchNamespaces, _ := resources.GetWatchNamespace()
			environmentVariables = append(environmentVariables, corev1.EnvVar{
				Name:  "WATCH_NAMESPACE",
				Value: watchNamespaces,
			})
		}
		if spec.Features.NamespaceScopeDenialLimit != 0 {
			environmentVariables = append(environmentVariables, corev1.EnvVar{
				Name:  "NAMESPACE_DENIAL_LIMIT",
				Value: strconv.Itoa(spec.Features.NamespaceScopeDenialLimit),
			})
		}
	}
	if spec.ChargebackRetentionPeriod != nil {
		environmentVariables = append(environmentVariables, corev1.EnvVar{
			Name:  "CONTRIBUTIONS_DATA_RETENTION",
			Value: strconv.Itoa(*spec.ChargebackRetentionPeriod),
		})
	}
	if !spec.IsURLBasedAuthEnabled() {
		environmentVariables = append(environmentVariables, corev1.EnvVar{
			Name:  "URL_AUTH_ENABLED",
			Value: "false",
		})
	}
	if spec.IsPrometheusQuerySourceEnabled() && resources.IsServiceCAAPI {
		environmentVariables = append(environmentVariables, corev1.EnvVar{
			Name:  "PROMETHEUS_QUERY_SOURCE_ENABLED",
			Value: "true",
		})
		url := spec.GetPrometheusQuerySourceURL()
		if url != "" {
			environmentVariables = append(environmentVariables, corev1.EnvVar{
				Name:  "thanos_url",
				Value: url,
			})
		}
	}
	if spec.Sender != nil {

		if spec.Sender.ClusterID != "" {
			environmentVariables = append(environmentVariables, []corev1.EnvVar{
				{
					Name:  "CLUSTER_ID",
					Value: spec.Sender.ClusterID,
				},
			}...)
		}

		if spec.Sender.ClusterName != "" {
			environmentVariables = append(environmentVariables, []corev1.EnvVar{
				{
					Name:  "CLUSTER_NAME",
					Value: spec.Sender.ClusterName,
				},
			}...)
		}

		environmentVariables = append(environmentVariables, []corev1.EnvVar{
			{
				Name:  "HUB_URL",
				Value: spec.Sender.ReporterURL,
			},
		}...)

		if spec.Sender.ValidateReporterCerts {
			environmentVariables = append(environmentVariables, []corev1.EnvVar{
				{
					Name:  "VALIDATE_REPORTER_CERTS",
					Value: "true",
				},
			}...)
		}

		// If workloads reporting feature is enabled, pass the env vars marking this to the operand
		if spec.Sender.Frequency != "" {
			environmentVariables = append(environmentVariables, []corev1.EnvVar{
				{
					Name:  "SENDER_WORKLOADS_INTERVAL",
					Value: spec.Sender.Frequency,
				},
			}...)
		}

	}

	if spec.EnvVariable != nil {
		for key, value := range spec.EnvVariable {
			environmentVariables = append(environmentVariables, corev1.EnvVar{
				Name:  key,
				Value: value,
			})
		}
	}
	return environmentVariables
}

func getProbeScheme(spec operatorv1alpha1.IBMLicensingSpec) corev1.URIScheme {
	if spec.HTTPSEnable {
		return "HTTPS"
	}
	return "HTTP"
}

func getProbeHandler(spec operatorv1alpha1.IBMLicensingSpec) corev1.ProbeHandler {
	var probeScheme = getProbeScheme(spec)
	return corev1.ProbeHandler{
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

func GetLicensingInitContainers(spec operatorv1alpha1.IBMLicensingSpec) []corev1.Container {
	containers := []corev1.Container{}
	if spec.IsMetering() {
		baseContainer := getLicensingContainerBase(spec)
		meteringSecretCheckContainer := corev1.Container{}
		baseContainer.DeepCopyInto(&meteringSecretCheckContainer)
		meteringSecretCheckContainer.Name = "metering-check-secret"
		meteringSecretCheckContainer.Command = []string{
			"sh",
			"-c",
			getMeteringSecretCheckScript(),
		}
		containers = append(containers, meteringSecretCheckContainer)
	}
	if resources.IsServiceCAAPI && spec.HTTPSEnable && spec.HTTPSCertsSource == operatorv1alpha1.OcpCertsSource {
		baseContainer := getLicensingContainerBase(spec)
		ocpSecretCheckContainer := corev1.Container{}

		baseContainer.DeepCopyInto(&ocpSecretCheckContainer)
		ocpSecretCheckContainer.Name = resources.OcpCheckString
		ocpSecretCheckContainer.Command = []string{
			"sh",
			"-c",
			resources.GetOCPSecretCheckScript(),
		}
		containers = append(containers, ocpSecretCheckContainer)

		if spec.IsPrometheusServiceNeeded() {
			baseContainer := getLicensingContainerBase(spec)
			ocpPrometheusSecretCheckContainer := corev1.Container{}

			baseContainer.DeepCopyInto(&ocpPrometheusSecretCheckContainer)
			ocpPrometheusSecretCheckContainer.Name = resources.OcpPrometheusCheckString
			ocpPrometheusSecretCheckContainer.Command = []string{
				"sh",
				"-c",
				resources.GetOCPPrometheusSecretCheckScript(),
			}
			containers = append(containers, ocpPrometheusSecretCheckContainer)
		}
	}
	return containers
}

func getLicensingContainerBase(spec operatorv1alpha1.IBMLicensingSpec) corev1.Container {
	container := resources.GetContainerBase(spec.Container)
	if spec.SecurityContext != nil {
		container.SecurityContext.RunAsUser = &spec.SecurityContext.RunAsUser
	}
	container.VolumeMounts = getLicensingVolumeMounts(spec)
	container.Env = getLicensingEnvironmentVariables(spec)
	container.Ports = getLicensingContainerPorts(spec)
	return container
}

func getLicensingContainerPorts(spec operatorv1alpha1.IBMLicensingSpec) []corev1.ContainerPort {
	ports := []corev1.ContainerPort{
		{
			ContainerPort: licensingServicePort.IntVal,
			Protocol:      corev1.ProtocolTCP,
		},
	}

	if spec.IsPrometheusServiceNeeded() {
		ports = append(ports, corev1.ContainerPort{
			ContainerPort: prometheusServicePort.IntVal,
			Protocol:      corev1.ProtocolTCP,
		})
	}

	return ports
}

func GetLicensingContainer(spec operatorv1alpha1.IBMLicensingSpec) []corev1.Container {
	var containers []corev1.Container

	licensingContainer := getLicensingContainerBase(spec)
	probeHandler := getProbeHandler(spec)
	licensingContainer.Name = "license-service"
	licensingContainer.LivenessProbe = resources.GetLivenessProbe(probeHandler)
	licensingContainer.ReadinessProbe = resources.GetReadinessProbe(probeHandler)
	containers = append(containers, licensingContainer)

	return containers
}
