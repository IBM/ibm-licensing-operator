//
// Copyright 2026 IBM Corporation
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
	"maps"

	routev1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	operatorv1alpha1 "github.com/IBM/ibm-licensing-operator/api/v1alpha1"
)

const (
	defaultGatewayClassName = "ibm-licensing"
	defaultHTTPPort         = int32(8080)
	defaultHTTPSPort        = int32(443)
	GatewayConfigMapName    = "ibm-licensing-gateway-api-config"
	kindService             = "Service"
	kindSecret              = "Secret"
	kindConfigMap           = "ConfigMap"
)

func GetGatewayName(instance *operatorv1alpha1.IBMLicensing) string {
	return "ibm-licensing-service-gateway"
}

func GetHTTPRouteName(instance *operatorv1alpha1.IBMLicensing) string {
	return "ibm-licensing"
}

func GetBackendTLSPolicyName(instance *operatorv1alpha1.IBMLicensing) string {
	return "licensing-backend-tls"
}

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
				Kind: kindService,
				Name: GetResourceName(instance),
			},
			Port: &routev1.RoutePort{
				TargetPort: licensingTargetPortName,
			},
			TLS: tls,
		},
	}
}

func newGatewayListener(name string, protocol gatewayv1.ProtocolType, port int32, tlsConfig *gatewayv1.ListenerTLSConfig) gatewayv1.Listener {
	listener := gatewayv1.Listener{
		Name:     gatewayv1.SectionName(name),
		Protocol: protocol,
		Port:     port,
		AllowedRoutes: &gatewayv1.AllowedRoutes{
			Namespaces: &gatewayv1.RouteNamespaces{From: ptr.To(gatewayv1.NamespacesFromSame)},
		},
	}
	if tlsConfig != nil {
		listener.TLS = tlsConfig
	}
	return listener
}

// mergeGatewayAnnotations merges instance-level Spec.Annotations with Gateway-specific
// GatewayOptions.Annotations. GatewayOptions.Annotations take precedence on key conflicts.
func mergeGatewayAnnotations(instance *operatorv1alpha1.IBMLicensing) map[string]string {
	options := instance.Spec.GatewayOptions
	if options == nil {
		options = &operatorv1alpha1.IBMLicensingGatewayOptions{}
	}

	if len(instance.Spec.Annotations) == 0 && len(options.Annotations) == 0 {
		return nil
	}

	merged := make(map[string]string, len(instance.Spec.Annotations)+len(options.Annotations))
	maps.Copy(merged, instance.Spec.Annotations)
	maps.Copy(merged, options.Annotations)

	return merged
}

func GetLicensingGateway(instance *operatorv1alpha1.IBMLicensing) *gatewayv1.Gateway {
	options := instance.Spec.GatewayOptions
	name := GetGatewayName(instance)

	if options == nil {
		options = &operatorv1alpha1.IBMLicensingGatewayOptions{}
	}

	className := defaultGatewayClassName
	if options.GatewayClassName != "" {
		className = options.GatewayClassName
	}

	httpsPort := defaultHTTPSPort
	if options.HTTPSPort != nil {
		httpsPort = *options.HTTPSPort
	}

	listeners := []gatewayv1.Listener{}

	tlsSecretName := options.TLSSecretName
	if tlsSecretName == "" {
		tlsSecretName = "ibm-license-service-cert-internal"
	}

	tlsConfig := &gatewayv1.ListenerTLSConfig{
		Mode: ptr.To(gatewayv1.TLSModeTerminate),
		CertificateRefs: []gatewayv1.SecretObjectReference{{
			Kind: ptr.To(gatewayv1.Kind(kindSecret)),
			Name: gatewayv1.ObjectName(tlsSecretName),
		}},
	}
	listeners = append(listeners, newGatewayListener("https", gatewayv1.HTTPSProtocolType, httpsPort, tlsConfig))

	return &gatewayv1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   instance.Spec.InstanceNamespace,
			Labels:      LabelsForMeta(instance),
			Annotations: mergeGatewayAnnotations(instance),
		},
		Spec: gatewayv1.GatewaySpec{
			GatewayClassName: gatewayv1.ObjectName(className),
			Listeners:        listeners,
		},
	}
}

func GetLicensingHTTPRoute(instance *operatorv1alpha1.IBMLicensing) *gatewayv1.HTTPRoute {
	path := "/" + GetResourceName(instance)
	routeName := GetHTTPRouteName(instance)
	gatewayName := GetGatewayName(instance)
	serviceName := GetResourceName(instance)

	options := instance.Spec.GatewayOptions
	if options == nil {
		options = &operatorv1alpha1.IBMLicensingGatewayOptions{}
	}

	return &gatewayv1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:        routeName,
			Namespace:   instance.Spec.InstanceNamespace,
			Labels:      LabelsForMeta(instance),
			Annotations: mergeGatewayAnnotations(instance),
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
							Kind: ptr.To(gatewayv1.Kind(kindService)),
							Name: gatewayv1.ObjectName(serviceName),
							Port: ptr.To(gatewayv1.PortNumber(defaultHTTPPort)),
						},
					},
				}},
			}},
		},
	}
}

func GetGatewayConfigMap(instance *operatorv1alpha1.IBMLicensing, internalCertData string) *corev1.ConfigMap {
	options := instance.Spec.GatewayOptions
	if options == nil {
		options = &operatorv1alpha1.IBMLicensingGatewayOptions{}
	}

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:        GatewayConfigMapName,
			Namespace:   instance.Spec.InstanceNamespace,
			Labels:      LabelsForMeta(instance),
			Annotations: mergeGatewayAnnotations(instance),
		},
		Data: map[string]string{
			"ca.crt": internalCertData,
		},
	}
}

func GetBackEndTLSPolicy(instance *operatorv1alpha1.IBMLicensing) *gatewayv1.BackendTLSPolicy {
	policyName := GetBackendTLSPolicyName(instance)
	serviceName := GetLicensingServiceName(instance)
	hostname := GetServiceHostname(instance)

	options := instance.Spec.GatewayOptions
	if options == nil {
		options = &operatorv1alpha1.IBMLicensingGatewayOptions{}
	}

	return &gatewayv1.BackendTLSPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:        policyName,
			Namespace:   instance.Spec.InstanceNamespace,
			Labels:      LabelsForMeta(instance),
			Annotations: mergeGatewayAnnotations(instance),
		},
		Spec: gatewayv1.BackendTLSPolicySpec{
			TargetRefs: []gatewayv1.LocalPolicyTargetReferenceWithSectionName{{
				LocalPolicyTargetReference: gatewayv1.LocalPolicyTargetReference{
					Group: gatewayv1.Group(""),
					Kind:  gatewayv1.Kind(kindService),
					Name:  gatewayv1.ObjectName(serviceName),
				},
			},
			},
			Validation: gatewayv1.BackendTLSPolicyValidation{
				Hostname: gatewayv1.PreciseHostname(hostname),
				CACertificateRefs: []gatewayv1.LocalObjectReference{{
					Group: "",
					Kind:  kindConfigMap,
					Name:  gatewayv1.ObjectName(GatewayConfigMapName),
				}},
			},
		},
	}

}
