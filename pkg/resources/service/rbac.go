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
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetServiceAccountName(instance *operatorv1alpha1.IBMLicensing) string {
	return GetResourceName(instance)
}

func GetLicensingServiceAccount(instance *operatorv1alpha1.IBMLicensing) *corev1.ServiceAccount {
	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetServiceAccountName(instance),
			Namespace: instance.Spec.InstanceNamespace,
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

func GetLicensingRole(instance *operatorv1alpha1.IBMLicensing) *rbacv1.Role {
	return &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetResourceName(instance),
			Namespace: instance.Spec.InstanceNamespace,
		},
		Rules: []rbacv1.PolicyRule{{
			Verbs:     []string{"create", "get", "list", "update"},
			APIGroups: []string{""},
			Resources: []string{"configmaps"},
		}},
	}
}

func GetLicensingRoleBinding(instance *operatorv1alpha1.IBMLicensing) *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetResourceName(instance),
			Namespace: instance.Spec.InstanceNamespace,
		},
		Subjects: []rbacv1.Subject{{
			APIGroup:  "",
			Kind:      "ServiceAccount",
			Name:      GetServiceAccountName(instance),
			Namespace: instance.Spec.InstanceNamespace,
		}},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     GetResourceName(instance),
		},
	}
}

func GetLicensingClusterRole(instance *operatorv1alpha1.IBMLicensing) *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: GetResourceName(instance),
		},
		Rules: []rbacv1.PolicyRule{{
			Verbs:     []string{"get", "list"},
			APIGroups: []string{""},
			Resources: []string{"pods", "nodes", "namespaces"},
		}},
	}
}

func GetLicensingClusterRoleBinding(instance *operatorv1alpha1.IBMLicensing) *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: GetResourceName(instance),
		},
		Subjects: []rbacv1.Subject{{
			APIGroup:  "",
			Kind:      "ServiceAccount",
			Name:      GetServiceAccountName(instance),
			Namespace: instance.Spec.InstanceNamespace,
		}},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     GetResourceName(instance),
		},
	}
}
