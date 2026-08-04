---
name: code-quality
description: Format, tidy, vet, lint, and secret-scan the ibm-licensing-operator codebase before committing or opening a PR. Runs the same checks the pre-commit git hook and CI enforce (golangci-lint, shellcheck, yamllint, mdl, go vet, detect-secrets). Use before every commit, when preparing a PR, or after any code/YAML/script/markdown change.
---

# code-quality

The single quality gate for this repo. Run it after making changes and before
committing - the installed pre-commit git hook runs `make code-dev` (which ends in
`make check`), and the Tekton PR pipeline runs the same linters, so passing locally
means the branch will pass CI checks.

## When to use

- Before committing or opening a PR (mandatory - a pre-commit hook runs `make code-dev`).
- After editing Go, shell, YAML, Dockerfile, or Markdown files.
- When a CI lint stage fails and you want to reproduce it locally.

## Commands

### Everyday check (recommended)

```bash
make code-dev
```

Runs, in order: `go mod tidy` → `go fmt`/`goimports` → `go vet` → `make check`
(all linters). This is the standard "am I ready to commit?" command.

### Individual pieces

```bash
make check      # all linters + go vet (the final step of code-dev)
make lint       # lint-all + vet: shellcheck, yamllint, copyright-banner, golangci-lint, mdl
make fmt        # format Go code (goimports)
make vet        # go vet ./...
```

### Secret scan (run before every commit)

```bash
make audit
```

Runs `detect-secrets` against the tree, updating `.secrets.baseline`, and opens the
audit view. Excludes lock/dependency files (`go.mod`, `go.sum`, `package-lock.json`,
etc.). If a **real** secret is flagged: remove it from code, use a Kubernetes Secret
or env var instead, and rotate the credential if it was ever committed. Only update
`.secrets.baseline` for genuine false positives.

## Linters that run

- **golangci-lint** (v2.11.2) - Go static analysis (staticcheck, govet, gosimple, …).
- **shellcheck** (v0.11.0) - shell scripts under `common/scripts/` and elsewhere.
- **yamllint** (1.37.1) - YAML syntax/formatting.
- **mdl** (0.15.0) - Markdown docs.
- **lint-copyright-banner** - every source file must carry the Apache 2.0 header.
- **go vet** - suspicious Go constructs.

## Prerequisites

Linters must be installed locally - see [[setup-tools]] (`make install-linters`).

## Notes

- The copyright-banner linter fails on new files missing the license header; copy the
  header from an existing file of the same type.
- If `make check` fails with a `goboringcrypto.h` / OpenSSL error, your Go toolchain is
  too old - install go 1.26+ and set `GOROOT` accordingly (see `CONTRIBUTING.md`).

## Related skills

- [[unit-test]] - run after quality passes.
- [[generate-manifests]] - regenerate CRDs/bundle if you changed the API.
- [[contributing]] - the full pre-PR checklist.
