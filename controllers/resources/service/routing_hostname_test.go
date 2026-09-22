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

package service

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestTruncateForDNSLabel_NoTruncationNeeded verifies that a namespace which already produces
// a label within the 63-character limit is returned unchanged.
func TestTruncateForDNSLabel_NoTruncationNeeded(t *testing.T) {
	routeName := "ibm-licensing-service-instance" // 30 chars
	namespace := "short-ns"                        // 8 chars → label = 39 chars (well under 63)

	result := TruncateForDNSLabel(routeName, namespace)

	assert.Equal(t, namespace, result, "namespace should be unchanged when label fits within 63 chars")
	label := routeName + "-" + result
	assert.LessOrEqual(t, len(label), maxDNSLabelLength,
		"resulting DNS label must not exceed %d characters", maxDNSLabelLength)
}

// TestTruncateForDNSLabel_ExactlyAtLimit verifies that a namespace producing a label of exactly
// 63 characters is returned unchanged (boundary: at limit, no truncation).
func TestTruncateForDNSLabel_ExactlyAtLimit(t *testing.T) {
	routeName := "ibm-licensing-service-instance" // 30 chars
	// label = routeName(30) + "-"(1) + namespace = 63 → namespace must be 32 chars
	namespace := strings.Repeat("a", maxDNSLabelLength-len(routeName)-1) // 32 chars

	result := TruncateForDNSLabel(routeName, namespace)

	assert.Equal(t, namespace, result, "namespace should be unchanged when label is exactly 63 chars")
	label := routeName + "-" + result
	assert.Equal(t, maxDNSLabelLength, len(label),
		"resulting DNS label should be exactly %d characters", maxDNSLabelLength)
}

// TestTruncateForDNSLabel_TruncationApplied verifies the ILS-3012 reproduction case:
// namespace "prd-dti-infra-watsonx-ibm-licensing" (35 chars) causes a 66-char label;
// the namespace must be truncated to 32 chars.
func TestTruncateForDNSLabel_TruncationApplied(t *testing.T) {
	routeName := "ibm-licensing-service-instance" // 30 chars
	namespace := "prd-dti-infra-watsonx-ibm-licensing"

	result := TruncateForDNSLabel(routeName, namespace)

	label := routeName + "-" + result
	assert.LessOrEqual(t, len(label), maxDNSLabelLength,
		"resulting DNS label must not exceed %d characters after truncation", maxDNSLabelLength)
	assert.Equal(t, maxDNSLabelLength-len(routeName)-1, len(result),
		"namespace should be truncated to exactly %d chars", maxDNSLabelLength-len(routeName)-1)
	assert.True(t, strings.HasPrefix(namespace, result),
		"truncated namespace should be a prefix of the original namespace")
}

// TestTruncateForDNSLabel_LongNamespace verifies that an arbitrarily long namespace is
// truncated correctly regardless of how far it exceeds the limit.
func TestTruncateForDNSLabel_LongNamespace(t *testing.T) {
	routeName := "ibm-licensing-service-instance" // 30 chars
	namespace := strings.Repeat("x", 100)

	result := TruncateForDNSLabel(routeName, namespace)

	label := routeName + "-" + result
	assert.LessOrEqual(t, len(label), maxDNSLabelLength,
		"resulting DNS label must not exceed %d characters for very long namespace", maxDNSLabelLength)
}

// TestTruncateForDNSLabel_EmptyNamespace verifies that an empty namespace is returned as-is.
func TestTruncateForDNSLabel_EmptyNamespace(t *testing.T) {
	routeName := "ibm-licensing-service-instance"
	namespace := ""

	result := TruncateForDNSLabel(routeName, namespace)

	assert.Equal(t, "", result, "empty namespace should be returned unchanged")
}

// TestGetLicensingRouteHostname_NoTruncation verifies that a short namespace produces the
// expected hostname without any truncation.
func TestGetLicensingRouteHostname_NoTruncation(t *testing.T) {
	routeName := "ibm-licensing-service-instance"
	namespace := "ibm-licensing"
	appsDomain := "apps.example.com"

	hostname := GetLicensingRouteHostname(routeName, namespace, appsDomain)

	expected := routeName + "-" + namespace + "." + appsDomain
	assert.Equal(t, expected, hostname)

	// First DNS label must be within limit
	label := strings.SplitN(hostname, ".", 2)[0]
	assert.LessOrEqual(t, len(label), maxDNSLabelLength,
		"first DNS label must not exceed %d characters", maxDNSLabelLength)
}

// TestGetLicensingRouteHostname_TruncationApplied verifies the ILS-3012 reproduction case end-to-end:
// the resulting hostname must have a first DNS label ≤ 63 characters.
func TestGetLicensingRouteHostname_TruncationApplied(t *testing.T) {
	routeName := "ibm-licensing-service-instance"
	namespace := "prd-dti-infra-watsonx-ibm-licensing"
	appsDomain := "apps.fgv.br"

	hostname := GetLicensingRouteHostname(routeName, namespace, appsDomain)

	// Split on first dot to get the DNS label
	parts := strings.SplitN(hostname, ".", 2)
	assert.Len(t, parts, 2, "hostname must contain at least one dot")
	label := parts[0]
	assert.LessOrEqual(t, len(label), maxDNSLabelLength,
		"first DNS label must not exceed %d characters (ILS-3012 regression)", maxDNSLabelLength)

	// Domain suffix must be preserved intact
	assert.True(t, strings.HasSuffix(hostname, "."+appsDomain),
		"hostname must end with .%s", appsDomain)

	// Route name prefix must be preserved intact
	assert.True(t, strings.HasPrefix(label, routeName+"-"),
		"hostname label must start with the route name prefix")
}

// TestGetLicensingRouteHostname_ExactLimit verifies that a namespace that produces exactly a
// 63-char label is passed through unchanged.
func TestGetLicensingRouteHostname_ExactLimit(t *testing.T) {
	routeName := "ibm-licensing-service-instance" // 30 chars
	namespace := strings.Repeat("b", 32)           // 30+1+32=63 exactly
	appsDomain := "apps.cluster.local"

	hostname := GetLicensingRouteHostname(routeName, namespace, appsDomain)

	label := strings.SplitN(hostname, ".", 2)[0]
	assert.Equal(t, maxDNSLabelLength, len(label),
		"first DNS label should be exactly %d chars at the boundary", maxDNSLabelLength)
	assert.Equal(t, routeName+"-"+namespace+"."+appsDomain, hostname,
		"hostname should be unmodified when namespace exactly fits the limit")
}
