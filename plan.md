# Plan: Generate Helm Chart from Deployed License Service Operator

## Overview
Create a script that installs IBM License Service on a Kubernetes cluster, extracts all created resources, and generates a new Helm chart. This allows us to create a standalone Helm chart that doesn't require the operator.

## Implementation Plan

### Phase 1: Create Resource Extraction Script

**Purpose**: Install IBM Licensing Service using Helm charts, extract all created resources, and organize them for generating a standalone Helm chart (save them to yaml files).

**Key Steps**:
1. Install IBM Licensing Service using existing Helm charts (`deploy/argo-cd/components/license-service/helm-cluster-scoped/`)
2. Wait for all resources to be ready
3. Extract all created resources using label selector: `app.kubernetes.io/managed-by=ibm-licensing-operator`
4. Clean up extracted resources (remove runtime fields like uid, resourceVersion, status etc.)
5. Organize resources into separate yaml files in `resources/` directory

**Script implemented in**: `scripts/extract-operator-resources.sh`

**Next Step**: Run the extraction script to generate actual resource YAML files, then analyze them before creating the templatization script.

**Key Functions**:
- `install_licensing_helm()`: Install LS using Helm charts from the repository
- `wait_for_resources()`: Wait for all resources to be ready
- `extract_resources()`: Extract all namespace-scoped and cluster-scoped resources using label selector
- `cleanup_resource()`: Remove runtime fields from extracted YAML files

**Resource Identification**:
Resources created by the IBM Licensing Operator have label `app.kubernetes.io/instance=ibm-licensing-service`

### Phase 2: Installation Method

#### 2.1 Install License Service Using Helm

The script will use the existing Helm charts from `deploy/argo-cd/components/license-service/helm-cluster-scoped/`

**Installation**:
The installation requires running the same helm template command twice to ensure all resources are properly created:

```bash
helm template ibm-licensing-cluster-scoped deploy/argo-cd/components/license-service/helm-cluster-scoped/ | kubectl apply -f -
helm template ibm-licensing-cluster-scoped deploy/argo-cd/components/license-service/helm-cluster-scoped/ | kubectl apply -f -
```

**Why run twice?**
- First run: Creates CRDs and initial resources
- Second run: Ensures all dependent resources are properly created after CRDs are established
- Wait for deployment to be ready after both runs

**What happens**:
1. Helm installs the IBM Licensing Operator
2. The operator automatically creates a default IBMLicensing CR instance
3. The operator reconciles and creates all necessary resources:
   - Deployment
   - Service
   - ServiceAccount
   - Role and RoleBinding
   - ClusterRole and ClusterRoleBinding
   - ConfigMaps
   - Secrets
   - Routes (on OpenShift)
   - etc.

### Phase 3: Templatize Extracted Resources and Handle Secrets

#### 3.1 Script: `scripts/templatize-resources.sh`

**Purpose**: Convert extracted static YAML files into Helm templates with proper templating and secret handling.

**Note**: This script will be created after running the extraction script and analyzing the actual resources that are created by the operator.

**Key Transformations**:

1. **Replace hardcoded namespaces**: Convert to Helm template variables
2. **Add conditional blocks**: For platform-specific resources (e.g., OpenShift Routes)
3. **Template environment variables**: Use values from values.yaml
4. **Implement secret handling strategies**: For TLS and token secrets

#### 3.2 Templatization Process

The script will perform these transformations on extracted resources:
- Replace hardcoded namespaces with Helm template variables
- Add conditional blocks for platform-specific resources (e.g., OpenShift Routes)
- Template resource limits and requests
- Template environment variables

#### 3.3 Identify Secret Types

**Purpose**: Analyze extracted secrets to determine which ones require special handling in the Helm chart.

**Secret types to handle**:

1. **TLS Secrets**:
   - Certificate and key pairs for HTTPS/TLS connections
   - Used for secure communication with License Service
   - Must be generated during installation and persist across upgrades

2. **Token Secrets**:
   - Random authentication tokens
   - API access tokens
   - Must be generated securely and persist across upgrades

#### 3.4 Helm Chart Strategies for Secrets

**Strategy 1: TLS Secret Handling**

