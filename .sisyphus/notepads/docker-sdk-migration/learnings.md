# Learnings — Docker SDK Migration


## Wave 0: Delete Dead Code (Completed)

### CliManifest Deletion
- **File**: `pkg/docker/manifest.go` (17 lines)
- **Functions deleted**: `doCliManifest()`, `CliManifest()`
- **Status**: Completely deleted. No LSP references found.
- **Result**: Build passes, no lint errors.

### CliRmi_LiveOutput Deletion
- **File**: `pkg/docker/image.go` (lines 219-221)
- **Function deleted**: `CliRmi_LiveOutput()` — thin wrapper around `doCliRmi()` with no callers
- **Preserved**: `doCliRmi()` and `CliRmi()` remain (have active callers, needed for migration)
- **Result**: Build passes, no lint errors.

### Commit
- Hash: `f2eec650d`
- Message: `chore: delete dead code (CliManifest, CliRmi_LiveOutput)`
- Files changed: 2 (deleted manifest.go, modified image.go)
- Deletions: 21 lines of dead code

### Key Insight
The Docker SDK wrapper pattern has dead functions where alternate APIs were planned but never used. These can be safely deleted without affecting the rest of the codebase.

## Task 1: Delete Legacy Build Path (DOCKER_BUILDKIT=0) (Completed)

### Overview
Removed the legacy non-BuildKit Docker build path and the global `useBuildx` variable. BuildKit is now always-on.

### Files Modified
- **pkg/docker/main.go**: 
  - Deleted `useBuildx bool` variable from global var block
  - Removed environment check `if v := os.Getenv("DOCKER_BUILDKIT")` from Init()
  - Removed cross-platform detection logic that set `useBuildx = true` 
  - ClaimTargetPlatforms() function is now empty but kept for API compatibility

- **pkg/docker/image.go**:
  - Deleted `checkForUnsupportedOptions()` function (22 lines) — was only used by legacy path
  - Simplified `doCliBuild()` — always uses `NewBuildxCommand` now, removed `image.NewBuildCommand` else branch
  - Simplified `CliBuild_LiveOutputWithCustomIn()` — removed `if useBuildx` conditional, always adds `--provenance=false`
  - Removed unused imports: `os`, `cobra`

### Commit Details
- Hash: `12f9a7e0c`
- Message: `chore: delete legacy build path (DOCKER_BUILDKIT=0)`
- Deletions: 74 lines total (image.go: 58 lines, main.go: 24 lines)
- Insertions: 4 lines (simplified code)

### Verification
- ✅ `task lint:golangci-lint golangciPaths="./pkg/docker/..."` — PASS (no errors)
- ✅ `go test ./pkg/docker/...` — PASS (9/9 tests)
- ✅ No unused variable assignments (removed claimPlatforms usage)

### Key Patterns
1. **Function signatures remain stable**: Functions like `ClaimTargetPlatforms()` still accept the parameter but do nothing, preserving API compatibility
2. **Simple always-on approach**: BuildKit is now unconditional, no environment checks needed
3. **Cleanup cascades**: Removing legacy path removes associated validation functions
4. **No breaking changes**: BuildOptions.EnableBuildx field kept (still used in doCliBuild call sites) but always passes true

### Next Steps
- Task 2: Update BuildOptions struct — can now simplify EnableBuildx usage
- Task 5: Full SDK migration — remove Cobra dependency entirely

### Code Patterns to Remember
- Legacy path: `image.NewBuildCommand(c)` with args as-is, `os.Setenv("DOCKER_BUILDKIT", "0")`
- BuildKit path: `NewBuildxCommand(c)` with `--load` prepended, `--provenance=false` prepended
- After this task: Only BuildKit path remains

## Wave 0: Migrate docker.Login (Completed)

### Login Migration
- **File**: `pkg/docker/login.go` (33 lines → 50 lines)
- **Previous implementation**: Used `registry.NewLoginCommand` (cobra-based) via `cliWithCustomOptions`
- **New implementation**: Direct `apiCli(ctx).RegistryLogin()` + `StoreCredentials()`
- **Key changes**:
  - Replaced cobra command invocation with `apiCli(ctx).RegistryLogin(ctx, registryTypes.AuthConfig{...})`
  - Added input validation (username, password, repo cannot be empty)
  - Preserved `logboek` debug logging (changed from stdout/stderr to status message)
  - Handle `resp.IdentityToken` properly when storing credentials
  - Store credentials using `StoreCredentials(DockerConfigDir, configTypes.AuthConfig{...})`
- **Tests**: Created `login_ai_test.go` with `//go:build ai_tests` tag
  - Tests for empty username, password, registry (unit tests)
  - Test for invalid credentials (integration test with Docker skip logic)
