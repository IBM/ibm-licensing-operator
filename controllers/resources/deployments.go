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

package resources

import (
	"reflect"
	"strings"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apieq "k8s.io/apimachinery/pkg/api/equality"
)

// To make linter happy
const containerErrorMessageStart = "Container "

func equalContainerLists(reqLogger *logr.Logger, containers1 []corev1.Container, containers2 []corev1.Container) bool {
	if len(containers1) != len(containers2) {
		(*reqLogger).Info("Deployment has wrong amount of containers")
		return false
	}
	if len(containers1) == 0 {
		return true
	}

	containersToBeChecked := map[*corev1.Container]*corev1.Container{}

	//map container with same names
	for i, container1 := range containers1 {
		foundContainer2 := false
		for j, container2 := range containers2 {
			if container1.Name == container2.Name {
				containersToBeChecked[&containers1[i]] = &containers2[j]
				foundContainer2 = true
				break
			}
		}
		if !foundContainer2 {
			return false
		}
	}

	potentialDifference := false
	for foundContainer, expectedContainer := range containersToBeChecked {
		if potentialDifference {
			break
		}
		potentialDifference = true
		if foundContainer.Image != expectedContainer.Image {
			(*reqLogger).Info(containerErrorMessageStart + foundContainer.Name + " wrong container image")
		} else if foundContainer.ImagePullPolicy != expectedContainer.ImagePullPolicy {
			(*reqLogger).Info(containerErrorMessageStart + foundContainer.Name + " wrong image pull policy")
		} else if !reflect.DeepEqual(foundContainer.Command, expectedContainer.Command) {
			(*reqLogger).Info(containerErrorMessageStart + foundContainer.Name + " wrong container command")
		} else if !reflect.DeepEqual(foundContainer.Ports, expectedContainer.Ports) {
			(*reqLogger).Info(containerErrorMessageStart + foundContainer.Name + " wrong containers ports")
		} else if !reflect.DeepEqual(foundContainer.VolumeMounts, expectedContainer.VolumeMounts) {
			(*reqLogger).Info(containerErrorMessageStart + foundContainer.Name + " wrong VolumeMounts in container")
		} else if !equalEnvVars(foundContainer.Env, expectedContainer.Env) { // DeepEqual requires same order of items, which results in false negatives, so we use custom comparison function
			(*reqLogger).Info(containerErrorMessageStart + foundContainer.Name + " wrong env variables in container")
		} else if !reflect.DeepEqual(foundContainer.SecurityContext, expectedContainer.SecurityContext) {
			(*reqLogger).Info(containerErrorMessageStart + foundContainer.Name + " wrong container security context")
		} else if (foundContainer.Resources.Limits == nil) || (foundContainer.Resources.Requests == nil) {
			(*reqLogger).Info(containerErrorMessageStart + foundContainer.Name + " wrong container Resources")
		} else if !(foundContainer.Resources.Limits.Cpu().Equal(*expectedContainer.Resources.Limits.Cpu()) &&
			foundContainer.Resources.Limits.Memory().Equal(*expectedContainer.Resources.Limits.Memory())) {
			(*reqLogger).Info(containerErrorMessageStart + foundContainer.Name + " wrong container Resources Limits")
		} else if !(foundContainer.Resources.Requests.Cpu().Equal(*expectedContainer.Resources.Requests.Cpu()) &&
			foundContainer.Resources.Requests.Memory().Equal(*expectedContainer.Resources.Requests.Memory()) &&
			foundContainer.Resources.Requests.StorageEphemeral().Equal(*expectedContainer.Resources.Requests.StorageEphemeral())) {
			(*reqLogger).Info("Container " + foundContainer.Name + " wrong container Resources Requests")
		} else if !equalHugePagesLimits(expectedContainer.Resources.Limits, foundContainer.Resources.Limits) {
			(*reqLogger).Info(containerErrorMessageStart + foundContainer.Name + " wrong hugepages limits")
		} else if !equalProbes(foundContainer.ReadinessProbe, expectedContainer.ReadinessProbe) {
			(*reqLogger).Info(containerErrorMessageStart + foundContainer.Name + " wrong container Readiness Probe")
		} else if !equalProbes(foundContainer.LivenessProbe, expectedContainer.LivenessProbe) {
			(*reqLogger).Info(containerErrorMessageStart + foundContainer.Name + " wrong container Liveness Probe")
		} else {
			potentialDifference = false
		}
	}
	return !potentialDifference
}

