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

	prometheusServicePort    = intstr.FromInt(8081)
	prometheusTargetPort     = intstr.FromInt(8081)
	prometheusTargetPortName = intstr.FromString("metrics")
)

func GetServices(instance *operatorv1alpha1.IBMLicensing) []*corev1.Service {
	var services []*corev1.Service
	services = append(services, GetLicensingService(instance))

	if s := GetPrometheusService(instance); s != nil {
		services = append(services, s)
	}

	return services
}

func GetLicensingServiceName(instance *operatorv1alpha1.IBMLicensing) string {
	return GetResourceName(instance)
}

func GetLicensingService(instance *operatorv1alpha1.IBMLicensing) *corev1.Service {
	metaLabels := LabelsForMeta(instance)
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        GetLicensingServiceName(instance),
			Namespace:   instance.Spec.InstanceNamespace,
			Labels:      metaLabels,
			Annotations: resources.AnnotateForService(instance.Spec.HTTPSCertsSource, instance.Spec.HTTPSEnable, LicenseServiceOCPCertName),
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Name:       licensingTargetPortName.String(),
					Port:       licensingServicePort.IntVal,
					TargetPort: licensingTargetPort,
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Selector: LabelsForSelector(instance),
		},
	}
}

func GetPrometheusServiceName() string {
	return "license-service-prometheus"
}

func GetPrometheusService(instance *operatorv1alpha1.IBMLicensing) *corev1.Service {
	if !instance.Spec.IsRHMPEnabled() {
		return nil
	}
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetPrometheusServiceName(),
			Namespace: instance.Spec.InstanceNamespace,
			Labels:    getPrometheusLabels(),
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Name:       prometheusTargetPortName.String(),
					Port:       prometheusServicePort.IntVal,
					TargetPort: prometheusTargetPort,
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Selector: LabelsForSelector(instance),
		},
	}
}

func getPrometheusLabels() map[string]string {
	labels := make(map[string]string)
	labels["release"] = "ibm-licensing-service-promethus"
	return labels
}
