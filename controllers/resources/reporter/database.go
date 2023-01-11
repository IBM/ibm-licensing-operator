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
	corev1 "k8s.io/api/core/v1"

	operatorv1alpha1 "github.com/IBM/ibm-licensing-operator/api/v1alpha1"
	"github.com/IBM/ibm-licensing-operator/controllers/resources"
)

func getDatabaseProbeHandler() corev1.ProbeHandler {
	return corev1.ProbeHandler{
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
	container.Env = getEnvVariable(instance.Spec)
	container.VolumeMounts = getDatabaseVolumeMounts()
	container.Name = DatabaseContainerName
	container.LivenessProbe = resources.GetLivenessProbe(getDatabaseProbeHandler())
	container.ReadinessProbe = resources.GetReadinessProbe(getDatabaseProbeHandler())
	return container
}
