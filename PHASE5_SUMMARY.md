# Phase 5 Summary — Go version, tooling, and cleanup

Completed: 2026-03-09

## Objective

Update the `go` directive in `go.mod` to match the installed toolchain, bump all
Makefile build tools to their latest stable releases, fix kustomize v5 compatibility
in config files, and verify the build and artifact generation pipeline is clean.

---

## Steps executed

### 5a. Update go directive

Updated `go.mod`:

```
go 1.25.7  →  go 1.26.1
```

Go 1.26.1 was already installed. After updating the directive, `go mod tidy` was run;
it correctly omitted a `toolchain` line since the installed toolchain matches the
`go` directive exactly (redundant in that case).

### 5b. Update Makefile tool versions

All versions were verified against the GitHub API (`/releases/latest`) on 2026-03-09:

| Tool | Before | After |
|------|--------|-------|
| `OPM_VERSION` | v1.26.2 | **v1.64.0** |
| `OPERATOR_SDK_VERSION` | v1.32.0 | **v1.42.0** |
| `YQ_VERSION` | v4.30.5 | **v4.52.4** |
| `KUSTOMIZE_VERSION` | v4.5.7 | **v5.8.1** |
| `CONTROLLER_GEN_VERSION` | v0.14.0 | **v0.20.1** |

Also updated the kustomize Go module install path in both `install-kustomize` and
`kustomize` Makefile targets:

```
sigs.k8s.io/kustomize/kustomize/v4@...  →  sigs.k8s.io/kustomize/kustomize/v5@...
```

#### Note on opm

The Makefile `opm` target still builds from source via `git clone` + `go build`.
This approach predates prebuilt binary releases. The flags used (`-mod=vendor`,
`-tags "json1"`) may no longer apply to v1.64.0. If `make opm` is needed,
consider switching to downloading the prebuilt binary from the operator-registry
releases page instead.

### 5c. go.mod cleanup

No action needed — all `replace` directives were already removed in phases 1–4.
`go mod tidy` confirmed the module graph is clean.

### 5d. kustomize v5 compatibility fix

`config/default/kustomization.yaml` used the `bases:` field which was removed in
kustomize v5 (deprecated in v4, removed in v5). Fixed by renaming to `resources:`:

```yaml
# Before
bases:
- ../crd
- ../rbac
- ../manager

# After
resources:
- ../crd
- ../rbac
- ../manager
```

All other kustomization files already used `resources:` and required no changes.

### 5e. Fix PROJECT file for operator-sdk v1.42.0

The `PROJECT` file declared `layout: go.kubebuilder.io/v2` and
`plugins: go.operator-sdk.io/v2-alpha`, both unsupported in operator-sdk v1.42.0
(which only supports `go.kubebuilder.io/v4`). Updated:

```yaml
# Before
layout:
- go.kubebuilder.io/v2
plugins:
  go.operator-sdk.io/v2-alpha: {}

# After
layout:
- go.kubebuilder.io/v4
```

The `plugins` block was removed entirely — it is not needed for Go operators in v4.

### 5f. Fix controller-gen paths to avoid scanning the module cache

The Makefile sets `GOPATH=$(PWD)/.go` (a local GOPATH inside the project directory).
controller-gen v0.20.1 with `paths="./..."` performs a literal directory walk, which
causes it to scan `.go/pkg/mod/` (the local module cache) as if it were project source.
This triggered attempts to resolve stale transitive dependencies over the network.

Fixed by replacing `paths="./..."` with targeted paths that cover only the project's
own packages:

| Target | Before | After |
|--------|--------|-------|
| `generate` | `paths="./..."` | `paths="./api/..."` |
| `manifests` | `paths="./..."` | `paths="./api/..." paths="./controllers/..."` |

The `generate` target only needs `./api/...` because DeepCopy code is generated
exclusively for API types. The `manifests` target adds `./controllers/...` to pick
up RBAC markers on controller reconcilers.

---

## Correct order for artifact generation

```
make generate manifests bundle
```

Dependency chain:

```
generate  →  controller-gen (generates DeepCopy Go code from API type markers)
manifests →  controller-gen + yq (generates CRD/RBAC YAML from Go types)
bundle    →  pre-bundle → manifests, operator-sdk, kustomize, yq
          →  update-roles-alm-example → alm-example → yq, jq
```

**Why this order matters:**

1. `generate` must run first — it regenerates `zz_generated.deepcopy.go` from type
   definitions. Subsequent steps depend on the Go code being consistent.
2. `manifests` must run before `bundle` — it regenerates CRD and RBAC YAML that
   `bundle` (via `pre-bundle`) packages into the OLM bundle. Running `manifests`
   after `bundle` would leave the bundle out of sync with the regenerated manifests.
3. `bundle` already declares `manifests` as a dependency (via `pre-bundle`), so
   Make will skip the duplicate execution of `manifests` — no double work.

---

## Verification

- `go mod tidy` — clean
- `go build ./...` — clean
- `go vet ./...` — clean
- `make generate` — clean
- `make manifests` — clean
- `make bundle` — clean; `operator-sdk bundle validate` reported `All validation tests have completed successfully`

---

## Files changed

| File | Change |
|------|--------|
| `go.mod` | `go 1.25.7` → `go 1.26.1`; `go mod tidy` run |
| `Makefile` | Tool versions bumped to latest stable; kustomize install path `/v4` → `/v5`; controller-gen `paths="./..."` → targeted per-directory paths |
| `config/default/kustomization.yaml` | `bases:` → `resources:` (kustomize v5 breaking change) |
| `PROJECT` | `layout: go.kubebuilder.io/v2` → `go.kubebuilder.io/v4`; removed unsupported `plugins` block |
