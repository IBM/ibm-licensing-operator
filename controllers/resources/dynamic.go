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

package resources

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// List any resource without having its model as a project dependency.
func ListResourcesDynamically(ctx context.Context, dynamic dynamic.Interface, group, version,
	resource, namespace string) ([]unstructured.Unstructured, error) {

	resourceID := schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resource,
	}
	list, err := dynamic.Resource(resourceID).Namespace(namespace).
		List(ctx, metav1.ListOptions{})

	if err != nil {
		return nil, err
	}

	return list.Items, nil
}

// Get any resource without having its model as a project dependency.
func GetResourceDynamically(ctx context.Context, dynamic dynamic.Interface, group, version, resource,
	name, namespace string) (*unstructured.Unstructured, error) {

	resourceID := schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resource,
	}
	obj, err := dynamic.Resource(resourceID).Namespace(namespace).
		Get(ctx, name, metav1.GetOptions{})

	if err != nil {
		return nil, err
	}

	return obj, nil
}
