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
	"time"

	"github.com/go-logr/logr"

	operatorv1alpha1 "github.com/ibm/ibm-licensing-operator/pkg/apis/operator/v1alpha1"
	res "github.com/ibm/ibm-licensing-operator/pkg/resources"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1 "k8s.io/api/extensions/v1beta1"
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

var (
	log                = logf.Log.WithName("controller_ibmlicensing")
	isOpenshiftCluster = true
)

//var isOldIngressVersion = false

type reconcileFunctionType = func(*operatorv1alpha1.IBMLicensing) (reconcile.Result, error)

// Add creates a new IBMLicensing Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileIBMLicensing{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

type ResourceObject interface {
	metav1.Object
	runtime.Object
}

func watchForResources(c controller.Controller, watchTypes []ResourceObject) error {
	for _, restype := range watchTypes {
		log.Info("Watching", "restype", restype)
		err := c.Watch(&source.Kind{Type: restype}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &operatorv1alpha1.IBMLicensing{},
		})
		if err != nil {
			return err
		}
	}
	return nil
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

	// Watch for changes to secondary resources
	err = watchForResources(c, []ResourceObject{
		&rbacv1.Role{},
		&rbacv1.RoleBinding{},
		&rbacv1.ClusterRole{},
		&rbacv1.ClusterRoleBinding{},
		&corev1.ServiceAccount{},
		&corev1.Secret{},
		&appsv1.Deployment{},
		&corev1.Service{},
		&extensionsv1.Ingress{},
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
		err = watchForResources(c, []ResourceObject{
			&routev1.Route{},
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
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileIBMLicensing) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request", request)
	reqLogger.Info("Reconciling IBMLicensing")

	// Fetch the IBMLicensing instance
	foundInstance := &operatorv1alpha1.IBMLicensing{}
	err := r.client.Get(context.TODO(), request.NamespacedName, foundInstance)
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
	instance := foundInstance.DeepCopy()
	instance.Spec.FillDefaultValues(isOpenshiftCluster)

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
		r.reconcileIngress,
	}

	for _, reconcileFunction := range reconcileFunctions {
		recResult, recErr = reconcileFunction.(reconcileFunctionType)(instance)
		if recErr != nil || recResult.Requeue {
			return recResult, recErr
		}
	}

	if isOpenshiftCluster {
		reconcileOpenShiftFunctions := []interface{}{
			r.reconcileRoute,
		}

		for _, reconcileFunction := range reconcileOpenShiftFunctions {
			recResult, recErr = reconcileFunction.(reconcileFunctionType)(instance)
			if recErr != nil || recResult.Requeue {
				return recResult, recErr
			}
		}
	}

	// Update status logic, using foundInstance, because we do not want to add filled default values to yaml
	return r.updateStatus(foundInstance, reqLogger)
}

func (r *ReconcileIBMLicensing) updateStatus(instance *operatorv1alpha1.IBMLicensing, reqLogger logr.Logger) (reconcile.Result, error) {
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(instance.Spec.InstanceNamespace),
		client.MatchingLabels(res.LabelsForLicensingPod(instance)),
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

	if !reflect.DeepEqual(podStatuses, instance.Status.LicensingPods) {
		reqLogger.Info("Updating IBMLicensing status", "Instance", instance)
		instance.Status.LicensingPods = podStatuses
		err := r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			reqLogger.Info("Failed to update pod status")
		}
	}

	reqLogger.Info("reconcile all done")
	return reconcile.Result{}, nil
}

func (r *ReconcileIBMLicensing) reconcileServiceAccount(instance *operatorv1alpha1.IBMLicensing) (reconcile.Result, error) {
	reqLogger := log.WithValues("reconcileServiceAccount", "Entry", "instance.GetName()", instance.GetName())
	expectedSA := res.GetLicensingServiceAccount(instance)
	foundSA := &corev1.ServiceAccount{}
	reconcileResult, err := r.reconcileResourceNamespacedExistence(instance, expectedSA, foundSA)
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
		// Spec updated - return and do not requeue as it might not consider extra values
		return reconcile.Result{}, nil
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileIBMLicensing) reconcileRole(instance *operatorv1alpha1.IBMLicensing) (reconcile.Result, error) {
	expectedRole := res.GetLicensingRole(instance)
	foundRole := &rbacv1.Role{}
	return r.reconcileResourceNamespacedExistence(instance, expectedRole, foundRole)
}

