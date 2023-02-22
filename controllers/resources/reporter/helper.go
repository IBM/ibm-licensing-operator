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

package reporter

import (
	"context"
	"fmt"
	"reflect"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	c "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	operatorv1alpha1 "github.com/IBM/ibm-licensing-operator/api/v1alpha1"
	res "github.com/IBM/ibm-licensing-operator/controllers/resources"
	"github.com/IBM/ibm-licensing-operator/version"
	odlm "github.com/IBM/operand-deployment-lifecycle-manager/api/v1alpha1"
)

const (
	DatabaseConfigSecretName = "license-service-hub-db-config"
	PostgresPasswordKey      = "POSTGRES_PASSWORD" // #nosec
	PostgresUserKey          = "POSTGRES_USER"
	PostgresDatabaseNameKey  = "POSTGRES_DB"
	PostgresPgDataKey        = "POSTGRES_PGDATA"

	DatabaseUser       = "postgres"
	DatabaseName       = "postgres"
	DatabaseMountPoint = "/var/lib/postgresql"
	PgData             = DatabaseMountPoint + "/pgdata"

	DatabaseContainerName = "database"
	ReceiverContainerName = "receiver"
	UIContainerName       = "reporter-ui"
	ReceiverPort          = 8080
	UIPort                = 3001

	LicenseReporterUIBase               = "ibm-license-service-reporter-ui"
	LicenseReporterResourceBase         = "ibm-license-service-reporter"
	LicenseReporterComponentName        = "ibm-license-service-reporter-svc"
	LicenseReporterReleaseName          = "ibm-license-service-reporter"
	LicenseReportOCPCertName            = "ibm-license-reporter-cert-internal"
	LicenseReportExternalCertName       = "ibm-license-reporter-cert"
	LicenseReportCustomExternalCertName = "ibm-licensing-reporter-certs"

	ZenConfigMapName          = "ibm-license-service-reporter-zen"
	LicenseReportBindInfoName = "ibm-license-service-reporter-bindinfo"
	OperandRegistry           = "common-service"
	ZenBindingName            = "public-zen-config"
)

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

func getReciverEnvVariables(spec operatorv1alpha1.IBMLicenseServiceReporterSpec) []corev1.EnvVar {
	environmentVariables := []corev1.EnvVar{
		{
			Name:  "HTTPS_CERTS_SOURCE",
			Value: string(operatorv1alpha1.ExternalCertsSource),
		},
	}
	if spec.IsDebug() {
		environmentVariables = append(environmentVariables, corev1.EnvVar{
			Name:  "logging.level.com.ibm",
			Value: "DEBUG",
		})
	}
	if spec.IsDebug() || spec.IsVerbose() {
		environmentVariables = append(environmentVariables, corev1.EnvVar{
			Name:  "SPRING_PROFILES_ACTIVE",
			Value: "verbose",
		})
	}
	if spec.EnvVariable != nil {
		for key, value := range spec.EnvVariable {
			environmentVariables = append(environmentVariables, corev1.EnvVar{
				Name:  key,
				Value: value,
			})
		}
	}
	return environmentVariables
}

func getEnvVariable(spec operatorv1alpha1.IBMLicenseServiceReporterSpec) []corev1.EnvVar {
	if spec.EnvVariable == nil {
		return nil
	}
	var environmentVariables = []corev1.EnvVar{}
	for key, value := range spec.EnvVariable {
		environmentVariables = append(environmentVariables, corev1.EnvVar{
			Name:  key,
			Value: value,
		})
	}
	return environmentVariables
}

func UpdateVersion(client client.Client, instance *operatorv1alpha1.IBMLicenseServiceReporter) error {
	if instance.Spec.Version != version.Version {
		instance.Spec.Version = version.Version
		return client.Update(context.TODO(), instance)
	}
	return nil
}

func AddSenderConfiguration(client client.Client, log logr.Logger) error {
	licensingList := &operatorv1alpha1.IBMLicensingList{}
	reqLogger := log.WithName("reconcileSenderConfiguration")

	operatorNs, _ := res.GetOperatorNamespace()
	listOpts := []c.ListOption{
		c.InNamespace(operatorNs),
	}

	err := client.List(context.TODO(), licensingList, listOpts...)
	if err != nil {
		reqLogger.Error(err, "Failed to get IBMLicensing resource")
		return err
	}
	if len(licensingList.Items) == 0 {
		reqLogger.Info("License Service not installed")
		return nil
	}

	for _, lic := range licensingList.Items {
		licensing := lic
		if licensing.Spec.SetDefaultSenderParameters() {
			err := client.Update(context.TODO(), &licensing)
			if err != nil {
				reqLogger.Error(err, fmt.Sprintf("Failed to configure sender for: %s", licensing.Name))
				return err
			}
			reqLogger.Info(fmt.Sprintf("Successfully configured sender for %s", licensing.Name))
		}
	}
	return nil
}

func ClearDefaultSenderConfiguration(client client.Client, log logr.Logger) {
	licensingList := &operatorv1alpha1.IBMLicensingList{}
	reqLogger := log.WithName("reconcileSenderConfiguration")

	operatorNs, _ := res.GetOperatorNamespace()
	listOpts := []c.ListOption{
		c.InNamespace(operatorNs),
	}

	err := client.List(context.TODO(), licensingList, listOpts...)
	if err != nil {
		reqLogger.Error(err, "Failed to get IBMLicensing resource")
		return
	}
	if len(licensingList.Items) == 0 {
		reqLogger.Info("License Service not installed")
		return
	}

	for _, lic := range licensingList.Items {
		licensing := lic
		if licensing.Spec.RemoveDefaultSenderParameters() {
			err := client.Update(context.TODO(), &licensing)
			if err != nil {
				reqLogger.Error(err, fmt.Sprintf("Failed to removed sender for: %s", licensing.Name))
				return
			}
			reqLogger.Info(fmt.Sprintf("Successfully removed sender for %s", licensing.Name))

		}
	}
}

func GetBindInfo(instance *operatorv1alpha1.IBMLicenseServiceReporter) *odlm.OperandBindInfo {
	return &odlm.OperandBindInfo{
		ObjectMeta: metav1.ObjectMeta{
			Name:      LicenseReportBindInfoName,
			Namespace: instance.GetNamespace(),
		},
		Spec: odlm.OperandBindInfoSpec{
			Operand:           res.OperatorName,
			Registry:          OperandRegistry,
			RegistryNamespace: instance.GetNamespace(),
			Description:       "Binding information that should be accessible to IBM License Service Reporter adopters",
			Bindings: map[string]odlm.SecretConfigmap{
				ZenBindingName: {
					Configmap: ZenConfigMapName,
				},
			},
		},
	}
}

func UpdateOperandBindInfoIfNeeded(reqLogger *logr.Logger, client client.Client, expectedBindInfo *odlm.OperandBindInfo,
	foundBindInfo *odlm.OperandBindInfo) (reconcile.Result, error) {

	if !reflect.DeepEqual(expectedBindInfo.Spec, foundBindInfo.Spec) {
		return res.UpdateResource(reqLogger, client, expectedBindInfo, foundBindInfo)
	}
	return reconcile.Result{}, nil
}
