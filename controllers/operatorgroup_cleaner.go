//
// Copyright 2024 IBM Corporation
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

package controllers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-logr/logr"
	operatorframeworkv1 "github.com/operator-framework/api/pkg/operators/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	res "github.com/IBM/ibm-licensing-operator/controllers/resources"
)

func RunRemoveStaleNamespacesFromOperatorGroupTask(ctx context.Context, logger *logr.Logger, client client.Client, reader client.Reader) {
	// Immediately run the task once before starting the ticker loop
	logger.Info("Running task of removing stale namespaces from OperatorGroup")
	removeStaleNamespacesFromOperatorGroup(logger, client, reader)

	ticker := time.NewTicker(time.Hour) // runs every hour

	for {
		select {
		case <-ticker.C:
			logger.Info("Running task of removing stale namespaces from OperatorGroup")
			removeStaleNamespacesFromOperatorGroup(logger, client, reader)
		case <-ctx.Done():
			logger.Info("Stopping task of removing stale namespaces from OperatorGroup")
			ticker.Stop()
			return
		}
	}
}

/*
Periodically checks and updates the targetNamespaces field in the OperatorGroup.

The OperatorGroup ibm-licensing-XXX contains a field called targetNamespaces, which specifies the list of namespaces
where the operator should watch for resources. If a namespace is deleted from the cluster but its name
still remains in the targetNamespaces field of the OperatorGroup, the operator may throw errors and fail
to reconcile resources properly.

To prevent such errors, this function periodically verifies whether the namespaces listed in targetNamespaces
still exist in the cluster. If any namespaces are found to be missing, they are removed from the targetNamespaces
list. This update causes the operator to restart and watch only existing namespaces.
*/
func removeStaleNamespacesFromOperatorGroup(logger *logr.Logger, client client.Client, reader client.Reader) {
	watchNamespaces, err := res.GetWatchNamespaceAsList()
	if err != nil {
		logger.Error(err, "Unable to get WATCH_NAMESPACE")
		return
	}

	operatorNamespace, err := res.GetOperatorNamespace()
	if err != nil {
		logger.Error(err, "Unable to get OPERATOR_NAMESPACE")
		return
	}

	for _, ns := range watchNamespaces {
		if !namespaceExists(logger, reader, ns) {
			logger.Info("Namespace does not exist: " + ns + " Attempting to remove it from OperatorGroup.")
			err := removeNamespaceFromOperatorGroup(logger, client, operatorNamespace, ns)
			if err != nil {
				logger.Error(err, "Failed to remove stale namespace from OperatorGroup: "+ns)
			}
		}
	}
}

/*
Checks if namespace with given name exits in the cluster.
*/
func namespaceExists(logger *logr.Logger, reader client.Reader, ns string) bool {
	namespace := &corev1.Namespace{}
	err := reader.Get(context.TODO(), client.ObjectKey{Name: ns}, namespace)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			logger.Error(err, "Failed to check namespace existence: "+ns)
			return false
		}
		return false
	}
	return true
}

/*
Looks for OperatorGroup from operator's namespace which name starts with "ibm-licensing-".
Removes given namespace from targetNamespaces list field if it contains the namespace.
*/
func removeNamespaceFromOperatorGroup(logger *logr.Logger, cli client.Client, namespace, namespaceToRemove string) error {
	operatorGroupList := &operatorframeworkv1.OperatorGroupList{}
	if err := cli.List(context.Background(), operatorGroupList, &client.ListOptions{Namespace: namespace}); err != nil {
		return fmt.Errorf("failed to list OperatorGroups: %v", err)
	}

	for i := range operatorGroupList.Items {
		operatorGroup := operatorGroupList.Items[i]
		if strings.HasPrefix(operatorGroup.Name, "ibm-licensing-") {
			logger.Info("Found OperatorGroup: " + operatorGroup.Name + " in namespace " + namespace)

			targetNamespaces := operatorGroup.Spec.TargetNamespaces
			var updatedNamespaces []string

			for _, ns := range targetNamespaces {
				if ns != namespaceToRemove {
					updatedNamespaces = append(updatedNamespaces, ns)
				}
			}

			if len(updatedNamespaces) == len(targetNamespaces) {
				continue
			}

			operatorGroup.Spec.TargetNamespaces = updatedNamespaces
			if err := cli.Update(context.TODO(), &operatorGroup); err != nil {
				return fmt.Errorf("failed to update OperatorGroup %s: %v", operatorGroup.Name, err)
			}
			logger.Info("Removed stale namespace " + namespaceToRemove + " from OperatorGroup" + operatorGroup.Name)
		}
	}
	return nil
}