func (r *ReconcileIBMLicensing) reconcileRoleBinding(instance *operatorv1alpha1.IBMLicensing) (reconcile.Result, error) {
	expectedRoleBinding := res.GetLicensingRoleBinding(instance)
	foundRoleBinding := &rbacv1.RoleBinding{}
	return r.reconcileResourceNamespacedExistence(instance, expectedRoleBinding, foundRoleBinding)
}

func (r *ReconcileIBMLicensing) reconcileClusterRole(instance *operatorv1alpha1.IBMLicensing) (reconcile.Result, error) {
	expectedClusterRole := res.GetLicensingClusterRole(instance)
	foundClusterRole := &rbacv1.ClusterRole{}
	return r.reconcileResourceClusterExistence(instance, expectedClusterRole, foundClusterRole)
}

func (r *ReconcileIBMLicensing) reconcileClusterRoleBinding(instance *operatorv1alpha1.IBMLicensing) (reconcile.Result, error) {
	expectedClusterRoleBinding := res.GetLicensingClusterRoleBinding(instance)
	foundClusterRoleBinding := &rbacv1.ClusterRoleBinding{}
	return r.reconcileResourceClusterExistence(instance, expectedClusterRoleBinding, foundClusterRoleBinding)
}

func (r *ReconcileIBMLicensing) reconcileNamespace(instance *operatorv1alpha1.IBMLicensing) (reconcile.Result, error) {
	expectedNamespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: instance.Spec.InstanceNamespace,
		},
	}
	foundNamespace := &corev1.Namespace{}
	return r.reconcileResourceClusterExistence(instance, expectedNamespace, foundNamespace)
}

func (r *ReconcileIBMLicensing) reconcileAPISecretToken(instance *operatorv1alpha1.IBMLicensing) (reconcile.Result, error) {
	metaLabels := res.LabelsForLicensingMeta(instance)
	expectedSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Spec.APISecretToken,
			Namespace: instance.Spec.InstanceNamespace,
			Labels:    metaLabels,
		},
		Type:       corev1.SecretTypeOpaque,
		StringData: map[string]string{"token": res.RandString(24)},
	}
	foundSecret := &corev1.Secret{}
	return r.reconcileResourceNamespacedExistence(instance, expectedSecret, foundSecret)
}

func (r *ReconcileIBMLicensing) reconcileService(instance *operatorv1alpha1.IBMLicensing) (reconcile.Result, error) {
	expectedService := res.GetLicensingService(instance)
	foundService := &corev1.Service{}
	return r.reconcileResourceNamespacedExistence(instance, expectedService, foundService)
}

