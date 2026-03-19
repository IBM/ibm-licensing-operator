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

	operatorv1alpha1 "github.com/IBM/ibm-licensing-operator/api/v1alpha1"
)

func TestGetLicensingVolumeMountsReporterDisabled(t *testing.T) {
	spec := operatorv1alpha1.IBMLicensingSpec{
		InstanceNamespace: "namespace",
		Datasource:        "datacollector",
	}
	volumeMounts := getLicensingVolumeMounts(spec)
	assert.Equal(t, 3, len(volumeMounts), "Sender is disabled, only 3 mountVolumes should be created.")
}

func TestGetLicensingVolumeMountsReporterEnabled(t *testing.T) {
	spec := operatorv1alpha1.IBMLicensingSpec{
		InstanceNamespace: "namespace",
		Datasource:        "datacollector",
		Sender:            &operatorv1alpha1.IBMLicensingSenderSpec{},
	}

	volumeMounts := getLicensingVolumeMounts(spec)
	assert.Equal(t, 4, len(volumeMounts), "Sender is enabled, 4 mountVolumes should be created, one additional for reporter token.")

	reporterTokenVolumeMount := volumeMounts[3]
	assert.Equal(t, ReporterTokenVolumeName, reporterTokenVolumeMount.Name, "Sender volume mount should have correct name.")
}

func TestGetLicensingVolumeMountsCertsValidationEnabledWithCerts(t *testing.T) {
	spec := operatorv1alpha1.IBMLicensingSpec{
		InstanceNamespace: "namespace",
		Datasource:        "datacollector",
		Sender: &operatorv1alpha1.IBMLicensingSenderSpec{
			ReporterCertsSecretName: "some-cert-name",
			ValidateReporterCerts:   true,
		},
	}

	volumeMounts := getLicensingVolumeMounts(spec)
	assert.Equal(t, 5, len(volumeMounts), "Sender is enabled, 5 mountVolumes should be created, one additional for reporter cert secret name.")

	reporterCertsVolumeMount := volumeMounts[4]
	assert.Equal(t, ReporterHTTPSCertsVolumeName, reporterCertsVolumeMount.Name, "License service reporter certificate volume mount should have correct name.")
}

func TestGetLicensingVolumeMountsCertsValidationEnabledWithoutCerts(t *testing.T) {
	spec := operatorv1alpha1.IBMLicensingSpec{
		InstanceNamespace: "namespace",
		Datasource:        "datacollector",
		Sender: &operatorv1alpha1.IBMLicensingSenderSpec{
			ValidateReporterCerts: true,
		},
	}

	volumes := getLicensingVolumeMounts(spec)
	assert.Equal(t, 4, len(volumes), "Sender is enabled, 4 volume mounts should be created, ReporterCertsSecretName isn't set so additional volume mount is not created.")
}

func TestGetLicensingVolumeMountsCertsValidationDisabledWithCerts(t *testing.T) {
	spec := operatorv1alpha1.IBMLicensingSpec{
		InstanceNamespace: "namespace",
		Datasource:        "datacollector",
		Sender: &operatorv1alpha1.IBMLicensingSenderSpec{
			ReporterCertsSecretName: "some-cert-name",
			ValidateReporterCerts:   false,
		},
	}

	volumes := getLicensingVolumes(spec)
	assert.Equal(t, 4, len(volumes), "Sender is enabled, 4 volume mounts should be created despite setting ReporterCertsSecretName because validation is disabled.")
}

func TestGetLicensingVolumesDisabled(t *testing.T) {
	spec := operatorv1alpha1.IBMLicensingSpec{
		InstanceNamespace: "namespace",
		Datasource:        "datacollector",
	}

	volumes := getLicensingVolumes(spec)
	assert.Equal(t, 3, len(volumes), "Sender is disabled, only 3 volumes should be created.")
}

func TestGetLicensingVolumesEnabled(t *testing.T) {
	spec := operatorv1alpha1.IBMLicensingSpec{
		InstanceNamespace: "namespace",
		Datasource:        "datacollector",
		Sender:            &operatorv1alpha1.IBMLicensingSenderSpec{},
	}

	volumes := getLicensingVolumes(spec)
	assert.Equal(t, 4, len(volumes), "Sender is enabled, 4 volumes should be created, one additional for reporter token.")

	reporterTokenVolume := volumes[3]
	assert.Equal(t, ReporterTokenVolumeName, reporterTokenVolume.Name, "Sender reporter token volume should have correct name.")
	assert.Equal(t, spec.GetDefaultReporterTokenName(), reporterTokenVolume.Secret.SecretName, "Sender reporter token secret not specified, volume should have default name.")

	spec.Sender.ReporterSecretToken = "someSecretName"
	volumes = getLicensingVolumes(spec)
	reporterTokenVolume = volumes[3]
	assert.Equal(t, "someSecretName", reporterTokenVolume.Secret.SecretName, "Sender reporter token should have correct name that was set in CR.")
}

