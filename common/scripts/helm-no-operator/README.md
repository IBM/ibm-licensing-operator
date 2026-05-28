# Helm Chart Generation Scripts (No Operator)

This directory contains scripts for generating a standalone Helm chart for IBM Licensing Service that can be deployed without the IBM Licensing Operator.

## Overview

These scripts automate the process of:
1. Extracting resources from a running IBM Licensing Service instance
2. Generating RBAC and CRD resources from the operator's manifests
3. Converting extracted resources into Helm templates
4. Building a complete Helm chart ready for deployment

## Scripts

### 1. `build-helm-chart.sh`

**Main orchestration script** that runs the entire Helm chart generation process.

**Usage:**
```bash
./common/scripts/helm-no-operator/build-helm-chart.sh
```

**What it does:**
- Executes all steps in sequence
- Provides colored output for easy tracking
- Cleans up temporary resources after completion
- Displays next steps after successful completion

**Output:**
- Generated Helm templates in `helm-no-operator/templates/`

---

### 2. `extract-cluster-resources.sh`

**Extracts resources from a running Kubernetes cluster** where IBM Licensing Service has been deployed via the operator.

---

### 3. `generate-resources.sh`

**Generates RBAC and CRD resources** from the operator's Kustomize manifests.

---

### 4. `template-resources.sh`

**Converts extracted YAML resources into Helm templates** with parameterized values.

---

## Complete Workflow

To generate a complete Helm chart from scratch:

```bash
# Run the main build script (recommended)
./common/scripts/helm-no-operator/build-helm-chart.sh
```

After running the build process, you need to check what was generated in helm-no-operator folder. If you want to adjust generated helm chart, you will need to adjust the generation scripts and re-run the build process.

## Testing the Generated Chart

```bash
# Validate the chart
helm lint ./helm-no-operator

# Install on cluster
helm install ibm-licensing ./helm-no-operator
```

## Requirements

### Tools
- `kubectl` - Kubernetes CLI
- `helm` - Helm package manager
- `sed` - Stream editor (macOS or Linux version)

### Cluster Access
- Active Kubernetes cluster connection
- Sufficient permissions to create namespaces and deploy resources

## Route management
If you are using OpenShift and want to use route to access Licensing Service endpoints, you will need to create the route manually after the helm installation. You can use ibm-license-service-cert-internal Secret to get the TLS certificate, but keep in mind that the service CA certificate, which issues the service certificates, is valid for 26 months and is automatically rotated when there is less than 13 months validity left. After rotation, the previous service CA configuration is still trusted until its expiration. This allows a grace period for all affected services to refresh their key material before the expiration.

Example route (make sure to adjust namespace and TLS configuration):
```yaml
apiVersion: route.openshift.io/v1
kind: Route
metadata:
  annotations: {}
  name: ibm-licensing-service-instance
  namespace: ibm-licensing
spec:
  port:
    targetPort: api-port
  tls:
    certificate: |-
      -----BEGIN CERTIFICATE-----
      <your-certificate>
      -----END CERTIFICATE-----
    destinationCACertificate: |-
      -----BEGIN CERTIFICATE-----
      <your-ca-certificate>
      -----END CERTIFICATE-----
    insecureEdgeTerminationPolicy: None
    key: |-
      -----BEGIN PRIVATE KEY-----
      <your-private-key>
      -----END PRIVATE KEY-----
    termination: reencrypt
  to:
    kind: Service
    name: ibm-licensing-service-instance
    weight: 100
```