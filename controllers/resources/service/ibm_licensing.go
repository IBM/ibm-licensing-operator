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

package service

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	operatorv1alpha1 "github.com/IBM/ibm-licensing-operator/api/v1alpha1"
)

func GetDefaultIBMLicensing() operatorv1alpha1.IBMLicensing {
	return operatorv1alpha1.IBMLicensing{
		ObjectMeta: metav1.ObjectMeta{
			Name: "instance",
		},
		Spec: operatorv1alpha1.IBMLicensingSpec{
			Datasource:  "datacollector",
			HTTPSEnable: true,
		},
	}
}
