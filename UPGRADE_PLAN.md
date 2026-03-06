# Dependency & Go Upgrade Plan — ibm-licensing-operator

## Current State (as of 2026-03-06)

### Go toolchain
| Item | Current | Latest |
|------|---------|--------|
| go directive in go.mod | 1.25.6 | 1.26.1 |
| Installed Go | 1.26.1 | — |

### Direct dependencies
| Module | Required | Effective (replace) | Latest |
|--------|----------|---------------------|--------|
| emperror.dev/errors | v0.8.0 | v0.8.0 | v0.8.1 |
| github.com/IBM/controller-filtered-cache | v0.3.5 | v0.3.5 | v0.3.6 |
| github.com/IBM/operand-deployment-lifecycle-manager | v1.21.0 | v1.21.0 | v1.23.5 |
| github.com/coreos/prometheus-operator | v0.41.0 | v0.41.0 | **MOVED** (see P2) |
| github.com/go-logr/logr | v1.2.4 | v1.2.4 | v1.4.3 |
| github.com/onsi/ginkgo/v2 | v2.9.2 | v2.9.2 | v2.28.1 |
| github.com/onsi/gomega | v1.27.4 | v1.27.4 | v1.39.1 |
| github.com/openshift/api | v0.0.0-20230306 | v0.0.0-20230306 | v0.0.0-20260306 |
| github.com/operator-framework/api | v0.17.7 | v0.17.7 | v0.41.0 |
| github.com/redhat-marketplace/redhat-marketplace-operator/v2 | v2.0.0-20230228 | v2.0.0-20230228 | v2.0.0-20260302 |
| github.com/stretchr/testify | v1.8.4 | v1.8.4 | v1.11.1 |
| go.uber.org/zap | v1.21.0 | v1.21.0 | v1.27.1 |
| k8s.io/api | v0.27.2 | **v0.25.7** (replace!) | v0.35.2 |
| k8s.io/apimachinery | v0.27.2 | **v0.25.7** (replace!) | v0.35.2 |
| k8s.io/client-go | v12.0.0+incompatible | **v0.25.7** (replace!) | v0.35.2 |
| k8s.io/utils | v0.0.0-20240502 | v0.0.0-20240502 | v0.0.0-20260210 |
| sigs.k8s.io/controller-runtime | v0.15.0 | **v0.12.3** (replace!) | v0.23.3 |

### Key indirect dependencies
| Module | go.mod version | Latest | Notes |
|--------|---------------|--------|-------|
| github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring | v0.57.0 | v0.89.0 | Will update with Phase 4a |
| github.com/prometheus/client_golang | v1.15.1 | v1.23.2 | Will update transitively |
| github.com/prometheus/common | v0.42.0 | v0.67.5 | Will update transitively |
| github.com/prometheus/procfs | v0.9.0 | v0.20.1 | Will update transitively |
| github.com/google/gnostic | v0.5.7-v3refs | — | **MOVED** → github.com/google/gnostic-models v0.7.1 |
| github.com/imdario/mergo | v0.3.12 | — | **MOVED** → dario.cat/mergo v1.0.2 |
| github.com/golang/protobuf | v1.5.3 | v1.5.4 | **DEPRECATED** → google.golang.org/protobuf |
| github.com/matttproud/golang_protobuf_extensions | v1.0.4 | v1.0.4 | **DEPRECATED** → merged into prometheus packages |
| go.uber.org/atomic | v1.9.0 | v1.11.0 | **DEPRECATED** → merged into zap/multierr |
| go.uber.org/multierr | v1.7.0 | v1.11.0 | Will update with zap upgrade |
| k8s.io/apiextensions-apiserver | v0.27.2 | v0.35.2 | Will update with k8s upgrade |
| k8s.io/component-base | v0.27.2 | v0.35.2 | Will update with k8s upgrade |
| k8s.io/klog (v1) | v1.0.0 | v1.0.0 | **DEPRECATED** → k8s.io/klog/v2 |
| k8s.io/klog/v2 | v2.90.1 | v2.130.1 | Will update with k8s upgrade |
| k8s.io/kube-openapi | v0.0.0-20230501 | v0.0.0-20260304 | Will update with k8s upgrade |
| golang.org/x/crypto | v0.45.0 | v0.48.0 | Will update transitively |
| golang.org/x/net | v0.47.0 | v0.51.0 | Will update transitively |
| golang.org/x/oauth2 | v0.27.0 | v0.35.0 | Will update transitively |
| golang.org/x/sys | v0.38.0 | v0.41.0 | Will update transitively |
| golang.org/x/term | v0.37.0 | v0.40.0 | Will update transitively |
| golang.org/x/text | v0.31.0 | v0.34.0 | Will update transitively |
| golang.org/x/time | v0.3.0 | v0.14.0 | Will update transitively |
| golang.org/x/tools | v0.38.0 | v0.42.0 | Will update transitively |
| gomodules.xyz/jsonpatch/v2 | v2.2.0 | v2.5.0 | Will update transitively |
| google.golang.org/protobuf | v1.33.0 | v1.36.11 | Will update transitively |
| sigs.k8s.io/json | v0.0.0-20221116 | v0.0.0-20250730 | Will update with k8s upgrade |
| sigs.k8s.io/structured-merge-diff/v4 | v4.2.3 | v4.7.0 | Will update with k8s upgrade |
| sigs.k8s.io/yaml | v1.3.0 | v1.6.0 | Will update transitively |
| github.com/gobuffalo/flect | v0.2.1 | v1.0.3 | Will update with controller-gen upgrade |
| github.com/spf13/cast | v1.4.1 | v1.10.0 | Will update transitively |
| github.com/spf13/pflag | v1.0.5 | v1.0.10 | Will update transitively |

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

