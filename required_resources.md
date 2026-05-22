# Required Resources for Helm Chart (Without Operator)

## Overview

This document outlines the resources needed for deploying the IBM Licensing Service using Helm charts without the operator.

## Required Resources

### Core Components

- **Operand Deployment**: `ibm-licensing-service-instance`
- **RBAC**: Role-based access control (sourced from Kustomize, operand RBAC only)
- **ConfigMaps**: Currently not required, as they are not mounted to the deployment, but may be added in the future for easier adoption
- **Secrets**:
  - `ibm-license-service-cert`
  - `ibm-licensing-token`
  - `ibm-licensing-upload-token`
- **Service**: `ibm-licensing-service-instance`
- **Route**: `ibm-licensing-service-instance`

### Currently Excluded Features

The following features are not included in the current implementation:
- Prometheus data import
- Sending data to Software Central
- Sending data to License Service Reporter (LSR)
- Gateway API
- Namespace scoping

**Example**: The `ibm-licensing-service-prometheus-cert` secret and `ibm-licensing-service-prometheus` service are currently skipped, because for now we don't support Prometheus data import. If we want to support it we need to add new field to the values.yaml and create those resources conditionally.

## Resource Details

### Service

The `ibm-licensing-service-instance` service has an annotation that creates the `ibm-license-service-cert-internal` secret.

### ConfigMaps

#### `ibm-licensing-info`
- Retrieves URL from the Custom Resource (CR), just like `ibm-licensing-upload-config`
- Not mounted anywhere in the deployment
- **Status**: Skipped

#### `ibm-licensing-upload-config`
- Retrieves certificate from the `ibm-license-service-cert-internal` secret (created by OpenShift)
- Creates service URL based on the CR (with HTTP or HTTPS)
- Could potentially be retrieved from the secret as well
- Not currently mounted in the deployment
- **Status**: Skipped

**Note**: No ConfigMaps are currently mounted in the deployment. The `ibm-licensing-upload-config` ConfigMap may be added in the future to simplify License Service usage.

### Secrets

#### Certificate Secrets

- **`ibm-license-service-cert-internal`**: Not created by the License Service operator
- **`ibm-license-service-cert`**: Created by the operator in `reconcileCertificateSecrets()`

#### Token Secrets

- **`ibm-licensing-default-reader-token`**: Currently skipped (not mounted)
- **`ibm-licensing-service-account-token`**: Currently skipped (not mounted)

**Note**: These two token secrets are relatively easy to add. They use annotations to allow Kubernetes to automatically create resources. For example, to obtain an API token for a ServiceAccount, create a new Secret with the special annotation `kubernetes.io/service-account.name`. One account name is hardcoded, while the other is read from the CR (can be read from `values.yaml` in Helm).

- **`ibm-licensing-token`**: Created by the operator (random string)
- **`ibm-licensing-upload-token`**: Created by the operator (random string)

### RBAC
Will be created by using kustomize and selecting specific resources only, similar to how `make generate-yaml-argo-cd` currently works.
RBAC for operator will be ignored, as we don't have operator in this helm chart.

### CRDs
Will be created by using kustomize and selecting specific resources only, similar to how `make generate-yaml-argo-cd` currently works.
CRD IBMLicensing is only used by operator and CRD IBMLicensingQuerySource is only used by prometheus, so those will be skipped.

### Mounted Resources (those are required for the Licesning Service to function properly)

The License Service deployment mounts the following resources:
- API token (secret)
- Upload token (secret)
- TLS certificate (secret)
- Empty temporary directory