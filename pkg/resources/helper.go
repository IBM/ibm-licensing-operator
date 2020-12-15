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

package resources

import (
	"context"
	"crypto/rand"
	"math/big"
	"reflect"
	"time"

	"github.com/operator-framework/operator-sdk/pkg/k8sutil"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/go-logr/logr"
	"github.com/ibm/ibm-licensing-operator/pkg/apis/operator/v1alpha1"
	servicecav1 "github.com/openshift/api/operator/v1"
	routev1 "github.com/openshift/api/route/v1"
	marketplacev1alpha1 "github.com/redhat-marketplace/redhat-marketplace-operator/pkg/apis/marketplace/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	c "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// cannot set to const due to k8s struct needing pointers to primitive types

var TrueVar = true
var FalseVar = false

var DefaultSecretMode int32 = 420
var Seconds60 int64 = 60

var IsRouteAPI = true
var IsServiceCAAPI = true
var RHMPEnabled = true
var IsRHMP = false

// Important product values needed for annotations
const LicensingProductName = "IBM Cloud Platform Common Services"
const LicensingProductID = "068a62892a1e4db39641342e592daa25"
const LicensingProductMetric = "FREE"

const randStringCharset string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const ocpCertSecretNameTag = "service.beta.openshift.io/serving-cert-secret-name" // #nosec
const OcpCheckString = "ocp-check-secret"

var randStringCharsetLength = big.NewInt(int64(len(randStringCharset)))

var annotationsForServicesToCheck = [...]string{ocpCertSecretNameTag}

type ResourceObject interface {
	metav1.Object
	runtime.Object
}

func RandString(length int) (string, error) {
	reader := rand.Reader
	outputStringByte := make([]byte, length)
	for i := 0; i < length; i++ {
		charIndex, err := rand.Int(reader, randStringCharsetLength)
		if err != nil {
			return "", err
		}
		outputStringByte[i] = randStringCharset[charIndex.Int64()]
	}
	return string(outputStringByte), nil
}

func Contains(s []corev1.LocalObjectReference, e corev1.LocalObjectReference) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func AnnotationsForPod() map[string]string {
	return map[string]string{"productName": LicensingProductName,
		"productID": LicensingProductID, "productMetric": LicensingProductMetric}
}

func WatchForResources(log logr.Logger, o runtime.Object, c controller.Controller, watchTypes []ResourceObject) error {
	for _, restype := range watchTypes {
		log.Info("Watching", "restype", reflect.TypeOf(restype).String())
		err := c.Watch(&source.Kind{Type: restype}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    o,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func GetSecretToken(name string, namespace string, secretKey string, metaLabels map[string]string) (*corev1.Secret, error) {
	randString, err := RandString(24)
	if err != nil {
		return nil, err
	}
	expectedSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    metaLabels,
		},
		Type:       corev1.SecretTypeOpaque,
		StringData: map[string]string{secretKey: randString},
	}
	return expectedSecret, nil
}

func AnnotateForService(httpCertSource v1alpha1.HTTPSCertsSource, isHTTPS bool, certName string) map[string]string {
	if IsServiceCAAPI && isHTTPS && httpCertSource == v1alpha1.OcpCertsSource {
		return map[string]string{ocpCertSecretNameTag: certName}
	}
	return map[string]string{}
}

func UpdateResource(reqLogger *logr.Logger, client c.Client,
	expectedResource ResourceObject, foundResource ResourceObject) (reconcile.Result, error) {
	resTypeString := reflect.TypeOf(expectedResource).String()
	(*reqLogger).Info("Updating " + resTypeString)
	err := client.Update(context.TODO(), expectedResource)
	if err != nil {
		// only need to delete resource as new will be recreated on next reconciliation
		(*reqLogger).Error(err, "Failed to update "+resTypeString+", deleting...", "Namespace", foundResource.GetNamespace(), "Name", foundResource.GetName())
		return DeleteResource(reqLogger, client, foundResource)
	}
	(*reqLogger).Info("Updated "+resTypeString+" successfully", "Namespace", expectedResource.GetNamespace(), "Name", expectedResource.GetName())
	// Resource updated - return and do not requeue as it might not consider extra values
	return reconcile.Result{}, nil
}

func UpdateServiceIfNeeded(reqLogger *logr.Logger, client c.Client, expectedService *corev1.Service, foundService *corev1.Service) (reconcile.Result, error) {
	for _, annotation := range annotationsForServicesToCheck {
		if foundService.Annotations[annotation] != expectedService.Annotations[annotation] {
			return UpdateResource(reqLogger, client, expectedService, foundService)
		}
	}
	return reconcile.Result{}, nil
}

func UpdateServiceMonitor(reqLogger *logr.Logger, client c.Client, expected, found *monitoringv1.ServiceMonitor) (reconcile.Result, error) {
	for _, annotation := range annotationsForServicesToCheck {
		if found.Annotations[annotation] != expected.Annotations[annotation] {
			return UpdateResource(reqLogger, client, found, expected)
		}
	}
	return reconcile.Result{}, nil
}

func DeleteResource(reqLogger *logr.Logger, client c.Client, foundResource ResourceObject) (reconcile.Result, error) {
	resTypeString := reflect.TypeOf(foundResource).String()
	err := client.Delete(context.TODO(), foundResource)
	if err != nil {
		(*reqLogger).Error(err, "Failed to delete "+resTypeString+" during recreation", "Namespace", foundResource.GetNamespace(), "Name", foundResource.GetName())
		return reconcile.Result{}, err
	}
	// Resource deleted successfully - return and requeue to create new one
	(*reqLogger).Info("Deleted "+resTypeString+" successfully", "Namespace", foundResource.GetNamespace(), "Name", foundResource.GetName())
	return reconcile.Result{Requeue: true, RequeueAfter: time.Second * 30}, nil
}

func GetOCPSecretCheckScript() string {
	script := `while true; do
  echo "$(date): Checking for ocp secret"
  ls /opt/licensing/certs/* && break
  echo "$(date): Required ocp secret not found ... try again in 30s"
  sleep 30
done
echo "$(date): All required secrets exist"
`
	return script
}

func UpdateCacheClusterExtensions(client c.Reader) error {
	namespace, err := k8sutil.GetWatchNamespace()
	if err != nil {
		return err
	}

	listOpts := []c.ListOption{
		c.InNamespace(namespace),
	}

	routeTestInstance := &routev1.Route{}
	err = client.List(context.TODO(), routeTestInstance, listOpts...)
	if err == nil {
		IsRouteAPI = true
	} else {
		IsRouteAPI = false
	}

	IsRHMP = checkRHMPPrereqs(client)

	serviceCAInstance := &servicecav1.ServiceCA{}
	err = client.List(context.TODO(), serviceCAInstance, listOpts...)
	if err == nil {
		IsServiceCAAPI = true
	} else {
		IsServiceCAAPI = false
	}
	return nil
}

func IsRHMPEnabledAndInstalled(rhmpEnabled bool) bool {
	return rhmpEnabled && IsRHMP
}

func checkRHMPPrereqs(client c.Reader) bool {
	mcList := &marketplacev1alpha1.MarketplaceConfigList{}
	err := client.List(context.TODO(), mcList, []c.ListOption{}...)
	return err == nil && len(mcList.Items) > 0
}
