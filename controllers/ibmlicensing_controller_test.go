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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestCheckReconcileLicensing(t *testing.T) {
}

// +kubebuilder:docs-gen:collapse=Imports

var _ = Describe("IBMLicensing controller", func() {
	const (
		name      = "instance-unitest"
		namespace = "ibm-common-services-unitest"
	)

	var (
		ctx context.Context

		instance  *operatorv1alpha1.IBMLicensing
		configKey types.NamespacedName
	)

	BeforeEach(func() {
		ctx = context.Background()

		By("Creating the Namespace")
		Expect(k8sClient.Create(ctx, NamespaceObj(namespace))).Should(Succeed())

	})

	AfterEach(func() {
		By("Deleting the Namespace")
		Expect(k8sClient.Delete(ctx, NamespaceObj(namespace))).Should(Succeed())

	})

	Context("Initializing IBMLicensing Status", func() {
		It("Should the status of IBMLicensing be Running", func() {
			By("Creating broken IBMLicensing")
			instance = IBMLicensingObj(name, namespace, "")
			Expect(k8sClient.Create(ctx, instance)).Should(MatchError(ContainSubstring("spec.datasource")))

			By("Creating broken IBMLicensing")
			instance = IBMLicensingObj(name, namespace, "datacollector1")
			Expect(k8sClient.Create(ctx, instance)).Should(MatchError(ContainSubstring("spec.datasource")))

			By("Creating the IBMLicensing")
			instance = IBMLicensingObj(name, namespace, "datacollector")
			Expect(k8sClient.Create(ctx, instance)).Should(Succeed())

			By("Checking status of the IBMLicensing")
			Eventually(func() operatorv1alpha1.IBMLicensingStatus {
				newInstance := &operatorv1alpha1.IBMLicensing{}

				configKey = types.NamespacedName{
					Name:      name,
					Namespace: namespace,
				}
				Expect(k8sClient.Get(ctx, configKey, newInstance)).Should(Succeed())

				return newInstance.Status
			}, timeout, interval).Should(Equal(operatorv1alpha1.IBMLicensingStatus{}))

			By("Cleaning up resources")
			Expect(k8sClient.Delete(ctx, instance)).Should(Succeed())
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

func NamespaceObj(name string) *corev1.Namespace {
	return &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}
