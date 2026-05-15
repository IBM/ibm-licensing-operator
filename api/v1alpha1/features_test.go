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

	"github.com/IBM/ibm-licensing-operator/api/v1alpha1/features"
)

func TestIsNamespaceDiscoveryEnabledFeaturesNil(t *testing.T) {
	spec := &IBMLicensingSpec{}
	assert.True(t, spec.IsNamespaceDiscoveryEnabled(),
		"Features unset should default namespace discovery to enabled.")
}

func TestIsNamespaceDiscoveryEnabledBlockNil(t *testing.T) {
	spec := &IBMLicensingSpec{Features: &Features{}}
	assert.True(t, spec.IsNamespaceDiscoveryEnabled(),
		"namespaceDiscovery block unset should default discovery to enabled.")
}

func TestIsNamespaceDiscoveryEnabledPointerNil(t *testing.T) {
	spec := &IBMLicensingSpec{Features: &Features{NamespaceDiscovery: &features.NamespaceDiscovery{}}}
	assert.True(t, spec.IsNamespaceDiscoveryEnabled(),
		"namespaceDiscovery.enabled unset should default discovery to enabled.")
}

func TestIsNamespaceDiscoveryEnabledExplicitTrue(t *testing.T) {
	spec := &IBMLicensingSpec{Features: &Features{NamespaceDiscovery: &features.NamespaceDiscovery{Enabled: new(true)}}}
	assert.True(t, spec.IsNamespaceDiscoveryEnabled(),
		"namespaceDiscovery.enabled=true should return true.")
}

func TestIsNamespaceDiscoveryEnabledExplicitFalse(t *testing.T) {
	spec := &IBMLicensingSpec{Features: &Features{NamespaceDiscovery: &features.NamespaceDiscovery{Enabled: new(false)}}}
	assert.False(t, spec.IsNamespaceDiscoveryEnabled(),
		"namespaceDiscovery.enabled=false should return false.")
}

func TestGetDiscoveryNamespacesFeaturesNil(t *testing.T) {
	spec := &IBMLicensingSpec{}
	assert.Nil(t, spec.GetDiscoveryNamespaces(),
		"Features unset should yield no discovery namespaces.")
}

func TestGetDiscoveryNamespacesBlockNil(t *testing.T) {
	spec := &IBMLicensingSpec{Features: &Features{}}
	assert.Nil(t, spec.GetDiscoveryNamespaces(),
		"namespaceDiscovery block unset should yield no discovery namespaces.")
}

func TestGetDiscoveryNamespacesSet(t *testing.T) {
	spec := &IBMLicensingSpec{Features: &Features{NamespaceDiscovery: &features.NamespaceDiscovery{
		Namespaces: []string{"ns-a", "ns-b"},
	}}}
	assert.Equal(t, []string{"ns-a", "ns-b"}, spec.GetDiscoveryNamespaces(),
		"GetDiscoveryNamespaces should return the configured list.")
}
