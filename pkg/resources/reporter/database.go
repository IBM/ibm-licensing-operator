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
	"github.com/ibm/ibm-licensing-operator/pkg/resources"
	corev1 "k8s.io/api/core/v1"
)

func getDatabaseEnvironmentVariables() []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name: "POSTGRES_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: DatabaseConfigSecretName,
					},
					Key: PostgresPasswordKey,
				},
			},
		},
		{
			Name: "POSTGRES_USER",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: DatabaseConfigSecretName,
					},
					Key: PostgresUserKey,
				},
			},
		},
		{
			Name: "DATABASE",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: DatabaseConfigSecretName,
					},
					Key: PostgresDatabaseNameKey,
				},
			},
		},
		{
			Name: "PGDATA",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: DatabaseConfigSecretName,
					},
					Key: PostgresPgDataKey,
				},
			},
		},
	}

}

func getDatabaseProbeHandler() corev1.Handler {
	return corev1.Handler{
		Exec: &corev1.ExecAction{
			Command: []string{
				"psql",
				"-w",
				"-U",
				DatabaseUser,
				"-d",
				DatabaseName,
				"-c",
				"SELECT 1",
			},
		},
	}
}

func GetDatabaseContainer(spec operatorv1alpha1.IBMLicenseServiceReporterSpec, instance *operatorv1alpha1.IBMLicenseServiceReporter) corev1.Container {
	container := resources.GetContainerBase(spec.DatabaseContainer)
	container.Env = getDatabaseEnvironmentVariables()
	container.Resources = instance.Spec.DatabaseContainer.Resources
	container.VolumeMounts = GetDatabaseVolumeMounts()
	container.Name = DatabaseContainerName
	container.LivenessProbe = resources.GetLivenessProbe(getDatabaseProbeHandler())
	container.ReadinessProbe = resources.GetReadinessProbe(getDatabaseProbeHandler())
	return container
}
