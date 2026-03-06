# Dependency Version Audit — ibm-licensing-operator

Checked on: 2026-03-06

## Direct Dependencies

| Module | go.mod version | Effective version* | Latest stable | Status |
|--------|---------------|-------------------|--------------|--------|
| `emperror.dev/errors` | v0.8.0 | v0.8.0 | v0.8.1 | Minor patch behind |
| `github.com/IBM/controller-filtered-cache` | v0.3.5 | v0.3.5 | v0.3.6 | Patch behind; last release May 2023, stale |
| `github.com/IBM/operand-deployment-lifecycle-manager` | v1.21.0 | v1.21.0 | v1.23.5 | Outdated |
| `github.com/coreos/prometheus-operator` | v0.41.0 | v0.41.0 | v0.89.0 | **MOVED** — import path archived; use `github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring` |
| `github.com/go-logr/logr` | v1.2.4 | v1.2.4 | v1.4.3 | Outdated |
| `github.com/onsi/ginkgo/v2` | v2.9.2 | v2.9.2 | v2.28.1 | Significantly outdated |
| `github.com/onsi/gomega` | v1.27.4 | v1.27.4 | v1.39.1 | Significantly outdated |
| `github.com/openshift/api` | v0.0.0-20230306181726-ab59d80e2b79 | v0.0.0-20230306 | v0.0.0-20260306 | ~3 years behind |
| `github.com/operator-framework/api` | v0.17.7 | v0.17.7 | v0.41.0 | Severely outdated |
| `github.com/redhat-marketplace/redhat-marketplace-operator/v2` | v2.0.0-20230228135942-40c6ba166b59 | v2.0.0-20230228 | v2.0.0-20260302 | ~3 years behind |
| `github.com/stretchr/testify` | v1.8.4 | v1.8.4 | v1.11.1 | Outdated |
| `github.com/pmezard/go-difflib` | v1.0.0 | v1.0.0 | v1.0.0 | Current |
| `go.uber.org/zap` | v1.21.0 | v1.21.0 | v1.27.1 | Outdated |
| `k8s.io/api` | v0.27.2 | **v0.25.7** (replaced!) | v0.35.2 | **Severely outdated** — replace directive pins to Kubernetes 1.25 API |
| `k8s.io/apimachinery` | v0.27.2 | **v0.25.7** (replaced!) | v0.35.2 | **Severely outdated** |
| `k8s.io/client-go` | v12.0.0+incompatible | **v0.25.7** (replaced!) | v0.35.2 | **Severely outdated** + legacy `+incompatible` version |
| `k8s.io/utils` | v0.0.0-20240502163921-fe8a2dddb1d0 | v0.0.0-20240502 | v0.0.0-20260210 | Outdated |
| `sigs.k8s.io/controller-runtime` | v0.15.0 | **v0.12.3** (replaced!) | v0.23.3 | **Severely outdated** — pinned to a ~3-year-old release |

\* After applying `replace` directives in go.mod.

## Indirect Dependencies

