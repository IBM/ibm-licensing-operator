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
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	odlm "github.com/IBM/operand-deployment-lifecycle-manager/api/v1alpha1"
	operatorframeworkv1 "github.com/operator-framework/api/pkg/operators/v1"
)

func TestGetLicensingOperatorGroupInNamespace(t *testing.T) {
	operatorNamespace := "ibm-licensing"
	scheme := runtime.NewScheme()
	odlm.AddToScheme(scheme)
	operatorframeworkv1.AddToScheme(scheme)

	t.Log("Given the need to get IBMLicensing OperatorGroup in certain namespace")
	{
		t.Log("\tTest 0:\tWhen there is licensing OperatorGroup in the namespace")
		{
			licensingOperatorGroup := OperatorGroupObj("ibm-licensing-og1", operatorNamespace, []string{operatorNamespace})
			operatorGroup := OperatorGroupObj("olm-default-og", operatorNamespace, []string{operatorNamespace})

			client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&operatorGroup, &licensingOperatorGroup).Build()

			foundOperatorGroup, err := GetLicensingOperatorGroupInNamespace(client, operatorNamespace)
			if err != nil {
				t.Fatalf("\t%s\tShould get licensing OperatorGroup without an error %s : %v", FAIL, operatorNamespace, err)
			}
			if !reflect.DeepEqual(foundOperatorGroup.TypeMeta, licensingOperatorGroup.TypeMeta) ||
				!reflect.DeepEqual(foundOperatorGroup.ObjectMeta, licensingOperatorGroup.ObjectMeta) ||
				!reflect.DeepEqual(foundOperatorGroup.Spec, licensingOperatorGroup.Spec) {
				t.Errorf("\t%s\tShould get licensing OperatorGroup", FAIL)
			} else {
				t.Logf("\t%s\tShould get licensing OperatorGroup", SUCCESS)
			}
		}

		t.Log("\tTest 1:\tWhen there is no licensing OperatorGroup in the namespace")
		{
			operatorGroup := OperatorGroupObj("olm-default-og", operatorNamespace, []string{operatorNamespace})

			client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&operatorGroup).Build()

			foundOperatorGroup, err := GetLicensingOperatorGroupInNamespace(client, operatorNamespace)
			if err != nil {
				t.Fatalf("\t%s\tShould not get an error %s : %v", FAIL, operatorNamespace, err)
			}
			if foundOperatorGroup != nil {
				t.Errorf("\t%s\tShould get nil instead of OperatorGroup object", FAIL)
			} else {
				t.Logf("\t%s\tShould get nil instead of OperatorGroup object", SUCCESS)
			}
		}

	}
}
