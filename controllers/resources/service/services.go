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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	operatorv1alpha1 "github.com/IBM/ibm-licensing-operator/api/v1alpha1"
	"github.com/IBM/ibm-licensing-operator/controllers/resources"
)

var (
	licensingServicePort    = intstr.FromInt(8080)
	licensingTargetPort     = intstr.FromInt(8080)
	licensingTargetPortName = intstr.FromString("api-port")

	prometheusServicePort    = intstr.FromInt(8081)
	prometheusTargetPort     = intstr.FromInt(8081)
	prometheusTargetPortName = intstr.FromString("metrics")

	usageServicePort    = intstr.FromInt(2112)
	usageTargetPort     = intstr.FromInt(2112)
	usageTargetPortName = intstr.FromString("usage-port")
)

func GetServices(instance *operatorv1alpha1.IBMLicensing) (expected []*corev1.Service, notExpected []*corev1.Service) {
	expected = append(expected, GetLicensingService(instance))

	prometheusService := GetPrometheusService(instance)
	if instance.Spec.IsPrometheusServiceNeeded() {
		expected = append(expected, prometheusService)
	} else {
		notExpected = append(notExpected, prometheusService)
	}

	usageService := GetUsageService(instance)
	if instance.Spec.UsageEnabled {
		expected = append(expected, usageService)
		return
	}
	notExpected = append(notExpected, usageService)
	return
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
			Annotations: resources.AnnotateForService(instance.Spec.HTTPSEnable, LicenseServiceInternalCertName),
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

func GetUsageServiceName() string {
	return UsageServiceName
}

func GetPrometheusServiceName() string {
	return PrometheusServiceName
}

func GetPrometheusService(instance *operatorv1alpha1.IBMLicensing) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        GetPrometheusServiceName(),
			Namespace:   instance.Spec.InstanceNamespace,
			Labels:      getPrometheusLabels(),
			Annotations: resources.AnnotateForService(instance.Spec.HTTPSEnable, PrometheusServiceOCPCertName),
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

func GetUsageService(instance *operatorv1alpha1.IBMLicensing) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetUsageServiceName(),
			Namespace: instance.Spec.InstanceNamespace,
			Labels:    getUsageServiceLabels(),
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Name:       usageTargetPortName.String(),
					Port:       usageServicePort.IntVal,
					TargetPort: usageTargetPort,
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Selector: LabelsForSelector(instance),
		},
	}
}

func getPrometheusLabels() map[string]string {
	labels := make(map[string]string)
	labels["release"] = ReleaseLabel
	return labels
}

func getUsageServiceLabels() map[string]string {
	labels := make(map[string]string)
	labels["release"] = ReleaseUsageLabel
	return labels
}
