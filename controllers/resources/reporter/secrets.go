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

package reporter

import (
	operatorv1alpha1 "github.com/ibm/ibm-licensing-operator/api/v1alpha1"
	"github.com/ibm/ibm-licensing-operator/controllers/resources"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const APIReciverSecretTokenKeyName = "token"

func GetAPISecretToken(instance *operatorv1alpha1.IBMLicenseServiceReporter) (*corev1.Secret, error) {
	return resources.GetSecretToken(instance.Spec.APISecretToken, instance.GetNamespace(), APIReciverSecretTokenKeyName, LabelsForMeta(instance))
}

func GetDatabaseSecret(instance *operatorv1alpha1.IBMLicenseServiceReporter) (*corev1.Secret, error) {
	metaLabels := LabelsForMeta(instance)
	randString, err := resources.RandString(8)
	if err != nil {
		return nil, err
	}
	expectedSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      DatabaseConfigSecretName,
			Namespace: instance.GetNamespace(),
			Labels:    metaLabels,
		},
		Type: corev1.SecretTypeOpaque,
		StringData: map[string]string{
			PostgresPasswordKey:     randString,
			PostgresUserKey:         DatabaseUser,
			PostgresDatabaseNameKey: DatabaseName,
			PostgresPgDataKey:       PgData,
		},
	}
	return expectedSecret, nil
}
