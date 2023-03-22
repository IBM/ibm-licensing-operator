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
	"reflect"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	res "github.com/IBM/ibm-licensing-operator/controllers/resources"
	svcres "github.com/IBM/ibm-licensing-operator/controllers/resources/service"
	odlm "github.com/IBM/operand-deployment-lifecycle-manager/api/v1alpha1"
)

func TestCheckOperandRequestReconciler(t *testing.T) {
}

var _ = Describe("OperandRequest controller", func() {
	const (
		name               = "operandrequest-test"
		operandRequestName = "ibm-licensing-opreq-1"
		lsBindInfoName     = "ibm-licensing-bindinfo"
		secret1Name        = lsBindInfoName + svcres.LicensingToken
		secret2Name        = lsBindInfoName + svcres.LicensingUploadToken
		cm1Name            = lsBindInfoName + svcres.LicensingInfo
		cm2Name            = lsBindInfoName + svcres.LicensingUploadConfig
	)

	var (
		ctx            context.Context
		operandRequest = res.OperandRequestObj(operandRequestName, opreqNamespace, res.OperatorName)
	)

	Context("(Setup) OperandRequest instance", func() {
		It("Should be created in namespace "+opreqNamespace, func() {

			Eventually(func() bool {
				operandRequestInstance := odlm.OperandRequest{}
				Expect(k8sClient.Create(ctx, operandRequest)).Should(Succeed())
				Expect(k8sClient.Get(ctx, types.NamespacedName{Name: operandRequestName, Namespace: opreqNamespace}, &operandRequestInstance)).Should(Succeed())
				return operandRequestInstance.Name == operandRequestName && reflect.DeepEqual(operandRequestInstance.Spec.Requests, operandRequest.Spec.Requests)
			}, timeout, interval).Should(BeTrue())
		})
	})

	Context("OperandRequest controller reconciliation", func() {
		It("Should expose connection details to the OperandRequest's namespace", func() {

			By("Copying " + svcres.LicensingToken + " Secret to OperandRequest's namespace")
			Eventually(func() bool {
				secret1 := corev1.Secret{}
				secret1Copy := corev1.Secret{}

				Expect(k8sClient.Get(ctx, types.NamespacedName{Name: svcres.LicensingToken, Namespace: namespace}, &secret1)).Should(Succeed())
				Expect(k8sClient.Get(ctx, types.NamespacedName{Name: secret1Name, Namespace: opreqNamespace}, &secret1Copy)).Should(Succeed())

				return res.CompareSecrets(&secret1, &secret1Copy)
			}, timeout, interval).Should(BeTrue())

			By("Copying " + svcres.LicensingUploadToken + " Secret to OperandRequest's namespace")
			Eventually(func() bool {
				secret2 := corev1.Secret{}
				secret2Copy := corev1.Secret{}

				Expect(k8sClient.Get(ctx, types.NamespacedName{Name: svcres.LicensingUploadToken, Namespace: namespace}, &secret2)).Should(Succeed())
				Expect(k8sClient.Get(ctx, types.NamespacedName{Name: secret2Name, Namespace: opreqNamespace}, &secret2Copy)).Should(Succeed())

				return res.CompareSecrets(&secret2, &secret2Copy)
			}, timeout, interval).Should(BeTrue())

			By("Copying " + svcres.LicensingInfo + " ConfigMap to OperandRequest's namespace")
			Eventually(func() bool {
				cm1 := corev1.ConfigMap{}
				cm1Copy := corev1.ConfigMap{}

				Expect(k8sClient.Get(ctx, types.NamespacedName{Name: svcres.LicensingInfo, Namespace: namespace}, &cm1)).Should(Succeed())
				Expect(k8sClient.Get(ctx, types.NamespacedName{Name: cm1Name, Namespace: opreqNamespace}, &cm1Copy)).Should(Succeed())

				return res.CompareConfigMap(&cm1, &cm1Copy)
			}, timeout, interval).Should(BeTrue())

			By("Copying " + svcres.LicensingUploadConfig + " ConfigMap to OperandRequest's namespace")
			Eventually(func() bool {
				cm2 := corev1.ConfigMap{}
				cm2Copy := corev1.ConfigMap{}

				Expect(k8sClient.Get(ctx, types.NamespacedName{Name: svcres.LicensingUploadConfig, Namespace: namespace}, &cm2)).Should(Succeed())
				Expect(k8sClient.Get(ctx, types.NamespacedName{Name: cm1Name, Namespace: opreqNamespace}, &cm2Copy)).Should(Succeed())

				return res.CompareConfigMap(&cm2, &cm2Copy)
			}, timeout, interval).Should(BeTrue())
		})
	})
})
