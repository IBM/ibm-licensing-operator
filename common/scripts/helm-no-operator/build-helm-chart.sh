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
# 1. Extract resources from a running cluster
# 2. Generate RBAC and CRD resources
# 3. Template resources into Helm templates
# 4. Clean up temporary resources directory

set -e -o pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../../.." && pwd)"
RESOURCES_DIR="${PROJECT_ROOT}/resources"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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

log_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

# Print banner
print_banner() {
    echo ""
    echo "=========================================="
    echo "  IBM Licensing Helm Chart Builder"
    echo "=========================================="
    echo ""
}

# Main execution
main() {
    print_banner
    
    log_info "Starting Helm chart build process..."
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
    echo ""
    log_info "Next steps:"
    log_info "  1. Review the generated templates"
    log_info "  2. Update helm-no-operator/values.yaml if needed"
    log_info "  3. Test the Helm chart with: helm template ibm-licensing ./helm-no-operator"
    log_info "  4. Install with: helm install ibm-licensing ./helm-no-operator"
    echo ""
}

# Run main function
main "$@"

# Made with Bob
