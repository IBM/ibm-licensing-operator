# Dependency & Go Upgrade Plan — ibm-licensing-operator

## Current State (as of 2026-03-06)

### Go toolchain
| Item | Current | Latest |
|------|---------|--------|
| go directive in go.mod | 1.25.6 | 1.26.1 |
| Installed Go | 1.26.1 | — |

### Direct dependencies (key)
| Module | Required | Effective (replace) | Latest |
|--------|----------|---------------------|--------|
| k8s.io/api | v0.27.2 | **v0.25.7** (replace!) | v0.35.2 |
| k8s.io/apimachinery | v0.27.2 | **v0.25.7** (replace!) | v0.35.2 |
| k8s.io/client-go | v12.0.0+incompatible | **v0.25.7** (replace!) | v0.35.2 |
| sigs.k8s.io/controller-runtime | v0.15.0 | **v0.12.3** (replace!) | v0.23.3 |
| github.com/coreos/prometheus-operator | v0.41.0 | v0.41.0 | **MOVED** (see §3) |
| github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring | — (indirect) | v0.57.0 | v0.89.0 |
| github.com/IBM/controller-filtered-cache | v0.3.5 | v0.3.5 | v0.3.6 |
| github.com/IBM/operand-deployment-lifecycle-manager | v1.21.0 | v1.21.0 | v1.23.5 |
| github.com/openshift/api | v0.0.0-20230306 | v0.0.0-20230306 | v0.0.0-20260306 |
| github.com/operator-framework/api | v0.17.7 | v0.17.7 | v0.41.0 |
| go.uber.org/zap | v1.21.0 | v1.21.0 | v1.27.1 |
| github.com/redhat-marketplace/redhat-marketplace-operator/v2 | v2.0.0-20230228 | v2.0.0-20230228 | — |

### Build tools (Makefile)
| Tool | Current | Notes |
|------|---------|-------|
| operator-sdk | v1.32.0 | Outdated |
| opm | v1.26.2 | Outdated |
| kustomize | v4.5.7 | Outdated (v5.x available) |
| controller-gen | v0.14.0 | Outdated |
| yq | v4.30.5 | Outdated |

---

## Root Problems

### P1 — replace directives defeat the require statements
`go.mod` has replace directives that pin `k8s.io/{api,apimachinery,client-go}` to
**v0.25.7** and `sigs.k8s.io/controller-runtime` to **v0.12.3**, despite the `require`
block listing much newer versions. The effective runtime versions are the pinned ones.

### P2 — Deprecated import path: coreos/prometheus-operator
`github.com/coreos/prometheus-operator` was moved to
`github.com/prometheus-operator/prometheus-operator`. The old module path is archived
and receives no security updates. Used in:
- `main.go`
- `controllers/ibmlicensing_controller.go`
- `controllers/resources/helper.go`
- `controllers/resources/service/service_monitor.go`

### P3 — Deprecated controller-runtime Options API
`ctrl.Options` fields used in `main.go` that changed in controller-runtime ≥ v0.15:
- `MetricsBindAddress` → `Metrics.BindAddress`
- `NewCache` builder function signature changed in v0.15+ (cache.Options struct)
- `Port` field handling changed

### P4 — IBM/controller-filtered-cache compatibility
`github.com/IBM/controller-filtered-cache` (v0.3.5, latest v0.3.6) was built against
old controller-runtime (v0.12.x). Its `MultiNamespacedFilteredCacheBuilder` signature
may change or break when upgrading controller-runtime. This library is no longer
actively maintained (last release May 2023). If it cannot be upgraded, it must be
replaced with the standard `cache.Options{DefaultNamespaces: ...}` approach
available natively in controller-runtime ≥ v0.15.

### P5 — k8s.io/client-go legacy incompatible version
`k8s.io/client-go v12.0.0+incompatible` in the require block is a fossil from before
the module was properly versioned. The replace directive fixes it to v0.25.7, but the
require line is misleading and should be corrected.

### P6 — Security and CVE exposure
Several transitive dependencies have known CVEs in their pinned versions
(e.g., old golang.org/x/net, golang.org/x/crypto). Upgrading the chain will bring
in patched versions.

---

## Upgrade Strategy

The upgrade is broken into **5 phases** ordered by risk and dependency. Each phase
should be committed and verified (build + tests) before proceeding to the next.

