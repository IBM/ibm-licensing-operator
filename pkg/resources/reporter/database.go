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

func GetDatabaseContainer(instance *operatorv1alpha1.IBMLicenseServiceReporter) corev1.Container {
	container := resources.GetContainerBase(instance.Spec.DatabaseContainer)
	container.EnvFrom = getDatabaseEnvFromSourceVariables()
	container.VolumeMounts = GetDatabaseVolumeMounts()
	container.Name = DatabaseContainerName
	container.LivenessProbe = resources.GetLivenessProbe(getDatabaseProbeHandler())
	container.ReadinessProbe = resources.GetReadinessProbe(getDatabaseProbeHandler())
	return container
}
