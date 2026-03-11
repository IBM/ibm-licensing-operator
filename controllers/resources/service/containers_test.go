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

package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"

	operatorv1alpha1 "github.com/IBM/ibm-licensing-operator/api/v1alpha1"
)

func TestGetLicensingEnvironmentVariablesCertsValidationDisabledWithCerts(t *testing.T) {
	spec := operatorv1alpha1.IBMLicensingSpec{
		InstanceNamespace: "namespace",
		Datasource:        "datacollector",
		Sender: &operatorv1alpha1.IBMLicensingSenderSpec{
			ReporterCertsSecretName: "some-cert-name",
			ValidateReporterCerts:   false,
		},
	}

	validateReporterCertsEnv := corev1.EnvVar{
		Name:  "VALIDATE_REPORTER_CERTS",
		Value: "true",
	}

	envVars := getLicensingEnvironmentVariables(spec)
	assert.False(t, Contains(envVars, validateReporterCertsEnv), "Sender ValidateReporterCerts is disabled, 'VALIDATE_REPORTER_CERTS' environemnt variable should not be added to Licensing pod.")
}

func TestGetLicensingEnvironmentVariablesCertsValidationEnabledWithCerts(t *testing.T) {
	spec := operatorv1alpha1.IBMLicensingSpec{
		InstanceNamespace: "namespace",
		Datasource:        "datacollector",
		Sender: &operatorv1alpha1.IBMLicensingSenderSpec{
			ReporterCertsSecretName: "some-cert-name",
			ValidateReporterCerts:   true,
		},
	}

	validateReporterCertsEnv := corev1.EnvVar{
		Name:  "VALIDATE_REPORTER_CERTS",
		Value: "true",
	}

	envVars := getLicensingEnvironmentVariables(spec)
	assert.True(t, Contains(envVars, validateReporterCertsEnv), "Sender ValidateReporterCerts is enabled, appropriate 'VALIDATE_REPORTER_CERTS' environemnt variable should be added to Licensing pod.")
}

func TestGetLicensingEnvironmentVariablesNamespaceScopingFeatureEnabled(t *testing.T) {
	spec := operatorv1alpha1.IBMLicensingSpec{
		InstanceNamespace: "namespace",
		Datasource:        "datacollector",
		Features: &operatorv1alpha1.Features{
			NamespaceScopeEnabled:     ptr.To(true),
			NamespaceScopeDenialLimit: 10,
		},
	}

	featureEnabledEnvVar := corev1.EnvVar{
		Name:  "NAMESPACE_SCOPE_ENABLED",
		Value: "true",
	}
	watchNamespacesEnvVar := corev1.EnvVar{
		Name:  "WATCH_NAMESPACE",
		Value: "ibm-licensing",
	}
	namespaceScopeDenialLimitEnvVar := corev1.EnvVar{
		Name:  "NAMESPACE_DENIAL_LIMIT",
		Value: "10",
	}

	envVars := getLicensingEnvironmentVariables(spec)
	assert.True(t, Contains(envVars, featureEnabledEnvVar), "Namespaces scoping feature is enabled, environemnt variable 'NAMESPACE_SCOPE_ENABLED' set to true should be added to Licensing pod.")
	assert.True(t, Contains(envVars, watchNamespacesEnvVar), "Namespaces scoping feature is enabled, appropriate 'WATCH_NAMESPACE' environemnt variable should be added to Licensing pod.")
	assert.True(t, Contains(envVars, namespaceScopeDenialLimitEnvVar), "Namespaces scoping feature is enabled, appropriate 'NAMESPACE_DENIAL_LIMIT' environemnt variable should be added to Licensing pod.")
}

// verifies that when SoftwareCentral is not configured, none of the SOFTWARE_CENTRAL_* environment variables are added to the Licensing pod.
func TestGetLicensingEnvironmentVariablesSoftwareCentralDisabled_NilSpec(t *testing.T) {
	spec := operatorv1alpha1.IBMLicensingSpec{
		InstanceNamespace: "namespace",
		Datasource:        "datacollector",
		SoftwareCentral:   nil,
	}

	envVars := getLicensingEnvironmentVariables(spec)
	assert.False(t, Contains(envVars, corev1.EnvVar{Name: "SOFTWARE_CENTRAL_ENABLED", Value: "false"}),
		"SoftwareCentral is nil, SOFTWARE_CENTRAL_ENABLED should not be added to Licensing pod.")
	assert.False(t, Contains(envVars, corev1.EnvVar{Name: "SOFTWARE_CENTRAL_URL", Value: softwareCentralProductionURL}),
		"SoftwareCentral is nil, SOFTWARE_CENTRAL_URL should not be added to Licensing pod.")
	assert.False(t, Contains(envVars, corev1.EnvVar{Name: "SOFTWARE_CENTRAL_FREQUENCY", Value: softwareCentralDefaultFrequency}),
		"SoftwareCentral is nil, SOFTWARE_CENTRAL_FREQUENCY should not be added to Licensing pod.")
}

