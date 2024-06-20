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

func TestGetLicensingVolumeMountsMeteringEnabled(t *testing.T) {
	spec := operatorv1alpha1.IBMLicensingSpec{
		InstanceNamespace: "namespace",
		Datasource:        "metering",
	}
	volumeMounts := getLicensingVolumeMounts(spec)
	assert.Equal(t, 4, len(volumeMounts), "Sender is disabled, only 3 mountVolumes should be created.")

	meteringAPICertsVolumeMount := volumeMounts[3]
	assert.Equal(t, MeteringAPICertsVolumeName, meteringAPICertsVolumeMount.Name, "Metering volume mount should have correct name.")
}

func TestGetLicensingVolumeMountsCertsValidationEnabledDefaultCert(t *testing.T) {
	spec := operatorv1alpha1.IBMLicensingSpec{
		InstanceNamespace: "namespace",
		Datasource:        "datacollector",
		Sender: &operatorv1alpha1.IBMLicensingSenderSpec{
			ValidateReporterCerts: true,
		},
	}

	volumes := getLicensingVolumeMounts(spec)
	assert.Equal(t, 4, len(volumes), "Sender with certificate validation is enabled, 4 volume mounts should be created, ReporterCertsSecretName so we will mount default OCP cert manager secret.")
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
	assert.Equal(t, 4, len(volumeMounts), "Sender with certificate validation is enabled, 4 mountVolumes should be created, one additional for reporter cert secret name.")

	reporterCertsVolumeMount := volumeMounts[3]
	assert.Equal(t, ReporterHTTPSCertsVolumeName, reporterCertsVolumeMount.Name, "License service reporter certificate volume mount should have correct name.")
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
	assert.Equal(t, 3, len(volumes), "Sender certificate validation is disabled, 3 volume mounts should be created despite setting ReporterCertsSecretName because validation is disabled.")
}

func TestGetLicensingVolumesDisabled(t *testing.T) {
	spec := operatorv1alpha1.IBMLicensingSpec{
		InstanceNamespace: "namespace",
		Datasource:        "datacollector",
	}

	volumes := getLicensingVolumes(spec)
	assert.Equal(t, 3, len(volumes), "Sender is disabled, only 3 volumes should be created.")
}

func TestGetLicensingVolumesCertsValidationEnabledDefaultCerts(t *testing.T) {
	spec := operatorv1alpha1.IBMLicensingSpec{
		InstanceNamespace: "namespace",
		Datasource:        "datacollector",
		Sender: &operatorv1alpha1.IBMLicensingSenderSpec{
			ValidateReporterCerts: true,
		},
	}

	volumes := getLicensingVolumes(spec)
	assert.Equal(t, 4, len(volumes), "Sender with certificate validation is enabled, 4 volumes should be created, ReporterCertsSecretName so we will create default OCP cert manager secret's volume.")
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
	assert.Equal(t, 4, len(volumes), "Sender certificate validation is enabled, 4 volumes should be created, one additional for reporter cert secret name.")

	reporterCertsVolume := volumes[3]
	assert.Equal(t, ReporterHTTPSCertsVolumeName, reporterCertsVolume.Name, "Sender reporter certs volume should have correct name.")
	assert.Equal(t, spec.Sender.ReporterCertsSecretName, reporterCertsVolume.Secret.SecretName, "Sender reporter volume should have provided certificate secret mounted.")
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
	assert.Equal(t, 3, len(volumes), "Sender is enabled, 3 volumes should be created despite setting ReporterCertsSecretName because validation is disabled.")
}
