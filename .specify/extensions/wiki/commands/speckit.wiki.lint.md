---
description: "Health-check the wiki: contradictions, orphan pages, stale claims, broken links, index drift"
---

# Lint the Wiki

The maintenance pass that keeps a compounding wiki from rotting. Karpathy's
LLM Wiki pattern calls for regular health checks — contradictions, orphaned
pages, stale claims, coverage gaps — because an unmaintained knowledge base
is abandoned, not merely imperfect.

Mechanical problems are fixed automatically (when configured); semantic
problems are **reported with suggested edits, never auto-rewritten**.

## User Input

```text
$ARGUMENTS
```

Optional scope: a page filename (lint one page and its neighbors) or a check
name from the list below (run only that check). Empty means the full pass.

## Steps

### 1. Resolve the wiki

Load configuration exactly as `/speckit.wiki.init` does. If
`WIKI_DIR/SCHEMA.md` does not exist, report that and stop. Read `SCHEMA.md`,
`INDEX.md`, and `sources.md`; read pages as each check requires.

### 2. Run the checks

| Check | Finds | Severity |
|-------|-------|----------|
| `index-drift` | pages missing from `INDEX.md`; index lines pointing at nonexistent files | mechanical |
| `links` | relative links to missing pages; citations naming unknown S-ids | mechanical |
| `orphans` | pages no other page links to (INDEX itself does not count) | structural |
| `contradictions` | `> ⚠ conflict:` markers still unresolved; pairs of pages sharing a source or link that assert incompatible claims | semantic |
| `stale` | pages whose `updated` is older than `lint.stale_after_days`; pages whose source was re-ingested after the page was last updated | semantic |
| `citations` | claims without a source ID, when `ingest.require_citations` is true | semantic |

Bound the contradiction check: compare only pages that share a source ID or a
link — never all pairs.

### 3. Apply mechanical fixes (per config)

If `lint.auto_fix` is `index-and-links`:

- regenerate `INDEX.md` from the actual `pages/` contents (grouped by
  frontmatter `type`, one line each);
- repair links whose target was renamed (when the target is unambiguous) and
  remove links to pages that no longer exist.

If `lint.auto_fix` is `none`, list these as suggested fixes instead. **Never**
auto-fix a semantic finding — do not rewrite prose, resolve conflicts, or
delete claims.

### 4. Write the report

Overwrite `WIKI_DIR/lint-report.md`:

```markdown
# Wiki Lint Report — <today's date>

| # | Check | Severity | Page | Finding | Suggested fix |
|---|-------|----------|------|---------|---------------|
| 1 | contradictions | semantic | payments-retries.md | S002 vs S007 on retry cap | re-ingest source S002; keep the newer claim, drop the marker |
```

One row per finding; a clean pass writes the header plus "No findings."

### 5. Report

- Counts per check, fixes applied vs. suggested.
- The two or three highest-value findings, verbatim from the table.
- Next step: the single action that clears the most severe finding
  (usually a re-ingest of a stale source, or a conflict for the user to
  resolve by editing the page).

## Guardrails

- Mechanical fixes only — `INDEX.md` and link targets. Page prose, claims,
  conflict markers, and `sources.md` history are never modified by lint.
- `lint-report.md` is the only file lint may overwrite wholesale.
- Findings must be verifiable: every row names the page and the exact claim
  or link it refers to.
- Stay within the configured token budget: read pages check-by-check, not
  the whole wiki at once.
