---
name: git-commit-message
description: Generates git commit messages formatted according to werf's contribution guidelines. Includes header with type/scope and a descriptive body. Use this skill when you need to create a commit message for staged changes.
---

# Skill: Git Commit Generation

This skill provides instructions for generating git commit messages that strictly adhere to werf's contribution guidelines.

## Instructions

When generating a commit message, follow the mandatory format and rules defined in [git-base](../git-base/SKILL.md).

- Header length must not exceed 72 characters.
- Output ONLY the commit message (header and body), with no additional text, quotes, or formatting.

### Format

```
<type>(<scope>): <subject>

<body>
```

Use the types, scopes (including nested), and subject rules from the base conventions.

### Body Requirements

As defined in the base conventions, the body MUST:
- Be separated from the header with a blank line.
- Use imperative, present tense.
- Include the **motivation** for the change.
- **Contrast** the change with previous behavior.

## Examples

### Good
```
feat(deploy/secrets): add support for external azure key vault

The previous implementation only supported local encrypted files. This change 
introduces a new secret manager that fetches values directly from Azure KV
to improve security in enterprise environments.
```

### Bad
```
Fix(deploy): Fixed the bug.
```
*(Reasons: capitalized type, past tense subject, no body, period at the end)*

## Workflow

1. Analyze `git diff --cached`.
2. Identify the primary `type` and `scope` based on affected files and logic (refer to [git-base](../git-base/SKILL.md)).
3. Formulate a `subject` line in lower-case imperative.
4. Write a `body` explaining "why" the change was made and what it replaces.