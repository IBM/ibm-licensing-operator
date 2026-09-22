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

package service

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"

	operatorv1alpha1 "github.com/IBM/ibm-licensing-operator/api/v1alpha1"
	res "github.com/IBM/ibm-licensing-operator/controllers/resources"
	"github.com/IBM/ibm-licensing-operator/version"
)

// maxDNSLabelLength is the maximum number of characters allowed in a single DNS label per RFC 1123.
const maxDNSLabelLength = 63

const (
	LicensingResourceBase                = "ibm-licensing-service"
	LicensingComponentName               = "ibm-licensing-service-svc"
	LicensingReleaseName                 = "ibm-licensing-service"
	LicenseServiceInternalCertName       = "ibm-license-service-cert-internal"
	PrometheusServiceOCPCertName         = "ibm-licensing-service-prometheus-cert"
	LicenseServiceExternalCertName       = "ibm-license-service-cert"
	LicenseServiceCustomExternalCertName = "ibm-licensing-certs"
	LicensingServiceAccount              = "ibm-license-service"
	LicensingServiceAccountRestricted    = "ibm-license-service-restricted"
	PrometheusServiceName                = "ibm-licensing-service-prometheus"
	PrometheusRHMPServiceMonitor         = "ibm-licensing-service-service-monitor"
	PrometheusAlertingServiceMonitor     = "ibm-licensing-service-service-monitor-alerting"

	LicensingServiceAppLabel = "ibm-licensing-service-instance"

	//goland:noinspection GoNameStartsWithPackageName
	ServiceMonitorSelectorLabel = "marketplace.redhat.com/metering"
	ReleaseLabel                = "ibm-licensing-service-prometheus"

	NamespaceScopeLabelKey   = "intent"
	NamespaceScopeLabelValue = "projected"

	//goland:noinspection GoNameStartsWithPackageName
	ServiceAccountSecretName        = "ibm-licensing-service-account-token"
	DefaultReaderTokenName          = "ibm-licensing-default-reader-token"
	DefaultReaderServiceAccountName = "ibm-licensing-default-reader"

	LicensingToken        = "ibm-licensing-token"
	LicensingInfo         = "ibm-licensing-info"
	LicensingUploadToken  = "ibm-licensing-upload-token"
	LicensingUploadConfig = "ibm-licensing-upload-config"

	ActiveCRState   = "ACTIVE"
	InactiveCRState = "INACTIVE"

	SchemeHTTP  = "http"
	SchemeHTTPS = "https"
)

func GetServiceAccountName(instance *operatorv1alpha1.IBMLicensing) string {
	if instance.Spec.IsNamespaceScopeEnabled() {
		return LicensingServiceAccountRestricted
	}
	return LicensingServiceAccount
}

func GetResourceName(instance *operatorv1alpha1.IBMLicensing) string {
	return LicensingResourceBase + "-" + instance.GetName()
}

// TruncateForDNSLabel truncates namespace so that routeName+"-"+namespace does not exceed maxDNSLabelLength (63).
// If the combined label already fits, namespace is returned unchanged.
// ILS-3012: OpenShift Route admission rejects spec.host when the first DNS label exceeds 63 characters.
func TruncateForDNSLabel(routeName, namespace string) string {
	separator := "-"
	maxNamespaceLen := maxDNSLabelLength - len(routeName) - len(separator)
	if maxNamespaceLen <= 0 {
		// routeName alone already fills the label — truncate namespace to empty string
		// (this should not happen in practice given ILS resource names)
		return ""
	}
	if len(namespace) <= maxNamespaceLen {
		return namespace
	}
	return namespace[:maxNamespaceLen]
}

// GetLicensingRouteHostname returns the explicit spec.host value for the ILS Route.
// It ensures the first DNS label (routeName+"-"+namespace) does not exceed 63 characters by
// truncating the namespace segment if necessary, then appends the cluster apps domain.
// appsDomain must be the bare domain (e.g. "apps.example.com") without a leading dot.
func GetLicensingRouteHostname(routeName, namespace, appsDomain string) string {
	truncatedNamespace := TruncateForDNSLabel(routeName, namespace)
	return routeName + "-" + truncatedNamespace + "." + appsDomain
}

func GetServiceURL(instance *operatorv1alpha1.IBMLicensing) string {
	var urlPrefix string
	if instance.Spec.HTTPSEnable {
		urlPrefix = "https://"
	} else {
		urlPrefix = "http://"
	}
	return urlPrefix + GetResourceName(instance) + "." + instance.Spec.InstanceNamespace + ".svc.cluster.local:" + licensingServicePort.String()
}

func GetServiceHostname(instance *operatorv1alpha1.IBMLicensing) string {
	return GetResourceName(instance) + "." + instance.Spec.InstanceNamespace + ".svc.cluster.local"
}

/*
MergeWithSpecLabels attaches spec labels to the provided map of predefined labels.

Helps cover some cases and optimize code (so that reconcile functions' `maybeAttachSpecLabels` only do the minimal
check) by pre-attaching the labels wherever `LabelsFor<Type>` functions are used (generally on resource creation).

In the future, labels addition in reconciliation should be gone and only present in this helper module (or similar).
However, this would require many changes to the general code flow and usage of update and create functionalities.
*/
func MergeWithSpecLabels(instance *operatorv1alpha1.IBMLicensing, labels map[string]string) map[string]string {
	if instance.Spec.Labels != nil {
		for key, value := range instance.Spec.Labels {
			labels[key] = value
		}
	}

	return labels
}

func LabelsForSelector(instance *operatorv1alpha1.IBMLicensing) map[string]string {
	return MergeWithSpecLabels(instance, map[string]string{
		"app":          GetResourceName(instance),
		"component":    LicensingComponentName,
		"licensing_cr": instance.GetName(),
	})
}

func LabelsForMeta(instance *operatorv1alpha1.IBMLicensing) map[string]string {
	return MergeWithSpecLabels(instance, map[string]string{
		"app.kubernetes.io/name":       GetResourceName(instance),
		"app.kubernetes.io/component":  LicensingComponentName,
		"app.kubernetes.io/managed-by": "operator",
		"app.kubernetes.io/instance":   LicensingReleaseName,
		res.LicensingReleaseLabelKey:   LicensingReleaseName,
	})
}

func LabelsForServiceMonitor(instance *operatorv1alpha1.IBMLicensing) map[string]string {
	return MergeWithSpecLabels(instance, map[string]string{ServiceMonitorSelectorLabel: "true"})
}

func LabelsForLicensingPod(instance *operatorv1alpha1.IBMLicensing) map[string]string {
	podLabels := LabelsForMeta(instance)
	selectorLabels := LabelsForSelector(instance)
	for key, value := range selectorLabels {
		podLabels[key] = value
	}
	return podLabels
}

func UpdateVersion(client client.Client, instance *operatorv1alpha1.IBMLicensing) error {
	if instance.Spec.Version != version.Version {
		instance.Spec.Version = version.Version
		return client.Update(context.TODO(), instance)
	}
	return nil
}
