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

package resources

import (
	operatorv1alpha1 "github.com/ibm/ibm-licensing-operator/pkg/apis/operator/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var (
	licensingServicePort    = intstr.FromInt(8080)
	licensingTargetPort     = intstr.FromInt(8080)
	licensingTargetPortName = intstr.FromString("http")
)

func getServiceSpec(instance *operatorv1alpha1.IBMLicensing) corev1.ServiceSpec {
	return corev1.ServiceSpec{
		Type: corev1.ServiceTypeClusterIP,
		Ports: []corev1.ServicePort{
			{
				Name:       licensingTargetPortName.String(),
				Port:       licensingServicePort.IntVal,
				TargetPort: licensingTargetPort,
				Protocol:   corev1.ProtocolTCP,
			},
		},
		Selector: LabelsForLicensingSelector(instance),
	}
}

func GetLicensingServiceName(instance *operatorv1alpha1.IBMLicensing) string {
	return GetResourceName(instance)
}

func GetLicensingService(instance *operatorv1alpha1.IBMLicensing) *corev1.Service {
	metaLabels := LabelsForLicensingMeta(instance)
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetLicensingServiceName(instance),
			Namespace: instance.Spec.InstanceNamespace,
			Labels:    metaLabels,
			// TODO: check if needed:
			// Annotations: map[string]string{"prometheus.io/scrape": "false", "prometheus.io/scheme": "http"},
		},
		Spec: getServiceSpec(instance),
	}
}
