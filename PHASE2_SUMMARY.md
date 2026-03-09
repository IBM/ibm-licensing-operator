# Phase 2+3 Summary — Upgrade k8s.io/* and controller-runtime

Completed: 2026-03-09

## Objective

Upgrade the core Kubernetes and controller-runtime dependencies from their
replace-pinned versions (k8s 1.25 / controller-runtime v0.12.3) to
controller-runtime v0.19.4 / k8s.io v0.31.4, and fix all resulting API
changes in the codebase.

## Why phases 2 and 3 were merged

The upgrade plan originally defined Phase 2 (update go.mod versions) and
Phase 3 (fix code for new controller-runtime API) as separate steps, with
Phase 2 expecting compilation errors resolved in Phase 3. In practice,
`go mod tidy` itself could not complete until the two blocking dependencies
(`controller-filtered-cache` and `redhat-marketplace-operator`) were removed
from go.mod, because both transitively import k8s.io packages that no longer
exist in v0.31 (e.g. `k8s.io/api/auditregistration/v1alpha1`,
`k8s.io/apimachinery/pkg/util/clock`, `github.com/googleapis/gnostic/OpenAPIv2`).
Removing those dependencies required code changes (Phase 3 / partial Phase 4f),
so the phases had to be executed together.

---

## go.mod changes

### Replace directives removed

All seven replace directives were removed:

| Directive | Reason |
|-----------|--------|
| `k8s.io/api => v0.25.7` | Pinned to Kubernetes 1.25; now using v0.31.4 directly |
| `k8s.io/apimachinery => v0.25.7` | Same |
| `k8s.io/client-go => v0.25.7` | Same; also resolved the `v12.0.0+incompatible` fossil |
| `sigs.k8s.io/controller-runtime => v0.12.3` | Pinned to 3-year-old release; now using v0.19.4 |
| `sigs.k8s.io/controller-runtime/pkg/cache => v0.10.0` | Sub-path replace no longer needed |
| `sigs.k8s.io/controller-runtime/pkg/client => v0.6.4` | Sub-path replace no longer needed |
| `cloud.google.com/go => v0.110.0` | Was a workaround for ambiguous import caused by redhat-marketplace-operator; no longer needed after that dependency was removed |
| `github.com/emicklei/go-restful/v3 => v3.10.1` | Was pinning to an older version; no longer needed |

### Version upgrades

| Module | Before (effective) | After |
|--------|--------------------|-------|
| `k8s.io/api` | v0.25.7 | v0.31.4 |
| `k8s.io/apimachinery` | v0.25.7 | v0.31.4 |
| `k8s.io/client-go` | v0.25.7 | v0.31.4 |
| `k8s.io/apiextensions-apiserver` | v0.27.2 | v0.31.4 |
| `sigs.k8s.io/controller-runtime` | v0.12.3 | v0.19.4 |
| `k8s.io/klog/v2` | v2.90.1 | v2.130.1 |
| `k8s.io/kube-openapi` | v0.0.0-20230501 | v0.0.0-20240228 |
| `github.com/go-logr/logr` | v1.2.4 | v1.4.2 |
| `go.uber.org/zap` | v1.21.0 | v1.26.0 |
| `go.uber.org/multierr` | v1.7.0 | v1.11.0 |

### Dependencies removed

| Module | Reason |
|--------|--------|
| `github.com/IBM/controller-filtered-cache` | Replaced by native `cache.Options` in controller-runtime v0.15+ |
| `github.com/redhat-marketplace/redhat-marketplace-operator/v2` | Types vendored locally (see justification below) |

### Deprecated/moved packages cleaned up by go mod tidy

These packages were present before the upgrade and are now gone:

- `github.com/google/gnostic` — replaced by `github.com/google/gnostic-models`
- `k8s.io/klog` v1 — superseded by `k8s.io/klog/v2`
- `go.uber.org/atomic` — merged into `go.uber.org/zap` and `multierr`
- `github.com/matttproud/golang_protobuf_extensions` — merged into prometheus client packages
- `cloud.google.com/go/compute/metadata` — no longer a transitive dependency
- ~20 other indirect dependencies that were only required by the removed packages

