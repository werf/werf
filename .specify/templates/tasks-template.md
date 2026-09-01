---

description: "Task list template for werf feature implementation"
---

# Tasks: [FEATURE NAME]

**Input**: Design documents from `/specs/[###-feature-name]/`

**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Include test tasks whenever the feature changes behavior, contracts, or CLI commands.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths and task commands in descriptions.

## Path Conventions

- **CLI commands**: `cmd/werf/<domain>/`
- **Business logic**: `pkg/<domain>/`
- **Unit tests**: co-located with source files as `*_test.go`
- **E2E tests**: `test/e2e/<domain>/`
- **Test helpers**: `test/pkg/`

## Build & Test Commands

- Build: `task build`
- Format: `task format`
- Unit tests: `task test:unit paths="./pkg/<domain>/..."`
- Lint: `task deps:install:golangci-lint` once per session, then `task lint`
- E2E: `task test:e2e paths="./test/e2e/<domain>/..." labelFilter="<label>"`
- Integration: `task test:integration` after required environment setup
- Mocks: `task mock:generate`, then `task mock:check`
- CLI help: `task doc:gen` after changing help text

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [ ] T001 Create project structure per implementation plan
- [ ] T002 Initialize [language] project with [framework] dependencies
- [ ] T003 [P] Configure linting and formatting tools

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [ ] T004 Define or update types in `pkg/<domain>/<types>.go`
- [ ] T005 [P] Implement core logic in `pkg/<domain>/<core>.go`
- [ ] T006 [P] Register CLI wiring in `cmd/werf/<domain>/<command>.go`
- [ ] T007 Add validation, action-oriented error context, and logging where required
- [ ] T008 Add or update generated mocks with `task mock:generate`

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - [Title] (Priority: P1) 🎯 MVP

**Goal**: [Brief description of what this story delivers]

**Independent Test**: [How to verify this story independently]

### Tests for User Story 1 (OPTIONAL - only if tests requested) ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T010 [P] [US1] Unit test for [core logic] in `pkg/<domain>/<file>_test.go`
- [ ] T011 [P] [US1] Unit test for [CLI command] in `cmd/werf/<domain>/<file>_test.go`

### Implementation for User Story 1

- [ ] T012 [P] [US1] Implement [core type/interface] in `pkg/<domain>/<file>.go`
- [ ] T013 [P] [US1] Implement [helper/utility] in `pkg/<domain>/<file>.go`
- [ ] T014 [US1] Implement [business logic] in `pkg/<domain>/<file>.go` (depends on T012, T013)
- [ ] T015 [US1] Implement CLI command in `cmd/werf/<domain>/<file>.go`
- [ ] T016 [US1] Add validation and error handling
- [ ] T017 [US1] Add logging for user story 1 operations

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently

---

## Phase 4: Additional User Stories

Repeat the User Story phase for each prioritized story. Each story MUST remain independently testable where the feature permits.

---

## Phase N: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] TXXX [P] Documentation updates in `docs/`
- [ ] TXXX Code cleanup and refactoring
- [ ] TXXX Performance optimization across all stories
- [ ] TXXX [P] Additional unit tests (if requested) in `pkg/<domain>/<file>_test.go`
- [ ] TXXX Security hardening
- [ ] TXXX [P] CLI help text generation (`task doc:gen`)
- [ ] TXXX [P] E2E tests with `paths="./test/e2e/..."` and `labelFilter="..."` (Ginkgo label filter). NEVER place `KEY=VALUE` after `--`.

---

## Final Phase: Quality Gates

- [ ] TXXX Run `task format`
- [ ] TXXX Run `task build`
- [ ] TXXX Run `task lint` after the linter prerequisite
- [ ] TXXX Run affected `task test:unit`
- [ ] TXXX Run applicable scoped E2E tests with `paths` and `labelFilter`
- [ ] TXXX Run `task test:integration` when applicable
- [ ] TXXX Run `task mock:check` when mocks changed
- [ ] TXXX Run `task doc:gen` when CLI help changed
- [ ] TXXX Run a CLI runtime smoke check when commands changed

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3+)**: All depend on Foundational phase completion
  - User stories can then proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 → P2 → P3)
- **Polish (Final Phase)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - May integrate with US1 but should be independently testable
- **User Story 3 (P3)**: Can start after Foundational (Phase 2) - May integrate with US1/US2 but should be independently testable

### Within Each User Story

- Tests (if included) MUST be written and FAIL before implementation
- Core types before business logic
- Business logic before CLI commands
- Story complete before moving to next priority

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel
- All Foundational tasks marked [P] can run in parallel (within Phase 2)
- Once Foundational phase completes, all user stories can start in parallel (if team capacity allows)
- All tests for a user story marked [P] can run in parallel
- Types within a story marked [P] can run in parallel
- Different user stories can be worked on in parallel by different team members

---

## Parallel Example: User Story 1

```bash
# Launch all tests for User Story 1 together (if tests requested):
Task: "task test:unit -- -run TestMyFunc ./pkg/<domain>/..."
Task: "task test:unit -- -run TestMyCommand ./cmd/werf/<domain>/..."

# Launch all types for User Story 1 together:
Task: "Implement core type in pkg/<domain>/<file>.go"
Task: "Implement helper in pkg/<domain>/<file>.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: Test User Story 1 independently
5. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Deploy/Demo (MVP!)
3. Add User Story 2 → Test independently → Deploy/Demo
4. Add User Story 3 → Test independently → Deploy/Demo
5. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1
   - Developer B: User Story 2
   - Developer C: User Story 3
3. Stories complete and integrate independently

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Verify tests fail before implementing
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Avoid: vague tasks, same file conflicts, cross-story dependencies that break independence
