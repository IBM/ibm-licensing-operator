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
	"fmt"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	operatorv1alpha1 "github.com/IBM/ibm-licensing-operator/api/v1alpha1"
)

func GetRHMPServiceMonitor(instance *operatorv1alpha1.IBMLicensing) *monitoringv1.ServiceMonitor {
	interval := "3h"
	name := PrometheusRHMPServiceMonitor
	tlsConfig := getTLSConfigForServiceMonitor(instance)
	metricRelabelConfigs := getMetricRelabelConfigsForRHMP()
	return GetServiceMonitor(instance, name, interval, tlsConfig, metricRelabelConfigs)
}

func GetAlertingServiceMonitor(instance *operatorv1alpha1.IBMLicensing) *monitoringv1.ServiceMonitor {
	interval := "5m"
	name := PrometheusAlertingServiceMonitor
	tlsConfig := getTLSConfigForServiceMonitor(instance)
	metricRelabelConfigs := getMetricRelabelConfigsForAlerting()
	return GetServiceMonitor(instance, name, interval, tlsConfig, metricRelabelConfigs)
}

func GetServiceMonitor(instance *operatorv1alpha1.IBMLicensing, name string, interval string,
	tlsConfig *monitoringv1.TLSConfig, metricRelabelConfigs []*monitoringv1.RelabelConfig) *monitoringv1.ServiceMonitor {

	return &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: instance.Spec.InstanceNamespace,
			Labels:    LabelsForServiceMonitor(),
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: getPrometheusLabels(),
			},
			Endpoints: []monitoringv1.Endpoint{
				{
					BearerTokenSecret: corev1.SecretKeySelector{
						Key: "",
					},
					Path:                 "/metrics",
					MetricRelabelConfigs: metricRelabelConfigs,
					Scheme:               getScheme(instance),
					TargetPort:           &prometheusTargetPort,
					TLSConfig:            tlsConfig,
					Interval:             interval,
					RelabelConfigs:       getRelabelConfigs(instance),
				},
			},
		},
	}
}

// initialize with all prometheus metrics from License Service operand monitoring service
var orderedAllPrometheusMetrics = []string{
	"cp4d_capability",
	"ibm_licensing_usage_daily_high_watermark",
	"product_license_usage",
	"product_license_usage_chargeback",
	"product_license_usage_details",
}

func getMetricRelabelConfigsForRHMP() []*monitoringv1.RelabelConfig {
	var usedMetrics = map[string]bool{
		"product_license_usage":            true,
		"product_license_usage_chargeback": true,
		"product_license_usage_details":    true,
		"cp4d_capability":                  true,
	}
	return getMetricRelabelConfigs(usedMetrics)
}

func getMetricRelabelConfigsForAlerting() []*monitoringv1.RelabelConfig {
	var usedMetrics = map[string]bool{
		"ibm_licensing_usage_daily_high_watermark": true,
	}
	return getMetricRelabelConfigs(usedMetrics)
}

// return metric relabel config that drops not used prometheus metrics
func getMetricRelabelConfigs(usedMetrics map[string]bool) []*monitoringv1.RelabelConfig {
	regex := "("
	isFirstMetric := true

	for _, metric := range orderedAllPrometheusMetrics {

		// drop metrics if used metrics doesn't have it
		if !usedMetrics[metric] {
			if isFirstMetric {
				isFirstMetric = false
			} else {
				regex += "|"
			}
			regex += metric
		}
	}
	regex += ")"

	var sourceLabels []string
	sourceLabels = append(sourceLabels, "__name__")

	relabelConfigs := make([]*monitoringv1.RelabelConfig, 0)
	relabelConfigs = append(relabelConfigs, &monitoringv1.RelabelConfig{
		Action:       "drop",
		Regex:        regex,
		SourceLabels: sourceLabels,
	})

	return relabelConfigs
}

func getRelabelConfigs(instance *operatorv1alpha1.IBMLicensing) []*monitoringv1.RelabelConfig {
	relabelConfigs := make([]*monitoringv1.RelabelConfig, 0)
	if instance.Spec.HTTPSEnable {
		relabelConfigs = append(relabelConfigs, &monitoringv1.RelabelConfig{
			Replacement: fmt.Sprintf("%s:%d", getServerName(instance), prometheusTargetPort.IntVal),
			TargetLabel: "__address__",
		})
	}
	return relabelConfigs
}

func getTLSConfigForServiceMonitor(instance *operatorv1alpha1.IBMLicensing) *monitoringv1.TLSConfig {
	if instance.Spec.HTTPSEnable {
		return &monitoringv1.TLSConfig{
			CA: monitoringv1.SecretOrConfigMap{
				Secret: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: ServiceAccountSecretName,
					},
					Key: "service-ca.crt",
				},
			},
		}
	}
	return nil
}

func getServerName(instance *operatorv1alpha1.IBMLicensing) string {
	return fmt.Sprintf("%s.%s.svc", GetPrometheusServiceName(), instance.Spec.InstanceNamespace)
}

func getScheme(instance *operatorv1alpha1.IBMLicensing) string {
	if instance.Spec.HTTPSEnable {
		return "https"
	}
	return "http"
}