---

## Code changes

### main.go — controller-runtime v0.19 API migration

controller-runtime v0.15 introduced breaking changes to `ctrl.Options`.
The following fields were removed and replaced:

```go
// BEFORE (v0.12.x API)
mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
    Scheme:             scheme,
    MetricsBindAddress: metricsAddr,
    Port:               9443,
    LeaderElection:     enableLeaderElection,
    LeaderElectionID:   "e1f51baf.ibm.com",
    NewCache:           cache.MultiNamespacedFilteredCacheBuilder(gvkLabelMap, watchNamespaces),
})

// AFTER (v0.19.x API)
mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
    Scheme: scheme,
    Metrics: metricsserver.Options{
        BindAddress: metricsAddr,
    },
    WebhookServer:    webhook.NewServer(webhook.Options{Port: 9443}),
    LeaderElection:   enableLeaderElection,
    LeaderElectionID: "e1f51baf.ibm.com",
    Cache: cache.Options{
        DefaultNamespaces: defaultNamespaces,
        ByObject: map[client.Object]cache.ByObject{
            &corev1.Secret{}:     {Label: licensingLabelSelector},
            &appsv1.Deployment{}: {Label: licensingLabelSelector},
            &corev1.Pod{}:        {Label: licensingLabelSelector},
        },
    },
})
```

This eliminates the dependency on `github.com/IBM/controller-filtered-cache`
entirely. The `cache.Options.DefaultNamespaces` field provides equivalent
multi-namespace watching, and `cache.Options.ByObject` with `Label` selectors
provides equivalent per-GVK label filtering. Both features are available
natively in controller-runtime since v0.15.

### controllers/suite_test.go

Same `MetricsBindAddress` → `Metrics` migration. Also removed the obsolete
`Namespace: ""` field (in v0.15+, not setting `DefaultNamespaces` already
means "watch all namespaces").

### Import path updates (6 files)

All imports of `github.com/redhat-marketplace/redhat-marketplace-operator/v2/apis/marketplace/...`
were replaced with local `github.com/IBM/ibm-licensing-operator/pkg/rhmp/...`.

---

## Vendoring redhat-marketplace-operator types — justification

### Why was this necessary?

The `github.com/redhat-marketplace/redhat-marketplace-operator/v2` module is
incompatible with k8s.io v0.31 in **both directions**:

1. **The old version (v2.0.0-20230228)** — which this operator was using —
   was built against k8s.io v0.25. Its transitive dependency graph pulls in
   packages that were removed in k8s.io v0.31:
   - `k8s.io/api/auditregistration/v1alpha1` (removed in Kubernetes 1.27)
   - `k8s.io/api/batch/v2alpha1` (removed in Kubernetes 1.27)
   - `k8s.io/api/settings/v1alpha1` (removed in Kubernetes 1.27)
   - `k8s.io/apimachinery/pkg/util/clock` (moved to `k8s.io/utils/clock`)
   - `k8s.io/client-go/metadata` (not available in the `v12.0.0+incompatible` require)
   - `github.com/googleapis/gnostic/OpenAPIv2` (moved to `gnostic-models`)
   - `sigs.k8s.io/controller-runtime/pkg/runtime/inject` (removed in controller-runtime v0.15)

   Because Go's module resolution requires all transitively imported packages
   to be resolvable, `go mod tidy` fails when this old version coexists with
   k8s.io v0.31 in the same module graph. This is not a theoretical issue —
   it was the actual blocker that prevented Phase 2 from completing.

2. **The latest version (v2.0.0-20260302)** — restructured its package layout
   and no longer contains the packages we import:
   - `apis/marketplace/v1beta1` — does not exist in the latest version
   - `apis/marketplace/common` — does not exist in the latest version

   `go mod tidy` reports: *"module found but does not contain package"*.

