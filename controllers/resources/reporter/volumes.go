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

package reporter

import (
	corev1 "k8s.io/api/core/v1"

	operatorv1alpha1 "github.com/IBM/ibm-licensing-operator/api/v1alpha1"
	"github.com/IBM/ibm-licensing-operator/controllers/resources"
)

const APISecretTokenVolumeName = "api-token"
const LicenseReporterHTTPSCertsVolumeName = "license-reporter-https-certs"
const DatabaseCredentialsVolumeName = "db-config" //nolint:gosec
const persistentVolumeClaimVolumeName = "data"
const ReceiverTmpVolumeName = "receiver-tmp"
const DatabaseTmpVolumeName = "db-tmp"
const DatabaseSocketsVolumeName = "db-sockets"

func getReceiverVolumeMounts() []corev1.VolumeMount {
	var volumeMounts = []corev1.VolumeMount{
		{
			Name:      APISecretTokenVolumeName,
			MountPath: "/opt/ibm/licensing",
			ReadOnly:  true,
		},
		{
			Name:      DatabaseCredentialsVolumeName,
			MountPath: "/opt/ibm/licensing/" + DatabaseConfigSecretName,
			ReadOnly:  true,
		},
		{
			Name:      ReceiverTmpVolumeName,
			MountPath: "/tmp",
			ReadOnly:  false,
		},
	}
	volumeMounts = append(volumeMounts, []corev1.VolumeMount{
		{
			Name:      LicenseReporterHTTPSCertsVolumeName,
			MountPath: "/opt/licensing/certs/",
			ReadOnly:  true,
		},
	}...)

	return volumeMounts
}

func getDatabaseVolumeMounts() []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{
			Name:      persistentVolumeClaimVolumeName,
			MountPath: DatabaseMountPoint,
		},
		{
			Name:      DatabaseCredentialsVolumeName,
			MountPath: "/opt/ibm/licensing/" + DatabaseConfigSecretName,
			ReadOnly:  true,
		},
		{
			Name:      DatabaseSocketsVolumeName,
			MountPath: "/var/run/",
			ReadOnly:  false,
		},
		{
			Name:      DatabaseTmpVolumeName,
			MountPath: "/tmp/",
			ReadOnly:  false,
		},
	}
}

func getLicenseServiceReporterVolumes(spec operatorv1alpha1.IBMLicenseServiceReporterSpec) []corev1.Volume {
	volumes := []corev1.Volume{

		{
			Name: APISecretTokenVolumeName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  spec.APISecretToken,
					DefaultMode: &resources.DefaultSecretMode,
				},
			},
		},
		{
			Name: persistentVolumeClaimVolumeName,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: PersistenceVolumeClaimName,
				},
			},
		},
		{
			Name: DatabaseCredentialsVolumeName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  DatabaseConfigSecretName,
					DefaultMode: &resources.DefaultSecretMode,
				},
			},
		},
		{
			Name: ReceiverTmpVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
		{
			Name: DatabaseTmpVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
		{
			Name: DatabaseSocketsVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	}

	volumes = append(volumes, resources.GetVolume(LicenseReporterHTTPSCertsVolumeName, LicenseReportOCPCertName))

	return volumes
}
