# AGENTS

All rules in this document are requirements — not suggestions. ALWAYS follow them.

werf is a CNCF Sandbox CLI tool to implement full-cycle CI/CD to Kubernetes. werf integrates into your CI system and leverages familiar and reliable technologies, such as Git, Dockerfile, Helm, and Buildah. werf uses [werf/nelm](https://github.com/werf/nelm) as its deployment engine.

## Highest-priority rule (MANDATORY)

- NEVER add comments unless they document a non-obvious public API or explain genuinely non-obvious logic. NEVER add comments that restate what the code does, repeat the field/function name, describe obvious error handling, or act as section separators. When in doubt, don't comment.
- ALWAYS use `task` commands for build/test/lint/format — NEVER raw `go build`, `go test`, `go vet`, `go fmt`, or `golangci-lint` directly.
- ALWAYS verify, don't assume — check the actual state before making changes.
- ALWAYS start with the simplest possible solution. If it works, stop. Add complexity only when justified by a concrete, current requirement — NEVER for hypothetical future needs.
- NEVER leave TODOs, stubs, or partial implementations.
- ALWAYS stay within the scope of what was asked. When asked to update a plan — only update the plan, don't change code. When asked to brainstorm/discuss — only discuss, don't write code. When asked to do X — do X and nothing else. NEVER make unsolicited changes.
- NEVER modify CHANGELOG.md, release notes, or other generated/workflow-managed files unless the user explicitly requests it.
- When deleting a block from structured data files (YAML, JSON, TOML), ALWAYS read surrounding lines to verify adjacent content (anchors, references, unrelated entries) is preserved.
- When removing content, ALWAYS clean up orphaned structural elements (comment separators, section headers, blank-line groups) that no longer serve a purpose.
- When renaming a type, function, or constant, ALWAYS rename all related local variables, parameters, and error messages that reference the old name. A rename is not complete until grep for the old name returns zero hits in affected packages.
- When removing a feature that has documentation in multiple languages (e.g. `pages_en/`, `pages_ru/`), ALWAYS apply the same removal to ALL language versions. NEVER assume English-only cleanup is sufficient.

## Code style

### Design (MANDATORY)

> The code style rules below are adapted from [CODESTYLE.md](CODESTYLE.md). If you are asked to update code style rules, update CODESTYLE.md first, then regenerate this section to match, using ALWAYS/NEVER/MUST phrasing.

- ALWAYS prefer stupid and simple over abstract and extendable.
- ALWAYS prefer a bit of duplication over complex abstractions.
- ALWAYS prefer clarity over brevity in names.
- ALWAYS minimize interfaces, generics, embedding.
- ALWAYS prefer fewer types. Prefer no types over few. Prefer data types over types with behavior.
- ALWAYS prefer functions over methods. ALWAYS prefer public fields over getters/setters.
- ALWAYS keep everything private/internal as much as possible.
- ALWAYS validate early, validate a lot. ALWAYS keep APIs stupid and minimal.
- NEVER prefer global state. ALWAYS prefer simplicity over micro-optimizations.
- ALWAYS use libraries for complex things instead of reinventing the wheel.
- NEVER add comments unless they document a non-obvious public API or explain genuinely non-obvious logic. NEVER add obvious/redundant comments, NEVER add comments restating what code does. When in doubt, don't comment.

### Conventions (MANDATORY)

> The code style rules below are adapted from [CODESTYLE.md](CODESTYLE.md). If you are asked to update code style rules, update CODESTYLE.md first, then regenerate this section to match, using ALWAYS/NEVER/MUST phrasing.

- All public functions/methods MUST accept `context.Context` as the first parameter.
- All arguments of a public function are required — passing nil not allowed.
- Optional arguments via `<FunctionName>Options` as the last argument. NEVER use functional options.
- Use guard clauses and early returns to keep the happy path unindented.
- Use `samber/lo` helpers: `lo.Filter`, `lo.Find`, `lo.Map`, `lo.Contains`, `lo.Ternary`, `lo.ToPtr`, `lo.Must`, etc.
- Constructors: `New<TypeName>[...]()`. No network/filesystem calls in constructors.
- Interfaces: ALWAYS add `var _ Animal = (*Dog)(nil)` compile-time check.
- Constants: avoid `iota`. Prefix enum constants with type name: `LogLevelDebug LogLevel = "debug"`.
- Errors: ALWAYS wrap with context: `fmt.Errorf("read config: %w", err)`. Describe what is being done, not what failed. Panic on programmer errors. Prefer one-line `if err := ...; err != nil`.

### Go standard guidelines (MANDATORY)

Follow [Effective Go](https://go.dev/doc/effective_go) and [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments). Commonly violated rules:

- NEVER use `this`/`self` as receiver names. Use 1-2 letter names, consistent across methods.
- NEVER discard errors with `_`. Indent error flow, not happy path.
- NEVER use dot imports.
- NEVER use named returns or naked returns.

## Commands (MANDATORY)

ALWAYS use these `task` commands. NEVER use raw `go build`, `go test`, `go fmt`, `go vet`, or `golangci-lint` directly. Pass extra args after `--` to forward them to the underlying command (e.g., `task test:unit -- -run TestMyFunc`).

- NEVER `go build` → ALWAYS `task build`. Builds binary to `./bin/`. Accepts `pkg=...`.
- NEVER `go test` → ALWAYS `task test:unit`. Accepts `paths="./pkg/..."`.
- NEVER `go test` (e2e) → ALWAYS `task test:e2e`. Accepts `paths="./pkg/..."` and `labelFilter="..."` (Ginkgo label filter). ALWAYS use `labelFilter` to target specific tests.
- NEVER `go test` (integration) → ALWAYS `task test:integration`. Legacy e2e tests.
- NEVER `go vet` → ALWAYS `task lint:golangci-lint`. golangci-lint includes vet checks. Accepts `golangciPaths="./pkg/..."`.
- NEVER `go fmt`/`gofmt` → ALWAYS `task format`. Accepts `paths="./pkg/..."`.
- NEVER `golangci-lint` → ALWAYS `task lint:golangci-lint`. Accepts `golangciPaths="./pkg/..."`.
- `task lint` — run all linters in parallel.
- `task enum:generate` — run enum generators.
- `task mock:generate` — run mock generators.
- `task mock:check` — verify generated mocks are up to date (runs `go generate -run mockgen` and diffs).
- `task doc:gen` — regenerate CLI reference docs. ALWAYS run after changing command descriptions, flags, or help text in Go source.

## Testing (MANDATORY)

- ALWAYS use `testify` (`assert`, `require`) when writing new tests.
- When writing tests as an AI agent → ALWAYS name the file `*_ai_test.go`, add `//go:build ai_tests` build tag, prefix test functions with `TestAI_`.
- ALWAYS place tests alongside source files, not in a separate directory.
- Test helpers go in `helpers_test.go` (or `helpers_ai_test.go` for AI-written helpers).
- Test fixtures go in `testdata/` subdirectory next to the tests.
- Shared test helpers are in `test/pkg/`.

## Reproducing failed CI tests locally (MANDATORY)

When a CI test fails and you need to reproduce it locally, follow this procedure.

### Step 1: Identify the CI job and map it to a task command

CI job names in `tests.yml` map to `task` commands as follows:

| CI job name | task command | default paths |
|---|---|---|
| `unit` | `task test:unit` | `./pkg ./cmd` |
| `integration_main` | `task test:integration` | `./test/legacy_e2e` |
| `integration_git` | `task test:integration` | `test/legacy_e2e/suites/build/stapel_image/git` |
| `e2e_simple` | `task test:e2e:simple` | `./test/e2e` |
| `e2e_complex` | `task test:e2e:complex` | `./test/e2e` |
| `e2e_extra` | `task test:e2e:extra` | `./test/e2e` |

All integration and e2e tests use Ginkgo. Unit tests also use Ginkgo (via `test:ginkgo`).

### Step 2: Narrow down to the failing test package

From the CI logs, find the test package path. This is the Go package path of the failing test, e.g. `./test/legacy_e2e/suites/build/stapel_image/build_phase`. Use it as the `paths=` variable.

### Step 3: Narrow down to the failing test case

From the CI logs, find the failing `It(...)` description text. Use Ginkgo's `--focus` flag (passed after `--`) to run only that test. The `--focus` value is a regex matched against the full Ginkgo test path (`Describe` + `Context` + `It` texts joined).

For e2e tests that use `Label(...)`, you can also use `labelFilter="..."` variable instead of `--focus`.

### Step 4: Build the command

Template:

```
task --yes -p <task-name> paths="<package-path>" -- --focus "<It description text>" -v
```

- `--yes` — auto-confirm remote taskfile downloads.
- `-p` — run in parallel mode (same as CI).
- `--` — separates `task` args from Ginkgo args.
- `--focus "<regex>"` — Ginkgo flag to run only matching tests.
- `-v` — verbose Ginkgo output.

### Examples

**Integration test** (from CI job `integration_main`, test `"should build install stage twice"`):
```
task --yes -p test:integration paths="./test/legacy_e2e/suites/build/stapel_image/build_phase" -- --focus "should build install stage twice" -v
```

**E2e test** (from CI job `e2e_simple`, test `"Simple build"`):
```
task --yes -p test:e2e:simple paths="./test/e2e/build" -- --focus "Simple build" -v
```

**E2e test with label filter** (all tests labeled `"build"`):
```
task --yes -p test:e2e labelFilter="build" paths="./test/e2e/build"
```

**Unit test** (a specific test in `./pkg/config`):
```
task --yes -p test:unit paths="./pkg/config" -- --focus "should parse config" -v
```

### Additional useful Ginkgo flags (pass after `--`)

- `--flake-attempts=N` — retry flaky tests N times (CI uses `--flake-attempts=3`).
- `--keep-going` — don't stop on first package failure.
- `--skip-package "path1,path2"` — exclude packages.
- `--until-it-fails` — run repeatedly until failure (useful for flaky test investigation).
- `--dry-run` — list tests without running them.

### Prerequisites

Before running integration or e2e tests, the local test environment must be set up:
```
task --yes deps:install:ginkgo
task --yes test:setup:environment
```

This creates a kind cluster and a local Docker registry. Clean up after with `task test:cleanup:environment`.

## PR review guidelines (MANDATORY)

- NEVER add new external dependencies without flagging to the user first.
- NEVER introduce breaking user-facing changes (not API changes) unless they are hidden behind a feature flag. Flag to the user first.
- NEVER introduce changes that may compromise security. Flag to the user first.

## Related repositories

- [werf/nelm](https://github.com/werf/nelm) — Deployment engine used by werf. Go-based Kubernetes deployment tool that manages Helm charts.
- [werf/3p-helm](https://github.com/werf/3p-helm) — Helm fork. Provides chart loading, rendering, and release primitives. Changes to Helm internals go here, not in werf.
- [werf/kubedog](https://github.com/werf/kubedog) — Kubernetes resource tracking library.
- [werf/common-go](https://github.com/werf/common-go) — Shared Go libraries (secrets, CLI utilities, locking).
