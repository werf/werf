# AGENTS

## Knowledge Discovery

When searching for information about the project or external Go packages, you MUST use Model Context Protocol (MCP) servers for all Go-related navigation and documentation:

- `gopls-mcp`: ALWAYS use for Go project-specific symbols, definitions, and code navigation instead of `grep` or `find_path`.
- `godoc-mcp`: ALWAYS use for documentation of standard library and external Go packages.

## Software Engineering Principles

- SOLID (SRP/OCP/LSP/ISP/DIP) for extensibility and safe change.
- DRY: one source of truth for business rules and schemas.
- KISS and YAGNI: prefer the simplest solution that meets current needs.
- Separation of Concerns; high cohesion, low coupling.
- Composition over inheritance; avoid deep hierarchies.
- Fail fast; validate inputs and assumptions early.
- Design for testability; keep code readable over clever.
- Security by design; least privilege and safe defaults.
- Observability by design: logs, metrics, traces for critical paths.

## BDD (Behavior-Driven Development) Process

I follow BDD as an iterative development process to ensure code meets requirements and remains maintainable:

1.  **Define Behavior**: Understand the expected behavior and create/update tests.
2.  **Iterative Implementation**:
    - Run **Tests** (see corresponding section) to confirm failure or verify progress.
    - Write/refine code to satisfy the requirements.
    - Repeat until all tests pass.
3.  **Verification**: Execute **Unit tests** and, optionally (at user's request), **E2E tests** to ensure no regressions and full compliance with the feature scope.
4.  **Finalize**: Once tests pass, perform **Formatting** (see section below) to ensure code quality and style consistency.

## Tests

### Unit tests

To run unit tests for a specific package, use:

```bash
task --yes test:unit paths=<pkg_dir>
```

Where:

- `pkg_dir` is the directory of the package you want to run the tests for.

### E2E tests

To run end-to-end (e2e) tests for a specific package, ALWAYS use the `labelFilter` variable to target specific tests:

```bash
task --yes test:e2e paths=<pkg_dir> labelFilter=<filter_expression>
```

Where:

- `filter_expression` is a Ginkgo label filter (e.g., `"v1alpha2 && !long"`) to focus on specific test scenarios (prevents running the entire suite).
- `pkg_dir` is the directory of the package you want to run the tests for.

## Formatting

Run formatting after you finish testing a package.

To format code for a specific package, use:

```bash
task --yes format paths=<pkg_dir>
```

Where:

- `pkg_dir` is the directory of the package you want to format.
