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
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	res "github.com/IBM/ibm-licensing-operator/controllers/resources"
	svcres "github.com/IBM/ibm-licensing-operator/controllers/resources/service"
)

func TestCheckOperandRequestReconciler(t *testing.T) {
}

var _ = Describe("OperandRequest controller", func() {
	const (
		name               = "operandrequest-test"
		operandRequestName = "ibm-licensing-opreq-1"
		lsBindInfoName     = "ibm-licensing-bindinfo"
		secret1Name        = lsBindInfoName + "-" + svcres.LicensingToken
		secret2Name        = lsBindInfoName + "-" + svcres.LicensingUploadToken
		cm1Name            = lsBindInfoName + "-" + svcres.LicensingInfo
		cm2Name            = lsBindInfoName + "-" + svcres.LicensingUploadConfig
	)

	operatorNamespace, _ = os.LookupEnv("OPERATOR_NAMESPACE")
	Expect(operatorNamespace).ToNot(BeEmpty())

	opreqNamespace, _ = os.LookupEnv("OPREQ_TEST_NAMESPACE")
	Expect(opreqNamespace).ToNot(BeEmpty())

	var (
		ctx            context.Context
		operandRequest = res.OperandRequestObj(operandRequestName, opreqNamespace, res.OperatorName)
		lsLabels       = map[string]string{"app": "ibm-licensing"}
		lsAnnotations  = map[string]string{"owned-by": "ibm-licensing"}
	)

	BeforeEach(func() {
		ctx = context.Background()
	})

	Context("(Setup) Licensing ConfigMaps and Secrets", func() {
		It("IbmLicensingToken Secret should be created in namespace "+operatorNamespace, func() {
			secret := corev1.Secret{}
			if err := k8sClient.Get(ctx, types.NamespacedName{Namespace: operatorNamespace, Name: svcres.LicensingToken}, &secret); err != nil && k8serr.IsNotFound(err) {
				lsTokenSecret := res.SecretObj(svcres.LicensingToken, operatorNamespace, map[string]string{"token": "aaaa"}, lsLabels, lsAnnotations)
				Expect(k8sClient.Create(ctx, &lsTokenSecret)).Should(Succeed())
			}
		})

		It("IbmLicensingUploadToken Secret should be created in namespace "+operatorNamespace, func() {
			secret := corev1.Secret{}
			if err := k8sClient.Get(ctx, types.NamespacedName{Namespace: operatorNamespace, Name: svcres.LicensingUploadToken}, &secret); err != nil && k8serr.IsNotFound(err) {
				lsUploadToken := res.SecretObj(svcres.LicensingUploadToken, operatorNamespace, map[string]string{"token": "bbbb"}, lsLabels, lsAnnotations)
				Expect(k8sClient.Create(ctx, &lsUploadToken)).Should(Succeed())
			}

		})

		It("IbmLicensingInfo ConfigMap should be created in namespace "+operatorNamespace, func() {
			cm := corev1.ConfigMap{}
			if err := k8sClient.Get(ctx, types.NamespacedName{Namespace: operatorNamespace, Name: svcres.LicensingInfo}, &cm); err != nil && k8serr.IsNotFound(err) {
				lsInfoCm := res.ConfigMapObj(svcres.LicensingInfo, operatorNamespace, map[string]string{"url": "https://ibm-licensing-service-instance"}, lsLabels, lsAnnotations)
				Expect(k8sClient.Create(ctx, &lsInfoCm)).Should(Succeed())
			}
		})

		It("IbmLicensingUploadConfig ConfigMap should be created in namespace "+operatorNamespace, func() {
			cm := corev1.ConfigMap{}
			if err := k8sClient.Get(ctx, types.NamespacedName{Namespace: operatorNamespace, Name: svcres.LicensingUploadConfig}, &cm); err != nil && k8serr.IsNotFound(err) {
				lsUploadConfigCm := res.ConfigMapObj(svcres.LicensingUploadConfig, operatorNamespace, map[string]string{"url": "https://ibm-licensing-service-instance"}, lsLabels, lsAnnotations)
				Expect(k8sClient.Create(ctx, &lsUploadConfigCm)).Should(Succeed())
			}
		})
	})

	Context("(Setup) OperandRequest instance", func() {
		It("Should be created in namespace "+opreqNamespace, func() {
			operandRequest = res.OperandRequestObj(operandRequestName, opreqNamespace, res.OperatorName)
			Expect(k8sClient.Create(ctx, &operandRequest)).Should(Succeed())
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
