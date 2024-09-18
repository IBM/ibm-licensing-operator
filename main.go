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

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	r "runtime"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	servicecav1 "github.com/openshift/api/operator/v1"
	routev1 "github.com/openshift/api/route/v1"
	meterdefv1beta1 "github.com/redhat-marketplace/redhat-marketplace-operator/v2/apis/marketplace/v1beta1"

	"go.uber.org/zap/zapcore"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	operatorframeworkv1 "github.com/operator-framework/api/pkg/operators/v1"

	cache "github.com/IBM/controller-filtered-cache/filteredcache"
	odlm "github.com/IBM/operand-deployment-lifecycle-manager/api/v1alpha1"

	operatorv1 "github.com/IBM/ibm-licensing-operator/api/v1"
	operatoribmcomv1alpha1 "github.com/IBM/ibm-licensing-operator/api/v1alpha1"
	"github.com/IBM/ibm-licensing-operator/controllers"
	res "github.com/IBM/ibm-licensing-operator/controllers/resources"
	"github.com/IBM/ibm-licensing-operator/version"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func printVersion() {
	setupLog.Info(fmt.Sprintf("Operator Version: %s", version.Version))
	setupLog.Info(fmt.Sprintf("Operator BuildDate: %s", readFile("/IMAGE_BUILDDATE")))
	setupLog.Info(fmt.Sprintf("Operator Commit: %s", readFile("/IMAGE_RELEASE")))
	setupLog.Info(fmt.Sprintf("Go Version: %s", r.Version()))
	setupLog.Info(fmt.Sprintf("Go OS/Arch: %s/%s", r.GOOS, r.GOARCH))
}

