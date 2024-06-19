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

package resources

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"regexp"
	"time"

	operatorv1alpha1 "github.com/IBM/ibm-licensing-operator/api/v1alpha1"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/go-logr/logr"
	servicecav1 "github.com/openshift/api/operator/v1"
	routev1 "github.com/openshift/api/route/v1"
	rhmp "github.com/redhat-marketplace/redhat-marketplace-operator/v2/apis/marketplace/v1beta1"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apieq "k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	c "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	odlm "github.com/IBM/operand-deployment-lifecycle-manager/api/v1alpha1"
)

// cannot set to const due to k8s struct needing pointers to primitive types
var (
	TrueVar  = true
	FalseVar = false

	DefaultSecretMode int32 = 420
	Seconds60         int64 = 60

	IsRouteAPI                 = true
	IsServiceCAAPI             = true
	IsAlertingEnabledByDefault = true
	RHMPEnabled                = false
	IsODLM                     = true

	PathType = networkingv1.PathTypeImplementationSpecific
)

const (
	// Important product values needed for annotations
	LicensingProductName   = "IBM Cloud Platform Common Services"
	LicensingProductID     = "068a62892a1e4db39641342e592daa25"
	LicensingProductMetric = "FREE"

	ocpCertSecretNameTag = "service.beta.openshift.io/serving-cert-secret-name" // #nosec

	OcpCheckString           = "ocp-check-secret"
	OcpPrometheusCheckString = "ocp-prometheus-check-secret"
	OperatorName             = "ibm-licensing-operator"
)

var annotationsForServicesToCheck = [...]string{ocpCertSecretNameTag}

type ResourceObject interface {
	metav1.Object
	runtime.Object
}

