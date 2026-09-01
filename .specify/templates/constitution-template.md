# werf Constitution

## Core Principles

### I. Simplicity Over Abstraction

Prefer simple, direct solutions over abstract or extensible designs. Prefer a small
amount of duplication over complex abstractions. Minimize interfaces, generics, and
embedding. Prefer functions over methods where practical, and prefer straightforward
data structures over types with unnecessary behavior. Complexity MUST be justified by
a concrete current requirement.

### II. Go Idiomatic Code

Follow Effective Go and Go Code Review Comments. All public functions and methods MUST
accept `context.Context` as the first parameter, and all public arguments are required.
Optional public arguments MUST use a `<FunctionName>Options` struct. Errors MUST be
wrapped with action-oriented context using `fmt.Errorf("doing something: %w", err)`.
Use guard clauses and early returns. Interface implementations MUST have compile-time
checks. Avoid `iota`, named returns, dot imports, and `this`/`self` receiver names.

### III. Minimal Public Surface

Keep implementation details private or internal whenever possible. Validate input early
and keep APIs minimal. Constructors MUST NOT perform network or filesystem operations
or other resource-intensive work. Comments MUST be limited to non-obvious public APIs
or genuinely non-obvious logic.

### IV. Test-Before-Merge

Tests MUST be placed alongside the source they verify. New Go tests MUST use Ginkgo and
Gomega and follow existing project test conventions. Mocks MUST be generated with the
project mock task rather than written by hand. A change MUST pass the applicable build,
formatting, lint, and test gates before merge; changes to CLI commands MUST include a
runtime smoke check in addition to unit tests.

### V. Conventional Commits

Commits MUST use the Conventional Commits format `type(scope): subject` and follow the
repository's contribution conventions, including a subject of no more than 72
characters. Branches and pull requests MUST follow the repository's documented
conventions. Contributions MUST preserve the project's DCO/sign-off requirements.

## Code Boundaries

| Layer | Path | Purpose |
|-------|------|---------|
| **CLI commands** | `cmd/werf/` | Cobra command tree and thin wiring layer |
| **Libraries** | `pkg/...` | Domain business logic and reusable components |
| **E2E tests** | `test/e2e/` | Ginkgo end-to-end test suites |
| **Legacy tests** | `test/legacy_e2e/` | Legacy integration tests |
| **Shared test helpers** | `test/pkg/` | Reusable test utilities and fixtures |

CLI packages MUST avoid owning business logic that belongs in `pkg/`. Packages under
`pkg/` MUST NOT depend on `cmd/` packages.

## Dependency Rules

- Internal packages under `pkg/` MAY import other `pkg/` subpackages when the dependency
  reflects a real domain relationship.
- `cmd/werf/` MAY import `pkg/` subpackages but MUST NOT be imported by `pkg/`.
- External dependencies MUST be managed through `go.mod`; adding one MUST be flagged
  for review before adoption.
- Forked dependencies MUST remain documented by the relevant `go.mod` `replace`
  directives and MUST NOT be silently replaced with upstream modules.

## Build & Quality Gates

Use project tasks instead of raw Go tooling:

- Formatting: `task format`.
- Build: `task build`.
- Lint: install the linter once with `task deps:install:golangci-lint`, then run
  `task lint` or the scoped lint task.
- Unit tests: `task test:unit`, preferably scoped while iterating.
- E2E tests: run `task test:e2e` with both `paths` and `labelFilter`; prepare the
  environment with `task test:setup:environment` when required by the platform or
  suite, and clean it with `task test:cleanup:environment` afterward.
- Integration tests: `task test:integration` after the required environment setup.
- Mocks: `task mock:generate`; verify them with `task mock:check`.
- CLI help changes: regenerate documentation with `task doc:gen`.

For Go changes, the final verification sequence MUST be `task format`, `task build`,
`task lint`, and `task test:unit`, followed by the applicable E2E/integration checks.

## Governance

This constitution supersedes conflicting project practices. Amendments MUST be made in
a pull request, include a Sync Impact Report, explain the semantic-version bump, and
update all affected Spec Kit templates and runtime guidance. Reviewers MUST verify
compliance with this constitution and require a documented justification for any
exception. `AGENTS.md`, `CODESTYLE.md`, and `CONTRIBUTING.md` provide operational
rules and MUST remain consistent with this document.

The constitution uses semantic versioning: MAJOR for incompatible governance changes,
MINOR for new or materially expanded principles or sections, and PATCH for clarifying
or non-semantic edits. The ratification date records adoption of this constitution;
the last amended date records the latest approved amendment.

**Version**: [CONSTITUTION_VERSION] | **Ratified**: [RATIFICATION_DATE] | **Last Amended**: [LAST_AMENDED_DATE]
