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
YQ="${PROJECT_ROOT}/bin/yq"

# Source shared logging utilities
# shellcheck source=common/scripts/helm-no-operator/logging.sh
source "${SCRIPT_DIR}/logging.sh"

check_prerequisites() {
    log_info "Checking prerequisites..."
    
    if [ ! -x "$YQ" ]; then
        log_error "yq not found at $YQ, run 'make install-yq' to install it."
        exit 1
    fi
    
    log_info "Prerequisites check passed"
}

setup_directories() {
    log_info "Setting up directories..."
    
    # Create output directory
    mkdir -p "$OUTPUT_DIR"
    
    log_info "Directories created"
}

template_secrets() {
    log_info "Templating secrets..."
    
    # Process first secret (ibm-licensing-token) - copy to output directory
    cp "$INPUT_DIR/secret-ibm-licensing-token.yaml" "$OUTPUT_DIR/secret-ibm-licensing-token.yaml"
    
    # Step 1: Use yq to add placeholders
    $YQ -i '.metadata.namespace = "sed-me-namespace"' "$OUTPUT_DIR/secret-ibm-licensing-token.yaml"
    TOKEN_FIELD=$($YQ '.data | keys | .[0]' "$OUTPUT_DIR/secret-ibm-licensing-token.yaml")
    $YQ -i ".data.$TOKEN_FIELD = \"sed-me-token\"" "$OUTPUT_DIR/secret-ibm-licensing-token.yaml"
    
    # Step 2: Use sed to replace placeholders with Helm templates
    sed -i '' "s/namespace: sed-me-namespace/namespace: {{ .Values.ibmLicensing.namespace }}/g" "$OUTPUT_DIR/secret-ibm-licensing-token.yaml"
    sed -i '' "s/sed-me-token/{{ randAlphaNum 24 | b64enc }}/g" "$OUTPUT_DIR/secret-ibm-licensing-token.yaml"
    
    # Process second secret (ibm-licensing-upload-token) - copy to output directory
    cp "$INPUT_DIR/secret-ibm-licensing-upload-token.yaml" "$OUTPUT_DIR/secret-ibm-licensing-upload-token.yaml"
    
    # Step 1: Use yq to add placeholders
    $YQ -i '.metadata.namespace = "sed-me-namespace"' "$OUTPUT_DIR/secret-ibm-licensing-upload-token.yaml"
    UPLOAD_TOKEN_FIELD=$($YQ '.data | keys | .[0]' "$OUTPUT_DIR/secret-ibm-licensing-upload-token.yaml")
    $YQ -i ".data.$UPLOAD_TOKEN_FIELD = \"sed-me-upload-token\"" "$OUTPUT_DIR/secret-ibm-licensing-upload-token.yaml"
    
    # Step 2: Use sed to replace placeholders with Helm templates
    sed -i '' "s/namespace: sed-me-namespace/namespace: {{ .Values.ibmLicensing.namespace }}/g" "$OUTPUT_DIR/secret-ibm-licensing-upload-token.yaml"
    sed -i '' "s/sed-me-upload-token/{{ randAlphaNum 24 | b64enc }}/g" "$OUTPUT_DIR/secret-ibm-licensing-upload-token.yaml"
    
    # Combine both secrets into final output
    cat "$OUTPUT_DIR/secret-ibm-licensing-token.yaml" > "$OUTPUT_DIR/secrets.yaml"
    echo "---" >> "$OUTPUT_DIR/secrets.yaml"
    cat "$OUTPUT_DIR/secret-ibm-licensing-upload-token.yaml" >> "$OUTPUT_DIR/secrets.yaml"
    
    # Remove separate secret files
    rm "$OUTPUT_DIR/secret-ibm-licensing-token.yaml"
    rm "$OUTPUT_DIR/secret-ibm-licensing-upload-token.yaml"
    
    log_info "secrets.yaml created"
}