// we could use reflection to have this method for all types but for now strings would be enough
func ListsEqualsLikeSets(list1 []string, list2 []string) bool {
	if list1 == nil {
		return list2 == nil
	}
	if len(list1) != len(list2) {
		return false
	}
	for _, item1 := range list1 {
		found := false
		for _, item2 := range list2 {
			if item2 == item1 {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func AnnotationsForPod(instance *operatorv1alpha1.IBMLicensing) map[string]string {
	return mergeWithSpecAnnotations(instance, map[string]string{
		"productName":   LicensingProductName,
		"productID":     LicensingProductID,
		"productMetric": LicensingProductMetric,
	})
}

func AnnotateForService(instance *operatorv1alpha1.IBMLicensing, certName string) map[string]string {
	if IsServiceCAAPI && instance.Spec.HTTPSEnable {
		return mergeWithSpecAnnotations(instance, map[string]string{ocpCertSecretNameTag: certName})
	}
	return mergeWithSpecAnnotations(instance, map[string]string{})
}

// Attach labels existing on the found resource to the expected resource
func attachExistingLabels(foundResource ResourceObject, expectedResource ResourceObject) {
	resourceLabels := foundResource.GetLabels()
	expectedLabels := expectedResource.GetLabels()

	if expectedLabels == nil {
		expectedLabels = map[string]string{}
	}

	for key, value := range resourceLabels {
		_, ok := expectedLabels[key]
		if !ok {
			expectedLabels[key] = value
		}
	}
}

// Attach annotations existing on the found resource to the expected resource
func attachExistingAnnotations(foundResource ResourceObject, expectedResource ResourceObject) {
	resourceAnnotations := foundResource.GetAnnotations()
	expectedAnnotations := expectedResource.GetAnnotations()

	if expectedAnnotations == nil {
		expectedAnnotations = map[string]string{}
	}

	for key, value := range resourceAnnotations {
		_, ok := expectedAnnotations[key]
		if !ok {
			expectedAnnotations[key] = value
		}
	}
}

func UpdateResource(reqLogger *logr.Logger, client c.Client,
	expectedResource ResourceObject, foundResource ResourceObject) (reconcile.Result, error) {
	resTypeString := reflect.TypeOf(expectedResource).String()
	(*reqLogger).Info("Updating " + resTypeString)
	expectedResource.SetResourceVersion(foundResource.GetResourceVersion())

	// Ensure persistence of existing metadata
	attachExistingLabels(foundResource, expectedResource)
	attachExistingAnnotations(foundResource, expectedResource)

	err := client.Update(context.TODO(), expectedResource)
	if err != nil {
		// only need to delete resource as new will be recreated on next reconciliation
		(*reqLogger).Info("Could not update "+resTypeString+", due to having not compatible changes between expected and updated resource, "+
			"will try to delete it and create new one...", "Namespace", foundResource.GetNamespace(), "Name", foundResource.GetName())

		result, err := DeleteResource(reqLogger, client, foundResource)
		if err != nil {
			(*reqLogger).Error(err, "Failed deleting the resource")
			return result, err
		}

		// Can't create a new resource with a resource version provided, so remove it
		expectedResource.SetResourceVersion("")

		// Recreate the resource immediately (to avoid losing e.g. previously existing labels)
		(*reqLogger).Info("Recreating "+resTypeString, "Namespace", foundResource.GetNamespace(), "Name", foundResource.GetName())
		return reconcile.Result{}, client.Create(context.TODO(), expectedResource)
	}
	(*reqLogger).Info("Updated "+resTypeString+" successfully", "Namespace", expectedResource.GetNamespace(), "Name", expectedResource.GetName())
	// Resource updated - return and do not requeue as it might not consider extra values
	return reconcile.Result{}, nil
}

func UpdateServiceIfNeeded(reqLogger *logr.Logger, client c.Client, expectedService *corev1.Service, foundService *corev1.Service) (reconcile.Result, error) {
	for _, annotation := range annotationsForServicesToCheck {
		if foundService.Annotations[annotation] != expectedService.Annotations[annotation] {
			expectedService.Spec.ClusterIP = foundService.Spec.ClusterIP
			return UpdateResource(reqLogger, client, expectedService, foundService)
		}
	}
	return reconcile.Result{}, nil
}

func UpdateServiceMonitor(reqLogger *logr.Logger, client c.Client, expected, found *monitoringv1.ServiceMonitor) (reconcile.Result, error) {
	if expected == nil || found == nil {
		err := errors.New("cannot update to empty service monitor")
		(*reqLogger).Error(err, "Expected and found service monitor cannot be nil.")
		return reconcile.Result{}, err
	}
	updateResource := func() (reconcile.Result, error) {
		return UpdateResource(reqLogger, client, expected, found)
	}
	for _, annotation := range annotationsForServicesToCheck {
		if found.Annotations[annotation] != expected.Annotations[annotation] {
			return updateResource()
		}
	}
	expectedSpec := expected.Spec
	foundSpec := found.Spec
	// we assume only one endpoint, if changed in expected service monitor then modify this method as well
	if len(expectedSpec.Endpoints) != 1 {
		err := errors.New("expected service monitor endpoints error")
		(*reqLogger).Error(
			err, "Expected service monitor should have 1 endpoint, change this function otherwise.",
			"Namespace", found.GetNamespace(), "Name", found.GetName())
		return reconcile.Result{}, err
	}
	if len(foundSpec.Endpoints) != len(expectedSpec.Endpoints) {
		// deleting will also cause updating, this needs to be done when in-place update cannot work
		return updateResource()
	}
	expectedEndpoint := expectedSpec.Endpoints[0]
	foundEndpoint := foundSpec.Endpoints[0]
	if expectedEndpoint.Scheme != foundEndpoint.Scheme {
		return updateResource()
	}
	if expectedEndpoint.TargetPort != nil {
		if foundEndpoint.TargetPort == nil ||
			expectedEndpoint.TargetPort.StrVal != foundEndpoint.TargetPort.StrVal ||
			expectedEndpoint.TargetPort.IntVal != foundEndpoint.TargetPort.IntVal {
			return updateResource()
		}
	} else {
		if foundEndpoint.TargetPort != nil {
			return updateResource()
		}
	}
	if expectedEndpoint.Interval != foundEndpoint.Interval || expectedEndpoint.Path != foundEndpoint.Path {
		return updateResource()
	}
	if expectedEndpoint.RelabelConfigs != nil {
		if foundEndpoint.RelabelConfigs == nil ||
			len(expectedEndpoint.RelabelConfigs) != len(foundEndpoint.RelabelConfigs) {
			return updateResource()
		}
		// we assume only one relabeling, if changed in expected service monitor then modify this method as well
		if len(expectedEndpoint.RelabelConfigs) != 1 {
			err := errors.New("expected service monitor relabeling error")
			(*reqLogger).Error(
				err, "Expected service monitor should have 1 relabeling, change this function otherwise.",
				"Namespace", found.GetNamespace(), "Name", found.GetName())
			return reconcile.Result{}, err
		}
		expectedRelabeling := expectedEndpoint.RelabelConfigs[0]
		foundRelabeling := foundEndpoint.RelabelConfigs[0]
		if expectedRelabeling.Replacement != foundRelabeling.Replacement ||
			expectedRelabeling.TargetLabel != foundRelabeling.TargetLabel {
			return updateResource()
		}
	} else {
		if foundEndpoint.RelabelConfigs != nil {
			return updateResource()
		}
	}
	result, done, err := checkMetricRelabelConfigs(reqLogger, expectedEndpoint, foundEndpoint, updateResource, found)
	if done {
		return result, err
	}
	if expectedEndpoint.TLSConfig != nil {
		if foundEndpoint.TLSConfig == nil ||
			!apieq.Semantic.DeepEqual(expectedEndpoint.TLSConfig, foundEndpoint.TLSConfig) {
			return updateResource()
		}
	} else {
		if foundEndpoint.TLSConfig != nil {
			return updateResource()
		}
	}
	if !apieq.Semantic.DeepEqual(expectedSpec.Selector, foundSpec.Selector) {
		return updateResource()
	}
	return reconcile.Result{}, nil
}

func checkMetricRelabelConfigs(reqLogger *logr.Logger, expectedEndpoint monitoringv1.Endpoint, foundEndpoint monitoringv1.Endpoint,
	updateResource func() (reconcile.Result, error), found *monitoringv1.ServiceMonitor) (reconcile.Result, bool, error) {
	if expectedEndpoint.MetricRelabelConfigs != nil {
		if foundEndpoint.MetricRelabelConfigs == nil ||
			len(expectedEndpoint.MetricRelabelConfigs) != len(foundEndpoint.MetricRelabelConfigs) {
			result, err := updateResource()
			return result, true, err
		}
		// we assume only one relabeling, if changed in expected service monitor then modify this method as well
		if len(expectedEndpoint.MetricRelabelConfigs) != 1 {
			err := errors.New("expected service monitor metric relabeling error")
			(*reqLogger).Error(
				err, "Expected service monitor should have 1 metric relabeling, change this function otherwise.",
				"Namespace", found.GetNamespace(), "Name", found.GetName())
			return reconcile.Result{}, true, err
		}
		expectedRelabeling := expectedEndpoint.MetricRelabelConfigs[0]
		foundRelabeling := foundEndpoint.MetricRelabelConfigs[0]
		if len(expectedRelabeling.SourceLabels) != 1 {
			err := errors.New("expected service monitor metric relabeling error")
			(*reqLogger).Error(
				err, "Expected service monitor should have 1 metric relabeling source label, change this function otherwise.",
				"Namespace", found.GetNamespace(), "Name", found.GetName())
			return reconcile.Result{}, true, err
		}
		if expectedRelabeling.Action != foundRelabeling.Action ||
			expectedRelabeling.Regex != foundRelabeling.Regex ||
			len(expectedRelabeling.SourceLabels) != len(foundRelabeling.SourceLabels) ||
			expectedRelabeling.SourceLabels[0] != foundRelabeling.SourceLabels[0] {
			result, err := updateResource()
			return result, true, err
		}
	} else {
		if foundEndpoint.MetricRelabelConfigs != nil {
			result, err := updateResource()
			return result, true, err
		}
	}
	return reconcile.Result{}, false, nil
}

func DeleteResource(reqLogger *logr.Logger, client c.Client, foundResource ResourceObject) (reconcile.Result, error) {
	resTypeString := reflect.TypeOf(foundResource).String()
	err := client.Delete(context.TODO(), foundResource)
	if err != nil {
		if apierrors.IsNotFound(err) {
			(*reqLogger).Info("Could not delete "+resTypeString+", as it was already deleted", "Namespace", foundResource.GetNamespace(), "Name", foundResource.GetName())
		} else {
			(*reqLogger).Error(err, "Failed to delete "+resTypeString+" during recreation", "Namespace", foundResource.GetNamespace(), "Name", foundResource.GetName())
			return reconcile.Result{}, err
		}
	} else {
		// Resource deleted successfully - return and requeue to create new one
		(*reqLogger).Info("Deleted "+resTypeString+" successfully", "Namespace", foundResource.GetNamespace(), "Name", foundResource.GetName())
	}
	return reconcile.Result{Requeue: true, RequeueAfter: time.Second * 30}, nil
}

func UpdateOwner(reqLogger *logr.Logger, client c.Client, owner ResourceObject) (reconcile.Result, error) {
	resTypeString := reflect.TypeOf(owner).String()
	err := client.Get(context.TODO(), types.NamespacedName{Name: owner.GetName(), Namespace: owner.GetNamespace()}, owner)
	if err != nil {
		(*reqLogger).Error(err, "Failed to update owner data "+resTypeString+"", "Namespace", owner.GetNamespace(), "Name", owner.GetName())
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
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

func GetOCPPrometheusSecretCheckScript() string {
	script := `while true; do
  echo "$(date): Checking for ocp prometheus secret"
  ls /opt/prometheus/certs/* && break
  echo "$(date): Required ocp prometheus secret not found ... try again in 30s"
  sleep 30
done
echo "$(date): All required secrets exist"
`
	return script
}

func UpdateCacheClusterExtensions(client c.Reader) error {
	namespace, err := GetOperatorNamespace()
	if err != nil {
		return errors.New("OPERATOR_NAMESPACE env not found")
	}

	listOpts := []c.ListOption{
		c.InNamespace(namespace),
	}

	MeterDefinitionCRD := &rhmp.MeterDefinitionList{}
	if err := client.List(context.TODO(), MeterDefinitionCRD, listOpts...); err == nil {
		RHMPEnabled = true
	} else {
		RHMPEnabled = false
	}

	routeTestInstance := &routev1.RouteList{}
	if err := client.List(context.TODO(), routeTestInstance, listOpts...); err == nil {
		IsRouteAPI = true
	} else {
		IsRouteAPI = false
	}

	serviceCAInstance := &servicecav1.ServiceCAList{}
	if err := client.List(context.TODO(), serviceCAInstance, listOpts...); err == nil {
		IsServiceCAAPI = true
		IsAlertingEnabledByDefault = true
	} else {
		IsServiceCAAPI = false
		IsAlertingEnabledByDefault = false
	}

	odlmTestInstance := &odlm.OperandBindInfoList{}
	if err := client.List(context.TODO(), odlmTestInstance, listOpts...); err == nil {
		IsODLM = true
	} else {
		IsODLM = false
	}

	return nil
}

// MapHasAllPairsFromOther checks if all key, value pairs present in the second param are in the first param
func MapHasAllPairsFromOther[K, V comparable](checked, allNeededPairs map[K]V) bool {
	for key, value := range allNeededPairs {
		if foundValue, ok := checked[key]; !ok || foundValue != value {
			return false
		}
	}
	return true
}

// Returns true if configmaps are equal in terms of stored data
func CompareConfigMapData(found, expected *corev1.ConfigMap) bool {
	return apieq.Semantic.DeepEqual(found.Data, expected.Data) && MapHasAllPairsFromOther(found.Labels, expected.Labels) && apieq.Semantic.DeepEqual(found.BinaryData, expected.BinaryData)
}

// Returns true if secrets are equal in terms of stored data
func CompareSecretsData(s1, s2 *corev1.Secret) bool {
	return apieq.Semantic.DeepEqual(s1.Data, s2.Data) && apieq.Semantic.DeepEqual(s1.Labels, s2.Labels) && apieq.Semantic.DeepEqual(s1.Type, s2.Type) && apieq.Semantic.DeepEqual(s1.StringData, s2.StringData)
}

// Returns true if routes are equal
func CompareRoutes(reqLogger logr.Logger, expectedRoute, foundRoute *routev1.Route) bool {
	if foundRoute.ObjectMeta.Name != expectedRoute.ObjectMeta.Name {
		reqLogger.Info("Names not equal", "old", foundRoute.ObjectMeta.Name, "new", expectedRoute.ObjectMeta.Name)
		return false
	}
	if foundRoute.Spec.To.Name != expectedRoute.Spec.To.Name {
		reqLogger.Info("Specs To Name not equal",
			"old", fmt.Sprintf("%v", foundRoute.Spec),
			"new", fmt.Sprintf("%v", expectedRoute.Spec))
		return false
	}
	if foundRoute.Spec.TLS == nil && expectedRoute.Spec.TLS != nil {
		reqLogger.Info("Found Route has empty TLS options, but Expected Route has not empty TLS options",
			"old", fmt.Sprintf("%v", foundRoute.Spec.TLS),
			"new", fmt.Sprintf("%v", getTLSDataAsString(expectedRoute)))
		return false
	}
	if foundRoute.Spec.TLS != nil && expectedRoute.Spec.TLS == nil {
		reqLogger.Info("Expected Route has empty TLS options, but Found Route has not empty TLS options",
			"old", fmt.Sprintf("%v", getTLSDataAsString(foundRoute)),
			"new", fmt.Sprintf("%v", expectedRoute.Spec.TLS))
		return false
	}
	if foundRoute.Spec.TLS != nil && expectedRoute.Spec.TLS != nil {
		if foundRoute.Spec.TLS.Termination != expectedRoute.Spec.TLS.Termination {
			reqLogger.Info("Expected Route has different TLS Termination option than Found Route",
				"old", fmt.Sprintf("%v", foundRoute.Spec.TLS.Termination),
				"new", fmt.Sprintf("%v", expectedRoute.Spec.TLS.Termination))
			return false
		}
		if foundRoute.Spec.TLS.InsecureEdgeTerminationPolicy != expectedRoute.Spec.TLS.InsecureEdgeTerminationPolicy {
			reqLogger.Info("Expected Route has different TLS InsecureEdgeTerminationPolicy option than Found Route",
				"old", fmt.Sprintf("%v", foundRoute.Spec.TLS.InsecureEdgeTerminationPolicy),
				"new", fmt.Sprintf("%v", expectedRoute.Spec.TLS.InsecureEdgeTerminationPolicy))
			return false
		}
		if !areTLSCertsSame(*expectedRoute.Spec.TLS, *foundRoute.Spec.TLS) {
			reqLogger.Info("Expected route has different certificate info in the TLS section than Found Route",
				"old", fmt.Sprintf("%v", getTLSDataAsString(foundRoute)),
				"new", fmt.Sprintf("%v", getTLSDataAsString(expectedRoute)))
			return false
		}
	}
	return true
}

func areTLSCertsSame(expected, found routev1.TLSConfig) bool {
	return (expected.CACertificate == found.CACertificate &&
		expected.Certificate == found.Certificate &&
		expected.Key == found.Key &&
		expected.DestinationCACertificate == found.DestinationCACertificate)
}

func GenerateSelfSignedCertSecret(namespacedName types.NamespacedName, dns []string) (*corev1.Secret, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	// Generate a pem block with the private key
	keyPem := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})

	commonName := ""
	if len(dns) > 0 {
		commonName = dns[0]
	}

	// need to generate a different serial number each execution
	serialNumber, _ := rand.Int(rand.Reader, big.NewInt(1000000))

	tml := x509.Certificate{
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: []string{"IBM"},
		},
		BasicConstraintsValid: true,
	}

	if dns != nil {
		tml.DNSNames = dns
	}

	cert, err := x509.CreateCertificate(rand.Reader, &tml, &tml, &key.PublicKey, key)
	if err != nil {
		return nil, err
	}

	certPem := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert,
	})

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namespacedName.Name,
			Namespace: namespacedName.Namespace,
			Labels: map[string]string{
				"release": "ibm-licensing-service",
			},
		},
		Data: map[string][]byte{
			"tls.crt": certPem,
			"tls.key": keyPem,
		},
		Type: corev1.SecretTypeTLS,
	}, nil
}

