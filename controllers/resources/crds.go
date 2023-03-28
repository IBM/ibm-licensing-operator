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
	"errors"
	"os"
	"reflect"
	"time"

	"github.com/go-logr/logr"

	meta "k8s.io/apimachinery/pkg/api/meta"
	c "sigs.k8s.io/controller-runtime/pkg/client"
)

// Returns true if CRD for provided resource exits
func DoesCRDExist(reader c.Reader, foundRes c.ObjectList) (bool, error) {
	namespace, err := GetOperatorNamespace()
	if err != nil {
		return false, errors.New("OPERATOR_NAMESPACE env not found")
	}
	listOpts := []c.ListOption{
		c.InNamespace(namespace),
	}

	if err := reader.List(context.TODO(), foundRes, listOpts...); err != nil {
		// If CRD is not present on the cluster, NoKindMatchError is returned
		kindMatchErr := &meta.NoKindMatchError{}
		if errors.As(err, &kindMatchErr) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// Restarts operator if specified CRD appears on the cluster
func RestartOnCRDCreation(logger *logr.Logger, reader c.Reader, foundRes c.ObjectList, reconcileInterval time.Duration) {
	resType := reflect.TypeOf(foundRes)
	reqLogger := logger.WithValues("action", "Checking for "+resType.String()+" CRD existence")
	for {
		if isCrdExists, _ := DoesCRDExist(reader, foundRes); isCrdExists {
			reqLogger.Info(resType.String() + " CRD found on cluster. Operator will be restarted to enable handling it")
			os.Exit(0)
		}
		time.Sleep(reconcileInterval)
	}
}
