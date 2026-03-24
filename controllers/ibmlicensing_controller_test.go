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

package controllers

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	routev1 "github.com/openshift/api/route/v1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	rhmp "github.com/IBM/ibm-licensing-operator/pkg/rhmp/v1beta1"

	operatorv1alpha1 "github.com/IBM/ibm-licensing-operator/api/v1alpha1"
	"github.com/IBM/ibm-licensing-operator/controllers/resources/service"
)

var _ = Describe("IBMLicensing controller", Ordered, func() {
	const (
		name = "instance"
	)

	var (
		ctx               context.Context
		instance          *operatorv1alpha1.IBMLicensing
		instanceForRemove = &operatorv1alpha1.IBMLicensing{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			}}
	)

	BeforeEach(func() {
		ctx = context.Background()
		k8sClient.Delete(ctx, instanceForRemove)
	})

	AfterEach(func() {
		k8sClient.Delete(ctx, instanceForRemove)
	})

	Context("Initializing IBMLicensing Status", func() {
		It("Should not create IBMLicensing instance", func() {
			By("Creating broken IBMLicensing without datasource")
			instance = &operatorv1alpha1.IBMLicensing{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Spec: operatorv1alpha1.IBMLicensingSpec{
					InstanceNamespace: namespace,
					License: &operatorv1alpha1.License{
						Accept: true,
					},
				},
			}

			Expect(k8sClient.Create(ctx, instance)).Should(MatchError(ContainSubstring("spec.datasource")))

			By("Creating broken IBMLicensing with wrong datasource")
			instance = &operatorv1alpha1.IBMLicensing{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Spec: operatorv1alpha1.IBMLicensingSpec{
					InstanceNamespace: namespace,
					Datasource:        "datacollector1",
					License: &operatorv1alpha1.License{
						Accept: true,
					},
				},
			}
			Expect(k8sClient.Create(ctx, instance)).Should(MatchError(ContainSubstring("spec.datasource")))
		})

		It("Should create specific license not accepted logs and events", func() {
			By("Creating IBMLicensing instance without license field")
			instance = &operatorv1alpha1.IBMLicensing{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Spec: operatorv1alpha1.IBMLicensingSpec{
					InstanceNamespace: namespace,
					Datasource:        "datacollector",
					Container: operatorv1alpha1.Container{
						ImagePullPolicy: v1.PullAlways,
					},
					IBMLicenseServiceBaseSpec: operatorv1alpha1.IBMLicenseServiceBaseSpec{
						ImagePullSecrets: []string{"artifactory-token"},
					},
				},
			}

			Expect(k8sClient.Create(ctx, instance)).Should(Succeed())

			newInstance := &operatorv1alpha1.IBMLicensing{}
			events := &v1.EventList{}

			By("Checking if license is not accepted")
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: instance.Name}, newInstance); err != nil {
					return true
				}
				return newInstance.Spec.IsLicenseAccepted()
			}, timeout, interval).Should(Equal(false))

			By("Checking if 'license not accepted' event was created successfully")
			Eventually(func() bool {
				Expect(k8sClient.List(ctx, events)).Should(Succeed())

				for _, event := range events.Items {
					if event.Message == operatorv1alpha1.LicenseNotAcceptedMessage {
						// Pass test if event was created correctly
						fmt.Println("License not accepted event found: " + event.Message)
						return true
					}
				}

				return false
			}, timeout, interval).Should(Equal(true))
		})

		It("Should create IBMLicensing instance HTTP", func() {
			By("Creating the IBMLicensing")
			newInstance := &operatorv1alpha1.IBMLicensing{}

			instance = &operatorv1alpha1.IBMLicensing{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Spec: operatorv1alpha1.IBMLicensingSpec{
					InstanceNamespace: namespace,
					Datasource:        "datacollector",
					Container: operatorv1alpha1.Container{
						ImagePullPolicy: v1.PullAlways,
					},
					IBMLicenseServiceBaseSpec: operatorv1alpha1.IBMLicenseServiceBaseSpec{
						ImagePullSecrets: []string{"artifactory-token"},
					},
					License: &operatorv1alpha1.License{
						Accept: true,
					},
				},
			}

			checkBasicRequirements(ctx, instance, newInstance)

			By("Checking if license is accepted")
			Eventually(func() bool {
				Expect(k8sClient.Get(ctx, types.NamespacedName{Name: instance.Name}, newInstance)).Should(Succeed())
				return newInstance.Spec.IsLicenseAccepted()
			}, timeout, interval).Should(Equal(true))
		})

		It("Should create IBMLicensing instance HTTPS", func() {
			By("Creating the IBMLicensing")
			newInstance := &operatorv1alpha1.IBMLicensing{}

			instance = &operatorv1alpha1.IBMLicensing{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Spec: operatorv1alpha1.IBMLicensingSpec{
					InstanceNamespace: namespace,
					Datasource:        "datacollector",
					HTTPSEnable:       true,
					Container: operatorv1alpha1.Container{
						ImagePullPolicy: v1.PullAlways,
					},
					IBMLicenseServiceBaseSpec: operatorv1alpha1.IBMLicenseServiceBaseSpec{
						ImagePullSecrets: []string{"artifactory-token"},
					},
					License: &operatorv1alpha1.License{
						Accept: true,
					},
				},
			}

			checkBasicRequirements(ctx, instance, newInstance)

		})

		It("Should create IBMLicensing instance with route", func() {
			if !ocp {
				Skip("for OCP ONLY")
			}

			By("Creating the IBMLicensing")
			newInstance := &operatorv1alpha1.IBMLicensing{}
			routeEnabled := true

			instance = &operatorv1alpha1.IBMLicensing{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Spec: operatorv1alpha1.IBMLicensingSpec{
					InstanceNamespace: namespace,
					Datasource:        "datacollector",
					HTTPSEnable:       false,
					RouteEnabled:      &routeEnabled,
					Container: operatorv1alpha1.Container{
						ImagePullPolicy: v1.PullAlways,
					},
					IBMLicenseServiceBaseSpec: operatorv1alpha1.IBMLicenseServiceBaseSpec{
						ImagePullSecrets: []string{"artifactory-token"},
					},
					License: &operatorv1alpha1.License{
						Accept: true,
					},
				},
			}

			checkBasicRequirements(ctx, instance, newInstance)

			By("Checking if route exists")
			Eventually(func() bool {
				route := &routev1.Route{}
				Expect(k8sClient.Get(ctx, types.NamespacedName{Name: service.GetResourceName(instance), Namespace: namespace}, route)).Should(Succeed())
				return route != nil
			}, timeout, interval).Should(BeTrue())

			By("Checking if chargeback is disabled")
			Eventually(func() bool {
				Expect(k8sClient.Get(ctx, types.NamespacedName{Name: instance.Name}, newInstance)).Should(Succeed())
				return newInstance.Spec.IsChargebackEnabled()
			}, timeout, interval).Should(Equal(false))

		})

		It("Should create IBMLicensing with RHMP enabled", func() {
			By("Creating the IBMLicensing")
			newInstance := &operatorv1alpha1.IBMLicensing{}
			rhmpEnabled := true

			instance = &operatorv1alpha1.IBMLicensing{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Spec: operatorv1alpha1.IBMLicensingSpec{
					InstanceNamespace: namespace,
					Datasource:        "datacollector",
					HTTPSEnable:       false,
					RHMPEnabled:       &rhmpEnabled,
					Container: operatorv1alpha1.Container{
						ImagePullPolicy: v1.PullAlways,
					},
					IBMLicenseServiceBaseSpec: operatorv1alpha1.IBMLicenseServiceBaseSpec{
						ImagePullSecrets: []string{"artifactory-token"},
					},
					License: &operatorv1alpha1.License{
						Accept: true,
					},
				},
			}

			checkBasicRequirements(ctx, instance, newInstance)

			By("Checking if prometheus service exists")
			Eventually(func() bool {
				prometheusService := &v1.Service{}
				Expect(k8sClient.Get(ctx, types.NamespacedName{Name: service.GetPrometheusServiceName(), Namespace: namespace}, prometheusService)).Should(Succeed())
				return prometheusService != nil
			}, timeout, interval).Should(BeTrue())

			By("Checking if service monitor exists")
			Eventually(func() bool {
				serviceMonitor := &monitoringv1.ServiceMonitor{}
				Expect(k8sClient.Get(ctx, types.NamespacedName{Name: service.PrometheusRHMPServiceMonitor, Namespace: namespace}, serviceMonitor)).Should(Succeed())
				return serviceMonitor != nil
			}, timeout, interval).Should(BeTrue())

			By("Checking if network policy exists")
			Eventually(func() bool {
				networkPolicy := &networkingv1.NetworkPolicy{}
				Expect(k8sClient.Get(ctx, types.NamespacedName{Name: service.GetNetworkPolicyName(newInstance), Namespace: namespace}, networkPolicy)).Should(Succeed())
				return networkPolicy != nil
			}, timeout, interval).Should(BeTrue())

			By("Checking if meter definition exists for cloudpak")
			Eventually(func() bool {
				meterDefinition := &rhmp.MeterDefinition{}
				Expect(k8sClient.Get(ctx, types.NamespacedName{Name: service.GetMeterDefinitionName(newInstance, "product"), Namespace: namespace}, meterDefinition)).Should(Succeed())
				return meterDefinition != nil
			}, timeout, interval).Should(BeTrue())

			By("Checking if meter definition exists for bundle product")
			Eventually(func() bool {
				meterDefinition := &rhmp.MeterDefinition{}
				Expect(k8sClient.Get(ctx, types.NamespacedName{Name: service.GetMeterDefinitionName(newInstance, "bundleproduct"), Namespace: namespace}, meterDefinition)).Should(Succeed())
				return meterDefinition != nil
			}, timeout, interval).Should(BeTrue())

			By("Checking if meter definition exists for chargeback")
			Eventually(func() bool {
				meterDefinition := &rhmp.MeterDefinition{}
				Expect(k8sClient.Get(ctx, types.NamespacedName{Name: service.GetMeterDefinitionName(newInstance, "chargeback"), Namespace: namespace}, meterDefinition)).Should(Succeed())
				return meterDefinition != nil
			}, timeout, interval).Should(BeTrue())

			By("Checking if chargeback is enabled")
			Eventually(func() bool {
				Expect(k8sClient.Get(ctx, types.NamespacedName{Name: instance.Name}, newInstance)).Should(Succeed())
				return newInstance.Spec.IsChargebackEnabled()
			}, timeout, interval).Should(Equal(true))
		})
	})

	Context("Gateway API reconciliation", func() {
		It("Should create Gateway resources when gateway is enabled", func() {
			By("Creating IBMLicensing with gateway enabled")
			newInstance := &operatorv1alpha1.IBMLicensing{}
			gatewayEnabled := true

			instance = &operatorv1alpha1.IBMLicensing{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Spec: operatorv1alpha1.IBMLicensingSpec{
					InstanceNamespace: namespace,
					Datasource:        "datacollector",
					HTTPSEnable:       true,
					GatewayEnabled:    &gatewayEnabled,
					Container: operatorv1alpha1.Container{
						ImagePullPolicy: v1.PullAlways,
					},
					IBMLicenseServiceBaseSpec: operatorv1alpha1.IBMLicenseServiceBaseSpec{
						ImagePullSecrets: []string{"artifactory-token"},
					},
					License: &operatorv1alpha1.License{
						Accept: true,
					},
				},
			}

			checkBasicRequirements(ctx, instance, newInstance)

			By("Checking if Gateway exists")
			Eventually(func() bool {
				gateway := &gatewayv1.Gateway{}
				err := k8sClient.Get(ctx, types.NamespacedName{Name: service.GatewayName, Namespace: namespace}, gateway)
				return err == nil && gateway != nil
			}, timeout, interval).Should(BeTrue())

			By("Verifying Gateway spec")
			Eventually(func() bool {
				gateway := &gatewayv1.Gateway{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: service.GatewayName, Namespace: namespace}, gateway); err != nil {
					return false
				}
				return string(gateway.Spec.GatewayClassName) == service.DefaultGatewayClassName &&
					len(gateway.Spec.Listeners) > 0
			}, timeout, interval).Should(BeTrue())

			By("Checking if HTTPRoute exists")
			Eventually(func() bool {
				httpRoute := &gatewayv1.HTTPRoute{}
				err := k8sClient.Get(ctx, types.NamespacedName{Name: service.HTTPRouteName, Namespace: namespace}, httpRoute)
				return err == nil && httpRoute != nil
			}, timeout, interval).Should(BeTrue())

			By("Verifying HTTPRoute spec")
			Eventually(func() bool {
				httpRoute := &gatewayv1.HTTPRoute{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: service.HTTPRouteName, Namespace: namespace}, httpRoute); err != nil {
					return false
				}
				return len(httpRoute.Spec.ParentRefs) > 0 &&
					string(httpRoute.Spec.ParentRefs[0].Name) == service.GatewayName &&
					len(httpRoute.Spec.Rules) > 0
			}, timeout, interval).Should(BeTrue())

			By("Checking if BackendTLSPolicy exists")
			Eventually(func() bool {
				policy := &gatewayv1.BackendTLSPolicy{}
				err := k8sClient.Get(ctx, types.NamespacedName{Name: service.BackendTLSPolicyName, Namespace: namespace}, policy)
				return err == nil && policy != nil
			}, timeout, interval).Should(BeTrue())

			By("Verifying BackendTLSPolicy spec")
			Eventually(func() bool {
				policy := &gatewayv1.BackendTLSPolicy{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: service.BackendTLSPolicyName, Namespace: namespace}, policy); err != nil {
					return false
				}
				return len(policy.Spec.TargetRefs) > 0 &&
					string(policy.Spec.TargetRefs[0].Name) == service.GetLicensingServiceName(instance)
			}, timeout, interval).Should(BeTrue())

			By("Checking if Gateway ConfigMap exists")
			Eventually(func() bool {
				configMap := &v1.ConfigMap{}
				err := k8sClient.Get(ctx, types.NamespacedName{Name: service.GatewayConfigMapName, Namespace: namespace}, configMap)
				return err == nil && configMap != nil
			}, timeout, interval).Should(BeTrue())

			By("Verifying Gateway ConfigMap contains certificate data")
			Eventually(func() bool {
				configMap := &v1.ConfigMap{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: service.GatewayConfigMapName, Namespace: namespace}, configMap); err != nil {
					return false
				}
				_, exists := configMap.Data["ca.crt"]
				return exists
			}, timeout, interval).Should(BeTrue())
		})

		It("Should create Gateway with custom options", func() {
			By("Creating IBMLicensing with custom gateway options")
			newInstance := &operatorv1alpha1.IBMLicensing{}
			gatewayEnabled := true
			customPort := int32(8443)
			customClassName := "custom-gateway-class"
			customTLSSecret := "custom-tls-secret"

			instance = &operatorv1alpha1.IBMLicensing{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Spec: operatorv1alpha1.IBMLicensingSpec{
					InstanceNamespace: namespace,
					Datasource:        "datacollector",
					HTTPSEnable:       true,
					GatewayEnabled:    &gatewayEnabled,
					GatewayOptions: &operatorv1alpha1.IBMLicensingGatewayOptions{
						GatewayClassName: customClassName,
						HTTPSPort:        &customPort,
						TLSSecretName:    customTLSSecret,
					},
					Container: operatorv1alpha1.Container{
						ImagePullPolicy: v1.PullAlways,
					},
					IBMLicenseServiceBaseSpec: operatorv1alpha1.IBMLicenseServiceBaseSpec{
						ImagePullSecrets: []string{"artifactory-token"},
					},
					License: &operatorv1alpha1.License{
						Accept: true,
					},
				},
			}

			checkBasicRequirements(ctx, instance, newInstance)

			By("Verifying Gateway uses custom class name")
			Eventually(func() bool {
				gateway := &gatewayv1.Gateway{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: service.GatewayName, Namespace: namespace}, gateway); err != nil {
					return false
				}
				return string(gateway.Spec.GatewayClassName) == customClassName
			}, timeout, interval).Should(BeTrue())

			By("Verifying Gateway uses custom HTTPS port")
			Eventually(func() bool {
				gateway := &gatewayv1.Gateway{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: service.GatewayName, Namespace: namespace}, gateway); err != nil {
					return false
				}
				for _, listener := range gateway.Spec.Listeners {
					if listener.Protocol == gatewayv1.HTTPSProtocolType {
						return listener.Port == gatewayv1.PortNumber(customPort)
					}
				}
				return false
			}, timeout, interval).Should(BeTrue())

			By("Verifying Gateway uses custom TLS secret")
			Eventually(func() bool {
				gateway := &gatewayv1.Gateway{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: service.GatewayName, Namespace: namespace}, gateway); err != nil {
					return false
				}
				for _, listener := range gateway.Spec.Listeners {
					if listener.TLS != nil && len(listener.TLS.CertificateRefs) > 0 {
						return string(listener.TLS.CertificateRefs[0].Name) == customTLSSecret
					}
				}
				return false
			}, timeout, interval).Should(BeTrue())
		})

		It("Should merge annotations correctly for Gateway resources", func() {
			By("Creating IBMLicensing with spec and gateway annotations")
			newInstance := &operatorv1alpha1.IBMLicensing{}
			gatewayEnabled := true

			instance = &operatorv1alpha1.IBMLicensing{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Spec: operatorv1alpha1.IBMLicensingSpec{
					InstanceNamespace: namespace,
					Datasource:        "datacollector",
					HTTPSEnable:       true,
					GatewayEnabled:    &gatewayEnabled,
					Annotations: map[string]string{
						"spec-annotation": "spec-value",
						"shared-key":      "spec-value",
					},
					GatewayOptions: &operatorv1alpha1.IBMLicensingGatewayOptions{
						Annotations: map[string]string{
							"gateway-annotation": "gateway-value",
							"shared-key":         "gateway-value", // Should override spec value
						},
					},
					Container: operatorv1alpha1.Container{
						ImagePullPolicy: v1.PullAlways,
					},
					IBMLicenseServiceBaseSpec: operatorv1alpha1.IBMLicenseServiceBaseSpec{
						ImagePullSecrets: []string{"artifactory-token"},
					},
					License: &operatorv1alpha1.License{
						Accept: true,
					},
				},
			}

			checkBasicRequirements(ctx, instance, newInstance)

			By("Verifying Gateway has merged annotations with gateway options taking precedence")
			Eventually(func() bool {
				gateway := &gatewayv1.Gateway{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: service.GatewayName, Namespace: namespace}, gateway); err != nil {
					return false
				}
				annotations := gateway.GetAnnotations()
				return annotations["spec-annotation"] == "spec-value" &&
					annotations["gateway-annotation"] == "gateway-value" &&
					annotations["shared-key"] == "gateway-value" // Gateway option should win
			}, timeout, interval).Should(BeTrue())

			By("Verifying HTTPRoute has merged annotations")
			Eventually(func() bool {
				httpRoute := &gatewayv1.HTTPRoute{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: service.HTTPRouteName, Namespace: namespace}, httpRoute); err != nil {
					return false
				}
				annotations := httpRoute.GetAnnotations()
				return annotations["spec-annotation"] == "spec-value" &&
					annotations["gateway-annotation"] == "gateway-value" &&
					annotations["shared-key"] == "gateway-value"
			}, timeout, interval).Should(BeTrue())
		})

		It("Should cleanup Gateway resources when gateway is disabled", func() {
			By("Creating IBMLicensing with gateway enabled first")
			newInstance := &operatorv1alpha1.IBMLicensing{}
			gatewayEnabled := true

			instance = &operatorv1alpha1.IBMLicensing{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Spec: operatorv1alpha1.IBMLicensingSpec{
					InstanceNamespace: namespace,
					Datasource:        "datacollector",
					HTTPSEnable:       true,
					GatewayEnabled:    &gatewayEnabled,
					Container: operatorv1alpha1.Container{
						ImagePullPolicy: v1.PullAlways,
					},
					IBMLicenseServiceBaseSpec: operatorv1alpha1.IBMLicenseServiceBaseSpec{
						ImagePullSecrets: []string{"artifactory-token"},
					},
					License: &operatorv1alpha1.License{
						Accept: true,
					},
				},
			}

			checkBasicRequirements(ctx, instance, newInstance)

			By("Verifying Gateway resources exist")
			Eventually(func() bool {
				gateway := &gatewayv1.Gateway{}
				return k8sClient.Get(ctx, types.NamespacedName{Name: service.GatewayName, Namespace: namespace}, gateway) == nil
			}, timeout, interval).Should(BeTrue())

			By("Disabling gateway")
			gatewayEnabled = false
			Expect(k8sClient.Get(ctx, types.NamespacedName{Name: instance.Name}, newInstance)).Should(Succeed())
			newInstance.Spec.GatewayEnabled = &gatewayEnabled
			Expect(k8sClient.Update(ctx, newInstance)).Should(Succeed())

			By("Verifying Gateway is deleted")
			Eventually(func() bool {
				gateway := &gatewayv1.Gateway{}
				err := k8sClient.Get(ctx, types.NamespacedName{Name: service.GatewayName, Namespace: namespace}, gateway)
				return err != nil
			}, timeout, interval).Should(BeTrue())

			By("Verifying HTTPRoute is deleted")
			Eventually(func() bool {
				httpRoute := &gatewayv1.HTTPRoute{}
				err := k8sClient.Get(ctx, types.NamespacedName{Name: service.HTTPRouteName, Namespace: namespace}, httpRoute)
				return err != nil
			}, timeout, interval).Should(BeTrue())

			By("Verifying BackendTLSPolicy is deleted")
			Eventually(func() bool {
				policy := &gatewayv1.BackendTLSPolicy{}
				err := k8sClient.Get(ctx, types.NamespacedName{Name: service.BackendTLSPolicyName, Namespace: namespace}, policy)
				return err != nil
			}, timeout, interval).Should(BeTrue())

			By("Verifying Gateway ConfigMap is deleted")
			Eventually(func() bool {
				configMap := &v1.ConfigMap{}
				err := k8sClient.Get(ctx, types.NamespacedName{Name: service.GatewayConfigMapName, Namespace: namespace}, configMap)
				return err != nil
			}, timeout, interval).Should(BeTrue())
		})

		It("Should handle Gateway with default values when options are nil", func() {
			By("Creating IBMLicensing with gateway enabled but no options")
			newInstance := &operatorv1alpha1.IBMLicensing{}
			gatewayEnabled := true

			instance = &operatorv1alpha1.IBMLicensing{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Spec: operatorv1alpha1.IBMLicensingSpec{
					InstanceNamespace: namespace,
					Datasource:        "datacollector",
					HTTPSEnable:       true,
					GatewayEnabled:    &gatewayEnabled,
					GatewayOptions:    nil, // Explicitly nil
					Container: operatorv1alpha1.Container{
						ImagePullPolicy: v1.PullAlways,
					},
					IBMLicenseServiceBaseSpec: operatorv1alpha1.IBMLicenseServiceBaseSpec{
						ImagePullSecrets: []string{"artifactory-token"},
					},
					License: &operatorv1alpha1.License{
						Accept: true,
					},
				},
			}

			checkBasicRequirements(ctx, instance, newInstance)

			By("Verifying Gateway uses default class name")
			Eventually(func() bool {
				gateway := &gatewayv1.Gateway{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: service.GatewayName, Namespace: namespace}, gateway); err != nil {
					return false
				}
				return string(gateway.Spec.GatewayClassName) == service.DefaultGatewayClassName
			}, timeout, interval).Should(BeTrue())

			By("Verifying Gateway uses default HTTPS port (443)")
			Eventually(func() bool {
				gateway := &gatewayv1.Gateway{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: service.GatewayName, Namespace: namespace}, gateway); err != nil {
					return false
				}
				for _, listener := range gateway.Spec.Listeners {
					if listener.Protocol == gatewayv1.HTTPSProtocolType {
						return listener.Port == gatewayv1.PortNumber(443)
					}
				}
				return false
			}, timeout, interval).Should(BeTrue())

			By("Verifying Gateway uses default TLS secret name")
			Eventually(func() bool {
				gateway := &gatewayv1.Gateway{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: service.GatewayName, Namespace: namespace}, gateway); err != nil {
					return false
				}
				for _, listener := range gateway.Spec.Listeners {
					if listener.TLS != nil && len(listener.TLS.CertificateRefs) > 0 {
						return string(listener.TLS.CertificateRefs[0].Name) == "ibm-license-service-cert-internal"
					}
				}
				return false
			}, timeout, interval).Should(BeTrue())
		})

		It("Should apply custom labels to Gateway resources", func() {
			By("Creating IBMLicensing with custom labels")
			newInstance := &operatorv1alpha1.IBMLicensing{}
			gatewayEnabled := true

			instance = &operatorv1alpha1.IBMLicensing{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Spec: operatorv1alpha1.IBMLicensingSpec{
					InstanceNamespace: namespace,
					Datasource:        "datacollector",
					HTTPSEnable:       true,
					GatewayEnabled:    &gatewayEnabled,
					Labels: map[string]string{
						"custom-label":     "custom-value",
						"environment":      "test",
						"team":             "platform",
					},
					Container: operatorv1alpha1.Container{
						ImagePullPolicy: v1.PullAlways,
					},
					IBMLicenseServiceBaseSpec: operatorv1alpha1.IBMLicenseServiceBaseSpec{
						ImagePullSecrets: []string{"artifactory-token"},
					},
					License: &operatorv1alpha1.License{
						Accept: true,
					},
				},
			}

			checkBasicRequirements(ctx, instance, newInstance)

			By("Verifying Gateway has custom labels")
			Eventually(func() bool {
				gateway := &gatewayv1.Gateway{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: service.GatewayName, Namespace: namespace}, gateway); err != nil {
					return false
				}
				labels := gateway.GetLabels()
				return labels["custom-label"] == "custom-value" &&
					labels["environment"] == "test" &&
					labels["team"] == "platform"
			}, timeout, interval).Should(BeTrue())

			By("Verifying Gateway has standard labels")
			Eventually(func() bool {
				gateway := &gatewayv1.Gateway{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: service.GatewayName, Namespace: namespace}, gateway); err != nil {
					return false
				}
				labels := gateway.GetLabels()
				_, hasAppLabel := labels["app"]
				_, hasReleaseLabel := labels["release"]
				return hasAppLabel && hasReleaseLabel
			}, timeout, interval).Should(BeTrue())

			By("Verifying HTTPRoute has custom labels")
			Eventually(func() bool {
				httpRoute := &gatewayv1.HTTPRoute{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: service.HTTPRouteName, Namespace: namespace}, httpRoute); err != nil {
					return false
				}
				labels := httpRoute.GetLabels()
				return labels["custom-label"] == "custom-value" &&
					labels["environment"] == "test" &&
					labels["team"] == "platform"
			}, timeout, interval).Should(BeTrue())

			By("Verifying BackendTLSPolicy has custom labels")
			Eventually(func() bool {
				policy := &gatewayv1.BackendTLSPolicy{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: service.BackendTLSPolicyName, Namespace: namespace}, policy); err != nil {
					return false
				}
				labels := policy.GetLabels()
				return labels["custom-label"] == "custom-value" &&
					labels["environment"] == "test" &&
					labels["team"] == "platform"
			}, timeout, interval).Should(BeTrue())

			By("Verifying Gateway ConfigMap has custom labels")
			Eventually(func() bool {
				configMap := &v1.ConfigMap{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: service.GatewayConfigMapName, Namespace: namespace}, configMap); err != nil {
					return false
				}
				labels := configMap.GetLabels()
				return labels["custom-label"] == "custom-value" &&
					labels["environment"] == "test" &&
					labels["team"] == "platform"
			}, timeout, interval).Should(BeTrue())
		})

		It("Should apply both labels and annotations to Gateway resources", func() {
			By("Creating IBMLicensing with both custom labels and annotations")
			newInstance := &operatorv1alpha1.IBMLicensing{}
			gatewayEnabled := true

			instance = &operatorv1alpha1.IBMLicensing{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Spec: operatorv1alpha1.IBMLicensingSpec{
					InstanceNamespace: namespace,
					Datasource:        "datacollector",
					HTTPSEnable:       true,
					GatewayEnabled:    &gatewayEnabled,
					Labels: map[string]string{
						"label-key": "label-value",
					},
					Annotations: map[string]string{
						"annotation-key": "annotation-value",
					},
					GatewayOptions: &operatorv1alpha1.IBMLicensingGatewayOptions{
						Annotations: map[string]string{
							"gateway-specific-annotation": "gateway-specific-value",
						},
					},
					Container: operatorv1alpha1.Container{
						ImagePullPolicy: v1.PullAlways,
					},
					IBMLicenseServiceBaseSpec: operatorv1alpha1.IBMLicenseServiceBaseSpec{
						ImagePullSecrets: []string{"artifactory-token"},
					},
					License: &operatorv1alpha1.License{
						Accept: true,
					},
				},
			}

			checkBasicRequirements(ctx, instance, newInstance)

			By("Verifying Gateway has both labels and annotations")
			Eventually(func() bool {
				gateway := &gatewayv1.Gateway{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: service.GatewayName, Namespace: namespace}, gateway); err != nil {
					return false
				}
				labels := gateway.GetLabels()
				annotations := gateway.GetAnnotations()
				return labels["label-key"] == "label-value" &&
					annotations["annotation-key"] == "annotation-value" &&
					annotations["gateway-specific-annotation"] == "gateway-specific-value"
			}, timeout, interval).Should(BeTrue())

			By("Verifying HTTPRoute has both labels and annotations")
			Eventually(func() bool {
				httpRoute := &gatewayv1.HTTPRoute{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: service.HTTPRouteName, Namespace: namespace}, httpRoute); err != nil {
					return false
				}
				labels := httpRoute.GetLabels()
				annotations := httpRoute.GetAnnotations()
				return labels["label-key"] == "label-value" &&
					annotations["annotation-key"] == "annotation-value" &&
					annotations["gateway-specific-annotation"] == "gateway-specific-value"
			}, timeout, interval).Should(BeTrue())
		})

		It("Should update Gateway resources when labels or annotations change", func() {
			By("Creating IBMLicensing with initial labels and annotations")
			newInstance := &operatorv1alpha1.IBMLicensing{}
			gatewayEnabled := true

			instance = &operatorv1alpha1.IBMLicensing{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Spec: operatorv1alpha1.IBMLicensingSpec{
					InstanceNamespace: namespace,
					Datasource:        "datacollector",
					HTTPSEnable:       true,
					GatewayEnabled:    &gatewayEnabled,
					Labels: map[string]string{
						"version": "v1",
					},
					Annotations: map[string]string{
						"description": "initial",
					},
					Container: operatorv1alpha1.Container{
						ImagePullPolicy: v1.PullAlways,
					},
					IBMLicenseServiceBaseSpec: operatorv1alpha1.IBMLicenseServiceBaseSpec{
						ImagePullSecrets: []string{"artifactory-token"},
					},
					License: &operatorv1alpha1.License{
						Accept: true,
					},
				},
			}

			checkBasicRequirements(ctx, instance, newInstance)

			By("Verifying initial labels and annotations")
			Eventually(func() bool {
				gateway := &gatewayv1.Gateway{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: service.GatewayName, Namespace: namespace}, gateway); err != nil {
					return false
				}
				labels := gateway.GetLabels()
				annotations := gateway.GetAnnotations()
				return labels["version"] == "v1" && annotations["description"] == "initial"
			}, timeout, interval).Should(BeTrue())

			By("Updating labels and annotations")
			Expect(k8sClient.Get(ctx, types.NamespacedName{Name: instance.Name}, newInstance)).Should(Succeed())
			newInstance.Spec.Labels = map[string]string{
				"version": "v2",
				"updated": "true",
			}
			newInstance.Spec.Annotations = map[string]string{
				"description": "updated",
				"timestamp":   "2026-03-24",
			}
			Expect(k8sClient.Update(ctx, newInstance)).Should(Succeed())

			By("Verifying updated labels and annotations on Gateway")
			Eventually(func() bool {
				gateway := &gatewayv1.Gateway{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: service.GatewayName, Namespace: namespace}, gateway); err != nil {
					return false
				}
				labels := gateway.GetLabels()
				annotations := gateway.GetAnnotations()
				return labels["version"] == "v2" &&
					labels["updated"] == "true" &&
					annotations["description"] == "updated" &&
					annotations["timestamp"] == "2026-03-24"
			}, timeout, interval).Should(BeTrue())

			By("Verifying updated labels and annotations on HTTPRoute")
			Eventually(func() bool {
				httpRoute := &gatewayv1.HTTPRoute{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: service.HTTPRouteName, Namespace: namespace}, httpRoute); err != nil {
					return false
				}
				labels := httpRoute.GetLabels()
				annotations := httpRoute.GetAnnotations()
				return labels["version"] == "v2" &&
					labels["updated"] == "true" &&
					annotations["description"] == "updated" &&
					annotations["timestamp"] == "2026-03-24"
			}, timeout, interval).Should(BeTrue())
		})
	})
})

func checkBasicRequirements(ctx context.Context, instance, newInstance *operatorv1alpha1.IBMLicensing) {
	Expect(k8sClient.Create(ctx, instance)).Should(Succeed())

	Eventually(func() int {
		k8sClient.Get(ctx, types.NamespacedName{Name: instance.Name}, newInstance)
		return len(newInstance.Status.LicensingPods)
	}, timeout, interval).Should(BeNumerically(">", 0))

	By("Checking status of the IBMLicensing")
	Eventually(func() v1.PodPhase {
		Expect(k8sClient.Get(ctx, types.NamespacedName{Name: instance.Name}, newInstance)).Should(Succeed())
		return newInstance.Status.LicensingPods[0].Phase
	}, timeout, interval).Should(Equal(v1.PodRunning))

	By("Checking if licensing-service exists")
	Eventually(func() bool {
		licensingService := &v1.Service{}
		Expect(k8sClient.Get(ctx, types.NamespacedName{Name: service.GetLicensingServiceName(instance), Namespace: namespace}, licensingService)).Should(Succeed())
		return licensingService != nil
	}, timeout, interval).Should(BeTrue())
}
