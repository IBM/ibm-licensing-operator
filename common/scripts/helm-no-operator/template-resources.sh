#!/bin/bash
# template-resources.sh
# Purpose: Template YAML files (secrets, deployment) from resources/final to Helm templates using yq and sed
# Usage: ./scripts/template-resources.sh

set -e

INPUT_DIR="./resources/final"
OUTPUT_DIR="./helm-no-operator/templates"
TEMP_DIR="./temp"
YQ="./bin/yq"

# Check if yq is available
if [ ! -x "$YQ" ]; then
    echo "[ERROR] yq not found at $YQ. Please run 'make install-yq' first."
    exit 1
fi

echo "[INFO] Templating secrets..."

# Create temp directory
mkdir -p "$TEMP_DIR"

# Process first secret (ibm-licensing-token)
cp "$INPUT_DIR/secret-ibm-licensing-token.yaml" "$TEMP_DIR/secret-ibm-licensing-token.yaml"

# Step 1: Use yq to add placeholder for token (yq works on valid YAML)
TOKEN_FIELD=$($YQ '.data | keys | .[0]' "$TEMP_DIR/secret-ibm-licensing-token.yaml")
$YQ -i ".data.$TOKEN_FIELD = \"sed-me-token\"" "$TEMP_DIR/secret-ibm-licensing-token.yaml"

# Step 2: Use sed to replace placeholders with Helm templates
sed -i '' "s/namespace: ibm-licensing/namespace: {{ .Values.ibmLicensing.namespace }}/g" "$TEMP_DIR/secret-ibm-licensing-token.yaml"
sed -i '' "s/sed-me-token/{{ randAlphaNum 32 | b64enc }}/g" "$TEMP_DIR/secret-ibm-licensing-token.yaml"

# Process second secret (ibm-licensing-upload-token)
cp "$INPUT_DIR/secret-ibm-licensing-upload-token.yaml" "$TEMP_DIR/secret-ibm-licensing-upload-token.yaml"

# Step 1: Use yq to add placeholder for token (yq works on valid YAML)
UPLOAD_TOKEN_FIELD=$($YQ '.data | keys | .[0]' "$TEMP_DIR/secret-ibm-licensing-upload-token.yaml")
$YQ -i ".data.$UPLOAD_TOKEN_FIELD = \"sed-me-upload-token\"" "$TEMP_DIR/secret-ibm-licensing-upload-token.yaml"

# Step 2: Use sed to replace placeholders with Helm templates
sed -i '' "s/namespace: ibm-licensing/namespace: {{ .Values.ibmLicensing.namespace }}/g" "$TEMP_DIR/secret-ibm-licensing-upload-token.yaml"
sed -i '' "s/sed-me-upload-token/{{ randAlphaNum 32 | b64enc }}/g" "$TEMP_DIR/secret-ibm-licensing-upload-token.yaml"

# Combine both secrets into final output
cat "$TEMP_DIR/secret-ibm-licensing-token.yaml" > "$OUTPUT_DIR/secrets.yaml"
echo "---" >> "$OUTPUT_DIR/secrets.yaml"
cat "$TEMP_DIR/secret-ibm-licensing-upload-token.yaml" >> "$OUTPUT_DIR/secrets.yaml"

echo "[INFO] ✓ secrets.yaml created"

# Process deployment
echo "[INFO] Templating deployment..."

cp "$INPUT_DIR/deployment-ibm-licensing-service-instance.yaml" "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"

# Step 1: Use yq to add placeholders for values that need templating
# Replace namespace
$YQ -i '.metadata.namespace = "sed-me-namespace"' "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"

# Replace image
$YQ -i '.spec.template.spec.containers[0].image = "sed-me-image"' "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
$YQ -i '.spec.template.spec.initContainers[0].image = "sed-me-image"' "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"

# Replace imagePullSecrets
$YQ -i '.spec.template.spec.imagePullSecrets[0].name = "sed-me-imagePullSecret"' "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"

# Replace environment variables in main container
$YQ -i '.spec.template.spec.containers[0].env[0].value = "sed-me-namespace"' "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
$YQ -i '.spec.template.spec.containers[0].env[1].value = "sed-me-datasource"' "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
$YQ -i '.spec.template.spec.containers[0].env[2].value = "sed-me-httpsEnable"' "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
$YQ -i '.spec.template.spec.containers[0].env[3].value = "sed-me-enableInstanaMetricCollection"' "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
$YQ -i '.spec.template.spec.containers[0].env[4].value = "sed-me-httpsCertsSource"' "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
$YQ -i '.spec.template.spec.containers[0].env[5].value = "sed-me-prometheusQuerySourceEnabled"' "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"

