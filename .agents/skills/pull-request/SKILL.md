---
name: pull-request
description: Generates Pull Request titles and descriptions according to werf conventions. Use when creating or updating a PR.
---

# Pull Request Conventions

## Defaults

- Always create PRs as draft (`gh pr create --draft`). The author marks it ready for review manually.

## Title

1. Read types, scopes, and formatting rules from `CONTRIBUTING.md#conventions`.
2. The PR title should mirror the header of the main commit in the PR.
3. Format the title as `<type>(<scope>): <subject>`.
4. Keep the total length ≤ 72 characters.
5. Nested scopes are allowed and encouraged, comma-separated: `fix(build, stapel, import): …`.
6. Subject: imperative present tense, no capitalized first letter, no trailing dot.

## Description

Match the description to the size of the diff. Two tiers, nothing in between.

**Small** — one logical change, under ~20 changed lines, no user-visible behavior change (test fix, typo, non-portable flag, dependency bump): no headings at all. One to three sentences — what was wrong, what it is now. If the title already says it, one sentence is the whole description.

**Everything else** — the structure below. A section earns its place only by telling the reviewer something the title and the diff do not. A section written to complete the template is slop. NEVER rename or substitute the sections.

```
## Summary

<1-3 sentence high-level overview of what the PR does. For a `fix`, lead with the
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
- *Summary* and *Why* answer different questions and NEVER restate each other. *Summary*: the observed behavior and what the PR does about it. *Why*: the root cause, and what leaving it alone costs — never a reworded list of the changes (that's *Key changes*). For the umask fix: *Summary* was "mode `0667` loses its only executable bit to the umask, so `execve` fails", *Why* was "the old mode worked by luck — the common umasks 022 and 002 don't touch the last digit".
- *Verification*: only the delta over CI — manual runs, hand-run e2e/real-cluster checks, the environment they required, and what could NOT be run. CI builds and runs the whole suite, so `task build`/`task test:unit` are noise unless a scoped local run is itself the point — then name it by its `task` command, never raw `go test`/`go vet`/`gofmt`. Plain list, not checkboxes — the reviewer is not the one ticking them.
- *Review focus / risks*: guide the reviewer — call out non-obvious consequences, large generated diffs, breaking changes, and any deliberately accepted limitation of the chosen approach.
- No AI-slop filler ("This PR improves the codebase…"). Every sentence must carry information.
- Length is a budget, not a target: under ~300 characters for a small PR, under ~1500 for a normal one. Cut any sentence that restates the title, another section, or the diff.
- The only code block worth its space is a repro the reviewer can paste (Dockerfile, werf.yaml, command). Never paste a terminal transcript of an error that the repro already produces.
- NEVER speculate in *Review focus / risks*. List what you know. "Probably fine because …" is filler — drop the whole bullet.
- NEVER include sensitive or customer-identifying details in the title or description: client/company names, internal hostnames or filesystem paths, private build tags or version suffixes, credentials. Describe the environment generically (e.g. "long-lived CI runners sharing WERF_HOME", not a customer's runner path).

## Output

When generating only the title (e.g. for `gh pr edit --title`), output ONLY the title, with no additional text, quotes, or formatting.