#!/bin/bash

# Script to extract resources created by IBM Licensing Operator
# This script installs the operator, waits for resources to be ready,
# and extracts all created resources for Helm chart generation

set -e -o pipefail

# Configuration
NAMESPACE="${NAMESPACE:-ibm-licensing}"
HELM_CHART_PATH="deploy/argo-cd/components/license-service/helm-cluster-scoped"
OUTPUT_DIR="resources"
# Label selector for License Service resources created by the operator
LABEL_SELECTOR="app.kubernetes.io/instance=ibm-licensing-service"
TIMEOUT=300  # 5 minutes timeout for resource readiness

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
    log_info "Waiting for operator deployment to be ready..."
    
    # Wait for operator deployment
    if ! kubectl wait --for=condition=available --timeout="${TIMEOUT}s" \
        deployment -n "${NAMESPACE}" -l app.kubernetes.io/name=ibm-licensing-operator 2>/dev/null; then
        log_warn "Operator deployment not found or not ready yet, continuing..."
    fi
    
    # Wait a bit more for the operator to reconcile and create operand resources
    log_info "Waiting for operator to reconcile and create operand resources..."
    sleep 30
    
    # Wait for License Service deployment using label selector
    log_info "Waiting for License Service deployment with label ${LABEL_SELECTOR}..."
    local retries=0
    local max_retries=10
    
    while [ $retries -lt $max_retries ]; do
        if kubectl get deployment -n "${NAMESPACE}" -l "${LABEL_SELECTOR}" 2>/dev/null | grep -q "ibm-licensing-service"; then
            log_info "License Service deployment found, waiting for it to be ready..."
            if kubectl wait --for=condition=available --timeout="${TIMEOUT}s" \
                deployment -n "${NAMESPACE}" -l "${LABEL_SELECTOR}" 2>/dev/null; then
                log_info "License Service deployment is ready"
                break
            fi
        fi
        
        retries=$((retries + 1))
        log_info "Waiting for License Service deployment... (attempt $retries/$max_retries)"
        sleep 15
    done
    
    if [ $retries -eq $max_retries ]; then
        log_warn "License Service deployment not ready after ${max_retries} attempts, but continuing with extraction..."
    fi
}

# Clean up runtime fields from YAML
cleanup_resource() {
    local input_file="$1"
    local output_file="$2"
    
    # Use yq if available, otherwise use kubectl
    if command -v yq &> /dev/null; then
        yq eval 'del(.metadata.uid, 
                     .metadata.resourceVersion, 
                     .metadata.generation, 
                     .metadata.creationTimestamp, 
                     .metadata.selfLink, 
                     .metadata.managedFields,
                     .metadata.ownerReferences,
                     .status)' "${input_file}" > "${output_file}"
    else
        # Fallback to kubectl with manual cleanup
        kubectl create --dry-run=client -f "${input_file}" -o yaml > "${output_file}" 2>/dev/null || cp "${input_file}" "${output_file}"
    fi
}

# Extract namespace-scoped resources created by the operator
extract_namespace_resources() {
    log_info "Extracting namespace-scoped resources from ${NAMESPACE}..."
    log_info "Using label selector: ${LABEL_SELECTOR}"
    
    mkdir -p "${OUTPUT_DIR}/namespace"
    
    # Resource types to extract
    local resource_types=(
        "deployment"
        "service"
        "serviceaccount"
        "role"
        "rolebinding"
        "configmap"
        "secret"
        "route"
        "servicemonitor"
        "ingress"
    )
    
    for resource_type in "${resource_types[@]}"; do
        log_info "Extracting ${resource_type}s with label ${LABEL_SELECTOR}..."
        
        # Get all resources of this type with the label selector
        local resources
        resources=$(kubectl get "${resource_type}" -n "${NAMESPACE}" \
            -l "${LABEL_SELECTOR}" -o name 2>/dev/null || echo "")
        
        if [ -z "${resources}" ]; then
            log_info "No ${resource_type}s found with label ${LABEL_SELECTOR}"
            continue
        fi
        
        # Extract each resource
        while IFS= read -r resource; do
            if [ -n "${resource}" ]; then
                local resource_name
                resource_name=$(echo "${resource}" | cut -d'/' -f2)
                local output_file="${OUTPUT_DIR}/namespace/${resource_type}-${resource_name}.yaml"
                
                log_info "Extracting ${resource}..."
                kubectl get "${resource}" -n "${NAMESPACE}" -o yaml > "${output_file}.tmp"
                cleanup_resource "${output_file}.tmp" "${output_file}"
                rm -f "${output_file}.tmp"
                
                log_info "Saved to ${output_file}"
            fi
        done <<< "${resources}"
    done
    
    log_info ""
    log_info "Note: ConfigMaps with 'owner=ibm-licensing' label are NOT extracted"
    log_info "These are created by the License Service deployment itself, not by the operator"
}

# Main execution
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
