#!/bin/bash

#
# Copyright 2026 IBM Corporation
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

# Script to extract resources created by IBM Licensing Operator
# This script installs the operator, waits for resources to be ready,
# and extracts all created resources into yaml files.

set -e -o pipefail

# Create paths to other folders/tools
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../../.." && pwd)"
LOCALBIN="${REPO_ROOT}/bin"
YQ="${LOCALBIN}/yq"

# Configuration
NAMESPACE="ibm-licensing"
HELM_CHART_PATH_CLUSTER_SCOPED="deploy/argo-cd/components/license-service/helm-cluster-scoped"
HELM_CHART_PATH="deploy/argo-cd/components/license-service/helm"
OUTPUT_DIR="resources"

# Source shared logging utilities
# shellcheck source=common/scripts/helm-no-operator/logging.sh
source "${SCRIPT_DIR}/logging.sh"

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl is not installed"
        exit 1
    fi
    
    if ! command -v helm &> /dev/null; then
        log_error "helm is not installed"
        exit 1
    fi
    
    if [ ! -x "${YQ}" ]; then
        log_error "yq not found at $YQ, run 'make install-yq' to install it."
        exit 1
    fi
    
    if ! kubectl cluster-info &> /dev/null; then
        log_error "Cannot connect to Kubernetes cluster"
        exit 1
    fi
    
    log_info "Prerequisites check passed"
}

# Create namespace if it doesn't exist
create_namespace() {
    log_info "Creating namespace ${NAMESPACE}..."
    # Use kubectl apply instead of kubectl create to make this operation idempotent.
    # This approach creates the namespace if it doesn't exist, or updates it if it does,
    # avoiding errors when the namespace already exists.
    kubectl create namespace "${NAMESPACE}" --dry-run=client -o yaml | kubectl apply -f -
}

# Install IBM Licensing Service using Helm
install_licensing_helm() {
    log_info "Installing IBM Licensing Service using Helm..."
    
    if [ ! -d "${HELM_CHART_PATH_CLUSTER_SCOPED}" ]; then
        log_error "Helm chart not found at ${HELM_CHART_PATH_CLUSTER_SCOPED}"
        exit 1
    fi

    if [ ! -d "${HELM_CHART_PATH}" ]; then
        log_error "Helm chart not found at ${HELM_CHART_PATH}"
        exit 1
    fi
    
    # Step 1: Install cluster-scoped resources (CRDs, ClusterRoles)
    # Note: This may fail for the CR because CRDs aren't ready yet - this is expected
    log_info "Installing cluster-scoped resources (CRDs, ClusterRoles)..."
    helm template ibm-licensing-cluster-scoped "${HELM_CHART_PATH_CLUSTER_SCOPED}" \
        --set ibmLicensing.spec.features.prometheusQuerySource.enabled=false \
        --set ibmLicensing.spec.features.alerting.enabled=false | kubectl apply -f - || true
    
    # Wait for CRDs to be established
    log_info "Waiting for CRDs to be established..."
    sleep 10
    
    # Step 2: Install namespace-scoped resources (Deployment, ServiceAccounts, Roles, CR)
    log_info "Installing namespace-scoped resources (Deployment, RBAC, CR)..."
    helm template ibm-licensing "${HELM_CHART_PATH}" \
        --namespace "${NAMESPACE}" \
        --set ibmLicensing.spec.features.prometheusQuerySource.enabled=false \
        --set ibmLicensing.spec.features.alerting.enabled=false | kubectl apply -f -
    
    log_info "Helm installation completed"
}

# Wait for deployment to be ready
wait_for_resources() {
    log_info "Waiting for License Service deployment ibm-licensing-service-instance..."
    
    # Poll for deployment creation (12 attempts with 10 second delay between retries)
    local attempts=1
    local max_attempts=12
    local wait_interval=10
    
    while [ $attempts -lt $max_attempts ]; do
        if kubectl get deployment ibm-licensing-service-instance -n "${NAMESPACE}" &>/dev/null; then
            log_info "Deployment found after $(((attempts - 1) * wait_interval))s"
            break
        fi
        
        log_info "Deployment not found yet, waiting ${wait_interval}s... (attempt $attempts/$max_attempts)"
        sleep "${wait_interval}"
        attempts=$((attempts + 1))
    done
    
    # Check if deployment was found
    if [ $attempts -eq $max_attempts ]; then
        log_error "Deployment not created after $((max_attempts * wait_interval))s timeout"
        exit 1
    fi
    
    # Wait for License Service deployment to be ready
    local TIMEOUT=150  # timeout for deployment readiness
    log_info "Waiting for deployment to become available..."
    if ! kubectl wait --for=condition=available --timeout="${TIMEOUT}s" \
        deployment/ibm-licensing-service-instance -n "${NAMESPACE}"; then
        log_error "License Service deployment not ready after ${TIMEOUT}s timeout"
        exit 1
    fi
    
    log_info "License Service deployment is ready"
}

