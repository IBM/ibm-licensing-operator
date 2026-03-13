# Dependency Version Audit — ibm-licensing-operator

**Original audit:** 2026-03-06 (pre-upgrade baseline)
**Updated to reflect final state after all upgrade phases:** 2026-03-09
**Build tool versions updated + linter script refactor:** 2026-03-13

## Direct Dependencies

| Module | go.mod version (final) | Notes |
|--------|------------------------|-------|
| `emperror.dev/errors` | v0.8.1 | At latest |
| `github.com/IBM/controller-filtered-cache` | **REMOVED** | Replaced by native `cache.Options` in controller-runtime ≥ v0.15 (Phase 3) |
| `github.com/IBM/operand-deployment-lifecycle-manager` | v1.23.5 | At latest |
| `github.com/coreos/prometheus-operator` | **REMOVED** | Archived import path; replaced by `github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring` (Phase 1) |
| `github.com/go-logr/logr` | v1.4.3 | At latest |
| `github.com/onsi/ginkgo/v2` | v2.28.1 | At latest |
| `github.com/onsi/gomega` | v1.39.1 | At latest |
| `github.com/openshift/api` | v0.0.0-20260306105915-ec7ab20aa8c4 | At latest (pseudo-version) |
| `github.com/operator-framework/api` | v0.41.0 | At latest |
| `github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring` | v0.89.0 | At latest; promoted to direct dependency (Phase 1) |
| `github.com/redhat-marketplace/redhat-marketplace-operator/v2` | **REMOVED** | Types vendored locally into `pkg/rhmp/` (Phase 2+3) |
| `github.com/stretchr/testify` | v1.11.1 | At latest |
| `go.uber.org/zap` | v1.27.1 | At latest |
| `k8s.io/api` | v0.35.1 | At latest stable (v0.35.x) |
| `k8s.io/apimachinery` | v0.35.1 | At latest stable (v0.35.x) |
| `k8s.io/client-go` | v0.35.1 | At latest stable (v0.35.x); legacy `+incompatible` version removed |
| `k8s.io/utils` | v0.0.0-20260108192941-914a6e750570 | At latest |
| `sigs.k8s.io/controller-runtime` | v0.23.1 | At latest stable |

All `replace` directives have been removed. Effective version now equals required version for all modules.

## Indirect Dependencies

| Module | go.mod version (final) | Notes |
|--------|------------------------|-------|
| `github.com/Masterminds/semver/v3` | v3.4.0 | At latest |
| `github.com/beorn7/perks` | v1.0.1 | At latest |
| `github.com/blang/semver/v4` | v4.0.0 | At latest |
| `github.com/cespare/xxhash/v2` | v2.3.0 | At latest |
| `github.com/davecgh/go-spew` | v1.1.2-0.20180830191138-d8f796af33cc | At latest |
| `github.com/deckarep/golang-set` | v1.7.1 | Minor patch behind (v1.8.0) |
| `github.com/emicklei/go-restful/v3` | v3.12.2 | Minor patch behind (v3.13.0) |
| `github.com/evanphx/json-patch` | v4.12.0+incompatible | Kept for transitive compat |
| `github.com/evanphx/json-patch/v5` | v5.9.11 | At latest |
| `github.com/fsnotify/fsnotify` | v1.9.0 | At latest |
| `github.com/fxamacker/cbor/v2` | v2.9.0 | New transitive dep (controller-runtime v0.23) |
| `github.com/go-logr/zapr` | v1.3.0 | At latest |
| `github.com/go-openapi/jsonpointer` | v0.22.4 | At latest |
| `github.com/go-openapi/jsonreference` | v0.21.4 | At latest |
| `github.com/go-openapi/swag` | v0.25.4 | At latest |
| `github.com/go-task/slim-sprig/v3` | v3.0.0 | At latest |
| `github.com/golang/protobuf` | **GONE** | Deprecated; superseded by `google.golang.org/protobuf` |
| `github.com/google/btree` | v1.1.3 | New transitive dep |
| `github.com/google/gnostic` | **GONE** | Moved to `github.com/google/gnostic-models` |
| `github.com/google/gnostic-models` | v0.7.1 | Replacement for gnostic |
| `github.com/google/go-cmp` | v0.7.0 | At latest |
| `github.com/google/pprof` | v0.0.0-20260115054156-294ebfa9ad83 | At latest |
| `github.com/google/uuid` | v1.6.0 | At latest |
| `github.com/imdario/mergo` | **GONE** | Moved to `dario.cat/mergo` (pulled transitively via controller-runtime) |
| `github.com/json-iterator/go` | v1.1.12 | At latest |
| `github.com/matttproud/golang_protobuf_extensions` | **GONE** | Deprecated; merged into prometheus client packages |
| `github.com/modern-go/concurrent` | v0.0.0-20180306012644-bacd9c7ef1dd | At latest |
| `github.com/modern-go/reflect2` | v1.0.3-0.20250322232337-35a7c28c31ee | At latest |
| `github.com/munnerz/goautoneg` | v0.0.0-20191010083416-a7dc8b61c822 | At latest |
| `github.com/pkg/errors` | v0.9.1 | Soft-deprecated (prefer stdlib errors); kept as transitive dep |
| `github.com/prometheus/client_golang` | v1.23.2 | At latest |
| `github.com/prometheus/client_model` | v0.6.2 | At latest |
| `github.com/prometheus/common` | v0.67.5 | At latest |
| `github.com/prometheus/procfs` | v0.19.2 | At latest |
| `github.com/sirupsen/logrus` | v1.9.4 | At latest |
| `github.com/spf13/pflag` | v1.0.10 | At latest |
| `github.com/x448/float16` | v0.8.4 | New transitive dep (cbor/v2) |
| `go.uber.org/atomic` | **GONE** | Deprecated; merged into `go.uber.org/multierr` and `zap` |
| `go.uber.org/multierr` | v1.11.0 | At latest |
| `go.yaml.in/yaml/v2` | v2.4.3 | New transitive dep (replaces gopkg.in/yaml.v2 transitively) |
| `go.yaml.in/yaml/v3` | v3.0.4 | New transitive dep |
| `golang.org/x/mod` | v0.33.0 | At latest |
| `golang.org/x/net` | v0.51.0 | At latest |
| `golang.org/x/oauth2` | v0.36.0 | At latest |
| `golang.org/x/sync` | v0.19.0 | At latest |
| `golang.org/x/sys` | v0.42.0 | At latest |
| `golang.org/x/term` | v0.40.0 | At latest |
| `golang.org/x/text` | v0.34.0 | At latest |
| `golang.org/x/time` | v0.15.0 | At latest |
| `golang.org/x/tools` | v0.42.0 | At latest |
| `gomodules.xyz/jsonpatch/v2` | v2.4.0 | At latest |
| `google.golang.org/protobuf` | v1.36.11 | At latest |
| `gopkg.in/evanphx/json-patch.v4` | v4.13.0 | New transitive dep |
| `gopkg.in/inf.v0` | v0.9.1 | At latest |
| `gopkg.in/yaml.v3` | v3.0.1 | At latest |
| `k8s.io/apiextensions-apiserver` | v0.35.1 | At latest stable (v0.35.x) |
| `k8s.io/component-base` | **GONE** | No longer required transitively |
| `k8s.io/klog` (v1) | **GONE** | Deprecated; superseded by `k8s.io/klog/v2` |
| `k8s.io/klog/v2` | v2.130.1 | At latest |
| `k8s.io/kube-openapi` | v0.0.0-20260127142750-a19766b6e2d4 | At latest |
| `sigs.k8s.io/json` | v0.0.0-20250730193827-2d320260d730 | At latest |
| `sigs.k8s.io/randfill` | v1.0.0 | New transitive dep |
| `sigs.k8s.io/structured-merge-diff/v4` | **GONE** | Replaced by v6 |
| `sigs.k8s.io/structured-merge-diff/v6` | v6.3.2-0.20260122202528-d9cc6641c482 | At latest |
| `sigs.k8s.io/yaml` | v1.6.0 | At latest |

