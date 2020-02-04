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
	"reflect"

	operatorv1alpha1 "github.com/ibm/ibm-licensing-operator/pkg/apis/operator/v1alpha1"
	res "github.com/ibm/ibm-licensing-operator/pkg/resources"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

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
	secondaryResourceTypes := []runtime.Object{
		&appsv1.Deployment{},
		&corev1.Service{},
		&corev1.Secret{},
	}

	for _, restype := range secondaryResourceTypes {
		log.Info("Watching", "restype", restype)
		err = c.Watch(&source.Kind{Type: restype}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &operatorv1alpha1.IBMLicensing{},
		})
		if err != nil {
			return err
		}
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
	reqLogger := log.WithValues("Request.Name", request.Name)
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

	// TODO: check if opVersion is needed at spec
	opVersion := instance.Spec.OperatorVersion
	reqLogger.Info("got IBMLicensing instance, version=" + opVersion + ", checking Service")

	var recResult reconcile.Result
	var recErr error

	// Reconcile the expected deployment
	recResult, recErr = r.reconcileDeployment(instance)
	if recErr != nil || recResult.Requeue {
		return recResult, recErr
	}

	// Reconcile the expected service
	recResult, recErr = r.reconcileService(instance)
	if recErr != nil || recResult.Requeue {
		return recResult, recErr
	}

	// Reconcile the expected APISecretToken
	recResult, recErr = r.reconcileAPISecretToken(instance)
	if recErr != nil || recResult.Requeue {
		return recResult, recErr
	}

	reqLogger.Info("reconcile all done")
	return reconcile.Result{}, nil
}

func (r *ReconcileIBMLicensing) reconcileAPISecretToken(instance *operatorv1alpha1.IBMLicensing) (reconcile.Result, error) {
	reqLogger := log.WithValues("APISecretToken", "Entry", "instance.GetName()", instance.GetName())
	metaLabels := res.LabelsForLicensingMeta(instance)

	expectedSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Spec.APISecretToken,
			Namespace: instance.GetNamespace(),
			Labels:    metaLabels,
		},
		Type:       corev1.SecretTypeOpaque,
		StringData: map[string]string{"token": res.RandString(24)},
	}
	// Set IBMLicensing instance as the owner and controller of the Service
	err := controllerutil.SetControllerReference(instance, expectedSecret, r.scheme)
	if err != nil {
		reqLogger.Error(err, "Failed to set owner for Secret APISecretToken")
		return reconcile.Result{}, err
	}

	currentAPISecret := &corev1.Secret{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Spec.APISecretToken, Namespace: instance.GetNamespace()}, currentAPISecret)
	if err != nil && errors.IsNotFound(err) {
		// APISecretToken does not exist
		reqLogger.Info("APISecretToken does not exist, creating secret: " + instance.Spec.APISecretToken)
		err = r.client.Create(context.TODO(), expectedSecret)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Secret", "Secret.Namespace", expectedSecret.Namespace, "Secret.Name", expectedSecret.Name)
			return reconcile.Result{}, err
		}
		// Secret created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get secret APISecretToken")
		return reconcile.Result{}, err
	} // do not compare stringdata and update secret as it is generated

	return reconcile.Result{}, nil
}

