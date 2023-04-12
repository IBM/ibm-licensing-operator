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
	"context"

	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	ctrl "sigs.k8s.io/controller-runtime"
)

// +kubebuilder:rbac:namespace=ibm-licensing,groups=operator.ibm.com,resources=namespacescope;namespacescope/finalizers;namespacescope/status,verbs=get;list;watch

func IsNamespaceScopeOperatorInstalled() (bool, error) {
	nssCr, err := getNamespaceScopeCR()
	if err != nil {
		if k8serr.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	return nssCr != nil, nil
}

// Get namespaces to watch from Namespace Scope Operator ConfigMap. Should only be used in CP2/CP3 coexistence scenario.
func getWatchNamespaceFromNssConfigMap() (string, error) {
	ctx := context.Background()
	config := ctrl.GetConfigOrDie()
	dynamicClient := dynamic.NewForConfigOrDie(config)
	namespace, err := GetOperatorNamespace()
	if err != nil {
		return "", err // TODO
	}

	// List Namespace Scope CRs in the namespace.
	nssCrList, err := ListResourcesDynamically(ctx, dynamicClient, "operator.ibm.com", "v1", "namespacescope", namespace)
	if err != nil {
		return "", err // TODO
	}

	var nssConfigmapName string
	if len(nssCrList) == 0 {
		return "", nil // TODO
	}

	// Find first CR with configmapName set.
	for _, nssCr := range nssCrList {
		if spec, ok := nssCr.Object["spec"]; ok {
			if cmName, ok := spec.(map[string]interface{})["configmapName"]; ok {
				nssConfigmapName = cmName.(string)
				break
			}
		}
	}
	if nssConfigmapName == "" {
		return "", nil // TODO
	}

	// Get Config Map with name specified in Namespace Scope CR
	nssCm, err := GetResourceDynamically(ctx, dynamicClient, nssConfigmapName, "", "v1", "configmap", namespace)
	if err != nil {
		return "", err
	}
	if nssCm == nil {
		return "", nil
	}

	if data, ok := nssCm.Object["data"]; ok {
		if namespaces, ok := data.(map[string]interface{})["namespaces"]; ok {
			return namespaces.(string), nil
		}
	}

	return "", nil // TODO
}

// Get first found Namespace Scope CR
func getNamespaceScopeCR() (*unstructured.Unstructured, error) {
	ctx := context.Background()
	config := ctrl.GetConfigOrDie()
	dynamicClient := dynamic.NewForConfigOrDie(config)
	namespace, err := GetOperatorNamespace()
	if err != nil {
		return nil, err
	}

	// List Namespace Scope Operator CRs in the namespace.
	nssCrList, err := ListResourcesDynamically(ctx, dynamicClient, "operator.ibm.com", "v1", "namespacescope", namespace)
	if err != nil {
		return nil, err
	}

	if len(nssCrList) == 0 {
		return nil, nil
	}

	return &nssCrList[0], nil
}
