# ILS-1821 Major Upgrade — Cache Breaking Change

The upgrade from controller-runtime v0.12.3 (via IBM/controller-filtered-cache) to v0.23.1 (ByObject) broke reconciliation.

**Why:** Old IBM filtered cache could fall through to the API server for Secret.Get() calls even when the secret lacked the label filter. New ByObject cache strictly returns "not found" for secrets without `release=ibm-licensing-service` label.

## Affected resources

1. `GetDefaultReaderToken` — `ibm-licensing-default-reader-token` — was missing label
2. `GetServiceAccountSecret` — `ibm-licensing-service-account-token` — was missing label
3. OCP: `ibm-license-service-cert-internal` — created by OpenShift ServiceCA, never has the label
4. OCP custom cert: `ibm-licensing-certs` — user-provided secret, never has the operator label

## Rule going forward

When adding new Secret-creating functions, always include `Labels: LabelsForMeta(instance)`. For resources created by external controllers (ServiceCA etc.) or user-provided secrets, use `r.Reader.Get()` instead of `r.Client.Get()`.

## Fixes applied

- `controllers/resources/service/secrets.go`: Added `Labels: LabelsForMeta(instance)` to both functions
- `controllers/ibmlicensing_controller.go`:
  - `reconcileRouteWithCertificates`: internal cert Get → `r.Reader.Get()`
  - `reconcileRouteWithCertificates`: external custom cert Get → `r.Reader.Get()` (user-provided, no operator label)
  - `reconcileConfigMaps`: internal cert Get → `r.Reader.Get()`
  - `reconcileResourceExistence`: on AlreadyExists, patch missing labels via Reader (migration for existing clusters)
