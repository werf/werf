# spec-kit-brownfield

A [Spec Kit](https://github.com/github/spec-kit) extension that bootstraps Specification-Driven Development for existing codebases — auto-discover architecture, generate a tailored constitution, and incrementally migrate features into the SDD workflow.

## Problem

Spec Kit's `specify init` creates generic templates that don't fit established projects. Teams with existing codebases face friction adopting SDD:

- Generic constitution doesn't reflect actual tech stack, architecture, or conventions
- Templates reference placeholder paths instead of real project modules
- Multi-module projects (monorepos) get no guidance on code boundaries
- No way to bring existing features into the SDD workflow retroactively
- Manual constitution creation is tedious and error-prone for large codebases

## Solution

The Brownfield Bootstrap extension adds four commands for adopting spec-kit in existing projects:

| Command | Purpose | Modifies Files? |
|---------|---------|-----------------|
| `/speckit.brownfield.scan` | Auto-discover project structure, tech stack, frameworks, and architecture patterns | No — read-only |
| `/speckit.brownfield.bootstrap` | Generate spec-kit configuration tailored to the existing codebase | Yes — creates/updates constitution, templates, AGENTS.md |
| `/speckit.brownfield.validate` | Verify bootstrap output matches actual project structure and conventions | No — read-only |
| `/speckit.brownfield.migrate` | Incrementally adopt SDD for existing features with reverse-engineered specs | Yes — creates spec.md, plan.md, tasks.md for existing features |

## Installation

```bash
specify extension add --from https://github.com/Quratulain-bilal/spec-kit-brownfield/archive/refs/tags/v1.0.0.zip
```

## How It Works

### Phase 1: Discover

`/speckit.brownfield.scan` analyzes your codebase to build a project profile:

```
Project Profile
├── Tech Stack: TypeScript (68%), Python (32%)
├── Frontend: React 18, Vite, TailwindCSS
├── Backend: FastAPI, SQLAlchemy, PostgreSQL
├── Architecture: Frontend + Backend (separated)
├── Modules: client/, server/, shared/
├── Testing: Jest (frontend), pytest (backend)
├── CI/CD: GitHub Actions
├── Branch Pattern: feat/*, fix/*, chore/*
└── Conventions: kebab-case (frontend), snake_case (backend)
```

### Phase 2: Configure

`/speckit.brownfield.bootstrap` generates spec-kit configuration from the profile:

- **Constitution**: Rules derived from actual codebase conventions — not generic boilerplate
- **Spec template**: Project-specific sections (e.g., "Database Migrations" for ORM projects)
- **Plan template**: Module-aware implementation phases (e.g., frontend/backend split)
- **Tasks template**: Real test commands and build steps from your actual toolchain
- **AGENTS.md**: Agent boundaries for multi-module projects

### Phase 3: Verify

`/speckit.brownfield.validate` checks that configuration matches reality:

- Verifies all directory references actually exist
- Confirms mentioned frameworks are in dependency files
- Samples files to validate naming convention rules
- Detects drift if project has changed since bootstrap

### Phase 4: Migrate

`/speckit.brownfield.migrate` brings existing features into SDD:

- Reverse-engineers spec.md from code behavior and test cases
- Reconstructs plan.md from actual implementation patterns
- Generates tasks.md with all tasks marked complete
- Identifies gaps: missing tests, error handling, documentation

## Workflow

```
specify init                         ← Initialize spec-kit (if not already done)
       │
       ▼
/speckit.brownfield.scan             ← Discover tech stack and architecture
       │
       ▼
/speckit.brownfield.bootstrap        ← Generate tailored configuration
       │
       ▼
/speckit.brownfield.validate         ← Verify configuration accuracy
       │
       ▼
/speckit.brownfield.migrate          ← Bring existing features into SDD
       │
       ▼
/speckit.specify                     ← Start new features with project-aware templates
```

## Supported Project Types

| Type | Detection |
|------|-----------|
| **Monolith** | Single source tree, one entry point |
| **Monorepo** | Workspace config, `packages/` or `apps/` directories |
| **Microservices** | Multiple Dockerfiles, service directories |
| **Frontend + Backend** | Separate `client/`/`server/` directories |
| **Library/Package** | `setup.py`, `lib/` directory, published package config |

## Supported Tech Stacks

| Category | Detected |
|----------|----------|
| **Languages** | TypeScript, JavaScript, Python, Go, Java, Rust, C#, Ruby, PHP |
| **Frontend** | React, Vue, Angular, Svelte, Next.js, Nuxt |
| **Backend** | Express, FastAPI, Django, Spring, Rails, Gin, ASP.NET |
| **Databases** | PostgreSQL, MySQL, MongoDB, SQLite, Redis |
| **Package managers** | npm, yarn, pnpm, pip, Poetry, Go modules, Cargo, Maven, Gradle |
| **CI/CD** | GitHub Actions, GitLab CI, CircleCI, Jenkins |
| **Testing** | Jest, pytest, Go test, JUnit, RSpec, PHPUnit |

## Hooks

The extension registers an optional hook:

- **after_init**: Offers to scan the project after `specify init` to auto-detect conventions

## Design Decisions

- **Scan before bootstrap** — configuration is always derived from actual codebase analysis, never guessed
- **Confirm before writing** — bootstrap and migrate always show a plan and wait for approval
- **Merge, don't replace** — if constitution or templates already exist, merge changes rather than overwriting
- **Mark migrated specs** — reverse-engineered specs include `status: migrated` to distinguish from fresh specs
- **Gap reporting** — migrate command actively identifies missing tests, error handling, and documentation
- **Module-aware** — all commands understand monorepo and multi-module project structures

## Requirements

- Spec Kit >= 0.4.0
- Git >= 2.0.0

## Related

- Issue [#1436](https://github.com/github/spec-kit/issues/1436) — Brownfield Bootstrap: SDD Workflow for Existing Projects (30+ reactions)

## License

MIT