## Build Tools (Makefile)

All tool versions are now defined exclusively in the Makefile. Previously `golangci-lint`, `goimports`, `shellcheck`, and `yamllint` had versions hardcoded in `install-linters-development.sh` (now removed). Each tool now has a dedicated install script under `common/scripts/`.

| Tool | Before ILS-1821 | After ILS-1821 | Current (2026-03-13) |
|------|----------------|----------------|----------------------|
| `operator-sdk` | v1.32.0 | v1.42.0 | **v1.42.1** |
| `opm` | v1.26.2 | v1.64.0 | v1.64.0 |
| `kustomize` | v4.5.7 | v5.8.1 | v5.8.1 |
| `controller-gen` | v0.14.0 | v0.20.1 | v0.20.1 |
| `yq` | v4.30.5 | v4.52.4 | v4.52.4 |
| `golangci-lint` | v2.11.2 (in script) | v2.11.2 (in script) | v2.11.2 (v2.11.3 hangs — regression) |
| `goimports` | v0.3.0 (in script) | v0.3.0 (in script) | **v0.43.0** |
| `shellcheck` | v0.8.0 (in script) | v0.8.0 (in script) | **v0.11.0** |
| `yamllint` | 1.28.0 (in script) | 1.28.0 (in script) | **1.37.1** |

## Summary

| Category | Count |
|----------|-------|
| At latest / current | ~45 |
| Minor patch behind | ~5 |
| Removed (deprecated/moved/replaced) | 10 |
| New transitive deps added | 7 |

### Deprecated packages — all resolved

| Package | Resolution |
|---------|------------|
| `github.com/coreos/prometheus-operator` | Removed; import path updated to `prometheus-operator/prometheus-operator` |
| `github.com/golang/protobuf` | Gone; superseded by `google.golang.org/protobuf` |
| `github.com/matttproud/golang_protobuf_extensions` | Gone; merged into prometheus client packages |
| `go.uber.org/atomic` | Gone; merged into `go.uber.org/multierr` and `zap` |
| `k8s.io/klog` (v1) | Gone; superseded by `k8s.io/klog/v2` |

### Packages with moved import paths — all resolved

| Old path | New path |
|----------|----------|
| `github.com/coreos/prometheus-operator` | `github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring` |
| `github.com/google/gnostic` | `github.com/google/gnostic-models` |
| `github.com/imdario/mergo` | `dario.cat/mergo` (pulled transitively) |

### replace directives — all removed

All `replace` directives that previously pinned k8s.io/\*, sigs.k8s.io/controller-runtime,
and github.com/IBM/controller-filtered-cache to old versions have been removed.
