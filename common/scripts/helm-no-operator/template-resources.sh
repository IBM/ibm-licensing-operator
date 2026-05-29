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

# Script to template YAML files into Helm templates.

set -e -o pipefail

# Create paths to other folders/tools
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../../.." && pwd)"
INPUT_DIR="${PROJECT_ROOT}/resources"
OUTPUT_DIR="${PROJECT_ROOT}/helm-no-operator/templates"
TEMP_DIR="${PROJECT_ROOT}/temp"
YQ="${PROJECT_ROOT}/bin/yq"

# Source shared logging utilities
source "${SCRIPT_DIR}/logging.sh"

check_prerequisites() {
    log_info "Checking prerequisites..."
    
    if [ ! -x "$YQ" ]; then
        log_error "yq not found at $YQ. Please run 'make install-yq' first."
        exit 1
    fi
    
    log_info "Prerequisites check passed"
}

setup_directories() {
    log_info "Setting up directories..."
    
    # Create temp directory
    mkdir -p "$TEMP_DIR"
    
    # Create output directory
    mkdir -p "$OUTPUT_DIR"
    
    log_info "Directories created"
}

template_secrets() {
    log_info "Templating secrets..."
    
    # Process first secret (ibm-licensing-token)
    cp "$INPUT_DIR/secret-ibm-licensing-token.yaml" "$TEMP_DIR/secret-ibm-licensing-token.yaml"
    
    # Step 1: Use yq to add placeholder for token
    TOKEN_FIELD=$($YQ '.data | keys | .[0]' "$TEMP_DIR/secret-ibm-licensing-token.yaml")
    $YQ -i ".data.$TOKEN_FIELD = \"sed-me-token\"" "$TEMP_DIR/secret-ibm-licensing-token.yaml"
    
    # Step 2: Use sed to replace placeholders with Helm templates
    sed -i '' "s/namespace: ibm-licensing/namespace: {{ .Values.ibmLicensing.namespace }}/g" "$TEMP_DIR/secret-ibm-licensing-token.yaml"
    sed -i '' "s/sed-me-token/{{ randAlphaNum 24 | b64enc }}/g" "$TEMP_DIR/secret-ibm-licensing-token.yaml"
    
    # Process second secret (ibm-licensing-upload-token)
    cp "$INPUT_DIR/secret-ibm-licensing-upload-token.yaml" "$TEMP_DIR/secret-ibm-licensing-upload-token.yaml"
    
    # Step 1: Use yq to add placeholder for token
    UPLOAD_TOKEN_FIELD=$($YQ '.data | keys | .[0]' "$TEMP_DIR/secret-ibm-licensing-upload-token.yaml")
    $YQ -i ".data.$UPLOAD_TOKEN_FIELD = \"sed-me-upload-token\"" "$TEMP_DIR/secret-ibm-licensing-upload-token.yaml"
    
    # Step 2: Use sed to replace placeholders with Helm templates
    sed -i '' "s/namespace: ibm-licensing/namespace: {{ .Values.ibmLicensing.namespace }}/g" "$TEMP_DIR/secret-ibm-licensing-upload-token.yaml"
    sed -i '' "s/sed-me-upload-token/{{ randAlphaNum 24 | b64enc }}/g" "$TEMP_DIR/secret-ibm-licensing-upload-token.yaml"
    
    # Combine both secrets into final output
    cat "$TEMP_DIR/secret-ibm-licensing-token.yaml" > "$OUTPUT_DIR/secrets.yaml"
    echo "---" >> "$OUTPUT_DIR/secrets.yaml"
    cat "$TEMP_DIR/secret-ibm-licensing-upload-token.yaml" >> "$OUTPUT_DIR/secrets.yaml"
    
    log_info "secrets.yaml created"
}

