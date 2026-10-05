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

/*
TestTruncateForDNSLabel_NoTruncationNeeded verifies that a namespace which already produces
a label within the 63-character limit is returned unchanged.
*/
func TestTruncateForDNSLabel_NoTruncationNeeded(t *testing.T) {
	routeName := "ibm-licensing-service-instance" // 30 chars
	namespace := "short-ns"                       // 8 chars -> label = 39 chars (well under 63)

	result, err := TruncateForDNSLabel(routeName, namespace)

	assert.NoError(t, err)
	assert.Equal(t, namespace, result, "namespace should be unchanged when label fits within 63 chars")
	label := routeName + "-" + result
	assert.LessOrEqual(t, len(label), maxDNSLabelLength,
		"resulting DNS label must not exceed %d characters", maxDNSLabelLength)
}

/*
TestTruncateForDNSLabel_ExactlyAtLimit verifies that a namespace producing a label of exactly
63 characters is returned unchanged (boundary: at limit, no truncation).
*/
func TestTruncateForDNSLabel_ExactlyAtLimit(t *testing.T) {
	routeName := "ibm-licensing-service-instance" // 30 chars
	// label = routeName(30) + "-"(1) + namespace = 63 -> namespace must be 32 chars
	namespace := strings.Repeat("a", maxDNSLabelLength-len(routeName)-1) // 32 chars

	result, err := TruncateForDNSLabel(routeName, namespace)

	assert.NoError(t, err)
	assert.Equal(t, namespace, result, "namespace should be unchanged when label is exactly 63 chars")
	label := routeName + "-" + result
	assert.Equal(t, maxDNSLabelLength, len(label),
		"resulting DNS label should be exactly %d characters", maxDNSLabelLength)
}

/*
TestTruncateForDNSLabel_TruncationApplied verifies the reproduction case where a long namespace
causes a DNS label that exceeds 63 characters. The namespace must be truncated to fit
within the RFC 1123 limit.
*/
func TestTruncateForDNSLabel_TruncationApplied(t *testing.T) {
	routeName := "ibm-licensing-service-instance" // 30 chars
	namespace := "prd-dti-infra-watsonx-ibm-licensing"

	result, err := TruncateForDNSLabel(routeName, namespace)

	assert.NoError(t, err)
	label := routeName + "-" + result
	assert.LessOrEqual(t, len(label), maxDNSLabelLength,
		"resulting DNS label must not exceed %d characters after truncation", maxDNSLabelLength)
	assert.Equal(t, maxDNSLabelLength-len(routeName)-1, len(result),
		"namespace should be truncated to exactly %d chars", maxDNSLabelLength-len(routeName)-1)
	assert.True(t, strings.HasPrefix(namespace, result),
		"truncated namespace should be a prefix of the original namespace")
}

/*
TestTruncateForDNSLabel_LongNamespace verifies that an arbitrarily long namespace is
truncated correctly regardless of how far it exceeds the limit.
*/
func TestTruncateForDNSLabel_LongNamespace(t *testing.T) {
	routeName := "ibm-licensing-service-instance" // 30 chars
	namespace := strings.Repeat("x", 100)

	result, err := TruncateForDNSLabel(routeName, namespace)

	assert.NoError(t, err)
	label := routeName + "-" + result
	assert.LessOrEqual(t, len(label), maxDNSLabelLength,
		"resulting DNS label must not exceed %d characters for very long namespace", maxDNSLabelLength)
}

/*
TestTruncateForDNSLabel_EmptyNamespace verifies that an empty namespace is returned as-is.
*/
func TestTruncateForDNSLabel_EmptyNamespace(t *testing.T) {
	routeName := "ibm-licensing-service-instance"
	namespace := ""

	result, err := TruncateForDNSLabel(routeName, namespace)

	assert.NoError(t, err)
	assert.Equal(t, "", result, "empty namespace should be returned unchanged")
}

