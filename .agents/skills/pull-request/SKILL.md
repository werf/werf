---
name: pull-request
description: Generates Pull Request titles and descriptions according to werf conventions. Use when creating or updating a PR.
---

# Pull Request Conventions

The description is the **spec** of the change: a reviewer who never opens the diff must be able to say what werf does differently after it, where it bites, and what was actually proven. The diff is the evidence, not the source.

Everything only true during the review — what you ran, where to look, what to do after the merge — goes in a comment. werf squashes, so the description becomes the commit body and outlives the review by years, while a comment never can: the kernel's `---` line, with GitHub doing the cutting. The `(#NNNN)` in the subject keeps the comment reachable from `git log`.

## Defaults

- Create PRs as draft (`gh pr create --draft`) and leave them draft. Only the user marks a PR ready.
- Post the review-time part as the first comment (`gh pr comment`), never in the description. Nothing to say — skip it; a comment holding only a mutation line is still required.
- When a push changes what the PR does, its evidence or its follow-up work, update the title, the description and that comment in the same step.

## Title

`<type>(<scope>): <subject>`, ≤ 72 characters, mirroring the header of the main commit. Types, scopes and subject rules come from `git-conventions` and `CONTRIBUTING.md#conventions`, including the rule that a `feat`/`fix` subject names the user-visible outcome, not the mechanism. Nested scopes comma-separated from the broadest: `fix(build, stapel, import): …`. For a dependency bump the title states werf's outcome, never the upstream changelog subject.

## Description

```
## Summary

<What werf does differently now, at most 3 sentences. For a `fix`: the observed wrong behavior,
plus a pasteable repro (command / werf.yaml / Dockerfile) when it reproduces from a clean
checkout, otherwise the precondition it needs — a race, a pre-existing host state. Never a
fabricated repro. For a `feat`: what the user can now do and the workflow that needed it.>

## What

- <One falsifiable behavior claim per line, with the condition that triggers it: "under umask 001
  the service script is mode 0755", never "handles umask correctly".>
- <Every user-visible surface added or changed, with its default: flag, annotation, env var,
  werf.yaml field, log or error text, exit code.>
- <BREAKING: a claim that breaks an existing setup names who it breaks and the way out — "every
  host without netavark fails at backend init; CONTAINERS_CONF_OVERRIDE cannot bring CNI back".>
- <VERIFIED: a check CI cannot repeat — hand-run, host-level, offline, against a cluster — named
  on the claim it settles. Anything CI does is not worth the line.>
- <UNVERIFIED: a claim nothing stands behind says so, and says what would settle it.>
- <What deliberately does NOT change, where a reader would expect it to.>

## Why

<The root cause, and what leaving it alone costs. Then the rejected alternative, when there is one
someone would argue for — "a repo-wide marker instead of a per-project one" is an alternative,
"buildah exposes no typed error" is the diff. Never a reworded Summary or claim list.>
```