---

## Phase 1 — Fix prometheus-operator import path
**Risk: Low | Effort: Low**

Replace every occurrence of `github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1`
with `github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1`.

Files to change:
- `main.go` (import alias `monitoringv1`)
- `controllers/ibmlicensing_controller.go`
- `controllers/resources/helper.go`
- `controllers/resources/service/service_monitor.go`

Update `go.mod`:
- Remove `github.com/coreos/prometheus-operator v0.41.0` from require
- Add `github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.57.0`
  (or latest compatible with current k8s version)

Verification: `go build ./...` must pass.

---

## Phase 2 — Upgrade k8s.io/* and controller-runtime together
**Risk: High | Effort: High**

These three must be upgraded in lockstep because controller-runtime requires specific
k8s.io minor versions.

Target versions (controller-runtime v0.19.x targets k8s v0.31.x):
| Module | Target |
|--------|--------|
| k8s.io/api | v0.31.x |
| k8s.io/apimachinery | v0.31.x |
| k8s.io/apiextensions-apiserver | v0.31.x |
| k8s.io/component-base | v0.31.x |
| k8s.io/client-go | v0.31.x |
| k8s.io/kube-openapi | compatible |
| sigs.k8s.io/controller-runtime | v0.19.4 |

Steps:
1. Remove all `replace` directives for `k8s.io/*` and `sigs.k8s.io/controller-runtime`.
2. Remove the `sigs.k8s.io/controller-runtime/pkg/cache` and
   `sigs.k8s.io/controller-runtime/pkg/client` sub-path replaces.
3. Update `require` to target versions above.
4. Run `go mod tidy`.

Verification: `go build ./...` — expect compilation errors from API changes (see Phase 3).

---

## Phase 3 — Fix code for new controller-runtime API
**Risk: Medium | Effort: Medium**

Changes required in `main.go` due to controller-runtime v0.15+ API changes:

```go
// BEFORE (v0.12.x)
mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
    Scheme:             scheme,
    MetricsBindAddress: metricsAddr,
    Port:               9443,
    LeaderElection:     enableLeaderElection,
    LeaderElectionID:   "e1f51baf.ibm.com",
    NewCache:           cache.MultiNamespacedFilteredCacheBuilder(gvkLabelMap, watchNamespaces),
})

// AFTER (v0.19.x)
mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
    Scheme: scheme,
    Metrics: metricsserver.Options{
        BindAddress: metricsAddr,
    },
    WebhookServer: webhook.NewServer(webhook.Options{Port: 9443}),
    LeaderElection:   enableLeaderElection,
    LeaderElectionID: "e1f51baf.ibm.com",
    Cache: cache.Options{
        DefaultNamespaces: buildNamespaceMap(watchNamespaces),
        ByObject: map[client.Object]cache.ByObject{
            &corev1.Secret{}:    {Label: labelSelectorForLicensing()},
            &appsv1.Deployment{}: {Label: labelSelectorForLicensing()},
            &corev1.Pod{}:       {Label: labelSelectorForLicensing()},
        },
    },
})
```

This eliminates the dependency on `github.com/IBM/controller-filtered-cache` entirely
(the standard cache.Options provides equivalent filtering since controller-runtime v0.15).

Additional changes:
- Add imports: `sigs.k8s.io/controller-runtime/pkg/metrics/server`,
  `sigs.k8s.io/controller-runtime/pkg/webhook`,
  `sigs.k8s.io/controller-runtime/pkg/cache`
- Remove import of `github.com/IBM/controller-filtered-cache/filteredcache`
- Remove `github.com/IBM/controller-filtered-cache` from `go.mod`
- Check all other controller-runtime API usage across controllers for deprecations
  (e.g., reconcile.Result fields, client options)

Verification: `go build ./...` and `go vet ./...` must pass.

---

## Phase 4 — Upgrade remaining direct dependencies
**Risk: Low-Medium | Effort: Medium**

Upgrade in order (run `go get <module>@<version>` + `go mod tidy` after each group):

### 4a. Prometheus operator
```
go get github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring@v0.89.0
```
Check for any API changes in `monitoringv1` types (RelabelConfig, Endpoint fields).

