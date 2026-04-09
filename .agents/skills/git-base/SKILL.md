---
name: git-base
description: Base git conventions for werf repository. Defines shared types, scopes, and formatting rules for commits, branches, and PRs.
---

# Base Git Conventions

This file contains the shared conventions for git operations in the werf repository. All git-related skills MUST adhere to these rules.

## Core Header Components

### Type
Must be exactly one of:
- **feat**: new features or capabilities.
- **fix**: bug fixes.
- **refactor**: code changes that neither fix a bug nor add a feature.
- **docs**: documentation updates or improvements.
- **test**: additions or corrections to tests.
- **chore**: updates to CI, dependencies, or development workflow.

### Scope
Scopes categorize the area of the project affected.

**End-user functionality scopes (can be nested):**
- `giterminism`
- `build` (nested: `stapel`, `dockerfile`, `docker`, `buildah`, `tagging`, `stages`)
- `deploy` (nested: `values`, `dependencies`, `secrets`, `templates`, `tracking`, `resource-order`, `resource-lifecycle`, `plan`)
- `bundle`, `cleanup`, `host-cleanup`, `run`, `kube-run`, `compose`, `ci-env`

**Development and Maintenance scopes:**
- `ci`, `release`, `dev`, `deps`

### Subject / Short Description
- Use the **imperative, present tense**: "change" not "changed" nor "changes".
- **Lower case**: do not capitalize the first letter.
- **No period**: do not end with a dot (`.`).

## Formatting Rules

| Context | Format | Scope Rule |
|---------|--------|------------|
| Commit | `<type>(<scope>): <subject>` | Nested scopes allowed (e.g., `build/stapel`) |
| PR Title | `<type>(<scope>): <subject>` | Nested scopes allowed |
| Branch | `<type>/<scope>/<short-description>` | **Top-level scope only** (e.g., `build`) |

## Body (for Commits)
- Separate from header by a blank line.
- Use imperative, present tense.
- Include **motivation** for the change.
- **Contrast** with previous behavior.