# Replace environment variables in init container
$YQ -i '.spec.template.spec.initContainers[0].env[0].value = "sed-me-namespace"' "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
$YQ -i '.spec.template.spec.initContainers[0].env[1].value = "sed-me-datasource"' "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
$YQ -i '.spec.template.spec.initContainers[0].env[2].value = "sed-me-httpsEnable"' "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
$YQ -i '.spec.template.spec.initContainers[0].env[3].value = "sed-me-enableInstanaMetricCollection"' "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
$YQ -i '.spec.template.spec.initContainers[0].env[4].value = "sed-me-httpsCertsSource"' "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
$YQ -i '.spec.template.spec.initContainers[0].env[5].value = "sed-me-prometheusQuerySourceEnabled"' "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"

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

# Replace imagePullSecrets BEFORE image (to avoid conflicts)
sed -i '' "s/name: sed-me-imagePullSecret/name: {{ .Values.global.imagePullSecret }}/g" "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"

# Replace image
sed -i '' "s|image: sed-me-image|image: {{ .Values.global.imagePullPrefix }}/{{ .Values.ibmLicensing.imageRegistryNamespaceOperand }}/ibm-licensing:4.2.23|g" "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"

# Replace environment variables
sed -i '' 's/value: sed-me-namespace/value: {{ .Values.ibmLicensing.namespace | quote }}/g' "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
sed -i '' 's/value: sed-me-datasource/value: {{ .Values.ibmLicensing.datasource | quote }}/g' "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
sed -i '' 's/value: "sed-me-httpsEnable"/value: {{ .Values.ibmLicensing.httpsEnable | quote }}/g' "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
sed -i '' 's/value: "sed-me-enableInstanaMetricCollection"/value: {{ .Values.ibmLicensing.enableInstanaMetricCollection | quote }}/g' "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
sed -i '' 's/value: sed-me-httpsCertsSource/value: {{ .Values.ibmLicensing.httpsCertsSource | quote }}/g' "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
sed -i '' 's/value: "sed-me-prometheusQuerySourceEnabled"/value: {{ .Values.ibmLicensing.prometheusQuerySourceEnabled | quote }}/g' "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"

# Replace resource limits and requests
sed -i '' "s/sed-me-cpu-limit/{{ .Values.ibmLicensing.resources.limits.cpu }}/g" "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
sed -i '' "s/sed-me-memory-limit/{{ .Values.ibmLicensing.resources.limits.memory }}/g" "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
sed -i '' "s/sed-me-cpu-request/{{ .Values.ibmLicensing.resources.requests.cpu }}/g" "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
sed -i '' "s/sed-me-memory-request/{{ .Values.ibmLicensing.resources.requests.memory }}/g" "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"
sed -i '' "s/sed-me-ephemeral-storage-request/{{ .Values.ibmLicensing.resources.requests.ephemeralStorage }}/g" "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml"

# Copy to output
cp "$TEMP_DIR/deployment-ibm-licensing-service-instance.yaml" "$OUTPUT_DIR/deployment.yaml"

echo "[INFO] ✓ deployment.yaml created"

# Process service
echo "[INFO] Templating service..."

cp "$INPUT_DIR/service-ibm-licensing-service-instance.yaml" "$TEMP_DIR/service-ibm-licensing-service-instance.yaml"

# Step 1: Use yq to add placeholder for namespace
$YQ -i '.metadata.namespace = "sed-me-namespace"' "$TEMP_DIR/service-ibm-licensing-service-instance.yaml"

# Step 2: Use sed to replace placeholder with Helm template
sed -i '' "s/namespace: sed-me-namespace/namespace: {{ .Values.ibmLicensing.namespace }}/g" "$TEMP_DIR/service-ibm-licensing-service-instance.yaml"

# Copy to output
cp "$TEMP_DIR/service-ibm-licensing-service-instance.yaml" "$OUTPUT_DIR/service.yaml"

echo "[INFO] ✓ service.yaml created"

# Process route
echo "[INFO] Templating route..."

cp "$INPUT_DIR/route-ibm-licensing-service-instance.yaml" "$TEMP_DIR/route-ibm-licensing-service-instance.yaml"

# Step 1: Use yq to add placeholder for namespace
$YQ -i '.metadata.namespace = "sed-me-namespace"' "$TEMP_DIR/route-ibm-licensing-service-instance.yaml"

