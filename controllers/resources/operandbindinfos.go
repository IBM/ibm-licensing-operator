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

	"emperror.dev/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	odlm "github.com/IBM/operand-deployment-lifecycle-manager/api/v1alpha1"
)

const LsBindInfoName = "ibm-licensing-bindinfo"

// +kubebuilder:rbac:namespace=ibm-licensing,groups="operator.ibm.com",resources=operandbindinfos,verbs=get;list;watch;delete

// Detect and delete existing IBM Licensing OperandBindInfo.
func DeleteBindInfoIfExists(ctx context.Context, reader client.Reader, writer client.Writer, namespace string) error {

	const retryTime = 10 * time.Second
	var err error
	bindinfo := odlm.OperandBindInfo{}
	retries := 3

	for retries > 0 {
		err = reader.Get(ctx, types.NamespacedName{Namespace: namespace, Name: LsBindInfoName}, &bindinfo)
		if err != nil {
			if apierrors.IsNotFound(err) {
				return nil
			}
			time.Sleep(retryTime)
			retries = retries - 1
			continue
		}

		err = writer.Delete(ctx, &bindinfo)
		if err != nil {
			time.Sleep(retryTime)
			retries = retries - 1
			continue
		}

		return nil
	}

	return errors.Wrap(err, "Could not delete "+LsBindInfoName+" after 3 retires")
}
