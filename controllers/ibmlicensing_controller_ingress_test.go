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

package controllers

import (
	"context"
	"errors"
	"testing"

	operatorv1alpha1 "github.com/IBM/ibm-licensing-operator/api/v1alpha1"
	"github.com/IBM/ibm-licensing-operator/controllers/resources/service"
	"go.uber.org/zap/zapcore"
	networkingv1 "k8s.io/api/networking/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	k8errors "k8s.io/apimachinery/pkg/api/errors"
)

const (
	SUCCESS = "\u2713"
	FAIL    = "\u2717"
)

// FakeClientWithGetError embeds the fake.Client and overrides the Get method to cause error
type FakeClientWithGetError struct {
	client.Client
	ErrorCausingObjNamespace string
	ErrorCausingObjName      string
}

var ErrFakeClientGet = errors.New("simulated error")

func (c FakeClientWithGetError) Get(ctx context.Context, key client.ObjectKey, obj client.Object) error {
	if (key.Name == c.ErrorCausingObjName) && (key.Namespace == c.ErrorCausingObjNamespace) {
		return ErrFakeClientGet
	}
	return c.Client.Get(ctx, key, obj)
}

func TestReconcileIngress(t *testing.T) {
	schema := runtime.NewScheme()
	operatorv1alpha1.AddToScheme(schema)
	networkingv1.AddToScheme(schema)

	t.Log("Given the need to reconcile IBMLicensing ingress")
	{
		t.Log("\tTest 0:\tWhen good ingress already exists")
		{
			// Set up your expectedIngress object
			trueVar := true
			instance := &operatorv1alpha1.IBMLicensing{
				// Fill in the necessary fields for your IBMLicensing instance
				ObjectMeta: v1.ObjectMeta{
					Name: "instance",
				},
				Spec: operatorv1alpha1.IBMLicensingSpec{
					IngressEnabled:    &trueVar,
					InstanceNamespace: "ibm-licensing",
				},
			}
			expectedIngress := service.GetLicensingIngress(instance)

			// Set up the fake client
			foundIngress := service.GetLicensingIngress(instance)
			fakeClient := fake.NewClientBuilder().WithScheme(schema).WithObjects(&foundIngress).Build()

			// Create the reconciler
			reconciler := IBMLicensingReconciler{
				Client: fakeClient,
				Log: zap.New(func(o *zap.Options) {
					o.Development = true
					o.TimeEncoder = zapcore.RFC3339TimeEncoder
				}),
				Scheme: schema,
			}

			// Call the reconcileIngress function
			result, err := reconciler.reconcileIngress(instance)

			// Check the result and error returned
			if err != nil {
				t.Fatalf("\t%s\tReconciling ingress when it exists and is correct should be without an error: %v", FAIL, err)
			}
			if result != (reconcile.Result{}) {
				t.Fatalf("\t%s\tReconciling ingress when it exists and is correct should return an empty reconcile result, but got: %v", FAIL, result)
			}

			// Assert the expected state of the ingress object
			foundIngress = networkingv1.Ingress{}
			namespacedName := types.NamespacedName{Name: expectedIngress.GetName(), Namespace: expectedIngress.GetNamespace()}
			err = reconciler.Client.Get(context.Background(), namespacedName, &foundIngress)
			if err != nil {
				t.Fatalf("\t%s\tFake k8s client should get correct foundIngress without an error: %v", FAIL, err)
			}
			if !service.IsIngressInDesiredState(foundIngress, expectedIngress, reconciler.Log) {
				t.Fatalf("\t%s\tFake k8s client should find correct foundIngress: %v, expectedIngress: %v", FAIL, foundIngress, expectedIngress)
			}
		}
	}
	t.Log("\tTest 1:\tWhen instance.Spec.IsIngressEnabled() returns true and an error occurs during reconciling")
	{
		// Set up your expectedIngress object
		trueVar := true
		instance := operatorv1alpha1.IBMLicensing{
			ObjectMeta: v1.ObjectMeta{
				Name: "instance",
			},
			// Fill in the necessary fields for your IBMLicensing instance
			Spec: operatorv1alpha1.IBMLicensingSpec{
				IngressEnabled:    &trueVar,
				InstanceNamespace: "ibm-licensing",
			},
		}

		// Set up the fake client
		foundIngress := service.GetLicensingIngress(&instance)
		fakeClient := fake.NewClientBuilder().WithScheme(schema).WithObjects(&foundIngress).Build()

		// Create the custom fake client
		customFakeClient := FakeClientWithGetError{Client: fakeClient, ErrorCausingObjNamespace: foundIngress.Namespace, ErrorCausingObjName: foundIngress.Name}

		// Create the reconciler
		reconciler := &IBMLicensingReconciler{
			Client: customFakeClient,
			Log: zap.New(func(o *zap.Options) {
				o.Development = true
				o.TimeEncoder = zapcore.RFC3339TimeEncoder
			}),
			Scheme: schema,
		}

		// Call the reconcileIngress function
		_, err := reconciler.reconcileIngress(&instance)

		// Check the result and error returned
		if err != ErrFakeClientGet {
			t.Fatalf("\t%s\tReconciling ingress should return an error when reconcileResourceNamespacedExistence fails", FAIL)
		}
	}
	t.Log("\tTest 2:\tWhen instance.Spec.IsIngressEnabled() returns true and the ingress needs an update")
	{
		// Set up your expectedIngress and foundIngress object
		trueVar := true
		instance := &operatorv1alpha1.IBMLicensing{
			ObjectMeta: v1.ObjectMeta{
				Name: "instance",
				UID:  "12345",
			},
			Spec: operatorv1alpha1.IBMLicensingSpec{
				IngressEnabled:    &trueVar,
				InstanceNamespace: "ibm-licensing",
			},
		}
		expectedIngress := service.GetLicensingIngress(instance)
		foundIngress := service.GetLicensingIngress(instance)

		// Modify foundIngress to require an update
		foundIngress.Spec.Rules[0].HTTP.Paths[0].Path = "/updated-path"

		// Set correct controller refs
		controllerutil.SetControllerReference(instance, &foundIngress, schema)
		controllerutil.SetControllerReference(instance, &expectedIngress, schema)

		// Set up the fake client with existing ingress that needs an update
		fakeClient := fake.NewClientBuilder().WithScheme(schema).WithObjects(&foundIngress).Build()

		// Create the reconciler
		reconciler := IBMLicensingReconciler{
			Client: fakeClient,
			Log: zap.New(func(o *zap.Options) {
				o.Development = true
				o.TimeEncoder = zapcore.RFC3339TimeEncoder
			}),
			Scheme: schema,
		}

		// Call the reconcileIngress function
		_, err := reconciler.reconcileIngress(instance)

		// Check the result and error returned
		if err != nil {
			t.Fatalf("\t%s\tReconciling ingress when it needs an update should be without an error: %v", FAIL, err)
		}

		// Assert the updated state of the ingress object
		foundIngress = networkingv1.Ingress{}
		namespacedName := types.NamespacedName{Name: expectedIngress.GetName(), Namespace: expectedIngress.GetNamespace()}
		err = reconciler.Client.Get(context.Background(), namespacedName, &foundIngress)
		if err != nil {
			t.Fatalf("\t%s\tFake k8s client should get updated foundIngress without an error: %v", FAIL, err)
		}
		if !service.IsIngressInDesiredState(foundIngress, expectedIngress, reconciler.Log) {
			t.Fatalf("\t%s\tFake k8s client should find updated foundIngress: %v, expectedIngress: %v", FAIL, foundIngress, expectedIngress)
		}
	}
	t.Log("\tTest 3:\tWhen instance.Spec.IsIngressEnabled() returns true and the ingress has a different owner and should not be updated")
	{
		// Set up your expectedIngress object
		trueVar := true
		instance := &operatorv1alpha1.IBMLicensing{
			ObjectMeta: v1.ObjectMeta{
				Name: "instance",
			},
			Spec: operatorv1alpha1.IBMLicensingSpec{
				IngressEnabled:    &trueVar,
				InstanceNamespace: "ibm-licensing",
			},
		}
		expectedIngress := service.GetLicensingIngress(instance)

		// Modify expectedIngress to have a different owner
		expectedIngress.SetOwnerReferences([]v1.OwnerReference{
			{
				APIVersion: "test-api/v1",
				Kind:       "TestOwner",
				Name:       "testowner",
				UID:        "12345",
			},
		})

		// Set up the fake client with existing ingress that has a different owner
		fakeClient := fake.NewClientBuilder().WithScheme(schema).WithObjects(&expectedIngress).Build()

		// Create the reconciler
		reconciler := IBMLicensingReconciler{
			Client: fakeClient,
			Log: zap.New(func(o *zap.Options) {
				o.Development = true
				o.TimeEncoder = zapcore.RFC3339TimeEncoder
			}),
			Scheme: schema,
		}

		// Call the reconcileIngress function
		result, err := reconciler.reconcileIngress(instance)

		// Check the result and error returned
		if err != nil {
			t.Fatalf("\t%s\tReconciling ingress when it has a different owner should be without an error: %v", FAIL, err)
		}
		if result != (reconcile.Result{}) {
			t.Fatalf("\t%s\tReconciling ingress when it has a different owner should return an empty reconcile result, but got: %v", FAIL, result)
		}

		// Assert the expected state of the ingress object
		foundIngress := networkingv1.Ingress{}
		namespacedName := types.NamespacedName{Name: expectedIngress.GetName(), Namespace: expectedIngress.GetNamespace()}
		err = reconciler.Client.Get(context.Background(), namespacedName, &foundIngress)
		if err != nil {
			t.Fatalf("\t%s\tFake k8s client should get foundIngress without an error: %v", FAIL, err)
		}
		if foundIngress.OwnerReferences[0].Name == instance.Name {
			t.Fatalf("\t%s\tIngress should not be updated when it has a different owner: %v", FAIL, foundIngress)
		}
	}
	t.Log("\tTest 4:\tWhen instance.Spec.IsIngressEnabled() returns false and the ingress does not exist (as expected)")
	{
		// Set up your instance object with IngressEnabled set to false
		falseVar := false
		instance := &operatorv1alpha1.IBMLicensing{
			ObjectMeta: v1.ObjectMeta{
				Name: "instance",
			},
			Spec: operatorv1alpha1.IBMLicensingSpec{
				IngressEnabled:    &falseVar,
				InstanceNamespace: "ibm-licensing",
			},
		}

		// Set up the fake client with no ingress object
		fakeClient := fake.NewClientBuilder().WithScheme(schema).Build()

		// Create the reconciler
		reconciler := IBMLicensingReconciler{
			Client: fakeClient,
			Log: zap.New(func(o *zap.Options) {
				o.Development = true
				o.TimeEncoder = zapcore.RFC3339TimeEncoder
			}),
			Scheme: schema,
		}

		// Call the reconcileIngress function
		result, err := reconciler.reconcileIngress(instance)

		// Check the result and error returned
		if err != nil {
			t.Fatalf("\t%s\tReconciling ingress when it does not exist and IngressEnabled is false should be without an error: %v", FAIL, err)
		}
		if result != (reconcile.Result{}) {
			t.Fatalf("\t%s\tReconciling ingress when it does not exist and IngressEnabled is false should return an empty reconcile result, but got: %v", FAIL, result)
		}

		// Assert that no ingress object was created
		foundIngresses := networkingv1.IngressList{}
		err = reconciler.Client.List(context.Background(), &foundIngresses)
		if err != nil {
			t.Fatalf("\t%s\tIngress listing should be possible: %v", FAIL, err)
		}
		if len(foundIngresses.Items) != 0 {
			t.Fatalf("\t%s\tIngress should not be created when IngressEnabled is false, ingresses: %v", FAIL, foundIngresses)
		}
	}
	t.Log("\tTest 5:\tWhen instance.Spec.IsIngressEnabled() returns false and an error occurs while getting the ingress")
	{
		// Set up your instance object with IngressEnabled set to false
		falseVar := false
		instance := operatorv1alpha1.IBMLicensing{
			ObjectMeta: v1.ObjectMeta{
				Name: "instance",
			},
			Spec: operatorv1alpha1.IBMLicensingSpec{
				IngressEnabled:    &falseVar,
				InstanceNamespace: "ibm-licensing",
			},
		}

		// Set up the fake client
		foundIngress := service.GetLicensingIngress(&instance)
		fakeClient := fake.NewClientBuilder().WithScheme(schema).WithObjects(&foundIngress).Build()

		// Create the custom fake client
		customFakeClient := FakeClientWithGetError{Client: fakeClient, ErrorCausingObjNamespace: foundIngress.Namespace, ErrorCausingObjName: foundIngress.Name}

		// Create the reconciler
		reconciler := &IBMLicensingReconciler{
			Client: customFakeClient,
			Log: zap.New(func(o *zap.Options) {
				o.Development = true
				o.TimeEncoder = zapcore.RFC3339TimeEncoder
			}),
			Scheme: schema,
		}

		// Call the reconcileIngress function
		result, err := reconciler.reconcileIngress(&instance)

		// Check the result and error returned
		if err != nil {
			t.Fatalf("\t%s\tReconciling ingress should return nil to allow continue, even thou there was a fail during reconcilation, got err: %v", FAIL, err)
		}
		if result != (reconcile.Result{}) {
			t.Fatalf("\t%s\tReconciling ingress should return nil to allow continue, even thou there was a fail during reconcilation, got result: %v", FAIL, result)
		}
	}
	t.Log("\tTest 6:\tWhen instance.Spec.IsIngressEnabled() returns false and the ingress exists but has a different owner and should not be deleted")
	{
		// Set up your instance object with IngressEnabled set to false
		falseVar := false
		instance := &operatorv1alpha1.IBMLicensing{
			ObjectMeta: v1.ObjectMeta{
				Name: "instance",
			},
			Spec: operatorv1alpha1.IBMLicensingSpec{
				IngressEnabled:    &falseVar,
				InstanceNamespace: "ibm-licensing",
			},
		}
		expectedIngress := service.GetLicensingIngress(instance)

		// Modify expectedIngress to have a different owner
		expectedIngress.SetOwnerReferences([]v1.OwnerReference{
			{
				APIVersion: "test-api/v1",
				Kind:       "TestOwner",
				Name:       "testowner",
				UID:        "12345",
			},
		})

		// Set up the fake client with existing ingress that has a different owner
		fakeClient := fake.NewClientBuilder().WithScheme(schema).WithObjects(&expectedIngress).Build()

		// Create the reconciler
		reconciler := IBMLicensingReconciler{
			Client: fakeClient,
			Log: zap.New(func(o *zap.Options) {
				o.Development = true
				o.TimeEncoder = zapcore.RFC3339TimeEncoder
			}),
			Scheme: schema,
		}

		// Call the reconcileIngress function
		result, err := reconciler.reconcileIngress(instance)

		// Check the result and error returned
		if err != nil {
			t.Fatalf("\t%s\tReconciling ingress when it has a different owner and IngressEnabled is false should be without an error: %v", FAIL, err)
		}
		if result != (reconcile.Result{}) {
			t.Fatalf("\t%s\tReconciling ingress when it has a different owner and IngressEnabled is false should return an empty reconcile result, but got: %v", FAIL, result)
		}

		// Assert the expected state of the ingress object
		foundIngress := networkingv1.Ingress{}
		namespacedName := types.NamespacedName{Name: expectedIngress.GetName(), Namespace: expectedIngress.GetNamespace()}
		err = reconciler.Client.Get(context.Background(), namespacedName, &foundIngress)
		if err != nil {
			t.Fatalf("\t%s\tFake k8s client should get foundIngress without an error: %v", FAIL, err)
		}
		if foundIngress.OwnerReferences[0].Name != "testowner" {
			t.Fatalf("\t%s\tIngress should not be deleted when it has a different owner: %v", FAIL, foundIngress)
		}
	}

	t.Log("\tTest 7:\tWhen instance.Spec.IsIngressEnabled() returns false, the ingress exists, and it should be deleted")
	{
		// Set up your instance object with IngressEnabled set to false
		falseVar := false
		instance := &operatorv1alpha1.IBMLicensing{
			ObjectMeta: v1.ObjectMeta{
				Name: "instance",
			},
			Spec: operatorv1alpha1.IBMLicensingSpec{
				IngressEnabled:    &falseVar,
				InstanceNamespace: "ibm-licensing",
			},
		}
		expectedIngress := service.GetLicensingIngress(instance)

		// Set correct controller refs
		controllerutil.SetControllerReference(instance, &expectedIngress, schema)

		// Set up the fake client with existing ingress
		fakeClient := fake.NewClientBuilder().WithScheme(schema).WithObjects(&expectedIngress).Build()

		// Create the reconciler
		reconciler := IBMLicensingReconciler{
			Client: fakeClient,
			Log: zap.New(func(o *zap.Options) {
				o.Development = true
				o.TimeEncoder = zapcore.RFC3339TimeEncoder
			}),
			Scheme: schema,
		}

		// Call the reconcileIngress function
		result, err := reconciler.reconcileIngress(instance)

		// Check the result and error returned
		if err != nil {
			t.Fatalf("\t%s\tReconciling ingress when it exists and IngressEnabled is false should be without an error: %v", FAIL, err)
		}
		if !result.Requeue {
			t.Fatalf("\t%s\tReconciling ingress when it exists and IngressEnabled is false should delete and return reconcile result with requeue, but got: %v", FAIL, result)
		}

		// Assert that the ingress object has been deleted
		foundIngress := networkingv1.Ingress{}
		namespacedName := types.NamespacedName{Name: expectedIngress.GetName(), Namespace: expectedIngress.GetNamespace()}
		err = reconciler.Client.Get(context.Background(), namespacedName, &foundIngress)
		if err == nil || !k8errors.IsNotFound(err) {
			t.Fatalf("\t%s\tIngress should be deleted when IngressEnabled is false: %v", FAIL, foundIngress)
		}
	}
}
