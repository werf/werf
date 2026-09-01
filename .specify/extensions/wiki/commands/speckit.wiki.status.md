---
description: "Compact wiki snapshot: counts, freshness, open lint issues, and one recommended next action"
---

# Wiki Status

Read-only snapshot of the wiki's state, ending in exactly **one** recommended
next action. This is the session-resume entry point: because the wiki lives
in files, a brand-new agent session rebuilds its working picture from this
command in one step — nothing depends on any previous conversation.

## User Input

```text
$ARGUMENTS
```

Optional: a page type (`decision`, `concept`, …) to list that type's pages,
or `full` for larger slices.

## Steps

### 1. Resolve the wiki

Load configuration exactly as `/speckit.wiki.init` does. If
`WIKI_DIR/SCHEMA.md` does not exist, report that no wiki exists yet and
recommend `/speckit.wiki.init <scope>` — that is the whole output.

### 2. Load slices, not files

Read `SCHEMA.md` (scope section only), `INDEX.md`, `sources.md`, and
`lint-report.md` if present. Do **not** read the pages themselves — the
snapshot is assembled from the index and registries.

### 3. Render the snapshot

- **Scope** — the wiki's mission sentence(s).
- **Pages** — count per type, and the most recently updated 5 (from INDEX +
  frontmatter dates when cheaply available).
- **Sources** — total registered; the 3 most recent ingests with dates.
- **Freshness** — sources re-ingested more recently than their dependent
  pages; pages older than `lint.stale_after_days`.
- **Open issues** — top rows from `lint-report.md` by severity
  (semantic > structural > mechanical), or "never linted" if absent.

### 4. Recommend exactly one next action

Priority order — pick the first that applies:

1. Unresolved conflict in the lint report → resolve it (name the page).
2. Never linted, or last lint older than the oldest ingest since → `/speckit.wiki.lint`.
3. No pages yet → `/speckit.wiki.ingest` (suggest the newest feature's `research.md`).
4. A feature shipped since the last ingest → ingest its artifacts.
5. Otherwise → `/speckit.wiki.query <the scope's most central open question>`
   to probe coverage.

## Guardrails

- **Writes nothing, ever.**
- Fits on one screen for the default invocation; `full` may triple the
  slices but still never loads page bodies.
- The recommendation must name a concrete command with concrete arguments —
  "consider maintaining the wiki" is not a recommendation.
