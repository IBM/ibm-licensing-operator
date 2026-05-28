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

# Script to build Helm chart for IBM Licensing Service (no operator)
# This script orchestrates the entire process:
# 1. Extract resources from a cluster
# 2. Generate RBAC and CRD resources using kustomize
# 3. Template resources into Helm templates
# 4. Clean up temporary resources directory

set -e -o pipefail

# Create path to resources directory based on script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../../.." && pwd)"
RESOURCES_DIR="${PROJECT_ROOT}/resources"

# Source shared logging utilities
source "${SCRIPT_DIR}/logging.sh"

# Main execution
main() {    
    log_info "Starting Helm chart without operator build process..."
    log_info "Project root: ${PROJECT_ROOT}"
    log_info "Resources directory: ${RESOURCES_DIR}"
    echo ""
    
    # Step 1: Extract cluster resources
    log_step "Step 1/4: Extracting resources from cluster..."
    if ! bash "${SCRIPT_DIR}/extract-cluster-resources.sh"; then
        log_error "Failed to extract cluster resources"
        exit 1
    fi
    echo ""
    
    # Step 2: Generate RBAC and CRD resources
    log_step "Step 2/4: Generating RBAC and CRD resources..."
    if ! bash "${SCRIPT_DIR}/generate-resources.sh"; then
        log_error "Failed to generate RBAC and CRD resources"
        exit 1
    fi
    echo ""
    
    # Step 3: Template resources into Helm templates
    log_step "Step 3/4: Templating resources into Helm templates..."
    if ! bash "${SCRIPT_DIR}/template-resources.sh"; then
        log_error "Failed to template resources"
        exit 1
    fi
    echo ""
    
    # Step 4: Clean up resources directory
    log_step "Step 4/4: Cleaning up temporary resources directory..."
    if [ -d "${RESOURCES_DIR}" ]; then
        log_info "Removing ${RESOURCES_DIR}..."
        rm -rf "${RESOURCES_DIR}"
        log_info "Resources directory removed successfully"
    else
        log_warn "Resources directory not found, skipping cleanup"
    fi
    echo ""
    
    # Success message
    echo "=========================================="
    log_info "✓ Helm chart build completed successfully!"
    echo "=========================================="
    echo ""
    log_info "Generated Helm templates are available in:"
    log_info "  ${PROJECT_ROOT}/helm-no-operator/templates/"
}

# Run main function
main "$@"