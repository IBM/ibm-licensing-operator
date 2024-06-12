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

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	routev1 "github.com/openshift/api/route/v1"
	rhmp "github.com/redhat-marketplace/redhat-marketplace-operator/v2/apis/marketplace/v1beta1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	operatorv1alpha1 "github.com/IBM/ibm-licensing-operator/api/v1alpha1"
	"github.com/IBM/ibm-licensing-operator/controllers/resources/service"
)

var _ = Describe("IBMLicensing controller", Ordered, func() {
	const (
		name = "instance-test"
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
