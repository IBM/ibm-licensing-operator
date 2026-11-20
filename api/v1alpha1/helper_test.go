//
// Copyright 2026 IBM Corporation
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

package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func TestIsNodeCpuCappingEnabledNil(t *testing.T) {
	spec := &IBMLicensingSpec{}
	assert.True(t, spec.IsNodeCpuCappingEnabled(),
		"NodeCpuCappingEnabled unset should default to true.")
}

func TestIsNodeCpuCappingEnabledExplicitTrue(t *testing.T) {
	spec := &IBMLicensingSpec{Features: &Features{NodeCpuCappingEnabled: new(true)}}
	assert.True(t, spec.IsNodeCpuCappingEnabled(),
		"NodeCpuCappingEnabled=true should return true.")
}

func TestIsNodeCpuCappingEnabledExplicitFalse(t *testing.T) {
	spec := &IBMLicensingSpec{Features: &Features{NodeCpuCappingEnabled: new(false)}}
	assert.False(t, spec.IsNodeCpuCappingEnabled(),
		"NodeCpuCappingEnabled=false should return false.")
}

func TestGetSanitizedExcludeNamespaceNoFeatures(t *testing.T) {
	spec := &IBMLicensingSpec{}
	assert.Equal(t, "", spec.GetSanitizedExcludeNamespace(),
		"No features block should return empty string.")
}

func TestGetSanitizedExcludeNamespaceEmptyString(t *testing.T) {
	spec := &IBMLicensingSpec{Features: &Features{ExcludeNamespace: ""}}
	assert.Equal(t, "", spec.GetSanitizedExcludeNamespace(),
		"Empty ExcludeNamespace should return empty string.")
}

func TestGetSanitizedExcludeNamespaceSingleNamespace(t *testing.T) {
	spec := &IBMLicensingSpec{Features: &Features{ExcludeNamespace: "namespace-a"}}
	assert.Equal(t, "namespace-a", spec.GetSanitizedExcludeNamespace(),
		"Single namespace should be returned as-is.")
}

func TestGetSanitizedExcludeNamespaceMultipleNamespaces(t *testing.T) {
	spec := &IBMLicensingSpec{Features: &Features{ExcludeNamespace: "namespace-a,namespace-b,namespace-c"}}
	assert.Equal(t, "namespace-a,namespace-b,namespace-c", spec.GetSanitizedExcludeNamespace(),
		"Multiple distinct namespaces should all be returned.")
}

func TestGetSanitizedExcludeNamespaceTrimsWhitespace(t *testing.T) {
	spec := &IBMLicensingSpec{Features: &Features{ExcludeNamespace: " namespace-a , namespace-b , namespace-c "}}
	assert.Equal(t, "namespace-a,namespace-b,namespace-c", spec.GetSanitizedExcludeNamespace(),
		"Whitespace around namespace names should be trimmed.")
}

func TestGetSanitizedExcludeNamespaceRemovesDuplicates(t *testing.T) {
	spec := &IBMLicensingSpec{Features: &Features{ExcludeNamespace: "namespace-a,namespace-b,namespace-a"}}
	assert.Equal(t, "namespace-a,namespace-b", spec.GetSanitizedExcludeNamespace(),
		"Duplicate namespaces should be removed, preserving first-seen order.")
}

func TestGetSanitizedExcludeNamespaceRemovesEmptyEntries(t *testing.T) {
	spec := &IBMLicensingSpec{Features: &Features{ExcludeNamespace: "namespace-a,,namespace-b,"}}
	assert.Equal(t, "namespace-a,namespace-b", spec.GetSanitizedExcludeNamespace(),
		"Empty entries from consecutive commas or trailing comma should be removed.")
}

func TestGetSanitizedExcludeNamespaceRegexPattern(t *testing.T) {
	spec := &IBMLicensingSpec{Features: &Features{ExcludeNamespace: "products-[a-zA-Z]+,database-[0-9]+"}}
	assert.Equal(t, "products-[a-zA-Z]+,database-[0-9]+", spec.GetSanitizedExcludeNamespace(),
		"Regex patterns should be preserved as-is.")
}

func TestGetSanitizedExcludeNamespaceWhitespaceOnlyEntry(t *testing.T) {
	spec := &IBMLicensingSpec{Features: &Features{ExcludeNamespace: "namespace-a,   ,namespace-b"}}
	assert.Equal(t, "namespace-a,namespace-b", spec.GetSanitizedExcludeNamespace(),
		"Whitespace-only entries should be treated as empty and dropped.")
}

