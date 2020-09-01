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

package reporter

import (
	operatorv1alpha1 "github.com/ibm/ibm-licensing-operator/pkg/apis/operator/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

const DatabaseConfigSecretName = "license-service-hub-db-config"
const PostgresPasswordKey = "POSTGRES_PASSWORD" // #nosec
const PostgresUserKey = "POSTGRES_USER"
const PostgresDatabaseNameKey = "POSTGRES_DB"
const PostgresPgDataKey = "POSTGRES_PGDATA"

const DatabaseUser = "postgres"
const DatabaseName = "postgres"
const DatabaseMountPoint = "/var/lib/postgresql"
const PgData = DatabaseMountPoint + "/pgdata"

const DatabaseContainerName = "database"
const ReceiverContainerName = "receiver"
const UIContainerName = "reporter-ui"
const ReceiverPort = 8080
const UIPort = 3001

const LicenseReporterUIBase = "ibm-license-service-reporter-ui"
const LicenseReporterResourceBase = "ibm-license-service-reporter"
const LicenseReporterComponentName = "ibm-license-service-reporter-svc"
const LicenseReporterReleaseName = "ibm-license-service-reporter"
const LicenseReportOCPCertName = "ibm-license-reporter-cert"

func GetResourceName(instance *operatorv1alpha1.IBMLicenseServiceReporter) string {
	return LicenseReporterResourceBase + "-" + instance.GetName()
}

func LabelsForSelector(instance *operatorv1alpha1.IBMLicenseServiceReporter) map[string]string {
	return map[string]string{"app": GetResourceName(instance), "component": LicenseReporterComponentName, "licensing_cr": instance.GetName()}
}

func LabelsForMeta(instance *operatorv1alpha1.IBMLicenseServiceReporter) map[string]string {
	return map[string]string{"app.kubernetes.io/name": GetResourceName(instance), "app.kubernetes.io/component": LicenseReporterComponentName,
		"app.kubernetes.io/managed-by": "operator", "app.kubernetes.io/instance": LicenseReporterReleaseName, "release": LicenseReporterReleaseName}
}

func LabelsForPod(instance *operatorv1alpha1.IBMLicenseServiceReporter) map[string]string {
	podLabels := LabelsForMeta(instance)
	selectorLabels := LabelsForSelector(instance)
	for key, value := range selectorLabels {
		podLabels[key] = value
	}
	return podLabels
}

func getDatabaseEnvFromSourceVariables() []corev1.EnvFromSource {
	return []corev1.EnvFromSource{
		{
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: DatabaseConfigSecretName,
				},
			},
		},
	}
}
