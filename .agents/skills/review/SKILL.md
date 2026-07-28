---
name: review
description: Code review of a pull request, branch, or diff. Covers technical, product, and risk perspectives in one pass and produces a consolidated report. Use when asked to review a PR, branch, or code changes.
---

# Code Review

Evidence-based and blunt. Every finding references a specific `file:line`, function, or component. NEVER sugarcoat, NEVER pad with praise, NEVER report a concern that is not grounded in the diff or the codebase. Style preferences are not defects.

## Before reviewing

1. Ask the user for numbered acceptance criteria (DoD). Block until received — nothing proceeds without them.
2. Get the diff for the change under review: `git diff origin/main...HEAD`, or the range/PR the user named. If the branch is not on `origin`, compare against local `main`.
3. Read every changed file **and its callers** — the diff alone hides breakage in consumers.
4. For 10+ changed files, split the reading across parallel `Agent` subagents by area (e.g. new files, storage/cleanup, build pipeline) and synthesize their findings yourself.

## Technical perspective

Code structure and correctness only — user impact belongs to the product perspective.

| Principle | What to check |
| :--- | :--- |
| SOLID | SRP per type, OCP for extensibility, ISP for interface size. |
| DRY | Duplicated logic, config, or error handling. |
| KISS/YAGNI | Unnecessary abstraction, generics, or interfaces. |
| Security | Least privilege, input validation, secret handling, container security. |
| Observability | Logs/metrics for critical paths (deploy, registry ops). |
| Testability | Can the change be validated without integration setup? |

Also check consistency with existing werf, nelm, Docker, and Container Registry patterns in the project, and with the conventions in `AGENTS.md` and `CODESTYLE.md`.

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

Derive risks from the technical and product findings plus the diff — including compound ones, where a technical flaw produces a product gap or an operational hazard. Probability is 0.0-1.0, severity is Critical/High/Medium/Low; be realistic, do not inflate. Every risk needs a concrete location.

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
- werf is a CLI tool: UX, error messages, and help text are part of the product.
- Build and test only via `task` commands, never raw Go tools.

## Output

Save the report to `reviews/<branch-with-slashes-replaced-by-dashes>/REPORT.md` and print it.

```markdown
# Code Review Report

**Branch:** `<branch>`
**Diff:** [X files, +Y/-Z lines]

## Verdict

- Technical: [2-3 sentences]
- Product: [2-3 sentences]
- Risk: [2-3 sentences]

## DoD Criteria

| Criteria | Met? | Evidence |
| :--- | :--- | :--- |
| [criterion] | ✅/⚠️/❌ | file:line |

## Best Practices

| Practice | Status | Comments |
| :--- | :--- | :--- |
| SOLID / DRY / KISS-YAGNI / Security / Observability / Testability | ✅/⚠️/❌ | file:line — one-liner |

## Issues

- **Critical** — blocking, with file:line
- **Major** — significant concern
- **Minor** — suggestion

## Risks

Sorted by severity, then by probability descending.

| № | Risk | Type | Probability | Severity | Location | Circumstances | Consequences |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |

## Risk Treatment

| Risk № | Severity | Strategy | Recommendation | Justification |
| :--- | :--- | :--- | :--- | :--- |
```

- **Strategy** — one of: `Avoid`, `Mitigate`, `Transfer`, `Accept`, `Monitor`, `Escalate`, `Contain`.
- **Recommendation** — concrete action with file:line references.
- A single risk may get several recommendations.

## Language

Headers in English, everything else in the user's language.