| Module | go.mod version | Latest stable | Status |
|--------|---------------|--------------|--------|
| `cloud.google.com/go/compute/metadata` | v0.3.0 | v0.9.0 | Outdated |
| `github.com/Masterminds/goutils` | v1.1.1 | v1.1.1 | Current |
| `github.com/Masterminds/semver/v3` | v3.1.1 | v3.4.0 | Outdated |
| `github.com/Masterminds/sprig/v3` | v3.2.2 | v3.3.0 | Minor version behind |
| `github.com/beorn7/perks` | v1.0.1 | v1.0.1 | Current |
| `github.com/blang/semver/v4` | v4.0.0 | v4.0.0 | Current |
| `github.com/cespare/xxhash` | v1.1.0 | v1.1.0 | Current (v2 preferred) |
| `github.com/cespare/xxhash/v2` | v2.2.0 | v2.3.0 | Minor patch behind |
| `github.com/davecgh/go-spew` | v1.1.1 | v1.1.1 | Current |
| `github.com/deckarep/golang-set` | v1.7.1 | v1.8.0 | Minor version behind |
| `github.com/emicklei/go-restful/v3` | v3.10.1 | v3.13.0 | Outdated |
| `github.com/evanphx/json-patch` | v4.12.0+incompatible | v5.9.11 (as `/v5`) | Outdated — proper v5 module available |
| `github.com/fsnotify/fsnotify` | v1.6.0 | v1.9.0 | Outdated |
| `github.com/go-logr/zapr` | v1.2.3 | v1.3.0 | Minor version behind |
| `github.com/go-openapi/jsonpointer` | v0.19.6 | v0.22.5 | Outdated |
| `github.com/go-openapi/jsonreference` | v0.20.1 | v0.21.5 | Outdated |
| `github.com/go-openapi/swag` | v0.22.3 | v0.25.5 | Outdated |
| `github.com/go-task/slim-sprig` | v0.0.0-20230315185526-52ccab3ef572 | v3.0.0 (as `/v3`) | Versioned module available |
| `github.com/gobuffalo/flect` | v0.2.1 | v1.0.3 | Significantly outdated |
| `github.com/gogo/protobuf` | v1.3.2 | v1.3.2 | Current |
| `github.com/golang/groupcache` | v0.0.0-20210331224755-41bb18bfe9da | v0.0.0-20241129 | Outdated |
| `github.com/golang/protobuf` | v1.5.3 | v1.5.4 | Patch behind; **DEPRECATED** — superseded by `google.golang.org/protobuf` |
| `github.com/google/gnostic` | v0.5.7-v3refs | — | **MOVED** → `github.com/google/gnostic-models` v0.7.1 |
| `github.com/google/go-cmp` | v0.6.0 | v0.7.0 | Minor version behind |
| `github.com/google/gofuzz` | v1.2.0 | v1.2.0 | Current |
| `github.com/google/pprof` | v0.0.0-20210407192527-94a9f03dee38 | v0.0.0-20260302 | Very outdated |
| `github.com/google/uuid` | v1.3.0 | v1.6.0 | Outdated |
| `github.com/huandu/xstrings` | v1.3.1 | v1.5.0 | Outdated |
| `github.com/imdario/mergo` | v0.3.12 | — | **MOVED** → `dario.cat/mergo` v1.0.2 |
| `github.com/josharian/intern` | v1.0.0 | v1.0.0 | Current |
| `github.com/json-iterator/go` | v1.1.12 | v1.1.12 | Current |
| `github.com/mailru/easyjson` | v0.7.7 | v0.9.1 | Outdated |
| `github.com/matttproud/golang_protobuf_extensions` | v1.0.4 | v1.0.4 | **DEPRECATED** — functionality merged into prometheus client packages |
| `github.com/mitchellh/copystructure` | v1.0.0 | v1.2.0 | Outdated |
| `github.com/mitchellh/reflectwalk` | v1.0.0 | v1.0.2 | Outdated |
| `github.com/modern-go/concurrent` | v0.0.0-20180306012644-bacd9c7ef1dd | v0.0.0-20180306 | Current |
| `github.com/modern-go/reflect2` | v1.0.2 | v1.0.2 | Current |
| `github.com/munnerz/goautoneg` | v0.0.0-20191010083416-a7dc8b61c822 | v0.0.0-20191010 | Current |
| `github.com/pkg/errors` | v0.9.1 | v0.9.1 | Current (soft-deprecated; prefer stdlib `errors`) |
| `github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring` | v0.57.0 | v0.89.0 | Significantly outdated |
| `github.com/prometheus/client_golang` | v1.15.1 | v1.23.2 | Significantly outdated |
| `github.com/prometheus/client_model` | v0.4.0 | v0.6.2 | Outdated |
| `github.com/prometheus/common` | v0.42.0 | v0.67.5 | Significantly outdated |
| `github.com/prometheus/procfs` | v0.9.0 | v0.20.1 | Significantly outdated |
| `github.com/rogpeppe/go-internal` | v1.11.0 | v1.14.1 | Outdated |
| `github.com/shopspring/decimal` | v1.2.0 | v1.4.0 | Outdated |
| `github.com/sirupsen/logrus` | v1.9.3 | v1.9.4 | Patch behind |
| `github.com/spf13/cast` | v1.4.1 | v1.10.0 | Significantly outdated |
| `github.com/spf13/pflag` | v1.0.5 | v1.0.10 | Outdated |
| `go.uber.org/atomic` | v1.9.0 | v1.11.0 | Outdated; **DEPRECATED** — merged into `go.uber.org/multierr` and `zap` |
| `go.uber.org/multierr` | v1.7.0 | v1.11.0 | Outdated |
| `golang.org/x/crypto` | v0.45.0 | v0.48.0 | Patch behind |
| `golang.org/x/net` | v0.47.0 | v0.51.0 | Patch behind |
| `golang.org/x/oauth2` | v0.27.0 | v0.35.0 | Outdated |
| `golang.org/x/sys` | v0.38.0 | v0.41.0 | Patch behind |
| `golang.org/x/term` | v0.37.0 | v0.40.0 | Patch behind |
| `golang.org/x/text` | v0.31.0 | v0.34.0 | Patch behind |
| `golang.org/x/time` | v0.3.0 | v0.14.0 | Outdated |
| `golang.org/x/tools` | v0.38.0 | v0.42.0 | Patch behind |
| `gomodules.xyz/jsonpatch/v2` | v2.2.0 | v2.5.0 | Outdated |
| `google.golang.org/protobuf` | v1.33.0 | v1.36.11 | Outdated |
| `gopkg.in/inf.v0` | v0.9.1 | v0.9.1 | Current |
| `gopkg.in/yaml.v2` | v2.4.0 | v2.4.0 | Current (yaml.v3 preferred) |
| `gopkg.in/yaml.v3` | v3.0.1 | v3.0.1 | Current |
| `k8s.io/apiextensions-apiserver` | v0.27.2 | v0.35.2 | Severely outdated |
| `k8s.io/component-base` | v0.27.2 | v0.35.2 | Severely outdated |
| `k8s.io/klog` | v1.0.0 | v1.0.0 | **DEPRECATED** — use `k8s.io/klog/v2` |
| `k8s.io/klog/v2` | v2.90.1 | v2.130.1 | Outdated |
| `k8s.io/kube-openapi` | v0.0.0-20230501164219-8b0f38b5fd1f | v0.0.0-20260304 | Outdated |
| `sigs.k8s.io/json` | v0.0.0-20221116044647-bc3834ca7abd | v0.0.0-20250730 | Very outdated |
| `sigs.k8s.io/structured-merge-diff/v4` | v4.2.3 | v4.7.0 | Outdated |
| `sigs.k8s.io/yaml` | v1.3.0 | v1.6.0 | Outdated |

