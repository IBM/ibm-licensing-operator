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
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	operatorv1alpha1 "github.com/IBM/ibm-licensing-operator/api/v1alpha1"

	"k8s.io/apimachinery/pkg/types"
)

func TestCheckReconcileLicenseReporter(t *testing.T) {
}

var _ = Describe("IBMLicenseServiceReporter controller", func() {
	const (
		name = "instance-rep-test"
	)

	var (
		ctx               context.Context
		instance          *operatorv1alpha1.IBMLicenseServiceReporter
		instanceForRemove = &operatorv1alpha1.IBMLicenseServiceReporter{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			}}
	)

	BeforeEach(func() {
		ctx = context.Background()
		k8sClient.Delete(ctx, instanceForRemove)
	})

	AfterEach(func() {
		k8sClient.Delete(ctx, instanceForRemove)
	})

	Context("Initializing IBMLicenseServiceReporter Status", func() {
		It("Should create IBMLicenseServiceReporter", func() {
			if !ocp {
				Skip("for OCP ONLY")
			}
			By("Creating the IBMLicenseServiceReporter")
			newInstance := &operatorv1alpha1.IBMLicenseServiceReporter{}

			instance = &operatorv1alpha1.IBMLicenseServiceReporter{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: operatorv1alpha1.IBMLicenseServiceReporterSpec{
					ReceiverContainer: operatorv1alpha1.Container{
						ImagePullPolicy: v1.PullAlways,
					},
					ReporterUIContainer: operatorv1alpha1.Container{
						ImagePullPolicy: v1.PullAlways,
					},
					DatabaseContainer: operatorv1alpha1.Container{
						ImagePullPolicy: v1.PullAlways,
					},
					IBMLicenseServiceBaseSpec: operatorv1alpha1.IBMLicenseServiceBaseSpec{
						ImagePullSecrets: []string{"artifactory-token"},
					},
				},
			}

			Expect(k8sClient.Create(ctx, instance)).Should(Succeed())

			Eventually(func() int {
				k8sClient.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, newInstance)
				return len(newInstance.Status.LicensingReporterPods)
			}, timeout, interval).Should(BeNumerically(">", 0))

			By("Checking status of the IBMLicensing is Pending because of UI")
			Eventually(func() v1.PodPhase {
				Expect(k8sClient.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, newInstance)).Should(Succeed())
				return newInstance.Status.LicensingReporterPods[0].Phase
			}, timeout, interval).Should(Equal(v1.PodRunning))
		})
	})
})