Generate TLS certificate during installation:
```yaml
# values.yaml
tls:
  enabled: true
  generate: true  # Auto-generate self-signed certificate
  secretName: "ibm-licensing-tls"
```

Implementation in template:
```yaml
{{- if and .Values.tls.enabled .Values.tls.generate }}
{{- if not (lookup "v1" "Secret" .Release.Namespace .Values.tls.secretName) }}
# Generate self-signed certificate using Helm's genSelfSignedCert function
{{- $cert := genSelfSignedCert .Values.tls.commonName nil nil 365 }}
apiVersion: v1
kind: Secret
type: kubernetes.io/tls
metadata:
  name: {{ .Values.tls.secretName }}
data:
  tls.crt: {{ $cert.Cert | b64enc }}
  tls.key: {{ $cert.Key | b64enc }}
{{- end }}
{{- end }}
```

**Strategy 2: Token Secret Handling**

Use Helm's `lookup` function to preserve existing tokens on upgrades:
```yaml
{{- $secret := lookup "v1" "Secret" .Release.Namespace "ibm-licensing-token" }}
{{- $token := "" }}
{{- if $secret }}
  {{- $token = index $secret.data "token" | b64dec }}
{{- else }}
  {{- $token = randAlphaNum 32 }}
{{- end }}
apiVersion: v1
kind: Secret
metadata:
  name: ibm-licensing-token
type: Opaque
data:
  token: {{ $token | b64enc }}
```

This approach:
- Generates a random token on first installation
- Preserves the existing token on upgrades
- Ensures token consistency across Helm operations

### Phase 4: Create Helm Chart Structure

#### 4.1 Directory Structure
```
helm-no-operator/
├── Chart.yaml
├── values.yaml
├── templates/
│   ├── deployment.yaml
│   ├── service.yaml
│   ├── serviceaccount.yaml
│   ├── role.yaml
│   ├── rolebinding.yaml
│   ├── configmaps.yaml
│   ├── secrets.yaml
│   └── routes.yaml
└── README.md
```

**Note**: During development, a temporary `resources/` directory will be used to store raw extracted YAMLs before templatization. This directory is not part of the final Helm chart and can be deleted after the chart is generated.

#### 4.2 Chart.yaml
```yaml
apiVersion: v2
name: ibm-licensing-no-operator
description: A Helm chart for IBM Licensing Service installation (without operator)
type: application
version: 4.2.21
appVersion: "4.2.21"
```

#### 4.3 values.yaml

The values.yaml will be based on the structure from `deploy/argo-cd/components/license-service/helm-cluster-scoped/values.yaml` but simplified for standalone deployment without the operator.

Example structure:
```yaml
---
global:
  licenseAccept: true
  imagePullPrefix: icr.io
  imagePullSecret: ibm-entitlement-key
  instanceNamespace: ""

ibmLicensing:
  imageRegistryNamespace: cpopen/cpfs
  enableRoutes: true
  
  # Environment variables for the operand
  env:
    httpsEnable: "true"
    datasource: "datacollector"
```

### Phase 5: Complete Automation Script

#### Script: `scripts/generate-helm-no-operator.sh`

**Complete workflow**:

1. Install IBM Licensing Service using Helm charts from `deploy/argo-cd/components/license-service/helm-cluster-scoped/`
2. Wait for all resources to be ready
3. Extract all resources from the cluster using label selector
4. Templatize resources and implement secret handling strategies
5. Generate complete Helm chart structure
6. Cleanup (optional)

**Output**:
- `helm-no-operator/` - Complete Helm chart ready to use (includes both namespace and cluster-scoped resources)

**Temporary files** (can be deleted after generation):
- `resources/` - Raw extracted resources used during chart generation

## Success Criteria

1. ✅ Script successfully installs LS operator
2. ✅ All resources are correctly extracted using label selector
3. ✅ Secrets are properly handled (TLS and token secrets)
4. ✅ TLS secrets are generated once and preserved on upgrades
5. ✅ Token secrets are generated once and preserved on upgrades
6. ✅ Generated Helm chart deploys successfully
7. ✅ Deployed LS functions identically to operator-managed version