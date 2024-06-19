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
	"testing"

	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	operatorframeworkv1 "github.com/operator-framework/api/pkg/operators/v1"
	apieq "k8s.io/apimachinery/pkg/api/equality"

	odlm "github.com/IBM/operand-deployment-lifecycle-manager/api/v1alpha1"
)

func TestGetLicensingOperatorGroupInNamespace(t *testing.T) {
	operatorNamespace := "ibm-licensing"
	odlm.AddToScheme(scheme.Scheme)
	operatorframeworkv1.AddToScheme(scheme.Scheme)

	t.Log("Given the need to get IBMLicensing OperatorGroup in certain namespace")
	{
		t.Log("\tTest 0:\tWhen there is licensing OperatorGroup in the namespace")
		{
			licensingOperatorGroup := OperatorGroupObj("ibm-licensing-og1", operatorNamespace,
				map[string]string{"olm.providedAPIs": "IBMLicensing.v1alpha1.operator.ibm.com"}, []string{operatorNamespace})
			operatorGroup := OperatorGroupObj("olm-default-og", operatorNamespace,
				map[string]string{"olm.providedAPIs": "Fake.v1alpha1.operator.ibm.com"}, []string{operatorNamespace})

			client := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(&operatorGroup, &licensingOperatorGroup).Build()

			foundOperatorGroup, err := GetLicensingOperatorGroupInNamespace(client, operatorNamespace)
			if err != nil {
				t.Fatalf("\t%s\tShould get licensing OperatorGroup without an error %s : %v", FAIL, operatorNamespace, err)
			}
			if foundOperatorGroup != nil && apieq.Semantic.DeepEqual(foundOperatorGroup, &licensingOperatorGroup) {
				t.Logf("\t%s\tShould get licensing OperatorGroup", SUCCESS)
			} else {
				t.Errorf("\t%s\tShould get licensing OperatorGroup", FAIL)
			}
		}

		t.Log("\tTest 1:\tWhen there is no licensing OperatorGroup in the namespace")
		{
			operatorGroup := OperatorGroupObj("olm-default-og", operatorNamespace,
				map[string]string{"olm.ProvidedAPIs": "Fake.v1alpha1.operator.ibm.com"}, []string{operatorNamespace})

			client := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(&operatorGroup).Build()

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
