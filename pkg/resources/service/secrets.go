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

package service

import (
	operatorv1alpha1 "github.com/ibm/ibm-licensing-operator/pkg/apis/operator/v1alpha1"
	res "github.com/ibm/ibm-licensing-operator/pkg/resources"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const APIUploadTokenName = "ibm-licensing-upload-token"
const APISecretTokenKeyName = "token"
const APIUploadTokenKeyName = "token-upload"

const ReporterSecretTokenKeyName = "token"

const UploadConfigMapKey = "url"

func GetAPISecretToken(instance *operatorv1alpha1.IBMLicensing) (*corev1.Secret, error) {
	return res.GetSecretToken(instance.Spec.APISecretToken, instance.Spec.InstanceNamespace, APISecretTokenKeyName, LabelsForMeta(instance))
}

func GetUploadToken(instance *operatorv1alpha1.IBMLicensing) (*corev1.Secret, error) {
	return res.GetSecretToken(APIUploadTokenName, instance.Spec.InstanceNamespace, APIUploadTokenKeyName, LabelsForMeta(instance))
}

func GetUploadConfigMap(instance *operatorv1alpha1.IBMLicensing) *corev1.ConfigMap {
	metaLabels := LabelsForMeta(instance)
	expectedCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ibm-licensing-upload-config",
			Namespace: instance.Spec.InstanceNamespace,
			Labels:    metaLabels,
		},
		Data: map[string]string{UploadConfigMapKey: GetUploadURL(instance)},
	}
	return expectedCM
}
