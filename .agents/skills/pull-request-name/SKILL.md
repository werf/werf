---
name: pull-request-name
description: Generates Pull Request titles following werf conventions. Use when creating or updating a Pull Request to ensure the title matches the primary commit header.
---

# Pull Request Naming

This skill provides instructions for generating Pull Request (PR) titles that strictly adhere to werf's contribution guidelines.

## Instructions

Each pull request title must clearly reflect the changes introduced and adhere to the shared header format. In most cases, it should mirror the header of the main commit in the PR.

### Constraints

- Total length must not exceed 72 characters.
- Output ONLY the Pull Request title, with no additional text, quotes, or formatting.

### Format

```
<type>(<scope>): <subject>
```

- **Type**: See `Types` in [../git-base/SKILL.md](../git-base/SKILL.md).
- **Scope**: Use specific scopes from [../git-base/SKILL.md](../git-base/SKILL.md). Nested scopes are allowed and encouraged for precision (e.g., `deploy/secrets`).
- **Subject**: Follow the imperative, lower-case, no-period rules defined in [../git-base/SKILL.md](../git-base/SKILL.md).

## Examples

- `feat(deploy/plan): add support for dry-run in specific namespaces`
- `fix(build/dockerfile): resolve caching issue with multi-stage builds`
- `docs(cleanup): clarify host-cleanup policy in readme`

## Workflow

1. Identify the primary `type` and most specific `scope` of the changes.
2. Formulate a `subject` that summarizes the PR's impact.
3. Ensure the title matches the header of the primary commit in the PR.
