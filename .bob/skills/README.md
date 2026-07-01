# IBM Licensing Operator - Bob Skills

Skills for the **bob** AI agent to work productively in the `ibm-licensing-operator`
repository. Each skill lives in its own directory as a `SKILL.md` with YAML frontmatter
(`name`, `description`) followed by task instructions, so bob can select the right one by
its `description`.

These skills wrap the repo's real `make` targets (source of truth: the `Makefile` and
`CONTRIBUTING.md`) and cover the developer inner loop. Release-only concerns - multi-arch
image assembly, OLM catalog builds, CSV publishing - are intentionally **not** skills:
those run in the Tekton pipeline (`.pipeline-config-*.yaml`), not on a developer machine.

## Skills

| Skill | Purpose | Key commands |
|-------|---------|--------------|
| **setup-tools** | Install / verify / clean the pinned local toolchain in `./bin`. | `make install-all-tools`, `make verify-installed-tools`, `make clean` |
| **code-quality** | Pre-commit gate: format, tidy, vet, lint, secret-scan. | `make code-dev`, `make check`, `make audit` |
| **unit-test** | Run controller tests (Ginkgo/envtest) against a cluster. | `make prepare-unit-test && make unit-test` |
| **generate-manifests** | Regenerate DeepCopy, CRDs, RBAC, OLM bundle, ArgoCD YAMLs after API changes. | `make generate manifests`, `make bundle` |
| **build-and-deploy** | Build / run / deploy the operator locally. | `make build`, `make run`, `make install && make deploy` |
| **build-helm-charts** | Build dev Helm charts (no-operator path). | `make build/helm-develop-all` |
| **contributing** | How to land quality changes: workflow, DCO, checklist, version bumps. | *(knowledge)* |
| **operator-sdk-guide** | How kubebuilder/operator-sdk works for this operator. | *(knowledge)* |

## Typical inner loop

1. **setup-tools** - once, or when a tool is missing.
2. Edit code (`api/`, `controllers/`, `config/`, …).
3. **generate-manifests** - if you changed the API or RBAC markers.
4. **code-quality** - format + lint + secret scan.
5. **unit-test** - validate controller/API behavior.
6. **build-and-deploy** - run or deploy to a cluster to try it out.
7. **contributing** - final pre-PR checklist and DCO sign-off.

Reach for **operator-sdk-guide** any time you need to understand the architecture (CRDs,
controllers, reconcile flow, codegen chain) before making a change.

## Skill format

Each `SKILL.md` starts with YAML frontmatter:

```markdown
---
name: <skill-name>
description: <what it does and when bob should use it>
---

# <skill-name>
...instructions, commands, prerequisites, notes, and [[links]] to related skills...
```

`index.json` is a machine-readable summary of the same set.

## Adding a skill

1. Create `.bob/skills/<skill-name>/SKILL.md` with the frontmatter above.
2. Prefer wrapping an existing `make` target over inventing new commands.
3. Add a row to the table here and an entry to `index.json`.
4. Cross-link related skills with `[[skill-name]]`.