func ShouldUpdateDeployment(
	reqLogger *logr.Logger,
	expectedSpec *corev1.PodTemplateSpec,
	foundSpec *corev1.PodTemplateSpec) bool {

	// this ensures that the annotation is always the same in both objects (to exclude it from deepEqual comparison)
	if _, ok := expectedSpec.Annotations["kubectl.kubernetes.io/restartedAt"]; ok {
		foundSpec.Annotations["kubectl.kubernetes.io/restartedAt"] = expectedSpec.Annotations["kubectl.kubernetes.io/restartedAt"]
	} else {
		delete(foundSpec.Annotations, "kubectl.kubernetes.io/restartedAt")
	}

	if !apieq.Semantic.DeepEqual(foundSpec.Spec.Volumes, expectedSpec.Spec.Volumes) {
		(*reqLogger).Info("Deployment has wrong volumes")
	} else if !reflect.DeepEqual(foundSpec.Spec.Affinity, expectedSpec.Spec.Affinity) {
		(*reqLogger).Info("Deployment has wrong affinity")
	} else if foundSpec.Spec.ServiceAccountName != expectedSpec.Spec.ServiceAccountName {
		(*reqLogger).Info("Deployment wrong service account name")
	} else if !reflect.DeepEqual(foundSpec.Annotations, expectedSpec.Annotations) {
		(*reqLogger).Info("Deployment has wrong spec template annotations")
	} else if !equalContainerLists(reqLogger, foundSpec.Spec.Containers, expectedSpec.Spec.Containers) {
		(*reqLogger).Info("Deployment wrong containers")
	} else if !equalContainerLists(reqLogger, foundSpec.Spec.InitContainers, expectedSpec.Spec.InitContainers) {
		(*reqLogger).Info("Deployment wrong init containers")
	} else {
		return !MapHasAllPairsFromOther(foundSpec.GetLabels(), expectedSpec.GetLabels())
	}
	return true
}

func equalProbes(probe1 *corev1.Probe, probe2 *corev1.Probe) bool {
	if probe1 == nil {
		return probe2 == nil
	} else if probe2 == nil {
		return false
	}
	// need to set thresholds for not set values
	if probe1.SuccessThreshold == 0 {
		probe1.SuccessThreshold = probe2.SuccessThreshold
	} else if probe2.SuccessThreshold == 0 {
		probe2.SuccessThreshold = probe1.SuccessThreshold
	}
	if probe1.FailureThreshold == 0 {
		probe1.FailureThreshold = probe2.FailureThreshold
	} else if probe2.FailureThreshold == 0 {
		probe2.FailureThreshold = probe1.FailureThreshold
	}
	return reflect.DeepEqual(probe1, probe2)
}

func equalEnvVars(envVarArr1, envVarArr2 []corev1.EnvVar) bool {
	if len(envVarArr1) != len(envVarArr2) {
		return false
	}

	for _, env1 := range envVarArr1 {
		contains := false
		for _, env2 := range envVarArr2 {
			if env1.Name == env2.Name && env1.Value == env2.Value && reflect.DeepEqual(env1.ValueFrom, env2.ValueFrom) {
				contains = true
				break
			}
		}
		if !contains {
			return contains
		}
	}
	return true
}

func equalVolumes(foundVolumes, expectedVolumes []corev1.Volume) bool {
	if len(expectedVolumes) != len(foundVolumes) {
		return false
	}
	for _, expectedVolume := range expectedVolumes {
		contains := false
		for _, foundVolume := range foundVolumes {
			if foundVolume.Name == expectedVolume.Name {
				if !reflect.DeepEqual(foundVolume, expectedVolume) {
					return false
				}
				contains = true
				break
			}
		}
		if !contains {
			return false
		}
	}
	return true
}

func equalHugePagesLimits(expected, actual corev1.ResourceList) bool {
	for k, v := range expected {
		if strings.HasPrefix(k.String(), corev1.ResourceHugePagesPrefix) && !apieq.Semantic.DeepEqual(actual[k], v) {
			return false
		}
	}

	return true
}
