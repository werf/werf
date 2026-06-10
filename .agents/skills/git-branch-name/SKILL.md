---
name: git-branch-name
description: Generates git branch names according to werf conventions. Use when creating a branch for a new task.
---

# Git Branch Name Generation

## Instructions

1. Read types, scopes, and naming rules from `CONTRIBUTING.md#conventions`.
2. Write the branch name in the following format:

   ```
   <type>/<scope>/<short-description>
   ```

3. Use **top-level scope only** — nested scopes are NOT allowed in branch names.
4. Keep the total length ≤ 50 characters.
5. Output ONLY the branch name, with no additional text, quotes, or formatting.