## Summary

| Category | Count |
|----------|-------|
| Current / at latest | 10 |
| Minor patch behind | 12 |
| Outdated (minor / several versions) | 24 |
| Severely outdated (major gap, core infra) | 7 |
| Deprecated (superseded by another package) | 5 |
| Moved to new import path | 3 |
| **Total** | **~56** |

### Severely outdated (core infrastructure)
- `k8s.io/api`, `k8s.io/apimachinery`, `k8s.io/client-go` — effectively at Kubernetes 1.25 (v0.25.7) due to `replace` directives; latest is v0.35.2
- `k8s.io/apiextensions-apiserver`, `k8s.io/component-base` — at v0.27.2; latest is v0.35.2
- `sigs.k8s.io/controller-runtime` — effectively at v0.12.3 due to `replace` directive; latest is v0.23.3
- `github.com/operator-framework/api` — at v0.17.7; latest is v0.41.0

### Deprecated packages in use
- `github.com/golang/protobuf` — superseded by `google.golang.org/protobuf`
- `github.com/matttproud/golang_protobuf_extensions` — merged into prometheus client packages
- `go.uber.org/atomic` — merged into `go.uber.org/multierr` and `zap`
- `k8s.io/klog` (v1) — superseded by `k8s.io/klog/v2`
- `github.com/pkg/errors` — stdlib `errors` + `fmt.Errorf` preferred

### Packages with new import paths
- `github.com/coreos/prometheus-operator` → `github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring`
- `github.com/google/gnostic` → `github.com/google/gnostic-models`
- `github.com/imdario/mergo` → `dario.cat/mergo`
