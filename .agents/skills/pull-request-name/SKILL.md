---
name: pull-request-name
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
5. Nested scopes are allowed and encouraged.

## Description

Use the following structure. Every section is **mandatory** — omit a section only when it genuinely does not apply (e.g. single-line typo fix may skip *Review focus / risks*).

```
## Summary

<1-3 sentence high-level overview of what the PR does and why it exists.>

## Key changes

- <concrete change 1>
- <concrete change 2>
- …

## Why

<Motivation: what problem this solves, what maintenance/UX/perf gain it brings.>

## Review focus / risks

- <area or file that deserves careful review>
- <potential risk or side-effect>
```

### Rules

- Language: match the project's primary language (English by default).
- Be specific: name files, modules, functions — not "updated some code".
- *Key changes*: group related items; use sub-bullets for detail when helpful.
- *Why*: explain the reason, not what changed (that's *Key changes*).
- *Review focus / risks*: guide the reviewer — call out non-obvious consequences, large generated diffs, breaking changes.
- No AI-slop filler ("This PR improves the codebase…"). Every sentence must carry information.

## Output

When generating only the title (e.g. for `gh pr edit --title`), output ONLY the title, with no additional text, quotes, or formatting.