template_deployment() {
    log_info "Templating deployment..."
    
    cp "$INPUT_DIR/deployment-ibm-licensing-service-instance.yaml" "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
    
    # Step 1: Use yq to add placeholders for values that need templating
    # Replace namespace
    $YQ -i '.metadata.namespace = "sed-me-namespace"' "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
    
    # Replace image
    $YQ -i '.spec.template.spec.containers[0].image = "sed-me-image"' "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
    $YQ -i '.spec.template.spec.initContainers[0].image = "sed-me-image"' "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
    
    # Remove imagePullSecrets (will be added conditionally later)
    $YQ -i 'del(.spec.template.spec.imagePullSecrets)' "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
    
    # Replace environment variables in main container
    $YQ -i '.spec.template.spec.containers[0].env[0].value = "sed-me-namespace"' "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
    $YQ -i '.spec.template.spec.containers[0].env[1].value = "sed-me-datasource"' "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
    $YQ -i '.spec.template.spec.containers[0].env[2].value = "sed-me-httpsEnable"' "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
    $YQ -i '.spec.template.spec.containers[0].env[3].value = "sed-me-enableInstanaMetricCollection"' "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
    $YQ -i '.spec.template.spec.containers[0].env[4].value = "sed-me-httpsCertsSource"' "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
    
    # Replace environment variables in init container
    $YQ -i '.spec.template.spec.initContainers[0].env[0].value = "sed-me-namespace"' "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
    $YQ -i '.spec.template.spec.initContainers[0].env[1].value = "sed-me-datasource"' "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
    $YQ -i '.spec.template.spec.initContainers[0].env[2].value = "sed-me-httpsEnable"' "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
    $YQ -i '.spec.template.spec.initContainers[0].env[3].value = "sed-me-enableInstanaMetricCollection"' "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
    $YQ -i '.spec.template.spec.initContainers[0].env[4].value = "sed-me-httpsCertsSource"' "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
    
    # Replace resource limits and requests in main container
    $YQ -i '.spec.template.spec.containers[0].resources.limits.cpu = "sed-me-cpu-limit"' "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
    $YQ -i '.spec.template.spec.containers[0].resources.limits.memory = "sed-me-memory-limit"' "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
    $YQ -i '.spec.template.spec.containers[0].resources.requests.cpu = "sed-me-cpu-request"' "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
    $YQ -i '.spec.template.spec.containers[0].resources.requests.memory = "sed-me-memory-request"' "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
    $YQ -i '.spec.template.spec.containers[0].resources.requests."ephemeral-storage" = "sed-me-ephemeral-storage-request"' "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
    
    # Replace resource limits and requests in init container
    $YQ -i '.spec.template.spec.initContainers[0].resources.limits.cpu = "sed-me-cpu-limit"' "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
    $YQ -i '.spec.template.spec.initContainers[0].resources.limits.memory = "sed-me-memory-limit"' "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
    $YQ -i '.spec.template.spec.initContainers[0].resources.requests.cpu = "sed-me-cpu-request"' "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
    $YQ -i '.spec.template.spec.initContainers[0].resources.requests.memory = "sed-me-memory-request"' "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
    $YQ -i '.spec.template.spec.initContainers[0].resources.requests."ephemeral-storage" = "sed-me-ephemeral-storage-request"' "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
    
    # Step 2: Use sed to replace placeholders with Helm templates
    # Replace namespace
    sed -i '' "s/namespace: sed-me-namespace/namespace: {{ .Values.ibmLicensing.namespace }}/g" "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
    
    # Replace image
    sed -i '' "s|image: sed-me-image|image: {{ .Values.global.imagePullPrefix }}/{{ .Values.ibmLicensing.imageRegistryNamespaceOperand }}/ibm-licensing:4.2.23|g" "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
    
    # Replace environment variables
    sed -i '' 's/value: sed-me-namespace/value: {{ .Values.ibmLicensing.namespace | quote }}/g' "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
    sed -i '' 's/value: sed-me-datasource/value: {{ .Values.ibmLicensing.datasource | quote }}/g' "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
    sed -i '' 's/value: "sed-me-httpsEnable"/value: {{ .Values.ibmLicensing.httpsEnable | quote }}/g' "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
    sed -i '' 's/value: "sed-me-enableInstanaMetricCollection"/value: {{ .Values.ibmLicensing.enableInstanaMetricCollection | quote }}/g' "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
    sed -i '' 's/value: sed-me-httpsCertsSource/value: {{ .Values.ibmLicensing.httpsCertsSource | quote }}/g' "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
    
    # Replace resource limits and requests
    sed -i '' "s/sed-me-cpu-limit/{{ .Values.ibmLicensing.resources.limits.cpu }}/g" "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
    sed -i '' "s/sed-me-memory-limit/{{ .Values.ibmLicensing.resources.limits.memory }}/g" "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
    sed -i '' "s/sed-me-cpu-request/{{ .Values.ibmLicensing.resources.requests.cpu }}/g" "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
    sed -i '' "s/sed-me-memory-request/{{ .Values.ibmLicensing.resources.requests.memory }}/g" "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
    sed -i '' "s/sed-me-ephemeral-storage-request/{{ .Values.ibmLicensing.resources.requests.ephemeralStorage }}/g" "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
    
    # Append conditional imagePullSecrets section
    cat "${PROJECT_ROOT}/common/makefile-generate/yaml-deployment-pull-secrets-part" >> "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
    
    # Copy to output
    cp "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml" "$OUTPUT_DIR/deployment.yaml"
    
    log_info "deployment.yaml created"
}

