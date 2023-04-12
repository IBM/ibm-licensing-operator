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
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	ctrl "sigs.k8s.io/controller-runtime"
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

	isNssInstalled, err := IsNamespaceScopeOperatorInstalled()
	if err != nil {
		return "", err
	}
	if isNssInstalled {
		nssNs, err := getWatchNamespaceFromNssConfigMap()
		if err != nil {
			return "", err
		}
		return nssNs, nil
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

// +kubebuilder:rbac:namespace=ibm-licensing,groups=operator.com.ibm,resources=namespacescope;namespacescope/finalizers;namespacescope/status,verbs=get;list;watch

func IsNamespaceScopeOperatorInstalled() (bool, error) {
	nssCr, err := getNamespaceScopeCR()
	if err != nil {
		if k8serr.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	return nssCr != nil, nil
}

// +kubebuilder:rbac:namespace=ibm-licensing,groups=operator.com.ibm,resources=namespacescope;namespacescope/finalizers;namespacescope/status,verbs=get;list;watch

// Get namespaces to watch from Namespace Scope Operator ConfigMap. Should only be used in CP2/CP3 coexistence scenario.
func getWatchNamespaceFromNssConfigMap() (string, error) {
	ctx := context.Background()
	config := ctrl.GetConfigOrDie()
	dynamicClient := dynamic.NewForConfigOrDie(config)
	namespace, err := GetOperatorNamespace()
	if err != nil {
		return "", err // TODO
	}

	// List Namespace Scope CRs in the namespace.
	nssCrList, err := ListResourcesDynamically(ctx, dynamicClient, "operator.com.ibm", "v1", "namespacescope", namespace)
	if err != nil {
		return "", err // TODO
	}

	var nssConfigmapName string
	if len(nssCrList) == 0 {
		return "", nil // TODO
	}

	// Find first CR with configmapName set.
	for _, nssCr := range nssCrList {
		if spec, ok := nssCr.Object["spec"]; ok {
			if cmName, ok := spec.(map[string]interface{})["configmapName"]; ok {
				nssConfigmapName = cmName.(string)
				break
			}
		}
	}
	if nssConfigmapName == "" {
		return "", nil // TODO
	}

	// Get Config Map with name specified in Namespace Scope CR
	nssCm, err := GetResourceDynamically(ctx, dynamicClient, nssConfigmapName, "", "v1", "configmap", namespace)
	if err != nil {
		return "", err
	}
	if nssCm == nil {
		return "", nil
	}

	if data, ok := nssCm.Object["data"]; ok {
		if namespaces, ok := data.(map[string]interface{})["namespaces"]; ok {
			return namespaces.(string), nil
		}
	}

	return "", nil // TODO
}

// +kubebuilder:rbac:namespace=ibm-licensing,groups=operator.com.ibm,resources=namespacescope;namespacescope/finalizers;namespacescope/status,verbs=get;list;watch

func getNamespaceScopeCR() (*unstructured.Unstructured, error) {
	ctx := context.Background()
	config := ctrl.GetConfigOrDie()
	dynamicClient := dynamic.NewForConfigOrDie(config)
	namespace, err := GetOperatorNamespace()
	if err != nil {
		return nil, err
	}

	// List Namespace Scope Operator CRs in the namespace.
	nssCrList, err := ListResourcesDynamically(ctx, dynamicClient, "operator.ibm.com", "v1", "namespacescope", namespace)
	if err != nil {
		return nil, err
	}

	if len(nssCrList) == 0 {
		return nil, nil
	}

	return &nssCrList[0], nil
}
