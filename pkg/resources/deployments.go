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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO: maybe add to ibmLicensingSpec struct later
var replicas = int32(1)

func GetLicensingDeployment(instance *operatorv1alpha1.IBMLicensing) *appsv1.Deployment {
	metaLabels := LabelsForLicensingMeta(instance)
	selectorLabels := LabelsForLicensingSelector(instance)
	podLabels := LabelsForLicensingPod(instance)
	// TODO: add init containers later
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetResourceName(instance),
			Namespace: instance.Spec.InstanceNamespace,
			Labels:    metaLabels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: selectorLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      podLabels,
					Annotations: AnnotationsForPod(),
				},
				Spec: corev1.PodSpec{
					Volumes: getLicensingVolumes(instance.Spec),
					Containers: []corev1.Container{
						GetLicensingContainer(instance.Spec),
					},
					TerminationGracePeriodSeconds: &seconds60,
					ServiceAccountName:            GetServiceAccountName(instance),
					Affinity: &corev1.Affinity{
						NodeAffinity: &corev1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
								NodeSelectorTerms: []corev1.NodeSelectorTerm{
									{
										MatchExpressions: []corev1.NodeSelectorRequirement{
											{
												Key:      "beta.kubernetes.io/arch",
												Operator: corev1.NodeSelectorOpIn,
												Values:   []string{"amd64"},
											},
										},
									},
								},
							},
						},
					},
					Tolerations: []corev1.Toleration{
						{
							Key:      "dedicated",
							Operator: corev1.TolerationOpExists,
							Effect:   corev1.TaintEffectNoSchedule,
						},
						{
							Key:      "CriticalAddonsOnly",
							Operator: corev1.TolerationOpExists,
						},
					},
				},
			},
		},
	}
}
