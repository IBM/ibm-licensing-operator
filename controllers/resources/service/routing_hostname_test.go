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

// TestGetLicensingRouteName_NoTruncationNeeded verifies that when the standard route name + namespace
// fits within the 63-char DNS label limit, the standard name "ibm-licensing-service-<instanceName>" is returned.
func TestGetLicensingRouteName_NoTruncationNeeded(t *testing.T) {
	instanceName := "instance"
	namespace := "ibm-licensing" // 31 + 1 + 13 = 45 <= 63

	result := GetLicensingRouteName(instanceName, namespace)

	assert.Equal(t, "ibm-licensing-service-instance", result)
	label := result + "-" + namespace
	assert.LessOrEqual(t, len(label), maxDNSLabelLength)
}

// TestGetLicensingRouteName_ShortenedPrefixApplied verifies the ILS-3012 reproduction case:
// namespace "prd-dti-infra-watsonx-ibm-licensing" (35 chars) with standard name (31 chars) gives 67 chars (> 63).
// It should switch to shortened prefix "ibm-ils-instance" (16 chars), giving 16 + 1 + 35 = 52 chars (<= 63).
func TestGetLicensingRouteName_ShortenedPrefixApplied(t *testing.T) {
	instanceName := "instance"
	namespace := "prd-dti-infra-watsonx-ibm-licensing"

	result := GetLicensingRouteName(instanceName, namespace)

	assert.Equal(t, "ibm-ils-instance", result)
	label := result + "-" + namespace
	assert.LessOrEqual(t, len(label), maxDNSLabelLength, "label must not exceed 63 characters")
	assert.Equal(t, 52, len(label))
}

// TestGetLicensingRouteName_Split3LongNamespace verifies the user's test cluster scenario:
// namespace "ibm-licensing-split3-very-long-name" (35 chars) with standard name gives 67 chars (> 63).
// It switches to "ibm-ils-instance" (16 chars), giving 16 + 1 + 35 = 52 chars (<= 63).
func TestGetLicensingRouteName_Split3LongNamespace(t *testing.T) {
	instanceName := "instance"
	namespace := "ibm-licensing-split3-very-long-name"

	result := GetLicensingRouteName(instanceName, namespace)

	assert.Equal(t, "ibm-ils-instance", result)
	label := result + "-" + namespace
	assert.LessOrEqual(t, len(label), maxDNSLabelLength, "label must not exceed 63 characters")
	assert.Equal(t, 52, len(label))
}

// TestGetLicensingRouteName_ExtremelyLongNamespace verifies that when namespace is exceptionally long
// (e.g. 50 chars), the route name is truncated further so label stays <= 63.
func TestGetLicensingRouteName_ExtremelyLongNamespace(t *testing.T) {
	instanceName := "instance"
	namespace := strings.Repeat("n", 50) // 50 chars

	result := GetLicensingRouteName(instanceName, namespace)

	label := result + "-" + namespace
	assert.LessOrEqual(t, len(label), maxDNSLabelLength, "resulting DNS label must not exceed 63 characters")
}

// TestGetLicensingRouteName_EmptyNamespace verifies that empty namespace produces standard name.
func TestGetLicensingRouteName_EmptyNamespace(t *testing.T) {
	instanceName := "instance"
	namespace := ""

	result := GetLicensingRouteName(instanceName, namespace)

	assert.Equal(t, "ibm-licensing-service-instance", result)
}
