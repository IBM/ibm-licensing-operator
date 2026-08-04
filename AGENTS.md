# AGENTS.md

Guidance for AI coding agents working in this repository.

## Repo-specific skills (read these first)

This repo ships task-oriented skills in **`.bob/skills/`**. Each is a `SKILL.md` selected by
its `description`. **Before starting a task, load the matching skill and follow it** — the
skills are the source of truth for the developer inner loop and wrap the real `make` targets
with detailed, up-to-date instructions. See `.bob/skills/index.json` for the full catalog.

| Task | Skill | Path |
|------|-------|------|
| Turn a task description into a detailed implementation plan (markdown file) | **implementation-plan** | `.bob/skills/implementation-plan/SKILL.md` |
| Install / verify / clean the pinned local toolchain | **setup-tools** | `.bob/skills/setup-tools/SKILL.md` |
| Format, tidy, vet, lint, secret-scan (pre-commit gate) | **code-quality** | `.bob/skills/code-quality/SKILL.md` |
| Run controller tests (Ginkgo/envtest against a cluster) | **unit-test** | `.bob/skills/unit-test/SKILL.md` |
| Review a pull request (via `gh`) and get a PASS / CHANGES REQUESTED verdict + fixes | **code-review** | `.bob/skills/code-review/SKILL.md` |
| Regenerate DeepCopy, CRDs, RBAC, OLM bundle, ArgoCD YAMLs | **generate-manifests** | `.bob/skills/generate-manifests/SKILL.md` |
| Build / run / deploy the operator locally | **build-and-deploy** | `.bob/skills/build-and-deploy/SKILL.md` |
| Build the stand-alone (no-operator) Helm charts | **build-helm-charts** | `.bob/skills/build-helm-charts/SKILL.md` |
| Contribution workflow, DCO, pre-PR checklist, version bumps | **contributing** | `.bob/skills/contributing/SKILL.md` |
| How operator-sdk/kubebuilder works here (architecture) | **operator-sdk-guide** | `.bob/skills/operator-sdk-guide/SKILL.md` |
| How License Service splits across the operator / operand / commons repos | **license-service-architecture** | `.bob/skills/license-service-architecture/SKILL.md` |
| Keep **this file** (`AGENTS.md`) in sync with the repo & skills | **refresh-agents-md** | `.bob/skills/refresh-agents-md/SKILL.md` |

Typical inner loop: **implementation-plan** (for non-trivial tasks) → **setup-tools** →
edit code → **generate-manifests** (if API/RBAC changed) → **code-quality** → **unit-test**
→ **code-review** → **build-and-deploy** → **contributing** (pre-PR).

### Planning a task (implementation-plan skill)

For a non-trivial task, use the **implementation-plan** skill to turn a task description into
a detailed, ordered plan grounded in the actual code, written to a markdown file. It plans
only — it does not edit source.

```text
# Ask the agent, e.g.:
"Use the implementation-plan skill to plan the task in <task-file.md>; save the plan to docs/plans/<name>.md."
"Use the implementation-plan skill for: <inline task description>. Write the plan to <path>."
```

