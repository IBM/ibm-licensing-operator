#!/bin/bash

# Script to extract resources created by IBM Licensing Operator
# This script installs the operator, waits for resources to be ready,
# and extracts all created resources for Helm chart generation

set -e -o pipefail

# Configuration
NAMESPACE="ibm-licensing"
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
        --set global.licenseAccept=true \
        --set ibmLicensing.spec.features.prometheusQuerySource.enabled=false \
        --set ibmLicensing.spec.features.alerting.enabled=false | kubectl apply -f - || true
    
    # Wait for CRDs to be established
    log_info "Waiting for CRDs to be established..."
    sleep 10
    
    # Second run: Ensure all dependent resources are created (including CR)
    log_info "Second helm template run (ensuring all resources are created)..."
    helm template ibm-licensing-cluster-scoped "${HELM_CHART_PATH}" \
        --namespace "${NAMESPACE}" \
        --set global.licenseAccept=true \
        --set ibmLicensing.spec.features.prometheusQuerySource.enabled=false \
        --set ibmLicensing.spec.features.alerting.enabled=false | kubectl apply -f -
    
    log_info "Helm installation completed"
}

# Wait for deployment to be ready
wait_for_resources() {
    log_info "Waiting for License Service deployment ibm-licensing-service-instance..."
    
    # Poll for deployment creation (12 attempts x 10 seconds = 120 seconds max)
    local attempts=0
    local max_attempts=12
    
    while [ $attempts -lt $max_attempts ]; do
        if kubectl get deployment ibm-licensing-service-instance -n "${NAMESPACE}" &>/dev/null; then
            log_info "Deployment found after $((attempts * 10))s"
            break
        fi
        
        attempts=$((attempts + 1))
        if [ $attempts -lt $max_attempts ]; then
            log_info "Deployment not found yet, waiting 10s... (attempt $attempts/$max_attempts)"
            sleep 10
        fi
    done
    
    # Check if deployment was found
    if [ $attempts -eq $max_attempts ]; then
        log_error "Deployment not created after $((max_attempts * 10))s timeout"
        exit 1
    fi
    
    # Wait for License Service deployment to be ready
    log_info "Waiting for deployment to become available..."
    if ! kubectl wait --for=condition=available --timeout="${TIMEOUT}s" \
        deployment/ibm-licensing-service-instance -n "${NAMESPACE}"; then
        log_error "License Service deployment not ready after ${TIMEOUT}s timeout"
        exit 1
    fi
    
    log_info "License Service deployment is ready"
}

# Clean up runtime and default fields from YAML using kubectl neat
# and additional yq processing for fields that kubectl neat doesn't remove
cleanup_resource() {
    local resource_type="$1"
    
    # First pass: kubectl neat to remove most runtime fields
    # Second pass: Remove additional runtime fields based on resource type using yq
    case "${resource_type}" in
        deployment)
            kubectl neat | yq eval 'del(.metadata.annotations."deployment.kubernetes.io/revision") |
                     del(.spec.progressDeadlineSeconds) |
                     del(.spec.revisionHistoryLimit) |
                     del(.spec.strategy) |
                     del(.spec.template.spec.dnsPolicy) |
                     del(.spec.template.spec.schedulerName) |
                     del(.spec.template.spec.serviceAccount) |
                     del(.spec.template.spec.containers[].terminationMessagePath) |
                     del(.spec.template.spec.containers[].terminationMessagePolicy) |
                     del(.spec.template.spec.initContainers[].terminationMessagePath) |
                     del(.spec.template.spec.initContainers[].terminationMessagePolicy)'
                     # TODO fix error I0526 15:32:45.073015   81401 warnings.go:110] "Warning: spec.template.spec.containers[0].ports[0]: duplicate port definition with spec.template.spec.initContainers[0].ports[0]"
            ;;
        service)
            kubectl neat | yq eval 'del(.metadata.annotations."service.alpha.openshift.io/serving-cert-signed-by") |
                     del(.metadata.annotations."service.beta.openshift.io/serving-cert-signed-by") |
                     del(.spec.clusterIP) |
                     del(.spec.clusterIPs) |
                     del(.spec.ipFamilies) |
                     del(.spec.ipFamilyPolicy)'
            ;;
        route)
            kubectl neat | yq eval 'del(.metadata.annotations."openshift.io/host.generated") |
                     del(.spec.host) |
                     del(.spec.wildcardPolicy)'
            ;;
        secret)
            kubectl neat
            ;;
        *)
            kubectl neat
            ;;
    esac
}

# Extract namespace-scoped resources created by the operator
# Only extracts required resources as specified in required_resources.md
extract_namespace_resources() {
    log_info "Extracting required namespace-scoped resources from ${NAMESPACE}..."
    
    mkdir -p "${OUTPUT_DIR}"
    
    # Define required resources by type and name, format: "resource_type:resource_name"
    local required_resources=(
        "deployment:ibm-licensing-service-instance"
        "service:ibm-licensing-service-instance"
        "route:ibm-licensing-service-instance"
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
            log_warn "${resource_type}/${resource_name} not found, skipping..."
            continue
        fi
        
        # Extract the resource and clean it up
        local output_file="${OUTPUT_DIR}/${resource_type}-${resource_name}.yaml"
        kubectl get "${resource_type}" "${resource_name}" -n "${NAMESPACE}" -o yaml | cleanup_resource "${resource_type}" > "${output_file}"
        
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