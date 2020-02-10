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
	rbacv1 "k8s.io/api/rbac/v1"
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

type reconcileFunctionType = func(*operatorv1alpha1.IBMLicensing) (reconcile.Result, error)

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
		&rbacv1.Role{},
		&rbacv1.RoleBinding{},
		&rbacv1.ClusterRole{},
		&rbacv1.ClusterRoleBinding{},
		&corev1.ServiceAccount{},
		&corev1.Secret{},
		&appsv1.Deployment{},
		&corev1.Service{},
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
	reqLogger := log.WithValues("Request", request)
	reqLogger.Info("Reconciling IBMLicensing")

	// Fetch the IBMLicensing instance
	instance := &operatorv1alpha1.IBMLicensing{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			// reqLogger.Info("IBMLicensing resource not found. Ignoring since object must be deleted")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		// reqLogger.Error(err, "Failed to get IBMLicensing")
		return reconcile.Result{}, err
	}

	var recResult reconcile.Result
	var recErr error

	reconcileFunctions := []interface{}{
		r.reconcileNamespace,
		r.reconcileServiceAccount,
		r.reconcileRole,
		r.reconcileRoleBinding,
		r.reconcileClusterRole,
		r.reconcileClusterRoleBinding,
		r.reconcileAPISecretToken,
		r.reconcileDeployment,
		r.reconcileService,
	}

	for _, reconcileFunction := range reconcileFunctions {
		recResult, recErr = reconcileFunction.(reconcileFunctionType)(instance)
		if recErr != nil || recResult.Requeue {
			return recResult, recErr
		}
	}

	reqLogger.Info("reconcile all done")
	return reconcile.Result{}, nil
}

func (r *ReconcileIBMLicensing) reconcileServiceAccount(instance *operatorv1alpha1.IBMLicensing) (reconcile.Result, error) {
	reqLogger := log.WithValues("reconcileServiceAccount", "Entry", "instance.GetName()", instance.GetName())
	expectedSA := res.GetLicensingServiceAccount(instance)
	foundSA := &corev1.ServiceAccount{}
	reconcileResult, err := r.reconcileResourceNamespacedExistance(instance, expectedSA, foundSA)
	if err != nil || reconcileResult.Requeue {
		return reconcileResult, err
	}
	// Check if found SA has all necessary Pull Secrets
	shouldUpdate := false
	for _, imagePullSecret := range expectedSA.ImagePullSecrets {
		if !res.Contains(foundSA.ImagePullSecrets, imagePullSecret) {
			foundSA.ImagePullSecrets = append(foundSA.ImagePullSecrets, imagePullSecret)
			shouldUpdate = true
		}
	}
	if shouldUpdate {
		//TODO: add updating deployment here
		reqLogger.Info("Updating ServiceAccount", "Updated ServiceAccount", foundSA)
		err = r.client.Update(context.TODO(), foundSA)
		if err != nil {
			reqLogger.Error(err, "Failed to update ServiceAccount, deleting...")
			err = r.client.Delete(context.TODO(), foundSA)
			if err != nil {
				reqLogger.Error(err, "Failed to delete ServiceAccount during recreation")
				return reconcile.Result{}, err
			}
			reqLogger.Info("Deleted ServiceAccount successfully")
			return reconcile.Result{Requeue: true}, nil
		}
		reqLogger.Info("Updated ServiceAccount successfully")
		// Spec updated - return and requeue
		return reconcile.Result{Requeue: true}, nil
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileIBMLicensing) reconcileRole(instance *operatorv1alpha1.IBMLicensing) (reconcile.Result, error) {
	expectedRole := res.GetLicensingRole(instance)
	foundRole := &rbacv1.Role{}
	return r.reconcileResourceNamespacedExistance(instance, expectedRole, foundRole)
}

func (r *ReconcileIBMLicensing) reconcileRoleBinding(instance *operatorv1alpha1.IBMLicensing) (reconcile.Result, error) {
	expectedRoleBinding := res.GetLicensingRoleBinding(instance)
	foundRoleBinding := &rbacv1.RoleBinding{}
	return r.reconcileResourceNamespacedExistance(instance, expectedRoleBinding, foundRoleBinding)
}

func (r *ReconcileIBMLicensing) reconcileClusterRole(instance *operatorv1alpha1.IBMLicensing) (reconcile.Result, error) {
	expectedClusterRole := res.GetLicensingClusterRole(instance)
	foundClusterRole := &rbacv1.ClusterRole{}
	return r.reconcileResourceClusterExistance(instance, expectedClusterRole, foundClusterRole)
}

func (r *ReconcileIBMLicensing) reconcileClusterRoleBinding(instance *operatorv1alpha1.IBMLicensing) (reconcile.Result, error) {
	expectedClusterRoleBinding := res.GetLicensingClusterRoleBinding(instance)
	foundClusterRoleBinding := &rbacv1.ClusterRoleBinding{}
	return r.reconcileResourceClusterExistance(instance, expectedClusterRoleBinding, foundClusterRoleBinding)
}

func (r *ReconcileIBMLicensing) reconcileNamespace(instance *operatorv1alpha1.IBMLicensing) (reconcile.Result, error) {
	expectedNamespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: instance.Spec.APINamespace,
		},
	}
	foundNamespace := &corev1.Namespace{}
	return r.reconcileResourceClusterExistance(instance, expectedNamespace, foundNamespace)
}