func (r *ReconcileIBMLicensing) reconcileService(instance *operatorv1alpha1.IBMLicensing) (reconcile.Result, error) {
	reqLogger := log.WithValues("reconcileService", "Entry", "instance.GetName()", instance.GetName())

	expectedService := res.GetLicensingService(instance)

	// Set IBMLicensing instance as the owner and controller of the Service
	err := controllerutil.SetControllerReference(instance, expectedService, r.scheme)
	if err != nil {
		reqLogger.Error(err, "Failed to set owner for Service")
		return reconcile.Result{}, err
	}

	foundService := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: res.GetResourceName(instance), Namespace: instance.GetNamespace()}, foundService)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Service", "Service.Namespace", expectedService.Namespace, "Service.Name", expectedService.Name)
		err = r.client.Create(context.TODO(), expectedService)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Service", "Service.Namespace", expectedService.Namespace, "Service.Name", expectedService.Name)
			return reconcile.Result{}, err
		}
		// Service created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get Service")
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileIBMLicensing) reconcileDeployment(instance *operatorv1alpha1.IBMLicensing) (reconcile.Result, error) {
	reqLogger := log.WithValues("reconcileDeployment", "Entry", "instance.GetName()", instance.GetName())

	expectedDeployment := res.GetLicensingDeployment(instance)

	// Set instance as the owner and controller of the Deployment
	err := controllerutil.SetControllerReference(instance, expectedDeployment, r.scheme)
	if err != nil {
		reqLogger.Error(err, "Failed to set owner for Deployment")
		return reconcile.Result{}, err
	}

	shouldUpdate := false
	foundDeployment := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: res.GetResourceName(instance), Namespace: instance.GetNamespace()}, foundDeployment)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Deployment", "Deployment.Namespace", expectedDeployment.Namespace, "Deployment.Name", expectedDeployment.Name)
		err = r.client.Create(context.TODO(), expectedDeployment)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Deployment", "Deployment.Namespace", expectedDeployment.Namespace, "Deployment.Name", expectedDeployment.Name)
			return reconcile.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get Deployment")
		return reconcile.Result{}, err
	} else if !reflect.DeepEqual(foundDeployment.Spec.Template.Spec.Volumes, expectedDeployment.Spec.Template.Spec.Volumes) {
		reqLogger.Info("Deployment has wrong volumes", "Deployment.Namespace", foundDeployment.Namespace,
			"Deployment.Name", foundDeployment.Name, "Deployment.Volumes", foundDeployment.Spec.Template.Spec.Volumes,
			"ExpectedDeployment.Volumes", expectedDeployment.Spec.Template.Spec.Volumes)
		shouldUpdate = true
	} else if len(foundDeployment.Spec.Template.Spec.Containers) != len(expectedDeployment.Spec.Template.Spec.Containers) {
		reqLogger.Info("Deployment has number of containers", "Deployment.Namespace", foundDeployment.Namespace,
			"Deployment.Name", foundDeployment.Name, "Deployment.Containers", foundDeployment.Spec.Template.Spec.Containers,
			"ExpectedDeployment.Containers", expectedDeployment.Spec.Template.Spec.Containers)
		shouldUpdate = true
	} else if !reflect.DeepEqual(foundDeployment.Spec.Template.Spec.Containers[0].Name, expectedDeployment.Spec.Template.Spec.Containers[0].Name) {
		reqLogger.Info("Deployment wrong spec error 3", "Deployment.Namespace", foundDeployment.Namespace, "Deployment.Name", foundDeployment.Name)
		shouldUpdate = true
	} else if !reflect.DeepEqual(foundDeployment.Spec.Template.Spec.Containers[0].Image, expectedDeployment.Spec.Template.Spec.Containers[0].Image) {
		reqLogger.Info("Deployment wrong spec error 4", "Deployment.Namespace", foundDeployment.Namespace, "Deployment.Name", foundDeployment.Name)
		shouldUpdate = true
	} else if !reflect.DeepEqual(foundDeployment.Spec.Template.Spec.Containers[0].Ports, expectedDeployment.Spec.Template.Spec.Containers[0].Ports) {
		reqLogger.Info("Deployment wrong containers ports", "Deployment.Namespace", foundDeployment.Namespace, "Deployment.Name", foundDeployment.Name,
			"Found Container Ports", foundDeployment.Spec.Template.Spec.Containers[0].Ports,
			"Expected Container Ports", expectedDeployment.Spec.Template.Spec.Containers[0].Ports)
		shouldUpdate = true
	} else if !reflect.DeepEqual(foundDeployment.Spec.Template.Spec.Containers[0].VolumeMounts, expectedDeployment.Spec.Template.Spec.Containers[0].VolumeMounts) {
		reqLogger.Info("Deployment wrong spec error 6", "Deployment.Namespace", foundDeployment.Namespace, "Deployment.Name", foundDeployment.Name)
		shouldUpdate = true
	}

	if shouldUpdate {
		// Spec is incorrect, update it and requeue
		reqLogger.Info("Found deployment spec is incorrect", "Found", foundDeployment.Name, "Expected", expectedDeployment.Name)
		refreshedDeployment := foundDeployment.DeepCopy()
		refreshedDeployment.Spec.Template.Spec.Volumes = expectedDeployment.Spec.Template.Spec.Volumes
		refreshedDeployment.Spec.Template.Spec.Containers = expectedDeployment.Spec.Template.Spec.Containers
		reqLogger.Info("Updating Deployment volumes to:", "RefreshedDeployment.Volumes", refreshedDeployment.Spec.Template.Spec.Volumes)
		err = r.client.Update(context.TODO(), refreshedDeployment)
		if err != nil {
			// only need to delete deployment as new will be recreated on next reconciliation
			reqLogger.Error(err, "Failed to update Deployment, deleting...", "Namespace", foundDeployment.Namespace, "Name", foundDeployment.Name)
			err = r.client.Delete(context.TODO(), foundDeployment)
			if err != nil {
				reqLogger.Error(err, "Failed to delete Deployment during recreation", "Namespace", foundDeployment.Namespace, "Name", foundDeployment.Name)
				return reconcile.Result{}, err
			}
			// Deployment deleted successfully - return and requeue to create new one
			reqLogger.Info("Deleted deployment successfully", "Deployment.Namespace", foundDeployment.Namespace, "Deployment.Name", foundDeployment.Name)
			return reconcile.Result{Requeue: true}, nil
		}
		reqLogger.Info("Updated deployment successfully", "Deployment.Namespace", refreshedDeployment.Namespace, "Deployment.Name", refreshedDeployment.Name)
		// Spec updated - return and requeue
		return reconcile.Result{Requeue: true}, nil
	}

	return reconcile.Result{}, nil
}
