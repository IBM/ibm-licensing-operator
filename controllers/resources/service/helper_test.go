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
	"context"
	"fmt"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/IBM/ibm-licensing-operator/controllers/resources"
)

// Required for failing deployment update call and recreating it
type failUpdateWrapper struct {
	client.Client
}

// Fail update call if the object is a deployment with the "fail-update" flag present
func (c *failUpdateWrapper) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	if deployment, ok := obj.(*appsv1.Deployment); ok {
		if _, exists := deployment.ObjectMeta.Labels["fail-update"]; exists {
			return fmt.Errorf("failed to update the deployment because of the fail-update flag present")
		}
	}

	return c.Client.Update(ctx, obj, opts...)
}

func TestUpdateResourceDoesNotRemoveExistingLabelsAndAnnotations(t *testing.T) {
	deploymentName := "deployment"
	secretName := "secret"
	namespace := "test"
	opts := zap.Options{
		Development: true,
		TimeEncoder: zapcore.RFC3339TimeEncoder,
	}
	logger := zap.New(zap.UseFlagOptions(&opts))

	// Create a found resource with some custom metadata not present in the expected resource
	foundSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        secretName,
			Namespace:   namespace,
			Labels:      map[string]string{"existing-label": "existing-value"},
			Annotations: map[string]string{"existing-annotation": "existing-value"},
		},
	}

	// Also create a deployment to test different types of resources with this functionality
	foundDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        deploymentName,
			Namespace:   namespace,
			Labels:      map[string]string{"existing-label": "existing-value", "fail-update": "true"},
			Annotations: map[string]string{"existing-annotation": "existing-value", "fail-update": "true"},
		},
	}

	// Build the client with the found resource
	fakeClient := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(foundSecret, foundDeployment).Build()
	wrappedClient := &failUpdateWrapper{fakeClient}

	// Create an expected resource with some expected metadata (different to the existing metadata)
	expectedSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        secretName,
			Namespace:   namespace,
			Labels:      map[string]string{"expected-label": "expected-value"},
			Annotations: map[string]string{"expected-annotation": "expected-value"},
		},
	}

	// Update the resource with an expected one and fetch the updated state from the cluster
	_, err := resources.UpdateResource(&logger, wrappedClient, expectedSecret, foundSecret)
	assert.NoError(t, err)
	err = wrappedClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: secretName}, foundSecret)
	assert.NoError(t, err)

	// Check both metadata values present
	assert.Equal(t, "existing-value", foundSecret.GetLabels()["existing-label"])
	assert.Equal(t, "expected-value", foundSecret.GetLabels()["expected-label"])
	assert.Equal(t, "existing-value", foundSecret.GetAnnotations()["existing-annotation"])
	assert.Equal(t, "expected-value", foundSecret.GetAnnotations()["expected-annotation"])

	// Create another expected resource with some updated metadata
	expectedSecret = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
			Labels: map[string]string{
				"new-expected-label": "expected-value",
				"expected-label":     "expected-value-updated",
			},
			Annotations: map[string]string{
				"new-expected-annotation": "expected-value",
				"expected-annotation":     "expected-value-updated",
			},
		},
	}

	// Update the resource with the expected one and fetch the updated state from the cluster
	_, err = resources.UpdateResource(&logger, wrappedClient, expectedSecret, foundSecret)
	assert.NoError(t, err)
	err = wrappedClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: secretName}, foundSecret)
	assert.NoError(t, err)

	// Check all metadata present and with the correct values
	assert.Equal(t, "existing-value", foundSecret.GetLabels()["existing-label"])
	assert.Equal(t, "expected-value-updated", foundSecret.GetLabels()["expected-label"])
	assert.Equal(t, "expected-value", foundSecret.GetLabels()["new-expected-label"])
	assert.Equal(t, "existing-value", foundSecret.GetAnnotations()["existing-annotation"])
	assert.Equal(t, "expected-value-updated", foundSecret.GetAnnotations()["expected-annotation"])
	assert.Equal(t, "expected-value", foundSecret.GetAnnotations()["new-expected-annotation"])

	// Create another expected resource but of different type
	expectedDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        deploymentName,
			Namespace:   namespace,
			Labels:      map[string]string{"expected-label": "expected-value"},
			Annotations: map[string]string{"expected-annotation": "expected-value"},
		},
	}

	// Update the resource with an expected one and fetch the updated state from the cluster
	_, err = resources.UpdateResource(&logger, wrappedClient, expectedDeployment, foundDeployment)
	assert.NoError(t, err)
	err = wrappedClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: deploymentName}, foundDeployment)
	assert.NoError(t, err)

	// Check both metadata values present
	assert.Equal(t, "existing-value", foundDeployment.GetLabels()["existing-label"])
	assert.Equal(t, "expected-value", foundDeployment.GetLabels()["expected-label"])
	assert.Equal(t, "existing-value", foundDeployment.GetAnnotations()["existing-annotation"])
	assert.Equal(t, "expected-value", foundDeployment.GetAnnotations()["expected-annotation"])
}
