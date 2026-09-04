---
description: "Answer a question from the wiki with page and source citations; flag coverage gaps"
---

# Query the Wiki

Answer a question **from the wiki pages, with citations** — or say plainly
that the wiki cannot answer it yet and what to ingest to fix that. This is the
**query** operation of Karpathy's LLM Wiki pattern, and it is also how you
test the wiki: an answer the pages cannot support is a coverage gap, not an
invitation to improvise.

## User Input

```text
$ARGUMENTS
```

`$ARGUMENTS` is the question. If empty, give a one-screen overview of what
the wiki currently knows (scope, page types and counts, notable pages) and
stop.

## Steps

### 1. Resolve the wiki

Load configuration exactly as `/speckit.wiki.init` does. If
`WIKI_DIR/SCHEMA.md` does not exist, report that no wiki exists and recommend
`/speckit.wiki.init` — do not answer the question from general knowledge.

### 2. Select pages

Read `INDEX.md` and pick the pages most relevant to the question — at most
`query.pages_slice`, staying within the `query.context_tokens` render budget.
Prefer pages whose type matches the question's shape (a "why did we…"
question → `decision` pages; a "how does X work" question → `component`
pages).

### 3. Answer with citations

Read only the selected pages. Compose the answer such that:

- every load-bearing statement names its wiki page and carries the page's
  source IDs, e.g. *…retries are idempotent by key
  ([idempotency-keys](wiki/pages/idempotency-keys.md), S003)*;
- conflicting page content (`> ⚠ conflict:` markers) is surfaced as
  conflicting, with both sides and their sources — never pick a winner
  silently;
- pages the answer relied on are listed at the end.

### 4. State coverage honestly

Close with one of:

- **Covered** — the pages fully support the answer.
- **Partial** — say exactly which part is unsupported, and recommend the
  concrete ingest that would close the gap
  (`/speckit.wiki.ingest <likely source>`).
- **Uncovered** — the wiki has nothing; say so, and recommend where the
  answer likely lives (a feature's `research.md`, a file, a URL) as an
  ingest target.

## Guardrails

- **Read-only.** This command writes nothing — no pages, no index, no registry.
- Answer only from wiki pages. General knowledge may be used to *phrase* the
  answer, never to *supply* facts the pages do not contain.
- If pages disagree, report the disagreement; resolving it is
  `/speckit.wiki.lint`'s and the user's job.
- Respect the slice and token caps — a focused answer from 5 pages beats a
  vague one from 50.
