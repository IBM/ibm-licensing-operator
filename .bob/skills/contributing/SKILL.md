---
name: contributing
description: How to contribute high-quality changes to the ibm-licensing-operator repo - the end-to-end workflow, DCO sign-off, the pre-PR quality checklist, when generated files must be committed, copyright headers, branching, and version bumps. Use when preparing a change, a commit, or a pull request, or when unsure what the repo expects before pushing.
---

# contributing

The quality bar and workflow for landing changes in this repo. Follow this before every
commit and PR. It combines `CONTRIBUTING.md`, `README.md`, and the repo's enforced hooks.

## Branching

- `master` - code for the **next** release.
- `develop` - current development.
- `latest-4.x` / release branches - already-released versions (the default PR base here
  is `latest-4.x`).

Confirm the intended base branch before opening a PR; don't assume `master`.

## Commit requirements

- **DCO sign-off is mandatory.** Every commit must have a `Signed-off-by:` line whose
  email matches the author. Use `git commit -s`. A probot check enforces this on PRs.
- Keep the Apache 2.0 copyright header on every source file - the `lint-copyright-banner`
  linter fails new files that lack it. Copy the header from a sibling file.

## The pre-PR checklist

Run these in order before opening a PR (an installed pre-commit git hook also runs
`make code-dev` on every commit):

1. **Quality gate** - `make code-dev` (tidy, fmt, vet, all linters). See [[code-quality]].
2. **Secret scan** - `make audit` (detect-secrets). See [[code-quality]].
3. **Regenerate if the API changed** - if you edited `api/` or `controllers/` markers,
   run `make bundle` and **commit the generated CRDs, RBAC, and bundle manifests** with
   your source change. See [[generate-manifests]].
4. **Tests** - `make unit-test` for controller/API changes. See [[unit-test]].
5. **Build check** - `make build` to confirm it compiles. See [[build-and-deploy]].

## What "good quality" means here

- **Generated files are part of the change.** Never hand-edit `zz_generated.deepcopy.go`,
  `config/crd/bases/*`, or `bundle/manifests/*` - regenerate them and commit the result.
- **Match the surrounding code.** Follow existing controller structure, error-handling,
  and logging patterns rather than introducing new idioms.
- **No secrets in code.** Use Kubernetes Secrets or env vars; `make audit` must be clean.
- **Keep `temp/` out of commits** - it holds intermediate build files and is gitignored.
- **Don't bump `CSV_VERSION` casually.** Version bumps use
  `common/scripts/next_csv.sh <current> <new> <old>` (e.g. `... 4.2.20 4.2.21 4.2.19`),
  which updates the CSV, skipRange, and related metadata consistently.

## Contributing a patch (from CONTRIBUTING.md)

1. Open an issue describing the proposed change.
2. Fork, develop, and test.
3. Commit with DCO (`-s`).
4. Open a PR against the correct base branch.

## Environment expectations

- go **1.26+**, `GO111MODULE=on`, `GOROOT` pointing at the 1.26 toolchain.
- All tools installed locally via `make install-all-tools` (see [[setup-tools]]).

## macOS git-hook tip

If IDE commits fail because the pre-commit hook can't find tools, extend `PATH`
inside `common/scripts/.githooks/pre-commit` with the output of `echo $PATH` from a
working terminal (keep this change local - don't push it).

## Related skills

- [[code-quality]], [[unit-test]], [[generate-manifests]], [[build-and-deploy]],
  [[setup-tools]], [[operator-sdk-guide]].
