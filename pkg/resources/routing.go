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
	routev1 "github.com/openshift/api/route/v1"
	extensionsv1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetLicensingRoute(instance *operatorv1alpha1.IBMLicensing) *routev1.Route {
	var tls *routev1.TLSConfig
	defaultRouteTLS := &routev1.TLSConfig{
		Termination:                   routev1.TLSTerminationPassthrough,
		InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyNone,
	}
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

func GetLicensingIngress(instance *operatorv1alpha1.IBMLicensing) *extensionsv1.Ingress {
	var (
		tls         []extensionsv1.IngressTLS
		path, host  string
		annotations map[string]string
	)
	options := instance.Spec.IngressOptions
	if options != nil {
		tls = options.TLS
		if options.Path != nil {
			path = *options.Path
		} else {
			path = "/" + GetResourceName(instance)
		}
		if options.Host != nil {
			host = *options.Host
		}
		annotations = options.Annotations
	}
	return &extensionsv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        GetResourceName(instance),
			Namespace:   instance.Spec.InstanceNamespace,
			Annotations: annotations,
		},
		Spec: extensionsv1.IngressSpec{
			TLS: tls,
			Rules: []extensionsv1.IngressRule{
				{
					Host: host,
					IngressRuleValue: extensionsv1.IngressRuleValue{
						HTTP: &extensionsv1.HTTPIngressRuleValue{
							Paths: []extensionsv1.HTTPIngressPath{
								{
									Path: path,
									Backend: extensionsv1.IngressBackend{
										ServiceName: GetLicensingServiceName(instance),
										ServicePort: licensingServicePort,
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
