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

func TestGetWatchedNamespacesEmpty(t *testing.T) {
	spec := &IBMLicensingSpec{}
	assert.Nil(t, spec.GetWatchedNamespaces(),
		"Empty watchedNamespaces should yield no namespaces.")
}

func TestGetWatchedNamespacesSingle(t *testing.T) {
	spec := &IBMLicensingSpec{WatchedNamespaces: "ns-a"}
	assert.Equal(t, []string{"ns-a"}, spec.GetWatchedNamespaces(),
		"A single namespace should be parsed into a one-element slice.")
}

func TestGetWatchedNamespacesMultiple(t *testing.T) {
	spec := &IBMLicensingSpec{WatchedNamespaces: "ns-a,ns-b,ns-c"}
	assert.Equal(t, []string{"ns-a", "ns-b", "ns-c"}, spec.GetWatchedNamespaces(),
		"A comma-separated list should be split into its elements.")
}

func TestGetWatchedNamespacesTrimsAndDropsEmpty(t *testing.T) {
	spec := &IBMLicensingSpec{WatchedNamespaces: " ns-a , ,ns-b ,"}
	assert.Equal(t, []string{"ns-a", "ns-b"}, spec.GetWatchedNamespaces(),
		"Whitespace should be trimmed and empty entries dropped.")
}
