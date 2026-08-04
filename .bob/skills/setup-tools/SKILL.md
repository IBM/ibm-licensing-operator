---
name: setup-tools
description: Install, verify, and clean the local development toolchain (operator-sdk, opm, controller-gen, kustomize, yq, linters, detect-secrets) for the ibm-licensing-operator repo. Use for first-time setup, onboarding, when a build fails because a tool is missing or the wrong version, or to reset a broken toolchain.
---

# setup-tools

Manage the pinned local toolchain that every other skill in this repo depends on.
All tools are installed into the repo-local `./bin/` directory (never system-wide),
so their versions match exactly what CI/Tekton uses.

## When to use

- **First-time setup** after cloning the repo.
- **Onboarding** a new developer environment.
- A build/lint/bundle target fails with `Required tool: <x> is not installed` or a
  version mismatch.
- You want a clean slate to fix tool version conflicts.

## Commands

Install (or top up) every required tool into `./bin/`:

```bash
make install-all-tools
```

Verify presence and print required-vs-installed versions:

```bash
make verify-installed-tools
```

Install only the linting tools (shellcheck, yamllint, golangci-lint, mdl):

```bash
make install-linters
```

Remove all binaries and installed tools (forces a clean reinstall next build):

```bash
make clean          # deletes ./bin/
```

## Pinned versions (source of truth: Makefile)

| Tool | Version |
|------|---------|
| operator-sdk | v1.42.1 |
| opm | v1.64.0 |
| controller-gen | v0.20.1 |
| kustomize | v5.8.1 |
| yq | v4.52.4 |
| golangci-lint | v2.11.2 |
| shellcheck | v0.11.0 |
| yamllint | 1.37.1 |
| mdl | 0.15.0 |
| goimports | v0.43.0 |
| detect-secrets | IBM fork, tracks `master` (installed via `install-detect-secrets.sh`) |
| helm | v4.1.4 (installed on demand by `make helm`, not by `install-all-tools`) |

Go toolchain: **go 1.26.5** (see `go.mod`).

## Notes

- `make clean` only removes `./bin/`. Follow it with `make install-all-tools` to rebuild.
- Some linters (yamllint via pip, mdl via gem) need Python and Ruby available on PATH.
- If tool versions drift, prefer `make clean && make install-all-tools` over manual fixes.

## Related skills

- [[code-quality]] - needs the linters installed here.
- [[generate-manifests]] - needs controller-gen, kustomize, yq, operator-sdk.
- [[contributing]] - describes the full development workflow.
