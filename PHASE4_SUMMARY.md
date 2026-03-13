# Phase 4 Summary — Upgrade remaining direct dependencies

Completed: 2026-03-09

## Objective

Upgrade all remaining direct dependencies that were not addressed in Phase 2+3,
fix any resulting API changes, and clean up deprecated/moved indirect packages.

---

## Cascading version jumps (beyond plan expectations)

The plan targeted conservative version bumps for each dependency in isolation.
In practice, the dependency graph forced larger jumps:

| Package | Plan target (after Phase 2) | Actual final version |
|---------|----------------------------|----------------------|
| `k8s.io/api` | v0.31.4 (already done) | **v0.35.1** |
| `k8s.io/apimachinery` | v0.31.4 | **v0.35.1** |
| `k8s.io/client-go` | v0.31.4 | **v0.35.1** |
| `k8s.io/apiextensions-apiserver` | v0.31.4 | **v0.35.1** |
| `sigs.k8s.io/controller-runtime` | v0.19.4 | **v0.23.1** |
| `sigs.k8s.io/structured-merge-diff` | v4.x | **v6.x** |

**Why the jumps occurred:**

- `prometheus-operator/pkg/apis/monitoring v0.89.0` requires `k8s.io >= v0.34` and
  `controller-runtime >= v0.22`. This triggered the first jump (k8s to v0.34.3,
  controller-runtime to v0.22.3).
- `github.com/openshift/api v0.0.0-20260306` requires `k8s.io >= v0.35`, pushing
  k8s to v0.35.1.
- `github.com/operator-framework/api v0.41.0` requires `controller-runtime >= v0.23`,
  pushing it to v0.23.1.

Despite jumping further than planned, `go build ./...` and `go vet ./...` remained
clean throughout because controller-runtime v0.22–v0.23 did not introduce
additional breaking API changes beyond what was already fixed in Phase 3.

---

## Steps executed

### 4a. prometheus-operator v0.57.0 → v0.89.0

```
go get github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring@v0.89.0
```

Transitive upgrades: `prometheus/client_golang` → v1.22.0, `prometheus/common` → v0.62.0,
`prometheus/procfs` updated; `matttproud/golang_protobuf_extensions` dropped.

