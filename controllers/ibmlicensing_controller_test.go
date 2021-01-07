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

package controllers

import (
	"context"
	"testing"

	operatorv1alpha1 "github.com/ibm/ibm-licensing-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestCheckReconcileLicensing(t *testing.T) {
}

var _ = Describe("IBMLicensing controller", func() {
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
				},
			}
			Expect(k8sClient.Create(ctx, instance)).Should(MatchError(ContainSubstring("spec.datasource")))
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
				},
			}

			Expect(k8sClient.Create(ctx, instance)).Should(Succeed())

			Eventually(func() int {
				k8sClient.Get(ctx, types.NamespacedName{Name: name}, newInstance)
				return len(newInstance.Status.LicensingPods)
			}, timeout, interval).Should(Equal(1))

			By("Checking status of the IBMLicensing")
			Eventually(func() v1.PodPhase {
				Expect(k8sClient.Get(ctx, types.NamespacedName{Name: name}, newInstance)).Should(Succeed())
				return newInstance.Status.LicensingPods[0].Phase
			}, timeout, interval).Should(Equal(v1.PodRunning))

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
				},
			}

			Expect(k8sClient.Create(ctx, instance)).Should(Succeed())

			Eventually(func() int {
				k8sClient.Get(ctx, types.NamespacedName{Name: name}, newInstance)
				return len(newInstance.Status.LicensingPods)
			}, timeout, interval).Should(Equal(1))

			By("Checking status of the IBMLicensing")
			Eventually(func() v1.PodPhase {
				Expect(k8sClient.Get(ctx, types.NamespacedName{Name: name}, newInstance)).Should(Succeed())
				return newInstance.Status.LicensingPods[0].Phase
			}, timeout, interval).Should(Equal(v1.PodRunning))

		})
	})
})
