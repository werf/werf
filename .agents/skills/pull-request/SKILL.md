---
name: pull-request
description: Generates Pull Request titles and descriptions according to werf conventions. Use when creating or updating a PR.
---

# Pull Request Conventions

## Defaults

- Always create PRs as draft (`gh pr create --draft`). The author marks it ready for review manually.
- When a pushed commit changes what the PR does, update the title and description in the same step. A description that describes an earlier state of the branch is what the reviewer reads.

## Title

1. Read types, scopes, and formatting rules from `CONTRIBUTING.md#conventions`.
2. The PR title should mirror the header of the main commit in the PR.
3. Format the title as `<type>(<scope>): <subject>`.
4. Keep the total length ≤ 72 characters.
5. Nested scopes are allowed and encouraged, comma-separated: `fix(build, stapel, import): …`.
6. Subject: imperative present tense, no capitalized first letter, no trailing dot.

## Description

Use the following structure. *Summary*, *Why* and *Verification* are **mandatory**. Omit *Key changes* when it would only restate *Summary*, and *Review focus / risks* when there is genuinely nothing to guide the reviewer to. NEVER rename or substitute the sections.

```
## Summary

<1-3 sentence high-level overview of what the PR does and why it exists. For a `fix`, lead with the
observed wrong behavior and add a minimal repro (Dockerfile / werf.yaml / command) plus expected vs actual.>

## Key changes

- <concrete change 1>
- <concrete change 2>
- …

## Why

<Motivation: what problem this solves, what maintenance/UX/perf gain it brings.>

## Verification

- <manual or hand-run e2e check, and the environment it needed>
- <what could not be run>

## Review focus / risks

- <area or file that deserves careful review>
- <potential risk or side-effect>
```

### Rules

- Language: English.
- Be specific about behavior, not "updated some code". Lead each bullet with what changed; add a path only when it helps navigation — a new file, a non-obvious location, or when the point is that the change plumbs through several layers. The Files tab already lists paths.
- *Key changes*: group related items by theme, not one bullet per file; use sub-bullets for detail when helpful.
- *Why*: explain the reason, not what changed (that's *Key changes*).
- *Verification*: only the delta over CI — manual runs, hand-run e2e/real-cluster checks, the environment they required, and what could NOT be run. CI builds and runs the whole suite, so `task build`/`task test:unit` are noise unless a scoped local run is itself the point — then name it by its `task` command, never raw `go test`/`go vet`/`gofmt`. Plain list, not checkboxes — the reviewer is not the one ticking them.
- *Review focus / risks*: guide the reviewer — call out non-obvious consequences, large generated diffs, breaking changes, and any deliberately accepted limitation of the chosen approach.
- No AI-slop filler ("This PR improves the codebase…"). Every sentence must carry information.
- NEVER include sensitive or customer-identifying details in the title or description: client/company names, internal hostnames or filesystem paths, private build tags or version suffixes, credentials. Describe the environment generically (e.g. "long-lived CI runners sharing WERF_HOME", not a customer's runner path).

## Output

When generating only the title (e.g. for `gh pr edit --title`), output ONLY the title, with no additional text, quotes, or formatting.