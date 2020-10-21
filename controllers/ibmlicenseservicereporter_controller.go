/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"time"

	routev1 "github.com/openshift/api/route/v1"

	"github.com/go-logr/logr"
	operatorv1alpha1 "github.com/ibm/ibm-licensing-operator/api/v1alpha1"
	res "github.com/ibm/ibm-licensing-operator/controllers/resources"
	"github.com/ibm/ibm-licensing-operator/controllers/resources/reporter"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1 "k8s.io/api/extensions/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type reconcileLRFunctionType = func(*operatorv1alpha1.IBMLicenseServiceReporter) (reconcile.Result, error)

// IBMLicenseServiceReporterReconciler reconciles a IBMLicenseServiceReporter object
type IBMLicenseServiceReporterReconciler struct {
	client.Client
	client.Reader
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=operator.ibm.com,resources=ibmlicenseservicereporters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operator.ibm.com,resources=ibmlicenseservicereporters/status,verbs=get;update;patch

func (r *IBMLicenseServiceReporterReconciler) Reconcile(request ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	reqLogger := r.Log.WithValues("ibmlicenseservicereporter", request.NamespacedName)

	var recResult reconcile.Result
	var recErr error

	reconcileFunctions := []interface{}{
		r.reconcileServiceAccount,
		r.reconcileRole,
		r.reconcileRoleBinding,
		r.reconcileAPISecretToken,
		r.reconcileDatabaseSecret,
		r.reconcilePersistentVolumeClaim,
		r.reconcileService,
		r.reconcileDeployment,
		r.reconcileReporterRoute,
		r.reconcileUIIngress,
		r.reconcileIngressProxy,
	}

	// Fetch the IBMLicenseServiceReporter instance
	foundInstance := &operatorv1alpha1.IBMLicenseServiceReporter{}
	err := r.Client.Get(context.TODO(), request.NamespacedName, foundInstance)
	if err != nil {
		if errors.IsNotFound(err) {
			res.IsReporterInstalled = false
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			// reqLogger.Info("IBMLicenseServiceReporter resource not found. Ignoring since object must be deleted")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	res.IsReporterInstalled = true

	instance := foundInstance.DeepCopy()

	err = reporter.UpdateVersion(r.Client, instance)
	if err != nil {
		reqLogger.Error(err, "Can not update version in CR")
	}

	res.UpdateCache(&reqLogger, r.Reader)

	err = instance.Spec.FillDefaultValues(reqLogger, r.Reader)
	if err != nil {
		return reconcile.Result{}, err
	}

	reqLogger.Info("got IBM License Service Reporter application, version=" + instance.Spec.Version)

	for _, reconcileFunction := range reconcileFunctions {
		recResult, recErr = reconcileFunction.(reconcileLRFunctionType)(instance)
		if recErr != nil || recResult.Requeue {
			return recResult, recErr
		}
	}

	// Update status logic, using foundInstance, because we do not want to add filled default values to yaml
	return r.updateStatus(foundInstance, reqLogger)
}

func (r *IBMLicenseServiceReporterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	log := r.Log.WithValues("SetupWithManager", "Entry")

	res.UpdateCache(&log, mgr.GetAPIReader())

	if res.IsRouteAPI {
		ctrl.NewControllerManagedBy(mgr).
			For(&operatorv1alpha1.IBMLicenseServiceReporter{}).
			Watches(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForObject{}).
			Watches(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForObject{}).
			Watches(&source.Kind{Type: &routev1.Route{}}, &handler.EnqueueRequestForObject{}).
			Complete(r)
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.IBMLicenseServiceReporter{}).
		Watches(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForObject{}).
		Watches(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForObject{}).
		Complete(r)

}

func (r *IBMLicenseServiceReporterReconciler) updateStatus(
	instance *operatorv1alpha1.IBMLicenseServiceReporter,
	reqLogger logr.Logger) (reconcile.Result, error) {
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(instance.GetNamespace()),
		client.MatchingLabels(reporter.LabelsForPod(instance)),
	}
	if err := r.Client.List(context.TODO(), podList, listOpts...); err != nil {
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
		reqLogger.Info("Updating IBMLicenseServiceReporter status")
		instance.Status.LicensingReporterPods = podStatuses
		err := r.Client.Status().Update(context.TODO(), instance)
		if err != nil {
			reqLogger.Info("Failed to update pod status")
		}
	}

	reqLogger.Info("reconcile all done")
	return reconcile.Result{}, nil
}

func (r *IBMLicenseServiceReporterReconciler) reconcileServiceAccount(instance *operatorv1alpha1.IBMLicenseServiceReporter) (reconcile.Result, error) {
	reqLogger := r.Log.WithValues("reconcileServiceAccount", "Entry", "instance.GetName()", instance.GetName())
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
		err = r.Client.Update(context.TODO(), foundSA)
		if err != nil {
			reqLogger.Error(err, "Failed to update ServiceAccount, deleting...")
			err = r.Client.Delete(context.TODO(), foundSA)
			if err != nil {
				reqLogger.Error(err, "Failed to delete ServiceAccount during recreation")
				return reconcile.Result{}, err
			}
			reqLogger.Info("Deleted ServiceAccount successfully")
			return reconcile.Result{Requeue: true, RequeueAfter: time.Minute}, nil
		}
		reqLogger.Info("Updated ServiceAccount successfully")
		// Spec updated - return and do not requeue as it might not consider extra values
		return reconcile.Result{}, nil
	}
	return reconcile.Result{}, nil
}

func (r *IBMLicenseServiceReporterReconciler) reconcileRole(instance *operatorv1alpha1.IBMLicenseServiceReporter) (reconcile.Result, error) {
	expectedRole := reporter.GetRole(instance)
	foundRole := &rbacv1.Role{}
	namespacedName := types.NamespacedName{Name: expectedRole.GetName(), Namespace: expectedRole.GetNamespace()}
	return r.reconcileResourceExistence(instance, expectedRole, foundRole, namespacedName)
}

func (r *IBMLicenseServiceReporterReconciler) reconcileRoleBinding(instance *operatorv1alpha1.IBMLicenseServiceReporter) (reconcile.Result, error) {
	expectedRoleBinding := reporter.GetRoleBinding(instance)
	foundRoleBinding := &rbacv1.RoleBinding{}
	namespacedName := types.NamespacedName{Name: expectedRoleBinding.GetName(), Namespace: expectedRoleBinding.GetNamespace()}
	return r.reconcileResourceExistence(instance, expectedRoleBinding, foundRoleBinding, namespacedName)
}

func (r *IBMLicenseServiceReporterReconciler) reconcilePersistentVolumeClaim(instance *operatorv1alpha1.IBMLicenseServiceReporter) (reconcile.Result, error) {

	expectedPVC := reporter.GetPersistenceVolumeClaim(instance)
	foundPVC := &corev1.PersistentVolumeClaim{}
	namespacedName := types.NamespacedName{Name: expectedPVC.GetName(), Namespace: expectedPVC.GetNamespace()}
	reconcileResult, err := r.reconcileResourceExistence(instance, expectedPVC, foundPVC, namespacedName)
	if err != nil || reconcileResult.Requeue {
		return reconcileResult, err
	}
	return reconcile.Result{}, nil

}

func (r *IBMLicenseServiceReporterReconciler) reconcileDatabaseSecret(instance *operatorv1alpha1.IBMLicenseServiceReporter) (reconcile.Result, error) {
	reqLogger := r.Log.WithValues("reconcileDatabaseSecret", "Entry", "instance.GetName()", instance.GetName())
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

func (r *IBMLicenseServiceReporterReconciler) reconcileAPISecretToken(instance *operatorv1alpha1.IBMLicenseServiceReporter) (reconcile.Result, error) {
	reqLogger := r.Log.WithValues("reconcileAPISecretToken", "Entry", "instance.GetName()", instance.GetName())
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

func (r *IBMLicenseServiceReporterReconciler) reconcileService(instance *operatorv1alpha1.IBMLicenseServiceReporter) (reconcile.Result, error) {
	reqLogger := r.Log.WithValues("reconcileService", "Entry", "instance.GetName()", instance.GetName())
	expectedService := reporter.GetService(instance)
	foundService := &corev1.Service{}
	namespacedName := types.NamespacedName{Name: expectedService.GetName(), Namespace: expectedService.GetNamespace()}
	reconcileResult, err := r.reconcileResourceExistence(instance, expectedService, foundService, namespacedName)
	if err != nil || reconcileResult.Requeue {
		return reconcileResult, err
	}
	return res.UpdateServiceIfNeeded(&reqLogger, r.Client, expectedService, foundService)
}

func (r *IBMLicenseServiceReporterReconciler) reconcileDeployment(instance *operatorv1alpha1.IBMLicenseServiceReporter) (reconcile.Result, error) {
	reqLogger := r.Log.WithValues("reconcileDeployment", "Entry", "instance.GetName()", instance.GetName())
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
	)

	if shouldUpdate {
		return res.UpdateResource(&reqLogger, r.Client, expectedDeployment, foundDeployment)
	}

	return reconcile.Result{}, nil
}

func (r *IBMLicenseServiceReporterReconciler) reconcileReporterRoute(instance *operatorv1alpha1.IBMLicenseServiceReporter) (reconcile.Result, error) {
	expectedRoute := reporter.GetReporterRoute(instance)
	foundRoute := &routev1.Route{}
	namespacedName := types.NamespacedName{Name: expectedRoute.GetName(), Namespace: expectedRoute.GetNamespace()}
	return r.reconcileResourceExistence(instance, expectedRoute, foundRoute, namespacedName)
}

func (r *IBMLicenseServiceReporterReconciler) reconcileUIIngress(instance *operatorv1alpha1.IBMLicenseServiceReporter) (reconcile.Result, error) {
	expectedIngress := reporter.GetUIIngress(instance)
	foundIngress := &extensionsv1.Ingress{}
	namespacedName := types.NamespacedName{Name: expectedIngress.GetName(), Namespace: expectedIngress.GetNamespace()}
	return r.reconcileResourceExistence(instance, expectedIngress, foundIngress, namespacedName)
}

func (r *IBMLicenseServiceReporterReconciler) reconcileIngressProxy(instance *operatorv1alpha1.IBMLicenseServiceReporter) (reconcile.Result, error) {
	expectedIngress := reporter.GetUIIngressProxy(instance)
	foundIngress := &extensionsv1.Ingress{}
	namespacedName := types.NamespacedName{Name: expectedIngress.GetName(), Namespace: expectedIngress.GetNamespace()}
	return r.reconcileResourceExistence(instance, expectedIngress, foundIngress, namespacedName)
}

func (r *IBMLicenseServiceReporterReconciler) reconcileResourceExistence(
	instance *operatorv1alpha1.IBMLicenseServiceReporter,
	expectedRes res.ResourceObject,
	foundRes runtime.Object,
	namespacedName types.NamespacedName) (reconcile.Result, error) {

	resType := reflect.TypeOf(expectedRes)
	reqLogger := r.Log.WithValues(resType.String(), "Entry", "instance.GetName()", instance.GetName())

	// expectedRes already set before and passed via parameter
	err := controllerutil.SetControllerReference(instance, expectedRes, r.Scheme)
	if err != nil {
		reqLogger.Error(err, "Failed to define expected resource")
		return reconcile.Result{}, err
	}

	// foundRes already initialized before and passed via parameter
	err = r.Client.Get(context.TODO(), namespacedName, foundRes)
	if err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info(resType.String()+" does not exist, trying creating new one", "Name", expectedRes.GetName(),
				"Namespace", expectedRes.GetNamespace())
			err = r.Client.Create(context.TODO(), expectedRes)
			if err != nil {
				if !errors.IsAlreadyExists(err) {
					reqLogger.Error(err, "Failed to create new "+resType.String(), "Name", expectedRes.GetName(),
						"Namespace", expectedRes.GetNamespace())
					return reconcile.Result{}, err
				}
			}
			// Created successfully, or already exists - return and requeue
			time.Sleep(time.Second * 5)
			return reconcile.Result{Requeue: true, RequeueAfter: time.Second}, nil
		}
		reqLogger.Error(err, "Failed to get "+resType.String(), "Name", expectedRes.GetName(),
			"Namespace", expectedRes.GetNamespace())
		return reconcile.Result{}, err
	}
	reqLogger.Info(resType.String() + " is correct!")
	return reconcile.Result{}, nil
}
