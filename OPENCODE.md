## Hard blocks (MANDATORY)

- When you are about to use the wrong tool — STOP. Use the correct tool listed below. Do NOT proceed with the wrong tool even if it seems faster.
- If you already used the wrong tool — STOP and redo the step with the correct tool. Do NOT continue with the result from the wrong tool.
- NEVER use `explore` for intent-based or behavioral queries (e.g. "how does X work?", "find the orchestration flow for Y"). `explore` can only use grep/glob/ast_grep. For intent/behavior queries, ALWAYS use `codealive_codebase_search` directly or `task(category="quick"/"deep")`.

## Code navigation (MANDATORY)

ALWAYS use **LSP** for navigating code. NEVER use `grep` for finding definitions, references, or implementations — LSP is semantically precise, `grep` matches strings blindly and gives false positives.

Default action when unsure: ALWAYS use LSP.
Tool priority order: `lsp(operation=...)` → `grep` (ONLY as a fallback when LSP returns no results).

IMPORTANT: OMO's built-in LSP tools (`lsp_goto_definition`, `lsp_find_references`, `lsp_symbols`, `lsp_diagnostics`, `lsp_prepare_rename`, `lsp_rename`) are DISABLED via `disabled_tools` in `opencode.json` because they use a separate gopls instance that conflicts with the OpenCode LSP. ALWAYS use the OpenCode `lsp(operation=...)` tool instead. NEVER use `lsp_*` prefixed tools — they will not be available.

- When you want to find where a function/type/variable is defined and you have a call site — NEVER `grep` for it. ALWAYS use `lsp` with `operation="goToDefinition"`. It jumps to the exact definition, even across packages.
- When you want to find where a symbol is defined but you don't have a call site — NEVER `grep` for it. ALWAYS use `lsp` with `operation="workspaceSymbol"`. Fall back to `grep` ONLY if LSP returns no results.
- When you want to find all usages of a symbol — NEVER `grep` for the symbol name. ALWAYS use `lsp` with `operation="findReferences"`. Grep will match comments, strings, and unrelated identifiers with the same name.
- When you want to understand what's in a file — NEVER scroll through it or `grep` for `func`. ALWAYS use `lsp` with `operation="documentSymbol"`. It returns the complete structure: functions, types, constants, variables.
- When you want to check a symbol's type or read its documentation — NEVER guess from context. ALWAYS use `lsp` with `operation="hover"`. It returns the exact type signature and godoc.
- When you want to find which types implement an interface — NEVER `grep` for type names. ALWAYS use `lsp` with `operation="goToImplementation"`. Grep cannot reliably find implicit Go interface implementations.
- When you want to trace what calls a function — NEVER `grep` for the function name. ALWAYS use `lsp` → `prepareCallHierarchy` → `incomingCalls`. Grep will miss method calls, aliased imports, and interface dispatch.
- When you want to trace what a function calls — NEVER read through the function body manually. ALWAYS use `lsp` → `prepareCallHierarchy` → `outgoingCalls`.
- When you want to rename a symbol — NEVER find-and-replace. ALWAYS use `ast_grep_replace` for safe AST-aware renaming.
- When you want to check for errors before building — NEVER skip this step. ALWAYS run `task build` and `task lint:golangci-lint`.

## Code search (MANDATORY)

ALWAYS use **CodeAlive MCP** (`codealive_codebase_search`, `codealive_codebase_consultant`) for semantic/intent-based code search. NEVER substitute with `grep` when the query is about intent or behavior — CodeAlive understands code semantics, `grep` only matches character sequences.

Default action when unsure: ALWAYS use `codealive_codebase_search`.
Tool priority order: `codealive_codebase_search` / `codealive_codebase_consultant` → `ast_grep_search` → `grep` / `glob`.

- ALWAYS call `codealive_get_data_sources` before any CodeAlive tool. Without this, CodeAlive calls will fail.
- When you want to find code by intent or behavior (e.g. "how does release planning work?", "where is the DAG built?") — NEVER `grep` for keywords. ALWAYS use `codealive_codebase_search`. Grep will miss relevant code that uses different terminology, and drown you in irrelevant string matches.
- When you want architectural advice or explanations (e.g. "why is the DAG built this way?", "how do these packages relate?") — NEVER guess from reading a few files. ALWAYS use `codealive_codebase_consultant`. It has indexed the entire codebase and understands cross-cutting concerns.
- When you want to find structural code patterns (e.g. all functions with a specific signature, all `fmt.Errorf(... %w ...)` calls, all interface implementations) — NEVER use `grep` with regex hacks. ALWAYS use `ast_grep_search`. It matches on the AST, not on text, so it won't be fooled by comments, strings, or formatting differences.
- ONLY fall back to `grep`/`glob` for simple literal matching (specific strings, config keys, error messages, annotation names). This is the ONLY valid use of `grep` in this codebase.
- When delegating code search to subagents — NEVER use `explore` for CodeAlive or LSP searches. The `explore` agent can only use grep/glob/ast_grep (OMO upstream limitation). ALWAYS use `task(category="quick")` or `task(category="deep")` for semantic search. The `librarian` agent has full CodeAlive/Context7 access and works correctly.
- NEVER use `explore` for intent-based or behavioral queries (e.g. "how does X work?", "find the orchestration flow for Y"). These require CodeAlive, which `explore` cannot access. ALWAYS use `task(category="quick")` or `task(category="deep")` instead, or do the `codealive_codebase_search` yourself. Reserve `explore` ONLY for literal pattern matching (specific identifiers, strings, config keys).

## Subagent reliability (MANDATORY)