### 4b. IBM ODLM
```
go get github.com/IBM/operand-deployment-lifecycle-manager@v1.23.5
```
Verify API compatibility in `controllers/resources/operandrequests.go` and
`controllers/resources/namespacescopes.go`.

### 4c. OpenShift API
```
go get github.com/openshift/api@v0.0.0-20260101000000  # latest stable tag or commit
```
Verify Route and ServiceCA types still compatible.

### 4d. operator-framework/api
```
go get github.com/operator-framework/api@v0.41.0
```
Verify `OperatorGroup` types still compatible in `controllers/operatorgroup_cleaner.go`.

### 4e. go.uber.org/zap
```
go get go.uber.org/zap@v1.27.1
```
The `zap.New(func(o *zap.Options){...})` functional options API is unchanged.
Verify `zapcore.RFC3339TimeEncoder` still available.

### 4f. redhat-marketplace-operator
This is a pseudo-version pinned to a specific commit. Evaluate whether:
- A newer tagged version exists and is compatible
- The dependency can be replaced by using only the CRD schema directly
  (the operator only uses `meterdefv1beta1.MeterDefinitionList` for scheme registration)

### 4g. General stdlib modernization
```
go get golang.org/x/net@latest golang.org/x/crypto@latest golang.org/x/sys@latest
```

Run `go mod tidy` after all updates.

---

## Phase 5 — Go version, tooling, and cleanup
**Risk: Low | Effort: Low**

### 5a. Update go directive
In `go.mod`, update:
```
go 1.26.1
```
Also add a `toolchain` directive if needed for reproducibility.

### 5b. Update Makefile tool versions
| Tool | Current | Target |
|------|---------|--------|
| operator-sdk | v1.32.0 | v1.40.0 (check latest) |
| opm | v1.26.2 | latest v1.x |
| kustomize | v4.5.7 | v5.x (note: v5 has breaking changes in Makefile targets) |
| controller-gen | v0.14.0 | v0.17.x (matches controller-runtime v0.19) |
| yq | v4.30.5 | v4.44.x |

**Note on kustomize v5**: The `kustomize/v4` → `kustomize/v5` bump changes the
`go install` path. Also verify kustomize config files are compatible with v5.

### 5c. Clean up go.mod
- Remove all remaining stale `replace` directives
- Remove `cloud.google.com/go` replace (was workaround for ambiguous import)
- Remove duplicate or unnecessary `replace` blocks
- Verify `go.sum` is consistent after `go mod tidy`

### 5d. Run full lint and test
```
make fmt vet lint
make unit-test
```
Fix any remaining issues.

---

## Testing Strategy

After each phase:
1. `go build ./...` — must pass
2. `go vet ./...` — must pass
3. `go test ./... -run TestUnit` (unit tests without cluster) — must pass

After Phase 4+5:
4. `make unit-test` (requires cluster) — must pass
5. Review generated CRDs: `make manifests` — check for drift
6. `make bundle` — verify bundle generation still works

---

## Risk Register

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| IBM/controller-filtered-cache incompatible with new controller-runtime | High | High | Replace with native cache.Options (Phase 3) |
| redhat-marketplace-operator has no newer tagged version | Medium | Medium | Use only CRD types, pin to working commit or use go:generate to copy type |
| IBM/operand-deployment-lifecycle-manager v1.23.5 pulls incompatible k8s version | Medium | Medium | Check go.mod of ODLM v1.23.5 before upgrading |
| operator-framework/api v0.41.0 OperatorGroup API changes | Low | Medium | Review changelog before upgrading |
| kustomize v4→v5 breaks bundle generation | Medium | Low | Test bundle generation after upgrade |
| prometheus-operator v0.89.0 monitoring type changes | Low | Low | Mostly additive; RelabelConfig may differ |

---

## File Change Summary

| File | Changes |
|------|---------|
| `go.mod` | Remove replace directives, update all versions |
| `go.sum` | Regenerated by go mod tidy |
| `main.go` | Fix ctrl.Options, remove filtered-cache import, fix prometheus path |
| `controllers/ibmlicensing_controller.go` | Fix prometheus import path |
| `controllers/resources/helper.go` | Fix prometheus import path |
| `controllers/resources/service/service_monitor.go` | Fix prometheus import path |
| `Makefile` | Update tool versions |

Additional files may need changes if ODLM, operator-framework, or openshift API
types have breaking changes discovered during compilation.
