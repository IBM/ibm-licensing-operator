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

# Script to generate RBAC resources and CRDs for IBM Licensing Service instance (operand)
# This script uses kustomize to build resources and extracts only instance-related RBAC and CRDs

set -e -o pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../../.." && pwd)"
OUTPUT_DIR="${OUTPUT_DIR:-${PROJECT_ROOT}/resources}"

# Tool paths (use from LOCALBIN if available, otherwise from PATH)
KUSTOMIZE="${KUSTOMIZE:-${PROJECT_ROOT}/bin/kustomize}"
YQ="${YQ:-${PROJECT_ROOT}/bin/yq}"

# Source shared logging utilities
source "${SCRIPT_DIR}/logging.sh"

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    if [ ! -x "${KUSTOMIZE}" ]; then
        log_error "kustomize not found at ${KUSTOMIZE}"
        log_error "Run 'make install-kustomize' to install it"
        exit 1
    fi
    
    if [ ! -x "${YQ}" ]; then
        log_error "yq not found at ${YQ}"
        log_error "Run 'make install-yq' to install it"
        exit 1
    fi
    
    log_info "Prerequisites check passed"
}

# Generate RBAC resources and CRDs
generate_rbac() {
    log_info "Generating instance RBAC resources and CRDs..."
    
    mkdir -p "${OUTPUT_DIR}"
    
    # Build with kustomize and save to temp file
    "${KUSTOMIZE}" build "${PROJECT_ROOT}/config/manifests" > "${OUTPUT_DIR}/tmp-resources.yaml"
    
    # Extract ClusterRole and ClusterRoleBinding for ibm-license-service
    (echo "---" && "${YQ}" 'select((.kind == "ClusterRole" or .kind == "ClusterRoleBinding") and .metadata.name == "ibm-license-service")' "${OUTPUT_DIR}/tmp-resources.yaml") > "${OUTPUT_DIR}/cluster-rbac.yaml"
    
    # Extract Role and RoleBinding for ibm-license-service
    (echo "---" && "${YQ}" 'select((.kind == "Role" or .kind == "RoleBinding") and .metadata.name == "ibm-license-service")' "${OUTPUT_DIR}/tmp-resources.yaml") > "${OUTPUT_DIR}/rbac.yaml"
    
    # Extract ServiceAccount for ibm-license-service
    (echo "---" && "${YQ}" 'select(.kind == "ServiceAccount" and .metadata.name == "ibm-license-service")' "${OUTPUT_DIR}/tmp-resources.yaml") > "${OUTPUT_DIR}/serviceaccounts.yaml"
    
    # Extract CustomResourceDefinitions (excluding IBMLicensing CRD, because that CRD is only used by operator)
    (echo "---" && "${YQ}" 'select(.kind == "CustomResourceDefinition" and .metadata.name != "ibmlicensings.operator.ibm.com")' "${OUTPUT_DIR}/tmp-resources.yaml") > "${OUTPUT_DIR}/crds.yaml"
    
    # Clean up temp file
    rm -f "${OUTPUT_DIR}/tmp-resources.yaml"
    
    log_info "RBAC resources and CRDs generated successfully"
}

# Main execution
main() {
    log_info "Starting instance RBAC and CRD generation..."
    log_info "Output directory: ${OUTPUT_DIR}"
    
    # Execute steps
    check_prerequisites
    generate_rbac
    
    log_info ""
    log_info "Instance RBAC and CRD generation completed!"
}

# Run main function
main "$@"
