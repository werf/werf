---
description: "Auto-discover project structure, tech stack, frameworks, and architecture patterns"
---

# Scan Project

Analyze an existing codebase to discover its technology stack, architecture patterns, module structure, and coding conventions. This produces a project profile that the bootstrap command uses to generate tailored spec-kit configuration.

## User Input

```text
$ARGUMENTS
```

You **MUST** consider the user input before proceeding (if not empty). The user may specify a subdirectory to scan (e.g., "backend/"), a focus area (e.g., "only frontend"), or request a specific depth of analysis.

## Prerequisites

1. Verify the current directory is a git repository
2. Verify this is an existing project with source code (not an empty repo)

## Outline

1. **Detect tech stack**: Identify languages, frameworks, and tools by scanning:

   | Signal | What to Check |
   |--------|--------------|
   | **Languages** | File extensions (`.py`, `.ts`, `.go`, `.java`, `.rs`, etc.) and their relative proportions |
   | **Package managers** | `package.json`, `requirements.txt`, `pyproject.toml`, `go.mod`, `Cargo.toml`, `pom.xml`, `build.gradle` |
   | **Frameworks** | Dependencies in package files (React, Django, Spring, Express, Rails, etc.) |
   | **Build tools** | `Makefile`, `webpack.config.js`, `vite.config.ts`, `Dockerfile`, `docker-compose.yml` |
   | **CI/CD** | `.github/workflows/`, `.gitlab-ci.yml`, `.circleci/`, `Jenkinsfile` |
   | **Testing** | Test directories, test frameworks in dependencies (`jest`, `pytest`, `go test`, `JUnit`) |

2. **Analyze architecture**: Identify the project's structural patterns:

   | Pattern | Indicators |
   |---------|-----------|
   | **Monolith** | Single source tree, one entry point, shared database config |
   | **Monorepo** | Multiple `package.json`/`go.mod` files, workspace config, `packages/` or `apps/` directories |
   | **Microservices** | Multiple Dockerfiles, service directories, API gateway config |
   | **Frontend + Backend** | Separate `client/`/`server/` or `frontend/`/`backend/` directories |
   | **Library/Package** | `setup.py`, `lib/` directory, published package config |
   | **MVC** | `models/`, `views/`, `controllers/` directories |
   | **Layered** | `domain/`, `application/`, `infrastructure/`, `presentation/` directories |

3. **Map module structure**: For monorepos and multi-module projects:
   - Identify each module/package/service and its purpose
   - Detect inter-module dependencies (imports, shared types)
   - Note module boundaries (what code belongs where)
   - Identify shared libraries or utilities

4. **Extract conventions**: Detect existing coding patterns:
   - **Naming**: File naming (camelCase, kebab-case, snake_case), directory naming
   - **Branching**: Existing branch names and patterns from `git branch -a`
   - **Commit style**: Recent commit message patterns from `git log --oneline -20`
   - **Testing**: Test file location (`__tests__/`, `*_test.go`, `test_*.py`), test naming
   - **Documentation**: README structure, inline docs, API docs

5. **Detect existing governance**: Check for files that indicate existing project standards:
   - `CONTRIBUTING.md`, `ARCHITECTURE.md`, `ADR/` (Architecture Decision Records)
   - `.editorconfig`, linter configs (`.eslintrc`, `.flake8`, `rustfmt.toml`)
   - `CLAUDE.md`, `AGENTS.md`, `.specify/` (existing spec-kit setup)

6. **Output project profile**:

   ```markdown
   # Project Profile

   ## Tech Stack
   | Category | Detected |
   |----------|----------|
   | **Primary language** | TypeScript (68%), Python (32%) |
   | **Frontend** | React 18, Vite, TailwindCSS |
   | **Backend** | FastAPI, SQLAlchemy, PostgreSQL |
   | **Testing** | Jest (frontend), pytest (backend) |
   | **CI/CD** | GitHub Actions |
   | **Package manager** | npm (frontend), pip (backend) |

   ## Architecture
   - **Pattern**: Frontend + Backend (separated)
   - **Frontend**: `client/` — React SPA
   - **Backend**: `server/` — FastAPI REST API
   - **Database**: PostgreSQL (via SQLAlchemy ORM)

   ## Module Map
   | Module | Path | Purpose | Dependencies |
   |--------|------|---------|-------------|
   | Frontend | `client/` | React SPA | Backend API |
   | Backend | `server/` | REST API | Database |
   | Shared | `shared/` | Type definitions | — |

   ## Conventions
   - **File naming**: kebab-case (frontend), snake_case (backend)
   - **Branch pattern**: `feat/*`, `fix/*`, `chore/*`
   - **Commit style**: Conventional Commits
   - **Test location**: `__tests__/` (frontend), `tests/` (backend)

   ## Existing Governance
   - ✅ CONTRIBUTING.md
   - ✅ .eslintrc.json
   - ❌ ARCHITECTURE.md
   - ❌ .specify/ (no spec-kit setup)

   ## Recommendations
   - Run `/speckit.brownfield.bootstrap` to generate tailored spec-kit configuration
   - Constitution should enforce: kebab-case files (frontend), snake_case (backend)
   - Feature specs should map to the frontend/backend split
   ```

## Rules

- **Read-only** — this command never modifies any files
- **Respect .gitignore** — never scan `node_modules/`, `vendor/`, `dist/`, `.venv/`, or other ignored directories
- **Proportional analysis** — report language percentages based on actual file counts or line counts
- **No assumptions** — only report what is actually detected in the codebase
- **Handle empty results** — if a category has nothing detected, say "Not detected" rather than guessing
