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
	"testing"

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

func TestUpdateResourceDoesNotRemoveExistingLabels(t *testing.T) {
	name := "resource"
	namespace := "test"
	opts := zap.Options{
		Development: true,
		TimeEncoder: zapcore.RFC3339TimeEncoder,
	}
	logger := zap.New(zap.UseFlagOptions(&opts))

	// Create a found resource with some custom label not present in the expected resource
	foundResource := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    map[string]string{"existing-label": "existing-value"},
		},
	}

	// Build the client with the found resource
	client := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(foundResource).Build()

	// Create an expected resource with some expected label (different to the existing label)
	expectedResource := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    map[string]string{"expected-label": "expected-value"},
		},
	}

	// Update the resource with an expected one and fetch the updated state from the cluster
	_, err := resources.UpdateResource(&logger, client, expectedResource, foundResource)
	assert.NoError(t, err)
	err = client.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: name}, foundResource)
	assert.NoError(t, err)

	// Check both labels present
	assert.Equal(t, "existing-value", foundResource.GetLabels()["existing-label"])
	assert.Equal(t, "expected-value", foundResource.GetLabels()["expected-label"])

	// Create another expected resource with some updated labels
	expectedResource = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"new-expected-label": "expected-value",
				"expected-label":     "expected-value-updated",
			},
		},
	}

	// Update the resource with the expected one and fetch the updated state from the cluster
	_, err = resources.UpdateResource(&logger, client, expectedResource, foundResource)
	assert.NoError(t, err)
	err = client.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: name}, foundResource)
	assert.NoError(t, err)

	// Check all labels present and with the correct values
	assert.Equal(t, "existing-value", foundResource.GetLabels()["existing-label"])
	assert.Equal(t, "expected-value-updated", foundResource.GetLabels()["expected-label"])
	assert.Equal(t, "expected-value", foundResource.GetLabels()["new-expected-label"])
}
