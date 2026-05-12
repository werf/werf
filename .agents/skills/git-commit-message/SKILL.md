---
name: git-commit-message
description: Generates git commit messages according to werf conventions. Use after staging changes.
---

# Git Commit Message Generation

## Instructions

1. Read types, scopes, subject rules, and body rules from `CONTRIBUTING.md#conventions`.
2. Analyze `git diff --cached` to determine the primary type and scope.
3. Use the following format:

   ```
   <type>(<scope>): <subject>

   <body>
   ```

4. Follow these rules:
   - **Header:** ≤ 72 characters. Nested scopes are allowed.
   - **Subject:** imperative, lower-case, no trailing period.
   - **Body:** imperative, include motivation for the change, contrast with previous behavior.
5. Output ONLY the commit message (header + body), with no additional text, quotes, or formatting.