func ProcessCerfiticateSecret(secret corev1.Secret) (cert, caCert, key string, err error) {

	certChain := string(secret.Data["tls.crt"])
	key = string(secret.Data["tls.key"])
	re := regexp.MustCompile("(?s)-----BEGIN CERTIFICATE-----.*?-----END CERTIFICATE-----")
	externalCerts := re.FindAllString(certChain, -1)

	if len(externalCerts) == 0 {
		err = errors.New("invalid certificate format under tls.crt section")
		return
	}

	cert = externalCerts[0]

	if len(externalCerts) == 2 {
		caCert = externalCerts[1]
	} else {
		caCert = ""
	}
	return
}

func ParseCertificate(rawCertData []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(rawCertData)

	if block != nil {
		return x509.ParseCertificate(block.Bytes)
	}

	return nil, errors.New("unable to decode pem block")
}

func getTLSDataAsString(route *routev1.Route) string {
	return fmt.Sprintf("{Termination: %v, InsecureEdgeTerminationPolicy: %v, Certificate: %s, CACertificate: %s, DestinationCACertificate: %s}",
		route.Spec.TLS.Termination, route.Spec.TLS.InsecureEdgeTerminationPolicy,
		route.Spec.TLS.Certificate, route.Spec.TLS.CACertificate, route.Spec.TLS.DestinationCACertificate)
}

/*
MergeWithSpecAnnotations attaches spec annotations to the provided map of predefined annotations.
*/
func mergeWithSpecAnnotations(instance *operatorv1alpha1.IBMLicensing, annotations map[string]string) map[string]string {
	if instance.Spec.Annotations != nil {
		for key, value := range instance.Spec.Annotations {
			annotations[key] = value
		}
	}

	return annotations
}
