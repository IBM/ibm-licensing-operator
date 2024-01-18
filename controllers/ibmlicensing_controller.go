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
	"fmt"
	"reflect"
	goruntime "runtime"
	"sort"
	"strings"
	"time"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/go-logr/logr"
	routev1 "github.com/openshift/api/route/v1"
	rhmp "github.com/redhat-marketplace/redhat-marketplace-operator/v2/apis/marketplace/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metaErrors "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	operatorv1alpha1 "github.com/IBM/ibm-licensing-operator/api/v1alpha1"
	res "github.com/IBM/ibm-licensing-operator/controllers/resources"
	"github.com/IBM/ibm-licensing-operator/controllers/resources/service"
)

type reconcileLSFunctionType = func(*operatorv1alpha1.IBMLicensing) (reconcile.Result, error)

func (r *IBMLicensingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := res.UpdateCacheClusterExtensions(mgr.GetAPIReader()); err != nil {
		r.Log.Error(err, "Error during checking K8s API")
	}

	watcher := ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.IBMLicensing{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{})

	if res.IsRouteAPI {
		watcher.Owns(&operatorv1alpha1.IBMLicenseServiceReporter{})
	}

	return watcher.Complete(r)
}

// blank assignment to verify that IBMLicensingReconciler implements reconcile.Reconciler
var _ reconcile.Reconciler = &IBMLicensingReconciler{}

// IBMLicensingReconciler reconciles a IBMLicensing object
type IBMLicensingReconciler struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client.Client
	client.Reader
	Log               logr.Logger
	Scheme            *runtime.Scheme
	OperatorNamespace string
}

// //kubebuilder:rbac:namespace=ibm-common-services,groups=,resources=pod,verbs=get;list;watch;create;update;patch;delete

// Reconcile reads that state of the cluster for a IBMLicensing object and makes changes based on the state read
// and what is in the IBMLicensing.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.

// +kubebuilder:rbac:namespace=ibm-common-services,groups=operator.ibm.com,resources=ibmlicensings;ibmlicensings/status;ibmlicensings/finalizers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:namespace=ibm-common-services,groups="apps",resources=deployments/finalizers,verbs=update
// +kubebuilder:rbac:namespace=ibm-common-services,groups=monitoring.coreos.com,resources=servicemonitors,verbs=get;create;watch;list;delete;update
// +kubebuilder:rbac:namespace=ibm-common-services,groups=rbac.authorization.k8s.io,resources=roles;rolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:namespace=ibm-common-services,groups="",resources=nodes;namespaces,verbs=get;list;watch
// +kubebuilder:rbac:namespace=ibm-common-services,groups=route.openshift.io,resources=routes;routes/custom-host,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:namespace=ibm-common-services,groups=marketplace.redhat.com,resources=meterdefinitions,verbs=get;list;create;update;watch
// +kubebuilder:rbac:namespace=ibm-common-services,groups=networking.k8s.io;extensions,resources=ingresses;networkpolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:namespace=ibm-common-services,groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:namespace=ibm-common-services,groups=apps,resources=daemonsets;replicasets;statefulsets,verbs=get;list;watch
// +kubebuilder:rbac:namespace=ibm-common-services,groups="",resources=pods;services;services/finalizers;endpoints;events;configmaps;secrets;serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operator.openshift.io,resources=servicecas,verbs=list
// +kubebuilder:rbac:groups=operator.ibm.com,resources=ibmlicensings;ibmlicensings/status;ibmlicensings/finalizers,verbs=get;list;watch;create;update;patch;delete

