---
name: git-conventions
description: werf conventions for branch names and commit messages. Use when creating a branch for a new task or committing staged changes.
---

# Git Conventions

Types and scopes are defined in `CONTRIBUTING.md#conventions` — that file is the source of truth, read it before composing either.

## Branch name

```
<type>/<scope>/<short-description>
```

- **Top-level scope only** — nested scopes are NOT allowed in branch names.
- `<short-description>`: kebab-case, concise.
- Total length ≤ 50 characters.

## Commit message

```
<type>(<scope>): <subject>

<body>
```

- Determine type and scope from `git diff --cached`.
- **Header:** ≤ 72 characters. Nested scopes are allowed.
- **Subject:** imperative, lower-case, no trailing period.
- **Body:** imperative; state the motivation for the change and contrast it with previous behavior.

## Output

Output ONLY the branch name or the commit message, with no additional text, quotes, or formatting.
