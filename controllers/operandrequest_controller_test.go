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
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	res "github.com/IBM/ibm-licensing-operator/controllers/resources"
	svcres "github.com/IBM/ibm-licensing-operator/controllers/resources/service"
)

var _ = Describe("OperandRequest controller", Ordered, func() {
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
		operandRequest = res.OperandRequestObj(operandRequestName, opreqNamespace, res.OperatorName)
		lsLabels       = map[string]string{"app": "ibm-licensing"}
		lsAnnotations  = map[string]string{"owned-by": "ibm-licensing"}
	)

	BeforeAll(func(ctx SpecContext) {
		secret := corev1.Secret{}
		if err := k8sRFromMgr.Get(ctx, types.NamespacedName{Namespace: operatorNamespace, Name: svcres.LicensingToken}, &secret); err != nil {
			Eventually(func() bool {
				lsTokenSecret := res.SecretObj(svcres.LicensingToken, operatorNamespace, map[string]string{"token": "aaaa"}, lsLabels, lsAnnotations)
				err := k8sCFromMgr.Create(ctx, &lsTokenSecret)
				return err == nil || k8serr.IsAlreadyExists(err)
			}).WithContext(ctx).Should(BeTrue())
		}

		secret = corev1.Secret{}
		if err := k8sRFromMgr.Get(ctx, types.NamespacedName{Namespace: operatorNamespace, Name: svcres.LicensingUploadToken}, &secret); err != nil {
			Eventually(func() bool {
				lsUploadToken := res.SecretObj(svcres.LicensingUploadToken, operatorNamespace, map[string]string{"token": "bbbb"}, lsLabels, lsAnnotations)
				err := k8sCFromMgr.Create(ctx, &lsUploadToken)
				return err == nil || k8serr.IsAlreadyExists(err)
			}).WithContext(ctx).Should(BeTrue())
		}

		cm := corev1.ConfigMap{}
		if err := k8sRFromMgr.Get(ctx, types.NamespacedName{Namespace: operatorNamespace, Name: svcres.LicensingUploadConfig}, &cm); err != nil {
			Eventually(func() bool {
				lsUploadConfigCm := res.ConfigMapObj(svcres.LicensingUploadConfig, operatorNamespace, map[string]string{"url": "https://ibm-licensing-service-instance"}, lsLabels, lsAnnotations)
				err := k8sCFromMgr.Create(ctx, &lsUploadConfigCm)
				return err == nil || k8serr.IsAlreadyExists(err)
			}).WithContext(ctx).Should(BeTrue())
		}

		cm = corev1.ConfigMap{}
		if err := k8sRFromMgr.Get(ctx, types.NamespacedName{Namespace: operatorNamespace, Name: svcres.LicensingInfo}, &cm); err != nil {
			Eventually(func() bool {
				lsInfoCm := res.ConfigMapObj(svcres.LicensingInfo, operatorNamespace, map[string]string{"url": "https://ibm-licensing-service-instance"}, lsLabels, lsAnnotations)
				err := k8sCFromMgr.Create(ctx, &lsInfoCm)
				return err == nil || k8serr.IsAlreadyExists(err)
			}).WithContext(ctx).Should(BeTrue())
		}

		operandRequest = res.OperandRequestObj(operandRequestName, opreqNamespace, res.OperatorName)
		err := k8sCFromMgr.Create(ctx, &operandRequest)
		Expect(err == nil || k8serr.IsAlreadyExists(err)).To(BeTrue())

	}, NodeTimeout(timeout))

	Context("OperandRequest controller reconciliation", func() {
		It("should expose connection details to the OperandRequest's namespace", func(ctx SpecContext) {

			By("Copying " + svcres.LicensingToken + " Secret to OperandRequest's namespace")
			Eventually(func() bool {
				secret1 := corev1.Secret{}
				secret1Copy := corev1.Secret{}

				if err := k8sRFromMgr.Get(ctx, types.NamespacedName{Name: svcres.LicensingToken, Namespace: namespace}, &secret1); err != nil {
					return false
				}
				if err := k8sRFromMgr.Get(ctx, types.NamespacedName{Name: secret1Name, Namespace: opreqNamespace}, &secret1Copy); err != nil {
					return false
				}

				return res.CompareSecretsData(&secret1, &secret1Copy)
			}).WithContext(ctx).Should(BeTrue())

			By("Copying " + svcres.LicensingUploadToken + " Secret to OperandRequest's namespace")
			Eventually(func() bool {
				secret2 := corev1.Secret{}
				secret2Copy := corev1.Secret{}

				if err := k8sRFromMgr.Get(ctx, types.NamespacedName{Name: svcres.LicensingUploadToken, Namespace: namespace}, &secret2); err != nil {
					return false
				}
				if err := k8sRFromMgr.Get(ctx, types.NamespacedName{Name: secret2Name, Namespace: opreqNamespace}, &secret2Copy); err != nil {
					return false
				}

				return res.CompareSecretsData(&secret2, &secret2Copy)
			}).WithContext(ctx).Should(BeTrue())

			By("Copying " + svcres.LicensingInfo + " ConfigMap to OperandRequest's namespace")
			Eventually(func() bool {
				cm1 := corev1.ConfigMap{}
				cm1Copy := corev1.ConfigMap{}

				if err := k8sRFromMgr.Get(ctx, types.NamespacedName{Name: svcres.LicensingInfo, Namespace: namespace}, &cm1); err != nil {
					return false
				}
				if err := k8sRFromMgr.Get(ctx, types.NamespacedName{Name: cm1Name, Namespace: opreqNamespace}, &cm1Copy); err != nil {
					return false
				}

				return res.CompareConfigMapData(&cm1, &cm1Copy)
			}).WithContext(ctx).Should(BeTrue())

			By("Copying " + svcres.LicensingUploadConfig + " ConfigMap to OperandRequest's namespace")
			Eventually(func() bool {
				cm2 := corev1.ConfigMap{}
				cm2Copy := corev1.ConfigMap{}

				if err := k8sRFromMgr.Get(ctx, types.NamespacedName{Name: svcres.LicensingUploadConfig, Namespace: namespace}, &cm2); err != nil {
					return false
				}
				if err := k8sRFromMgr.Get(ctx, types.NamespacedName{Name: cm2Name, Namespace: opreqNamespace}, &cm2Copy); err != nil {
					return false
				}

				return res.CompareConfigMapData(&cm2, &cm2Copy)
			}).WithContext(ctx).Should(BeTrue())
		})
	})
})
