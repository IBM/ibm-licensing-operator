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

	if *instance.Spec.RHMPEnabled {
		services = append(services, &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:        GetLicensingServiceName(instance) + "-8081",
				Namespace:   instance.Spec.InstanceNamespace,
				Labels:      metaLabels,
				Annotations: resources.AnnotateForService(instance.Spec.HTTPSCertsSource, instance.Spec.HTTPSEnable, LicenseServiceOCPCertName),
			},
			Spec: getServiceSpec(instance, monitorServicePort, monitorTargetPort),
		})
	}

	return services
}
