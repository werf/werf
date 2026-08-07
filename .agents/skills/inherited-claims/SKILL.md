---
name: inherited-claims
description: How to treat conclusions inherited from an earlier session — handover notes, prepared comments, verdict files, plans. Use whenever a task starts from context you did not derive yourself, or resumes work described by a previous session.
---

# Inherited Claims

A conclusion carried in from somewhere else is a claim, not a finding — no matter how confident its wording or how detailed its citations. Handover notes, prepared comment texts, verdict files, plans, an issue's own body, a previous agent's summary: all of them were true about a codebase that may no longer exist.

The repo itself carries claims too: a merged PR's title and body, a changelog entry, a migration guide, a release note. Those describe what someone intended to do, not necessarily what shipped.

## Rules

- Re-derive every inherited claim from the primary source — the code, the API, the config, the actual command output — before it reaches anywhere a human will read it: a tracker comment, a PR description, a card, a message to the user.
- "X was removed/changed in version N" is a claim even when a merged PR title and the shipped documentation both say it. Check the call site at the released ref — `git show <tag>:<path> | grep <symbol>` — before ruling a behavior out or in. Removing the flag that *disables* a check is not removing the check; a PR can land its own description's headline only in part, and the guide written from that description then documents behavior the code never had (paid in full: a check documented as removed in a migration guide, still executing in the released binary, blamed for nothing until its cost showed up as a 3.6-hour rebuild).
- Re-derive the *reason*, not just the verdict. "Close this as obsolete" may be right while the sentence explaining why is invented; publishing the invented sentence is the damage.
- When an inherited note states a convention and a skill covers the same ground, the skill wins. Load the skill and follow it; do not act on the remembered version.
- Quote what you verified with a `file:line`, a flag default, a response body. If a claim cannot be checked cheaply, drop it from the text rather than softening it — a hedge still reads as an assertion.
- Batching is where this fails. A pile of pre-written items feels like work already done; it is a pile of unverified claims, and approving them as a batch means publishing them as a batch.

## Why this matters more than it looks

Being wrong in public, in the user's name, costs more than having nothing to say. A stale claim in a tracker is read as the maintainer's position, gets quoted back, and outlives the session that produced it.
