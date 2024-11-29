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

package controllers

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/utils/strings/slices"
	c "sigs.k8s.io/controller-runtime/pkg/client"

	res "github.com/IBM/ibm-licensing-operator/controllers/resources"

	odlm "github.com/IBM/operand-deployment-lifecycle-manager/api/v1alpha1"
)

// Looks for OperandRequests (that have binding for ibm-licensing-operator) in other namespaces

// +kubebuilder:rbac:namespace=ibm-licensing,groups=operators.coreos.com,resources=operatorgroups,verbs=get;list;patch;update;watch

func DiscoverOperandRequests(logger *logr.Logger, writer c.Writer, reader c.Reader, watchNamespace []string, namespaceScopeSemaphore chan bool) {
	var nssEnabled, prevNssEnabledState bool
	var operandRequestList odlm.OperandRequestList
	var namespaceListToExtend []string

	operatorNamespace, err := res.GetOperatorNamespace()
	if err != nil {
		logger.Error(err, "Could not retrieve operator namespace. Discovering OperandRequests will be disabled")
		return
	}

	// synchronize with reconciliation of IBMLicensing instance
	nssEnabled = <-namespaceScopeSemaphore

	for {
		prevNssEnabledState = nssEnabled
		select {
		case nssEnabled = <-namespaceScopeSemaphore:
			if nssEnabled != prevNssEnabledState {
				if nssEnabled {
					logger.Info("Namespace scope enabled. Cluster-wide discovering OperandRequests disabled")
				} else {
					logger.Info("Namespace scope disabled. Cluster-wide discovering OperandRequests enabled")
				}
			}
		default:
		}

		if nssEnabled {
			time.Sleep(30 * time.Second)
			continue
		}

		operandRequestList = odlm.OperandRequestList{}
		err := reader.List(context.TODO(), &operandRequestList)
		if err != nil {
			logger.Error(err, "Could not list OperandRequests from cluster")
		}

		namespaceListToExtend = []string{}
		for _, operandRequest := range operandRequestList.Items {
			if !isOperandRequestNamespaceValid(logger, reader, operandRequest) {
				continue
			}
			if hasBinding := res.HasOperandRequestBindingForLicensing(operandRequest); hasBinding {
				if !slices.Contains(watchNamespace, operandRequest.Namespace) && !slices.Contains(namespaceListToExtend, operandRequest.Namespace) {
					logger.Info("OperandRequest for "+res.OperatorName+" detected. IBMLicensing OperatorGroup will be extended", "OperandRequest", operandRequest.Name, "Namespace", operandRequest.Namespace)
					namespaceListToExtend = append(namespaceListToExtend, operandRequest.Namespace)
				}
			}
		}

		if len(namespaceListToExtend) > 0 {
			licensingOperatorGroup, err := res.GetLicensingOperatorGroupInNamespace(reader, operatorNamespace)
			if err != nil {
				logger.Error(err, "An error occurred while retrieving IBMLicensing OperatorGroup")
			} else if licensingOperatorGroup != nil {
				logger.Info("Extending IBMLicensing OperatorGroup with namespaces", "OperatorGroup", licensingOperatorGroup.Name, "NamespaceList", namespaceListToExtend)
				licensingOperatorGroup = res.ExtendOperatorGroupWithNamespaceList(namespaceListToExtend, licensingOperatorGroup)
				err := writer.Update(context.TODO(), licensingOperatorGroup)
				if err != nil {
					logger.Error(err, "An error occurred while extending IBMLicensing OperatorGroup", "OperatorGroup", licensingOperatorGroup.Name, "Namespace", operatorNamespace)
				}
			} else {
				logger.Info("OperatorGroup for IBMLicensing operator not found", "Namespace", operatorNamespace)
			}
		}

		time.Sleep(30 * time.Second)
	}
}

/*
It can happen that there is an OperandRequest in a non-existent namespace, which causes LS operator to restart constantly.
Such scenario was found when namespace was force deleted, while OperandRequest had pending finalization due to removed Pod providing finalization logic.
To prevent such case, we are additionally checking if the namespace actually exists when processing OperandRequests
*/
func isOperandRequestNamespaceValid(logger *logr.Logger, reader c.Reader, operandRequest odlm.OperandRequest) bool {
	if namespaceActive, err := namespaceActive(reader, operandRequest.Namespace); err != nil {
		logger.Error(err, "Failed to check namespace existence: "+operandRequest.Namespace)
		return false
	} else if !namespaceActive {
		logger.Info("OperandRequest's namespace was not found in the cluster or is terminating. It will not be added to OperatorGroup", "OperandRequest", operandRequest.Name, "Namespace", operandRequest.Namespace)
		return false
	}
	return true
}
