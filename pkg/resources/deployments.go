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
	"reflect"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
)

func ShouldUpdateDeployment(
	reqLogger *logr.Logger,
	expectedSpec *corev1.PodTemplateSpec,
	foundSpec *corev1.PodTemplateSpec,
	isMetering bool) bool {
	// TODO: this should be refactored in some nice way where you only declare which parameters needs to be correct between resources
	shouldUpdate := true
	if !reflect.DeepEqual(foundSpec.Spec.Volumes, expectedSpec.Spec.Volumes) {
		(*reqLogger).Info("Deployment has wrong volumes")
	} else if !reflect.DeepEqual(foundSpec.Spec.Affinity, expectedSpec.Spec.Affinity) {
		(*reqLogger).Info("Deployment has wrong affinity")
	} else if foundSpec.Spec.ServiceAccountName != expectedSpec.Spec.ServiceAccountName {
		(*reqLogger).Info("Deployment wrong service account name")
	} else if len(foundSpec.Spec.Containers) != len(expectedSpec.Spec.Containers) {
		(*reqLogger).Info("Deployment has wrong number of containers")
	} else if len(foundSpec.Spec.InitContainers) != len(expectedSpec.Spec.InitContainers) {
		(*reqLogger).Info("Deployment has wrong number of init containers")
	} else if !reflect.DeepEqual(foundSpec.Annotations, expectedSpec.Annotations) {
		(*reqLogger).Info("Deployment has wrong spec template annotations")
	} else {
		shouldUpdate = false
		containersToBeChecked := map[*corev1.Container]*corev1.Container{&foundSpec.Spec.Containers[0]: &expectedSpec.Spec.Containers[0]}
		if isMetering {
			containersToBeChecked[&foundSpec.Spec.InitContainers[0]] = &expectedSpec.Spec.InitContainers[0]
		}
		for foundContainer, expectedContainer := range containersToBeChecked {
			if shouldUpdate {
				break
			}
			shouldUpdate = true
			if foundContainer.Name != expectedContainer.Name {
				(*reqLogger).Info("Deployment wrong container name")
			} else if foundContainer.Image != expectedContainer.Image {
				(*reqLogger).Info("Deployment wrong container image")
			} else if foundContainer.ImagePullPolicy != expectedContainer.ImagePullPolicy {
				(*reqLogger).Info("Deployment wrong image pull policy")
			} else if !reflect.DeepEqual(foundContainer.Command, expectedContainer.Command) {
				(*reqLogger).Info("Deployment wrong container command")
			} else if !reflect.DeepEqual(foundContainer.Ports, expectedContainer.Ports) {
				(*reqLogger).Info("Deployment wrong containers ports")
			} else if !reflect.DeepEqual(foundContainer.VolumeMounts, expectedContainer.VolumeMounts) {
				(*reqLogger).Info("Deployment wrong VolumeMounts in container")
			} else if !reflect.DeepEqual(foundContainer.Env, expectedContainer.Env) {
				(*reqLogger).Info("Deployment wrong env variables in container")
			} else if !reflect.DeepEqual(foundContainer.SecurityContext, expectedContainer.SecurityContext) {
				(*reqLogger).Info("Deployment wrong container security context")
			} else if (foundContainer.Resources.Limits == nil) || (foundContainer.Resources.Requests == nil) {
				(*reqLogger).Info("Deployment wrong container Resources")
			} else if !(foundContainer.Resources.Limits.Cpu().Equal(*expectedContainer.Resources.Limits.Cpu()) &&
				foundContainer.Resources.Limits.Memory().Equal(*expectedContainer.Resources.Limits.Memory())) {
				(*reqLogger).Info("Deployment wrong container Resources Limits")
			} else if !(foundContainer.Resources.Requests.Cpu().Equal(*expectedContainer.Resources.Requests.Cpu()) &&
				foundContainer.Resources.Requests.Memory().Equal(*expectedContainer.Resources.Requests.Memory())) {
				(*reqLogger).Info("Deployment wrong container Resources Requests")
			} else {
				shouldUpdate = false
			}
		}
	}
	return shouldUpdate
}
