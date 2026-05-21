---
name: pull-request-name
description: Generates Pull Request titles according to werf conventions. Use when creating or updating a PR.
---

# Pull Request Naming

## Instructions

1. Read types, scopes, and formatting rules from `CONTRIBUTING.md#conventions`.
2. The PR title should mirror the header of the main commit in the PR.
3. Format the title as `<type>(<scope>): <subject>`.
4. Keep the total length ≤ 72 characters.
5. Nested scopes are allowed and encouraged.
6. Output ONLY the title, with no additional text, quotes, or formatting.