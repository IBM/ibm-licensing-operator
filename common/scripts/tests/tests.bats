#!/usr/bin/env bats
#
# Copyright 2020 IBM Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http:#www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

@test "Create namespace ibm-common-services" {
  kubectl create namespace ibm-common-services
  [ "$?" -eq 0 ]
}

@test "Build Operator" {
  make build
  [ "$?" -eq 0 ]
}

@test "Apply CRD and RBAC" {
  kubectl apply -f ./deploy/crds/operator.ibm.com_ibmlicenseservicereporters_crd.yaml
  [ "$?" -eq 0 ]

  kubectl apply -f ./deploy/crds/operator.ibm.com_ibmlicensings_crd.yaml
  [ "$?" -eq 0 ]

  kubectl apply -f ./deploy/service_account.yaml -n ibm-common-services
  [ "$?" -eq 0 ]

  kubectl apply -f ./deploy/role.yaml
  [ "$?" -eq 0 ]

  kubectl apply -f ./deploy/role_binding.yaml 
  [ "$?" -eq 0 ]
}

@test "Run Operator in backgroud" {
  ./operator-sdk run --watch-namespace ibm-common-services --local > ./operator-sdk_logs.txt 2>&1 &
}

@test "List all POD in cluster" {
  results="$(kubectl get pods --all-namespaces | wc -l)"
  [ "$results" -gt 0 ]
}


@test "Wait 10s for checking pod in ibm-common-services. List should be empty" {
  sleep 10
  kubectl get pods -n ibm-common-services
  results="$(kubectl get pods -n ibm-common-services | wc -l)"
  [ "$results" -eq "0" ]
}


@test "Load CR for LS" {
cat <<EOF | kubectl apply -f -
  apiVersion: operator.ibm.com/v1alpha1
  kind: IBMLicensing
  metadata:
    name: instance
  spec:
    apiSecretToken: ibm-licensing-token
    datasource: datacollector
    httpsEnable: true
    instanceNamespace: ibm-common-services
EOF
  [ "$?" -eq "0" ]
}

@test "Wait 180s for checking pod in ibm-common-services. List should be one POD" {
  sleep 180
  kubectl get pods -n ibm-common-services | grep ibm-licensing-service-instance
  results="$(kubectl get pods -n ibm-common-services | grep ibm-licensing-service-instance | wc -l)"
  [ $results -eq "1" ]

  results="$(kubectl get pods -n ibm-common-services | grep ibm-licensing-service-instance |grep Running |grep '1/1'| wc -l)"
  [ $results -eq "1" ]

}

@test "Remove CR from IBMLicensing" {
  kubectl delete IBMLicensing --all
  [ $? -eq 0 ]
}

@test "Wait 120s for checking pod in ibm-common-services. List should be empty" {
  sleep 120
  kubectl get pods -n ibm-common-services
  results="$(kubectl get pods -n ibm-common-services | grep ibm-licensing-service-instance | wc -l)"
  [ $results -eq "0" ]
}
