---
name: session-retro
description: "Analyze the current session for harness-worthy lessons — repeated corrections, discovered conventions, skill bugs — and turn them into concrete repo changes: docs, skills, task targets, linter rules, CI checks. Use at the end of a session, when asked to reflect/retro, or when invoked as /session-retro."
---

# Session Retro

A session carries signal about how the harness should behave next time: corrections, discovered
conventions, skills that turned out wrong, workflow friction. It evaporates when the conversation
ends unless something writes it down. This skill is that step.

The output is a *harness* change — a change to whatever tells the next session how to work. Often
that is prose (`AGENTS.md`, `CODESTYLE.md`, `CONTRIBUTING.md`, a skill under `.agents/skills`), but
it is just as legitimately a `Taskfile.dist.yaml` target, a `.golangci.yml` rule, a CI check, an
issue/PR template, or a `docs/` page. Pick the landing spot from the finding, not from a list.

## 1. Scan the session

- **Corrections**: the user said "no", "not like that", or redid your work. What rule would have prevented it?
- **Repeated explanations**: anything explained more than once — the current instructions don't cover it.
- **Discovered conventions**: facts that came from the user or from reading the repo, not from any doc — build quirks, naming schemes, "we always do X here".
- **Skill bugs found in use**: a skill that gave wrong guidance, missed a step, or referenced a stale path or command.
- **Workflow decisions**: the user picked one approach over another ("always do it this way from now on") — a durable preference, not a one-off call.
- **Friction**: a check that was slow, awkward, or easy to forget, or a step done by hand that a `task` target could do.
- **Near-misses**: something caught just before landing (wrong branch, guessed path, unverified assumption) — the cheapest lesson, the cost is already paid.
- **Context waste**: tool calls that cost a lot of context for little return — a whole-file read where a grep would do, the same file read twice, a raw `task build`/`task test` log pasted in full, a subagent spawned for a one-line lookup, or a wide search done inline instead of delegated. A generic `bash` call where a purpose-built tool exists is the expensive one. Each maps to a rule.
- **Skill usage**: which skills actually fired, and whether it mattered. A skill whose body was never opened during the work it governs guided nothing, however apt its name.
- **Prompt friction**: the request as posed cost tokens — a file the user already knew and you hunted for, a constraint revealed after you built the wrong thing, work redone that one clarifying question would have prevented. If a question you should have asked would have caught it, that is a rule for you; otherwise it is feedback for the user, not a file.

Those three are denominated in tokens, so measure them instead of recalling them: per-message usage lives in the session transcript (`~/.pi/agent/sessions/<cwd-slug>/*.jsonl`, or `~/.claude/projects/<cwd-slug>/*.jsonl`). Aggregate inside the analysis script and print only the top consumers — dumping per-message rows into the conversation costs more than the finding is worth. Match skills by **file path** (`skills/<name>/SKILL.md`), never the bare name, and keep frontmatter-only, body-read, and edited apart. With no transcript available, drop these cuts rather than estimating from memory.

Ignore one-off task specifics, anything an existing doc, skill, or check already covers, and raw totals with no attributable cause — a per-tool call-count table is not a finding.

## 2. Classify each finding

Ask two questions: *who needs to know this*, and *can a machine enforce it instead of a human
remembering it?*

| Finding is about… | Goes into |
|---|---|
| A rule a tool can check | A `task` target, a `.golangci.yml` rule, or a CI check — ALWAYS prefer this over a sentence asking someone to remember |
| How to work in this repo — commands, verification, scope discipline | `AGENTS.md` |
| Go design or naming convention | `CODESTYLE.md` |
| Commit/branch/PR types and scopes | `CONTRIBUTING.md` |
| A whole reusable procedure | A skill in `.agents/skills` |
| Behavior a werf user hits, not an agent | `docs/`, or command help text (then `task doc:gen`) |
| One-off, won't recur | Nothing |

Search before writing: most findings are a missing line in something that already exists, not a new
file. Extend the closest existing skill or section rather than creating a near-duplicate.

## 3. Draft the change

- Make the smallest edit that closes the gap; don't rewrite unrelated sections.
- Match the file's existing tone — `AGENTS.md` and `CODESTYLE.md` are terse, imperative, bulleted.
- A rule belongs in exactly one place. Cross-reference instead of duplicating; a copied rule drifts.
- Don't restate in prose what a `task` command or linter already enforces.
- When a tool has become self-documenting — its own descriptions now carry the contract — TRIM the skill that taught workarounds for it instead of layering notes on top. A skill keeps only what the tool cannot know: local conventions and domain norms.

## 4. Confirm before applying

- Low-risk (typo, stale path or command, clarifying a sentence): apply directly.
- A new mandatory rule, a changed workflow, a new skill, or any change to tooling and CI: state the proposed wording and rationale, and confirm before writing.

## 5. Apply through the repo's own conventions

Harness files are versioned like code: branch and open a PR per `git-conventions` and
`pull-request`. NEVER push to a release branch (`main`, `3`, `2`, `1.2`) directly. A change to
`Taskfile.dist.yaml`, `.golangci.yml`, or CI must be run once before it is proposed — an unverified
check is worse than none.

## 6. Report

List what changed, file by file, and where each finding was routed — including the ones dropped as
one-off, so nothing is silently omitted. Report skill usage as a short table — skill, tier reached,
body reads, and one line of impact: what it changed, or what it would have prevented. Report
context-waste and prompt-friction findings with their measured cost and the turn they came from,
even when they produce no file.
