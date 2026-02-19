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
	"fmt"

	routev1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	operatorv1alpha1 "github.com/IBM/ibm-licensing-operator/api/v1alpha1"
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

func GetLicensingGateway(instance *operatorv1alpha1.IBMLicensing) *gatewayv1.Gateway {
	options := instance.Spec.GatewayOptions
	name := GetResourceName(instance) + "-gateway"
	className := "ibm-licensing"
	var annotations map[string]string

	if options != nil {
		if options.GatewayClassName != "" {
			className = options.GatewayClassName
		}
		annotations = options.Annotations
	}

	listeners := []gatewayv1.Listener{
		{
			Name:     "http",
			Protocol: gatewayv1.HTTPProtocolType,
			Port:     gatewayv1.PortNumber(80),
			AllowedRoutes: &gatewayv1.AllowedRoutes{
				Namespaces: &gatewayv1.RouteNamespaces{From: ptr.To(gatewayv1.NamespacesFromAll)},
			},
		},
	}

	if options != nil && options.TLSSecretName != "" {
		listeners = append(listeners, gatewayv1.Listener{
			Name:     "https",
			Protocol: gatewayv1.HTTPSProtocolType,
			Port:     gatewayv1.PortNumber(8080),
			TLS: &gatewayv1.ListenerTLSConfig{
				Mode: ptr.To(gatewayv1.TLSModeTerminate),
				CertificateRefs: []gatewayv1.SecretObjectReference{{
					Kind: ptr.To(gatewayv1.Kind("Secret")),
					Name: gatewayv1.ObjectName(options.TLSSecretName),
				}},
			},
		})
	}

	return &gatewayv1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   instance.Spec.InstanceNamespace,
			Annotations: annotations,
		},
		Spec: gatewayv1.GatewaySpec{
			GatewayClassName: gatewayv1.ObjectName(className),
			Listeners:        listeners,
		},
	}

}

func GetLicensingHTTPRoute(instance *operatorv1alpha1.IBMLicensing) *gatewayv1.HTTPRoute {
	path := "/" + GetResourceName(instance)
	routeName := GetResourceName(instance) + "-route"
	gatewayName := GetResourceName(instance) + "-gateway"
	serviceName := GetResourceName(instance)

	return &gatewayv1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:      routeName,
			Namespace: instance.Spec.InstanceNamespace,
		},
		Spec: gatewayv1.HTTPRouteSpec{
			CommonRouteSpec: gatewayv1.CommonRouteSpec{
				ParentRefs: []gatewayv1.ParentReference{{
					Name: gatewayv1.ObjectName(gatewayName),
				}},
			},
			Rules: []gatewayv1.HTTPRouteRule{{
				Matches: []gatewayv1.HTTPRouteMatch{{
					Path: &gatewayv1.HTTPPathMatch{
						Type:  ptr.To(gatewayv1.PathMatchPathPrefix),
						Value: ptr.To(path),
					},
				}},
				Filters: []gatewayv1.HTTPRouteFilter{{
					Type: gatewayv1.HTTPRouteFilterURLRewrite,
					URLRewrite: &gatewayv1.HTTPURLRewriteFilter{
						Path: &gatewayv1.HTTPPathModifier{
							Type:               gatewayv1.PrefixMatchHTTPPathModifier,
							ReplacePrefixMatch: ptr.To("/"),
						},
					},
				}},

				BackendRefs: []gatewayv1.HTTPBackendRef{{
					BackendRef: gatewayv1.BackendRef{
						BackendObjectReference: gatewayv1.BackendObjectReference{
							Kind: ptr.To(gatewayv1.Kind("Service")),
							Name: gatewayv1.ObjectName(serviceName),
							Port: ptr.To(gatewayv1.PortNumber(8080)),
						},
					},
				}},
			}},
		},
	}
}

func GetGatewayConfigMap(instance *operatorv1alpha1.IBMLicensing, internalCertData string) *corev1.ConfigMap {
	newConfigMapName := "ibm-licensing-gateway-api-config"

	expectedCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:        newConfigMapName,
			Namespace:   instance.Spec.InstanceNamespace,
			Annotations: instance.Spec.Annotations,
		},
		Data: map[string]string{
			"ca.crt": internalCertData,
		},
	}
	return expectedCM
}

func GetBackEndTLSPolicy(instance *operatorv1alpha1.IBMLicensing) *gatewayv1.BackendTLSPolicy {
	policyName := GetResourceName(instance) + "-backend-tls"
	targetConfigMapName := "ibm-licensing-gateway-api-config"
	serviceName := GetLicensingServiceName(instance)
	hostname := fmt.Sprintf("%s.%s.svc.cluster.local", serviceName, instance.Spec.InstanceNamespace)

	return &gatewayv1.BackendTLSPolicy{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "gateway.networking.k8s.io/v1alpha3",
			Kind:       "BackendTLSPolicy",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      policyName,
			Namespace: instance.Spec.InstanceNamespace,
		},
		Spec: gatewayv1.BackendTLSPolicySpec{
			TargetRefs: []gatewayv1.LocalPolicyTargetReferenceWithSectionName{{
				LocalPolicyTargetReference: gatewayv1.LocalPolicyTargetReference{
					Group: gatewayv1.Group(""),
					Kind:  gatewayv1.Kind("Service"),
					Name:  gatewayv1.ObjectName(serviceName),
				},
			},
			},
			Validation: gatewayv1.BackendTLSPolicyValidation{
				Hostname: gatewayv1.PreciseHostname(hostname),
				CACertificateRefs: []gatewayv1.LocalObjectReference{{
					Group: "",
					Kind:  "ConfigMap",
					Name:  gatewayv1.ObjectName(targetConfigMapName),
				}},
			},
		},
	}

}

// Deprecated: GetLicensingIngress is deprecated and not currently used
// func GetLicensingIngress(instance *operatorv1alpha1.IBMLicensing) *networkingv1.Ingress {
// 	var (
// 		tls              []networkingv1.IngressTLS
// 		path, host       string
// 		annotations      map[string]string
// 		ingressClassName *string
// 	)
// 	path = "/" + GetResourceName(instance)
// 	options := instance.Spec.GatewayOptions
// 	if options != nil {
// 		tls = options.TLS
// 		if options.Path != nil {
// 			path = *options.Path
// 		}
// 		if options.Host != nil {
// 			host = *options.Host
// 		}
// 		ingressClassName = options.IngressClassName
// 		annotations = options.Annotations
// 	}
// 	return &networkingv1.Ingress{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:        GetResourceName(instance),
// 			Namespace:   instance.Spec.InstanceNamespace,
// 			Annotations: annotations,
// 		},
// 		Spec: networkingv1.IngressSpec{
// 			TLS:              tls,
// 			IngressClassName: ingressClassName,
// 			Rules: []networkingv1.IngressRule{
// 				{
// 					Host: host,
// 					IngressRuleValue: networkingv1.IngressRuleValue{
// 						HTTP: &networkingv1.HTTPIngressRuleValue{
// 							Paths: []networkingv1.HTTPIngressPath{
// 								{
// 									Path:     path,
// 									PathType: &resources.PathType,
// 									Backend: networkingv1.IngressBackend{
// 										Service: &networkingv1.IngressServiceBackend{
// 											Name: GetLicensingServiceName(instance),
// 											Port: networkingv1.ServiceBackendPort{
// 												Number: licensingServicePort.IntVal,
// 											},
// 										},
// 									},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 		},
// 	}
// }
