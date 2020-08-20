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

package ibmlicenseservicereporter

import (
	"context"
	"reflect"
	"time"

	extensionsv1 "k8s.io/api/extensions/v1beta1"

	"github.com/go-logr/logr"
	operatorv1alpha1 "github.com/ibm/ibm-licensing-operator/pkg/apis/operator/v1alpha1"
	res "github.com/ibm/ibm-licensing-operator/pkg/resources"
	"github.com/ibm/ibm-licensing-operator/pkg/resources/reporter"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metaErrors "k8s.io/apimachinery/pkg/api/meta"
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

var log = logf.Log.WithName("controller_ibmlicenseservicereporter")
var isOpenshiftCluster = false

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new IBMLicenseServiceReporter Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileIBMLicenseServiceReporter{client: mgr.GetClient(), reader: mgr.GetAPIReader(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("ibmlicenseservicereporter-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource IBMLicenseServiceReporter
	err = c.Watch(&source.Kind{Type: &operatorv1alpha1.IBMLicenseServiceReporter{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resources
	err = res.WatchForResources(log, &operatorv1alpha1.IBMLicensing{}, c, []res.ResourceObject{
		&appsv1.Deployment{},
		&corev1.Service{},
	})
	if err != nil {
		return err
	}

	routeTestInstance := &routev1.Route{}
	err = mgr.GetClient().Get(context.TODO(), types.NamespacedName{}, routeTestInstance)
	if err != nil && metaErrors.IsNoMatchError(err) {
		log.Error(err, "Route CR not found, assuming not on OpenShift Cluster, restart operator if this is wrong")
		isOpenshiftCluster = false
	}

	if isOpenshiftCluster {
		// Watch for changes to openshift resources if on OC
		err = res.WatchForResources(log, &operatorv1alpha1.IBMLicensing{}, c, []res.ResourceObject{
			&routev1.Route{},
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// blank assignment to verify that ReconcileIBMLicenseServiceReporter implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileIBMLicenseServiceReporter{}

// ReconcileIBMLicenseServiceReporter reconciles a IBMLicenseServiceReporter object
type ReconcileIBMLicenseServiceReporter struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	reader client.Reader
	scheme *runtime.Scheme
}

type reconcileFunctionType = func(*operatorv1alpha1.IBMLicenseServiceReporter) (reconcile.Result, error)

// Reconcile reads that state of the cluster for a IBMLicenseServiceReporter object and makes changes based on the state read
// and what is in the IBMLicenseServiceReporter.Spec
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileIBMLicenseServiceReporter) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request", request)
	reqLogger.Info("Reconciling IBMLicenseServiceReporter")

	var recResult reconcile.Result
	var recErr error

	reconcileFunctions := []interface{}{
		r.reconcileServiceAccount,
		r.reconcileRole,
		r.reconcileRoleBinding,
		r.reconcileAPISecretToken,
		r.reconcileDatabaseSecret,
		r.reconcilePersistentVolumeClaim,
		r.reconcileDeployment,
		r.reconcileService,
		r.reconcileReporterRoute,
		r.reconcileUIIngress,
		r.reconcileIngressProxy,
	}

	// Fetch the IBMLicenseServiceReporter instance
	foundInstance := &operatorv1alpha1.IBMLicenseServiceReporter{}
	err := r.client.Get(context.TODO(), request.NamespacedName, foundInstance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			// reqLogger.Info("IBMLicenseServiceReporter resource not found. Ignoring since object must be deleted")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	instance := foundInstance.DeepCopy()
	reqLogger.Info("got IBM License Service Reporter application, version=" + instance.Spec.Version)
	err = instance.Spec.FillDefaultValues(reqLogger, r.reader)
	if err != nil {
		return reconcile.Result{}, err
	}

	for _, reconcileFunction := range reconcileFunctions {
		recResult, recErr = reconcileFunction.(reconcileFunctionType)(instance)
		if recErr != nil || recResult.Requeue {
			return recResult, recErr
		}
	}

	// Update status logic, using foundInstance, because we do not want to add filled default values to yaml
	return r.updateStatus(foundInstance, reqLogger)
}

func (r *ReconcileIBMLicenseServiceReporter) updateStatus(
	instance *operatorv1alpha1.IBMLicenseServiceReporter,
	reqLogger logr.Logger) (reconcile.Result, error) {
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(instance.GetNamespace()),
		client.MatchingLabels(reporter.LabelsForPod(instance)),
	}
	if err := r.client.List(context.TODO(), podList, listOpts...); err != nil {
		reqLogger.Error(err, "Failed to list pods")
		return reconcile.Result{}, err
	}

	var podStatuses []corev1.PodStatus
	for _, pod := range podList.Items {
		if pod.Status.Conditions != nil {
			i := 0
			for _, podCondition := range pod.Status.Conditions {
				if (podCondition.LastProbeTime == metav1.Time{Time: time.Time{}}) {
					// Time{} is treated as null and causes error at status update so value need to be changed to some other default empty value
					pod.Status.Conditions[i].LastProbeTime = metav1.Time{
						Time: time.Unix(0, 1),
					}
				}
				i++
			}
		}
		podStatuses = append(podStatuses, pod.Status)
	}

	if !reflect.DeepEqual(podStatuses, instance.Status.LicensingReporterPods) {
		reqLogger.Info("Updating IBMLicenseServiceReporter status", "Instance", instance)
		instance.Status.LicensingReporterPods = podStatuses
		err := r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			reqLogger.Info("Failed to update pod status")
		}
	}

	reqLogger.Info("reconcile all done")
	return reconcile.Result{}, nil
}

func (r *ReconcileIBMLicenseServiceReporter) reconcileServiceAccount(instance *operatorv1alpha1.IBMLicenseServiceReporter) (reconcile.Result, error) {
	reqLogger := log.WithValues("reconcileServiceAccount", "Entry", "instance.GetName()", instance.GetName())
	expectedSA := reporter.GetServiceAccount(instance)
	foundSA := &corev1.ServiceAccount{}
	namespacedName := types.NamespacedName{Name: expectedSA.GetName(), Namespace: expectedSA.GetNamespace()}
	reconcileResult, err := r.reconcileResourceExistence(instance, expectedSA, foundSA, namespacedName)
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
		// Spec updated - return and do not requeue as it might not consider extra values
		return reconcile.Result{}, nil
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileIBMLicenseServiceReporter) reconcileRole(instance *operatorv1alpha1.IBMLicenseServiceReporter) (reconcile.Result, error) {
	expectedRole := reporter.GetRole(instance)
	foundRole := &rbacv1.Role{}
	namespacedName := types.NamespacedName{Name: expectedRole.GetName(), Namespace: expectedRole.GetNamespace()}
	return r.reconcileResourceExistence(instance, expectedRole, foundRole, namespacedName)
}

func (r *ReconcileIBMLicenseServiceReporter) reconcileRoleBinding(instance *operatorv1alpha1.IBMLicenseServiceReporter) (reconcile.Result, error) {
	expectedRoleBinding := reporter.GetRoleBinding(instance)
	foundRoleBinding := &rbacv1.RoleBinding{}
	namespacedName := types.NamespacedName{Name: expectedRoleBinding.GetName(), Namespace: expectedRoleBinding.GetNamespace()}
	return r.reconcileResourceExistence(instance, expectedRoleBinding, foundRoleBinding, namespacedName)
}

func (r *ReconcileIBMLicenseServiceReporter) reconcilePersistentVolumeClaim(instance *operatorv1alpha1.IBMLicenseServiceReporter) (reconcile.Result, error) {

	expectedPVC := reporter.GetPersistenceVolumeClaim(instance)
	foundPVC := &corev1.PersistentVolumeClaim{}
	namespacedName := types.NamespacedName{Name: expectedPVC.GetName(), Namespace: expectedPVC.GetNamespace()}
	reconcileResult, err := r.reconcileResourceExistence(instance, expectedPVC, foundPVC, namespacedName)
	if err != nil || reconcileResult.Requeue {
		return reconcileResult, err
	}
	return reconcile.Result{}, nil

}

func (r *ReconcileIBMLicenseServiceReporter) reconcileDatabaseSecret(instance *operatorv1alpha1.IBMLicenseServiceReporter) (reconcile.Result, error) {
	reqLogger := log.WithValues("reconcileDatabaseSecret", "Entry", "instance.GetName()", instance.GetName())
	expectedSecret, err := reporter.GetDatabaseSecret(instance)
	if err != nil {
		reqLogger.Info("Failed to get expected secret")
		return reconcile.Result{
			Requeue:      true,
			RequeueAfter: time.Minute,
		}, err
	}
	foundSecret := &corev1.Secret{}
	namespacedName := types.NamespacedName{Name: expectedSecret.GetName(), Namespace: expectedSecret.GetNamespace()}
	return r.reconcileResourceExistence(instance, expectedSecret, foundSecret, namespacedName)
}

func (r *ReconcileIBMLicenseServiceReporter) reconcileAPISecretToken(instance *operatorv1alpha1.IBMLicenseServiceReporter) (reconcile.Result, error) {
	reqLogger := log.WithValues("reconcileAPISecretToken", "Entry", "instance.GetName()", instance.GetName())
	expectedSecret, err := reporter.GetAPISecretToken(instance)
	if err != nil {
		reqLogger.Info("Failed to get expected secret")
		return reconcile.Result{
			Requeue:      true,
			RequeueAfter: time.Minute,
		}, err
	}
	foundSecret := &corev1.Secret{}
	namespacedName := types.NamespacedName{Name: expectedSecret.GetName(), Namespace: expectedSecret.GetNamespace()}
	return r.reconcileResourceExistence(instance, expectedSecret, foundSecret, namespacedName)
}

func (r *ReconcileIBMLicenseServiceReporter) reconcileService(instance *operatorv1alpha1.IBMLicenseServiceReporter) (reconcile.Result, error) {
	expectedService := reporter.GetService(instance)
	foundService := &corev1.Service{}
	namespacedName := types.NamespacedName{Name: expectedService.GetName(), Namespace: expectedService.GetNamespace()}
	return r.reconcileResourceExistence(instance, expectedService, foundService, namespacedName)
}

func (r *ReconcileIBMLicenseServiceReporter) reconcileDeployment(instance *operatorv1alpha1.IBMLicenseServiceReporter) (reconcile.Result, error) {
	reqLogger := log.WithValues("reconcileDeployment", "Entry", "instance.GetName()", instance.GetName())
	expectedDeployment := reporter.GetDeployment(instance)
	foundDeployment := &appsv1.Deployment{}
	namespacedName := types.NamespacedName{Name: expectedDeployment.GetName(), Namespace: expectedDeployment.GetNamespace()}
	reconcileResult, err := r.reconcileResourceExistence(instance, expectedDeployment, foundDeployment, namespacedName)
	if err != nil || reconcileResult.Requeue {
		return reconcileResult, err
	}

	shouldUpdate := res.ShouldUpdateDeployment(
		&reqLogger,
		&expectedDeployment.Spec.Template,
		&foundDeployment.Spec.Template,
		false,
	)

	if shouldUpdate {
		refreshedDeployment := foundDeployment.DeepCopy()
		refreshedDeployment.Spec.Template.Spec.Volumes = expectedDeployment.Spec.Template.Spec.Volumes
		refreshedDeployment.Spec.Template.Spec.Containers = expectedDeployment.Spec.Template.Spec.Containers
		refreshedDeployment.Spec.Template.Spec.InitContainers = expectedDeployment.Spec.Template.Spec.InitContainers
		refreshedDeployment.Spec.Template.Spec.ServiceAccountName = expectedDeployment.Spec.Template.Spec.ServiceAccountName
		refreshedDeployment.Spec.Template.Annotations = expectedDeployment.Spec.Template.Annotations
		reqLogger.Info("Updating Deployment Spec to", "RefreshedDeployment.Spec", refreshedDeployment.Spec)
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
		// Spec updated - return and do not requeue as it might not consider extra values
		return reconcile.Result{}, nil
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileIBMLicenseServiceReporter) reconcileReporterRoute(instance *operatorv1alpha1.IBMLicenseServiceReporter) (reconcile.Result, error) {
	expectedRoute := reporter.GetReporterRoute(instance)
	foundRoute := &routev1.Route{}
	namespacedName := types.NamespacedName{Name: expectedRoute.GetName(), Namespace: expectedRoute.GetNamespace()}
	return r.reconcileResourceExistence(instance, expectedRoute, foundRoute, namespacedName)
}

func (r *ReconcileIBMLicenseServiceReporter) reconcileUIIngress(instance *operatorv1alpha1.IBMLicenseServiceReporter) (reconcile.Result, error) {
	expectedIngress := reporter.GetUIIngress(instance)
	foundIngress := &extensionsv1.Ingress{}
	namespacedName := types.NamespacedName{Name: expectedIngress.GetName(), Namespace: expectedIngress.GetNamespace()}
	return r.reconcileResourceExistence(instance, expectedIngress, foundIngress, namespacedName)
}

func (r *ReconcileIBMLicenseServiceReporter) reconcileIngressProxy(instance *operatorv1alpha1.IBMLicenseServiceReporter) (reconcile.Result, error) {
	expectedIngress := reporter.GetUIIngressProxy(instance)
	foundIngress := &extensionsv1.Ingress{}
	namespacedName := types.NamespacedName{Name: expectedIngress.GetName(), Namespace: expectedIngress.GetNamespace()}
	return r.reconcileResourceExistence(instance, expectedIngress, foundIngress, namespacedName)
}

func (r *ReconcileIBMLicenseServiceReporter) reconcileResourceExistence(
	instance *operatorv1alpha1.IBMLicenseServiceReporter,
	expectedRes res.ResourceObject,
	foundRes runtime.Object,
	namespacedName types.NamespacedName) (reconcile.Result, error) {

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
