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

package reporter

import (
	operatorv1alpha1 "github.com/ibm/ibm-licensing-operator/api/v1alpha1"
	routev1 "github.com/openshift/api/route/v1"
	extensionsv1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetReporterRoute(instance *operatorv1alpha1.IBMLicenseServiceReporter) *routev1.Route {
	return &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      LicenseReporterResourceBase,
			Namespace: instance.GetNamespace(),
			Labels:    LabelsForMeta(instance),
		},
		Spec: routev1.RouteSpec{
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: LicenseReporterResourceBase,
			},
			Port: &routev1.RoutePort{
				TargetPort: receiverTargetPortName,
			},
			TLS: &routev1.TLSConfig{
				Termination: routev1.TLSTerminationPassthrough,
			},
		},
	}
}

func annotationsForIngress() map[string]string {
	return map[string]string{"icp.management.ibm.com/auth-type": "access-token", "kubernetes.io/ingress.class": "ibm-icp-management"}
}

func GetUIIngress(instance *operatorv1alpha1.IBMLicenseServiceReporter) *extensionsv1.Ingress {
	return &extensionsv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        LicenseReporterUIBase,
			Namespace:   instance.GetNamespace(),
			Labels:      LabelsForMeta(instance),
			Annotations: annotationsForIngress(),
		},
		Spec: extensionsv1.IngressSpec{
			Rules: []extensionsv1.IngressRule{
				{
					IngressRuleValue: extensionsv1.IngressRuleValue{
						HTTP: &extensionsv1.HTTPIngressRuleValue{
							Paths: []extensionsv1.HTTPIngressPath{
								{
									Path: "/license-service-reporter",
									Backend: extensionsv1.IngressBackend{
										ServiceName: LicenseReporterResourceBase,
										ServicePort: reporterUIServicePort,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func GetUIIngressProxy(instance *operatorv1alpha1.IBMLicenseServiceReporter) *extensionsv1.Ingress {
	return &extensionsv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      LicenseReporterUIBase + "-proxy",
			Namespace: instance.GetNamespace(),
			Labels:    LabelsForMeta(instance),
		},
		Spec: extensionsv1.IngressSpec{
			Rules: []extensionsv1.IngressRule{
				{
					IngressRuleValue: extensionsv1.IngressRuleValue{
						HTTP: &extensionsv1.HTTPIngressRuleValue{
							Paths: []extensionsv1.HTTPIngressPath{{
								Path: "/console",
								Backend: extensionsv1.IngressBackend{
									ServiceName: LicenseReporterResourceBase,
									ServicePort: reporterUIServicePort,
								},
							}},
						},
					},
				},
			},
		},
	}
}
