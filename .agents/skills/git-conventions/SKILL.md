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
- **Header:** ≤ 72 characters. Nested scopes are allowed, comma-separated: `fix(build, stapel, import): …`.
- **Subject:** imperative, lower-case, no trailing period.
- **Subject of a `feat`/`fix`:** the observed outcome, never the mechanism. `CHANGELOG.md` is generated from these subjects verbatim, so the reader is a werf user who has never seen the code: name the symptom that goes away or what becomes possible, and leave function names, internal identifiers and the how to the body. `make service script executable regardless of umask` describes the patch; `stop stapel builds failing under a custom umask` describes what the user hit. For `refactor`/`test`/`chore` the reader is a developer and the mechanism is the right subject.
- **Body:** imperative; state the motivation for the change and contrast it with previous behavior. This is where the mechanism and the root cause go.
- NEVER include sensitive or customer-identifying details: client/company names, internal hostnames or filesystem paths, private build tags or version suffixes, credentials. Describe environments generically.

## Before starting work

- Create the topic branch BEFORE the first commit, not before the push. `main`, `3`, `2`, `1.2` are release branches: a commit landed on one has to be moved by hand afterwards and its message amended along with it, and until someone asks for a PR nothing reveals it is on the wrong branch.
- `git fetch` and compare your base against `origin/<base>` BEFORE writing code, not at push time. A moved base can have refactored the very file you are about to edit, and the whole diff then has to be re-ported by hand during the rebase.

## Before staging

- Check `git status` for unrelated untracked files before staging. This worktree carries local working files (`.dev/`, scratch notes, orchestrator state), so prefer explicit paths over `git add -A` — a blanket add sweeps them in, and untracking later costs an extra commit. An orchestrator or helper commit command stages broadly — inspect `git status` BEFORE invoking it, not after.

## Merging a branch into another

- Create the branch that will carry the merge BEFORE running `git merge`, never after resolving the conflicts. Switching branches with `git checkout -b` while a merge is in progress silently discards `MERGE_HEAD` once the index is clean, and the following `git commit` records a single-parent commit: the content is merged, the ancestry is not, so every PR based on the target still sees the source's commits as unmerged and conflicts on the same files again.
- Before pushing a merge, ALWAYS check it has two parents: `git log -1 --format=%p` must print two hashes, and `git merge-base --is-ancestor origin/<source> HEAD` must succeed. A merge message proves nothing — it is just text.
- To repair ancestry when a reviewed merge result is already on the target, merge the source again with `-s ours` and verify `git rev-parse HEAD^{tree}` is unchanged. Re-running a plain `git merge` would resurrect what the resolution deliberately dropped.

## Before pushing

- ALWAYS check the current branch (`git branch --show-current`) — a stale local `main` or someone else's WIP branch is easy to miss.
- NEVER push to a release branch (`main`, `3`, `2`, `1.2`) directly — branch from the current `origin/<base>` and open a PR.
- When a commit carries recorded fixtures or logs, scrub every class of identifier — uuids, numeric ids, initials, avatar urls — not just names and emails. A name-only pass leaves pseudonymous ids that map back to people through internal tables, and rewriting history afterwards is the expensive path.
- If a push is rejected, don't retry with force. Find out why the ref diverged first: `--force-with-lease` is also rejected as `stale info` when there is no remote-tracking ref for the branch (e.g. after pushing by URL) — that is a missing lease baseline, not a diverged history, and the fix is `--force-with-lease=<ref>:<sha>`.

## Output

Output ONLY the branch name or the commit message, with no additional text, quotes, or formatting.
