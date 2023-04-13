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

	"k8s.io/client-go/dynamic"
	ctrl "sigs.k8s.io/controller-runtime"
)

const namespaceScopeCmName = "namespace-scope"

// +kubebuilder:rbac:namespace=ibm-licensing,groups=operator.ibm.com,resources=namespacescopes;namespacescopes/finalizers;namespacescopes/status,verbs=get;list;watch

// Return true if Namespace Scope Config Map is available in operator's namespace.
func IsNamespaceScopeOperatorAvailable() (bool, error) {
	ctx := context.Background()
	config := ctrl.GetConfigOrDie()
	dynamicClient := dynamic.NewForConfigOrDie(config)
	namespace, err := GetOperatorNamespace()
	if err != nil {
		return false, err
	}

	nssCrExists, err := isNamespaceScopeCmExists(ctx, dynamicClient, namespace)
	if err != nil {
		return false, err
	}

	return nssCrExists, nil
}

func isNamespaceScopeCmExists(ctx context.Context, client dynamic.Interface, namespace string) (bool, error) {
	nssCm, err := GetResourceDynamically(ctx, client, "", "v1", "configmaps", namespaceScopeCmName, namespace)
	if err != nil {
		return false, err
	}
	if nssCm == nil {
		return false, nil
	}

	return true, nil
}
