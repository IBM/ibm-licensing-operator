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
	"strings"

	v1 "github.com/operator-framework/api/pkg/operators/v1"

	c "sigs.k8s.io/controller-runtime/pkg/client"
)

const ibmLicensingPrefix = "IBMLicensing"

// Returns first found OperatorGroup with `ibm-licensing` in name, otherwise nil
func GetLicensingOperatorGroupInNamespace(reader c.Reader, namespace string) (*v1.OperatorGroup, error) {

	operatorGroupList := v1.OperatorGroupList{}
	listOpts := []c.ListOption{
		c.InNamespace(namespace),
	}

	err := reader.List(context.TODO(), &operatorGroupList, listOpts...)
	if err != nil {
		return nil, err
	}

	var foundOperatorGroup v1.OperatorGroup

	for _, operatorGroup := range operatorGroupList.Items {
		if val, exists := operatorGroup.Annotations["olm.providedAPIs"]; exists {
			if strings.Contains(val, ibmLicensingPrefix) {
				foundOperatorGroup = operatorGroup
				return &foundOperatorGroup, nil
			}
		}
	}

	return nil, nil
}

func ExtendOperatorGroupWithNamespaceList(namespaceList []string, operatorGroup *v1.OperatorGroup) *v1.OperatorGroup {
	operatorGroup.Spec.TargetNamespaces = append(operatorGroup.Spec.TargetNamespaces, namespaceList...)
	return operatorGroup
}