func readFile(filename string) string {
	content, err := os.ReadFile(filename)
	if err != nil {
		setupLog.Info(fmt.Sprintf("Can not read: %s", filename))
		return ""
	}

	return string(content)[:len(string(content))-1]
}

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(operatoribmcomv1alpha1.AddToScheme(scheme))

	utilruntime.Must(routev1.AddToScheme(scheme))

	utilruntime.Must(servicecav1.AddToScheme(scheme))

	utilruntime.Must(monitoringv1.AddToScheme(scheme))

	utilruntime.Must(networkingv1.AddToScheme(scheme))

	utilruntime.Must(meterdefv1beta1.AddToScheme(scheme))

	utilruntime.Must(odlm.AddToScheme(scheme))

	utilruntime.Must(operatorv1.AddToScheme(scheme))

	utilruntime.Must(operatorframeworkv1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var routinesToCancel []func()
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	ctrl.SetLogger(zap.New(func(o *zap.Options) {
		o.Development = true
		o.TimeEncoder = zapcore.RFC3339TimeEncoder
	}))

	printVersion()

	operatorNamespace, err := res.GetOperatorNamespace()
	if err != nil {
		setupLog.Error(err, "unable to get OPERATOR_NAMESPACE")
	}

	watchNamespaces, err := res.GetWatchNamespaceAsList()
	if err != nil {
		setupLog.Error(err, "unable to get WATCH_NAMESPACE")
		setupLog.Info("Manager will watch and manage resources only in operator namespace")
		watchNamespaces = []string{operatorNamespace}
	}

	gvkLabelMap := map[schema.GroupVersionKind]cache.Selector{
		corev1.SchemeGroupVersion.WithKind("Secret"): {
			LabelSelector: "release in (ibm-licensing-service)",
		},
		appsv1.SchemeGroupVersion.WithKind("Deployment"): {
			LabelSelector: "release in (ibm-licensing-service)",
		},
		corev1.SchemeGroupVersion.WithKind("Pod"): {
			LabelSelector: "release in (ibm-licensing-service)",
		},
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "e1f51baf.ibm.com",
		NewCache:           cache.MultiNamespacedFilteredCacheBuilder(gvkLabelMap, watchNamespaces),
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// 1-size channel for communicating namespace scope status between IBMLicensing controller and operandrequest-discovery goroutine
	nssEnabledSemaphore := make(chan bool, 1)

	controller := &controllers.IBMLicensingReconciler{
		Client:                  mgr.GetClient(),
		Reader:                  mgr.GetAPIReader(),
		Log:                     ctrl.Log.WithName("controllers").WithName("IBMLicensing"),
		Scheme:                  mgr.GetScheme(),
		Recorder:                mgr.GetEventRecorderFor("IBMLicensing"),
		OperatorNamespace:       operatorNamespace,
		NamespaceScopeSemaphore: nssEnabledSemaphore,
	}
	if err = controller.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "IBMLicensing")
		os.Exit(1)
	}

	operandRequestList := odlm.OperandRequestList{}
	opreqControllerEnabled, err := res.DoesCRDExist(mgr.GetAPIReader(), &operandRequestList)
	if err != nil {
		setupLog.Error(err, "An error occurred while checking for CRD existence. OperandRequest controller will not be started")
	}

	if opreqControllerEnabled {
		if err = (&controllers.OperandRequestReconciler{
			Client:            mgr.GetClient(),
			Reader:            mgr.GetAPIReader(),
			Log:               ctrl.Log.WithName("controllers").WithName("OperandRequest"),
			Scheme:            mgr.GetScheme(),
			OperatorNamespace: operatorNamespace,
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "OperandRequest")
			os.Exit(1)
		}
		crdLogger := ctrl.Log.WithName("operandrequest-discovery")
		// In Cloud Pak 2.0/3.0 coexistence scenario, License Service Operator 4.x.x leverages Namespace Scope Operator and must not modify OperatorGroup.
		isNssActive, err := res.IsNamespaceScopeOperatorAvailable(context.Background(), mgr.GetAPIReader(), operatorNamespace)
		if err != nil {
			setupLog.Error(err, "Error occurred while detecting Namespace Scope Operator")
		}
		if isNssActive {
			setupLog.Info("Namespace Scope ConfigMap detected. operandrequest-discovery disabled")
		} else {
			go controllers.DiscoverOperandRequests(&crdLogger, mgr.GetClient(), mgr.GetAPIReader(), watchNamespaces, nssEnabledSemaphore)

			logger := ctrl.Log.WithName("operatorgroup-namespaces-watcher")
			removeStaleNamespacesTaskCtx, cancelRemoveStaleNamespacesTask := context.WithCancel(context.Background())
			go controllers.RunRemoveStaleNamespacesFromOperatorGroupTask(removeStaleNamespacesTaskCtx, &logger, mgr.GetClient(), mgr.GetAPIReader())
			routinesToCancel = append(routinesToCancel, cancelRemoveStaleNamespacesTask)
		}
	} else {
		logger := ctrl.Log.WithName("crd-watcher").WithName("OperandRequest")
		// Set custom time duration for CRD watcher (in seconds)
		reconcileInterval, err := res.GetCrdReconcileInterval()
		if err != nil {
			setupLog.Error(err, "Incorrect reconcile interval set. Defaulting to 300s", "crd-watcher", "OperandRequest")
		}
		go res.RestartOnCRDCreation(&logger, mgr.GetClient(), &operandRequestList, reconcileInterval)
	}

	// If OperandBindInfo CRD exists, try to find ibm-licensing-bindinfo and delete it.
	operandBindInfoList := odlm.OperandBindInfoList{}
	bindInfoCrdExists, err := res.DoesCRDExist(mgr.GetAPIReader(), &operandBindInfoList)
	if err != nil {
		setupLog.Error(err, "An error occurred while checking for OperandBindInfo CRD existence")
	}

	if bindInfoCrdExists {
		go func() {
			err := res.DeleteBindInfoIfExists(context.TODO(), mgr.GetAPIReader(), mgr.GetClient(), operatorNamespace)
			if err != nil {
				ctrl.Log.Error(err, "An error occurred while detecting and deleting "+res.LsBindInfoName)
			} else {
				ctrl.Log.Info(res.LsBindInfoName + " deleted")
			}
		}()
	}

	// +kubebuilder:scaffold:builder

	setupLog.Info("Creating first instance.")
	_ = controller.CreateDefaultInstance(true)

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		for _, cancelRoutine := range routinesToCancel {
			cancelRoutine()
		}
		os.Exit(1)
	}
}
