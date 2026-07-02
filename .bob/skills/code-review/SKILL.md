---
name: code-review
description: Perform a thorough, precise code review of a change entirely inside the agent thread - no GitHub or network access required. Give it a base branch and a feature branch (and optionally a short description of the change); it reviews the git diff between them, runs an explicit review loop, decides PASS or CHANGES REQUESTED, and returns a prioritized list of required fixes. Use when asked to review a branch, a PR's changes, or "the diff" before merging. Repo-agnostic.
---

# code-review

A self-contained code-review skill. It reviews the **diff between two git branches**
directly in this agent thread - it does **not** need GitHub access, a PR number, or any
network. You supply the branches; it produces a verdict and a fix list.

This skill is intentionally generic and portable: it makes no assumptions about the
language, framework, or this specific repo. When a repo *does* ship its own conventions
(linters, quality gates, contribution rules), fold those in - here they live in the
[[code-quality]] and [[contributing]] skills.

## Inputs

| Parameter | Required | Meaning |
|-----------|----------|---------|
| `base` | **yes** | The branch the change will merge **into** (e.g. `main`, `latest-4.x`). The baseline. |
| `feature` | **yes** | The branch **under review** containing the change. |
| `description` | no | A short summary of intent: what the change is meant to do, linked issue, risk areas to focus on. Improves precision - use it if given. |

If either branch is missing, ask for it before starting. Do not guess the base branch.

## Establish the diff

Review the **merge-base diff**, not a raw two-dot diff, so unrelated commits already on
`base` are excluded and you see exactly what `feature` introduces:

```bash
git fetch --all --quiet 2>/dev/null || true          # best-effort; skip if offline
git merge-base <base> <feature>                       # confirm the branches share history
git diff --merge-base <base> <feature>                # the review surface (equivalent to base...feature)
git diff --merge-base <base> <feature> --stat         # scope overview: files + churn
git log --oneline <base>..<feature>                   # the commits being introduced
```

Read the **full diff**, and open the surrounding code of any non-trivial hunk with the
file tools - a diff hunk alone hides callers, invariants, and the rest of the function.
Never review from the `--stat` summary alone.

## The review loop

Run this loop deliberately. Do not shortcut to a verdict.

1. **Understand intent.** From the `description`, commit messages, and diff, state in one
   or two sentences what this change is trying to do. If intent is unclear and no
   description was given, note it - unclear intent is itself a review finding.

2. **Map the blast radius.** List the files/functions touched and what depends on them.
   For each non-trivial hunk, read enough surrounding code to judge it in context
   (callers, error paths, the rest of the function, tests that cover it).

3. **Review each hunk against the checklist below.** For every concern, record: the
   `file:line`, what's wrong, why it matters, and the concrete fix. Distinguish facts you
   verified in the code from suspicions you could not confirm.

4. **Look for what's missing, not only what's present.** Absent tests, unhandled errors,
   an updated caller that a signature change requires, docs/generated files that should
   have changed alongside the code, a config/flag left unset.

5. **Self-check before verdict.** Re-read your findings and drop or downgrade anything you
   cannot point to a concrete failure case for. A precise review is one where every
   *blocking* finding has a plausible, statable way it goes wrong. Avoid style nitpicks
   unless a repo linter would actually reject them.

6. **Decide and report** (next section).

Iterate steps 2-4 until you have covered every changed file. If the diff is large, work
file-by-file and keep a running findings list so nothing is dropped.

## Review checklist

Weigh these; not all apply to every change.

- **Correctness** - logic errors, off-by-one, wrong operators/conditions, incorrect
  assumptions, broken control flow, misuse of APIs.
- **Edge cases** - nil/null/empty, zero, negative, boundary values, overflow, empty
  collections, unexpected input, concurrency/ordering, partial failure.
- **Error handling** - errors swallowed, ignored, or wrapped without context; missing
  rollback/cleanup; panics/exceptions on reachable paths; resource leaks (files, handles,
  connections, goroutines/threads).
- **Security** - injection, unvalidated input, authz/authn gaps, secrets in code or logs,
  unsafe deserialization, path traversal, SSRF, missing least-privilege.
- **Concurrency** - data races, deadlocks, unsynchronized shared state, context/cancellation
  not honored.
- **Contracts & compatibility** - changed function/API signatures with un-updated callers,
  breaking changes to public interfaces, schema/serialization/back-compat breaks.
- **Tests** - are new/changed code paths covered? Do tests assert behavior (not just run)?
  Are failure paths tested?
- **Readability & maintainability** - naming, dead code, needless complexity, duplication
  that should be reused, missing or misleading comments where intent is non-obvious.
- **Consistency** - does the change match the surrounding code's idioms, patterns, and
  conventions? (Read neighbors to judge this.)
- **Performance** - obvious inefficiency on a hot path: N+1, unnecessary allocation/copy,
  repeated work that should be hoisted or cached. Don't speculate about micro-perf.
- **Docs & generated artifacts** - user-facing docs, changelogs, and any generated files
  the change should have regenerated.

## Severity levels

Classify every finding:

- **BLOCKER** - must fix before merge: correctness bug, security hole, data loss, breaks
  callers/build/tests.
- **MAJOR** - should fix: likely bug under some inputs, missing error handling, missing
  test for meaningful new logic, meaningful maintainability problem.
- **MINOR** - nice to fix: readability, small consistency issues, weak naming.
- **NIT** - optional/style; never blocks on its own.

## Verdict rule

- **CHANGES REQUESTED** if there is **any BLOCKER or MAJOR** finding.
- **PASS** if there are no BLOCKER/MAJOR findings (MINOR/NIT may remain, listed as
  non-blocking).

Be decisive - always emit exactly one verdict.

## Output format

Report in the thread using this structure:

```
## Code review: <feature> → <base>

**Verdict: PASS** | **CHANGES REQUESTED**

<one-paragraph summary: what the change does and the overall assessment>

### Required fixes (blocking)
1. [BLOCKER] <file>:<line> - <what's wrong>. Fix: <concrete change>.
2. [MAJOR]   <file>:<line> - <what's wrong>. Fix: <concrete change>.

### Non-blocking suggestions
- [MINOR] <file>:<line> - <suggestion>.
- [NIT]   <file>:<line> - <suggestion>.

### Notes / could not verify
- <anything you flagged but couldn't confirm, or assumptions made>
```

If the verdict is **PASS**, the "Required fixes (blocking)" section reads `None.`
Every blocking item must be actionable: a reviewer should be able to apply the fix from
the description alone.

## Principles

- **Precision over volume.** A few real, well-explained findings beat a long list of
  guesses. Every blocking finding names a concrete way the code fails.
- **Read the context, not just the diff.** Confirm callers, invariants, and tests in the
  surrounding code before asserting a bug.
- **Review the change, not the whole codebase.** Pre-existing issues outside the diff are
  out of scope unless the change makes them materially worse.
- **Be specific.** Always cite `file:line` and give the fix, never just "this looks wrong".

## Related skills

- [[code-quality]] - run the repo's actual linters/formatters/secret-scan; a clean review
  still needs the mechanical gate to pass.
- [[unit-test]] - verify the change's tests actually pass.
- [[contributing]] - the repo's pre-PR checklist the change must satisfy before merge.
