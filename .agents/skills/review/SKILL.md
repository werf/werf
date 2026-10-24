---
name: review
description: Code review of a pull request, branch, or diff. Covers technical, product, and risk perspectives in one pass and produces a consolidated report. Use when asked to review a PR, branch, or code changes.
---

# Code Review

Evidence-based and blunt. Every finding references a specific `file:line`, function, or component. NEVER sugarcoat, NEVER pad with praise, NEVER report a concern that is not grounded in the diff or the codebase. Style preferences are not defects — but a violation of `AGENTS.md` or `CODESTYLE.md` is a convention finding, not a preference.

## If you are the diff's author

If you wrote the diff you are reviewing now, say so explicitly in the report's Verdict and treat this pass as necessary but insufficient. `challenge-review/SKILL.md` covers why self-review inherits your own design assumptions and what to recommend instead — read it, don't re-derive it here.

## Before reviewing

1. Recover acceptance criteria from the user request, current task contract, PR description and linked issue. Mark derived criteria `(inferred)` and proceed. Ask only about unresolved choices that materially change the work; do not ask the user to repeat criteria already supplied.
2. Resolve the base first: `git fetch`, then diff against the branch the PR actually targets. werf maintains release branches (`1.2`, `2.63`, `3`, …), so `main` is the wrong base for a backport. State the resolved base in the report. For uncommitted work, review `git diff` / `git diff --cached` instead.
3. If the PR head is not checked out, take every `file:line` from that commit's blob (`git show <head>:<path>`), never from the worktree. The worktree sits on the base, so any file the PR itself changed has different line numbers there, and a comment anchored on a worktree line number lands on the wrong line — or is rejected outright.
4. Read every changed file. Then trace callers of the changed exported symbols whose signature or behavior changed, and of anything crossing a persistence boundary — via LSP call hierarchy and references, not grep. LSP indexes the worktree, so when the head is not checked out it answers about the base: add a worktree at the head first, or read blobs and say in the report that call tracing was limited.
5. For 10+ changed files, split the reading by area (e.g. new files, storage/cleanup, build pipeline) across subagents if your harness has them, and synthesize the findings yourself.
6. For runtime-affecting changes, require `task build` and `task test:unit` on the candidate in a supported environment. Apply the repeat-review rules below when reusing that evidence. Prose-only edits need no build. NEVER run `task format`: it would rewrite the diff under review.

## Repeat reviews

The first review covers the whole change. A subsequent review starts from the
previously reviewed SHA, its open findings and the delta to the current candidate;
trace affected callers and invariants beyond the diff. Verify the previous scope
and evidence before reusing them. A prior ACCEPT never approves a new SHA.

One full check per candidate and relevant environment is the default, with a named
owner and inspectable revision, commands, exit statuses and logs. Reviewers validate
that evidence and independently exercise changed behavior; they need not each rerun
the entire suite. Record the reason before expanding: changed dependencies, base,
build/test infrastructure, broad behavior impact, missing or unreliable evidence,
or explicit repository/user requirements. A rebase requires inspecting the base
and conflict-resolution changes as well as the patch.

Documentation-only deltas need no code build unless they affect executable or
generated inputs. Runtime-affecting changes still require the repository's final
checks. For deletion, migration, security or concurrency changes, retain direct
inspection and targeted failure-path checks and meaningful mutations; shared green
logs cannot replace these. Report current SHA, reused evidence and its limits,
new checks, remaining findings and verdict.

## Technical perspective

Code structure and correctness only — user impact belongs to the product perspective.

- Conventions: `AGENTS.md` and `CODESTYLE.md` are the standard. This project prefers a bit of duplication over abstraction and minimizes interfaces and generics — flag deviations in either direction, and NEVER report duplication as a DRY defect on its own.
- Correctness: error wrapping and discarded errors, context propagation and cancellation, goroutine and errgroup ownership, nil map writes, typed-nil interfaces.
- Security: least privilege, input validation, secret handling, container security.
- Observability: when deploy or registry operations fail, is the cause visible in the logs?
- Testability: can the change be exercised without a cluster or a registry?
- Consistency with the werf, nelm, Docker, and Container Registry patterns already in the project.

