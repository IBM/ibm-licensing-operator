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
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/utils/strings/slices"
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

	var namespacesToRemove []string
	for _, ns := range watchNamespaces {
		if namespaceActive, err := namespaceActive(reader, ns); err != nil {
			logger.Error(err, "Failed to check namespace existence: "+ns)
			return
		} else if !namespaceActive {
			namespacesToRemove = append(namespacesToRemove, ns)
			logger.Info("Namespace does not exist or is terminating: " + ns + " Namespace marked for removal.")
		}
	}
	if err = removeNamespaceFromOperatorGroup(logger, client, reader, operatorNamespace, namespacesToRemove); err != nil {
		logger.Error(err, "Failed to remove stale namespaces from OperatorGroup")
	}
}

/*
Checks if namespace with given name exits and is active in the cluster.
*/
func namespaceActive(reader client.Reader, ns string) (bool, error) {
	namespace := &corev1.Namespace{}
	if err := reader.Get(context.Background(), client.ObjectKey{Name: ns}, namespace); err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return namespace.Status.Phase == corev1.NamespaceActive, nil
}

/*
Looks for OperatorGroup from operator's namespace.
Removes given namespace from targetNamespaces list field if it contains the namespace.
*/
func removeNamespaceFromOperatorGroup(logger *logr.Logger, cli client.Client, reader client.Reader, namespace string, namespacesToRemove []string) error {
	licensingOperatorGroup, err := res.GetLicensingOperatorGroupInNamespace(reader, namespace)
	if err != nil {
		logger.Error(err, "An error occurred while retrieving IBMLicensing OperatorGroup")
		return err
	} else if licensingOperatorGroup == nil {
		logger.Info("OperatorGroup not found in namespace " + namespace)
		return nil
	}

	targetNamespaces := licensingOperatorGroup.Spec.TargetNamespaces
	var updatedNamespaces []string

	for _, ns := range targetNamespaces {
		if !slices.Contains(namespacesToRemove, ns) {
			updatedNamespaces = append(updatedNamespaces, ns)
		}
	}

	if len(updatedNamespaces) == len(targetNamespaces) {
		return nil
	}

	licensingOperatorGroup.Spec.TargetNamespaces = updatedNamespaces
	if err := cli.Update(context.Background(), licensingOperatorGroup); err != nil {
		return fmt.Errorf("failed to update OperatorGroup %s: %v", licensingOperatorGroup.Name, err)
	}
	logger.Info("Removed stale namespaces from OperatorGroup: " + licensingOperatorGroup.Name)

	return nil
}
