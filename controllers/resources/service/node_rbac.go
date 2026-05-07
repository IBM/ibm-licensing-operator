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

package service

import (
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	operatorv1alpha1 "github.com/IBM/ibm-licensing-operator/api/v1alpha1"
)

// IBMLicensing is cluster-scoped and exists once per cluster, so a single
// shared ClusterRole + ClusterRoleBinding is created and reused regardless of
// the CR's name.
const NodeAccessClusterRoleName = "ibm-license-service-nodes"

func GetExpectedNodeClusterRole(instance *operatorv1alpha1.IBMLicensing) *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:   NodeAccessClusterRoleName,
			Labels: LabelsForMeta(instance),
		},
		Rules: []rbacv1.PolicyRule{{
			APIGroups: []string{""},
			Resources: []string{"nodes"},
			Verbs:     []string{"get", "list"},
		}},
	}
}

func GetExpectedNodeClusterRoleBinding(instance *operatorv1alpha1.IBMLicensing) *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:   NodeAccessClusterRoleName,
			Labels: LabelsForMeta(instance),
		},
		Subjects: []rbacv1.Subject{{
			Kind:      "ServiceAccount",
			Name:      GetServiceAccountName(instance),
			Namespace: instance.Spec.InstanceNamespace,
		}},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     NodeAccessClusterRoleName,
		},
	}
}