### Why couldn't we use a newer compatible version?

There is no version of `redhat-marketplace-operator/v2` that simultaneously:
- Contains the `apis/marketplace/v1beta1` and `apis/marketplace/common` packages
- Is compatible with k8s.io v0.31+ and controller-runtime v0.15+

The module jumped from old k8s (v0.25-era) packages directly to a restructured
layout that removed the API packages entirely. There is no intermediate release
that bridges the gap.

### Why not use replace directives?

Adding replace directives for all the broken transitive packages
(`k8s.io/api/auditregistration/v1alpha1`, etc.) is not feasible because:
- These are not separate modules — they are sub-packages within `k8s.io/api`
  which is already at v0.31.4. You cannot replace sub-packages of a module.
- The `sigs.k8s.io/controller-runtime/pkg/runtime/inject` package was deleted
  entirely in controller-runtime v0.15 and cannot be substituted.

### What was vendored?

Only the CRD type definitions actually used by ibm-licensing-operator were
copied into `pkg/rhmp/`. The vendored code is minimal:

| File | Contents |
|------|----------|
| `pkg/rhmp/common/types.go` | `GroupVersionKind`, `WorkloadType`, `WorkloadTypeService`, `MetricType`, `NamespacedNameReference` |
| `pkg/rhmp/v1beta1/types.go` | `MeterDefinition`, `MeterDefinitionList`, `MeterDefinitionSpec`, `ResourceFilter`, `NamespaceFilter`, `OwnerCRDFilter`, `MeterWorkload`, and supporting types |
| `pkg/rhmp/v1beta1/groupversion.go` | `SchemeBuilder`, `AddToScheme`, `GroupVersion` for CRD scheme registration |
| `pkg/rhmp/v1beta1/zz_generated.deepcopy.go` | `DeepCopyObject` and `DeepCopyInto` implementations required by `runtime.Object` interface |

The `MeterDefinitionStatus` type was intentionally left empty because the
operator never reads or writes status fields — it only creates
`MeterDefinition` objects and registers the type with the scheme.

### Future considerations

The original upgrade plan (Phase 4f) suggested evaluating whether the
dependency could be replaced by using only the CRD schema directly. This
vendoring accomplishes exactly that. If the redhat-marketplace-operator
publishes a new standalone API types module compatible with modern k8s
versions, the vendored types can be replaced with that module.

---

## Verification

All checks passed after changes:

- `go mod tidy` — clean (no stale dependencies)
- `go build ./...` — clean
- `go vet ./...` — clean
- No references to `controller-filtered-cache` in any `.go` file
- No references to `redhat-marketplace-operator` in go.mod
- No replace directives remaining in go.mod

## Files changed

| File | Change |
|------|--------|
| `go.mod` | Removed all replace directives; upgraded k8s.io/* to v0.31.4, controller-runtime to v0.19.4; removed controller-filtered-cache and redhat-marketplace-operator dependencies |
| `go.sum` | Regenerated by `go mod tidy` |
| `main.go` | Migrated to controller-runtime v0.19 API (`Metrics`, `WebhookServer`, `Cache` options); replaced controller-filtered-cache with native `cache.Options`; updated rhmp import |
| `controllers/suite_test.go` | Same `ctrl.Options` migration; updated rhmp import |
| `controllers/ibmlicensing_controller.go` | Updated rhmp import |
| `controllers/ibmlicensing_controller_test.go` | Updated rhmp import |
| `controllers/resources/helper.go` | Updated rhmp import |
| `controllers/resources/service/meter_definition.go` | Updated rhmp + rhmpcommon imports |
| `pkg/rhmp/common/types.go` | **New** — vendored common types |
| `pkg/rhmp/v1beta1/types.go` | **New** — vendored CRD types |
| `pkg/rhmp/v1beta1/groupversion.go` | **New** — scheme registration |
| `pkg/rhmp/v1beta1/zz_generated.deepcopy.go` | **New** — DeepCopy implementations |
