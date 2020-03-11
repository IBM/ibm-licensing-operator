//
// Copyright 2020 IBM Corporation
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

package resources

import (
	operatorv1alpha1 "github.com/ibm/ibm-licensing-operator/pkg/apis/operator/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

const APISecretTokenVolumeName = "api-token"
const MeteringAPICertsVolumeName = "metering-api-certs"
const LicensingHTTPSCertsVolumeName = "licensing-https-certs"

func getLicensingVolumeMounts(spec operatorv1alpha1.IBMLicensingSpec) []corev1.VolumeMount {
	var volumeMounts = []corev1.VolumeMount{
		{
			Name:      APISecretTokenVolumeName,
			MountPath: "/opt/ibm/licensing",
			ReadOnly:  true,
		},
	}
	if spec.HTTPSEnable {
		if spec.HTTPSCertsSource == "custom" {
			volumeMounts = append(volumeMounts, []corev1.VolumeMount{
				{
					Name:      LicensingHTTPSCertsVolumeName,
					MountPath: "/opt/licensing/certs/",
					ReadOnly:  true,
				},
			}...)
		}
	}
	if spec.IsMetering() {
		volumeMounts = append(volumeMounts, []corev1.VolumeMount{
			{
				Name:      MeteringAPICertsVolumeName,
				MountPath: "/opt/metering/certs/",
				ReadOnly:  true,
			},
		}...)
	}
	return volumeMounts
}

func getLicensingVolumes(spec operatorv1alpha1.IBMLicensingSpec) []corev1.Volume {
	volumes := []corev1.Volume{}

	apiSecretTokenVolume := corev1.Volume{
		Name: APISecretTokenVolumeName,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName:  spec.APISecretToken,
				DefaultMode: &defaultSecretMode,
			},
		},
	}

	volumes = append(volumes, apiSecretTokenVolume)

	if spec.IsMetering() {
		meteringAPICertVolume := corev1.Volume{
			Name: MeteringAPICertsVolumeName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  "icp-metering-api-secret",
					DefaultMode: &defaultSecretMode,
					Optional:    &TrueVar,
				},
			},
		}

		volumes = append(volumes, meteringAPICertVolume)
	}

	if spec.HTTPSEnable {
		if spec.HTTPSCertsSource == "custom" {
			licensingHTTPSCertsVolume := corev1.Volume{
				Name: LicensingHTTPSCertsVolumeName,
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName:  "ibm-licensing-certs",
						DefaultMode: &defaultSecretMode,
						Optional:    &TrueVar,
					},
				},
			}

			volumes = append(volumes, licensingHTTPSCertsVolume)
		}
	}

	return volumes
}