Inputs: `task` (a path to a description file, or inline text — what/where/constraints) and
`output` (where to save the plan; the caller chooses — the skill asks if it's omitted).
`context` is optional. See `.bob/skills/implementation-plan/SKILL.md`.

### Reviewing a pull request (code-review skill)

To review a PR in the agent thread, use the **code-review** skill. It uses the `gh` CLI to
fetch the PR and its diff, runs an explicit review loop, and returns a
PASS / CHANGES REQUESTED verdict plus a prioritized list of required fixes.

```text
# Ask the agent, e.g.:
"Use the code-review skill to review PR #1234."
"Use the code-review skill to review the PR open from branch <branch>."
# Optionally add a short description of the change to focus the review.
```

Inputs: a `pr` number (preferred) or a `branch` (the skill finds the open PR from it via
`gh pr list --head`); `description` is optional. Requires `gh` installed and authenticated
(`gh auth status`). See `.bob/skills/code-review/SKILL.md`.

> Keep this file current with the **refresh-agents-md** skill
> (`.bob/skills/refresh-agents-md/SKILL.md`) whenever the Makefile targets, toolchain,
> architecture, versioning, or the `.bob/skills/` catalog change.

## Overview

`ibm-licensing-operator` is a Kubernetes/OpenShift operator (Go, built with operator-sdk /
kubebuilder v4) that installs and manages **License Service**, which collects license-usage
data for IBM containerized products. It runs either as part of IBM Cloud Pak foundational
services or stand-alone (the "no-operator" Helm path).

This repo is **only the operator**. The License Service application it deploys (the *operand*)
and the shared library that operand builds on (the *commons*) live in **separate repositories**
with independent release cycles — the operator references the operand only as a pinned
container image. See the **license-service-architecture** skill for how the three fit together
(and don't assume the other repos are checked out locally).

## Toolchain

All build/lint/codegen tools are pinned in the `Makefile` (versions at the top) and installed
into the gitignored `./bin` directory, which is prepended to `PATH` by the Makefile. Do **not**
rely on globally installed tools — always go through `make`. See the **setup-tools** skill.

```bash
make install-all-tools      # install/verify the entire pinned toolchain into ./bin
make verify-installed-tools # check tool versions match
make clean                  # remove ./bin
```

Key tools: `controller-gen`, `kustomize`, `operator-sdk`, `opm`, `yq`, `helm`, `golangci-lint`,
`goimports`, `detect-secrets`, plus `shellcheck`/`yamllint`/`mdl` for non-Go linting.

## Common commands

Each group below has a corresponding skill in `.bob/skills/` with fuller instructions.

```bash
# Quality gate — run before every commit / PR (mirrors CI & the pre-commit hook). Skill: code-quality
make code-dev        # go mod tidy + fmt + goimports + vet + make check
make check           # all linters (lint-all) + go vet
make audit           # detect-secrets scan against .secrets.baseline

# Codegen — MUST run after changing anything in api/ or +kubebuilder/RBAC markers. Skill: generate-manifests
make generate        # regenerate zz_generated.deepcopy.go (DeepCopy methods)
make manifests       # regenerate CRDs (config/crd), RBAC (config/rbac)
make bundle          # regenerate the OLM bundle/ + CSV (uses operator-sdk)
# Always commit generated files together with the API change.

# Build & run. Skill: build-and-deploy
make build           # build operator binary to bin/ibm-licensing-operator (via common/scripts/gobuild.sh)
make run             # go run ./main.go against ~/.kube/config (NAMESPACE=<ns> to override)
make install         # kustomize build config/crd | kubectl apply  (install CRDs)
make deploy          # deploy the operator via config/default
make build-push-image-development  # build + push image to the scratch registry

# Helm (stand-alone "no-operator" path). Skill: build-helm-charts
make build/helm-develop-all   # build LS, LSR (reporter), LSS (scanner) dev charts
```

## Tests

Controller tests use **Ginkgo + envtest** and run against a **real cluster**
(`USE_EXISTING_CLUSTER=true`) — there is no fully mocked control plane. You need a reachable
Kubernetes/OpenShift cluster (`~/.kube/config`). See the **unit-test** skill.

```bash
make prepare-unit-test   # create namespaces, apply CRDs + required external CRDs (ODLM,
                         # Prometheus, gateway-api, meterdefinitions) into the cluster
make unit-test           # go test ./controllers/... -coverprofile cover.out -timeout 30m
```

Run a single package / focused spec:

```bash
# after prepare-unit-test and exporting the same env vars the target sets
# (OPERATOR_NAMESPACE, WATCH_NAMESPACE, NAMESPACE, IBM_LICENSING_IMAGE, USE_EXISTING_CLUSTER=true)
go test ./controllers/... -run TestAPIs -ginkgo.focus="<spec text>" -v
```

## Architecture

For a deeper walkthrough of the operator-sdk/kubebuilder structure, see the
**operator-sdk-guide** skill (`.bob/skills/operator-sdk-guide/SKILL.md`).

**Entrypoint — `main.go`:** registers all API schemes, builds a controller-runtime manager
with a **namespace-scoped cache** (watches `WATCH_NAMESPACE`; Secrets/Deployments/Pods are
label-filtered to `release in (ibm-licensing-service)`; gateway-api objects cached only in the
operator namespace). It probes the cluster at startup for optional CRDs (gateway-api, ODLM
OperandRequest/OperandBindInfo, OperatorGroup, Namespace Scope) and wires controllers/goroutines
conditionally based on what exists and on the active CR's `features.operandRequestsEnabled` flag.

**APIs — `api/`:**
- `api/v1alpha1` — the primary `IBMLicensing` CRD (cluster-scoped) plus `IBMLicensingMetadata`;
  `features/` holds optional-feature specs (auth, alerting, hyper-threading, prometheus query
  source). `IBMLicensingSpec` uses `PreserveUnknownFields` and inlines several embedded structs.
- `api/v1` — `IBMLicensingDefinition`, `IBMLicensingQuerySource`.
- `zz_generated.deepcopy.go` files are generated by `make generate` — never edit by hand.

**Controllers — `controllers/`:**
- `ibmlicensing_controller.go` — the main reconciler. `Reconcile` selects a single **active**
  IBMLicensing instance (oldest CR wins), then runs a sequence of focused `reconcile*` methods
  (tokens/secrets, configmaps, services, service monitors, network policy, deployment,
  certificates, route, and gateway-api exposure). Each returns `(reconcile.Result, error)`.
- `operandrequest_controller.go` + `operandrequest_discovery.go` — ODLM OperandRequest support
  (only started when the ODLM CRD exists and the feature flag is on).
- `operatorgroup_cleaner.go` — background task removing stale namespaces from the OperatorGroup
  (skipped when Namespace Scope Operator is active, per Cloud Pak coexistence rules).
- `controllers/resources/` — the resource-builder layer: pure(ish) functions that construct
  desired k8s objects (`deployments.go`, `containers.go`, `envs.go`, `volumes.go`, `crds.go`,
  cluster-capability detection like `IsGatewayAPI`, `DoesCRDExist`, `UpdateResource`, etc.).
  Most controller logic that builds or diffs a k8s object belongs here, not in the reconciler.

**`pkg/rhmp/`** — vendored-style Red Hat Marketplace meterdefinition types used by the scheme.

## Manifests, bundle & config

- `config/` — kustomize sources (crd, rbac, manager, default, samples, manifests). These are the
  inputs; `make manifests`/`make bundle` regenerate the outputs.
- `bundle/` + `bundle.Dockerfile` — the generated OLM bundle (CSV, CRDs, metadata).
- `deploy/` — raw YAMLs and an ArgoCD deployment path (`make generate-yaml-argo-cd`).
- `helm-no-operator/`, `helm-migration/` — Helm charts for the stand-alone / migration paths.
- `common/` — shared IBM build scripts (`Makefile.common.mk`, `scripts/`) used by the Makefile.

## Versioning

Operator version lives in two places kept in sync: `CSV_VERSION` in the `Makefile` and
`Version` in `version/version.go`. Bumping the version is a dedicated step (see the **contributing**
skill and the "Version bump" section of `CONTRIBUTING.md`) and regenerates the bundle.

## Workflow notes

See the **contributing** skill (`.bob/skills/contributing/SKILL.md`) for the full checklist.

- Commits require **DCO sign-off** (`git commit -s`).
- Run `make code-dev` (or at least `make check`) before pushing; a pre-commit hook enforces it.
- When you change the API or RBAC markers, regenerate **and commit** CRDs + bundle manifests
  in the same change (**generate-manifests** skill).
- Release-only concerns (multi-arch image assembly, catalog/CSV publishing) run in the Tekton
  pipeline (`.pipeline-config-*.yaml`), not on a dev machine.
