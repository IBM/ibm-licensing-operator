# Plan: Generate Helm Chart from Deployed License Service Operator

## Overview
Create a script that installs IBM License Service on a Kubernetes cluster, extracts all created resources, and generates a new Helm chart. This allows us to create a standalone Helm chart that doesn't require the operator.

## Implementation Plan

### Phase 1: Create Resource Extraction Script

#### Script: `scripts/extract-ls-resources.sh`

**Purpose**: Install IBM Licensing Service using Helm charts, extract all created resources, and organize them for generating a standalone Helm chart.

**Key Steps**:
1. Install IBM Licensing Service using existing Helm charts (`deploy/argo-cd/components/license-service/helm-cluster-scoped/`)
2. Wait for all resources to be ready
3. Extract all created resources using label selector: `app.kubernetes.io/managed-by=ibm-licensing-operator`
4. Clean up extracted resources (remove runtime fields like uid, resourceVersion, status etc.)
5. Organize resources into directory structure

**Script will be implemented in**: `scripts/extract-operator-resources.sh`

**Key Functions**:
- `install_licensing_helm()`: Install LS using Helm charts from the repository
- `wait_for_resources()`: Wait for all resources to be ready
- `extract_resources()`: Extract all namespace-scoped and cluster-scoped resources using label selector
- `cleanup_resource()`: Remove runtime fields from extracted YAML files

**Resource Identification**:
All resources created by the IBM Licensing Operator are labeled with:
- `app.kubernetes.io/managed-by=ibm-licensing-operator`

This label will be used to identify and extract all operator-managed resources.

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

All these resources will be labeled with `app.kubernetes.io/managed-by=ibm-licensing-operator`

### Phase 3: Resource Extraction Details

#### 3.1 Namespace-Scoped Resources to Extract

All namespace-scoped resources will be extracted using the label selector:
`app.kubernetes.io/managed-by=ibm-licensing-operator`

**Resources to extract**:
- **Deployment**: License Service deployment
- **Service**: Service exposing the License Service
- **ServiceAccount**: Service account for the operand
- **Role**: Namespace-scoped permissions
- **RoleBinding**: Binding the role to the service account
- **ConfigMaps**: All configuration maps created by the operator
- **Secrets**: Any secrets created by the operator
- **Routes** (OpenShift only): Routes for external access
- **Any other resources**: Any other resources created by the operator

**Extraction process**:
1. Use kubectl with label selector to get all resources of each type
2. Remove runtime fields (uid, resourceVersion, status, managedFields, etc.)
3. Save each resource to a separate YAML file in `resources/` directory

#### 3.2 Cluster-Scoped Resources to Extract

Cluster-scoped resources will also be identified using the label selector where applicable:
`app.kubernetes.io/managed-by=ibm-licensing-operator`

**Resources to extract**:
- **ClusterRole**: Cluster-wide permissions for the operand
- **ClusterRoleBinding**: Binding the cluster role to the service account

**Note**: If cluster-scoped resources don't have the label, they will be identified by name pattern (e.g., resources containing `ibm-licensing`).

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
│   ├── clusterrole.yaml
│   ├── clusterrolebinding.yaml
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
  
  # Image configuration
  image:
    repository: icr.io/cpopen/cpfs/ibm-licensing
    tag: 4.2.21
    pullPolicy: IfNotPresent
  
  # Environment variables
  env:
    httpsEnable: "true"
    datasource: "datacollector"
```

### Phase 5: Analyze and Handle Secrets

#### 5.1 Identify Secret Types

**Purpose**: Analyze extracted secrets to determine which ones require special handling in the Helm chart.

**Secret types to handle**:

1. **TLS Secrets**:
   - Certificate and key pairs for HTTPS/TLS connections
   - Used for secure communication with License Service
   - Must be created before Helm installation or generated during installation

2. **Token Secrets**:
   - Random authentication tokens
   - API access tokens
   - Must be generated securely and persist across upgrades

#### 5.2 Helm Chart Strategies for Secrets

**Strategy 1: TLS Secret Handling**

Option A - Use existing TLS secret:
```yaml
# values.yaml
tls:
  enabled: true
  secretName: "ibm-licensing-tls"  # Reference to pre-existing secret
```

Option B - Generate TLS certificate during installation:
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

**Strategy 3: External Secret References**

For production environments, allow referencing externally managed secrets:
```yaml
# values.yaml
secrets:
  tls:
    create: false  # Don't create, use existing
    secretName: "my-existing-tls-secret"
  token:
    create: false  # Don't create, use existing
    secretName: "my-existing-token-secret"
```

### Phase 6: Templatize Extracted Resources

#### 6.1 Script: `scripts/templatize-resources.sh`

**Purpose**: Convert extracted static YAML files into Helm templates with proper templating.

**Key Transformations**:

1. **Replace hardcoded namespaces**: Convert to Helm template variables
2. **Replace image references**: Use values from values.yaml
3. **Add conditional blocks**: For platform-specific resources (e.g., OpenShift Routes)
4. **Template environment variables**: Use values from values.yaml

#### 6.2 Templatization Process

The script will perform these transformations on extracted resources:
- Replace hardcoded namespaces with Helm template variables
- Replace image references with values from values.yaml
- Add conditional blocks for platform-specific resources (e.g., OpenShift Routes)
- Template resource limits and requests
- Template environment variables

### Phase 7: Complete Automation Script

#### Script: `scripts/generate-helm-no-operator.sh`

**Complete workflow**:

1. Install IBM Licensing Service using Helm charts from `deploy/argo-cd/components/license-service/helm-cluster-scoped/`
2. Wait for the operator to auto-create the IBMLicensing instance
3. Wait for all resources to be ready
4. Extract all resources from the cluster using label selector
5. Analyze secrets and implement proper handling strategies
6. Templatize resources for the new Helm chart
7. Cleanup (optional)

**Output**:
- `helm-no-operator/` - Complete Helm chart ready to use (includes both namespace and cluster-scoped resources)

**Temporary files** (can be deleted after generation):
- `resources/` - Raw extracted resources used during chart generation

## Success Criteria

1. ✅ Script successfully installs LS operator and creates instance
2. ✅ All resources are correctly extracted using label selector
3. ✅ Secrets are properly handled (TLS and token secrets)
4. ✅ TLS secrets can be auto-generated or referenced externally
5. ✅ Token secrets are generated once and preserved on upgrades
6. ✅ Generated Helm chart deploys successfully
7. ✅ Deployed LS functions identically to operator-managed version