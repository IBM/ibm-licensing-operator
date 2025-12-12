#!/usr/bin/env bash
set -euo pipefail


echo "Build:"
make build

echo "Create cluster for Scorecard tests"
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
make scorecard 2>&1 | tee ./scorecard_logs.txt

# echo "Create cluster for unit tests:"
# ./build_scripts/create_cluster.sh

echo "Test Unit Operator - License Service:"
export SUFIX=$RANDOM
export USE_EXISTING_CLUSTER=true
export KUBECONFIG=$HOME/.kube/config  # Explicitly set KUBECONFIG
echo "Using KUBECONFIG: $KUBECONFIG"
cat $KUBECONFIG | grep "server:"  # Verify it's correct
make unit-test 2>&1 | tee ./unittest_logs.txt

echo "Check all pods"
kubectl config set-context kind-tests
kubectl describe pods --all-namespaces  > ./pods.txt 2>&1
