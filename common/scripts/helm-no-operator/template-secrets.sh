#!/bin/bash
# template-secrets.sh
# Purpose: Template secret YAML files from resources/final to Helm templates using yq and sed
# Usage: ./scripts/template-secrets.sh

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

# Clean up temp files
rm -rf "$TEMP_DIR"

echo "[INFO] ✓ secrets.yaml created"
