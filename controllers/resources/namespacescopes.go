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

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const namespaceScopeCmName = "namespace-scope"

// Return true if Namespace Scope Config Map is available in operator's namespace.
func IsNamespaceScopeOperatorAvailable(ctx context.Context, reader client.Reader, namespace string) (bool, error) {
	nssCm := corev1.ConfigMap{}
	err := reader.Get(ctx, types.NamespacedName{Namespace: namespace, Name: namespaceScopeCmName}, &nssCm)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
