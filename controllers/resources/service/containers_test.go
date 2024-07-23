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

func Contains[T comparable](s []T, e T) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}
