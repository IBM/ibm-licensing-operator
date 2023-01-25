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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	operatorv1alpha1 "github.com/IBM/ibm-licensing-operator/api/v1alpha1"
	res "github.com/IBM/ibm-licensing-operator/controllers/resources"
)

var replicas = int32(1)

func GetDeployment(instance *operatorv1alpha1.IBMLicenseServiceReporter) *appsv1.Deployment {
	metaLabels := LabelsForMeta(instance)
	selectorLabels := LabelsForSelector(instance)
	podLabels := LabelsForPod(instance)

	var imagePullSecrets []corev1.LocalObjectReference
	if instance.Spec.ImagePullSecrets != nil {
		for _, pullSecret := range instance.Spec.ImagePullSecrets {
			imagePullSecrets = append(imagePullSecrets, corev1.LocalObjectReference{Name: pullSecret})
		}
	}

	containers := []corev1.Container{
		GetDatabaseContainer(instance),
		GetReceiverContainer(instance),
	}
	if res.IsUIEnabled {
		containers = append(containers, GetReporterUIContainer(instance))
	}
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetResourceName(instance),
			Namespace: instance.GetNamespace(),
			Labels:    metaLabels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: selectorLabels,
			},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RecreateDeploymentStrategyType,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      podLabels,
					Annotations: res.AnnotationsForPod(),
				},
				Spec: corev1.PodSpec{
					Volumes:                       getLicenseServiceReporterVolumes(instance.Spec),
					InitContainers:                GetLicenseReporterInitContainers(instance),
					Containers:                    containers,
					TerminationGracePeriodSeconds: &res.Seconds60,
					ServiceAccountName:            GetServiceAccountName(instance),
					ImagePullSecrets:              imagePullSecrets,
					Affinity: &corev1.Affinity{
						NodeAffinity: &corev1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
								NodeSelectorTerms: []corev1.NodeSelectorTerm{
									{
										MatchExpressions: []corev1.NodeSelectorRequirement{
											{
												Key:      "kubernetes.io/arch",
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
	return deployment
}