func TestGetLicensingVolumesCertsValidationEnabledWithCerts(t *testing.T) {
	spec := operatorv1alpha1.IBMLicensingSpec{
		InstanceNamespace: "namespace",
		Datasource:        "datacollector",
		Sender: &operatorv1alpha1.IBMLicensingSenderSpec{
			ReporterCertsSecretName: "some-cert-name",
			ValidateReporterCerts:   true,
		},
	}

	volumes := getLicensingVolumes(spec)
	assert.Equal(t, 5, len(volumes), "Sender is enabled, 5 volumes should be created, one additional for reporter cert secret name.")

	reporterCertsVolume := volumes[4]
	assert.Equal(t, ReporterHTTPSCertsVolumeName, reporterCertsVolume.Name, "Sender reporter certs volume should have correct name.")
	assert.Equal(t, spec.Sender.ReporterCertsSecretName, reporterCertsVolume.Secret.SecretName, "Sender reporter volume should have provided certificate secret mounted.")
}

func TestGetLicensingVolumesCertsValidationEnabledWithoutCerts(t *testing.T) {
	spec := operatorv1alpha1.IBMLicensingSpec{
		InstanceNamespace: "namespace",
		Datasource:        "datacollector",
		Sender: &operatorv1alpha1.IBMLicensingSenderSpec{
			ValidateReporterCerts: true,
		},
	}

	volumes := getLicensingVolumes(spec)
	assert.Equal(t, 4, len(volumes), "Sender is enabled, 4 volumes should be created, ReporterCertsSecretName isn't set so additional volume is not created.")
}

func TestGetLicensingVolumesCertsValidationDisabledWithCerts(t *testing.T) {
	spec := operatorv1alpha1.IBMLicensingSpec{
		InstanceNamespace: "namespace",
		Datasource:        "datacollector",
		Sender: &operatorv1alpha1.IBMLicensingSenderSpec{
			ReporterCertsSecretName: "some-cert-name",
			ValidateReporterCerts:   false,
		},
	}

	volumes := getLicensingVolumes(spec)
	assert.Equal(t, 4, len(volumes), "Sender is enabled, 4 volumes should be created despite setting ReporterCertsSecretName because validation is disabled.")
}

// verifies that when SoftwareCentral is enabled, an additional volume mount for the entitlement key is added.
func TestGetLicensingVolumeMountsSoftwareCentralEnabled(t *testing.T) {
	spec := operatorv1alpha1.IBMLicensingSpec{
		InstanceNamespace: "namespace",
		Datasource:        "datacollector",
		SoftwareCentral: &operatorv1alpha1.IBMLicensingSoftwareCentralSpec{
			Enable:               true,
			EntitlementKeySecret: "my-entitlement-secret",
		},
	}

	volumeMounts := getLicensingVolumeMounts(spec)
	assert.Equal(t, 4, len(volumeMounts), "SoftwareCentral is enabled, 4 volume mounts should be created (3 base + entitlement key).")

	swcVolumeMount := volumeMounts[3]
	assert.Equal(t, SoftwareCentralEntitlementKeyVolumeName, swcVolumeMount.Name,
		"Software Central entitlement key volume mount should have correct name.")
	assert.Equal(t, "/opt/ibm/licensing/swc-entitlement-key", swcVolumeMount.MountPath,
		"Software Central entitlement key volume mount should have correct mount path.")
	assert.True(t, swcVolumeMount.ReadOnly, "Software Central entitlement key volume mount should be read-only.")
}

// verifies that when SoftwareCentral is enabled, an additional volume for the entitlement key secret is added.
func TestGetLicensingVolumesSoftwareCentralEnabled(t *testing.T) {
	spec := operatorv1alpha1.IBMLicensingSpec{
		InstanceNamespace: "namespace",
		Datasource:        "datacollector",
		SoftwareCentral: &operatorv1alpha1.IBMLicensingSoftwareCentralSpec{
			Enable:               true,
			EntitlementKeySecret: "my-entitlement-secret",
		},
	}

	volumes := getLicensingVolumes(spec)
	assert.Equal(t, 4, len(volumes), "SoftwareCentral is enabled, 4 volumes should be created (3 base + entitlement key).")

	swcVolume := volumes[3]
	assert.Equal(t, SoftwareCentralEntitlementKeyVolumeName, swcVolume.Name,
		"Software Central entitlement key volume should have correct name.")
	assert.Equal(t, "my-entitlement-secret", swcVolume.Secret.SecretName,
		"Software Central entitlement key volume should reference the configured secret name.")
}
