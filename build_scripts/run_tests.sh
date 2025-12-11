#!/usr/bin/env bash
set -euo pipefail


echo "Build:"
make build

echo "Create cluster for Scorecard tests"
# kind create cluster --image kindest/node:v1.24.15
# kind get clusters
# kubectl config get-contexts
# kubectl config set-context kind-kind
./build_scripts/create_cluster.sh

echo "Install OLM:"
kubectl apply -f https://github.com/operator-framework/operator-lifecycle-manager/releases/download/v0.19.1/crds.yaml
kubectl apply -f https://github.com/operator-framework/operator-lifecycle-manager/releases/download/v0.19.1/olm.yaml

echo "Deploy Operators YAML:"
kubectl create namespace ${LICENSING_NAMESPACE}
n=0; until ((n >= 60)); do kubectl -n ${LICENSING_NAMESPACE} get serviceaccount default -o name && break; n=$((n + 1)); sleep 60; done; ((n < 60))
kubectl create secret docker-registry my-registry-token -n ${LICENSING_NAMESPACE} --docker-server=docker-na-public.artifactory.swg-devops.com --docker-username=${ARTIFACTORY_USERNAME} --docker-password=${ARTIFACTORY_TOKEN}
kubectl apply -f ./bundle/manifests/operator.ibm.com_ibmlicensings.yaml
kubectl apply -f ./config/rbac/service_account.yaml -n ${LICENSING_NAMESPACE}
kubectl -n ${LICENSING_NAMESPACE} patch serviceaccount ibm-licensing-operator -p '{"imagePullSecrets": [{"name": "my-registry-token"}]}'
kubectl apply -f ./config/rbac/role.yaml
kubectl apply -f ./config/rbac/role_binding.yaml
kubectl apply -f ./config/rbac/role_operands.yaml
kubectl get sa -n ${LICENSING_NAMESPACE}

echo "Run Scorecard tests:"
export PATH=`pwd`:$PATH
set -o pipefail
make scorecard 2>&1 | tee ./scorecard_logs.txt

echo "Create cluster for unit tests:"
# cp ./common/scripts/tests/kind_config.yaml ./
# kind create cluster --image kindest/node:${{ matrix.k8s }} --config ./kind_config.yaml --name tests
# kubectl config set-context kind-tests        
# kubectl get nodes
./build_scripts/create_cluster.sh

echo "Test Unit Operator - License Service:"
export PATH=`pwd`:$PATH
export SUFIX=$RANDOM
export USE_EXISTING_CLUSTER=true
set -o pipefail
make unit-test 2>&1 | tee ./unittest_logs_${{ matrix.k8s }}.txt

echo "Check all pods"
export PATH=`pwd`:$PATH
kubectl config set-context kind-tests
kubectl describe pods --all-namespaces  > ./pods_${{ matrix.k8s }}.txt 2>&1
