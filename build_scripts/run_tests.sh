#!/usr/bin/env bash
#
# Copyright 2023 IBM Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
set -euo pipefail


echo "Build:"
make build

echo "Create cluster for Scorecard tests"
./build_scripts/create_cluster.sh

echo "Install OLM:"
kubectl apply --server-side -f https://github.com/operator-framework/operator-lifecycle-manager/releases/download/v0.41.0/crds.yaml
kubectl apply --server-side -f https://github.com/operator-framework/operator-lifecycle-manager/releases/download/v0.41.0/olm.yaml

echo "Deploy Operators YAML:"
kubectl create namespace "${LICENSING_NAMESPACE}"
n=0; until ((n >= 60)); do kubectl -n "${LICENSING_NAMESPACE}" get serviceaccount default -o name && break; n=$((n + 1)); sleep 60; done; ((n < 60))
kubectl create secret docker-registry my-registry-token -n "${LICENSING_NAMESPACE}" --docker-server=docker-na-public.artifactory.swg-devops.com --docker-username="${ARTIFACTORY_USERNAME}" --docker-password="${ARTIFACTORY_TOKEN}"
kubectl apply -f ./bundle/manifests/operator.ibm.com_ibmlicensings.yaml
kubectl apply -f ./config/rbac/service_account.yaml -n "${LICENSING_NAMESPACE}"
kubectl -n "${LICENSING_NAMESPACE}" patch serviceaccount ibm-licensing-operator -p '{"imagePullSecrets": [{"name": "my-registry-token"}]}'
kubectl apply -f ./config/rbac/role.yaml
kubectl apply -f ./config/rbac/role_binding.yaml
kubectl apply -f ./config/rbac/role_operands.yaml
kubectl get sa -n "${LICENSING_NAMESPACE}"

echo "Run Scorecard tests:"
make scorecard 2>&1 | tee ./scorecard_logs.txt

echo "Test Unit Operator - License Service:"
export SUFIX=$RANDOM
export USE_EXISTING_CLUSTER=true
export KUBECONFIG=$HOME/.kube/config
echo "Using KUBECONFIG: $KUBECONFIG"
grep "server:" "$KUBECONFIG"
make unit-test 2>&1 | tee ./unittest_logs.txt

echo "Check all pods"
kubectl describe pods --all-namespaces  > ./pods.txt 2>&1
