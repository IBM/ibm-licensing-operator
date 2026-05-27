# Helm Chart Templating Plan for IBM Licensing Service

## Overview

This document outlines the plan for creating a **bash script** that will automate the templating of YAML files located in `./resources/final` to create a Helm chart. The script will transform static YAML files into Helm templates with configurable values, making these resources reusable across different environments.

**Goal**: Create a bash script (`template-helm-chart.sh`) that will:
1. Read the YAML files from `./resources/final`
2. Replace hardcoded values with Helm template expressions
3. Generate random tokens for secrets
4. Create a complete Helm chart structure with templates, values.yaml, and Chart.yaml
5. Output the templated Helm chart to a target directory

## Source Files Analysis

The following files in `./resources/final` need to be templated:

1. **deployment-ibm-licensing-service-instance.yaml** - Main deployment with containers and init containers
2. **secret-ibm-licensing-token.yaml** - API token secret
3. **secret-ibm-licensing-upload-token.yaml** - Upload token secret
4. **service-ibm-licensing-service-instance.yaml** - Service definition
5. **route-ibm-licensing-service-instance.yaml** - OpenShift route with TLS
6. **rbac.yaml** - Namespace-scoped Role and RoleBinding
7. **cluster-rbac.yaml** - ClusterRole and ClusterRoleBinding
8. **serviceaccounts.yaml** - ServiceAccount definition

## Key Templating Requirements

### 1. Deployment Environment Variables (High Priority)

The deployment contains multiple environment variables that need templating:

- `NAMESPACE` - Currently hardcoded to "ibm-licensing"
- `DATASOURCE` - Currently "datacollector"
- `HTTPS_ENABLE` - Currently "true"
- `ENABLE_INSTANA_METRIC_COLLECTION` - Currently "false"
- `HTTPS_CERTS_SOURCE` - Currently "external"
- `PROMETHEUS_QUERY_SOURCE_ENABLED` - Currently "true"

### 2. Secret Token Generation (High Priority)

Both secrets contain base64-encoded tokens that should be randomly generated:

- `ibm-licensing-token` - Contains `token` field with value: `eW9LSkZtTEd1Zlp1aDV3UWd6U2J2dTRj`
- `ibm-licensing-upload-token` - Contains `token-upload` field with value: `RkFVQmVCdnJGZ3FUemxEcTc0RTlDT2hR`

**Strategy**: Use Helm's `randAlphaNum` function to generate random tokens and base64 encode them.

### 3. Other Configurable Parameters

- **Namespace**: Currently hardcoded to "ibm-licensing" throughout all files
- **Image**: `docker-na-public.artifactory.swg-devops.com/hyc-cloud-private-integration-docker-local/ibmcom/ibm-licensing:4.2.23`
- **Image Pull Policy**: `IfNotPresent`
- **Image Pull Secrets**: `ibm-entitlement-key`
- **Replicas**: Currently 1
- **Resource Limits/Requests**: CPU and memory values
- **Service Account**: `ibm-license-service`
- **Labels**: Multiple app.kubernetes.io labels
- **Route TLS Configuration**: Certificate, key, and CA certificate (currently hardcoded)

## Implementation Plan

### Phase 1: Chart Structure Setup

- [ ] Create Helm chart directory structure:
  ```
  ibm-licensing-service/
  ├── Chart.yaml
  ├── values.yaml
  ├── templates/
  │   ├── _helpers.tpl
  │   ├── deployment.yaml
  │   ├── service.yaml
  │   ├── route.yaml
  │   ├── secrets.yaml
  │   ├── serviceaccount.yaml
  │   ├── rbac.yaml
  │   └── cluster-rbac.yaml
  └── README.md
  ```

- [ ] Create `Chart.yaml` with metadata:
  - Name: ibm-licensing-service
  - Version: 4.2.23 (matching current image version)
  - Description, maintainers, keywords
  - Reference existing chart at `deploy/argo-cd/components/license-service/helm-cluster-scoped/Chart.yaml`

### Phase 2: Values.yaml Design

- [ ] Design comprehensive `values.yaml` structure matching existing chart format:

```yaml
# Global settings (matches existing chart structure)
global:
  imagePullPrefix: docker-na-public.artifactory.swg-devops.com
  imagePullSecret: ibm-entitlement-key
  licenseAccept: true

# IBM Licensing configuration (matches existing chart structure)
ibmLicensing:
  # Namespace configuration
  namespace: ibm-licensing
  
  # Image registry configuration
  imageRegistryNamespaceOperator: hyc-cloud-private-integration-docker-local/ibmcom
  imageRegistryNamespaceOperand: hyc-cloud-private-integration-docker-local/ibmcom
  
  # RBAC configuration
  createRBAC: true
  
  # Version and image configuration
  version: 4.2.23
  
  # Environment variables
  datasource: datacollector
  httpsEnable: true
  enableInstanaMetricCollection: false
  httpsCertsSource: external
  prometheusQuerySourceEnabled: true
  
  
  # Resource limits
  resources:
    limits:
      cpu: 500m
      memory: 1Gi
    requests:
      cpu: 200m
      memory: 256Mi
      ephemeralStorage: 256Mi
  
  # Route configuration (OpenShift)
  route:
    enabled: true

  # Labels
  labels:
    app.kubernetes.io/component: ibm-licensing-service-svc
    app.kubernetes.io/instance: ibm-licensing-service
    app.kubernetes.io/managed-by: operator
    app.kubernetes.io/name: ibm-licensing-service-instance
    release: ibm-licensing-service
```

