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

package resources

import (
	v1 "github.com/operator-framework/api/pkg/operators/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	odlm "github.com/IBM/operand-deployment-lifecycle-manager/api/v1alpha1"
)

const (
	SUCCESS = "\u2713"
	FAIL    = "\u2717"
)

func OperatorGroupObj(name, namespace string, annotations map[string]string, targetNamespaces []string) v1.OperatorGroup {
	return v1.OperatorGroup{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "operators.coreos.com/v1",
			Kind:       "OperatorGroup",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Annotations: annotations,
		},
		Spec: v1.OperatorGroupSpec{
			TargetNamespaces: targetNamespaces,
		},
	}
}

func OperandRequestObj(name, namespace, requestedOperandName string) odlm.OperandRequest {
	return odlm.OperandRequest{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "operator.ibm.com/v1alpha1",
			Kind:       "OperandRequest",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: odlm.OperandRequestSpec{
			Requests: []odlm.Request{
				{
					Registry:          "registry-name",
					RegistryNamespace: "registry-namespace",
					Operands: []odlm.Operand{
						{
							Name: requestedOperandName,
						},
					},
				},
			},
		},
	}
}

func OperandRequestObjWithBindings(name, namespace, requestedOperandName string) odlm.OperandRequest {
	return odlm.OperandRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: odlm.OperandRequestSpec{
			Requests: []odlm.Request{
				{
					Registry:          "registry-name",
					RegistryNamespace: "registry-namespace",
					Operands: []odlm.Operand{
						{
							Name: requestedOperandName,
							Bindings: map[string]odlm.SecretConfigmap{
								"public-api-data": {
									Secret:    "secret1",
									Configmap: "cm1",
								},
								"public-api-token": {
									Secret: "secret1",
								},
								"public-api-upload": {
									Secret:    "secret2",
									Configmap: "cm2",
								},
							},
						},
					},
				},
			},
		},
	}
}

func LicensingOperandBindInfo(name, namespace string) odlm.OperandBindInfo {
	return odlm.OperandBindInfo{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: odlm.OperandBindInfoSpec{
			Bindings: map[string]odlm.SecretConfigmap{
				"public-api-data": {
					Configmap: "ibm-licensing-info",
					Secret:    "ibm-licensing-token",
				},
				"public-api-token": {
					Secret: "ibm-licensing-token",
				},
				"public-api-upload": {
					Configmap: "ibm-licensing-upload-config",
					Secret:    "ibm-licensing-upload-token",
				},
			},
			Description: "Binding information that should be accessible to licensing adopters",
			Operand:     "ibm-licensing-operator",
			Registry:    "common-service",
		},
	}
}

func SecretObj(name, namespace string, data, labels, annotations map[string]string) corev1.Secret {
	return corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		StringData: data,
		Type:       corev1.SecretTypeOpaque,
	}
}

func ConfigMapObj(name, namespace string, data, labels, annotations map[string]string) corev1.ConfigMap {
	return corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Data: data,
	}
}
