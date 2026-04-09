name: git-branch-name
description: Generates git branch names according to werf's contribution guidelines. Use when starting a new task, creating a branch for a fix, feature, or refactor.
---

# Git Branch Name Generation

This skill provides instructions for generating git branch names that strictly adhere to werf's contribution guidelines based on `CONTRIBUTING.md`.

## Instructions

When generating a branch name, you MUST use the following format:

- Total length must not exceed 50 characters.
- Output ONLY the branch name, with no additional text, quotes, or formatting.

```
<type>/<scope>/<short-description>
```

### Constraints from GIT_BASE

Refer to [Base Git Conventions](../git-base/SKILL.md) for core definitions.

- **Type**: Must be one of the types defined in `git-base` (e.g., `feat`, `fix`, `refactor`, `docs`, `test`, `chore`).
- **Scope**: Use ONLY the **top-level scope**. Contrary to commits and PRs, nested scopes are NOT allowed in branch names (e.g., use `build`, not `build/stapel`).
- **Short Description**: Use a concise, hyphen-separated phrase in `kebab-case`.

## Examples

### Good
- `feat/deploy/azure-kv-support`
- `fix/build/buildah-storage-limit`
- `docs/giterminism/fix-typos`

### Bad
- `feat/deploy/secrets/add-vault` (Reason: nested scope `deploy/secrets` is not allowed in branch names)
- `fix/Bug_In_Engine` (Reason: not kebab-case, contains uppercase)
- `new-feature-branch` (Reason: missing type and scope)

## Workflow

1. Identify the primary `type` of work being done from `git-base`.
2. Select the relevant top-level `scope` (do not use nesting).
3. Create a short `kebab-case` description of the task.
4. Join the components with forward slashes: `type/scope/description`.