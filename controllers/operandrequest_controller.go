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
	"regexp"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
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

// OperandRequestReconciler reconciles a OperandRequest object
type OperandRequestReconciler struct {
	client.Client
	client.Reader
	Log               logr.Logger
	Scheme            *runtime.Scheme
	OperatorNamespace string
}

var (
	operandBindInfoInfix, _ = regexp.Compile(`^(.*)opbi(.*)$`)
)

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

// Reconcile reads that state of the cluster for a OperandRequest object and copies shared Config Maps and Secrets
// to OperandRequest's namespace
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.

//+kubebuilder:rbac:groups=operator.ibm.com,resources=operandrequests;operandrequests/finalizers,verbs=get;list;watch;create;update;patch
//+kubebuilder:rbac:groups=operator.ibm.com,resources=operandrequests/status,verbs=get;list
//+kubebuilder:rbac:groups="",resources=configmaps;secrets,verbs=get;list;watch;create;update;patch;delete

func (r *OperandRequestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	reqLogger := r.Log.WithValues("operandrequest", req.NamespacedName)
	reqLogger.Info("Reconciling OperandRequest")

	if err := res.UpdateCacheClusterExtensions(r.Reader); err != nil {
		reqLogger.Error(err, "Error during checking K8s API")
	}

	// Fetch the OperandRequest instance
	operandRequest := odlm.OperandRequest{}
	if err := r.Client.Get(context.TODO(), req.NamespacedName, &operandRequest); err != nil {
		// Error reading the object - requeue the request.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var infoConfigMapName, tokenSecretName, uploadConfigName, uploadTokenName string
	var requeueTokenSec, requeueUploadSec, requeueInfoCm, requeueUploadCm bool
	var err error

	licensingOpreqHandled := false
	for _, request := range operandRequest.Spec.Requests {
		for _, operand := range request.Operands {
			if operand.Name == res.OperatorName {

				for key, binding := range operand.Bindings {
					if key == "public-api-data" {
						infoConfigMapName = binding.Configmap
						tokenSecretName = binding.Secret
					}

					if key == "public-api-token" {
						tokenSecretName = binding.Secret
					}

					if key == "public-api-upload" {
						uploadConfigName = binding.Configmap
						uploadTokenName = binding.Secret
					}
				}

				requeueTokenSec, err = r.copySecret(ctx, req, svcres.LicensingToken, tokenSecretName, r.OperatorNamespace, operandRequest.Namespace, &operandRequest)
				if err != nil {
					reqLogger.Error(err, "Cannot copy Secret %s to Namespace %s", svcres.LicensingToken, operandRequest.Namespace)
				}
				if requeueTokenSec {
					return reconcile.Result{Requeue: true}, err
				}

				requeueUploadSec, err = r.copySecret(ctx, req, svcres.LicensingUploadToken, uploadTokenName, r.OperatorNamespace, operandRequest.Namespace, &operandRequest)
				if err != nil {
					reqLogger.Error(err, "Cannot copy Secret %s to Namespace %s", svcres.LicensingUploadToken, operandRequest.Namespace)
				}
				if requeueUploadSec {
					return reconcile.Result{Requeue: true}, err
				}

				requeueInfoCm, err = r.copyConfigMap(ctx, req, svcres.LicensingInfo, infoConfigMapName, r.OperatorNamespace, operandRequest.Namespace, &operandRequest)
				if err != nil {
					reqLogger.Error(err, "Cannot copy ConfigMap %s to Namespace %s", svcres.LicensingInfo, operandRequest.Namespace)
				}
				if requeueInfoCm {
					return reconcile.Result{Requeue: true}, err
				}

				requeueUploadCm, err = r.copyConfigMap(ctx, req, svcres.LicensingUploadConfig, uploadConfigName, r.OperatorNamespace, operandRequest.Namespace, &operandRequest)
				if err != nil {
					reqLogger.Error(err, "Cannot copy ConfigMap %s to Namespace %s", svcres.LicensingUploadConfig, operandRequest.Namespace)
				}
				if requeueUploadCm {
					return reconcile.Result{Requeue: true}, err
				}

				licensingOpreqHandled = true
				break
			}
		}
		if licensingOpreqHandled {
			break
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
		targetName = requestInstance.Name + "-" + sourceName
	}

	secret := &corev1.Secret{}
	if err := r.Client.Get(ctx, types.NamespacedName{Name: sourceName, Namespace: sourceNs}, secret); err != nil {
		if apierrors.IsNotFound(err) {
			reqLogger.Info("Secret %s is not found from the namespace %s", sourceName, sourceNs) // TODO
			return true, nil
		}
		reqLogger.Error(err, "failed to get Secret %s/%s", sourceNs, sourceName)
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

	secretCopy := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      targetName,
			Namespace: targetNs,
			Labels:    secretLabel,
		},
		Type:       secret.Type,
		Data:       secret.Data,
		StringData: secret.StringData,
	}
	// Set the OperandRequest as the controller of the Secret
	if err := controllerutil.SetControllerReference(requestInstance, secretCopy, r.Scheme); err != nil {
		reqLogger.Error(err, "failed to set OperandRequest %s as the owner of Secret %s", requestInstance.Name, targetName)
		return false, err
	}

	if err := r.Create(ctx, secretCopy); err != nil {
		if apierrors.IsAlreadyExists(err) {
			// If already exist, update the Secret
			existingSecret := &corev1.Secret{}
			if err := r.Client.Get(ctx, types.NamespacedName{Namespace: targetNs, Name: targetName}, existingSecret); err != nil {
				reqLogger.Error(err, "failed to get secret %s/%s", targetNs, targetName)
				return false, err
			}
			if needUpdate := compareSecret(secretCopy, existingSecret); needUpdate {
				if err := r.Update(ctx, secretCopy); err != nil {
					reqLogger.Error(err, "failed to update secret %s/%s", targetNs, targetName)
					return false, err
				}
			}
		} else {
			reqLogger.Error(err, "failed to create secret %s/%s", targetNs, targetName)
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
		targetName = requestInstance.Name + "-" + sourceName
	}

	cm := &corev1.ConfigMap{}
	if err := r.Client.Get(ctx, types.NamespacedName{Name: sourceName, Namespace: sourceNs}, cm); err != nil {
		if apierrors.IsNotFound(err) {
			reqLogger.Info("Configmap %s/%s is not found", sourceNs, sourceName)
			return true, nil
		}
		reqLogger.Error(err, "failed to get ConfigMap %s/%s", sourceNs, sourceName)
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

	cmCopy := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      targetName,
			Namespace: targetNs,
			Labels:    cmLabel,
		},
		Data:       cm.Data,
		BinaryData: cm.BinaryData,
	}
	// Set the OperandRequest as the controller of the configmap
	if err := controllerutil.SetControllerReference(requestInstance, cmCopy, r.Scheme); err != nil {
		reqLogger.Error(err, "failed to set OperandRequest %s as the owner of ConfigMap %s", requestInstance.Name, sourceName)
		return false, err
	}

	// Create the ConfigMap in the OperandRequest namespace
	if err := r.Create(ctx, cmCopy); err != nil {
		if apierrors.IsAlreadyExists(err) {
			// If already exist, update the ConfigMap
			existingCm := &corev1.ConfigMap{}
			if err := r.Client.Get(ctx, types.NamespacedName{Namespace: targetNs, Name: targetName}, existingCm); err != nil {
				reqLogger.Error(err, "failed to get ConfigMap %s/%s", targetNs, targetName)
				return false, err
			}
			if needUpdate := compareConfigMap(cmCopy, existingCm); needUpdate {
				if err := r.Update(ctx, cmCopy); err != nil {
					reqLogger.Error(err, "failed to update ConfigMap %s/%s", targetNs, sourceName)
					return false, err
				}
			}
		} else {
			reqLogger.Error(err, "failed to create ConfigMap %s/%s", targetNs, sourceName)
			return false, err
		}
	}

	// Update the operand Configmap
	if err := r.Update(ctx, cm); err != nil {
		reqLogger.Error(err, "failed to update ConfigMap %s/%s", cm.Namespace, cm.Name)
		return false, err
	}

	return false, nil
}

func compareSecret(secret *corev1.Secret, existingSecret *corev1.Secret) (needUpdate bool) {
	return !equality.Semantic.DeepEqual(secret.GetLabels(), existingSecret.GetLabels()) ||
		!equality.Semantic.DeepEqual(secret.Type, existingSecret.Type) ||
		!equality.Semantic.DeepEqual(secret.Data, existingSecret.Data) ||
		!equality.Semantic.DeepEqual(secret.StringData, existingSecret.StringData)
}

func compareConfigMap(configMap *corev1.ConfigMap, existingConfigMap *corev1.ConfigMap) (needUpdate bool) {
	return !equality.Semantic.DeepEqual(configMap.GetLabels(), existingConfigMap.GetLabels()) ||
		!equality.Semantic.DeepEqual(configMap.Data, existingConfigMap.Data) ||
		!equality.Semantic.DeepEqual(configMap.BinaryData, existingConfigMap.BinaryData)
}