//nolint:revive
func (r *IBMLicensingReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {

	reqLogger := r.Log.WithValues("ibmlicensing", req.NamespacedName)
	reqLogger.Info("Reconciling IBMLicensing")
	goruntime.GC()

	if err := res.UpdateCacheClusterExtensions(r.Reader); err != nil {
		reqLogger.Error(err, "Error during checking K8s API")
	}

	// Check if there are already IBMLicensing instances created
	ibmlicensingList := &operatorv1alpha1.IBMLicensingList{}
	if err := r.Client.List(context.TODO(), ibmlicensingList); err != nil {
		// Error when looking for IMBLicensing objects - requeue
		reqLogger.Error(err, "Couldn't retrieve IBMLicensing objects. Retrying.")
		return reconcile.Result{}, err
	}

	// Fetch the IBMLicensing instance
	foundInstance := &operatorv1alpha1.IBMLicensing{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, foundInstance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile req.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			// reqLogger.Info("IBMLicensing resource not found. Ignoring since object must be deleted")

			// In case of deleting active instance, detect new one
			if !hasIBMLicensingListActiveInstance(ibmlicensingList) {
				return reconcile.Result{}, r.findAndMarkActiveIBMLicensing(ibmlicensingList)
			}

			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the req.
		// reqLogger.Error(err, "Failed to get IBMLicensing")
		return reconcile.Result{}, err
	}

	instance := foundInstance.DeepCopy()

	// Check if there are any active CR or if they are properly marked (field .State)
	if !hasIBMLicensingListActiveInstance(ibmlicensingList) || instance.Status.State == "" {
		err := r.findAndMarkActiveIBMLicensing(ibmlicensingList)
		if err != nil {
			reqLogger.Error(err, "Failed to update IBMLicensing CR status.")
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	}

	// Ignore reconciliation if CR is 'inactive'
	if instance.Status.State == service.InactiveCRState {
		reqLogger.Info("Ignoring reconciliation because CR's status is " + instance.Status.State)
		return reconcile.Result{}, nil
	}

	// Ignore reconciliation if CR version is v4/CP3
	if strings.HasPrefix(instance.Spec.Version, "4") {
		reqLogger.Info("Detected newer version of License Service running. Ignoring reconciliation. IBMLicensing CR instance spec.version=" + instance.Spec.Version)
		return reconcile.Result{}, nil
	}

	err = service.UpdateVersion(r.Client, instance)
	if err != nil {
		reqLogger.Error(err, "Can not update version in CR")
	}

	err = instance.Spec.FillDefaultValues(reqLogger, res.IsServiceCAAPI, res.IsRouteAPI, res.RHMPEnabled,
		res.IsAlertingEnabledByDefault, r.OperatorNamespace)
	if err != nil {
		return reconcile.Result{}, err
	}
	r.controllerStatus(instance)

	reqLogger.Info("got IBM License Service application, version=" + instance.Spec.Version)

	var recResult reconcile.Result

	reconcileFunctions := []interface{}{
		r.reconcileAPISecretToken,
		r.reconcileUploadToken,
		r.reconcileDefaultReaderToken,
		r.reconcileServiceAccountToken,
		r.reconcileServices,
		r.reconcileIngress,
		r.reconcileRouteWithoutCertificates,
		r.reconcileCertificateSecrets,
		r.reconcileRouteWithCertificates,
		r.reconcileConfigMaps,
		r.reconcileDeployment,
		r.reconcileNetworkPolicy,
		r.reconcileRHMPServiceMonitor,
		r.reconcileAlertingServiceMonitor,
		r.reconcileMeterDefinition,
	}

	for _, reconcileFunction := range reconcileFunctions {
		recResult, err = reconcileFunction.(reconcileLSFunctionType)(instance)
		if err != nil || recResult.Requeue {
			return recResult, err
		}
	}

	// Update status logic, using foundInstance, because we do not want to add filled default values to yaml
	return r.updateStatus(foundInstance, reqLogger)
}

func (r *IBMLicensingReconciler) findAndMarkActiveIBMLicensing(ibmlicensingList *operatorv1alpha1.IBMLicensingList) error {
	if ibmlicensingList.Items == nil || len(ibmlicensingList.Items) == 0 {
		return nil
	}

	// Sort by creation timestamp
	sort.SliceStable(ibmlicensingList.Items, func(i, j int) bool {
		return ibmlicensingList.Items[i].ObjectMeta.CreationTimestamp.Time.Before(ibmlicensingList.Items[j].ObjectMeta.CreationTimestamp.Time)
	})

	// First element is oldest one and should only be active
	initialInstance := ibmlicensingList.Items[0]

	var cr operatorv1alpha1.IBMLicensing
	// Mark all CRs states depending on their creation time
	for _, cr = range ibmlicensingList.Items {
		// Only firstly created instance is marked as 'active' and will be reconciled
		if cr.UID == initialInstance.UID {
			r.Log.Info("Due to having first creation timestamp the active IBMLicensing instance CR is named: " + cr.Name)
			cr.Status.State = service.ActiveCRState
		} else {
			r.Log.Info("Other IBMLicensing instance already exists and is active! Ignoring CR: " + cr.Name)
			// CR should be marked as 'inactive' and ignored during next reconciliation
			if cr.Status.State != service.InactiveCRState {
				cr.Status.State = service.InactiveCRState
			}
		}
		err := r.Client.Status().Update(context.TODO(), &cr)
		if err != nil {
			return err
		}
	}

	return nil
}

func hasIBMLicensingListActiveInstance(ibmlicensingList *operatorv1alpha1.IBMLicensingList) bool {
	// Iterate over the ibmlicensingList items and check if there is any active CR
	for _, s := range ibmlicensingList.Items {
		if s.Status.State == service.ActiveCRState {
			return true
		}
	}
	return false
}

func (r *IBMLicensingReconciler) updateStatus(instance *operatorv1alpha1.IBMLicensing, reqLogger logr.Logger) (reconcile.Result, error) {
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(instance.Spec.InstanceNamespace),
		client.MatchingLabels(service.LabelsForLicensingPod(instance)),
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

	var featuresStatuses operatorv1alpha1.IBMLicensingFeaturesStatus

	var rhmpEnabled bool
	if instance.Spec.RHMPEnabled == nil {
		rhmpEnabled = res.RHMPEnabled
	} else {
		rhmpEnabled = *instance.Spec.RHMPEnabled
	}

	featuresStatuses.RHMPEnabled = &rhmpEnabled

	if !reflect.DeepEqual(podStatuses, instance.Status.LicensingPods) || !reflect.DeepEqual(featuresStatuses, instance.Status.Features) {
		reqLogger.Info("Updating IBMLicensing status")
		instance.Status.LicensingPods = podStatuses
		instance.Status.Features = featuresStatuses
		err := r.Client.Status().Update(context.TODO(), instance)
		if err != nil {
			reqLogger.Info("Failed to update pod status, this does not affect License Service")
		}
	}

	reqLogger.Info("reconcile all done")
	return reconcile.Result{}, nil
}

func (r *IBMLicensingReconciler) reconcileAPISecretToken(instance *operatorv1alpha1.IBMLicensing) (reconcile.Result, error) {
	reqLogger := r.Log.WithValues("reconcileAPISecretToken", "Entry", "instance.GetName()", instance.GetName())
	expectedSecret, err := service.GetAPISecretToken(instance)
	if err != nil {
		reqLogger.Info("Failed to get expected secret")
		return reconcile.Result{
			Requeue:      true,
			RequeueAfter: time.Minute,
		}, err
	}
	foundSecret := &corev1.Secret{}
	return r.reconcileResourceNamespacedExistence(instance, expectedSecret, foundSecret)
}

// default reader token is not created by default since kubernetes 1.24, we need to ensure it is always generated
// having two default reader tokens for previous k8s is not a problem, you can use either one, and both will be cleaned
func (r *IBMLicensingReconciler) reconcileDefaultReaderToken(instance *operatorv1alpha1.IBMLicensing) (reconcile.Result, error) {
	reqLogger := r.Log.WithValues("reconcileDefaultReaderToken", "Entry", "instance.GetName()", instance.GetName())
	expectedSecret, err := service.GetDefaultReaderToken(instance)
	if err != nil {
		reqLogger.Info("Failed to get expected secret")
		return reconcile.Result{
			Requeue:      true,
			RequeueAfter: time.Minute,
		}, err
	}
	foundSecret := &corev1.Secret{}
	result, err := r.reconcileResourceNamespacedExistence(instance, expectedSecret, foundSecret)
	if err != nil || result.Requeue {
		return result, err
	}
	if expectedSecret.Annotations[service.ServiceAccountSecretAnnotationKey] !=
		foundSecret.Annotations[service.ServiceAccountSecretAnnotationKey] {
		err = r.Client.Delete(context.TODO(), foundSecret)
		if err != nil {
			reqLogger.Error(err, "Failed to delete ServiceAccount secret due to wrong annotations.")
			return reconcile.Result{}, err
		}
		return reconcile.Result{
			Requeue:      true,
			RequeueAfter: time.Minute,
		}, err
	}
	return result, err
}

func (r *IBMLicensingReconciler) reconcileServiceAccountToken(instance *operatorv1alpha1.IBMLicensing) (reconcile.Result, error) {
	if instance.Spec.IsAlertingEnabled() {
		reqLogger := r.Log.WithValues("reconcileServiceAccountToken", "Entry", "instance.GetName()", instance.GetName())
		expectedSecret, err := service.GetServiceAccountSecret(instance)
		if err != nil {
			reqLogger.Info("Failed to get expected secret")
			return reconcile.Result{
				Requeue:      true,
				RequeueAfter: time.Minute,
			}, err
		}
		foundSecret := &corev1.Secret{}
		result, err := r.reconcileResourceNamespacedExistence(instance, expectedSecret, foundSecret)
		if err != nil || result.Requeue {
			return result, err
		}
		if expectedSecret.Annotations[service.ServiceAccountSecretAnnotationKey] !=
			foundSecret.Annotations[service.ServiceAccountSecretAnnotationKey] {
			err = r.Client.Delete(context.TODO(), foundSecret)
			if err != nil {
				reqLogger.Error(err, "Failed to delete ServiceAccount secret due to wrong annotations.")
				return reconcile.Result{}, err
			}
			return reconcile.Result{
				Requeue:      true,
				RequeueAfter: time.Minute,
			}, err
		}
		return result, err
	}
	return reconcile.Result{}, nil
}

func (r *IBMLicensingReconciler) reconcileUploadToken(instance *operatorv1alpha1.IBMLicensing) (reconcile.Result, error) {
	reqLogger := r.Log.WithValues("reconcileUploadToken", "Entry", "instance.GetName()", instance.GetName())
	expectedSecret, err := service.GetUploadToken(instance)
	if err != nil {
		reqLogger.Info("Failed to get expected secret")
		return reconcile.Result{
			Requeue:      true,
			RequeueAfter: time.Minute,
		}, err
	}
	foundSecret := &corev1.Secret{}
	return r.reconcileResourceNamespacedExistence(instance, expectedSecret, foundSecret)
}

func (r *IBMLicensingReconciler) reconcileConfigMaps(instance *operatorv1alpha1.IBMLicensing) (reconcile.Result, error) {
	reqLogger := r.Log.WithValues("reconcileConfigMaps", "Entry", "instance.GetName()", instance.GetName())

	internalCertificate := &corev1.Secret{}
	certificateNamespacedName := types.NamespacedName{Namespace: instance.Spec.InstanceNamespace, Name: service.LicenseServiceInternalCertName}

	if err := r.Client.Get(context.TODO(), certificateNamespacedName, internalCertificate); err != nil {
		// Generate certificate only when route/ingress is enabled
		if instance.Spec.IsRouteEnabled() || instance.Spec.IsIngressEnabled() {
			r.Log.WithValues("cert name", certificateNamespacedName).Info("certificate secret not existing. Generating self signed certificate")
			return reconcile.Result{Requeue: true}, err
		}

		// Skip verification of certificates when route/ingress is disabled
		return reconcile.Result{}, nil
	}

	expectedCMs := []*corev1.ConfigMap{
		service.GetUploadConfigMap(instance, string(internalCertificate.Data["tls.crt"])),
		service.GetInfoConfigMap(instance),
	}
	for _, expectedCM := range expectedCMs {
		foundCM := &corev1.ConfigMap{}
		reconcileResult, err := r.reconcileResourceNamespacedExistence(instance, expectedCM, foundCM)
		if err != nil || reconcileResult.Requeue {
			return reconcileResult, err
		}
		if !res.CompareConfigMap(expectedCM, foundCM) {
			if updateReconcileResult, err := res.UpdateResource(&reqLogger, r.Client, expectedCM, foundCM); err != nil || updateReconcileResult.Requeue {
				return updateReconcileResult, err
			}
		}

	}
	return reconcile.Result{}, nil
}

func (r *IBMLicensingReconciler) reconcileServices(instance *operatorv1alpha1.IBMLicensing) (reconcile.Result, error) {
	var (
		result reconcile.Result
		err    error
	)
	reqLogger := r.Log.WithValues("reconcileServices", "Entry", "instance.GetName()", instance.GetName())
	expected, notExpected := service.GetServices(instance)
	found := &corev1.Service{}
	for _, es := range expected {
		result, err = r.reconcileResourceNamespacedExistence(instance, es, found)
		if err != nil || result.Requeue {
			return result, err
		}
		result, err = res.UpdateServiceIfNeeded(&reqLogger, r.Client, es, found)
	}

	for _, ne := range notExpected {
		result, err = r.reconcileNamespacedResourceWhichShouldNotExist(instance, ne, found)
		if err != nil || result.Requeue {
			return result, err
		}
	}

	return result, err
}

func (r *IBMLicensingReconciler) reconcileRHMPServiceMonitor(instance *operatorv1alpha1.IBMLicensing) (reconcile.Result, error) {
	expectedServiceMonitor := service.GetRHMPServiceMonitor(instance)
	shouldDelete := !instance.Spec.IsRHMPEnabled()
	return r.reconcileServiceMonitor(instance, expectedServiceMonitor, shouldDelete)
}

func (r *IBMLicensingReconciler) reconcileAlertingServiceMonitor(instance *operatorv1alpha1.IBMLicensing) (reconcile.Result, error) {
	expectedServiceMonitor := service.GetAlertingServiceMonitor(instance)
	shouldDelete := !instance.Spec.IsAlertingEnabled()
	return r.reconcileServiceMonitor(instance, expectedServiceMonitor, shouldDelete)
}

func (r *IBMLicensingReconciler) reconcileServiceMonitor(instance *operatorv1alpha1.IBMLicensing,
	expectedServiceMonitor *monitoringv1.ServiceMonitor, shouldDelete bool) (reconcile.Result, error) {

	reqLogger := r.Log.WithValues("reconcileServiceMonitor", "Entry", "instance.GetName()", instance.GetName(),
		"expectedServiceMonitor.GetName()", expectedServiceMonitor.GetName())
	foundServiceMonitor := &monitoringv1.ServiceMonitor{}
	if shouldDelete {
		reconcileResult, err := r.reconcileNamespacedResourceWhichShouldNotExist(
			instance, expectedServiceMonitor, foundServiceMonitor)
		if err != nil || reconcileResult.Requeue {
			return reconcileResult, err
		}
		return reconcile.Result{}, nil
	}

	owner := service.GetPrometheusService(instance)
	result, err := res.UpdateOwner(&reqLogger, r.Client, owner)
	if err != nil || result.Requeue {
		return result, err
	}
	result, err = r.reconcileResourceNamespacedExistenceWithCustomController(instance, owner, expectedServiceMonitor, foundServiceMonitor)
	if err != nil || result.Requeue {
		return result, err
	}
	result, err = res.UpdateServiceMonitor(&reqLogger, r.Client, expectedServiceMonitor, foundServiceMonitor)

	return result, err
}

func (r *IBMLicensingReconciler) reconcileNetworkPolicy(instance *operatorv1alpha1.IBMLicensing) (reconcile.Result, error) {
	if instance.Spec.IsPrometheusServiceNeeded() {
		reqLogger := r.Log.WithValues("reconcileNetworkPolicy", "Entry", "instance.GetName()", instance.GetName())
		expected := service.GetNetworkPolicy(instance)
		owner := service.GetPrometheusService(instance)
		result, err := res.UpdateOwner(&reqLogger, r.Client, owner)
		if err != nil || result.Requeue {
			return result, err
		}
		found := &networkingv1.NetworkPolicy{}
		result, err = r.reconcileResourceNamespacedExistenceWithCustomController(instance, owner, expected, found)
		if err != nil || result.Requeue {
			return result, err
		}
		result, err = res.UpdateResource(&reqLogger, r.Client, expected, found)

		return result, err
	}
	return reconcile.Result{}, nil
}

func (r *IBMLicensingReconciler) reconcileDeployment(instance *operatorv1alpha1.IBMLicensing) (reconcile.Result, error) {
	reqLogger := r.Log.WithValues("reconcileDeployment", "Entry", "instance.GetName()", instance.GetName())
	expectedDeployment := service.GetLicensingDeployment(instance)

	foundDeployment := &appsv1.Deployment{}
	reconcileResult, err := r.reconcileResourceNamespacedExistence(instance, expectedDeployment, foundDeployment)
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

func (r *IBMLicensingReconciler) reconcileCertificateSecrets(instance *operatorv1alpha1.IBMLicensing) (reconcile.Result, error) {
	var namespacedName types.NamespacedName
	var hostname []string
	var rolloutPods bool

	if res.IsRouteAPI && instance.Spec.IsRouteEnabled() {
		// for backward compatibility, we treat the "ocp" HTTPSCertsSource same as "self-signed"
		if instance.Spec.HTTPSCertsSource == "custom" {
			r.Log.Info("Skipping external certificate reconciliation - custom certificate set")
			return reconcile.Result{}, nil
		}

		r.Log.Info("Reconciling external certificate")

		routeNamespacedName := types.NamespacedName{Namespace: instance.Spec.InstanceNamespace, Name: service.GetResourceName(instance)}
		route := &routev1.Route{}
		if err := r.Client.Get(context.TODO(), routeNamespacedName, route); err != nil {
			r.Log.Error(err, "Cannot get route")
			return reconcile.Result{Requeue: true}, err
		}

		namespacedName = types.NamespacedName{Namespace: instance.Spec.InstanceNamespace, Name: service.LicenseServiceExternalCertName}
		hostname = []string{route.Spec.Host}
		rolloutPods = false
	} else {
		// skip certificate creation only for OCP environment if route is disabled
		if res.IsServiceCAAPI {
			r.Log.Info("Skipping certificate creation for OCP - route is disabled via configuration")
			return reconcile.Result{}, nil
		}
	}

	// Reconcile internal certificate only on non-OCP environments
	if !res.IsServiceCAAPI {
		r.Log.Info("Reconciling internal certificate")

		namespacedName = types.NamespacedName{Namespace: instance.Spec.InstanceNamespace, Name: service.LicenseServiceInternalCertName}

		hostname = make([]string, 2)
		hostname[0] = fmt.Sprintf("%s.%s.svc", service.GetResourceName(instance), instance.Spec.InstanceNamespace)
		hostname[1] = fmt.Sprintf("%s.%s.svc.cluster.local", service.GetResourceName(instance), instance.Spec.InstanceNamespace)

		rolloutPods = true
	}

	return r.reconcileSelfSignedCertificate(instance, namespacedName, hostname, rolloutPods)
}

func (r *IBMLicensingReconciler) reconcileRouteWithCertificates(instance *operatorv1alpha1.IBMLicensing) (reconcile.Result, error) {
	if res.IsRouteAPI && instance.Spec.IsRouteEnabled() {
		r.Log.Info("Reconciling route with certificate")
		externalCertSecret := corev1.Secret{}
		var externalCertName string
		if instance.Spec.HTTPSCertsSource == "custom" {
			externalCertName = service.LicenseServiceCustomExternalCertName
		} else {
			externalCertName = service.LicenseServiceExternalCertName
		}

		externalNamespacedName := types.NamespacedName{Namespace: instance.Spec.InstanceNamespace, Name: externalCertName}
		if err := r.Client.Get(context.TODO(), externalNamespacedName, &externalCertSecret); err != nil {
			r.Log.Error(err, "Cannot retrieve external certificate from secret")
			return reconcile.Result{Requeue: true}, nil
		}

		internalCertSecret := corev1.Secret{}
		internalNamespacedName := types.NamespacedName{Namespace: instance.Spec.InstanceNamespace, Name: service.LicenseServiceInternalCertName}
		if err := r.Client.Get(context.TODO(), internalNamespacedName, &internalCertSecret); err != nil {
			r.Log.Error(err, "Cannot retrieve external certificate from secret")
			return reconcile.Result{Requeue: true}, nil
		}

		cert, caCert, key, err := res.ProcessCerfiticateSecret(externalCertSecret)
		if err != nil {
			r.Log.Error(err, "Invalid certificate format in secret, retrying")
			return reconcile.Result{Requeue: true, RequeueAfter: 30 * time.Second}, err
		}
		_, destinationCaCert, _, err := res.ProcessCerfiticateSecret(internalCertSecret)
		if err != nil {
			r.Log.Error(err, "Invalid certificate format in secret, retrying")
			return reconcile.Result{Requeue: true, RequeueAfter: 30 * time.Second}, err
		}

		defaultRouteTLS := &routev1.TLSConfig{
			Termination:                   routev1.TLSTerminationReencrypt,
			InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyNone,
			Certificate:                   cert,
			CACertificate:                 caCert,
			Key:                           key,
			DestinationCACertificate:      destinationCaCert,
		}
		return r.reconcileRouteWithTLS(instance, defaultRouteTLS)
	}
	return reconcile.Result{}, nil
}

func (r *IBMLicensingReconciler) reconcileRouteWithoutCertificates(instance *operatorv1alpha1.IBMLicensing) (reconcile.Result, error) {
	defaultRouteTLS := &routev1.TLSConfig{
		Termination:                   routev1.TLSTerminationReencrypt,
		InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyNone,
	}

	route := &routev1.Route{}
	expectedRoute := service.GetLicensingRoute(instance, defaultRouteTLS)

	if res.IsRouteAPI && instance.Spec.IsRouteEnabled() {
		routeNamespacedName := types.NamespacedName{Namespace: instance.Spec.InstanceNamespace, Name: service.GetResourceName(instance)}
		if err := r.Client.Get(context.TODO(), routeNamespacedName, route); err != nil {
			r.Log.Info("Route does not exist, reconciling route without certificates")

			defaultRouteTLS := &routev1.TLSConfig{
				Termination:                   routev1.TLSTerminationReencrypt,
				InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyNone,
			}
			return r.reconcileRouteWithTLS(instance, defaultRouteTLS)
		}
	} else {
		r.Log.Info("Route is disabled, deleting current route if exists")
		reconcileResult, err := r.reconcileNamespacedResourceWhichShouldNotExist(instance, expectedRoute, route)
		if err != nil || reconcileResult.Requeue {
			return reconcileResult, err
		}
	}
	return reconcile.Result{}, nil
}

func (r *IBMLicensingReconciler) reconcileRouteWithTLS(instance *operatorv1alpha1.IBMLicensing, defaultRouteTLS *routev1.TLSConfig) (reconcile.Result, error) {
	if res.IsRouteAPI && instance.Spec.IsRouteEnabled() {
		expectedRoute := service.GetLicensingRoute(instance, defaultRouteTLS)
		foundRoute := &routev1.Route{}
		reconcileResult, err := r.reconcileResourceNamespacedExistence(instance, expectedRoute, foundRoute)
		if err != nil || reconcileResult.Requeue {
			return reconcileResult, err
		}
		reqLogger := r.Log.WithValues("reconcileRoute", "Entry", "instance.GetName()", instance.GetName())

		if !res.CompareRoutes(reqLogger, expectedRoute, foundRoute) {
			//route tls cannot be updated, that is why we delete and create
			reconcileResult, err = res.DeleteResource(&reqLogger, r.Client, foundRoute)
			if err != nil {
				return reconcileResult, err
			}
			time.Sleep(time.Second * 10)
			foundRoute = &routev1.Route{}
			reconcileResult, err = r.reconcileResourceNamespacedExistence(instance, expectedRoute, foundRoute)
			if err != nil || reconcileResult.Requeue {
				return reconcileResult, err
			}
		}
	}
	return reconcile.Result{}, nil
}

func (r *IBMLicensingReconciler) reconcileIngress(instance *operatorv1alpha1.IBMLicensing) (reconcile.Result, error) {
	expectedIngress := service.GetLicensingIngress(instance)
	foundIngress := &networkingv1.Ingress{}

	if instance.Spec.IsIngressEnabled() {
		reconcileResult, err := r.reconcileResourceNamespacedExistence(instance, expectedIngress, foundIngress)
		if err != nil || reconcileResult.Requeue {
			return reconcileResult, err
		}
		possibleUpdateNeeded := true
		reqLogger := r.Log.WithValues("reconcileIngress", "Entry", "instance.GetName()", instance.GetName())
		if foundIngress.ObjectMeta.Name != expectedIngress.ObjectMeta.Name {
			reqLogger.Info("Names not equal", "old", foundIngress.ObjectMeta.Name, "new", expectedIngress.ObjectMeta.Name)
		} else if !reflect.DeepEqual(foundIngress.ObjectMeta.Labels, expectedIngress.ObjectMeta.Labels) {
			reqLogger.Info("Labels not equal",
				"old", fmt.Sprintf("%v", foundIngress.ObjectMeta.Labels),
				"new", fmt.Sprintf("%v", expectedIngress.ObjectMeta.Labels))
		} else if !reflect.DeepEqual(foundIngress.ObjectMeta.Annotations, expectedIngress.ObjectMeta.Annotations) {
			reqLogger.Info("Annotations not equal",
				"old", fmt.Sprintf("%v", foundIngress.ObjectMeta.Annotations),
				"new", fmt.Sprintf("%v", expectedIngress.ObjectMeta.Annotations))
		} else if !reflect.DeepEqual(foundIngress.Spec, expectedIngress.Spec) {
			reqLogger.Info("Specs not equal",
				"old", fmt.Sprintf("%v", foundIngress.Spec),
				"new", fmt.Sprintf("%v", expectedIngress.Spec))
		} else {
			possibleUpdateNeeded = false
		}
		if possibleUpdateNeeded {
			return res.UpdateResource(&reqLogger, r.Client, expectedIngress, foundIngress)
		}
	} else {
		r.Log.Info("Ingress is disabled, deleting current ingress if exists")
		reconcileResult, err := r.reconcileNamespacedResourceWhichShouldNotExist(instance, expectedIngress, foundIngress)
		if err != nil || reconcileResult.Requeue {
			return reconcileResult, err
		}
	}
	return reconcile.Result{}, nil
}

func (r *IBMLicensingReconciler) reconcileMeterDefinition(instance *operatorv1alpha1.IBMLicensing) (reconcile.Result, error) {
	if !instance.Spec.IsRHMPEnabled() {
		return reconcile.Result{}, nil
	}
	reqLogger := r.Log.WithValues("reconcileMeterDefinition", "Entry", "instance.GetName()", instance.GetName())
	expectedMeterDefinitionList := service.GetMeterDefinitionList(instance)
	found := &rhmp.MeterDefinition{}
	owner := service.GetPrometheusService(instance)
	result, err := res.UpdateOwner(&r.Log, r.Client, owner)
	if err != nil || result.Requeue {
		return result, err
	}
	for _, expected := range expectedMeterDefinitionList {
		result, err := r.reconcileResourceNamespacedExistenceWithCustomController(instance, owner, expected, found)
		if err != nil || result.Requeue {
			return result, err
		}
		possibleUpdateNeeded := true
		foundMeters := found.Spec.Meters
		var foundMeter *rhmp.MeterWorkload
		if len(foundMeters) == 1 {
			foundMeter = &foundMeters[0]
		}
		if foundMeter != nil {
			expectedMeter := expected.Spec.Meters[0]
			// check basic parameters
			if found.ObjectMeta.Name != expected.ObjectMeta.Name {
				reqLogger.Info("Names not equal", "old", found.ObjectMeta.Name, "new", expected.ObjectMeta.Name)
			} else if found.Spec.Kind != expected.Spec.Kind {
				reqLogger.Info("Found wrong Kind")
			} else if foundMeter.Query != expectedMeter.Query {
				reqLogger.Info("Found MeterDefinition with wrong Query",
					"old", fmt.Sprintf("%v", foundMeter.Query),
					"new", fmt.Sprintf("%v", expectedMeter.Query))
			} else if foundMeter.Aggregation != expectedMeter.Aggregation {
				reqLogger.Info("Found MeterDefinition with wrong Aggregation",
					"old", fmt.Sprintf("%v", foundMeter.Aggregation),
					"new", fmt.Sprintf("%v", expectedMeter.Aggregation))
			} else if foundMeter.Name != expectedMeter.Name {
				reqLogger.Info("Found MeterDefinition with wrong Name",
					"old", fmt.Sprintf("%v", foundMeter.Name),
					"new", fmt.Sprintf("%v", expectedMeter.Name))
			} else if foundMeter.ValueLabelOverride != expectedMeter.ValueLabelOverride {
				reqLogger.Info("Found MeterDefinition with wrong ValueLabelOverride",
					"old", fmt.Sprintf("%v", foundMeter.ValueLabelOverride),
					"new", fmt.Sprintf("%v", expectedMeter.ValueLabelOverride))
			} else if foundMeter.Metric != expectedMeter.Metric {
				reqLogger.Info("Found MeterDefinition with wrong Metric",
					"old", fmt.Sprintf("%v", foundMeter.Metric),
					"new", fmt.Sprintf("%v", expectedMeter.Metric))
			} else if foundMeter.WorkloadType != expectedMeter.WorkloadType {
				reqLogger.Info("Found MeterDefinition with wrong WorkloadType",
					"old", fmt.Sprintf("%v", foundMeter.WorkloadType),
					"new", fmt.Sprintf("%v", expectedMeter.WorkloadType))
			} else if foundMeter.DateLabelOverride != expectedMeter.DateLabelOverride {
				reqLogger.Info("Found MeterDefinition with wrong DateLabelOverride",
					"old", fmt.Sprintf("%v", foundMeter.DateLabelOverride),
					"new", fmt.Sprintf("%v", expectedMeter.DateLabelOverride))
			} else {
				possibleUpdateNeeded = false
			}
			if !possibleUpdateNeeded {
				if !res.ListsEqualsLikeSets(expectedMeter.GroupBy, foundMeter.GroupBy) {
					reqLogger.Info("Found meter groupBy has wrong params",
						"old", fmt.Sprintf("%v", foundMeter.GroupBy),
						"new", fmt.Sprintf("%v", expectedMeter.GroupBy))
					possibleUpdateNeeded = true
				}
			}
		} else {
			reqLogger.Info("Found MeterDefinition without Meter")
		}
		if possibleUpdateNeeded {
			return res.UpdateResource(&reqLogger, r.Client, expected, found)
		}
	}
	return reconcile.Result{}, nil
}

func (r *IBMLicensingReconciler) reconcileResourceNamespacedExistence(
	instance *operatorv1alpha1.IBMLicensing, expectedRes res.ResourceObject, foundRes client.Object) (reconcile.Result, error) {

	namespacedName := types.NamespacedName{Name: expectedRes.GetName(), Namespace: expectedRes.GetNamespace()}
	return r.reconcileResourceExistence(instance, instance, expectedRes, foundRes, namespacedName)
}

func (r *IBMLicensingReconciler) reconcileResourceNamespacedExistenceWithCustomController(
	instance *operatorv1alpha1.IBMLicensing, controller, expectedRes res.ResourceObject, foundRes client.Object) (reconcile.Result, error) {

	namespacedName := types.NamespacedName{Name: expectedRes.GetName(), Namespace: expectedRes.GetNamespace()}
	return r.reconcileResourceExistence(instance, controller, expectedRes, foundRes, namespacedName)
}

func (r *IBMLicensingReconciler) reconcileResourceExistence(
	instance *operatorv1alpha1.IBMLicensing,
	controller metav1.Object,
	expectedRes res.ResourceObject,
	foundRes client.Object,
	namespacedName types.NamespacedName) (reconcile.Result, error) {

	resType := reflect.TypeOf(expectedRes)
	reqLogger := r.Log.WithValues(resType.String(), "Entry", "instance.GetName()", instance.GetName(), "expectedRes.getName()", expectedRes.GetName())

	// expectedRes already set before and passed via parameter
	err := controllerutil.SetControllerReference(controller, expectedRes, r.Scheme)
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
	reqLogger.Info(resType.String() + " exists!")
	return reconcile.Result{}, nil
}

func (r *IBMLicensingReconciler) reconcileNamespacedResourceWhichShouldNotExist(
	instance *operatorv1alpha1.IBMLicensing, expectedRes res.ResourceObject, foundRes client.Object) (reconcile.Result, error) {

	namespacedName := types.NamespacedName{Name: expectedRes.GetName(), Namespace: expectedRes.GetNamespace()}
	return r.reconcileResourceWhichShouldNotExist(instance, expectedRes, foundRes, namespacedName)
}

func (r *IBMLicensingReconciler) reconcileResourceWhichShouldNotExist(
	instance *operatorv1alpha1.IBMLicensing,
	expectedRes res.ResourceObject,
	foundRes client.Object,
	namespacedName types.NamespacedName) (reconcile.Result, error) {

	resType := reflect.TypeOf(expectedRes)
	reqLogger := r.Log.WithValues(resType.String(), "Entry", "instance.GetName()", instance.GetName())

	err := r.Client.Get(context.TODO(), namespacedName, foundRes)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		} else if metaErrors.IsNoMatchError(err) {
			return reconcile.Result{}, nil
		}
		reqLogger.Error(err, "Failed to get "+resType.String(), "Name", expectedRes.GetName(),
			"Namespace", expectedRes.GetNamespace())
		return reconcile.Result{}, nil
	}
	return res.DeleteResource(&reqLogger, r.Client, expectedRes)
}

func (r *IBMLicensingReconciler) getSelfSignedCertWithOwnerReference(
	instance *operatorv1alpha1.IBMLicensing,
	namespacedName types.NamespacedName,
	dns []string) (*corev1.Secret, error) {

	secret, err := res.GenerateSelfSignedCertSecret(namespacedName, dns)
	if err != nil {
		r.Log.Error(err, "Error when generating self signed certificate")
	}
	err = controllerutil.SetControllerReference(instance, secret, r.Scheme)
	if err != nil {
		r.Log.Error(err, "Failed to set owner reference in secret")
		return nil, err
	}
	return secret, nil
}

func (r *IBMLicensingReconciler) controllerStatus(instance *operatorv1alpha1.IBMLicensing) {
	if res.IsRouteAPI {
		r.Log.Info("Route feature is enabled")
	} else {
		r.Log.Info("Route feature is disabled")
	}
	if res.IsServiceCAAPI {
		r.Log.Info("ServiceCA feature is enabled")
	} else {
		r.Log.Info("ServiceCA feature is disabled")
	}
	if instance.Spec.IsRHMPEnabled() {
		r.Log.Info("RHMP is enabled")
	} else {
		r.Log.Info("RHMP is disabled")
	}
	if instance.Spec.IsAlertingEnabled() {
		r.Log.Info("Alerting is enabled")
	} else {
		r.Log.Info("Alerting is disabled")
	}
	if instance.Spec.UsageEnabled {
		r.Log.Info("Usage container is enabled")
	}
	if instance.Spec.IsNamespaceScopeEnabled() {
		r.Log.Info("Namespace scope restriction is enabled")
	}

}

func (r *IBMLicensingReconciler) reconcileSelfSignedCertificate(instance *operatorv1alpha1.IBMLicensing, secretNsName types.NamespacedName, hostname []string, rolloutPods bool) (reconcile.Result, error) {
	certSecret := &corev1.Secret{}

	if err := r.Client.Get(context.TODO(), secretNsName, certSecret); err != nil {
		r.Log.WithValues("cert name", secretNsName).Info("certificate secret not existing. Generating self signed certificate")

		secret, err := r.getSelfSignedCertWithOwnerReference(instance, secretNsName, hostname)
		if err != nil {
			r.Log.Error(err, "Error generating self signed certificate")
			return reconcile.Result{Requeue: true}, err
		}

		if err := r.Client.Create(context.TODO(), secret); err != nil {
			r.Log.Error(err, "Error creating self signed certificate")
			return reconcile.Result{Requeue: true}, err
		}
		if rolloutPods {
			deploymentNsName := types.NamespacedName{
				Name:      service.GetResourceName(instance),
				Namespace: instance.Spec.InstanceNamespace,
			}

			if err := r.rolloutRestartDeployment(deploymentNsName); err != nil {
				r.Log.Info("Failed to roll update deployment")
				return reconcile.Result{Requeue: true}, err
			}
		}

		return reconcile.Result{}, nil
	}
	// checking certificate
	cert, err := res.ParseCertificate(certSecret.Data["tls.crt"])
	reqLogger := r.Log.WithValues("reconcileCertificate", "Entry", "instance.GetName()", instance.GetName())

	regenerateCertificate := false

	// if improper x509 certificate
	if err != nil {
		r.Log.Error(err, "Improper x509 certificate in secret")
		regenerateCertificate = true
	}
	// if certificate is expired
	if cert.NotAfter.Before(time.Now().AddDate(0, 0, 90)) {
		r.Log.Info("Self signed certificate is expiring in less than 90 days.")
		regenerateCertificate = true
	}
	// if certificate is not issued to the proper host
	if err := cert.VerifyHostname(hostname[0]); err != nil {
		r.Log.Info("Certificate not issued to a proper hostname.")
		regenerateCertificate = true
	}

	if regenerateCertificate {
		r.Log.Info("Regenerating certificate")
		secret, err := r.getSelfSignedCertWithOwnerReference(instance, secretNsName, hostname)
		if err != nil {
			r.Log.Error(err, "Error creating self signed certificate")
			return reconcile.Result{Requeue: true}, err

		}
		result, err2 := res.UpdateResource(&reqLogger, r.Client, secret, certSecret)
		if err2 != nil {
			return result, err
		}

		if rolloutPods {
			deploymentNsName := types.NamespacedName{
				Name:      service.GetResourceName(instance),
				Namespace: instance.Spec.InstanceNamespace,
			}

			if err := r.rolloutRestartDeployment(deploymentNsName); err != nil {
				r.Log.Info("Failed to roll update deployment")
				return reconcile.Result{Requeue: true}, err
			}
		}

		return result, nil
	}
	r.Log.Info("*v1.Certificate exists!")
	return reconcile.Result{}, nil
}

func (r *IBMLicensingReconciler) rolloutRestartDeployment(deploymentNsName types.NamespacedName) error {
	r.Log.Info("Performing rolling restart of deployment")
	data := fmt.Sprintf(`{"spec":{"template":{"metadata":{"annotations":{"kubectl.kubernetes.io/restartedAt":"%s"}}}}}`, time.Now().String())
	patch := []byte(data)

	r.Log.Info(data)

	return r.Client.Patch(context.TODO(), &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: deploymentNsName.Namespace,
			Name:      deploymentNsName.Name,
		},
	}, client.RawPatch(types.MergePatchType, patch))
}