// verifies that when SoftwareCentral is enabled, SOFTWARE_CENTRAL_ENABLED, SOFTWARE_CENTRAL_URL and SOFTWARE_CENTRAL_FREQUENCY environment variables are added.
func TestGetLicensingEnvironmentVariablesSoftwareCentralEnabled(t *testing.T) {
	spec := operatorv1alpha1.IBMLicensingSpec{
		InstanceNamespace: "namespace",
		Datasource:        "datacollector",
		SoftwareCentral: &operatorv1alpha1.IBMLicensingSoftwareCentralSpec{
			Enable: true,
		},
	}

	envVars := getLicensingEnvironmentVariables(spec)
	assert.True(t, Contains(envVars, corev1.EnvVar{Name: "SOFTWARE_CENTRAL_ENABLED", Value: "true"}),
		"SoftwareCentral is enabled, SOFTWARE_CENTRAL_ENABLED=true should be added to Licensing pod.")
	assert.True(t, Contains(envVars, corev1.EnvVar{Name: "SOFTWARE_CENTRAL_URL", Value: softwareCentralProductionURL}),
		"SoftwareCentral is enabled with Sandbox=false, SOFTWARE_CENTRAL_URL should point to production URL.")
	assert.True(t, Contains(envVars, corev1.EnvVar{Name: "SOFTWARE_CENTRAL_FREQUENCY", Value: softwareCentralDefaultFrequency}),
		"SoftwareCentral is enabled, SOFTWARE_CENTRAL_FREQUENCY should be set to the provided value.")
}

// verifies that when SoftwareCentral is enabled with Sandbox=true, the SOFTWARE_CENTRAL_URL environment variable points to the sandbox URL.
func TestGetLicensingEnvironmentVariablesSoftwareCentralEnabled_SandboxTrue(t *testing.T) {
	spec := operatorv1alpha1.IBMLicensingSpec{
		InstanceNamespace: "namespace",
		Datasource:        "datacollector",
		SoftwareCentral: &operatorv1alpha1.IBMLicensingSoftwareCentralSpec{
			Enable:  true,
			Sandbox: true,
		},
	}

	envVars := getLicensingEnvironmentVariables(spec)
	assert.True(t, Contains(envVars, corev1.EnvVar{Name: "SOFTWARE_CENTRAL_URL", Value: softwareCentralSandboxURL}),
		"SoftwareCentral.Sandbox is true, SOFTWARE_CENTRAL_URL should point to the sandbox URL.")
}

// verifies that when SoftwareCentral is enabled with a custom cron frequency, the SOFTWARE_CENTRAL_FREQUENCY environment variable reflects
// the custom value rather than the default.
func TestGetLicensingEnvironmentVariablesSoftwareCentralEnabled_WithCustomFrequency(t *testing.T) {
	spec := operatorv1alpha1.IBMLicensingSpec{
		InstanceNamespace: "namespace",
		Datasource:        "datacollector",
		SoftwareCentral: &operatorv1alpha1.IBMLicensingSoftwareCentralSpec{
			Enable:    true,
			Frequency: "0 12 * * *",
		},
	}

	envVars := getLicensingEnvironmentVariables(spec)
	assert.True(t, Contains(envVars, corev1.EnvVar{Name: "SOFTWARE_CENTRAL_FREQUENCY", Value: "0 12 * * *"}),
		"SoftwareCentral is enabled with custom frequency, SOFTWARE_CENTRAL_FREQUENCY should reflect the custom value.")
}

func TestGetSoftwareCentralFrequencyDefaultValue(t *testing.T) {
	// Test default frequency
	softwareCentralSpec := &operatorv1alpha1.IBMLicensingSoftwareCentralSpec{}
	assert.Equal(t, "0 5 0 * * *", getSoftwareCentralFrequency(softwareCentralSpec))

	// Test frequency with 5 characters
	softwareCentralSpec = &operatorv1alpha1.IBMLicensingSoftwareCentralSpec{
		Frequency: "* * * * *",
	}
	assert.Equal(t, "0 * * * * *", getSoftwareCentralFrequency(softwareCentralSpec))

	// Test frequency with 6 characters
	softwareCentralSpec = &operatorv1alpha1.IBMLicensingSoftwareCentralSpec{
		Frequency: "*/20 * * * * *",
	}
	assert.Equal(t, "*/20 * * * * *", getSoftwareCentralFrequency(softwareCentralSpec))

}

func Contains[T comparable](s []T, e T) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}
