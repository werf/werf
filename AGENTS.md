# AGENTS

All rules in this document are requirements — not suggestions. ALWAYS follow them.

werf is a CNCF Sandbox CLI tool to implement full-cycle CI/CD to Kubernetes. werf integrates into your CI system and leverages familiar and reliable technologies, such as Git, Dockerfile, Helm, and Buildah. werf uses [werf/nelm](https://github.com/werf/nelm) as its deployment engine.

## Repository map

- `cmd/werf` — CLI commands and flags. Deploy commands call nelm directly.
- `pkg/build`, `pkg/dockerfile`, `pkg/stapel`, `pkg/container_backend`, `pkg/buildah` — image build: stage assembly, Dockerfile and Stapel builders, Docker and Buildah backends. Buildah code is split into `*_linux.go` and `*_others.go`.
- `pkg/storage`, `pkg/image`, `pkg/docker_registry` — stage storage, image metadata, container registry clients.
- `pkg/cleaning`, `pkg/host_cleaning` — registry cleanup and local cache cleanup.
- `pkg/deploy` — bundles and Helm chart extenders; the deployment itself is driven by nelm.
- `pkg/config`, `pkg/giterminism_manager`, `pkg/git_repo`, `pkg/true_git` — `werf.yaml` parsing, giterminism, git access.
- `test/e2e` — e2e suites, `test/legacy_e2e` — what `task test:integration` runs, `test/pkg` — shared test helpers.
- `.agents/skills` — mandatory agent skills: branch and commit conventions, PR format, code review. `.claude/skills` is a symlink to it.

## Highest-priority rule (MANDATORY)

- NEVER add comments unless they document a non-obvious public API or explain genuinely non-obvious logic. NEVER add comments that restate what the code does, repeat the field/function name, describe obvious error handling, or act as section separators. When in doubt, don't comment.
- ALWAYS use `task` commands for build/test/lint/format — NEVER raw `go build`, `go test`, `go vet`, `go fmt`, or `golangci-lint` directly.
- ALWAYS read the matching skill in `.agents/skills/` BEFORE the action it governs and follow it verbatim: `git-conventions/SKILL.md` before naming a branch or writing a commit message, `pull-request/SKILL.md` before creating or updating a PR (title, description, draft by default), `review/SKILL.md` before reviewing code. These files are the source of truth and are NOT duplicated here.
- ALWAYS verify, don't assume — check the actual state before making changes.
- ALWAYS start with the simplest possible solution. If it works, stop. Add complexity only when justified by a concrete, current requirement — NEVER for hypothetical future needs.
- NEVER leave TODOs, stubs, or partial implementations.
- ALWAYS stay within the scope of what was asked. When asked to update a plan — only update the plan, don't change code. When asked to brainstorm/discuss — only discuss, don't write code. When asked to do X — do X and nothing else. NEVER make unsolicited changes.
- NEVER modify CHANGELOG.md, release notes, or other generated/workflow-managed files unless the user explicitly requests it.
- When deleting a block from structured data files (YAML, JSON, TOML), ALWAYS read surrounding lines to verify adjacent content (anchors, references, unrelated entries) is preserved.
- When removing content, ALWAYS clean up orphaned structural elements (comment separators, section headers, blank-line groups) that no longer serve a purpose.
- When renaming a type, function, or constant, ALWAYS rename all related local variables, parameters, and error messages that reference the old name. A rename is not complete until grep for the old name returns zero hits in affected packages.
- When removing a feature that has documentation in multiple languages (e.g. `pages_en/`, `pages_ru/`), ALWAYS apply the same removal to ALL language versions. NEVER assume English-only cleanup is sufficient.
- NEVER trust LSP/gopls diagnostics from unrelated files as proof of build failure. The ONLY source of truth for compilation is `task build`. LSP often reports false errors due to stale cache or incomplete workspace indexing.
- If you encounter errors in files OUTSIDE your task scope — STOP and report to the orchestrator. NEVER fix them yourself. Unsolicited fixes to unrelated files cause scope creep and may introduce regressions.

## Code style (MANDATORY)

[CODESTYLE.md](CODESTYLE.md) is the source of truth for design and conventions — it is short, read it before writing Go. Rules that get broken most often:

- Public functions and methods MUST take `context.Context` first, and all their arguments are required — passing nil is not allowed.
- Optional arguments go into a `<FunctionName>Options` struct passed last. NEVER use the functional options pattern.
- Every interface implementation MUST have a compile-time check: `var _ Animal = (*Dog)(nil)`.
- ALWAYS wrap errors with what was being done: `fmt.Errorf("read config: %w", err)`, not `"cannot read config"`.
- Avoid `iota`; prefix enum constants with the type name: `LogLevelDebug LogLevel = "debug"`.
- Use guard clauses and early returns to keep the happy path unindented.
- Use `samber/lo` helpers (`lo.Filter`, `lo.Map`, `lo.Ternary`, `lo.ToPtr`, …) when the standard library has no equivalent.

Follow [Effective Go](https://go.dev/doc/effective_go) and [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments). Violated here most often: NEVER use `this`/`self` as a receiver name, NEVER discard errors with `_`, NEVER use dot imports, NEVER use named or naked returns.

## Code navigation (MANDATORY)

- ALWAYS use LSP (`goToDefinition`, `findReferences`, `documentSymbol`, `hover`, `goToImplementation`, call hierarchy) to find definitions, usages, implementations, and callers. `grep` matches strings blindly: it hits comments and unrelated identifiers, and misses interface dispatch and aliased imports.
- Use `grep` ONLY for literal text — config keys, error message strings, annotation names.
- If your harness has a semantic code-search tool, prefer it over `grep` for intent-based questions ("how does X work"). If it does not, read the code: NEVER substitute keyword grepping for understanding.

## Commands (MANDATORY)

ALWAYS use these `task` commands. NEVER use raw `go build`, `go test`, `go fmt`, `go vet`, or `golangci-lint` directly. Pass extra args after `--` to forward them to the underlying command (e.g., `task test:unit -- -run TestMyFunc`).

- NEVER `go build` → ALWAYS `task build`. Builds binary to `./bin/`. Accepts `pkg=...`.
- NEVER `go test` → ALWAYS `task test:unit`. Accepts `paths="./pkg/..."`.
- NEVER `go test` (e2e) → ALWAYS `task test:e2e` with `paths="./pkg/..."` and `labelFilter="..."` (Ginkgo label filter) to target specific tests.
- NEVER `go test` (integration) → ALWAYS `task test:integration`. Legacy e2e tests.
- NEVER `go vet` → ALWAYS `task lint:golangci-lint`. golangci-lint includes vet checks. Accepts `golangciPaths="./pkg/..."`.
- NEVER `go fmt`/`gofmt` → ALWAYS `task format`. Accepts `paths="./pkg/..."`.
- NEVER `golangci-lint` → ALWAYS `task lint:golangci-lint`. Accepts `golangciPaths="./pkg/..."`.
- `task lint` — run all linters in parallel.
- `task enum:generate` — run enum generators.
- `task mock:generate` — run mock generators.
- `task mock:check` — verify generated mocks are up to date (runs `go generate -run mockgen` and diffs).
- `task doc:gen` — regenerate CLI reference docs. ALWAYS run after changing command descriptions, flags, or help text in Go source.

`format` and `lint*` come from a remote taskfile ([werf/common-ci](https://github.com/werf/common-ci)), so they need `TASK_X_REMOTE_TASKFILES=1` and network access.

## Verifying changes (MANDATORY)

After changing Go code, run these in order — `task format` mutates files, so it goes first:

1. `task format`
2. `task build`
3. `task lint`
4. `task test:unit`

NEVER assume a change compiles. While iterating, scope the slow steps (`task lint:golangci-lint golangciPaths="./pkg/foo/..."`, `task test:unit paths="./pkg/foo/..."`), then run them unscoped before handing the work over.

On macOS `task build` produces a **non-CGO** binary — the Buildah backend is only built for linux/amd64 (`task build:dev:linux:amd64:cgo`), so Buildah changes cannot be compiled or exercised locally. Unit tests run anywhere; e2e and integration tests need Linux with Docker and kind (`task test:setup:environment`).

## Testing (MANDATORY)

- ALWAYS use Ginkgo and Gomega when writing new tests. Prefer table-driven tests with `DescribeTable`.
- ALWAYS place tests alongside source files, not in a separate directory.
- Test helpers go in `helpers_test.go`.
- Test fixtures go in `testdata/` subdirectory next to the tests.
- Shared test helpers are in `test/pkg/`.

## PR review guidelines (MANDATORY)

- NEVER add new external dependencies without flagging to the user first.
- NEVER introduce breaking user-facing changes (not API changes) unless they are hidden behind a feature flag. Flag to the user first.
- NEVER introduce changes that may compromise security. Flag to the user first.

## Self-improvement

When a mistake was caused by a rule missing from AGENTS.md or CODESTYLE.md, propose that concrete rule to the user instead of silently swallowing the lesson.

## Related repositories

- [werf/nelm](https://github.com/werf/nelm) — Deployment engine used by werf. Go-based Kubernetes deployment tool that manages Helm charts.
- [werf/kubedog](https://github.com/werf/kubedog) — Kubernetes resource tracking library.
- [werf/common-go](https://github.com/werf/common-go) — Shared Go libraries (secrets, CLI utilities, locking).

`nelm`, `3p-helm`, `kubedog`, and `common-go` are ordinary versioned dependencies: fixing something inside them means a PR in that repository plus a version bump here — NEVER a local patch.

`go.mod` also has a `replace` block pointing several dependencies at forks, including `spf13/cobra` → `andremueller/cobra` and `containers/buildah`, `deislabs/oras`, `docker/buildx` → `werf/3p-*`. ALWAYS check that block before trusting upstream documentation for these libraries.
