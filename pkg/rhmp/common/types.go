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

// Package common contains shared types originally from
// github.com/redhat-marketplace/redhat-marketplace-operator/v2/apis/marketplace/common.
// Vendored locally to decouple from the upstream module which is incompatible
// with k8s.io v0.31+.
package common

// GroupVersionKind identifies a Kubernetes CRD by API version and kind.
type GroupVersionKind struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
}

// WorkloadType identifies the type of workload.
type WorkloadType string

const (
	WorkloadTypePod     WorkloadType = "Pod"
	WorkloadTypeService WorkloadType = "Service"
	WorkloadTypePVC     WorkloadType = "PersistentVolumeClaim"
)

// MetricType identifies the type of metric a meter definition reports.
type MetricType string

// NamespacedNameReference is a reference to a namespaced resource.
type NamespacedNameReference struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}
