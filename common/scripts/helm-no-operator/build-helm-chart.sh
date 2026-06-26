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
# 2. Generate CRD resources using kustomize
# 3. Template resources into Helm templates
# 4. Clean up temporary resources directory

set -e -o pipefail

# Create paths to other folders/tools
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../../.." && pwd)"
RESOURCES_DIR="${PROJECT_ROOT}/resources"

# Source shared logging utilities
# shellcheck source=common/scripts/helm-no-operator/logging.sh
source "${SCRIPT_DIR}/logging.sh"

# Main execution
main() {    
    log_info "Starting Helm chart without operator build process..."
    log_info "Project root: ${PROJECT_ROOT}"
    log_info "Resources directory: ${RESOURCES_DIR}"
    echo ""
    
    # Step 1: Extract cluster resources
    log_step "Step 1/5: Extracting resources from cluster..."
    if ! bash "${SCRIPT_DIR}/extract-cluster-resources.sh"; then
        log_error "Failed to extract cluster resources"
        exit 1
    fi
    echo ""
    
    # Step 2: Generate CRD resources
    log_step "Step 2/5: Generating CRD resources..."
    if ! bash "${SCRIPT_DIR}/generate-resources.sh"; then
        log_error "Failed to generate CRD resources"
        exit 1
    fi
    echo ""
    
    # Step 3: Template resources into Helm templates
    log_step "Step 3/5: Templating resources into Helm templates..."
    if ! bash "${SCRIPT_DIR}/template-resources.sh"; then
        log_error "Failed to template resources"
        exit 1
    fi
    echo ""
    
    # Step 4: Clean up resources directory
    log_step "Step 4/5: Cleaning up temporary resources directory..."
    if [ -d "${RESOURCES_DIR}" ]; then
        log_info "Removing ${RESOURCES_DIR}..."
        rm -rf "${RESOURCES_DIR}"
        log_info "Resources directory removed successfully"
    else
        log_warn "Resources directory not found, skipping cleanup"
    fi
    echo ""

    # Step 5: Update other resources using IBM Bob
    log_step "Step 5/5: Syncing RBAC from source-of-truth chart..."
    bob -y "Sync cluster-rbac.yaml, rbac-watch-namespace.yaml, rbac.yaml, serviceaccounts.yaml, _helpers.tpl from ./helm-no-operator/templates/ with their counterparts in ./deploy/argo-cd/components/license-service/helm-cluster-scoped/templates/.
    
    Rules:
    1. Keep only RBAC for the operand (instance deployment). Drop every resource whose name contains 'operator' or that is only needed by the operator controller (for example ClusterRole ibm-licensing-operator, Role ibm-licensing-operator, ClusterRole ibm-licensing-opreqs-role, and their bindings).
    2. Preserve all Helm conditionals ({{- if ... }}/{{- end }}) exactly as they appear in the source files.
    3. Do not add or remove any other content."
    
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