func TestGetTolerationsNilReceiver(t *testing.T) {
	var config *OperandPodConfig
	assert.Nil(t, config.GetTolerations(), "GetTolerations on a nil OperandPodConfig should return nil.")
}

func TestGetTolerations(t *testing.T) {
	tolerations := []corev1.Toleration{
		{Key: "dedicated", Operator: corev1.TolerationOpExists, Effect: corev1.TaintEffectNoSchedule},
	}

	tests := []struct {
		name   string
		config *OperandPodConfig
		want   []corev1.Toleration
	}{
		{"unset tolerations", &OperandPodConfig{}, nil},
		{"empty tolerations", &OperandPodConfig{Tolerations: []corev1.Toleration{}}, []corev1.Toleration{}},
		{"set tolerations are returned unchanged", &OperandPodConfig{Tolerations: tolerations}, tolerations},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.config.GetTolerations())
		})
	}
}

func TestValidateTolerations(t *testing.T) {
	seconds := int64(30)
	tests := []struct {
		name        string
		config      *OperandPodConfig
		wantErrPart string
	}{
		{"nil config", nil, ""},
		{"unset tolerations", &OperandPodConfig{}, ""},
		{"Exists with effect", &OperandPodConfig{Tolerations: []corev1.Toleration{
			{Key: "dedicated", Operator: corev1.TolerationOpExists, Effect: corev1.TaintEffectNoSchedule}}}, ""},
		{"Equal with value", &OperandPodConfig{Tolerations: []corev1.Toleration{
			{Key: "example.com/key", Operator: corev1.TolerationOpEqual, Value: "v1", Effect: corev1.TaintEffectNoExecute}}}, ""},
		{"empty operator defaults to Equal", &OperandPodConfig{Tolerations: []corev1.Toleration{
			{Key: "k", Value: "v"}}}, ""},
		{"Exists without key tolerates everything", &OperandPodConfig{Tolerations: []corev1.Toleration{
			{Operator: corev1.TolerationOpExists}}}, ""},
		{"tolerationSeconds with NoExecute", &OperandPodConfig{Tolerations: []corev1.Toleration{
			{Key: "k", Operator: corev1.TolerationOpExists, Effect: corev1.TaintEffectNoExecute, TolerationSeconds: &seconds}}}, ""},
		{"invalid operator", &OperandPodConfig{Tolerations: []corev1.Toleration{
			{Key: "bad", Operator: "Foo", Effect: corev1.TaintEffectNoSchedule}}}, "tolerations[0].operator: unsupported value \"Foo\""},
		{"invalid effect", &OperandPodConfig{Tolerations: []corev1.Toleration{
			{Key: "bad", Operator: corev1.TolerationOpExists, Effect: "Foo"}}}, "tolerations[0].effect: unsupported value \"Foo\""},
		{"Equal without key", &OperandPodConfig{Tolerations: []corev1.Toleration{
			{Operator: corev1.TolerationOpEqual, Value: "v"}}}, "operator must be Exists when key is empty"},
		{"Exists with value", &OperandPodConfig{Tolerations: []corev1.Toleration{
			{Key: "k", Operator: corev1.TolerationOpExists, Value: "v"}}}, "value must be empty when operator is Exists"},
		{"invalid key", &OperandPodConfig{Tolerations: []corev1.Toleration{
			{Key: "not a valid key!", Operator: corev1.TolerationOpExists}}}, "tolerations[0].key"},
		{"invalid value", &OperandPodConfig{Tolerations: []corev1.Toleration{
			{Key: "k", Operator: corev1.TolerationOpEqual, Value: "not valid!"}}}, "tolerations[0].value"},
		{"tolerationSeconds without NoExecute", &OperandPodConfig{Tolerations: []corev1.Toleration{
			{Key: "k", Operator: corev1.TolerationOpExists, Effect: corev1.TaintEffectNoSchedule, TolerationSeconds: &seconds}}},
			"effect must be NoExecute when tolerationSeconds is set"},
		{"error points at the failing index", &OperandPodConfig{Tolerations: []corev1.Toleration{
			{Key: "ok", Operator: corev1.TolerationOpExists},
			{Key: "bad", Operator: "Foo"}}}, "tolerations[1].operator"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.ValidateTolerations()
			if tt.wantErrPart == "" {
				assert.NoError(t, err)
				return
			}
			assert.ErrorContains(t, err, tt.wantErrPart)
		})
	}
}
