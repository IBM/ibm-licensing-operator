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
	"path/filepath"
	"testing"
	"time"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	servicecav1 "github.com/openshift/api/operator/v1"
	routev1 "github.com/openshift/api/route/v1"
	operatorframeworkv1 "github.com/operator-framework/api/pkg/operators/v1"
	meterdefv1beta1 "github.com/redhat-marketplace/redhat-marketplace-operator/v2/apis/marketplace/v1beta1"
	"go.uber.org/zap/zapcore"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	odlm "github.com/IBM/operand-deployment-lifecycle-manager/api/v1alpha1"

	operatoribmcomv1alpha1 "github.com/IBM/ibm-licensing-operator/api/v1alpha1"
	// +kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	cfg               *rest.Config
	k8sClient         client.Client
	k8sCFromMgr       client.Client
	k8sRFromMgr       client.Reader
	testEnv           *envtest.Environment
	namespace         string
	operatorNamespace string
	opreqNamespace    string
	ocp               bool
	timeout           = time.Second * 300
	interval          = time.Second * 5
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	suiteConfig, reporterConfig := GinkgoConfiguration()
	reporterConfig.FullTrace = true

	RunSpecs(t, "Controller suite", suiteConfig, reporterConfig)
}

var _ = BeforeSuite(func() {

	logf.SetLogger(zap.New(func(o *zap.Options) {
		o.Development = true
		o.TimeEncoder = zapcore.RFC3339TimeEncoder
		o.DestWriter = GinkgoWriter
	}))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("..", "config", "crd", "bases")},
	}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	err = operatoribmcomv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = operatoribmcomv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = corev1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = routev1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = servicecav1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = monitoringv1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = odlm.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = networkingv1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = meterdefv1beta1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = operatorframeworkv1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	operatorNamespace, _ = os.LookupEnv("OPERATOR_NAMESPACE")
	Expect(operatorNamespace).ToNot(BeEmpty())

	namespace, _ = os.LookupEnv("NAMESPACE")
	Expect(namespace).ToNot(BeEmpty())

	opreqNamespace, _ = os.LookupEnv("OPREQ_TEST_NAMESPACE")
	Expect(opreqNamespace).ToNot(BeEmpty())

	// +kubebuilder:scaffold:scheme
	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sClient).ToNot(BeNil())

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme.Scheme,
		MetricsBindAddress: "0",
		Namespace:          "",
	})
	Expect(err).ToNot(HaveOccurred())

	k8sCFromMgr = mgr.GetClient()
	k8sRFromMgr = mgr.GetAPIReader()

	nssEnabledSemaphore := make(chan bool, 1)

	err = (&IBMLicensingReconciler{
		Client:                  mgr.GetClient(),
		Reader:                  mgr.GetAPIReader(),
		Log:                     ctrl.Log.WithName("controllers").WithName("IBMLicensing"),
		Scheme:                  mgr.GetScheme(),
		Recorder:                mgr.GetEventRecorderFor("IBMLicensing"),
		OperatorNamespace:       operatorNamespace,
		NamespaceScopeSemaphore: nssEnabledSemaphore,
	}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	err = (&OperandRequestReconciler{
		Client:            mgr.GetClient(),
		Reader:            mgr.GetAPIReader(),
		Log:               ctrl.Log.WithName("controllers").WithName("OperandRequest"),
		Scheme:            mgr.GetScheme(),
		OperatorNamespace: operatorNamespace,
	}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	k8sClient = mgr.GetClient()
	Expect(k8sClient).ToNot(BeNil())

	ocpEnvVar, _ := os.LookupEnv("OCP")
	if ocpEnvVar == "" {
		ocp = false
	} else {
		ocp = true
	}

	go func() {
		defer GinkgoRecover()
		err = mgr.Start(ctrl.SetupSignalHandler())
		Expect(err).ToNot(HaveOccurred())
	}()
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})
