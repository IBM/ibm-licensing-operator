#!/bin/bash

# Script to extract resources created by IBM Licensing Operator
# This script installs the operator, waits for resources to be ready,
# and extracts all created resources for Helm chart generation

set -e -o pipefail

# Configuration
NAMESPACE="${NAMESPACE:-ibm-licensing}"
HELM_CHART_PATH="deploy/argo-cd/components/license-service/helm-cluster-scoped"
OUTPUT_DIR="resources"
TIMEOUT=300  # 5 minutes timeout for resource readiness

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
    
    if ! kubectl neat --help &> /dev/null; then
        log_error "kubectl neat plugin is not installed"
        log_error "Install it with: kubectl krew install neat"
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
    kubectl create namespace "${NAMESPACE}" --dry-run=client -o yaml | kubectl apply -f -
}

# Install IBM Licensing Service using Helm
install_licensing_helm() {
    log_info "Installing IBM Licensing Service using Helm..."
    
    if [ ! -d "${HELM_CHART_PATH}" ]; then
        log_error "Helm chart not found at ${HELM_CHART_PATH}"
        exit 1
    fi
    
    # First run: Create CRDs and initial resources
    # Note: This may fail for the CR because CRDs aren't ready yet - this is expected
    log_info "First helm template run (creating CRDs and initial resources)..."
    helm template ibm-licensing-cluster-scoped "${HELM_CHART_PATH}" \
        --namespace "${NAMESPACE}" \
        --set global.licenseAccept=true | kubectl apply -f - || true
    
    # Wait for CRDs to be established
    log_info "Waiting for CRDs to be established..."
    sleep 10
    
    # Second run: Ensure all dependent resources are created (including CR)
    log_info "Second helm template run (ensuring all resources are created)..."
    helm template ibm-licensing-cluster-scoped "${HELM_CHART_PATH}" \
        --namespace "${NAMESPACE}" \
        --set global.licenseAccept=true | kubectl apply -f -
    
    log_info "Helm installation completed"
}

# Wait for deployment to be ready
wait_for_resources() {
    log_info "Waiting for License Service deployment ibm-licensing-service-instance..."
    
    # Wait for License Service deployment (operator will create it)
    if ! kubectl wait --for=condition=available --timeout="${TIMEOUT}s" \
        deployment/ibm-licensing-service-instance -n "${NAMESPACE}" 2>/dev/null; then
        log_error "License Service deployment not ready after ${TIMEOUT}s timeout"
        exit 1
    fi
    
    log_info "License Service deployment is ready"
}

# Clean up runtime and default fields from YAML using kubectl neat
cleanup_resource() {
    local input_file="$1"
    local output_file="$2"
    
    kubectl neat -f "${input_file}" > "${output_file}"
}

# Extract namespace-scoped resources created by the operator
# Only extracts required resources as specified in required_resources.md
extract_namespace_resources() {
    log_info "Extracting required namespace-scoped resources from ${NAMESPACE}..."
    
    mkdir -p "${OUTPUT_DIR}/cluster"
    
    # Define required resources by type and name (from required_resources.md)
    # Format: "resource_type:resource_name"
    local required_resources=(
        # Core Components
        "deployment:ibm-licensing-service-instance"
        "service:ibm-licensing-service-instance"
        "route:ibm-licensing-service-instance"
        
        # Secrets (required for deployment)
        "secret:ibm-license-service-cert"
        "secret:ibm-licensing-token"
        "secret:ibm-licensing-upload-token"
    )
    
    log_info "Extracting ${#required_resources[@]} required resources..."
    
    for resource_spec in "${required_resources[@]}"; do
        # Skip comments
        if [[ "${resource_spec}" =~ ^[[:space:]]*# ]]; then
            continue
        fi
        
        # Parse resource type and name
        local resource_type
        local resource_name
        resource_type=$(echo "${resource_spec}" | cut -d':' -f1)
        resource_name=$(echo "${resource_spec}" | cut -d':' -f2)
        
        log_info "Extracting ${resource_type}/${resource_name}..."
        
        # Check if resource exists
        if ! kubectl get "${resource_type}" "${resource_name}" -n "${NAMESPACE}" &>/dev/null; then
            log_warn "${resource_type}/${resource_name} not found, skipping..."
            continue
        fi
        
        # Extract the resource and clean it up with kubectl neat
        local output_file="${OUTPUT_DIR}/cluster/${resource_type}-${resource_name}.yaml"
        kubectl get "${resource_type}" "${resource_name}" -n "${NAMESPACE}" -o yaml | kubectl neat > "${output_file}"
        
        log_info "Saved to ${output_file}"
    done
    
    log_info ""
    log_info "Note: Only required resources from required_resources.md are extracted"
    log_info "RBAC resources will be sourced from Kustomize (operand RBAC only)"
    log_info "ConfigMaps are currently skipped (not mounted in deployment)"
    log_info "Prometheus-related resources are skipped (feature not supported yet)"
}

# ============================================================================
# Logging utilities
# ============================================================================

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
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
    log_info ""
    log_info "Next steps:"
    log_info "1. Review extracted resources in ${OUTPUT_DIR}/"
    log_info "2. Run templatization script to convert to Helm templates"
    log_info "3. Test the generated Helm chart"
}

# Run main function
main "$@"

# Made with Bob