template_deployment() {
    log_info "Templating deployment..."
    
    # Copy to output directory and modify in place
    cp "$INPUT_DIR/deployment-ibm-licensing-service-instance.yaml" "$OUTPUT_DIR/deployment.yaml"
    
    # Step 1: Use yq to add placeholders
    # Replace namespace
    $YQ -i '.metadata.namespace = "sed-me-namespace"' "$OUTPUT_DIR/deployment.yaml"
    
    # Replace image
    $YQ -i '.spec.template.spec.containers[0].image = "sed-me-image"' "$OUTPUT_DIR/deployment.yaml"
    $YQ -i '.spec.template.spec.initContainers[0].image = "sed-me-image"' "$OUTPUT_DIR/deployment.yaml"
    
    # Remove imagePullSecrets (will be added conditionally later)
    $YQ -i 'del(.spec.template.spec.imagePullSecrets)' "$OUTPUT_DIR/deployment.yaml"
    
    # Replace environment variables in main container
    $YQ -i '.spec.template.spec.containers[0].env[0].value = "sed-me-namespace"' "$OUTPUT_DIR/deployment.yaml"
    $YQ -i '.spec.template.spec.containers[0].env[1].value = "sed-me-datasource"' "$OUTPUT_DIR/deployment.yaml"
    $YQ -i '.spec.template.spec.containers[0].env[2].value = "sed-me-httpsEnable"' "$OUTPUT_DIR/deployment.yaml"
    $YQ -i '.spec.template.spec.containers[0].env[3].value = "sed-me-enableInstanaMetricCollection"' "$OUTPUT_DIR/deployment.yaml"
    $YQ -i '.spec.template.spec.containers[0].env[4].value = "sed-me-httpsCertsSource"' "$OUTPUT_DIR/deployment.yaml"
    
    # Add new environment variables to main container
    $YQ -i '.spec.template.spec.containers[0].env += [{"name": "NAMESPACE_SCOPE_ENABLED", "value": "sed-me-nssEnabled"}]' "$OUTPUT_DIR/deployment.yaml"
    $YQ -i '.spec.template.spec.containers[0].env += [{"name": "WATCH_NAMESPACE", "value": "sed-me-watchNamespace"}]' "$OUTPUT_DIR/deployment.yaml"
    $YQ -i '.spec.template.spec.containers[0].env += [{"name": "NODE_CPU_CAPPING_ENABLED", "value": "sed-me-nodeCpuCappingEnabled"}]' "$OUTPUT_DIR/deployment.yaml"
    $YQ -i '.spec.template.spec.containers[0].env += [{"name": "KUBE_RBAC_AUTH_ENABLED", "value": "sed-me-kubeRBACAuthEnabled"}]' "$OUTPUT_DIR/deployment.yaml"
    $YQ -i '.spec.template.spec.containers[0].env += [{"name": "CUSTOM_RESOURCES_ENABLED", "value": "sed-me-customResourcesEnabled"}]' "$OUTPUT_DIR/deployment.yaml"
    
    # Replace environment variables in init container
    $YQ -i '.spec.template.spec.initContainers[0].env[0].value = "sed-me-namespace"' "$OUTPUT_DIR/deployment.yaml"
    $YQ -i '.spec.template.spec.initContainers[0].env[1].value = "sed-me-datasource"' "$OUTPUT_DIR/deployment.yaml"
    $YQ -i '.spec.template.spec.initContainers[0].env[2].value = "sed-me-httpsEnable"' "$OUTPUT_DIR/deployment.yaml"
    $YQ -i '.spec.template.spec.initContainers[0].env[3].value = "sed-me-enableInstanaMetricCollection"' "$OUTPUT_DIR/deployment.yaml"
    $YQ -i '.spec.template.spec.initContainers[0].env[4].value = "sed-me-httpsCertsSource"' "$OUTPUT_DIR/deployment.yaml"
    
    # Add new environment variables to init container
    $YQ -i '.spec.template.spec.initContainers[0].env += [{"name": "NAMESPACE_SCOPE_ENABLED", "value": "sed-me-nssEnabled"}]' "$OUTPUT_DIR/deployment.yaml"
    $YQ -i '.spec.template.spec.initContainers[0].env += [{"name": "WATCH_NAMESPACE", "value": "sed-me-watchNamespace"}]' "$OUTPUT_DIR/deployment.yaml"
    $YQ -i '.spec.template.spec.initContainers[0].env += [{"name": "NODE_CPU_CAPPING_ENABLED", "value": "sed-me-nodeCpuCappingEnabled"}]' "$OUTPUT_DIR/deployment.yaml"
    $YQ -i '.spec.template.spec.initContainers[0].env += [{"name": "KUBE_RBAC_AUTH_ENABLED", "value": "sed-me-kubeRBACAuthEnabled"}]' "$OUTPUT_DIR/deployment.yaml"
    $YQ -i '.spec.template.spec.initContainers[0].env += [{"name": "CUSTOM_RESOURCES_ENABLED", "value": "sed-me-customResourcesEnabled"}]' "$OUTPUT_DIR/deployment.yaml"
    
    # Replace resource limits and requests in main container
    $YQ -i '.spec.template.spec.containers[0].resources.limits.cpu = "sed-me-cpu-limit"' "$OUTPUT_DIR/deployment.yaml"
    $YQ -i '.spec.template.spec.containers[0].resources.limits.memory = "sed-me-memory-limit"' "$OUTPUT_DIR/deployment.yaml"
    $YQ -i '.spec.template.spec.containers[0].resources.requests.cpu = "sed-me-cpu-request"' "$OUTPUT_DIR/deployment.yaml"
    $YQ -i '.spec.template.spec.containers[0].resources.requests.memory = "sed-me-memory-request"' "$OUTPUT_DIR/deployment.yaml"
    $YQ -i '.spec.template.spec.containers[0].resources.requests."ephemeral-storage" = "sed-me-ephemeral-storage-request"' "$OUTPUT_DIR/deployment.yaml"
    
    # Replace resource limits and requests in init container
    $YQ -i '.spec.template.spec.initContainers[0].resources.limits.cpu = "sed-me-cpu-limit"' "$OUTPUT_DIR/deployment.yaml"
    $YQ -i '.spec.template.spec.initContainers[0].resources.limits.memory = "sed-me-memory-limit"' "$OUTPUT_DIR/deployment.yaml"
    $YQ -i '.spec.template.spec.initContainers[0].resources.requests.cpu = "sed-me-cpu-request"' "$OUTPUT_DIR/deployment.yaml"
    $YQ -i '.spec.template.spec.initContainers[0].resources.requests.memory = "sed-me-memory-request"' "$OUTPUT_DIR/deployment.yaml"
    $YQ -i '.spec.template.spec.initContainers[0].resources.requests."ephemeral-storage" = "sed-me-ephemeral-storage-request"' "$OUTPUT_DIR/deployment.yaml"
    
    # Update managed-by label from operator to Helm for pod template
    $YQ -i '.spec.template.metadata.labels."app.kubernetes.io/managed-by" = "Helm"' "$OUTPUT_DIR/deployment.yaml"
    
    # Step 2: Use sed to replace placeholders with Helm templates
    # Replace namespace
    sed -i '' "s/namespace: sed-me-namespace/namespace: {{ .Values.ibmLicensing.namespace }}/g" "$OUTPUT_DIR/deployment.yaml"
    
    # Replace image
    sed -i '' "s|image: sed-me-image|image: {{ .Values.global.imagePullPrefix }}/{{ .Values.ibmLicensing.imageRegistryNamespaceOperand }}/ibm-licensing:{{ .Values.ibmLicensing.ibmLicensingVersion }}|g" "$OUTPUT_DIR/deployment.yaml"
    
    # Replace environment variables
    sed -i '' 's/value: sed-me-namespace/value: {{ .Values.ibmLicensing.namespace | quote }}/g' "$OUTPUT_DIR/deployment.yaml"
    sed -i '' 's/value: sed-me-datasource/value: {{ .Values.ibmLicensing.spec.datasource | quote }}/g' "$OUTPUT_DIR/deployment.yaml"
    sed -i '' 's/value: "sed-me-httpsEnable"/value: {{ .Values.ibmLicensing.spec.httpsEnable | quote }}/g' "$OUTPUT_DIR/deployment.yaml"
    sed -i '' 's/value: "sed-me-enableInstanaMetricCollection"/value: {{ .Values.ibmLicensing.spec.enableInstanaMetricCollection | quote }}/g' "$OUTPUT_DIR/deployment.yaml"
    sed -i '' 's/value: sed-me-httpsCertsSource/value: {{ .Values.ibmLicensing.spec.httpsCertsSource | quote }}/g' "$OUTPUT_DIR/deployment.yaml"
    sed -i '' 's/value: sed-me-nssEnabled/value: {{ .Values.ibmLicensing.spec.features.nssEnabled | quote }}/g' "$OUTPUT_DIR/deployment.yaml"
    sed -i '' 's/value: sed-me-watchNamespace/value: {{ .Values.ibmLicensing.watchNamespace | quote }}/g' "$OUTPUT_DIR/deployment.yaml"
    sed -i '' 's/value: sed-me-nodeCpuCappingEnabled/value: {{ .Values.ibmLicensing.spec.features.nodeCpuCappingEnabled | quote }}/g' "$OUTPUT_DIR/deployment.yaml"
    sed -i '' 's/value: sed-me-kubeRBACAuthEnabled/value: {{ .Values.ibmLicensing.spec.features.kubeRBACAuthEnabled | quote }}/g' "$OUTPUT_DIR/deployment.yaml"
    sed -i '' 's/value: sed-me-customResourcesEnabled/value: {{ .Values.ibmLicensing.spec.features.customResourcesEnabled | quote }}/g' "$OUTPUT_DIR/deployment.yaml"
    
    # Replace resource limits and requests
    sed -i '' "s/sed-me-cpu-limit/{{ .Values.ibmLicensing.spec.resources.limits.cpu }}/g" "$OUTPUT_DIR/deployment.yaml"
    sed -i '' "s/sed-me-memory-limit/{{ .Values.ibmLicensing.spec.resources.limits.memory }}/g" "$OUTPUT_DIR/deployment.yaml"
    sed -i '' "s/sed-me-cpu-request/{{ .Values.ibmLicensing.spec.resources.requests.cpu }}/g" "$OUTPUT_DIR/deployment.yaml"
    sed -i '' "s/sed-me-memory-request/{{ .Values.ibmLicensing.spec.resources.requests.memory }}/g" "$OUTPUT_DIR/deployment.yaml"
    sed -i '' "s/sed-me-ephemeral-storage-request/{{ .Values.ibmLicensing.spec.resources.requests.ephemeralStorage }}/g" "$OUTPUT_DIR/deployment.yaml"
    
    # Replace serviceAccountName with helper template
    sed -i '' 's/serviceAccountName: ibm-license-service$/serviceAccountName: {{ include "ibm-licensing.operandServiceAccount" . }}/g' "$OUTPUT_DIR/deployment.yaml"
    
    # Append conditional imagePullSecrets section
    cat "${PROJECT_ROOT}/common/makefile-generate/yaml-deployment-pull-secrets-part" >> "$OUTPUT_DIR/deployment.yaml"
    
    log_info "deployment.yaml created"
}

template_service() {
    log_info "Templating service..."
    
    # Copy to output directory and modify in place
    cp "$INPUT_DIR/service-ibm-licensing-service-instance.yaml" "$OUTPUT_DIR/service.yaml"
    
    # Step 1: Use yq to add placeholders
    $YQ -i '.metadata.namespace = "sed-me-namespace"' "$OUTPUT_DIR/service.yaml"
    
    # Step 2: Use sed to replace placeholders with Helm templates
    sed -i '' "s/namespace: sed-me-namespace/namespace: {{ .Values.ibmLicensing.namespace }}/g" "$OUTPUT_DIR/service.yaml"
    
    log_info "service.yaml created"
}

template_crds() {
    log_info "Templating CRDs..."
    
    # CRDs are cluster-scoped resources and typically don't need templating
    # Just copy them directly to the output directory
    cp "$INPUT_DIR/crd.yaml" "$OUTPUT_DIR/crd.yaml"
    
    log_info "crd.yaml created"
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

    log_info ""
    log_info "Resource templating completed successfully!"
    log_info "Templated resources are available in: ${OUTPUT_DIR}/"
}

# Run main function
main "$@"
