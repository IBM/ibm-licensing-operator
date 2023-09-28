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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	operatorv1alpha1 "github.com/IBM/ibm-licensing-operator/api/v1alpha1"
	"github.com/IBM/ibm-licensing-operator/controllers/resources"
)

var (
	reporterUIServicePort    = intstr.FromInt(UIPort)
	reporterUITargetPort     = intstr.FromInt(UIPort)
	reporterUITargetPortName = intstr.FromString("reporter-ui-port")
	receiverServicePort      = intstr.FromInt(ReceiverPort)
	receiverTargetPort       = intstr.FromInt(ReceiverPort)
	receiverTargetPortName   = intstr.FromString("receiver-port")
)

func getServiceSpec(instance *operatorv1alpha1.IBMLicenseServiceReporter) corev1.ServiceSpec {
	return corev1.ServiceSpec{
		Type: corev1.ServiceTypeClusterIP,
		Ports: []corev1.ServicePort{
			{
				Name:       reporterUITargetPortName.String(),
				Port:       reporterUIServicePort.IntVal,
				TargetPort: reporterUITargetPort,
				Protocol:   corev1.ProtocolTCP,
			},
			{
				Name:       receiverTargetPortName.String(),
				Port:       receiverServicePort.IntVal,
				TargetPort: receiverTargetPort,
				Protocol:   corev1.ProtocolTCP,
			},
		},
		Selector: LabelsForSelector(instance),
	}
}

func GetService(instance *operatorv1alpha1.IBMLicenseServiceReporter) *corev1.Service {
	metaLabels := LabelsForMeta(instance)
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        LicenseReporterResourceBase,
			Namespace:   instance.GetNamespace(),
			Labels:      metaLabels,
			Annotations: resources.AnnotateForService(true, LicenseReportOCPCertName),
		},
		Spec: getServiceSpec(instance),
	}
}
