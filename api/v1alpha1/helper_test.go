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