Cover the ones the diff actually touches; stay silent about the rest.

## Tests as evidence

Passing tests are evidence with limits, not proof. Read `test-the-tests/SKILL.md` and mutate plausible regressions for load-bearing tests of changed behavior. In a repeat review, cover the delta and affected invariants rather than repeating mutations of unchanged behavior with valid evidence. A prose-only edit without executable behavior needs no synthetic mutation test.

For a non-trivial or high-risk diff, or a diff that touches tests or verification infrastructure, also read `challenge-review/SKILL.md` as an independent challenge pass. It covers check-gaming detection (weakened assertions, quietly skipped tests, mocked-out critical behavior, and more) in one place, so this list doesn't drift from it again.

## Product perspective

What the change does for the user — not how the code is written.

- User impact: CLI UX, error messages, flag names, defaults, output formatting, breaking changes.
- Completeness: edge cases (dry-run, force, conflicting flags, empty states).
- Consistency: matches existing werf CLI conventions and nelm behavior.
- CLI surface: every flag needs its `WERF_*` env counterpart, a renamed or removed flag needs a deprecation path, and exit codes plus machine-readable output (`--build-report-path`, `--save-deploy-report`) are parsed by users' CI — changing that schema is a breaking change.
- Documentation: CLI reference pages are generated, so the fix for a stale one is `task doc:gen`, never a hand edit, and a hand-edited `CHANGELOG.md` is itself a defect (release-please owns it). Feature docs under `pages_en` need their `pages_ru` counterpart.

## Risks

Derive risks from the technical and product findings plus the diff — including compound ones, where a technical flaw produces a product gap or an operational hazard. Likelihood is Likely/Possible/Unlikely, severity is Critical/High/Medium/Low; be realistic, do not inflate. Every risk needs a concrete location.

Classify each risk as Technical, Security, UX/Product, or Operational, and report risks only when they exist — an empty matrix is noise. A `go.mod` bump of `nelm`, `kubedog`, or `common-go` carries the widest blast radius here: it silently changes deploy behavior for everyone.

## Gotchas

- werf uses [nelm](https://github.com/werf/nelm) as its deploy engine — evaluate against nelm patterns, not generic Helm.
- Stage digests and content-based tags: if a change alters what goes into a digest, every user's cache invalidates and their tags move, which breaks rollback. Say so explicitly.
- Registry cleanup is destructive — check that `--dry-run` and the keep policies still hold.
- Giterminism: any new read of uncommitted state MUST go through `giterminism_manager`.
- `*_linux.go` / `*_others.go` pairs must stay in sync, as must the Buildah and Docker backends — and a reviewer on macOS cannot compile the Buildah side at all.
- Persisted formats (stage metadata, bundles, storage records) need backward compatibility.
- `go.mod` replaces cobra and oras with `werf/3p-*` forks — upstream documentation is not authoritative for them.
- Build and test only via `task` commands, never raw Go tools.
- The PR description outlives the review — werf squashes, so it becomes the commit body. Audit its claims against your findings: a safety property asserted there that a finding contradicts is itself a defect, and inline comments anchored to code lines never reach whoever reads it.

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

| Criteria | Inferred? | Met? | Evidence |
| :--- | :--- | :--- | :--- |
| [criterion] | yes/no | ✅/⚠️/❌ | file:line |

## Issues

- **Critical** — blocking, with file:line
- **Major** — significant concern
- **Minor** — suggestion

## Risks

Sorted by severity, then by likelihood.

| № | Risk | Type | Likelihood | Severity | Location | Circumstances | Consequences | Recommendation |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |

## Not verified

- What was not built, run, or reachable — and why (Buildah paths do not compile on macOS, e2e needs Linux with kind).
```

- **Recommendation** — the concrete action, with file:line references.

## Language

Headers in English, everything else in the user's language.