### P6 — Moved indirect packages
Two indirect packages have moved to new import paths and must be tracked:
- `github.com/google/gnostic` → `github.com/google/gnostic-models` v0.7.1
- `github.com/imdario/mergo` → `dario.cat/mergo` v1.0.2

These are pulled transitively by k8s.io/* and will be resolved automatically
after the k8s upgrade, but the old paths must not appear in `go.mod` after tidy.

### P7 — Deprecated indirect packages
Four indirect packages are deprecated and should disappear after `go mod tidy`
once their dependents are upgraded:
- `go.uber.org/atomic` — merged into `go.uber.org/zap`/`multierr` (drops after Phase 4e)
- `k8s.io/klog` v1 — superseded by `k8s.io/klog/v2` (drops after Phase 2)
- `github.com/golang/protobuf` — superseded by `google.golang.org/protobuf` (drops after Phase 2)
- `github.com/matttproud/golang_protobuf_extensions` — merged into prometheus packages (drops after Phase 4a)

### P8 — Security and CVE exposure
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
Transitive updates expected: `prometheus/client_golang` → v1.23.2, `prometheus/common` → v0.67.5,
`prometheus/procfs` → v0.20.1, `prometheus/client_model` → v0.6.2,
`matttproud/golang_protobuf_extensions` should drop.

### 4b. IBM ODLM
```
go get github.com/IBM/operand-deployment-lifecycle-manager@v1.23.5
```
Verify API compatibility in `controllers/resources/operandrequests.go` and
`controllers/resources/namespacescopes.go`.

### 4c. OpenShift API
```
go get github.com/openshift/api@v0.0.0-20260306105915-ec7ab20aa8c4
```
Verify Route and ServiceCA types still compatible.

### 4d. operator-framework/api
```
go get github.com/operator-framework/api@v0.41.0
```
Verify `OperatorGroup` types still compatible in `controllers/operatorgroup_cleaner.go`.

### 4e. go.uber.org/zap and multierr
```
go get go.uber.org/zap@v1.27.1
go get go.uber.org/multierr@v1.11.0
```
The `zap.New(func(o *zap.Options){...})` functional options API is unchanged.
Verify `zapcore.RFC3339TimeEncoder` still available.
`go.uber.org/atomic` should drop from go.mod after tidy (deprecated, merged into these).

### 4f. redhat-marketplace-operator
Latest pseudo-version is `v2.0.0-20260302222345-0e3562b432b3`. Evaluate whether:
- The newer pseudo-version is compatible
- The dependency can be replaced by using only the CRD schema directly
  (the operator only uses `meterdefv1beta1.MeterDefinitionList` for scheme registration)

### 4g. Remaining direct dependencies
```
go get emperror.dev/errors@v0.8.1
go get github.com/go-logr/logr@v1.4.3
go get github.com/onsi/ginkgo/v2@v2.28.1
go get github.com/onsi/gomega@v1.39.1
go get github.com/stretchr/testify@v1.11.1
```

### 4h. golang.org/x/* packages
```
go get golang.org/x/crypto@latest
go get golang.org/x/net@latest
go get golang.org/x/oauth2@latest
go get golang.org/x/sys@latest
go get golang.org/x/term@latest
go get golang.org/x/text@latest
go get golang.org/x/time@latest
go get golang.org/x/tools@latest
```

### 4i. Verify deprecated/moved indirect packages are gone
After `go mod tidy`, confirm these no longer appear in `go.mod`:
- `go.uber.org/atomic` (deprecated)
- `k8s.io/klog` v1 (deprecated)
- `github.com/golang/protobuf` (deprecated)
- `github.com/matttproud/golang_protobuf_extensions` (deprecated)
- `github.com/google/gnostic` (moved to `gnostic-models`)
- `github.com/imdario/mergo` (moved to `dario.cat/mergo`)

If any remain, they are still required by a dependency that has not yet been
upgraded — investigate and resolve before proceeding.

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
| Moved packages (gnostic, mergo) still referenced after tidy | Low | Low | Explicitly verify absence in go.mod after Phase 2 tidy |
| Deprecated indirect packages not dropped by tidy | Low | Low | If any remain, identify which dependency still requires them and upgrade it |

---

## File Change Summary

| File | Changes |
|------|---------|
| `go.mod` | Remove replace directives, update all versions, remove deprecated entries |
| `go.sum` | Regenerated by go mod tidy |
| `main.go` | Fix ctrl.Options, remove filtered-cache import, fix prometheus path |
| `controllers/ibmlicensing_controller.go` | Fix prometheus import path |
| `controllers/resources/helper.go` | Fix prometheus import path |
| `controllers/resources/service/service_monitor.go` | Fix prometheus import path |
| `Makefile` | Update tool versions |

Additional files may need changes if ODLM, operator-framework, openshift API,
or prometheus-operator v0.89.0 types have breaking changes discovered during compilation.
