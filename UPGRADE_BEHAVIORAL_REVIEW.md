# Behavioral Review: ILS-1821 Major Upgrade

## Context

The IBM Licensing Operator underwent a major dependency upgrade (branch `ILS-1821-major-upgrade`).
17 commits from `ae222b8` to `6128349` upgraded:
- controller-runtime v0.12 → v0.23
- Kubernetes libs v0.25 → v0.35
- prometheus-operator v0.57 → v0.89
- Go 1.25 → 1.26

This document categorizes every change as: **intentional behavioral change**, **bug**, or **pure API adaptation** (no behavioral change).

## Scope

6 modified Go source files + 3 new vendored type files + 3 test files.

---

## 1. Manager Setup (`main.go`)

| Change | Category |
|--------|----------|
| Cache: `controller-filtered-cache` → native `cache.Options.ByObject` with `DefaultNamespaces` | **Intentional** — same GVKs (Secret, Deployment, Pod) with same label selector. New cache returns "not found" for objects missing the label (no API server fall-through). |
| Metrics: `MetricsBindAddress` → `metricsserver.Options{BindAddress}` | **Pure API adaptation** — same address `:8080` |
| Webhook: `Port: 9443` → `webhook.NewServer(webhook.Options{Port: 9443})` | **Pure API adaptation** — same port |
| RHMP scheme registration: import path changed to local `pkg/rhmp/v1beta1` | **Pure API adaptation** — registers same `GroupVersion` with same types |

## 2. Reconciliation Controller (`controllers/ibmlicensing_controller.go`)

| Location | Change | Category |
|----------|--------|----------|
| Line 593 | `r.Client.Get` → `r.Reader.Get` for internal cert in `reconcileConfigMaps` | **Intentional** — ServiceCA-created cert lacks the label, invisible to ByObject cache. Reader bypasses cache. |
| Line 835 | `r.Client.Get` → `r.Reader.Get` for internal cert in `reconcileRouteWithCertificates` | **Intentional** — same reason. Also fixes log message from "external" to "internal" certificate. |
| Lines 746-758 | New `replicasMismatch` check before `ShouldUpdateDeployment` | **Intentional** — old code only compared PodTemplateSpec, ignoring replica count drift. `GetLicensingDeployment` always sets `Spec.Replicas` explicitly (not nil), so nil-panic is not a risk. |
| Lines 1098-1120 | Upgrade migration: label patching in `reconcileResourceExistence` when `AlreadyExists` | **Intentional** — patches labels on pre-upgrade resources that lack them. Idempotent. |
| Imports | `coreos/prometheus-operator` → `prometheus-operator/prometheus-operator`, `redhat-marketplace` → local `pkg/rhmp/v1beta1` | **Pure API adaptation** |

## 3. Helper Functions (`controllers/resources/helper.go`)

The diff was import-only (6 lines), but the underlying types changed, creating **silent type changes** in comparisons.

### Bugs Found and Fixed

| Line | Comparison | Old Type | New Type | Issue | Status |
|------|-----------|----------|----------|-------|--------|
| 238 | `expectedEndpoint.Scheme != foundEndpoint.Scheme` | `string` | `*Scheme` | **BUG: pointer comparison**. Two `*Scheme` pointers with same value but different allocations always compare as not-equal. Caused unnecessary ServiceMonitor update on **every reconcile**. | **FIXED** — now uses `equalStringPointers()` |
| 270 | `expectedRelabeling.Replacement != foundRelabeling.Replacement` | `string` | `*string` | **BUG: pointer comparison**. Same issue. Triggered unnecessary update every reconcile when HTTPS enabled. | **FIXED** — now uses `equalStringPointers()` |

### Safe Comparisons (no issue)

| Line | Field | Type | Why Safe |
|------|-------|------|----------|
| 252 | `Interval` | `Duration` (named `string`) | Value comparison works correctly |
| 271 | `TargetLabel` | `string` | Unchanged, value comparison |
| 324 | `Action` | `string` | Unchanged, value comparison |
| 325 | `Regex` | `string` | Unchanged, value comparison |
| 327 | `SourceLabels[0]` | `LabelName` (named `string`) | Value comparison works correctly |
| 283-285 | `TLSConfig` | `*TLSConfig` | Uses `apieq.Semantic.DeepEqual` — correct |

## 4. ServiceMonitor Construction (`controllers/resources/service/service_monitor.go`)

All changes are **pure API adaptations**:
- `BearerTokenSecret` field removed (was explicit zero-value)
- `TLSConfig` wrapped in `HTTPConfigWithProxyAndTLSFiles` (inline JSON, same serialized path)
- `Interval` cast to `monitoringv1.Duration`
- `getScheme()` returns `*monitoringv1.Scheme` instead of `string`
- `Replacement` uses `*string` instead of `string`
- `[]*RelabelConfig` → `[]RelabelConfig`
- `[]string` → `[]monitoringv1.LabelName`

## 5. MeterDefinition Construction (`controllers/resources/service/meter_definition.go`)

**Pure API adaptation** — import paths only. Vendored types in `pkg/rhmp/` are field-for-field compatible.

## 6. Secret Management (`controllers/resources/service/secrets.go`)

**Intentional behavioral change**. `GetDefaultReaderToken` and `GetServiceAccountSecret` now include `Labels: LabelsForMeta(instance)`. Required for the new ByObject cache to see these secrets. Adding labels has no side effects on secret functionality.

## 7. Vendored RHMP Types (`pkg/rhmp/`)

New files for dependency decoupling. Registers `marketplace.redhat.com/v1beta1` with `MeterDefinition` and `MeterDefinitionList`. Types pinned to Feb 2023 upstream version. Not a behavioral change.

## 8. Test Files

| File | Change Type | Test Coverage Impact |
|------|------------|---------------------|
| `ibmlicensing_controller_test.go` | Import-only | **None** |
| `suite_test.go` | Import + API adaptation | **None** — but tests don't use ByObject cache, so cache-related bugs aren't testable |
| `operatorgroups_test.go` | Assertion weakened (DeepEqual → name-only) | **Minor** — sufficient for lookup verification but less strict |

---

## Summary

### Intentional Behavioral Changes (5)
1. Cache strategy: `controller-filtered-cache` → native `cache.Options.ByObject` (stricter: no API server fall-through)
2. `r.Client.Get` → `r.Reader.Get` for ServiceCA internal cert (2 locations)
3. Labels added to 2 ServiceAccountToken secrets for cache visibility
4. Upgrade migration: label patching for pre-upgrade resources lacking labels
5. Replica count mismatch detection in deployment reconciliation

### Bugs Found and Fixed (2)
1. **`helper.go:238`** — `Scheme` pointer comparison → value comparison via `equalStringPointers()`
2. **`helper.go:270`** — `Replacement` pointer comparison → value comparison via `equalStringPointers()`

### Pure API Adaptations (14 items, no behavioral change)
Import path changes, type casts, struct nesting, slice element types, removed zero-value fields, metrics/webhook config restructuring.
