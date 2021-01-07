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
	"time"

	operatorv1alpha1 "github.com/ibm/ibm-licensing-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestCheckReconcileLicensing(t *testing.T) {
}

// +kubebuilder:docs-gen:collapse=Imports

var _ = Describe("IBMLicensing controller", func() {
	const (
		name      = "instance-test"
		namespace = "ibm-common-services"
	)

	var (
		ctx context.Context

		instance  *operatorv1alpha1.IBMLicensing
		configKey types.NamespacedName
	)

	BeforeEach(func() {
		ctx = context.Background()
		instance = IBMLicensingObj(name, namespace, "datacollector")
	})

	AfterEach(func() {
		By("Cleaning up resources")
		k8sClient.Delete(ctx, instance)
	})

	Context("Initializing IBMLicensing Status", func() {
		It("Should not create IBMLicensing instance", func() {
			By("Creating broken IBMLicensing")
			instance = IBMLicensingObj(name, namespace, "")
			Expect(k8sClient.Create(ctx, instance)).Should(MatchError(ContainSubstring("spec.datasource")))

			By("Creating broken IBMLicensing")
			instance = IBMLicensingObj(name, namespace, "datacollector1")
			Expect(k8sClient.Create(ctx, instance)).Should(MatchError(ContainSubstring("spec.datasource")))
		})

		It("Should create IBMLicensing instance", func() {
			By("Creating the IBMLicensing")
			Expect(k8sClient.Create(ctx, instance)).Should(Succeed())

			time.Sleep(time.Second * 5)

			By("Checking status of the IBMLicensing")
			Eventually(func() operatorv1alpha1.IBMLicensingStatus {
				newInstance := &operatorv1alpha1.IBMLicensing{}

				configKey = types.NamespacedName{
					Name: name,
				}
				Expect(k8sClient.Get(ctx, configKey, newInstance)).Should(Succeed())

				return newInstance.Status
			}, timeout, interval).Should(Equal(operatorv1alpha1.IBMLicensingStatus{}))

		})
	})
})

func IBMLicensingObj(name, namespace string, datasource string) *operatorv1alpha1.IBMLicensing {

	if datasource == "" {
		return &operatorv1alpha1.IBMLicensing{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Spec: operatorv1alpha1.IBMLicensingSpec{
				InstanceNamespace: namespace,
			},
		}
	}
	return &operatorv1alpha1.IBMLicensing{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: operatorv1alpha1.IBMLicensingSpec{
			InstanceNamespace: namespace,
			Datasource:        datasource,
		},
	}

}
