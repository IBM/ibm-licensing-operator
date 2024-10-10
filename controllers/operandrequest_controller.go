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
	"regexp"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	res "github.com/IBM/ibm-licensing-operator/controllers/resources"
	svcres "github.com/IBM/ibm-licensing-operator/controllers/resources/service"
	odlm "github.com/IBM/operand-deployment-lifecycle-manager/api/v1alpha1"
)

var operandBindInfoInfix, _ = regexp.Compile(`^(.*)(opbi|bindinfo)(.*)$`)

// OperandRequestReconciler reconciles a OperandRequest object
type OperandRequestReconciler struct {
	client.Client
	client.Reader
	Log               logr.Logger
	Scheme            *runtime.Scheme
	OperatorNamespace string
}

// SetupWithManager sets up the controller with the Manager.
func (r *OperandRequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := res.UpdateCacheClusterExtensions(mgr.GetAPIReader()); err != nil {
		r.Log.Error(err, "Error during checking K8s API")
	}

	watcher := ctrl.NewControllerManagedBy(mgr).
		For(&odlm.OperandRequest{}).
		WithEventFilter(ignoreDeletionPredicate())

	return watcher.Complete(r)
}

func ignoreDeletionPredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// Ignore updates to CR status in which case metadata.Generation does not change
			return e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration()
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			// Evaluates to false if the object has been confirmed deleted.
			return !e.DeleteStateUnknown
		},
	}
}

// Reconcile reads that state of the cluster for an OperandRequest object and copies shared Config Maps and Secrets
// to OperandRequest's namespace
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.

// +kubebuilder:rbac:groups=operator.ibm.com,resources=operandrequests;operandrequests/finalizers;operandrequests/status,verbs=get;list;patch;update;watch
// +kubebuilder:rbac:namespace=ibm-licensing,groups="",resources=configmaps;secrets,verbs=get;list;watch;create;update;patch;delete

/*
UpdateOperandRequestWithPhase sets a specific phase of the operand request's member field.

This is different to the overall request phase, as we only care about settings phases of related bindings.

This function may return an error for example if the Update API call doesn't work. Licensing does not have permissions
to recreate a resource (delete + create) and only patching/updating is supported.
*/
func (r *OperandRequestReconciler) UpdateOperandRequestWithPhase(
	reqLogger logr.Logger,
	operandRequest *odlm.OperandRequest,
	phase odlm.ServicePhase,
) {
	reqLogger.Info(fmt.Sprintf("Updating %s operand request with phase %s", operandRequest.Name, phase))

	shouldAddMember := true
	for index, member := range operandRequest.Status.Members {
		if member.Name == res.OperatorName {
			operandRequest.Status.Members[index].Phase = odlm.MemberPhase{
				OperatorPhase: member.Phase.OperatorPhase,
				OperandPhase:  phase,
			}
			shouldAddMember = false
			break
		}
	}

	// No member status for licensing was found, so add it
	if shouldAddMember {
		operandRequest.Status.Members = append(operandRequest.Status.Members, odlm.MemberStatus{
			Name: res.OperatorName,
			Phase: odlm.MemberPhase{
				OperandPhase: phase,
			},
		})
	}

	// Updating status requires the Status() call, otherwise the resource's status will not be updated
	if err := r.Client.Status().Update(context.Background(), operandRequest); err != nil {
		reqLogger.Error(
			err,
			"Couldn't update operand request status (this may cause some dependent components to fail)",
			"Name", operandRequest.Name,
			"Phase", phase,
		)
	}
}

