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

package features

// +k8s:deepcopy-gen=true
type NamespaceDiscovery struct {
	// Enables cluster-wide namespace discovery on the operand. When true or
	// nil, the operand discovers and monitors every namespace on the cluster
	// (today's behavior). When false, the operand monitors only the namespaces
	// listed in Namespaces and issues no cluster-wide list calls.
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// Explicit list of namespaces the operand should monitor. Takes effect only
	// when Enabled is false. Ignored (with an info log) when discovery is on.
	// +optional
	Namespaces []string `json:"namespaces,omitempty"`
}
