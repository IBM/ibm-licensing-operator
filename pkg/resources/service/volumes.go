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

package service

import (
	operatorv1alpha1 "github.com/ibm/ibm-licensing-operator/pkg/apis/operator/v1alpha1"
	res "github.com/ibm/ibm-licensing-operator/pkg/resources"
	corev1 "k8s.io/api/core/v1"
)

const APISecretTokenVolumeName = "api-token"
const APIUploadTokenVolumeName = "token-upload"
const MeteringAPICertsVolumeName = "metering-api-certs"
const LicensingHTTPSCertsVolumeName = "licensing-https-certs"

func getLicensingVolumeMounts(spec operatorv1alpha1.IBMLicensingSpec, isOpenShift bool) []corev1.VolumeMount {
	var volumeMounts = []corev1.VolumeMount{
		{
			Name:      APISecretTokenVolumeName,
			MountPath: "/opt/ibm/licensing/" + APISecretTokenKeyName,
			SubPath:   APISecretTokenKeyName,
			ReadOnly:  true,
		},
		{
			Name:      APIUploadTokenVolumeName,
			MountPath: "/opt/ibm/licensing/" + APIUploadTokenKeyName,
			SubPath:   APIUploadTokenKeyName,
			ReadOnly:  true,
		},
	}
	if spec.HTTPSEnable {
		if spec.HTTPSCertsSource == "custom" || (isOpenShift && spec.HTTPSCertsSource == res.Ocp) {
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

func getLicensingVolumes(spec operatorv1alpha1.IBMLicensingSpec, isOpenShift bool) []corev1.Volume {
	var volumes []corev1.Volume

	apiSecretTokenVolume := corev1.Volume{
		Name: APISecretTokenVolumeName,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName:  spec.APISecretToken,
				DefaultMode: &res.DefaultSecretMode,
			},
		},
	}

	volumes = append(volumes, apiSecretTokenVolume)

	apiUploadTokenVolume := corev1.Volume{
		Name: APIUploadTokenVolumeName,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName:  APIUploadTokenName,
				DefaultMode: &res.DefaultSecretMode,
			},
		},
	}

	volumes = append(volumes, apiUploadTokenVolume)

	if spec.IsMetering() {
		meteringAPICertVolume := corev1.Volume{
			Name: MeteringAPICertsVolumeName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  "icp-metering-api-secret",
					DefaultMode: &res.DefaultSecretMode,
					Optional:    &res.TrueVar,
				},
			},
		}

		volumes = append(volumes, meteringAPICertVolume)
	}

	if spec.HTTPSEnable {
		if spec.HTTPSCertsSource == "custom" {
			volumes = append(volumes, res.GetVolume(LicensingHTTPSCertsVolumeName, "ibm-licensing-certs"))
		} else if isOpenShift && spec.HTTPSCertsSource == res.Ocp {
			volumes = append(volumes, res.GetVolume(LicensingHTTPSCertsVolumeName, LiceseServiceOCPCertName))
		}
	}

	return volumes
}
