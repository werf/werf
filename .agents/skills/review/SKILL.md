---
name: review
description: Code review of a pull request, branch, or diff. Covers technical, product, and risk perspectives in one pass and produces a consolidated report. Use when asked to review a PR, branch, or code changes.
---

# Code Review

Evidence-based and blunt. Every finding references a specific `file:line`, function, or component. NEVER sugarcoat, NEVER pad with praise, NEVER report a concern that is not grounded in the diff or the codebase. Style preferences are not defects.

## Before reviewing

1. Ask the user for numbered acceptance criteria (DoD). If there are none, derive them from the PR description or the linked issue, mark them `(inferred)`, and proceed — do not stall, and do not invent criteria silently.
2. Resolve the base first: `git fetch`, then diff against the branch the PR actually targets. werf maintains release branches (`1.2`, `2.63`, `3`, …), so `main` is the wrong base for a backport. State the resolved base in the report. For uncommitted work, review `git diff` / `git diff --cached` instead.
3. Read every changed file. Then trace callers of the changed exported symbols whose signature or behavior changed, and of anything crossing a persistence boundary — via LSP call hierarchy and references, not grep.
4. For 10+ changed files, split the reading by area (e.g. new files, storage/cleanup, build pipeline) across subagents if your harness has them, and synthesize the findings yourself.

## Technical perspective

Code structure and correctness only — user impact belongs to the product perspective.

- Conventions: `AGENTS.md` and `CODESTYLE.md` are the standard. This project prefers a bit of duplication over abstraction and minimizes interfaces and generics — flag deviations in either direction, and NEVER report duplication as a DRY defect on its own.
- Correctness: error wrapping and discarded errors, context propagation and cancellation, goroutine and errgroup ownership, nil map writes, typed-nil interfaces.
- Security: least privilege, input validation, secret handling, container security.
- Observability: when deploy or registry operations fail, is the cause visible in the logs?
- Testability: can the change be exercised without a cluster or a registry?
- Consistency with the werf, nelm, Docker, and Container Registry patterns already in the project.

Cover the ones the diff actually touches; stay silent about the rest.

## When the diff was written by an agent

Passing tests, high coverage, and the author's confidence are not evidence of correctness. Ask what evidence would fail if the implementation were wrong.

Would each load-bearing test fail if:

- the core behavior were removed;
- a condition were inverted;
- the return value were a constant;
- a side effect happened zero or two times?

Name the smallest mutation that should be tried. A suite that cannot detect a plausible fault is weak even when green.

High risk by default:

- weakened or deleted assertions;
- golden files updated without an explained behavioral change;
- skipped, quarantined, or filtered tests;
- lowered quality thresholds;
- test-only branches or CI-environment detection;
- fixture-specific hardcoding;
- critical behavior covered only by mocks;
- verification scripts changed together with the implementation.

## Product perspective

What the change does for the user — not how the code is written.

- User impact: CLI UX, error messages, flag names, defaults, output formatting, breaking changes.
- Completeness: edge cases (dry-run, force, conflicting flags, empty states).
- Consistency: matches existing werf CLI conventions and nelm behavior.
- Documentation: help text or docs updated where user-facing behavior changed.

## Risks

Derive risks from the technical and product findings plus the diff — including compound ones, where a technical flaw produces a product gap or an operational hazard. Likelihood is Likely/Possible/Unlikely, severity is Critical/High/Medium/Low; be realistic, do not inflate. Every risk needs a concrete location.

| Type | Covers |
| :--- | :--- |
| Technical | Architecture, performance, maintainability, testability, missing observability |
| Security | Vulnerabilities, privilege escalation, data exposure |
| UX/Product | User confusion, incomplete features, breaking changes, any change to nelm |
| Operational | Deployment issues, monitoring gaps, failure modes, registry data loss |

## Gotchas

- werf uses [nelm](https://github.com/werf/nelm) as its deploy engine — evaluate against nelm patterns, not generic Helm. A change in nelm behavior affects every werf deployment.
- Content-based tagging: tag logic affects cache invalidation and registry cleanup; users depend on predictable tags for rollback.
- Registry cleanup is destructive — users rely on dry-run modes.
- Stage digests: if a change alters what goes into a digest, every user's cache invalidates silently. Say so explicitly.
- Giterminism: any new read of uncommitted state MUST go through `giterminism_manager`.
- `*_linux.go` / `*_others.go` pairs must stay in sync, as must the Buildah and Docker backends — and a reviewer on macOS cannot compile the Buildah side at all.
- Persisted formats (stage metadata, bundles, storage records) need backward compatibility.
- werf is a CLI tool: UX, error messages, and help text are part of the product.
- Build and test only via `task` commands, never raw Go tools.

## Output

Print the report. Do not write it into the repository unless the user asks for a file.

```markdown
# Code Review Report

**Base:** `<resolved base branch>`
**Diff:** [X files, +Y/-Z lines]

## Verdict

- Technical: [up to 3 sentences, or `no findings`]
- Product: [up to 3 sentences, or `no findings`]
- Risk: [up to 3 sentences, or `no findings`]

## DoD Criteria

| Criteria | Met? | Evidence |
| :--- | :--- | :--- |
| [criterion] | ✅/⚠️/❌ | file:line |

## Issues

- **Critical** — blocking, with file:line
- **Major** — significant concern
- **Minor** — suggestion

## Risks

Sorted by severity, then by likelihood.

| № | Risk | Type | Likelihood | Severity | Location | Circumstances | Consequences |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |

## Risk Treatment

| Risk № | Recommendation |
| :--- | :--- |
```

- **Recommendation** — the concrete action, with file:line references. A single risk may get several.

## Language

Headers in English, everything else in the user's language.
