# Concepts — from Karpathy's LLM Wiki to a Spec Kit extension

This document maps the source pattern to the extension's design and records
where — and why — the extension deliberately deviates.

## The source pattern

Karpathy's ["LLM Wiki" gist](https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f)
describes knowledge bases where an LLM *maintains* a persistent wiki rather
than retrieving from raw documents on demand. Its core claim: retrieval-only
systems have **no accumulation** — the model rediscovers knowledge from
scratch on every question — while a maintained wiki is a *persistent,
compounding artifact*.

Three layers:

1. **Raw sources** — immutable documents; the ground truth.
2. **The wiki** — LLM-generated markdown pages with cross-references; the
   working knowledge.
3. **The schema** — a configuration document defining the wiki's structure
   and maintenance workflows; the contract the LLM follows.

Three operations:

- **Ingest** — a new source arrives; the LLM reads it and updates the ~10–15
  related pages.
- **Query** — questions are answered from synthesized wiki content, with
  citations.
- **Lint** — regular health checks: contradictions, orphaned pages, stale
  claims, coverage gaps.

Division of labor: the human curates sources and asks questions; the LLM does
the bookkeeping — summarizing, cross-referencing, consistency maintenance.

## The mapping

| Gist concept | Extension realization |
|---|---|
| Raw sources | `wiki/sources.md` — an append-only registry (S-ids) of feature artifacts, files, and URLs; pointed to, never copied |
| Wiki pages | `wiki/pages/*.md` with frontmatter (`title`, `type`, `sources`, `updated`) and relative cross-links |
| The schema | `wiki/SCHEMA.md` — page-type taxonomy, naming/linking/citation rules, maintenance workflows; user-editable, command-obeyed |
| Ingest | `/speckit.wiki.ingest` — register → extract → cap-bounded page updates → index maintenance |
| Query | `/speckit.wiki.query` — slice-bounded, citation-required, honest coverage verdict |
| Lint | `/speckit.wiki.lint` — six checks; mechanical auto-fix, semantic report-only |
| "Human curates, LLM bookkeeps" | You choose what to ingest and when to resolve conflicts; commands handle citations, cross-links, index, and reports |

## Deliberate differences from the gist

**Project wiki, not personal knowledge base.** The gist's motivating cases
are personal research and reading notes. This extension targets a *repository*:
the wiki is committed, PR-reviewed, and shared by every teammate and agent.
That is why pages are typed (`decision` pages make "why" reviewable) and why
everything is plain relative-linked markdown.

**Spec Kit artifacts are first-class sources.** The default ingest target is
the active feature's `research.md` and plan decisions, and the two hooks
(`after_plan`, `after_implement`) fire at the moments spec-driven development
produces knowledge. The gist has no equivalent, because it has no surrounding
workflow.

**Hard caps everywhere.** The gist says ingest updates "10–15 related pages";
this extension makes the cap explicit configuration (`max_pages_per_ingest`,
`page_max_words`, `pages_slice`, `context_tokens`) and instructs commands to
defer overflow visibly rather than sprawl. Prompt-only systems need their
limits written down.

**Citations are mandatory, conflicts are preserved.** The gist asks for
citations in query answers; this extension requires a source ID on every
*written claim* (`require_citations`) and forbids silent overwrites — a
contradicting ingest produces a visible `> ⚠ conflict:` marker that lint
tracks until a human resolves it. In a shared repo, "the wiki quietly changed
its mind" is worse than a flagged disagreement.

**Lint splits mechanical from semantic.** Index drift and dead links are
deterministic and safe to auto-fix; contradictions and staleness are
judgment calls. The gist's lint is one maintenance notion; here the split is
the safety boundary: prose is never machine-rewritten.

**Status as a first-class operation.** The gist assumes a continuous
operator; agent sessions die (restart, compaction, window overflow).
`/speckit.wiki.status` is the session-resume entry point, mirroring the
pattern proven by the Research Harness extension: files are the memory,
context is rebuilt on demand.

## Relationship to OpenWiki

[OpenWiki](https://github.com/langchain-ai/openwiki) (LangChain) applies the
same maintained-docs idea to **code documentation**: an agent generates and
refreshes `openwiki/` docs *derived from the repository itself*, typically on
a schedule in CI.

The boundary between the two systems is the boundary of derivability:

- **Derivable from the code** (what a module does, how components connect):
  OpenWiki territory — regenerate it whenever the code changes.
- **Not derivable from the code** (why the alternative was rejected, the
  vendor limit that shaped the design, what the outage taught, external
  facts that were verified once): LLM Wiki territory — it must be
  *accumulated*, because no regeneration can recover it.

They compose: an OpenWiki page can even be an ingest source when a code-level
fact deserves a permanent, citable home next to the decision it influenced.

## Invariants (the short version)

1. `init` creates; it never overwrites.
2. Knowledge enters only through `ingest`; every claim carries a source ID.
3. Sources are immutable; the wiki never edits `spec.md` / `plan.md` /
   `tasks.md` or any ingested file.
4. Conflicts are marked, never silently resolved.
5. `query` and `status` are read-only; `lint` rewrites only `INDEX.md`,
   link targets, and `lint-report.md`.
6. Commands load slices (index + affected pages), never the whole wiki.