Required code changes — see [Code changes](#code-changes) below.

### 4b. IBM ODLM v1.21.0 → v1.23.5

```
go get github.com/IBM/operand-deployment-lifecycle-manager@v1.23.5
```

No code changes required. API surface used by the operator
(`controllers/resources/operandrequests.go`, `controllers/resources/namespacescopes.go`)
remained compatible.

### 4c. OpenShift API → v0.0.0-20260306105915

```
go get github.com/openshift/api@v0.0.0-20260306105915-ec7ab20aa8c4
```

No code changes required. `Route` and `ServiceCA` types used in the operator
remained compatible.

### 4d. operator-framework/api v0.17.7 → v0.41.0

```
go get github.com/operator-framework/api@v0.41.0
```

No code changes required. `OperatorGroup` types used in
`controllers/operatorgroup_cleaner.go` remained compatible.

### 4e. go.uber.org/zap v1.26.0 → v1.27.1

```
go get go.uber.org/zap@v1.27.1
```

No code changes required. The `zap.New(func(o *zap.Options){...})` functional
options API and `zapcore.RFC3339TimeEncoder` are unchanged.
`go.uber.org/multierr` stayed at v1.11.0 (already at target).

### 4f. redhat-marketplace-operator

Already handled in Phase 2+3 — types vendored locally into `pkg/rhmp/`,
dependency removed from `go.mod`.

### 4g. Remaining direct dependencies

```
go get emperror.dev/errors@v0.8.1
go get github.com/go-logr/logr@v1.4.3
go get github.com/onsi/ginkgo/v2@v2.28.1
go get github.com/onsi/gomega@v1.39.1
go get github.com/stretchr/testify@v1.11.1
```

No code changes required.

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

All pinned to latest to pick up CVE fixes (especially in `x/net` and `x/crypto`).

### 4i. Deprecated/moved packages — verified gone

All of the following were absent from `go.mod` after `go mod tidy`:

| Package | Status |
|---------|--------|
| `go.uber.org/atomic` | Gone — merged into `zap` / `multierr` |
| `k8s.io/klog` v1 | Gone — superseded by `k8s.io/klog/v2` |
| `github.com/golang/protobuf` | Gone — superseded by `google.golang.org/protobuf` |
| `github.com/matttproud/golang_protobuf_extensions` | Gone — merged into prometheus client |
| `github.com/google/gnostic` | Gone — replaced by `github.com/google/gnostic-models` |
| `github.com/imdario/mergo` | Gone — replaced by `dario.cat/mergo` (via controller-runtime) |

---

## Code changes

### controllers/resources/service/service_monitor.go

This was the only file requiring code changes. The `prometheus-operator/monitoring/v1`
API changed significantly between v0.57.0 and v0.89.0.

#### 1. `BearerTokenSecret` removed from `Endpoint` struct literal

**Before:**
```go
Endpoints: []monitoringv1.Endpoint{
    {
        BearerTokenSecret: corev1.SecretKeySelector{Key: ""},
        ...
    },
},
```

**After:** field removed entirely.

**Why:** In v0.89.0, `BearerTokenSecret` moved from a direct field on `Endpoint`
into an embedded struct chain: `Endpoint` → `HTTPConfigWithProxyAndTLSFiles` →
`HTTPConfigWithoutTLS.BearerTokenSecret`. Go does not allow setting promoted
(embedded) fields by name in a struct literal — you must name the embedding struct
explicitly. Since the old value was an empty `SecretKeySelector{Key: ""}` (a no-op
that disables bearer token auth), the correct fix is to simply remove it; `nil`
is the default and has the same meaning.

#### 2. `TLSConfig` moved into embedded struct

**Before:**
```go
Endpoint{
    TLSConfig: tlsConfig,
}
```

**After:**
```go
Endpoint{
    HTTPConfigWithProxyAndTLSFiles: monitoringv1.HTTPConfigWithProxyAndTLSFiles{
        HTTPConfigWithTLSFiles: monitoringv1.HTTPConfigWithTLSFiles{
            TLSConfig: tlsConfig,
        },
    },
}
```

**Why:** Same restructuring — `TLSConfig` moved from a direct `Endpoint` field
into the embedded `HTTPConfigWithTLSFiles`. The type is still `*monitoringv1.TLSConfig`
(which still embeds `SafeTLSConfig`), so `getTLSConfigForServiceMonitor` required
no changes.

#### 3. `[]*RelabelConfig` → `[]RelabelConfig` (value slice)

**Before:**
```go
func getMetricRelabelConfigs(...) []*monitoringv1.RelabelConfig {
    relabelConfigs = append(relabelConfigs, &monitoringv1.RelabelConfig{...})
    return relabelConfigs
}
```

**After:**
```go
func getMetricRelabelConfigs(...) []monitoringv1.RelabelConfig {
    relabelConfigs = append(relabelConfigs, monitoringv1.RelabelConfig{...})
    return relabelConfigs
}
```

Applied to `getMetricRelabelConfigs`, `getMetricRelabelConfigsForRHMP`,
`getMetricRelabelConfigsForAlerting`, `getRelabelConfigs`, and the `GetServiceMonitor`
parameter signature.

#### 4. `Scheme` field: `string` → `*monitoringv1.Scheme`

**Before:**
```go
func getScheme(instance *operatorv1alpha1.IBMLicensing) string {
    if instance.Spec.HTTPSEnable {
        return "https"
    }
    return "http"
}
```

**After:**
```go
func getScheme(instance *operatorv1alpha1.IBMLicensing) *monitoringv1.Scheme {
    var s monitoringv1.Scheme
    if instance.Spec.HTTPSEnable {
        s = monitoringv1.SchemeHTTPS
    } else {
        s = monitoringv1.SchemeHTTP
    }
    return &s
}
```

**Why:** `Scheme` is now a distinct type (`type Scheme string`) with enum validation.
The valid values are `monitoringv1.SchemeHTTP` (`"HTTP"`) and
`monitoringv1.SchemeHTTPS` (`"HTTPS"`) — note uppercase, unlike the old raw strings
`"http"` / `"https"`. The field also changed from `string` to `*Scheme`.

#### 5. `RelabelConfig.Replacement`: `string` → `*string`

**Before:**
```go
monitoringv1.RelabelConfig{
    Replacement: fmt.Sprintf("%s:%d", getServerName(instance), prometheusTargetPort.IntVal),
    TargetLabel: "__address__",
}
```

**After:**
```go
replacement := fmt.Sprintf("%s:%d", getServerName(instance), prometheusTargetPort.IntVal)
monitoringv1.RelabelConfig{
    Replacement: &replacement,
    TargetLabel: "__address__",
}
```

---

## Final go.mod direct dependencies

| Module | Before Phase 4 | After Phase 4 |
|--------|---------------|---------------|
| `emperror.dev/errors` | v0.8.0 | v0.8.1 |
| `github.com/IBM/operand-deployment-lifecycle-manager` | v1.21.0 | **v1.23.5** |
| `github.com/go-logr/logr` | v1.4.2 | v1.4.3 |
| `github.com/onsi/ginkgo/v2` | v2.19.0 | **v2.28.1** |
| `github.com/onsi/gomega` | v1.33.1 | **v1.39.1** |
| `github.com/openshift/api` | v0.0.0-20230306 | **v0.0.0-20260306** |
| `github.com/operator-framework/api` | v0.17.7 | **v0.41.0** |
| `github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring` | v0.57.0 | **v0.89.0** |
| `github.com/stretchr/testify` | v1.9.0 | **v1.11.1** |
| `go.uber.org/zap` | v1.26.0 | **v1.27.1** |
| `k8s.io/api` | v0.31.4 | **v0.35.1** |
| `k8s.io/apimachinery` | v0.31.4 | **v0.35.1** |
| `k8s.io/client-go` | v0.31.4 | **v0.35.1** |
| `k8s.io/utils` | v0.0.0-20240711 | v0.0.0-20260108 |
| `sigs.k8s.io/controller-runtime` | v0.19.4 | **v0.23.1** |

---

## Verification

All checks passed after Phase 4:

- `go mod tidy` — clean (no stale dependencies)
- `go build ./...` — clean
- `go vet ./...` — clean
- No deprecated packages in `go.mod`
- No replace directives in `go.mod`

## Files changed

| File | Change |
|------|--------|
| `go.mod` | All direct dependencies upgraded; several new transitive packages added; deprecated ones removed |
| `go.sum` | Regenerated by `go mod tidy` |
| `controllers/resources/service/service_monitor.go` | Migrated to prometheus-operator v0.89.0 API (removed `BearerTokenSecret`, moved `TLSConfig`, changed `[]*RelabelConfig` to `[]RelabelConfig`, typed `Scheme`, `*string` Replacement) |