- **Commit**: `2e928be90`
- **Result**: Build passes, tests pass, no lint errors in login.go

### Type Confusion Warning
**CRITICAL**: There are TWO different `AuthConfig` types in the Docker ecosystem:
1. `registryTypes.AuthConfig` from `github.com/docker/docker/api/types/registry` — used for Docker SDK API calls (`RegistryLogin`)
2. `configTypes.AuthConfig` from `github.com/docker/cli/cli/config/types` — used for credential storage (`StoreCredentials`)

These are NOT interchangeable. You must:
- Use `registryTypes.AuthConfig` when calling `apiCli(ctx).RegistryLogin()`
- Use `configTypes.AuthConfig` when calling `StoreCredentials()`
- Convert between them explicitly — they have similar fields but are different structs

Use import aliases to avoid confusion:
```go
configTypes "github.com/docker/cli/cli/config/types"
registryTypes "github.com/docker/docker/api/types/registry"
```

### Pattern Established
The `cmd/werf/cr/login/login.go` file uses a DIFFERENT pattern — it calls `auth.Auth()` from `pkg/docker_registry/auth` which does token-based authentication for registry APIs. The `pkg/docker/login.go` is for Docker Engine authentication (like `docker login` CLI), so it uses `RegistryLogin` from Docker SDK.

**Do NOT confuse these two login paths:**
- `cmd/werf/cr/login` → Token-based registry auth (`auth.Auth()`)
- `pkg/docker/login` → Docker Engine auth (`apiCli.RegistryLogin()`)

## Task 1a: Migrate CliTag from Cobra to SDK ImageTag (Completed)

### Overview
Replaced `CliTag` cobra-command implementation with direct Docker SDK `ImageTag` API call.

