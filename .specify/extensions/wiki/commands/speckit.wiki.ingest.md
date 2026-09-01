---
description: "Ingest a source (feature artifacts, file, or URL) and update the related wiki pages with citations"
---

# Ingest a Source into the Wiki

The only operation that writes knowledge into the wiki. Reads one source,
extracts what is worth keeping, and folds it into the related pages — with a
citation on every claim, cross-links between pages, and a hard cap on how
many pages one ingest may touch.

This is the **ingest** operation of Karpathy's LLM Wiki pattern: the human
curates sources; the LLM does the bookkeeping — summarizing, cross-referencing,
and keeping related pages consistent — so knowledge compounds instead of being
rediscovered from scratch on every question.

## User Input

```text
$ARGUMENTS
```

`$ARGUMENTS` names the source: a file path, a directory, or a URL. If empty,
default to the active feature's artifacts — `research.md` plus the decision
sections of `plan.md` (resolve the feature directory from the current git
branch, else the most recently modified directory under `specs/`).
`key=value` tokens (e.g. `max_pages_per_ingest=6`) are configuration overrides.

## Steps

### 1. Resolve configuration and the wiki

Load configuration exactly as `/speckit.wiki.init` does (config file →
`SPECKIT_WIKI_*` env vars → `key=value` arguments). If `WIKI_DIR/SCHEMA.md`
does not exist, create the skeleton first per the init command's step 3, and
say so in the report. Read `SCHEMA.md` — its rules override this prompt's
defaults where they conflict.

### 2. Register the source

In `sources.md`, dedup by normalized path/URL:

- New source → append the next `S-id` with type (`feature-artifact`, `file`,
  `url`), today's date in both date columns, and an empty pages list.
- Known source → update `Last ingested` only. Re-ingesting is the normal way
  to refresh the wiki after a source changed.

### 3. Read the source and extract wiki-worthy knowledge

Read the source once. Extract discrete items worth keeping — knowledge that
outlives the source's moment:

- decisions and their rejected alternatives
- constraints, limits, and gotchas that will bite again
- domain concepts and how this project uses them
- how a component actually works (when the source proves it)
- verified external facts

Skip transient status ("tests currently failing"), speculation, and anything
already on a page unchanged. Summarize — never paste bulk source content.

### 4. Choose the affected pages

Load `INDEX.md` (never all pages). Map each extracted item to an existing
page by topic, or to a new page typed per the schema's page-type table.
**Hard cap: at most `ingest.max_pages_per_ingest` pages created or updated in
one run.** If items would exceed the cap, ingest the most valuable ones and
list the remainder in the report as a suggested follow-up ingest.

### 5. Update the pages

Create `pages/` if missing. For each affected page (read only those):

- Merge the new items where they belong; update statements the source has
  made stale.
- Cite the source on every new or changed claim: `… (S007)`. When
  `ingest.require_citations` is true, no uncited claim may be written.
- If the new source **contradicts** an existing cited claim, keep both under
  a marker — `> ⚠ conflict: S002 says X; S007 says Y` — and flag it in the
  report. Never silently overwrite a cited claim.
- Cross-link: link new pages from the related existing pages you touched, and
  back. A new page unreachable from any other page is an orphan — avoid
  creating one.
- Update frontmatter: `updated` to today, `sources` to include the S-id.
- If a page exceeds `ingest.page_max_words`, split it per the schema and link
  both halves.

### 6. Update the registry and index

- `sources.md`: fill the source's `Pages touched` with the page filenames.
- `INDEX.md`: add/adjust one line per created or renamed page, grouped by
  type, and remove the "_No pages yet_" placeholder if present.

### 7. Report

- The source and its `S-id` (new or re-ingested).
- Pages created and updated — one line each on what changed.
- Conflicts flagged (if any) and items deferred by the page cap (if any).
- Next step: `/speckit.wiki.lint` if conflicts were flagged, otherwise
  `/speckit.wiki.query <question>` to test what the wiki now knows.

## Guardrails

- Sources are immutable — never edit the ingested file, and never edit
  `spec.md`, `plan.md`, or `tasks.md` from this command.
- Respect the page cap and word cap; splitting and deferring beat sprawling.
- Every claim written carries a source citation; conflicts are kept visible,
  not resolved by deletion.
- Do not load every page "for context" — INDEX plus the affected pages only.
