# Helm Chart Generation Scripts (No Operator)

This directory contains scripts for generating a standalone Helm chart for IBM Licensing Service deployed without the IBM Licensing Operator.

## Overview

These scripts automate the process of:
1. Extracting resources created by IBM Licensing Service operator into yaml files
2. Generating CRD and instance RBAC resources from the operator's manifests
3. Converting extracted resources into Helm templates

## Scripts

### 1. `build-helm-chart.sh`

**Main orchestration script** that runs the entire Helm chart generation process.

**Usage:**
```bash
./common/scripts/helm-no-operator/build-helm-chart.sh
```

**What it does:**
- Executes all steps in sequence
- Cleans up temporary resources after completion

**Output:**
- Generated Helm templates in `helm-no-operator/templates/`

---

### 2. `extract-cluster-resources.sh`

**Extracts resources created by the operator from a running Kubernetes cluster** where IBM Licensing Service has been deployed.

---

### 3. `generate-resources.sh`

**Generates CRD and instance RBAC resources** from the operator's Kustomize manifests.

---

### 4. `template-resources.sh`

**Converts extracted YAML resources into Helm templates** with parameterized values.

---

## Complete Workflow

To generate a complete Helm chart from scratch:

```bash
# Run the main build script
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
- `sed` - Stream editor

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
```