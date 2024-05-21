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
	routev1 "github.com/openshift/api/route/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	operatorv1alpha1 "github.com/IBM/ibm-licensing-operator/api/v1alpha1"
	"github.com/IBM/ibm-licensing-operator/controllers/resources"
)

func GetLicensingRoute(instance *operatorv1alpha1.IBMLicensing, defaultRouteTLS *routev1.TLSConfig) *routev1.Route {
	var tls *routev1.TLSConfig

	if instance.Spec.RouteOptions != nil {
		if instance.Spec.RouteOptions.TLS == nil {
			if instance.Spec.HTTPSEnable {
				tls = defaultRouteTLS
			}
		} else {
			tls = instance.Spec.RouteOptions.TLS
		}
	} else {
		if instance.Spec.HTTPSEnable {
			tls = defaultRouteTLS
		}
	}
	return &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetResourceName(instance),
			Namespace: instance.Spec.InstanceNamespace,
		},
		Spec: routev1.RouteSpec{
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: GetResourceName(instance),
			},
			Port: &routev1.RoutePort{
				TargetPort: licensingTargetPortName,
			},
			TLS: tls,
		},
	}
}

func GetLicensingIngress(instance *operatorv1alpha1.IBMLicensing) *networkingv1.Ingress {
	var (
		tls              []networkingv1.IngressTLS
		path, host       string
		annotations      map[string]string
		ingressClassName *string
	)
	path = "/" + GetResourceName(instance)
	options := instance.Spec.IngressOptions
	if options != nil {
		tls = options.TLS
		if options.Path != nil {
			path = *options.Path
		}
		if options.Host != nil {
			host = *options.Host
		}
		ingressClassName = options.IngressClassName
		annotations = options.Annotations
	}
	return &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        GetResourceName(instance),
			Namespace:   instance.Spec.InstanceNamespace,
			Annotations: annotations,
		},
		Spec: networkingv1.IngressSpec{
			TLS:              tls,
			IngressClassName: ingressClassName,
			Rules: []networkingv1.IngressRule{
				{
					Host: host,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     path,
									PathType: &resources.PathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: GetLicensingServiceName(instance),
											Port: networkingv1.ServiceBackendPort{
												Number: licensingServicePort.IntVal,
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
