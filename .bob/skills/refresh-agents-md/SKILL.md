---
name: refresh-agents-md
description: Keep the root AGENTS.md agent guide accurate and in sync with the repo - refresh it after changes to the Makefile targets, toolchain versions, api/ or controllers/ architecture, config/bundle layout, the versioning scheme, or the .bob/skills/ catalog. Use when AGENTS.md is stale, when onboarding guidance drifts from reality, or after adding/removing/renaming a skill.
---

# refresh-agents-md

`AGENTS.md` (repo root) is the entrypoint guide for AI coding agents. It is a hand-maintained
summary that points at the authoritative sources - the `Makefile`, `CONTRIBUTING.md`, the code,
and the `.bob/skills/` catalog. It drifts as the repo changes, so refresh it deliberately rather
than letting it rot.

## When to use

- A `make` target used in AGENTS.md was added, removed, or renamed.
- A pinned tool or its version changed (top of the `Makefile`).
- The architecture moved (new/renamed dirs under `api/`, `controllers/`, `pkg/`, `config/`).
- The bundle/config/deploy layout or the versioning scheme changed.
- **A skill was added, removed, or renamed in `.bob/skills/`** - the skills table in AGENTS.md
  and the mentions of skills must match the catalog exactly.
- AGENTS.md simply reads as stale or contradicts current behavior.

## How to refresh

Work from the sources of truth, not from memory:

1. **Skills catalog.** List `.bob/skills/*/SKILL.md` and read `.bob/skills/index.json`.
   Make the skills table near the top of AGENTS.md match one-to-one:
   every skill directory has a row (task, skill name, `.bob/skills/<name>/SKILL.md` path) and no
   row points at a skill that no longer exists.

   ```bash
   ls .bob/skills/*/SKILL.md
   ```

2. **Commands.** Cross-check every `make` target quoted in AGENTS.md against the `Makefile`:

   ```bash
   grep -nE '^[a-zA-Z0-9/_-]+:' Makefile        # target names
   make help                                     # documented targets + descriptions
   ```

   Fix renamed/removed targets; keep the inline `Skill: <name>` pointers on each command group.

3. **Toolchain.** Reconcile the tool list against the pinned versions at the top of the
   `Makefile` (the `*_VERSION` vars and the `LOCALBIN` tool definitions).

4. **Architecture.** Verify the `main.go` / `api/` / `controllers/` / `controllers/resources/`
   descriptions still hold (dir listing + a skim of `main.go` and the reconciler). Update any
   moved or renamed paths.

5. **Versioning & workflow.** Confirm the version still lives in `Makefile` `CSV_VERSION` and
   `version/version.go`, and that the workflow notes match `CONTRIBUTING.md` (DCO, pre-commit,
   commit-generated-files).

## Invariants to preserve

- **No agent-vendor-specific references.** AGENTS.md is vendor-neutral: no product/assistant
  names, no `CLAUDE.md`-style naming. Keep the header generic ("AI coding agents").
- **Point, don't duplicate.** AGENTS.md summarizes and links to `.bob/skills/`, `Makefile`, and
  `CONTRIBUTING.md`; it is not the place to copy their full contents. When in doubt, add a
  pointer to the relevant skill instead of expanding prose.
- **The skills table is the contract** between AGENTS.md and `.bob/skills/` - keep it exact.

## After editing

```bash
grep -in claude AGENTS.md          # must return nothing (vendor-neutral guard)
make lint-markdown                 # if touching markdown lint is in scope; see code-quality
```

## Related skills

- [[contributing]] - the workflow/checklist AGENTS.md summarizes.
- [[generate-manifests]] - the codegen chain AGENTS.md references.
- [[operator-sdk-guide]] - the architecture AGENTS.md summarizes.