/*
TestTruncateForDNSLabel_RouteNameTooLong verifies that an instance name long enough to
exhaust the 63-character DNS label budget on its own returns an error.
*/
func TestTruncateForDNSLabel_RouteNameTooLong(t *testing.T) {
	// 63 chars — fills the entire label budget leaving no room for "-" + namespace
	routeName := strings.Repeat("a", maxDNSLabelLength)
	namespace := "some-namespace"

	result, err := TruncateForDNSLabel(routeName, namespace)

	assert.Error(t, err, "should return an error when routeName exhausts the DNS label budget")
	assert.Empty(t, result)
}

/*
TestGetLicensingRouteHostname_NoTruncation verifies that a short namespace produces the
expected hostname without any truncation.
*/
func TestGetLicensingRouteHostname_NoTruncation(t *testing.T) {
	routeName := "ibm-licensing-service-instance"
	namespace := "ibm-licensing"
	appsDomain := "apps.example.com"

	hostname, err := GetLicensingRouteHostname(routeName, namespace, appsDomain)

	assert.NoError(t, err)
	expected := routeName + "-" + namespace + "." + appsDomain
	assert.Equal(t, expected, hostname)

	// First DNS label must be within limit
	label := strings.SplitN(hostname, ".", 2)[0]
	assert.LessOrEqual(t, len(label), maxDNSLabelLength,
		"first DNS label must not exceed %d characters", maxDNSLabelLength)
}

/*
TestGetLicensingRouteHostname_TruncationApplied verifies the end-to-end truncation case:
the resulting hostname must have a first DNS label of at most 63 characters.
*/
func TestGetLicensingRouteHostname_TruncationApplied(t *testing.T) {
	routeName := "ibm-licensing-service-instance"
	namespace := "prd-dti-infra-watsonx-ibm-licensing"
	appsDomain := "apps.fgv.br"

	hostname, err := GetLicensingRouteHostname(routeName, namespace, appsDomain)

	assert.NoError(t, err)
	parts := strings.SplitN(hostname, ".", 2)
	assert.Len(t, parts, 2, "hostname must contain at least one dot")
	label := parts[0]
	assert.LessOrEqual(t, len(label), maxDNSLabelLength,
		"first DNS label must not exceed %d characters", maxDNSLabelLength)

	assert.True(t, strings.HasSuffix(hostname, "."+appsDomain),
		"hostname must end with .%s", appsDomain)

	assert.True(t, strings.HasPrefix(label, routeName+"-"),
		"hostname label must start with the route name prefix")
}

/*
TestGetLicensingRouteHostname_Split3LongNamespace verifies that a long namespace used in
a known test environment is correctly truncated to produce a valid first DNS label.
*/
func TestGetLicensingRouteHostname_Split3LongNamespace(t *testing.T) {
	routeName := "ibm-licensing-service-instance"
	namespace := "ibm-licensing-split3-very-long-name"
	appsDomain := "apps.dpermus6.cp.fyre.ibm.com"

	hostname, err := GetLicensingRouteHostname(routeName, namespace, appsDomain)

	assert.NoError(t, err)
	parts := strings.SplitN(hostname, ".", 2)
	assert.Len(t, parts, 2)
	label := parts[0]
	assert.LessOrEqual(t, len(label), maxDNSLabelLength,
		"first DNS label must not exceed 63 characters")
	assert.Equal(t, maxDNSLabelLength, len(label))
	assert.True(t, strings.HasSuffix(hostname, "."+appsDomain))
}

/*
TestMaxNamespaceLengthForRouteProbe verifies that MaxNamespaceLengthForRouteProbe equals 61,
allowing a 1-character probe Route ("p") with hyphen ("p-<namespace>") to fit within the
63-character RFC 1123 DNS label limit (1 + 1 + 61 = 63).
*/
func TestMaxNamespaceLengthForRouteProbe(t *testing.T) {
	probeName := "p"
	maxAllowedProbeLabelLen := len(probeName) + 1 + MaxNamespaceLengthForRouteProbe
	assert.Equal(t, maxDNSLabelLength, maxAllowedProbeLabelLen, "probe route label with MaxNamespaceLengthForRouteProbe must equal 63 chars")
	assert.Equal(t, 61, MaxNamespaceLengthForRouteProbe)
}