func (r *ReconcileIBMLicensing) reconcileAPISecretToken(instance *operatorv1alpha1.IBMLicensing) (reconcile.Result, error) {
	metaLabels := res.LabelsForLicensingMeta(instance)
	expectedSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Spec.APISecretToken,
			Namespace: instance.Spec.APINamespace,
			Labels:    metaLabels,
		},
		Type:       corev1.SecretTypeOpaque,
		StringData: map[string]string{"token": res.RandString(24)},
	}
	foundSecret := &corev1.Secret{}
	return r.reconcileResourceNamespacedExistance(instance, expectedSecret, foundSecret)
}

func (r *ReconcileIBMLicensing) reconcileService(instance *operatorv1alpha1.IBMLicensing) (reconcile.Result, error) {
	expectedService := res.GetLicensingService(instance)
	foundService := &corev1.Service{}
	return r.reconcileResourceNamespacedExistance(instance, expectedService, foundService)
}

func (r *ReconcileIBMLicensing) reconcileDeployment(instance *operatorv1alpha1.IBMLicensing) (reconcile.Result, error) {
	reqLogger := log.WithValues("reconcileDeployment", "Entry", "instance.GetName()", instance.GetName())
	expectedDeployment := res.GetLicensingDeployment(instance)
	shouldUpdate := false
	foundDeployment := &appsv1.Deployment{}
	reconcileResult, err := r.reconcileResourceNamespacedExistance(instance, expectedDeployment, foundDeployment)
	if err != nil || reconcileResult.Requeue {
		return reconcileResult, err
	}
	foundSpec := foundDeployment.Spec.Template.Spec
	expectedSpec := expectedDeployment.Spec.Template.Spec
	if !reflect.DeepEqual(foundSpec.Volumes, expectedSpec.Volumes) {
		reqLogger.Info("Deployment has wrong volumes", "Deployment.Namespace", foundDeployment.Namespace,
			"Deployment.Name", foundDeployment.Name, "Deployment.Volumes", foundSpec.Volumes,
			"ExpectedDeployment.Volumes", expectedSpec.Volumes)
		shouldUpdate = true
	} else if len(foundSpec.Containers) != len(expectedSpec.Containers) {
		reqLogger.Info("Deployment has number of containers", "Deployment.Namespace", foundDeployment.Namespace,
			"Deployment.Name", foundDeployment.Name, "Deployment.Containers", foundSpec.Containers,
			"ExpectedDeployment.Containers", expectedSpec.Containers)
		shouldUpdate = true
	} else if !reflect.DeepEqual(foundSpec.Containers[0].Name, expectedSpec.Containers[0].Name) {
		reqLogger.Info("Deployment wrong container name", "Deployment.Namespace", foundDeployment.Namespace, "Deployment.Name", foundDeployment.Name,
			"Container.Name", foundSpec.Containers[0].Name,
			"ExpectedContainer.Name", expectedSpec.Containers[0].Name)
		shouldUpdate = true
	} else if !reflect.DeepEqual(foundSpec.Containers[0].Image, expectedSpec.Containers[0].Image) {
		reqLogger.Info("Deployment wrong container image", "Deployment.Namespace", foundDeployment.Namespace, "Deployment.Name", foundDeployment.Name,
			"Container.Image", foundSpec.Containers[0].Image,
			"ExpectedContainer.Image", expectedSpec.Containers[0].Image)
		shouldUpdate = true
	} else if !reflect.DeepEqual(foundSpec.Containers[0].Ports, expectedSpec.Containers[0].Ports) {
		reqLogger.Info("Deployment wrong containers ports", "Deployment.Namespace", foundDeployment.Namespace, "Deployment.Name", foundDeployment.Name,
			"Found Container Ports", foundSpec.Containers[0].Ports,
			"Expected Container Ports", expectedSpec.Containers[0].Ports)
		shouldUpdate = true
	} else if !reflect.DeepEqual(foundSpec.Containers[0].VolumeMounts, expectedSpec.Containers[0].VolumeMounts) {
		reqLogger.Info("Deployment wrong VolumeMounts in container", "Deployment.Namespace", foundDeployment.Namespace, "Deployment.Name", foundDeployment.Name)
		shouldUpdate = true
	} else if !reflect.DeepEqual(foundSpec.Containers[0].Env, expectedSpec.Containers[0].Env) {
		reqLogger.Info("Deployment wrong env variables in container", "Deployment.Namespace", foundDeployment.Namespace,
			"Deployment.Name", foundDeployment.Name, "Deployment.Containers", foundSpec.Containers,
			"ExpectedDeployment.Containers", expectedSpec.Containers)
		shouldUpdate = true
	} else if foundSpec.ServiceAccountName != expectedSpec.ServiceAccountName {
		reqLogger.Info("Deployment wrong service account name", "Deployment.Namespace", foundDeployment.Namespace,
			"Deployment.Name", foundDeployment.Name, "Deployment.SA", foundSpec.ServiceAccountName,
			"ExpectedDeployment.SA", expectedSpec.ServiceAccountName)
		shouldUpdate = true
	} else if !reflect.DeepEqual(foundSpec.Containers[0].SecurityContext, expectedSpec.Containers[0].SecurityContext) {
		reqLogger.Info("Deployment wrong container security context", "Deployment.Namespace", foundDeployment.Namespace,
			"Deployment.Name", foundDeployment.Name, "Deployment.SC", foundSpec.Containers[0].SecurityContext,
			"ExpectedDeployment.SC", expectedSpec.Containers[0].SecurityContext)
		shouldUpdate = true
	}

	if shouldUpdate {
		// Spec is incorrect, update it and requeue
		reqLogger.Info("Found deployment spec is incorrect", "Found", foundDeployment.Name, "Expected", expectedDeployment.Name)
		refreshedDeployment := foundDeployment.DeepCopy()
		refreshedDeployment.Spec.Template.Spec.Volumes = expectedDeployment.Spec.Template.Spec.Volumes
		refreshedDeployment.Spec.Template.Spec.Containers = expectedDeployment.Spec.Template.Spec.Containers
		refreshedDeployment.Spec.Template.Spec.ServiceAccountName = expectedDeployment.Spec.Template.Spec.ServiceAccountName
		reqLogger.Info("Updating Deployment volumes to", "RefreshedDeployment.Volumes", refreshedDeployment.Spec.Template.Spec.Volumes)
		reqLogger.Info("Updating Deployment containers to", "RefreshedDeployment.Containers", refreshedDeployment.Spec.Template.Spec.Containers)
		reqLogger.Info("Updating Deployment SA to", "RefreshedDeployment.ServiceAccount", refreshedDeployment.Spec.Template.Spec.ServiceAccountName)
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

type ResourceObject interface {
	metav1.Object
	runtime.Object
}

func (r *ReconcileIBMLicensing) reconcileResourceNamespacedExistance(
	instance *operatorv1alpha1.IBMLicensing, expectedRes ResourceObject, foundRes runtime.Object) (reconcile.Result, error) {

	namespacedName := types.NamespacedName{Name: expectedRes.GetName(), Namespace: expectedRes.GetNamespace()}
	return r.reconcileResourceExistance(instance, expectedRes, foundRes, namespacedName)
}

func (r *ReconcileIBMLicensing) reconcileResourceClusterExistance(
	instance *operatorv1alpha1.IBMLicensing, expectedRes ResourceObject, foundRes runtime.Object) (reconcile.Result, error) {

	namespacedName := types.NamespacedName{Name: expectedRes.GetName()}
	return r.reconcileResourceExistance(instance, expectedRes, foundRes, namespacedName)
}

func (r *ReconcileIBMLicensing) reconcileResourceExistance(
	instance *operatorv1alpha1.IBMLicensing, expectedRes ResourceObject, foundRes runtime.Object, namespacedName types.NamespacedName) (reconcile.Result, error) {

	resType := reflect.TypeOf(expectedRes)
	reqLogger := log.WithValues(resType.String(), "Entry", "instance.GetName()", instance.GetName())

	// expectedRes already set before and passed via parameter
	err := controllerutil.SetControllerReference(instance, expectedRes, r.scheme)
	if err != nil {
		reqLogger.Error(err, "Failed to define expected resource")
		return reconcile.Result{}, err
	}

	// foundRes already initialized before and passed via parameter
	err = r.client.Get(context.TODO(), namespacedName, foundRes)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info(resType.String()+" does not exist, trying creating new one", "Name", expectedRes.GetName(),
			"Namespace", expectedRes.GetNamespace())
		err = r.client.Create(context.TODO(), expectedRes)
		if err != nil {
			reqLogger.Error(err, "Failed to create new "+resType.String(), "Name", expectedRes.GetName(),
				"Namespace", expectedRes.GetNamespace())
			return reconcile.Result{}, err
		}
		// Created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get "+resType.String(), "Name", expectedRes.GetName(),
			"Namespace", expectedRes.GetNamespace())
		return reconcile.Result{}, err
	} else {
		reqLogger.Info(resType.String() + " is correct!")
	}
	return reconcile.Result{}, nil
}