# Clean up runtime and default fields from YAML
cleanup_resource() {
    local resource_type="$1"
    
    # Common deletions for all resources (applied first)
    local common_deletes='del(.metadata.creationTimestamp) |
                          del(.metadata.generation) |
                          del(.metadata.resourceVersion) |
                          del(.metadata.uid) |
                          del(.metadata.ownerReferences) |
                          del(.status)'
    
    # Apply common deletions, then resource-specific deletions
    case "${resource_type}" in
        deployment)
            "${YQ}" eval "${common_deletes} |
                     del(.metadata.annotations.\"deployment.kubernetes.io/revision\") |
                     del(.spec.progressDeadlineSeconds) |
                     del(.spec.revisionHistoryLimit) |
                     del(.spec.strategy) |
                     del(.spec.template.spec.dnsPolicy) |
                     del(.spec.template.spec.schedulerName) |
                     del(.spec.template.spec.serviceAccount) |
                     del(.spec.template.spec.containers[].terminationMessagePath) |
                     del(.spec.template.spec.containers[].terminationMessagePolicy) |
                     del(.spec.template.spec.initContainers[].terminationMessagePath) |
                     del(.spec.template.spec.initContainers[].terminationMessagePolicy) |
                     del(.spec.template.spec.initContainers[].ports)"
            ;;
        service)
            "${YQ}" eval "${common_deletes} |
                     del(.metadata.annotations.\"service.alpha.openshift.io/serving-cert-signed-by\") |
                     del(.metadata.annotations.\"service.beta.openshift.io/serving-cert-signed-by\") |
                     del(.spec.clusterIP) |
                     del(.spec.clusterIPs) |
                     del(.spec.internalTrafficPolicy) |
                     del(.spec.ipFamilies) |
                     del(.spec.ipFamilyPolicy) |
                     del(.spec.sessionAffinity)"
            ;;
        *)
            "${YQ}" eval "${common_deletes}"
            ;;
    esac
}

# Extract resources created by the operator
extract_namespace_resources() {
    log_info "Extracting required namespace-scoped resources from ${NAMESPACE}..."
    
    mkdir -p "${OUTPUT_DIR}"
    
    # Define required resources by type and name, format: "resource_type:resource_name"
    local required_resources=(
        "deployment:ibm-licensing-service-instance"
        "service:ibm-licensing-service-instance"
        "secret:ibm-licensing-token"
        "secret:ibm-licensing-upload-token"
    )
    
    log_info "Extracting ${#required_resources[@]} required resources..."
    
    for resource_spec in "${required_resources[@]}"; do
        # Parse resource type and name
        local resource_type
        local resource_name
        resource_type=$(echo "${resource_spec}" | cut -d':' -f1)
        resource_name=$(echo "${resource_spec}" | cut -d':' -f2)
        
        log_info "Extracting ${resource_type}/${resource_name}..."
        
        # Check if resource exists
        if ! kubectl get "${resource_type}" "${resource_name}" -n "${NAMESPACE}" &>/dev/null; then
            log_error "${resource_type}/${resource_name} not found"
            exit 1
        fi
        
        # Extract the resource and clean it up
        local output_file="${OUTPUT_DIR}/${resource_type}-${resource_name}.yaml"
        kubectl get "${resource_type}" "${resource_name}" -n "${NAMESPACE}" -o yaml | cleanup_resource "${resource_type}" > "${output_file}"
        log_info "Saved to ${output_file}"
    done
    
    log_info ""
}

# ============================================================================
# Main execution
# ============================================================================

main() {
    log_info "Starting resource extraction process..."
    log_info "Namespace: ${NAMESPACE}"
    log_info "Output directory: ${OUTPUT_DIR}"
    
    # Clean up output directory if it exists
    if [ -d "${OUTPUT_DIR}" ]; then
        log_warn "Output directory ${OUTPUT_DIR} already exists, cleaning up..."
        rm -rf "${OUTPUT_DIR}"
    fi
    
    mkdir -p "${OUTPUT_DIR}"
    
    # Execute steps
    check_prerequisites
    create_namespace
    install_licensing_helm
    wait_for_resources
    extract_namespace_resources
    
    log_info "Resource extraction completed successfully!"
    log_info "Extracted resources are available in: ${OUTPUT_DIR}/"
}

# Run main function
main "$@"