Background subagents (`explore`, `librarian`) have known reliability issues. Tasks can silently vanish (task ID becomes unfetchable), stall (prompt received but no tool calls made), or crash without producing an error status. NEVER depend on a single background agent for critical information.

- ALWAYS fire at least 2 `explore` agents when information is critical. Use different search strategies (e.g. one with grep, one with ast_grep + glob). If one fails, the other covers.
- ALWAYS have a direct-tool fallback ready. After firing background agents, immediately start your own parallel search with `codealive_codebase_search`, `LSP`, or `grep`. Do NOT wait idle for agent results.
- ALWAYS treat `background_output` returning "Task not found" as a silent failure, not a timing issue. The task is gone — move on to your fallback.
- When an agent's task ID vanishes or shows `status: running` with only the initial prompt message after 30+ seconds, assume it stalled. Do NOT keep polling — use your own tools instead.
- For this codebase, PREFER direct tools (`codealive_codebase_search` + `LSP` + `read`) over `explore` agents for targeted queries. Direct tools are 100% reliable and 3-10x faster. Reserve `explore` agents ONLY for broad multi-angle discovery where 3+ different search patterns are needed simultaneously.
- When delegating to `librarian`, ALWAYS use maximally directive prompts that name the exact tools to call and the exact queries to run. Open-ended prompts cause the librarian to spend 40-50s "thinking" before making any tool calls. BAD: "Find the best approach for X." GOOD: "Use `context7_resolve-library-id` for 'helm', then `context7_query-docs` for 'hook lifecycle annotations helm.sh/hook'. Return the raw documentation."

## Delegating with tool constraints (MANDATORY)

Subagents are stateless — they do NOT read OPENCODE.md or AGENTS.md. They only know what you pass in the `prompt=` parameter. When delegating tasks that involve code search, navigation, or external lookups, ALWAYS include the relevant tool rules from this file in the delegation prompt. Without this, subagents will use wrong tools (e.g. `grep` instead of CodeAlive, guessing APIs instead of using Context7).

- When delegating code search — ALWAYS include: "Use `codealive_codebase_search` for intent/behavioral queries. Call `codealive_get_data_sources` first. Use LSP `documentSymbol` to understand file structure. NEVER use grep for finding definitions or references."
- When delegating external knowledge lookup — ALWAYS include: "Use `context7_resolve-library-id` + `context7_query-docs` for library docs. Use `grep_app_searchGitHub` for real-world usage patterns. Use `websearch_web_search_exa` for current information. NEVER guess APIs from training data."
- When delegating to `librarian` — ALWAYS use directive prompts: "Search NOW for X and return results", not "Find the best approach for X". Librarian may get stuck planning instead of executing if the prompt is too open-ended.

## External knowledge (MANDATORY)

NEVER guess at APIs — ALWAYS look them up. Using wrong API signatures wastes time on compilation errors and subtle bugs.

Default action when unsure: ALWAYS use `context7_resolve-library-id` + `context7_query-docs`.
Tool priority order: `lsp` with `operation="goToDefinition"` → Context7 → `grep_app_searchGitHub` → `websearch_web_search_exa`. If you have a URL, ALWAYS use `webfetch`.

- When you want to check a Go type signature or read godoc for a dependency — NEVER guess from memory or training data. ALWAYS use `lsp` with `operation="goToDefinition"` to navigate to the actual source. Training data may be outdated or wrong.
- When you want library documentation, guides, or API examples — NEVER rely on training data. ALWAYS use `context7_resolve-library-id` + `context7_query-docs`. Context7 has up-to-date docs; your training data may be stale.
- When you want real-world usage patterns (how do other projects use this library?) — NEVER invent patterns. ALWAYS use `grep_app_searchGitHub`. It searches real code from real repositories.
- When you need current information, recent changes, or anything that might have changed after your training cutoff — NEVER answer from memory. ALWAYS use `websearch_web_search_exa`.
- When you have a specific URL to read (docs page, GitHub issue, PR) — NEVER summarize from memory. ALWAYS use `webfetch` to retrieve the actual content.

## Cluster inspection (MANDATORY)

Default action when unsure: ALWAYS use Kubernetes MCP tools.
Tool priority order: Kubernetes MCP tools ONLY.

- When you want to inspect Kubernetes cluster state (pods, deployments, services, logs, events) — NEVER run raw `kubectl` via Bash. ALWAYS use Kubernetes MCP tools (`kubernetes_kubectl_get`, `kubernetes_kubectl_describe`, `kubernetes_kubectl_logs`, etc.). MCP tools return structured data; raw kubectl output is harder to parse and error-prone.

## Verifying changes (MANDATORY)

ALWAYS verify after making changes, in this order. NEVER skip steps. NEVER assume "it probably compiles."

Default action when unsure: ALWAYS run the full verification pipeline.
Tool priority order: `task format` → `task build` → `task lint:golangci-lint` → `task test:unit`.

1. ALWAYS run `task format` first — it mutates files, so other checks must run after it.
2. ALWAYS run `task build` — verify it compiles. NEVER assume your changes compile without checking.
3. ALWAYS run `task lint:golangci-lint` — verify linting passes. NEVER ignore lint errors.
4. ALWAYS run `task test:unit` — verify tests pass. NEVER skip tests.

Scope verification with `golangciPaths=` for focused lint changes (e.g. `task lint:golangci-lint golangciPaths="./pkg/foo/..."`). Use `paths=` for other tasks. ALWAYS run full-project verification at the end of a task.

When changes affect CLI commands, deployment logic, or Kubernetes interactions, ALSO verify against the local dev cluster:

- ALWAYS run `./bin/werf` against the cluster to deploy/test.
- ALWAYS use Kubernetes MCP tools to inspect cluster state (see "Cluster inspection" above).
