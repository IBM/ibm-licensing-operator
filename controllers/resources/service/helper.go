//
// Copyright 2021 IBM Corporation
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

	operatorv1alpha1 "github.com/ibm/ibm-licensing-operator/api/v1alpha1"
	"github.com/ibm/ibm-licensing-operator/version"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const LicensingResourceBase = "ibm-licensing-service"
const LicensingComponentName = "ibm-licensing-service-svc"
const LicensingReleaseName = "ibm-licensing-service"
const LicenseServiceOCPCertName = "ibm-license-service-cert"
const PrometheusServiceOCPCertName = "ibm-licensing-service-prometheus-cert"
const LicensingServiceAccount = "ibm-license-service"
const LicensingServiceAccountRestricted = "ibm-license-service-restricted"
const UsageServiceName = "ibm-licensing-service-usage"
const PrometheusServiceName = "ibm-licensing-service-prometheus"
const PrometheusServiceMonitor = "ibm-licensing-service-service-monitor"

const LicensingServiceAppLabel = "ibm-licensing-service-instance"
const ServiceMonitorSelectorLabel = "marketplace.redhat.com/metering"
const ReleaseLabel = "ibm-licensing-service-prometheus"
const ReleaseUsageLabel = "ibm-licensing-service-usage"

const NamespaceScopeLabelKey = "intent"
const NamespaceScopeLabelValue = "projected"

func GetResourceName(instance *operatorv1alpha1.IBMLicensing) string {
	return LicensingResourceBase + "-" + instance.GetName()
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

func LabelsForSelector(instance *operatorv1alpha1.IBMLicensing) map[string]string {
	return map[string]string{"app": GetResourceName(instance), "component": LicensingComponentName, "licensing_cr": instance.GetName()}
}

func LabelsForMeta(instance *operatorv1alpha1.IBMLicensing) map[string]string {
	return map[string]string{"app.kubernetes.io/name": GetResourceName(instance), "app.kubernetes.io/component": LicensingComponentName,
		"app.kubernetes.io/managed-by": "operator", "app.kubernetes.io/instance": LicensingReleaseName, "release": LicensingReleaseName}
}

func LabelsForServiceMonitor() map[string]string {
	return map[string]string{
		ServiceMonitorSelectorLabel: "true"}
}

func LabelsForLicensingPod(instance *operatorv1alpha1.IBMLicensing) map[string]string {
	podLabels := LabelsForMeta(instance)
	selectorLabels := LabelsForSelector(instance)
	for key, value := range selectorLabels {
		podLabels[key] = value
	}
	if instance.Spec.IsNamespaceScopeEnabled() {
		podLabels[NamespaceScopeLabelKey] = NamespaceScopeLabelValue
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
