# Reconcile Loop Fix — controller-runtime Cache Breaking Change

## Root Cause

The upgrade from `controller-runtime v0.12.3` + `IBM/controller-filtered-cache` to `v0.23.1`
with native `ByObject` cache changed a critical behavior:

**Old behavior**: The IBM custom filtered cache could fall-through to the API server for `Get()`
calls even when a Secret lacked the `release=ibm-licensing-service` label.

**New behavior**: `ByObject` cache strictly returns "not found" for any Secret without that label —
no API server fallback.

This caused `reconcileDefaultReaderToken` (3rd in the reconcile list) to enter an infinite loop:
1. `r.Client.Get()` → "not found" (secret exists but lacks label → not in cache)
2. `r.Client.Create()` → "already exists"
3. `time.Sleep(5s)` + `return Requeue: true`
4. **Never reaches `updateStatus()` → "reconcile all done" never logged**

## Affected Resources

| Secret name | Function | Problem |
|---|---|---|
| `ibm-licensing-default-reader-token` | `GetDefaultReaderToken` | Missing `release` label |
| `ibm-licensing-service-account-token` | `GetServiceAccountSecret` | Missing `release` label |
| `ibm-license-service-cert-internal` | Created by OpenShift ServiceCA | External controller, label never added |

## Three Fixes Applied

### 1. `controllers/resources/service/secrets.go` — missing labels

Added `Labels: LabelsForMeta(instance)` to `GetDefaultReaderToken` and `GetServiceAccountSecret`.
These were the only secret-creating functions missing the `release=ibm-licensing-service` label
required by the new `ByObject` filtered cache.

### 2. `controllers/ibmlicensing_controller.go` — OCP internal cert

Changed internal cert (`ibm-license-service-cert-internal`) fetches in:
- `reconcileRouteWithCertificates`
- `reconcileConfigMaps`

…to use `r.Reader.Get()` instead of `r.Client.Get()`. On OCP, this cert is created by OpenShift
ServiceCA without the label, so it is never present in the filtered informer cache.

### 3. `controllers/ibmlicensing_controller.go` — upgrade migration

When `Create` returns `AlreadyExists` inside `reconcileResourceExistence`, the code now:
1. Uses `r.Reader.Get()` (bypasses cache) to fetch the existing resource
2. Patches any labels from the expected resource onto it
3. Calls `r.Client.Update()` to persist the labels

This automatically handles existing clusters that already have the unlabeled secrets from
the pre-upgrade operator — without requiring manual intervention.

## Rule for Future Work

When adding new Secret-creating functions, always include `Labels: LabelsForMeta(instance)`.
For Secrets created by external controllers (ServiceCA, Kubernetes token controller, etc.),
use `r.Reader.Get()` instead of `r.Client.Get()` to bypass the label-filtered cache.
