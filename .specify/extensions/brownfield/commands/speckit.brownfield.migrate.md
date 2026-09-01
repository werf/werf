---
description: "Incrementally adopt SDD for existing features with reverse-engineered specs"
---

# Migrate Existing Features

Reverse-engineer spec-kit artifacts (spec.md, plan.md, tasks.md) for features that were built before spec-kit was adopted. This brings existing work into the SDD workflow so teams can track, refine, and extend features using spec-kit commands.

## User Input

```text
$ARGUMENTS
```

You **MUST** consider the user input before proceeding (if not empty). The user may specify a feature or module to migrate (e.g., "auth system", "payments module"), a branch name, or "all" to migrate everything.

## Prerequisites

1. Verify a spec-kit project exists by checking for `.specify/` directory
2. Verify git is available and the project is a git repository
3. Verify the project has existing source code to migrate (not an empty project)
4. Verify constitution exists (recommend running `/speckit.brownfield.bootstrap` first if missing)

## Outline

1. **Identify migration targets**: Determine what to migrate based on user input:

   | Input | Action |
   |-------|--------|
   | Specific feature name | Locate the feature in the codebase by searching for related files, modules, or directories |
   | Specific branch name | Analyze the branch's commits and changed files to identify the feature scope |
   | Module path | Treat the entire module as a single feature to migrate |
   | `all` | List all identifiable features and let the user select which to migrate |
   | No input | Show a list of detected features and ask the user to pick one |

2. **Detect feature boundaries**: For each migration target, determine its scope:
   - **Files**: Which source files implement this feature
   - **Tests**: Which test files cover this feature
   - **Dependencies**: What other modules or services this feature depends on
   - **API surface**: Endpoints, functions, or interfaces exposed by this feature
   - **Database**: Migrations, models, or schema changes related to this feature

3. **Reverse-engineer spec.md**: Analyze the code to reconstruct what the feature does:
   - **User scenarios**: Infer from test cases, route handlers, and UI components
   - **Requirements**: Extract from code behavior, validation rules, and error handling
   - **Success criteria**: Derive from test assertions and acceptance patterns
   - **Assumptions**: Note any hardcoded values, environment dependencies, or implicit requirements
   - Mark the spec as `status: migrated` to distinguish from specs created through the normal workflow

4. **Reverse-engineer plan.md**: Reconstruct the implementation approach:
   - **Technical context**: Actual frameworks, libraries, and patterns used
   - **Project structure**: Where the feature's code lives in the project
   - **Complexity assessment**: Based on file count, line count, and dependency depth

5. **Reverse-engineer tasks.md**: Create a task list reflecting what was actually built:
   - Each major component or module becomes a task group
   - Mark all tasks as `[x]` (completed) since the feature already exists
   - Include test tasks based on actual test files found
   - Note any gaps: code without tests, features without error handling

6. **Create feature branch and artifacts**: For each migrated feature:
   - Create a feature directory: `specs/{feature-name}/`
   - Write `spec.md`, `plan.md`, and `tasks.md` into the feature directory
   - Do **not** create a git branch — the feature already exists on its branch or main

7. **Present migration plan**: Show what will be created before writing:

   ```markdown
   # Migration Plan: User Authentication

   ## Detected Scope
   | Category | Files | Lines |
   |----------|-------|-------|
   | Source | 8 files | ~420 lines |
   | Tests | 3 files | ~180 lines |
   | Migrations | 2 files | ~45 lines |

   ## Artifacts to Generate
   | File | Content |
   |------|---------|
   | `specs/user-auth/spec.md` | 4 user scenarios, 12 requirements, 6 success criteria |
   | `specs/user-auth/plan.md` | 3 implementation phases, 8 technical decisions |
   | `specs/user-auth/tasks.md` | 14 tasks (all completed), 2 gaps identified |

   ## Gaps Found
   - ⚠️ No error handling tests for expired tokens
   - ⚠️ No rate limiting on login endpoint

   Proceed with migration?
   ```

8. **Execute migration**: After user confirmation, write all artifacts.

9. **Report**:

   ```markdown
   # Migration Complete: User Authentication

   | Artifact | Status |
   |----------|--------|
   | spec.md | ✅ Created — 4 scenarios, 12 requirements |
   | plan.md | ✅ Created — 3 phases |
   | tasks.md | ✅ Created — 14/14 tasks complete |

   ## Identified Gaps
   1. No error handling tests for expired tokens → consider `/speckit.specify` for a follow-up feature
   2. No rate limiting on login endpoint → consider `/speckit.bugfix.report` to track

   ## Next Steps
   - Review generated artifacts in `specs/user-auth/`
   - Use `/speckit.refine.update` to adjust any inaccurate specs
   - Use `/speckit.specify` for new features — they'll follow the same SDD workflow
   - Run `/speckit.brownfield.migrate` again for additional features
   ```

## Rules

- **Always confirm before writing** — show the migration plan and wait for user approval
- **Honest assessment** — if the code is unclear or poorly documented, say so in the spec rather than inventing explanations
- **Mark as migrated** — all migrated specs must include `status: migrated` to distinguish from fresh specs
- **Identify gaps** — actively look for missing tests, error handling, or documentation and report them
- **Non-destructive** — never modify existing source code, only create spec artifacts
- **One feature at a time** — for "all" input, migrate features sequentially with confirmation between each
- **Respect constitution** — generated artifacts must follow the project's constitution rules