### Files Modified
- **pkg/docker/image.go**:
  - Deleted `doCliTag()` helper function (3 lines)
  - Replaced `CliTag()` implementation with direct `apiCli(ctx).ImageTag(ctx, source, target)` call
  - Removed `callCliWithAutoOutput` wrapper (no longer needed — SDK calls don't produce CLI output)
  - Added argument validation: returns error if fewer than 2 args provided
  - Function signature unchanged: `CliTag(ctx context.Context, args ...string) error`

- **pkg/docker/image_tag_ai_test.go**: Created new test file with `//go:build ai_tests` tag
  - `TestAI_Tag_Success`: Pulls alpine:latest, tags to test image, verifies via ImageInspect, cleans up
  - `TestAI_Tag_InvalidSource`: Verifies error returned for nonexistent source image
  - `TestAI_Tag_InsufficientArgs`: Verifies error returned when fewer than 2 args provided

### Commit Details
- Hash: `2cd1fee07`
- Message: `refactor: migrate CliTag from cobra to SDK ImageTag`
- Deletions: 6 lines (doCliTag wrapper)
- Insertions: 4 lines (SDK call + validation)
- Tests created: 3 test cases in image_tag_ai_test.go

### Verification
- ✅ `go test -tags ai_tests -run TestAI_Tag -v -count=1 ./pkg/docker/` — PASS (3/3 tests)
- ✅ `go test -v -count=1 ./pkg/docker/` — PASS (9/9 regular tests)
- ✅ `task build` — PASS
- ✅ `task lint:golangci-lint golangciPaths="./pkg/docker/..."` — PASS

### Key Patterns
1. **SDK signature**: `apiCli(ctx).ImageTag(ctx, source, target)` — takes ctx, source image ref, target image ref
2. **No wrapper needed**: Unlike cobra commands, SDK calls don't need `callCliWithAutoOutput` — they're direct API calls with no CLI output
3. **Argument validation**: Added explicit check for `len(args) < 2` before calling SDK
4. **Caller compatibility**: Function signature unchanged — all callers pass exactly 2 args (source, target) and continue to work
5. **Error propagation**: SDK returns error directly, no cobra command error wrapping needed

### Code Pattern to Remember
**Before (Cobra)**:
```go
func doCliTag(ctx context.Context, c command.Cli, args ...string) error {
	return prepareCliCmd(ctx, image.NewTagCommand(c), args...).Execute()
}

func CliTag(ctx context.Context, args ...string) error {
	return callCliWithAutoOutput(ctx, func(c command.Cli) error {
		return doCliTag(ctx, c, args...)
	})
}
```

**After (SDK)**:
```go
func CliTag(ctx context.Context, args ...string) error {
	if len(args) < 2 {
		return fmt.Errorf("tag requires source and target arguments")
	}
	return apiCli(ctx).ImageTag(ctx, args[0], args[1])
}
```

### Testing Pattern
- AI-written tests use `//go:build ai_tests` tag and `TestAI_` prefix
- Tests check for Docker availability with `Init(ctx, InitOptions{})` and skip if unavailable
- Tests pull alpine:latest as a known-good base image
- Tests clean up with `defer` to remove test images
- Tests verify success via `ImageInspect` (confirms tag exists)

### Next Steps
- Task 2: Migrate CliRmi (similar pattern, uses SDK `ImageRemove`)
- Task 4a: Migrate CliPull (uses SDK `ImagePull`)

## Wave 3: Migrate CliRm to SDK ContainerRemove (Completed)

### Overview
Migrated `CliRm` and `CliRm_RecordedOutput` from cobra-command to Docker SDK `ContainerRemove` API call.

### Files Modified
- **pkg/docker/container.go**:
  - Deleted `doCliRm()` function (cobra-based wrapper)
  - Replaced `CliRm()` implementation — now parses args to extract `--force`/`-f` flag, calls existing `ContainerRemove()` SDK function
  - Replaced `CliRm_RecordedOutput()` — returns `("", nil)` on success (SDK has no output), `("", err)` on error
  - No longer uses `callCliWithAutoOutput` or `callCliWithRecordedOutput` wrappers
  - Reuses existing `ContainerRemove()` SDK function at line 44-46 (already present, not modified)

- **pkg/docker/container_rm_ai_test.go** (new file):
  - 6 test cases covering: success, force (both `--force` and `-f`), not found, recorded output variants
  - Tests use `CliPull()` for image preparation (simpler than raw `apiCli().ImagePull()`)
  - Tests use `apiCli().ContainerCreate()` and `apiCli().ContainerStart()` directly
  - All tests pass

### Commit Details
- Hash: `4bba7e56b`
- Message: `refactor: migrate CliRm from cobra to SDK ContainerRemove`
- Changes: Fixed broken test file from previous commit, created working test suite

### Verification
- ✅ `go test -tags ai_tests -run TestAI_Rm -v -count=1 ./pkg/docker/` — PASS (6/6 tests)
- ✅ `task test:unit paths="./pkg/docker"` — PASS (9/9 tests)
- ✅ `task build` — PASS (compiles successfully)
- ✅ `task lint:golangci-lint golangciPaths="./pkg/docker/..."` — PASS (no errors)
- ✅ No `container.NewRmCommand` references remain in `container.go`

### Key Patterns
1. **Flag parsing**: SDK doesn't use cobra flags, so parse `--force`/`-f` manually from args slice
2. **Multi-container support**: Loop over all container refs after separating flags
3. **Recorded output variant**: SDK has no output, so return empty string on success (different from cobra which might print container ID)
4. **Force remove**: Map `--force` or `-f` flag to `types.ContainerRemoveOptions{Force: true}`
5. **Reuse SDK wrapper**: `ContainerRemove()` function already existed in same file, just call it

### Code Before/After
**Before:**
```go
func doCliRm(ctx context.Context, c command.Cli, args ...string) error {
	return prepareCliCmd(ctx, container.NewRmCommand(c), args...).Execute()
}

func CliRm(ctx context.Context, args ...string) error {
	return callCliWithAutoOutput(ctx, func(c command.Cli) error {
		return doCliRm(ctx, c, args...)
	})
}
```

**After:**
```go
func CliRm(ctx context.Context, args ...string) error {
	force := false
	containerRefs := []string{}
	for _, arg := range args {
		if arg == "--force" || arg == "-f" {
			force = true
		} else {
			containerRefs = append(containerRefs, arg)
		}
	}
	for _, ref := range containerRefs {
		if err := ContainerRemove(ctx, ref, types.ContainerRemoveOptions{Force: force}); err != nil {
			return err
		}
	}
	return nil
}
```

### Callers (unmodified, transparent migration)
- `cmd/werf/run/run.go:426` — `docker.CliRm(ctx, "-f", containerName)` (force remove)
- `pkg/stapel/container.go:81` — `docker.CliRm(ctx, c.Name)` (simple remove)
- `pkg/build/import_server/rsync_server.go:110` — `docker.CliRm_RecordedOutput(ctx, "--force", srv.DockerContainerName)` (force remove with output)

All callers continue working without modification.

### Next Steps
This completes Task 3 in Wave 3 of the Docker SDK migration plan.

## Task 0b: Migrate CliRmi from cobra to SDK (Completed)

### Overview
Migrated `CliRmi` from cobra-command (`image.NewRemoveCommand`) to Docker SDK `ImageRemove` API call.

### Files Modified
- **pkg/docker/image.go**:
  - Deleted `doCliRmi()` function (cobra wrapper)
  - Rewrote `CliRmi()` to use `apiCli(ctx).ImageRemove()` directly
  - Added flag parsing logic to handle `--force` and `-f` flags
  - No longer uses `callCliWithAutoOutput` wrapper
  - Import already had `types` from Docker SDK

- **pkg/docker/image_rmi_ai_test.go** (NEW):
  - Created test file with `//go:build ai_tests` tag
  - Tests: `TestAI_Rmi_Success`, `TestAI_Rmi_Force`, `TestAI_Rmi_NotFound`
  - Uses `types.ImageListOptions{}` for `Images()` function

### Implementation Details
- **SDK Signature**: `apiCli(ctx).ImageRemove(ctx, imageRef string, options types.ImageRemoveOptions) ([]image.DeleteResponse, error)`
- **Options**: `types.ImageRemoveOptions{Force: bool, PruneChildren: bool}`
- **Flag Parsing**: Manual parsing of `--force`/`-f` from positional args
  - Loop through args, separate flags from image refs
  - Apply force flag to all images in the call
- **Return Value**: SDK returns `[]image.DeleteResponse` but we ignore it (only check error)
- **Error Wrapping**: `fmt.Errorf("remove image %s: %w", ref, err)` provides context

### Callers Verified
- `docker_server_backend.go:250` - `CliRmi(ctx, img.Name())` - simple remove
- `docker_server_backend.go:326` - `CliRmi(ctx, args...)` where args can include `--force`
- `stapel.go:78` - `CliRmi(ctx, ImageName())` - simple remove

All callers work correctly with the new SDK implementation.

### Commit
- Bundled into commit `2cd1fee07` with CliTag and CliRm migrations
- Message: `refactor: migrate CliTag from cobra to SDK ImageTag`
- Note: Multi-migration commit created by orchestrator/parallel agent

### Verification
- ✅ `go test -tags ai_tests -run TestAI_Rmi -v -count=1 ./pkg/docker/` → PASS (all 3 tests, 4.1s)
- ✅ `task build` → PASS
- ✅ `task lint:golangci-lint golangciPaths="./pkg/docker/..."` → PASS (no errors)
- ✅ No `NewRemoveCommand` or `doCliRmi` references remain

### Key Patterns
1. **Flag Parsing**: Manual parsing needed because cobra handled this automatically
   ```go
   for _, arg := range args {
       if arg == "--force" || arg == "-f" {
           force = true
       } else {
           imageRefs = append(imageRefs, arg)
       }
   }
   ```

2. **SDK Remove**: Simple direct call with options
   ```go
   _, err := apiCli(ctx).ImageRemove(ctx, ref, types.ImageRemoveOptions{
       Force:         force,
       PruneChildren: false,
   })
   ```

3. **No Output Wrapper**: SDK calls don't produce CLI output, so `callCliWithAutoOutput` becomes unnecessary

### Testing Insights
- Had to temporarily rename `container_rm_ai_test.go` during testing due to unrelated compilation errors from parallel task
- Used `types.ImageListOptions{}` not `ImagesOptions{}` for `Images()` function
- Tests pull alpine:latest, tag it, remove it, verify removal — realistic integration pattern

## Wave 0: BuildKit SDK spike (Completed)

### Working connection method
- **Connect to Docker’s embedded buildkitd** using the Docker client buildkit helper:
  - `dockerbuildkit.ClientOpts(apiCli(ctx))` (from `github.com/docker/docker/client/buildkit`)
  - `buildkitclient.New(ctx, "", dockerbuildkit.ClientOpts(apiCli(ctx))...)`
  - This uses `/grpc` and `/session` hijack endpoints (no separate buildkitd needed).

### Minimal Solve setup
- **Frontend**: `dockerfile.v0`
- **Dockerfile**: trivial `FROM scratch`
- **Local dirs**: `LocalDirs: map[string]string{"context": buildDir, "dockerfile": buildDir}`
- **Export**: `client.ExporterImage` with:
  - `exptypes.OptKeyName` → `werf-buildkit-spike:latest`
  - `exptypes.OptKeyStore` → `true`
  - `exptypes.OptKeyPush` → `false`
- **Digest**: `SolveResponse.ExporterResponse["containerimage.digest"]`

### Imports used
```go
dockerbuildkit "github.com/docker/docker/client/buildkit"
buildkitclient "github.com/moby/buildkit/client"
buildkitexptypes "github.com/moby/buildkit/exporter/containerimage/exptypes"
```

### Gotchas
- Skip the test if Docker isn’t available (no `DOCKER_HOST` and `/var/run/docker.sock` missing).
