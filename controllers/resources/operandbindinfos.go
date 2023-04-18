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
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	odlm "github.com/IBM/operand-deployment-lifecycle-manager/api/v1alpha1"
)

const LsBindInfoName = "ibm-licensing-bindinfo"

// Detect and delete existing IBM Licensing OperandBindInfo. It can still be left on cluster left after upgrade from 1.x.x to 4.x.x
func DeleteLicensingOperandBindInfo(ctx context.Context, client client.Client, namespace string) {

	bindinfo := odlm.OperandBindInfo{}

	for {
		time.Sleep(5 * time.Second)

		err := client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: LsBindInfoName}, &bindinfo)
		if err != nil {
			if apierrors.IsNotFound(err) {
				return
			}
			continue
		}

		err = client.Delete(ctx, &bindinfo)
		if err != nil {
			continue
		}

		return
	}
}