# Step 2: Remove hardcoded TLS certificates
$YQ -i 'del(.spec.tls.certificate)' "$TEMP_DIR/route-ibm-licensing-service-instance.yaml"
$YQ -i 'del(.spec.tls.key)' "$TEMP_DIR/route-ibm-licensing-service-instance.yaml"
$YQ -i 'del(.spec.tls.destinationCACertificate)' "$TEMP_DIR/route-ibm-licensing-service-instance.yaml"

# Step 3: Use sed to replace placeholder with Helm template
sed -i '' "s/namespace: sed-me-namespace/namespace: {{ .Values.ibmLicensing.namespace }}/g" "$TEMP_DIR/route-ibm-licensing-service-instance.yaml"

# Copy to output
cp "$TEMP_DIR/route-ibm-licensing-service-instance.yaml" "$OUTPUT_DIR/route.yaml"

echo "[INFO] ✓ route.yaml created"

# Process CRDs
echo "[INFO] Templating CRDs..."

# CRDs are cluster-scoped resources and typically don't need templating
# Just copy them directly to the output directory
cp "$INPUT_DIR/crds.yaml" "$OUTPUT_DIR/crds.yaml"

echo "[INFO] ✓ crds.yaml created"

# Process ServiceAccount
echo "[INFO] Templating serviceaccount..."

cp "$INPUT_DIR/serviceaccounts.yaml" "$TEMP_DIR/serviceaccounts.yaml"

# Step 1: Use yq to add placeholder for namespace
$YQ -i '.metadata.namespace = "sed-me-namespace"' "$TEMP_DIR/serviceaccounts.yaml"

# Step 2: Use sed to replace placeholder with Helm template
sed -i '' "s/namespace: sed-me-namespace/namespace: {{ .Values.ibmLicensing.namespace }}/g" "$TEMP_DIR/serviceaccounts.yaml"

# Copy to output
cp "$TEMP_DIR/serviceaccounts.yaml" "$OUTPUT_DIR/serviceaccount.yaml"

echo "[INFO] ✓ serviceaccount.yaml created"

# Process RBAC (Role and RoleBinding)
echo "[INFO] Templating rbac..."

cp "$INPUT_DIR/rbac.yaml" "$TEMP_DIR/rbac.yaml"

# Step 1: Use yq to add placeholders for namespace in both Role and RoleBinding
$YQ -i '(select(.kind == "Role") | .metadata.namespace) = "sed-me-namespace"' "$TEMP_DIR/rbac.yaml"
$YQ -i '(select(.kind == "RoleBinding") | .metadata.namespace) = "sed-me-namespace"' "$TEMP_DIR/rbac.yaml"
$YQ -i '(select(.kind == "RoleBinding") | .subjects[0].namespace) = "sed-me-namespace"' "$TEMP_DIR/rbac.yaml"

# Step 2: Use sed to replace placeholders with Helm templates
sed -i '' "s/namespace: sed-me-namespace/namespace: {{ .Values.ibmLicensing.namespace }}/g" "$TEMP_DIR/rbac.yaml"

# Copy to output
cp "$TEMP_DIR/rbac.yaml" "$OUTPUT_DIR/rbac.yaml"

echo "[INFO] ✓ rbac.yaml created"

# Process Cluster RBAC (ClusterRole and ClusterRoleBinding)
echo "[INFO] Templating cluster-rbac..."

cp "$INPUT_DIR/cluster-rbac.yaml" "$TEMP_DIR/cluster-rbac.yaml"

# Step 1: Use yq to add placeholders for namespace in ClusterRoleBinding
# Note: ClusterRole doesn't have namespace, but ClusterRoleBinding metadata and subjects do
$YQ -i '(select(.kind == "ClusterRoleBinding") | .metadata.namespace) = "sed-me-namespace"' "$TEMP_DIR/cluster-rbac.yaml"
$YQ -i '(select(.kind == "ClusterRoleBinding") | .subjects[0].namespace) = "sed-me-namespace"' "$TEMP_DIR/cluster-rbac.yaml"

# Step 2: Use sed to replace placeholders with Helm templates
sed -i '' "s/namespace: sed-me-namespace/namespace: {{ .Values.ibmLicensing.namespace }}/g" "$TEMP_DIR/cluster-rbac.yaml"

# Copy to output
cp "$TEMP_DIR/cluster-rbac.yaml" "$OUTPUT_DIR/cluster-rbac.yaml"

echo "[INFO] ✓ cluster-rbac.yaml created"

# Clean up temp files
rm -rf "$TEMP_DIR"
