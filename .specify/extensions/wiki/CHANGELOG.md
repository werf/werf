# Changelog

All notable changes to the LLM Wiki extension are documented here.
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-07-03

### Added

- Initial release, adapting the LLM Wiki pattern from Andrej Karpathy's
  ["LLM Wiki" gist](https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f)
  to spec-driven development: a persistent, LLM-maintained, compounding
  project wiki with three layers (raw sources → wiki pages → schema).
- `/speckit.wiki.init` — create the wiki skeleton: `SCHEMA.md` (structure and
  maintenance rules), `INDEX.md` (page directory), `sources.md` (source registry).
- `/speckit.wiki.ingest` — register a source (feature artifacts, file, or URL),
  extract wiki-worthy knowledge, and update the related pages with per-claim
  source citations, cross-links, conflict markers, and a hard per-run page cap.
- `/speckit.wiki.query` — answer questions strictly from wiki pages with page
  and source citations; report coverage gaps as concrete ingest suggestions.
- `/speckit.wiki.lint` — the maintenance pass: index drift, broken links,
  orphan pages, contradictions, stale claims, uncited claims; mechanical fixes
  applied per config, semantic issues reported with suggested edits only.
- `/speckit.wiki.status` — read-only snapshot (counts, freshness, open lint
  issues) plus exactly one recommended next action; the session-resume entry point.
- Optional hooks: `after_plan` → `speckit.wiki.ingest`,
  `after_implement` → `speckit.wiki.ingest`.
- `config-template.yml` for wiki location, ingest caps, query slices, and lint policy.
