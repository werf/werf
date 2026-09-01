---
description: "Generate spec-kit configuration tailored to the existing codebase"
---

# Bootstrap Spec-Kit

Generate a customized spec-kit configuration for an existing codebase. Uses the project profile from `/speckit.brownfield.scan` (or performs a scan if none exists) to create a constitution, templates, and agent configuration that match the project's actual architecture, tech stack, and conventions.

## User Input

```text
$ARGUMENTS
```

You **MUST** consider the user input before proceeding (if not empty). The user may specify preferences (e.g., "strict TDD", "minimal constitution"), a target directory for a monorepo module, or request specific template customizations.

## Prerequisites

1. Verify the current directory is a git repository
2. Verify a spec-kit project exists by checking for `.specify/` directory (run `specify init` first if missing)
3. Check if a project profile exists from a previous scan — if not, run a scan first

## Outline

1. **Load or generate project profile**: Check if `/speckit.brownfield.scan` has been run:
   - If a project profile exists, use it
   - If not, perform an inline scan to gather tech stack, architecture, and conventions
   - Confirm the profile with the user before proceeding

2. **Generate constitution**: Create `.specify/memory/constitution.md` tailored to the project:

   The constitution **MUST** include:
   - **Project identity**: Name, purpose, primary language(s), architecture pattern
   - **Code boundaries**: Which directories contain which types of code (e.g., "frontend code lives in `client/`, backend in `server/`")
   - **Naming conventions**: File naming, variable naming, branch naming as detected
   - **Testing requirements**: Test framework, test location, coverage expectations
   - **Dependency rules**: How modules depend on each other, what imports are allowed
   - **Quality gates**: Linting, formatting, CI checks that must pass

   The constitution **MUST NOT**:
   - Override existing project standards without user confirmation
   - Invent conventions that don't exist in the codebase
   - Include generic boilerplate unrelated to the actual project

3. **Customize spec template**: Modify `.specify/templates/spec-template.md` to reflect the project:
   - Add project-specific sections (e.g., "Database Migrations" for projects with ORMs)
   - Include architecture-aware requirements (e.g., "Frontend Requirements" and "API Requirements" for full-stack projects)
   - Reference actual module paths instead of generic placeholders

4. **Customize plan template**: Modify `.specify/templates/plan-template.md` to reflect the project:
   - Include module-aware implementation sections (e.g., separate phases for frontend/backend)
   - Reference actual test frameworks and build tools
   - Include project-specific complexity factors

5. **Customize tasks template**: Modify `.specify/templates/tasks-template.md` to reflect the project:
   - Task phases should map to the project's actual module structure
   - Include project-specific setup tasks (e.g., database migration, dependency install)
   - Reference actual test commands (e.g., `npm test`, `pytest`, `go test ./...`)

6. **Generate AGENTS.md** (if multi-module): For monorepos and multi-module projects:
   - Define agent boundaries per module
   - Specify which agent owns which directories
   - Set up inter-agent communication rules

7. **Present changes**: Show the user what will be created or modified:

   ```markdown
   # Bootstrap Plan

   | File | Action | Description |
   |------|--------|-------------|
   | `.specify/memory/constitution.md` | Create | Project-specific constitution with detected conventions |
   | `.specify/templates/spec-template.md` | Modify | Add project-specific sections (Database Migrations, API Contract) |
   | `.specify/templates/plan-template.md` | Modify | Add module-aware phases (frontend, backend, shared) |
   | `.specify/templates/tasks-template.md` | Modify | Add actual test commands and build steps |
   | `AGENTS.md` | Create | Agent boundaries for frontend and backend modules |

   Proceed with bootstrap? (confirm before writing)
   ```

8. **Execute bootstrap**: After user confirmation, write all files.

9. **Report**:

   ```markdown
   # Bootstrap Complete

   | Artifact | Status |
   |----------|--------|
   | Constitution | ✅ Created — 12 rules from detected conventions |
   | Spec template | ✅ Customized — added Database Migrations, API Contract sections |
   | Plan template | ✅ Customized — frontend/backend phase split |
   | Tasks template | ✅ Customized — actual test commands included |
   | AGENTS.md | ✅ Created — 2 agents (frontend, backend) |

   ## Next Steps
   - Review `.specify/memory/constitution.md` and adjust any rules
   - Run `/speckit.brownfield.validate` to verify configuration matches project
   - Run `/speckit.brownfield.migrate` to reverse-engineer specs for existing features
   - Start new features with `/speckit.specify` — templates are now project-aware
   ```

## Rules

- **Always confirm before writing** — show the bootstrap plan and wait for approval
- **Never overwrite without asking** — if constitution or templates already exist, show a diff and ask
- **Derive from reality** — every constitution rule must trace to something detected in the codebase
- **No invented conventions** — if the project has no consistent pattern for something, say so instead of guessing
- **Respect existing spec-kit setup** — if `.specify/` already has customizations, merge rather than replace
- **Module-aware** — for monorepos, generate configuration that respects module boundaries
