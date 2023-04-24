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
	"context"
	"testing"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	odlm "github.com/IBM/operand-deployment-lifecycle-manager/api/v1alpha1"
)

func TestDeleteBindInfoIfExists(t *testing.T) {
	namespace := "ibm-licensing"
	odlm.AddToScheme(scheme.Scheme)

	t.Log("Given the need to delete IBM Licensing OperandBindInfo in certain namespace")
	{
		t.Log("\tTest 0:\tWhen there is licensing OperandBindInfo in the namespace")
		{
			operandBindInfo := LicensingOperandBindInfo(LsBindInfoName, namespace)
			client := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(&operandBindInfo).Build()

			exisitngBindInfo := odlm.OperandBindInfo{}
			err := client.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: LsBindInfoName}, &exisitngBindInfo)
			if err != nil {
				t.Fatalf("\t%s\tFake Client setup error. Could not verify OperandBindInfo existence", FAIL)
			}

			err = DeleteBindInfoIfExists(context.Background(), client, client, namespace)
			if err != nil {
				t.Fatalf("\t%s\tShould not return an error during deleting OperandBindInfo", FAIL)
			}

			expectedBindInfo := odlm.OperandBindInfo{}
			err = client.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: LsBindInfoName}, &expectedBindInfo)
			if err != nil {
				if apierrors.IsNotFound(err) {
					t.Logf("\t%s\tShould delete found licensing OperandBindInfo", SUCCESS)
				} else {
					t.Errorf("\t%s\tShould delete found licensing OperandBindInfo without an error : %v", FAIL, err)
				}
			} else {
				t.Errorf("\t%s\tShould delete found licensing OperandBindInfo", FAIL)

			}
		}

		t.Log("\tTest 1:\tWhen there is no licensing OperandBindInfo in the namespace")
		{
			client := fake.NewClientBuilder().WithScheme(scheme.Scheme).Build()
			err := DeleteBindInfoIfExists(context.Background(), client, client, namespace)
			if err != nil {
				t.Errorf("\t%s\tShould not return an error when OperandBindInfo does not exist : %v", FAIL, err)
			} else {
				t.Logf("\t%s\tShould not return an error when OperandBindInfo does not exist", SUCCESS)
			}
		}
	}
}
