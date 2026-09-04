# Implementation Plan: [FEATURE]

**Branch**: `[###-feature-name]` | **Date**: [DATE] | **Spec**: [link]

**Input**: Feature specification from `/specs/[###-feature-name]/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See
`.specify/templates/plan-template.md` for the execution workflow.

## Summary

[Extract from feature spec: primary requirement + technical approach from research]

## Technical Context

**Language/Version**: Go 1.25.5 (toolchain Go 1.25.8)

**Primary Dependencies**:
- Container building: `containers/buildah`, `containers/storage`, `containers/image`
- Kubernetes deployment: `werf/nelm`, `werf/kubedog`, Helm chart primitives
- Kubernetes client: `k8s.io/client-go`, `k8s.io/api`, `k8s.io/apimachinery`
- Container registry: `google/go-containerregistry`, `aws/aws-sdk-go-v2`
- Utilities: `samber/lo`, `werf/common-go`, `go-git/go-git`, `docker/docker`

**Storage**: OCI container registries, local Git repositories, and Buildah container storage

**Testing**: Ginkgo + Gomega for Go tests; project task-based unit, integration, and E2E suites

**Target Platform**: Linux and supported developer platforms; Kubernetes clusters where applicable

**Project Type**: CLI tool (Go binary via `cmd/werf/main.go`)

**Performance Goals**: Efficient image builds, registry operations, deployment tracking, and stage caching

**Constraints**: Self-contained CLI; Git, Helm, Kubernetes, and OCI-compatible workflows; platform-specific Buildah limitations MUST be documented

**Scale/Scope**: Single CLI binary with commands across build, deploy, cleanup, SBOM, configuration, and auxiliary domains

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

[Gates determined based on constitution file]

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
└── tasks.md
```

### Source Code (repository root)

```text
cmd/werf/       # CLI entry point and command wiring
pkg/            # Domain business logic and reusable components
test/e2e/       # Ginkgo end-to-end suites
test/legacy_e2e/ # Legacy integration suites
test/pkg/       # Shared test helpers and fixtures
```

**Structure Decision**: Keep command wiring in `cmd/werf/<domain>/` and implementation
in the relevant `pkg/<domain>/` package. Place unit tests alongside source files.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [violation] | [current need] | [reason] |
