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
	operatorv1alpha1 "github.com/ibm/ibm-licensing-operator/pkg/apis/operator/v1alpha1"
	"github.com/ibm/ibm-licensing-operator/pkg/resources"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var (
	licensingServicePort    = intstr.FromInt(8080)
	licensingTargetPort     = intstr.FromInt(8080)
	licensingTargetPortName = intstr.FromString("api-port")

	monitorServicePort = intstr.FromInt(8081)
	monitorTargetPort  = intstr.FromInt(8081)
)

func getServiceSpec(instance *operatorv1alpha1.IBMLicensing, port, target intstr.IntOrString) corev1.ServiceSpec {
	return corev1.ServiceSpec{
		Type: corev1.ServiceTypeClusterIP,
		Ports: []corev1.ServicePort{
			{
				Name:       licensingTargetPortName.String(),
				Port:       port.IntVal,
				TargetPort: target,
				Protocol:   corev1.ProtocolTCP,
			},
		},
		Selector: LabelsForSelector(instance),
	}
}

func GetLicensingServiceName(instance *operatorv1alpha1.IBMLicensing) string {
	return GetResourceName(instance)
}

func GetLicensingServices(instance *operatorv1alpha1.IBMLicensing) []*corev1.Service {
	metaLabels := LabelsForMeta(instance)
	services := []*corev1.Service{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:        GetLicensingServiceName(instance),
				Namespace:   instance.Spec.InstanceNamespace,
				Labels:      metaLabels,
				Annotations: resources.AnnotateForService(instance.Spec.HTTPSCertsSource, instance.Spec.HTTPSEnable, LicenseServiceOCPCertName),
			},
			Spec: getServiceSpec(instance, licensingServicePort, licensingTargetPort),
		},
	}

	if s := GetPrometheusService(instance); s != nil {
		services = append(services, s)
	}

	return services
}

func GetPrometheusService(instance *operatorv1alpha1.IBMLicensing) *corev1.Service {
	if !*instance.Spec.RHMPEnabled {
		return nil
	}
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "ibm-licensing-service-promethus",
			Namespace:   instance.Spec.InstanceNamespace,
			Labels:      getPrometheusLabels(instance),
			Annotations: resources.AnnotateForService(instance.Spec.HTTPSCertsSource, instance.Spec.HTTPSEnable, LicenseServiceOCPCertName),
		},
		Spec: getServiceSpec(instance, monitorServicePort, monitorTargetPort),
	}
}

func GetServiceMonitor(instance *operatorv1alpha1.IBMLicensing) *monitoringv1.ServiceMonitor {
	return &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ibm-licensing-service-promethus",
			Namespace: instance.Spec.InstanceNamespace,
			Labels: map[string]string{
				"app":                             "ibm-licesnisng-promethus",
				"marketplace.redhat.com/metering": "true",
				"release":                         "prometheus",
			},
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{"unique.label": "unique"},
			},
			Endpoints: []monitoringv1.Endpoint{
				{
					BearerTokenSecret: corev1.SecretKeySelector{
						Key: "",
					},
					Interval:   "15s",
					Path:       "/metrics",
					Scheme:     "https",
					TargetPort: &monitorTargetPort,
					TLSConfig: &monitoringv1.TLSConfig{
						InsecureSkipVerify: true,
					},
				},
			},
		},
	}
}

func getPrometheusLabels(instance *operatorv1alpha1.IBMLicensing) map[string]string {
	labels := LabelsForMeta(instance)
	labels["unique.label"] = "unique"
	return labels
}
