package resources

import (
	"fmt"
	"os"
	"strings"
)

// GetWatchNamespace returns the Namespace the operator should be watching for changes
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

// GetWatchNamespaceList returns the Namespace the operator should be watching for changes in form of list
func GetWatchNamespaceList() ([]string, error) {

	ns, err := GetWatchNamespace()
	if err != nil {
		return nil, err
	}

	return strings.Split(ns, ","), nil
}

// GetOperatorNamespace returns the Namespace the operator should be watching for changes
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
