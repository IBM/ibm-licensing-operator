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