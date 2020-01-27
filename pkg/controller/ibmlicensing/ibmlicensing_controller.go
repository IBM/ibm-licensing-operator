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

package ibmlicensing

import (
	"context"

	operatorv1alpha1 "github.com/ibm/ibm-licensing-operator/pkg/apis/operator/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

//TODO: determine if this should not be somewhere else in future:
const licensingResourceName = "licensing-service"

var log = logf.Log.WithName("controller_ibmlicensing")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new IBMLicensing Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileIBMLicensing{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("ibmlicensing-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource IBMLicensing
	err = c.Watch(&source.Kind{Type: &operatorv1alpha1.IBMLicensing{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource "Deployment" and requeue the owner IBMLicensing
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.IBMLicensing{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource "Service" and requeue the owner IBMLicensing
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.IBMLicensing{},
	})
	if err != nil {
		return err
	}
	return nil
}

// blank assignment to verify that ReconcileIBMLicensing implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileIBMLicensing{}

// ReconcileIBMLicensing reconciles a IBMLicensing object
type ReconcileIBMLicensing struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a IBMLicensing object and makes changes based on the state read
// and what is in the IBMLicensing.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Deployment and Service for IBM Licensing
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileIBMLicensing) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling IBMLicensing")

	// Fetch the IBMLicensing instance
	instance := &operatorv1alpha1.IBMLicensing{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			reqLogger.Info("IBMLicensing resource not found. Ignoring since object must be deleted")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		reqLogger.Error(err, "Failed to get IBMLicensing")
		return reconcile.Result{}, err
	}

	opVersion := instance.Spec.OperatorVersion
	reqLogger.Info("got IBMLicensing instance, version=" + opVersion + ", checking Service")

	// Check if the service already exists
	currentService := &corev1.Service{}

	err = r.client.Get(context.TODO(), types.NamespacedName{Name: licensingResourceName, Namespace: instance.Namespace}, currentService)
	// In case error is cause by non existing service we will create one:
	if err != nil && errors.IsNotFound(err) {
		// Define a new service
		newService := r.newServiceForLicensingCR(instance)
		reqLogger.Info("Creating a new Service", "Service.Namespace", newService.Namespace, "Service.Name", newService.Name)
		err = r.client.Create(context.TODO(), newService)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Service", "Service.Namespace", newService.Namespace, "Service.Name", newService.Name)
			return reconcile.Result{}, err
		}
		// Service created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get Service")
		return reconcile.Result{}, err
	}
	reqLogger.Info("got Service, checking Deployment")

	// Check if the deployment already exists, if not create a new one
	// currentDeployment := &appsv1.Deployment{}

	// err = r.client.Get(context.TODO(), types.NamespacedName{Name: licensingResourceName, Namespace: instance.Namespace}, currentDeployment)
	// if err != nil && errors.IsNotFound(err) {
	// 	// Define a new deployment
	// 	newDeployment := r.newDeploymentForLicensingCR(instance)
	// 	reqLogger.Info("Creating a new Deployment", "Deployment.Namespace", newDeployment.Namespace, "Deployment.Name", newDeployment.Name)
	// 	err = r.client.Create(context.TODO(), newDeployment)
	// 	if err != nil {
	// 		reqLogger.Error(err, "Failed to create new Deployment", "Deployment.Namespace", newDeployment.Namespace, "Deployment.Name", newDeployment.Name)
	// 		return reconcile.Result{}, err
	// 	}
	// 	// Deployment created successfully - return and requeue
	// 	return reconcile.Result{Requeue: true}, nil
	// } else if err != nil {
	// 	reqLogger.Error(err, "Failed to get Deployment")
	// 	return reconcile.Result{}, err
	// }
	// reqLogger.Info("got Deployment")

	// OLD CODE VERSION WAS HERE

	reqLogger.Info("all done")
	return reconcile.Result{}, nil
}

func (r *ReconcileIBMLicensing) newServiceForLicensingCR(instance *operatorv1alpha1.IBMLicensing) *corev1.Service {
	reqLogger := log.WithValues("serviceForLicensing", "Entry", "instance.Name", instance.Name)
	metaLabels := labelsForLicensingMeta(licensingResourceName)
	selectorLabels := labelsForLicensingSelector(instance.Name, licensingResourceName)

	reqLogger.Info("New Service Entry")
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      licensingResourceName,
			Namespace: instance.Namespace,
			Labels:    metaLabels,
			// Annotations: map[string]string{"prometheus.io/scrape": "false", "prometheus.io/scheme": "http"},
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       8080,
					TargetPort: intstr.FromInt(8080),
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Selector: selectorLabels,
		},
	}

	// Set IBMLicensing instance as the owner and controller of the Service
	err := controllerutil.SetControllerReference(instance, service, r.scheme)
	if err != nil {
		reqLogger.Error(err, "Failed to set owner for Service")
		return nil
	}
	return service
}

func labelsForLicensingSelector(instanceName string, appName string) map[string]string {
	return map[string]string{"app": appName, "component": "ibmlicensingsvc", "licensing_cr": instanceName}
}

func labelsForLicensingMeta(appName string) map[string]string {
	return map[string]string{"app.kubernetes.io/name": appName, "app.kubernetes.io/component": "ibmlicensingsvc", "release": "licensing"}
}

// !! OLD CODE with Pod creation
// // Define a new Pod object
// nonsense logic because newPod is only when pod does not exists
// pod := newPodForCR(instance)

// // Set IBMLicensing instance as the owner and controller
// if err := controllerutil.SetControllerReference(instance, pod, r.scheme); err != nil {
// 	return reconcile.Result{}, err
// }

// // Check if this Pod already exists
// found := &corev1.Pod{}
// err = r.client.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, found)
// if err != nil && errors.IsNotFound(err) {
// 	reqLogger.Info("Creating a new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
// 	err = r.client.Create(context.TODO(), pod)
// 	if err != nil {
// 		return reconcile.Result{}, err
// 	}

// 	// Pod created successfully - don't requeue
// 	return reconcile.Result{}, nil
// } else if err != nil {
// 	return reconcile.Result{}, err
// }

// // Pod already exists - don't requeue
// reqLogger.Info("Skip reconcile: Pod already exists", "Pod.Namespace", found.Namespace, "Pod.Name", found.Name)
// return reconcile.Result{}, nil

// }

// // newPodForCR returns a busybox pod with the same name/namespace as the cr
// func newPodForCR(cr *operatorv1alpha1.IBMLicensing) *corev1.Pod {
// 	labels := map[string]string{
// 		"app": cr.Name,
// 	}
// 	return &corev1.Pod{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:      cr.Name + "-pod",
// 			Namespace: cr.Namespace,
// 			Labels:    labels,
// 		},
// 		Spec: corev1.PodSpec{
// 			Containers: []corev1.Container{
// 				{
// 					Name:    "busybox",
// 					Image:   "busybox",
// 					Command: []string{"sleep", "3600"},
// 				},
// 			},
// 		},
// 	}
// }

// !! END OLD CODE with Pod creation
