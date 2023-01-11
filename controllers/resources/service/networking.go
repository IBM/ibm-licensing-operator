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
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	operatorv1alpha1 "github.com/IBM/ibm-licensing-operator/api/v1alpha1"
)

func GetNetworkPolicyName(instance *operatorv1alpha1.IBMLicensing) string {
	return GetResourceName(instance)
}

func GetNetworkPolicy(instance *operatorv1alpha1.IBMLicensing) *networkingv1.NetworkPolicy {
	protocol := corev1.ProtocolTCP
	prometheusPorts := []networkingv1.NetworkPolicyPort{
		{
			Port:     &prometheusServicePort,
			Protocol: &protocol,
		},
	}
	defaultPrometheusPeer := []networkingv1.NetworkPolicyPeer{
		{
			NamespaceSelector: getOneKeyValueLabelSelector(
				"marketplace.redhat.com/operator", "true"),
			PodSelector: getOneKeyValueLabelSelector(
				"prometheus", "rhm-marketplaceconfig-meterbase"),
		},
	}
	userWorkloadPrometheusPeer := []networkingv1.NetworkPolicyPeer{
		{
			NamespaceSelector: getOneKeyValueLabelSelector(
				"name", "openshift-user-workload-monitoring"),
			PodSelector: getOneKeyValueLabelSelector(
				"prometheus", "user-workload"),
		},
	}
	return &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetNetworkPolicyName(instance),
			Namespace: instance.Spec.InstanceNamespace,
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: getNetworkPolicyPodSelector(),
			PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress},
			Ingress: []networkingv1.NetworkPolicyIngressRule{
				{
					Ports: prometheusPorts,
					From:  defaultPrometheusPeer,
				},
				{
					Ports: prometheusPorts,
					From:  userWorkloadPrometheusPeer,
				},
				{
					Ports: []networkingv1.NetworkPolicyPort{
						{
							Port:     &licensingServicePort,
							Protocol: &protocol,
						},
					},
					From: []networkingv1.NetworkPolicyPeer{
						{
							NamespaceSelector: &metav1.LabelSelector{},
						},
					},
				},
			},
		},
	}
}

func getNetworkPolicyPodSelector() metav1.LabelSelector {
	return metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": LicensingServiceAppLabel,
		},
	}
}

func getOneKeyValueLabelSelector(key string, value string) *metav1.LabelSelector {
	return &metav1.LabelSelector{
		MatchLabels: map[string]string{
			key: value,
		},
	}
}