func (r *OperandRequestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqLogger := r.Log.WithValues("operandrequest", req.NamespacedName)

	if err := res.UpdateCacheClusterExtensions(r.Reader); err != nil {
		reqLogger.Error(err, "Error during checking K8s API")
	}

	// Fetch the OperandRequest instance
	operandRequest := odlm.OperandRequest{}
	if err := r.Client.Get(ctx, req.NamespacedName, &operandRequest); err != nil {
		// Error reading the object - requeue the request.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if isForLicensing := res.HasOperandRequestBindingForLicensing(operandRequest); !isForLicensing {
		return ctrl.Result{}, nil
	}

	reqLogger.Info("Reconciling OperandRequest")

	for _, request := range operandRequest.Spec.Requests {
		var infoConfigMapName, tokenSecretName, tokenSecretName2, uploadConfigName, uploadTokenName string
		var requeueTokenSec, requeueToken2Sec, requeueUploadSec, requeueInfoCm, requeueUploadCm bool
		var err error
		for _, operand := range request.Operands {
			if operand.Name == res.OperatorName {
				r.UpdateOperandRequestWithPhase(reqLogger, &operandRequest, odlm.ServiceCreating)

				// In case something failed but requeue was not requested
				operandRequestFailedCopy := false

				for key, binding := range operand.Bindings {
					if key == "public-api-data" {
						infoConfigMapName = binding.Configmap
						tokenSecretName = binding.Secret
					}

					if key == "public-api-token" {
						tokenSecretName2 = binding.Secret
					}

					if key == "public-api-upload" {
						uploadConfigName = binding.Configmap
						uploadTokenName = binding.Secret
					}
				}

				requeueTokenSec, err = r.copySecret(ctx, req, svcres.LicensingToken, tokenSecretName, r.OperatorNamespace, operandRequest.Namespace, &operandRequest)
				if err != nil {
					reqLogger.Error(err, "Cannot copy Secret", "name", svcres.LicensingToken, "namespace", operandRequest.Namespace)
					operandRequestFailedCopy = true
					r.UpdateOperandRequestWithPhase(reqLogger, &operandRequest, odlm.ServiceFailed)
				}
				if requeueTokenSec {
					return reconcile.Result{Requeue: true}, err
				}

				requeueToken2Sec, err = r.copySecret(ctx, req, svcres.LicensingToken, tokenSecretName2, r.OperatorNamespace, operandRequest.Namespace, &operandRequest)
				if err != nil {
					reqLogger.Error(err, "Cannot copy Secret", "name", svcres.LicensingToken, "namespace", operandRequest.Namespace)
					operandRequestFailedCopy = true
					r.UpdateOperandRequestWithPhase(reqLogger, &operandRequest, odlm.ServiceFailed)
				}
				if requeueToken2Sec {
					return reconcile.Result{Requeue: true}, err
				}

				requeueUploadSec, err = r.copySecret(ctx, req, svcres.LicensingUploadToken, uploadTokenName, r.OperatorNamespace, operandRequest.Namespace, &operandRequest)
				if err != nil {
					reqLogger.Error(err, "Cannot copy Secret", "name", svcres.LicensingUploadToken, "namespace", operandRequest.Namespace)
					operandRequestFailedCopy = true
					r.UpdateOperandRequestWithPhase(reqLogger, &operandRequest, odlm.ServiceFailed)
				}
				if requeueUploadSec {
					return reconcile.Result{Requeue: true}, err
				}

				requeueInfoCm, err = r.copyConfigMap(ctx, req, svcres.LicensingInfo, infoConfigMapName, r.OperatorNamespace, operandRequest.Namespace, &operandRequest)
				if err != nil {
					reqLogger.Error(err, "Cannot copy ConfigMap", svcres.LicensingInfo, "namespace", operandRequest.Namespace)
					operandRequestFailedCopy = true
					r.UpdateOperandRequestWithPhase(reqLogger, &operandRequest, odlm.ServiceFailed)
				}
				if requeueInfoCm {
					return reconcile.Result{Requeue: true}, err
				}

				requeueUploadCm, err = r.copyConfigMap(ctx, req, svcres.LicensingUploadConfig, uploadConfigName, r.OperatorNamespace, operandRequest.Namespace, &operandRequest)
				if err != nil {
					reqLogger.Error(err, "Cannot copy ConfigMap", "name", svcres.LicensingUploadConfig, "namespace", operandRequest.Namespace)
					operandRequestFailedCopy = true
					r.UpdateOperandRequestWithPhase(reqLogger, &operandRequest, odlm.ServiceFailed)
				}
				if requeueUploadCm {
					return reconcile.Result{Requeue: true}, err
				}

				// Set the status as Running - the operand request has provided all requested objects
				// Any status could be chosen here, but to avoid confusion with updates/creation, Running is used
				// IBM License Service Scanner is also expecting this status here when reconciling connection to License Service
				if !operandRequestFailedCopy {
					r.UpdateOperandRequestWithPhase(reqLogger, &operandRequest, odlm.ServiceRunning)
				}
			}
		}
	}

	reqLogger.Info("reconcile all done")
	return ctrl.Result{}, nil
}

