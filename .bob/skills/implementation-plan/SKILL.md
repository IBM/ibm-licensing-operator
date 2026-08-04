---
name: implementation-plan
description: Turn a task description into a detailed, actionable implementation plan written to a markdown file. Give it the task (a path to a description file, or inline prompt text - what needs to be done, where, and the constraints) and where to save the plan; the skill reads the task, grounds itself in the actual codebase, runs a planning loop, and writes a structured, step-by-step plan to the file the caller chose. It plans only - it does not implement. Use when asked to "plan", "write an implementation plan", "design an approach", or "break down" a task before coding. Repo-agnostic.
---

# implementation-plan

Produces a **detailed implementation plan** for a task and writes it to a **markdown file**.
The caller supplies the task description and the output location; this skill investigates the
codebase, thinks the change through, and saves a concrete, ordered plan someone (human or
agent) can execute step by step.

This skill **plans, it does not implement.** It changes no source files - its only output is
the plan document at the path the caller chose. It is intentionally generic: it makes no
assumptions about language, framework, or this repo. When repo conventions matter to the plan
(build/lint/test gates, codegen, architecture), pull them from the repo's own skills - here
[[operator-sdk-guide]], [[generate-manifests]], [[code-quality]], [[unit-test]], and
[[contributing]].

## Inputs

| Parameter | Required | Meaning |
|-----------|----------|---------|
| `task` | **yes** | The task description. Either a **path to a file** containing it, or **inline prompt text**. Should cover *what* needs to be done, *where*, and any *constraints*. |
| `output` | **yes** | Where to write the plan markdown file - the caller decides the path/filename. |
| `context` | no | Extra pointers: relevant dirs/files, tickets, prior art, must-not-touch areas, deadlines. Use if given. |

Take these **from the prompt that invoked this skill.**

- If `task` names a file, **read that file** as the source of truth. If it is inline text,
  use it directly. If the task is missing or too vague to plan against, ask the caller to
  expand it before starting - do not invent scope.
- If `output` is not given, **ask the caller where to save the plan** (suggest a sensible
  default such as `docs/plans/<task-slug>.md` or a path next to the task file). **Never** pick
  a location silently and never overwrite an existing file without confirming.

## The planning loop

Run this deliberately before writing anything to disk.

1. **Absorb the task.** Read the task description (and any `context`) fully. Restate, in your
   own words, the goal, the explicit constraints, and what "done" means. List anything
   ambiguous.

2. **Resolve ambiguity.** If open questions materially change the approach, ask the caller
   now rather than guessing. Only proceed on assumptions for low-risk gaps, and record every
   such assumption in the plan.

3. **Ground in the codebase.** Do **not** plan from imagination. Explore the repo: find the
   files, functions, modules, tests, and configs the task touches; read enough of them to
   know how the current code actually works and what patterns/conventions to follow. Note
   existing code to reuse and callers/dependents that a change will ripple to.

4. **Shape the approach.** Decide the overall strategy. If there are real alternatives, weigh
   them briefly and pick one with a one-line rationale (record the rejected options). Prefer
   the smallest change that satisfies the task and matches existing patterns.

5. **Decompose into ordered steps.** Break the work into concrete, independently checkable
   steps in execution order. Each step names the specific file(s)/area, what to change, and
   how to tell it worked. Sequence so the tree builds/tests at as many step boundaries as
   practical.

6. **Plan verification and risk.** Define how the change will be tested (new/updated tests,
   manual checks, the repo's test/lint/codegen gates) and call out risks, edge cases,
   migrations, and rollback.

7. **Self-review, then write.** Re-read the plan against the task: is every requirement and
   constraint covered? Is each step concrete enough to act on without re-deriving your
   research? Trim vagueness. Then write the markdown file to `output`.

Iterate steps 3-5 until the whole task is covered. For a large task, keep a running outline
so no requirement is dropped.

## Plan document structure

Write the file with this structure (drop sections that genuinely don't apply; don't pad):

```markdown
# Implementation plan: <concise task title>

## Summary
<2-4 sentences: what this change does and why.>

## Task & constraints
- **Goal:** <what must be true when done>
- **Constraints:** <hard requirements: compatibility, must-not-touch, perf, deadlines>
- **In scope / Out of scope:** <explicit boundaries>

## Assumptions & open questions
- <assumptions made, and questions still needing an answer>

## Affected areas
<files / modules / packages this touches, and dependents that ripple - with paths.>

## Approach
<the chosen strategy and the key design decisions; briefly, alternatives considered
and why rejected.>

## Implementation steps
1. **<step title>** — <file(s)/area>. <what to change>. <how to verify this step>.
2. **<step title>** — ...
   <ordered, concrete, independently checkable.>

## Testing & verification
<tests to add/update, manual checks, and the repo's build/lint/codegen/test gates to run.>

## Risks & rollback
<what could go wrong, edge cases, migrations, and how to back the change out.>

## Definition of done
- [ ] <checklist tying back to the goal and constraints>
```

## Principles

- **Ground every claim in the code.** A plan built from reading the actual repo beats a
  plausible-sounding one; cite real `file`/paths, not guesses.
- **Concrete over vague.** "Add field `X` to struct `Y` in `path/z.go` and update its two
  callers" - not "update the model". Each step should be actionable without redoing research.
- **Smallest change that fits.** Follow existing patterns; don't propose refactors the task
  didn't ask for. Flag necessary incidental cleanup separately.
- **Surface uncertainty.** State assumptions and open questions explicitly instead of hiding
  them behind confident prose.
- **Plan only.** Do not edit source. The single deliverable is the plan file at `output`.

## After writing

- Confirm the file was written to the caller's `output` path and report that path back.
- Give a one-line summary of the plan and flag any open questions that still need the
  caller's decision.

## Related skills

- [[operator-sdk-guide]] - architecture context to ground a plan that touches the operator.
- [[generate-manifests]] - if the plan changes the API/RBAC, the codegen it must include.
- [[code-quality]] / [[unit-test]] - the gates a plan's "Testing & verification" should name.
- [[code-review]] - reviews the change once the plan is implemented.
- [[contributing]] - the pre-PR checklist a plan should end up satisfying.