- *Summary* and *Why* are the pair that collides, because for a `fix` the symptom and its cause are neighbours in one chain: what is visible from outside goes in *Summary*, why the code did that and what leaving it costs goes in *Why*. When the symptom cannot be named without its mechanism — a stale cache entry, a missing host binary — name it in *Summary* anyway.
- Every risk is a claim; *Review focus* is for where to look, not for what breaks.
- Breaking claims go first, in a `### Breaking` group, or lead their line with `BREAKING:` when there is only one. `VERIFIED:` and `UNVERIFIED:` stay inside the claim they qualify and are never grouped — a claim that both breaks and is unproven is the one a group would tear in half. Capitals, never bold or emoji — `git log` shows both literally, and only a word greps.
- A change breaking in the conventional-commits sense also needs the `BREAKING CHANGE:` footer, which release-please turns into a major version bump. Never add it on your own — propose it, the bump is the user's call.
- A claim is one short line and the list is one level deep. A qualifier that will not fit — the way out, the mechanism a claim's safety rests on, `UNVERIFIED:` — becomes the next line of the same group. Group by user workflow or surface, never by file, under `###` headings once *What* passes eight claims.
- Behavior is what is observable from outside: a command's output, an exit code, a file's mode, a resource in the cluster, a rebuild that no longer happens. Function names, refactored call paths and file lists belong to the diff. Debug output emitted incidentally is not a surface; when diagnostic output is the deliverable it is — then name the switch that turns it on, the lines it adds and where they appear, and the invariant that nothing is emitted or measured without it.
- With no user-visible behavior at all, *What* states the invariant: what must not change, and what a reviewer checks it against. For a build or packaging change that is a developer-visible contract — build tags, embedded assets, `task` targets. For a speed change it is the work that no longer happens plus the workload the numbers came from; a wall-clock figure alone is not a claim.
- Length follows the behavior surface, never the diff. A *What* that will not fit is a PR that should be split, not a description to trim.
- These heading names, verbatim. A section with nothing to say is omitted, never renamed or padded. For one obvious change — nothing a reader would observe, and nothing a `BREAKING:`, `VERIFIED:` or `UNVERIFIED:` marker would qualify — drop the headings: one to three sentences, a marker leading the sentence it qualifies. A typo or a wording fix qualifies; a change to a published artifact, to what a pipeline emits, or to an instruction an agent follows does not, and a change needing `BREAKING:` never does.
- A bump of `werf/nelm`, `werf/kubedog` or `werf/common-go` claims every user-visible change between the old pin and the new one, read from the commit range and never from the target's release notes — a pseudo-version pin sits mid-release, so the range crosses commits those notes never mention. One `gh api repos/<owner>/<repo>/compare/<old>...<new>` answers it in a second; a bump is trivial only after that range comes back with no user-visible change, and then the sentences name the range and say so.
- An edit under `.agents/skills`, `AGENTS.md` or `CODESTYLE.md` is trivial only when it cannot change what an agent produces: a typo, a broken link, reflowed text. A changed, added or removed instruction takes the full form, and *What* states the delta in the artifact the agent generates.
- A bot-generated description — release-please, dependabot, renovate — is the artifact and is regenerated on every push. Never rewrite it; what a human has to add goes in the comment.
- When the behavior surface already has a versioned home in this repo — a migration guide, a reference page — *What* names that document and adds only what is not in it, unless this PR is the change to that document — then *What* states what changes for whoever follows it and names the file as the authority for the rest, because an index of the edits is not a spec. Check the document really covers the claims before delegating to it; a guide that lags the branch turns the delegation into a false statement.
- The description is self-contained: someone with no access to the issue, the discussion or the diff must be able to follow it. Link the issue and summarize what a claim rests on, no more.
- The only code block worth its space is a repro the reader can paste, never a transcript of output that repro already produces. An error message users will search for is quoted inline, in the claim that says when they see it.
- English. No speculation, no "this PR improves the codebase".
- `Fixes`/`Closes` keywords live in the description — GitHub does not auto-close from a comment. A reference to `werf/nelm`, `werf/kubedog` or `werf/common-go` does not auto-close at all: link it, and put closing it by hand in the comment's *Follow-up*. When a known commit introduced the bug, pin it: `Fixes: <12+ chars of sha> ("the commit subject")`.
- Sensitive details are governed by `git-conventions` and apply here unchanged, plus: a public project may be named as the workload that motivated a change, a customer may not.

## Self-check against the diff

Before handing the PR over to the user, map it both ways:

- every behavior-changing hunk lands on a claim in *What*, and a supporting hunk (test, doc, error wrapping, refactor) lands on the claim it serves — an unmapped hunk is scope creep or a missing claim;
- every claim lands on code — a claim with nothing behind it is unfinished work. For a documentation change this map runs against the existing code that makes each claim true.

A branch too large to walk hunk by hunk is mapped commit by commit, and the comment says so.

## The comment

```
## Verification

- <manual or hand-run check CI cannot make, and the environment it needed>
- Mutation: <what was broken in the code> → <test that failed>.
- Not run: <check that would have covered a claim, and why it could not run>

## Review focus

- <where to look: an area that deserves care, a large generated diff>

## Follow-up

- [ ] <action outside this diff, naming where it happens (`owner/repo`, a file, a command)>
- [ ] BLOCKER: <the same, but it has to land before this PR is merged>
```

- These heading names, verbatim. A block with nothing to say is left out, and a comment with nothing in any block is not posted. The comment never stands in for a missing description: a fact that belongs in *What* means *What* has to be written.
- *Verification* is the mechanics: the command, the environment, what it showed. The attestation itself — that a check CI cannot repeat was made — rides on the claim as `VERIFIED:`, and this block is where it is spelled out; never the same sentence twice. Only the delta over CI. CI builds and runs the whole suite, so `task build`/`task test:unit` are noise unless a scoped local run is itself the point. Which claims are unproven belongs to the claims, never written twice.
- A new or changed test needs its mutation named. When the mutation was not run, say so and name the one to try (`test-the-tests`).
- A `Not run:` line names the check and what stopped it; the claim itself says it is unproven. If the line would only restate the claim, drop it. It earns its place only when the missing check leaves a claim unverified.
- *Review focus* never speculates — "probably fine because …" is not a risk, drop the line. A safety property the author could not establish stays here as a question — "confirm a cached git stage cannot resurrect content after a force-push" — because as a claim it asserts the very safety nobody checked.
- *Follow-up*: checkboxes, one action per line, everything this diff cannot contain — after the merge, or before it with a leading `BLOCKER:` for a dependency that has to land first. Mark what must not be skipped; when a release note is the only mitigation for a breaking change, say so on the line. No speculative and no already-done items. Public repositories only: a private harness or any path under `~` is never a line in the PR.

## Output

When generating only the title (e.g. for `gh pr edit --title`), output ONLY the title, with no additional text, quotes, or formatting.
