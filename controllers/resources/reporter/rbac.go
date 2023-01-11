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
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	operatorv1alpha1 "github.com/IBM/ibm-licensing-operator/api/v1alpha1"
)

func GetServiceAccountName(instance *operatorv1alpha1.IBMLicenseServiceReporter) string {
	return GetResourceName(instance)
}

func GetServiceAccount(instance *operatorv1alpha1.IBMLicenseServiceReporter) *corev1.ServiceAccount {
	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetServiceAccountName(instance),
			Namespace: instance.GetNamespace(),
		},
	}
	if instance.Spec.ImagePullSecrets != nil {
		serviceAccount.ImagePullSecrets = []corev1.LocalObjectReference{}
		for _, imagePullSecret := range instance.Spec.ImagePullSecrets {
			serviceAccount.ImagePullSecrets = append(serviceAccount.ImagePullSecrets, corev1.LocalObjectReference{Name: imagePullSecret})
		}
	}
	return serviceAccount
}

func GetRole(instance *operatorv1alpha1.IBMLicenseServiceReporter) *rbacv1.Role {
	return &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetResourceName(instance),
			Namespace: instance.GetNamespace(),
		},
		Rules: []rbacv1.PolicyRule{{
			Verbs:     []string{"create", "get", "list", "update"},
			APIGroups: []string{""},
			Resources: []string{"configmaps"},
		}},
	}
}

func GetRoleBinding(instance *operatorv1alpha1.IBMLicenseServiceReporter) *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetResourceName(instance),
			Namespace: instance.GetNamespace(),
		},
		Subjects: []rbacv1.Subject{{
			APIGroup:  "",
			Kind:      "ServiceAccount",
			Name:      GetServiceAccountName(instance),
			Namespace: instance.GetNamespace(),
		}},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     GetResourceName(instance),
		},
	}
}
