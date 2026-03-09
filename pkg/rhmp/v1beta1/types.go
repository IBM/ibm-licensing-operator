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

// Package v1beta1 contains vendored CRD types originally from
// github.com/redhat-marketplace/redhat-marketplace-operator/v2/apis/marketplace/v1beta1.
// Only the types used by ibm-licensing-operator are included.
package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/IBM/ibm-licensing-operator/pkg/rhmp/common"
)

// +kubebuilder:object:root=true
type MeterDefinition struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MeterDefinitionSpec   `json:"spec,omitempty"`
	Status MeterDefinitionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type MeterDefinitionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MeterDefinition `json:"items"`
}

type MeterDefinitionSpec struct {
	Group           string                         `json:"group"`
	Kind            string                         `json:"kind"`
	ResourceFilters []ResourceFilter               `json:"resourceFilters"`
	Meters          []MeterWorkload                `json:"meters"`
	InstalledBy     *common.NamespacedNameReference `json:"installedBy,omitempty"`
}

// MeterDefinitionStatus is intentionally kept minimal; the operator never
// reads or writes status fields for this vendored type.
type MeterDefinitionStatus struct{}

type ResourceFilter struct {
	Namespace    *NamespaceFilter    `json:"namespace,omitempty"`
	OwnerCRD     *OwnerCRDFilter     `json:"ownerCRD,omitempty"`
	Label        *LabelFilter        `json:"label,omitempty"`
	Annotation   *AnnotationFilter   `json:"annotation,omitempty"`
	WorkloadType common.WorkloadType `json:"workloadType"`
}

type NamespaceFilter struct {
	UseOperatorGroup bool                  `json:"useOperatorGroup"`
	LabelSelector    *metav1.LabelSelector `json:"labelSelector,omitempty"`
}

type OwnerCRDFilter struct {
	common.GroupVersionKind `json:",inline"`
}

type LabelFilter struct {
	LabelSelector *metav1.LabelSelector `json:"labelSelector,omitempty"`
}

type AnnotationFilter struct {
	AnnotationSelector *metav1.LabelSelector `json:"annotationSelector,omitempty"`
}

type MeterWorkload struct {
	Metric             string              `json:"metricId"`
	Name               string              `json:"name,omitempty"`
	Description        string              `json:"description,omitempty"`
	WorkloadType       common.WorkloadType `json:"workloadType"`
	MetricType         common.MetricType   `json:"metricType,omitempty"`
	GroupBy            []string            `json:"groupBy,omitempty"`
	Without            []string            `json:"without,omitempty"`
	Aggregation        string              `json:"aggregation"`
	Period             *metav1.Duration    `json:"period,omitempty"`
	Query              string              `json:"query"`
	Label              string              `json:"label,omitempty"`
	Unit               string              `json:"unit,omitempty"`
	DateLabelOverride  string              `json:"dateLabelOverride,omitempty"`
	ValueLabelOverride string              `json:"valueLabelOverride,omitempty"`
}
