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

package resources

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// GetWatchNamespace returns the Namespace the operator should be watching for changes.
func GetWatchNamespace() (string, error) {
	// WatchNamespaceEnvVar is the constant for env variable WATCH_NAMESPACE
	// which specifies the Namespace to watch.
	// An empty value means the operator is running with cluster scope.
	var watchNamespaceEnvVar = "WATCH_NAMESPACE"

	ns, found := os.LookupEnv(watchNamespaceEnvVar)
	if !found {
		return "", fmt.Errorf("%s must be set", watchNamespaceEnvVar)
	}

	return ns, nil
}

// GetWatchNamespaceList returns list of namespaces operator should watch for changes.
func GetWatchNamespaceAsList() ([]string, error) {

	ns, err := GetWatchNamespace()
	if err != nil {
		return nil, err
	}

	return strings.Split(ns, ","), nil
}

// GetOperatorNamespace returns the Namespace the operator should be watching for changes.
func GetOperatorNamespace() (string, error) {
	// OperatorNamespaceEnvVar is the constant for env variable OPERATOR_NAMESPACE
	// which describes the namespace where operator is working.
	// An empty value means the operator is running with cluster scope.
	var operatorNamespaceEnvVar = "OPERATOR_NAMESPACE"

	ns, found := os.LookupEnv(operatorNamespaceEnvVar)
	if !found {
		return "", fmt.Errorf("%s must be set", operatorNamespaceEnvVar)
	}
	return ns, nil
}

// GetCrdReconcileInterval returns time duration in seconds for requested CRD watching. Defaults to 300s.
func GetCrdReconcileInterval() (time.Duration, error) {
	crdReconcileEnvVar := "CRD_RECONCILE_INTERVAL"

	defaultReconcileInterval := 300 * time.Second
	env, found := os.LookupEnv(crdReconcileEnvVar)

	if found {
		envVal, err := strconv.Atoi(env)
		if err != nil {
			return defaultReconcileInterval, fmt.Errorf("%s must be a natural number", crdReconcileEnvVar)
		}
		return time.Duration(envVal) * time.Second, nil
	}
	return defaultReconcileInterval, nil
}