template_service() {
    log_info "Templating service..."
    
    cp "$INPUT_DIR/service-ibm-licensing-service-instance.yaml" "$TEMP_DIR/service-ibm-licensing-service-instance.yaml"
    
    # Step 1: Use yq to add placeholder for namespace
    $YQ -i '.metadata.namespace = "sed-me-namespace"' "$TEMP_DIR/service-ibm-licensing-service-instance.yaml"
    
    # Step 2: Use sed to replace placeholder with Helm template
    sed -i '' "s/namespace: sed-me-namespace/namespace: {{ .Values.ibmLicensing.namespace }}/g" "$TEMP_DIR/service-ibm-licensing-service-instance.yaml"
    
    # Copy to output
    cp "$TEMP_DIR/service-ibm-licensing-service-instance.yaml" "$OUTPUT_DIR/service.yaml"
    
    log_info "service.yaml created"
}

template_crds() {
    log_info "Templating CRDs..."
    
    # CRDs are cluster-scoped resources and typically don't need templating
    # Just copy them directly to the output directory
    cp "$INPUT_DIR/crds.yaml" "$OUTPUT_DIR/crds.yaml"
    
    log_info "crds.yaml created"
}

template_serviceaccount() {
    log_info "Templating serviceaccount..."
    
    cp "$INPUT_DIR/serviceaccounts.yaml" "$TEMP_DIR/serviceaccounts.yaml"
    
    # Step 1: Use yq to add placeholder for namespace
    $YQ -i '.metadata.namespace = "sed-me-namespace"' "$TEMP_DIR/serviceaccounts.yaml"
    
    # Step 2: Use sed to replace placeholder with Helm template
    sed -i '' "s/namespace: sed-me-namespace/namespace: {{ .Values.ibmLicensing.namespace }}/g" "$TEMP_DIR/serviceaccounts.yaml"
    
    # Copy to output
    cp "$TEMP_DIR/serviceaccounts.yaml" "$OUTPUT_DIR/serviceaccount.yaml"
    
    log_info "serviceaccount.yaml created"
}

template_rbac() {
    log_info "Templating rbac..."
    
    cp "$INPUT_DIR/rbac.yaml" "$TEMP_DIR/rbac.yaml"
    
    # Step 1: Use yq to add placeholders for namespace in both Role and RoleBinding
    $YQ -i '(select(.kind == "Role") | .metadata.namespace) = "sed-me-namespace"' "$TEMP_DIR/rbac.yaml"
    $YQ -i '(select(.kind == "RoleBinding") | .metadata.namespace) = "sed-me-namespace"' "$TEMP_DIR/rbac.yaml"
    $YQ -i '(select(.kind == "RoleBinding") | .subjects[0].namespace) = "sed-me-namespace"' "$TEMP_DIR/rbac.yaml"
    
    # Step 2: Use sed to replace placeholders with Helm templates
    sed -i '' "s/namespace: sed-me-namespace/namespace: {{ .Values.ibmLicensing.namespace }}/g" "$TEMP_DIR/rbac.yaml"
    
    # Copy to output
    cp "$TEMP_DIR/rbac.yaml" "$OUTPUT_DIR/rbac.yaml"
    
    log_info "rbac.yaml created"
}

template_cluster_rbac() {
    log_info "Templating cluster-rbac..."
    
    cp "$INPUT_DIR/cluster-rbac.yaml" "$TEMP_DIR/cluster-rbac.yaml"
    
    # Step 1: Use yq to add placeholders for namespace in ClusterRoleBinding
    # Note: ClusterRole doesn't have namespace, but ClusterRoleBinding metadata and subjects do
    $YQ -i '(select(.kind == "ClusterRoleBinding") | .metadata.namespace) = "sed-me-namespace"' "$TEMP_DIR/cluster-rbac.yaml"
    $YQ -i '(select(.kind == "ClusterRoleBinding") | .subjects[0].namespace) = "sed-me-namespace"' "$TEMP_DIR/cluster-rbac.yaml"
    
    # Step 2: Use sed to replace placeholders with Helm templates
    sed -i '' "s/namespace: sed-me-namespace/namespace: {{ .Values.ibmLicensing.namespace }}/g" "$TEMP_DIR/cluster-rbac.yaml"
    
    # Copy to output
    cp "$TEMP_DIR/cluster-rbac.yaml" "$OUTPUT_DIR/cluster-rbac.yaml"
    
    log_info "cluster-rbac.yaml created"
}

# Clean up temporary files
cleanup_directories() {
    log_info "Cleaning up temporary files..."
    rm -rf "$TEMP_DIR"
    log_info "Cleanup completed"
}

main() {
    log_info "Starting resource templating process..."
    log_info "Input directory: ${INPUT_DIR}"
    log_info "Output directory: ${OUTPUT_DIR}"
    
    # Execute steps
    check_prerequisites
    setup_directories
    template_secrets
    template_deployment
    template_service
    template_crds
    template_serviceaccount
    template_rbac
    template_cluster_rbac
    cleanup_directories
    
    log_info ""
    log_info "Resource templating completed successfully!"
    log_info "Templated resources are available in: ${OUTPUT_DIR}/"
}

# Run main function
main "$@"
