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

package testutils

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var (
	defaultMode int32 = 420
	trueVar           = true
)

var Envs1 = []corev1.EnvVar{
	{Name: "env1", Value: "val1"},
	{Name: "env2", Value: "env2"},
}

var Envs2 = []corev1.EnvVar{
	{Name: "env1", Value: "val1"},
	{Name: "env2", Value: "env2"},
	{Name: "env3", Value: "env3"},
}

var Envs3 = []corev1.EnvVar{
	{Name: "env", Value: "val"},
}

var Envs5 = []corev1.EnvVar{
	{Name: "env1", Value: "val3"},
	{Name: "env2", Value: "env2"},
}

var Envs6 = []corev1.EnvVar{
	{Name: "env1", Value: "val1"},
	{Name: "env2", Value: "env2", ValueFrom: &corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "secret-a"}}}},
}

var Envs7 = []corev1.EnvVar{
	{Name: "env1", Value: "val1"},
	{Name: "env2", Value: "env2", ValueFrom: &corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "secret-b"}}}},
}

var Volumes1 = []corev1.Volume{
	{Name: "Volume A", VolumeSource: corev1.VolumeSource{
		EmptyDir: &corev1.EmptyDirVolumeSource{SizeLimit: resource.NewQuantity(600*1024*1024, resource.BinarySI)},
	}},
	{Name: "Volume B", VolumeSource: corev1.VolumeSource{
		Secret: &corev1.SecretVolumeSource{
			SecretName:  "Secret B",
			DefaultMode: &defaultMode,
			Optional:    &trueVar,
		},
	}},
}

var Volumes1DiffOrder = []corev1.Volume{
	{Name: "Volume B", VolumeSource: corev1.VolumeSource{
		Secret: &corev1.SecretVolumeSource{
			SecretName:  "Secret B",
			DefaultMode: &defaultMode,
			Optional:    &trueVar,
		},
	}},
	{Name: "Volume A", VolumeSource: corev1.VolumeSource{
		EmptyDir: &corev1.EmptyDirVolumeSource{SizeLimit: resource.NewQuantity(600*1024*1024, resource.BinarySI)},
	}},
}

var Volumes1AdditionalVolume = []corev1.Volume{
	{Name: "Volume A", VolumeSource: corev1.VolumeSource{
		EmptyDir: &corev1.EmptyDirVolumeSource{SizeLimit: resource.NewQuantity(600*1024*1024, resource.BinarySI)},
	}},
	{Name: "Volume B", VolumeSource: corev1.VolumeSource{
		Secret: &corev1.SecretVolumeSource{
			SecretName:  "Secret B",
			DefaultMode: &defaultMode,
			Optional:    &trueVar,
		},
	}},
	{Name: "Volume C", VolumeSource: corev1.VolumeSource{
		EmptyDir: &corev1.EmptyDirVolumeSource{SizeLimit: resource.NewQuantity(600*1024*1024, resource.BinarySI)},
	}},
}

var Volumes1DiffEmptyDirSize = []corev1.Volume{
	{Name: "Volume A", VolumeSource: corev1.VolumeSource{
		EmptyDir: &corev1.EmptyDirVolumeSource{SizeLimit: resource.NewQuantity(800*1024*1024, resource.BinarySI)},
	}},
	{Name: "Volume B", VolumeSource: corev1.VolumeSource{
		Secret: &corev1.SecretVolumeSource{
			SecretName:  "Secret B",
			DefaultMode: &defaultMode,
			Optional:    &trueVar,
		},
	}},
}

var Volumes2 = []corev1.Volume{
	{Name: "Volume C", VolumeSource: corev1.VolumeSource{
		EmptyDir: &corev1.EmptyDirVolumeSource{SizeLimit: resource.NewQuantity(600*1024*1024, resource.BinarySI)},
	}},
	{Name: "Volume B", VolumeSource: corev1.VolumeSource{
		Secret: &corev1.SecretVolumeSource{
			SecretName:  "Secret B",
			DefaultMode: &defaultMode,
			Optional:    &trueVar,
		},
	}},
}

var Volumes3 = []corev1.Volume{
	{Name: "Volume B", VolumeSource: corev1.VolumeSource{
		Secret: &corev1.SecretVolumeSource{
			SecretName:  "Secret B",
			DefaultMode: &defaultMode,
			Optional:    &trueVar,
		},
	}},
}

var Volumes4 = []corev1.Volume{
	{Name: "Volume A", VolumeSource: corev1.VolumeSource{
		EmptyDir: &corev1.EmptyDirVolumeSource{SizeLimit: resource.NewQuantity(600*1024*1024, resource.BinarySI)},
	}},
	{Name: "Volume B", VolumeSource: corev1.VolumeSource{
		Secret: &corev1.SecretVolumeSource{
			SecretName:  "Secret B",
			DefaultMode: &defaultMode,
		},
	}},
}