func (r *ReconcileIBMLicensing) reconcileDeployment(instance *operatorv1alpha1.IBMLicensing) (reconcile.Result, error) {
	reqLogger := log.WithValues("reconcileDeployment", "Entry", "instance.GetName()", instance.GetName())
	expectedDeployment := res.GetLicensingDeployment(instance)
	shouldUpdate := true
	foundDeployment := &appsv1.Deployment{}
	reconcileResult, err := r.reconcileResourceNamespacedExistence(instance, expectedDeployment, foundDeployment)
	if err != nil || reconcileResult.Requeue {
		return reconcileResult, err
	}
	// TODO: this should be refactored in some nice way where you only declare which parameters needs to be correct between resources
	foundSpec := foundDeployment.Spec.Template.Spec
	expectedSpec := expectedDeployment.Spec.Template.Spec
	if !reflect.DeepEqual(foundSpec.Volumes, expectedSpec.Volumes) {
		reqLogger.Info("Deployment has wrong volumes")
	} else if foundSpec.ServiceAccountName != expectedSpec.ServiceAccountName {
		reqLogger.Info("Deployment wrong service account name")
	} else if len(foundSpec.Containers) != len(expectedSpec.Containers) {
		reqLogger.Info("Deployment has wrong number of containers")
	} else if len(foundSpec.InitContainers) != len(expectedSpec.InitContainers) {
		reqLogger.Info("Deployment has wrong number of init containers")
	} else {
		shouldUpdate = false
		containersToBeChecked := map[*corev1.Container]corev1.Container{&foundSpec.Containers[0]: expectedSpec.Containers[0]}
		if instance.Spec.IsMetering() {
			containersToBeChecked[&foundSpec.InitContainers[0]] = expectedSpec.InitContainers[0]
		}
		for foundContainer, expectedContainer := range containersToBeChecked {
			if shouldUpdate {
				break
			}
			shouldUpdate = true
			if foundContainer.Name != expectedContainer.Name {
				reqLogger.Info("Deployment wrong container name")
			} else if foundContainer.Image != expectedContainer.Image {
				reqLogger.Info("Deployment wrong container image")
			} else if foundContainer.ImagePullPolicy != expectedContainer.ImagePullPolicy {
				reqLogger.Info("Deployment wrong image pull policy")
			} else if !reflect.DeepEqual(foundContainer.Command, expectedContainer.Command) {
				reqLogger.Info("Deployment wrong container command")
			} else if !reflect.DeepEqual(foundContainer.Ports, expectedContainer.Ports) {
				reqLogger.Info("Deployment wrong containers ports")
			} else if !reflect.DeepEqual(foundContainer.VolumeMounts, expectedContainer.VolumeMounts) {
				reqLogger.Info("Deployment wrong VolumeMounts in container")
			} else if !reflect.DeepEqual(foundContainer.Env, expectedContainer.Env) {
				reqLogger.Info("Deployment wrong env variables in container")
			} else if !reflect.DeepEqual(foundContainer.SecurityContext, expectedContainer.SecurityContext) {
				reqLogger.Info("Deployment wrong container security context")
			} else if (foundContainer.Resources.Limits == nil) || (foundContainer.Resources.Requests == nil) {
				reqLogger.Info("Deployment wrong container Resources")
			} else if !(foundContainer.Resources.Limits.Cpu().Equal(*expectedContainer.Resources.Limits.Cpu()) &&
				foundContainer.Resources.Limits.Memory().Equal(*expectedContainer.Resources.Limits.Memory())) {
				reqLogger.Info("Deployment wrong container Resources Limits")
			} else if !(foundContainer.Resources.Requests.Cpu().Equal(*expectedContainer.Resources.Requests.Cpu()) &&
				foundContainer.Resources.Requests.Memory().Equal(*expectedContainer.Resources.Requests.Memory())) {
				reqLogger.Info("Deployment wrong container Resources Requests")
			} else {
				shouldUpdate = false
			}
		}
	}

	if shouldUpdate {
		refreshedDeployment := foundDeployment.DeepCopy()
		refreshedDeployment.Spec.Template.Spec.Volumes = expectedDeployment.Spec.Template.Spec.Volumes
		refreshedDeployment.Spec.Template.Spec.Containers = expectedDeployment.Spec.Template.Spec.Containers
		refreshedDeployment.Spec.Template.Spec.InitContainers = expectedDeployment.Spec.Template.Spec.InitContainers
		refreshedDeployment.Spec.Template.Spec.ServiceAccountName = expectedDeployment.Spec.Template.Spec.ServiceAccountName
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

func (r *ReconcileIBMLicensing) reconcileRoute(instance *operatorv1alpha1.IBMLicensing) (reconcile.Result, error) {
	if instance.Spec.IsRouteEnabled() {
		expectedRoute := res.GetLicensingRoute(instance)
		foundRoute := &routev1.Route{}
		return r.reconcileResourceNamespacedExistence(instance, expectedRoute, foundRoute)
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileIBMLicensing) reconcileIngress(instance *operatorv1alpha1.IBMLicensing) (reconcile.Result, error) {
	if instance.Spec.IsIngressEnabled() {
		expectedIngress := res.GetLicensingIngress(instance)
		foundIngress := &extensionsv1.Ingress{}
		return r.reconcileResourceNamespacedExistence(instance, expectedIngress, foundIngress)
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileIBMLicensing) reconcileResourceNamespacedExistence(
	instance *operatorv1alpha1.IBMLicensing, expectedRes ResourceObject, foundRes runtime.Object) (reconcile.Result, error) {

	namespacedName := types.NamespacedName{Name: expectedRes.GetName(), Namespace: expectedRes.GetNamespace()}
	return r.reconcileResourceExistence(instance, expectedRes, foundRes, namespacedName)
}

func (r *ReconcileIBMLicensing) reconcileResourceClusterExistence(
	instance *operatorv1alpha1.IBMLicensing, expectedRes ResourceObject, foundRes runtime.Object) (reconcile.Result, error) {

	namespacedName := types.NamespacedName{Name: expectedRes.GetName()}
	return r.reconcileResourceExistence(instance, expectedRes, foundRes, namespacedName)
}

func (r *ReconcileIBMLicensing) reconcileResourceExistence(
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
