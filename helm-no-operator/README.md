# IBM License Service Helm Chart (no-operator)

Installs IBM License Service directly on a Red Hat OpenShift cluster without the IBM Licensing Operator.

## Supported platforms

Red Hat OpenShift Container Platform 4.10 or newer on Linux x86_64, ppc64le, and s390x.

## Prerequisites

- Helm 3
- Cluster-admin rights (CRDs and ClusterRoles are installed by this chart)
- An image pull secret in the target namespace (default name: `ibm-entitlement-key`)

## Configuration

### Global

| Parameter | Description | Default |
|-----------|-------------|---------|
| `global.imagePullPrefix` | Registry prefix for all images | `icr.io` |
| `global.imagePullSecret` | Name of the image pull secret | `ibm-entitlement-key` |

### Core

| Parameter | Description | Default |
|-----------|-------------|---------|
| `ibmLicensing.namespace` | Namespace to deploy License Service into | `ibm-licensing` |
| `ibmLicensing.watchNamespace` | Namespace(s) watched for license data collection | `ibm-licensing` |
| `ibmLicensing.excludeNamespace` | Comma-separated namespaces excluded from collection (ignored when `nssEnabled: true`) | `""` |
| `ibmLicensing.imageRegistryNamespaceOperand` | Registry sub-path for the operand image | `cpopen/cpfs` |
| `ibmLicensing.ibmLicensingVersion` | Operand image tag | `4.2.25` |

### RBAC

| Parameter | Description | Default |
|-----------|-------------|---------|
| `ibmLicensing.createRBAC` | Create ServiceAccounts, Roles, RoleBindings, ClusterRoles, and ClusterRoleBindings. Set to `false` when a cluster administrator pre-creates all RBAC resources | `true` |
| `ibmLicensing.createRBACInWatchedNamespaces` | Create RBAC in each watched namespace (required when `nssEnabled: true`) | `true` |
| `ibmLicensing.createRBACReaderRole` | Create `ibm-licensing-default-reader` ClusterRole and binding for API consumers | `true` |

### Features

| Parameter | Description | Default |
|-----------|-------------|---------|
| `ibmLicensing.spec.datasource` | Data source mode | `datacollector` |
| `ibmLicensing.spec.httpsEnable` | Serve the API over HTTPS | `true` |
| `ibmLicensing.spec.httpsCertsSource` | TLS certificate source | `external` |
| `ibmLicensing.spec.enableInstanaMetricCollection` | Enable Instana metric collection | `false` |
| `ibmLicensing.spec.features.nssEnabled` | Enable namespace-scope mode. When `true`, the restricted service account and per-namespace RBAC are used | `false` |
| `ibmLicensing.spec.features.nodeCpuCappingEnabled` | Include node-level CPU capping | `true` |
| `ibmLicensing.spec.features.kubeRBACAuthEnabled` | Protect the API with kube RBAC token review | `true` |
| `ibmLicensing.spec.features.customResourcesEnabled` | Read `IBMLicensing*` custom resources | `true` |

### Resources

| Parameter | Description | Default |
|-----------|-------------|---------|
| `ibmLicensing.spec.resources.limits.cpu` | CPU limit for the container and init container | `500m` |
| `ibmLicensing.spec.resources.limits.memory` | Memory limit | `1Gi` |
| `ibmLicensing.spec.resources.requests.cpu` | CPU request | `200m` |
| `ibmLicensing.spec.resources.requests.memory` | Memory request | `256Mi` |
| `ibmLicensing.spec.resources.requests.ephemeralStorage` | Ephemeral storage request | `256Mi` |

### Pod scheduling & identity

| Parameter | Description | Default |
|-----------|-------------|---------|
| `ibmLicensing.operand.serviceAccount.name` | Service account name override. Default resolves to `ibm-license-service` (or `ibm-license-service-restricted` when `nssEnabled: true`). Intended for use with `createRBAC: false` when the SA is pre-created | `""` |
| `ibmLicensing.operand.nodeSelector` | Node selector key/value pairs | `{}` |
| `ibmLicensing.operand.affinity` | Affinity rules **merged** with the IBM default `kubernetes.io/arch` node affinity. User rules take precedence on key conflicts | `{}` |

### Pod metadata

| Parameter | Description | Default |
|-----------|-------------|---------|
| `ibmLicensing.operand.podAnnotations` | Extra annotations added to the pod. IBM product annotations (`productID`, `productName`, `productMetric`) cannot be overridden | `{}` |
| `ibmLicensing.operand.podLabels` | Extra labels added to the pod. IBM selector labels (`app`, `component`, `licensing_cr`, etc.) cannot be overridden | `{}` |
