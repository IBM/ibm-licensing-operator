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
	operatorv1alpha1 "github.com/ibm/ibm-licensing-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	v1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetNetworkPolicyName(instance *operatorv1alpha1.IBMLicensing) string {
	return GetResourceName(instance)
}

func GetNetworkPolicy(instance *operatorv1alpha1.IBMLicensing) *v1beta1.NetworkPolicy {
	protocol := corev1.ProtocolTCP
	return &v1beta1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetNetworkPolicyName(instance),
			Namespace: instance.Spec.InstanceNamespace,
		},
		Spec: v1beta1.NetworkPolicySpec{
			PodSelector: getNetworkPolicyPodSelector(),
			PolicyTypes: []v1beta1.PolicyType{v1beta1.PolicyTypeIngress},
			Ingress: []v1beta1.NetworkPolicyIngressRule{
				{
					Ports: []v1beta1.NetworkPolicyPort{
						{
							Port:     &prometheusServicePort,
							Protocol: &protocol,
						},
					},
					From: []v1beta1.NetworkPolicyPeer{
						{
							NamespaceSelector: getNetworkPolicyFromNamespaceSelector(),
							PodSelector:       getNetworkPolicyFromPodSelector(),
						},
					},
				},
				{
					Ports: []v1beta1.NetworkPolicyPort{
						{
							Port:     &licensingServicePort,
							Protocol: &protocol,
						},
					},
					From: []v1beta1.NetworkPolicyPeer{
						{
							NamespaceSelector: &metav1.LabelSelector{},
							PodSelector:       &metav1.LabelSelector{},
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

func getNetworkPolicyFromNamespaceSelector() *metav1.LabelSelector {
	return &metav1.LabelSelector{
		MatchLabels: map[string]string{
			MarketplaceMonitoringLabel: "true",
		},
	}
}
func getNetworkPolicyFromPodSelector() *metav1.LabelSelector {
	return &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"prometheus": MeterbaseLabel,
		},
	}
}