**Note**: This structure maintains compatibility with the existing chart by using the same top-level keys (`global` and `ibmLicensing`). All configuration is directly under `ibmLicensing` without a `spec` key since this chart deploys resources directly without a custom resource.

### Phase 3: Template Creation

#### 3.1 Helper Templates (_helpers.tpl)

- [ ] Create common label templates
- [ ] Create name generation templates
- [ ] Create image reference template
- [ ] Create selector labels template

#### 3.2 Deployment Template

- [ ] Template metadata (name, namespace, labels)
- [ ] Template spec.replicas from values
- [ ] Template container image from values
- [ ] **Template all environment variables** from values:
  ```yaml
  - name: NAMESPACE
    value: {{ .Values.namespace | quote }}
  - name: DATASOURCE
    value: {{ .Values.env.datasource | quote }}
  - name: HTTPS_ENABLE
    value: {{ .Values.env.httpsEnable | quote }}
  - name: ENABLE_INSTANA_METRIC_COLLECTION
    value: {{ .Values.env.enableInstanaMetricCollection | quote }}
  - name: HTTPS_CERTS_SOURCE
    value: {{ .Values.env.httpsCertsSource | quote }}
  - name: PROMETHEUS_QUERY_SOURCE_ENABLED
    value: {{ .Values.env.prometheusQuerySourceEnabled | quote }}
  ```
- [ ] Template resource limits/requests
- [ ] Template probes configuration
- [ ] Template volume mounts and volumes
- [ ] Template init container with same env var templating
- [ ] Template service account name
- [ ] Template image pull secrets
- [ ] Template affinity and tolerations

#### 3.3 Secret Templates

- [ ] Create secret template for API token with random generation:
  ```yaml
  apiVersion: v1
  kind: Secret
  metadata:
    name: ibm-licensing-token
    namespace: {{ .Values.namespace }}
    labels:
      {{- include "ibm-licensing.labels" . | nindent 4 }}
  type: Opaque
  data:
    token: {{ .Values.secrets.apiToken | default (randAlphaNum 32 | b64enc) | b64enc }}
  ```

- [ ] Create secret template for upload token with random generation:
  ```yaml
  apiVersion: v1
  kind: Secret
  metadata:
    name: ibm-licensing-upload-token
    namespace: {{ .Values.namespace }}
    labels:
      {{- include "ibm-licensing.labels" . | nindent 4 }}
  type: Opaque
  data:
    token-upload: {{ .Values.secrets.uploadToken | default (randAlphaNum 32 | b64enc) | b64enc }}
  ```

**Important**: Use `lookup` function to check if secrets already exist to avoid regenerating tokens on upgrades.

#### 3.4 Service Template

- [ ] Template metadata (name, namespace, labels, annotations)
- [ ] Template service type and ports
- [ ] Template selector labels

#### 3.5 Route Template (OpenShift)

- [ ] Add conditional rendering based on `route.enabled`
- [ ] Template metadata
- [ ] Template TLS configuration
- [ ] Handle certificate generation or use provided values
- [ ] Template target service reference

#### 3.6 RBAC Templates

- [ ] Template ServiceAccount with conditional creation
- [ ] Template Role with namespace-scoped permissions
- [ ] Template RoleBinding
- [ ] Template ClusterRole with cluster-wide permissions (conditional)
- [ ] Template ClusterRoleBinding (conditional)
- [ ] All RBAC resources should reference templated namespace and service account

### Phase 4: Documentation

- [ ] Create comprehensive README.md with:
  - Installation instructions
  - Configuration options documentation
  - Examples for different scenarios
  - Upgrade instructions
  - Token management guidance

- [ ] Add inline comments in templates explaining complex logic

- [ ] Document the random token generation approach and security considerations

### Phase 5: Testing & Validation

- [ ] Test chart installation with default values
- [ ] Test with custom values
- [ ] Verify environment variables are correctly templated
- [ ] Verify secrets are generated with random tokens
- [ ] Test upgrade scenario (ensure tokens persist)
- [ ] Validate RBAC permissions
- [ ] Test on different Kubernetes/OpenShift versions

## Key Considerations

### Security

1. **Token Generation**: Tokens should be randomly generated on first install but persist on upgrades
2. **Secret Management**: Consider using `lookup` function to check existing secrets
3. **RBAC**: Ensure minimal required permissions are granted

## Reference

Reference the existing chart at `deploy/argo-cd/components/license-service/helm-cluster-scoped/` for:
- Chart metadata structure
- Values organization patterns
- Label conventions
- Documentation style

## Success Criteria

- [ ] All hardcoded values are templated
- [ ] Environment variables are configurable via values.yaml
- [ ] Secrets contain randomly generated tokens
- [ ] Chart can be installed with `helm install`
- [ ] Chart can be upgraded without breaking changes
- [ ] Documentation is complete and clear
- [ ] All resources are properly labeled and annotated

**Note**: This plan focuses on creating a production-ready Helm chart with proper templating, security considerations, and maintainability. The implementation should follow Helm best practices and be compatible with both OpenShift and standard Kubernetes environments where applicable.