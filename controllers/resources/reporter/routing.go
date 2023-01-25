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
	routev1 "github.com/openshift/api/route/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	operatorv1alpha1 "github.com/IBM/ibm-licensing-operator/api/v1alpha1"
	"github.com/IBM/ibm-licensing-operator/controllers/resources"
)

func annotationsForReporterRoute() map[string]string {
	return map[string]string{"haproxy.router.openshift.io/timeout": "90s"}
}

func GetReporterRoute(instance *operatorv1alpha1.IBMLicenseServiceReporter, defaultRouteTLS *routev1.TLSConfig) *routev1.Route {
	var tls *routev1.TLSConfig

	if instance.Spec.RouteOptions != nil {
		if instance.Spec.RouteOptions.TLS != nil {
			tls = instance.Spec.RouteOptions.TLS
		}
	} else {
		tls = defaultRouteTLS
	}

	return &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:        LicenseReporterResourceBase,
			Namespace:   instance.GetNamespace(),
			Labels:      LabelsForMeta(instance),
			Annotations: annotationsForReporterRoute(),
		},
		Spec: routev1.RouteSpec{
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: LicenseReporterResourceBase,
			},
			Port: &routev1.RoutePort{
				TargetPort: receiverTargetPortName,
			},
			TLS: tls,
		},
	}
}

func annotationsForIngress() map[string]string {
	return map[string]string{"icp.management.ibm.com/auth-type": "access-token", "kubernetes.io/ingress.class": "ibm-icp-management"}
}

func GetUIIngress(instance *operatorv1alpha1.IBMLicenseServiceReporter) *networkingv1.Ingress {
	return &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        LicenseReporterUIBase,
			Namespace:   instance.GetNamespace(),
			Labels:      LabelsForMeta(instance),
			Annotations: annotationsForIngress(),
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/license-service-reporter",
									PathType: &resources.PathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: LicenseReporterResourceBase,
											Port: networkingv1.ServiceBackendPort{
												Number: UIPort,
											},
										},
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

func GetUIIngressProxy(instance *operatorv1alpha1.IBMLicenseServiceReporter) *networkingv1.Ingress {
	return &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      LicenseReporterUIBase + "-proxy",
			Namespace: instance.GetNamespace(),
			Labels:    LabelsForMeta(instance),
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{{
								Path:     "/console",
								PathType: &resources.PathType,
								Backend: networkingv1.IngressBackend{
									Service: &networkingv1.IngressServiceBackend{
										Name: LicenseReporterResourceBase,
										Port: networkingv1.ServiceBackendPort{
											Number: UIPort,
										},
									},
								},
							}},
						},
					},
				},
			},
		},
	}
}