// Copy secret `sourceName` from source namespace `sourceNs` to target namespace `targetNs`
func (r *OperandRequestReconciler) copySecret(ctx context.Context, req reconcile.Request, sourceName, targetName, sourceNs, targetNs string,
	requestInstance *odlm.OperandRequest) (requeue bool, err error) {
	reqLogger := r.Log.WithValues("operandrequest", req.NamespacedName)

	if sourceName == "" || sourceNs == "" || targetNs == "" {
		return false, nil
	}

	if sourceName == targetName && sourceNs == targetNs {
		return false, nil
	}

	if targetName == "" {
		targetName = res.LsBindInfoName + "-" + sourceName
	}

	secret := corev1.Secret{}
	if err := r.Client.Get(ctx, types.NamespacedName{Name: sourceName, Namespace: sourceNs}, &secret); err != nil {
		if apierrors.IsNotFound(err) {
			reqLogger.Info("Secret not found", "name", sourceName, "namespace", sourceNs)
			return true, nil
		}
		reqLogger.Error(err, "failed to get Secret", "name", sourceName, "namespace", sourceNs)
		return false, err
	}
	// Create the Secret to the OperandRequest namespace
	secretLabel := make(map[string]string)
	// Copy from the original labels to the target labels
	for k, v := range secret.Labels {
		if operandBindInfoInfix.MatchString(k) {
			continue
		}
		secretLabel[k] = v
	}

	secretCopy := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        targetName,
			Namespace:   targetNs,
			Labels:      secretLabel,
			Annotations: secret.Annotations,
		},
		Type:       secret.Type,
		Data:       secret.Data,
		StringData: secret.StringData,
	}
	// Set the OperandRequest as the controller of the Secret
	if err := controllerutil.SetControllerReference(requestInstance, &secretCopy, r.Scheme); err != nil {
		reqLogger.Error(err, "failed to set OperandRequest as the owner of Secret", "OperandRequestName", requestInstance.Name, "SecretName", secretCopy.Name, "namespace", targetNs)
		return false, err
	}

	if err := r.Create(ctx, &secretCopy); err != nil {
		if apierrors.IsAlreadyExists(err) {
			// If already exist, update the Secret
			existingSecret := corev1.Secret{}
			if err := r.Client.Get(ctx, types.NamespacedName{Namespace: targetNs, Name: targetName}, &existingSecret); err != nil {
				reqLogger.Error(err, "failed to get Secret", "name", targetName, "namespace", targetNs)
				return false, err
			}
			// Update existing Secret only if it has same name and set of labels
			if needUpdate := !res.CompareSecretsData(&secretCopy, &existingSecret); needUpdate {
				if err := r.Update(ctx, &secretCopy); err != nil {
					reqLogger.Error(err, "failed to update Secret", "name", targetName, "namespace", targetNs)
					return false, err
				}
			}
		} else {
			reqLogger.Error(err, "failed to create Secret", "name", targetName, "namespace", targetNs)
			return false, err
		}
	}

	return false, nil
}

// Copy configmap `sourceName` from namespace `sourceNs` to namespace `targetNs`
// and rename it to `targetName`
func (r *OperandRequestReconciler) copyConfigMap(ctx context.Context, req reconcile.Request, sourceName, targetName, sourceNs, targetNs string,
	requestInstance *odlm.OperandRequest) (requeue bool, err error) {
	reqLogger := r.Log.WithValues("operandrequest", req.NamespacedName)

	if sourceName == "" || sourceNs == "" || targetNs == "" {
		return false, nil
	}

	if sourceName == targetName && sourceNs == targetNs {
		return false, nil
	}

	if targetName == "" {
		targetName = res.LsBindInfoName + "-" + sourceName
	}

	cm := corev1.ConfigMap{}
	if err := r.Client.Get(ctx, types.NamespacedName{Name: sourceName, Namespace: sourceNs}, &cm); err != nil {
		if apierrors.IsNotFound(err) {
			reqLogger.Info("Configmap not found", "name", sourceName, "namespace", sourceNs)
			return true, nil
		}
		reqLogger.Error(err, "failed to get ConfigMap", "name", sourceName, "namespace", sourceNs)
		return false, err
	}
	// Create the ConfigMap to the OperandRequest namespace
	cmLabel := make(map[string]string)
	// Copy from the original labels to the target labels
	for k, v := range cm.Labels {
		if operandBindInfoInfix.MatchString(k) {
			continue
		}
		cmLabel[k] = v
	}

	cmCopy := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:        targetName,
			Namespace:   targetNs,
			Labels:      cmLabel,
			Annotations: cm.Annotations,
		},
		Data:       cm.Data,
		BinaryData: cm.BinaryData,
	}
	// Set the OperandRequest as the controller of the configmap
	if err := controllerutil.SetControllerReference(requestInstance, &cmCopy, r.Scheme); err != nil {
		reqLogger.Error(err, "failed to set OperandRequest as an owner of a ConfigMap", "OperandRequestName", requestInstance.Name, "ConfigMapName", cmCopy.Name, "namespace", targetNs)
		return false, err
	}

	// Create the ConfigMap in the OperandRequest namespace
	if err := r.Create(ctx, &cmCopy); err != nil {
		if apierrors.IsAlreadyExists(err) {
			// If already exist, update the ConfigMap
			existingCm := corev1.ConfigMap{}
			if err := r.Client.Get(ctx, types.NamespacedName{Namespace: targetNs, Name: targetName}, &existingCm); err != nil {
				reqLogger.Error(err, "failed to get ConfigMap", "name", targetName, "namespace", targetNs)
				return false, err
			}
			// Update existing ConfigMap only if it has same name and set of labels
			if needUpdate := !res.CompareConfigMapData(&existingCm, &cmCopy); needUpdate {
				if err := r.Update(ctx, &cmCopy); err != nil {
					reqLogger.Error(err, "failed to update ConfigMap", "name", targetName, "namespace", targetNs)
					return false, err
				}
			}
		} else {
			reqLogger.Error(err, "failed to create ConfigMap", "name", targetName, "namespace", targetNs)
			return false, err
		}
	}

	return false, nil
}
