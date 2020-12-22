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
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	operatorv1alpha1 "github.com/ibm/ibm-licensing-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetServiceMonitorName() string {
	return PrometheusServiceMonitor
}

func GetServiceMonitor(instance *operatorv1alpha1.IBMLicensing) *monitoringv1.ServiceMonitor {
	return &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetServiceMonitorName(),
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
					Path:       "/metrics",
					Scheme:     getScheme(instance),
					TargetPort: &prometheusTargetPort,
					TLSConfig:  getTLSConfig(instance),
				},
			},
		},
	}
}

func getTLSConfig(instance *operatorv1alpha1.IBMLicensing) *monitoringv1.TLSConfig {
	if instance.Spec.HTTPSEnable {
		return &monitoringv1.TLSConfig{
			InsecureSkipVerify: true,
		}
	}
	return nil
}

func getScheme(instance *operatorv1alpha1.IBMLicensing) string {
	if instance.Spec.HTTPSEnable {
		return "https"
	}
	return "http"
}
