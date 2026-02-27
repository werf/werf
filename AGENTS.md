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

## Testing (MANDATORY)

- ALWAYS use `testify` (`assert`, `require`) when writing new tests.
- When writing tests as an AI agent → ALWAYS name the file `*_ai_test.go`, add `//go:build ai_tests` build tag, prefix test functions with `TestAI_`.
- ALWAYS place tests alongside source files, not in a separate directory.
- Test helpers go in `helpers_test.go` (or `helpers_ai_test.go` for AI-written helpers).
- Test fixtures go in `testdata/` subdirectory next to the tests.
- Shared test helpers are in `test/pkg/`.

## PR review guidelines (MANDATORY)

- NEVER add new external dependencies without flagging to the user first.
- NEVER introduce breaking user-facing changes (not API changes) unless they are hidden behind a feature flag. Flag to the user first.
- NEVER introduce changes that may compromise security. Flag to the user first.

## Related repositories

- [werf/nelm](https://github.com/werf/nelm) — Deployment engine used by werf. Go-based Kubernetes deployment tool that manages Helm charts.
- [werf/3p-helm](https://github.com/werf/3p-helm) — Helm fork. Provides chart loading, rendering, and release primitives. Changes to Helm internals go here, not in werf.
- [werf/kubedog](https://github.com/werf/kubedog) — Kubernetes resource tracking library.
- [werf/common-go](https://github.com/werf/common-go) — Shared Go libraries (secrets, CLI utilities, locking).
