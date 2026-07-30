---
title: Migration from v2 to v3
permalink: resources/migration_from_v2_to_v3.html
toc: false
---

## Breaking changes in v3.0

Key changes:
1. Command `werf build` (and other build-triggering commands) now default `--final-images-only` to `true`, consistent with `converge`, `render`, `export`, `plan`, `lint`, and `bundle publish`. Pass `--final-images-only=false` to also build orphan non-final images not referenced by any final image.
1. Command `werf config list` now defaults `--final-images-only` to `true` and drops the deprecated `--images-only` alias. Pass `--final-images-only=false` to restore the old output.
1. The `artifact` image directive is no longer supported. Use `image` with `final: false` instead.
1. The `--synchronization`/`-S` flag and `WERF_SYNCHRONIZATION` env var are removed with no replacement; stage tags are content-addressable, so the distributed lock manager is no longer needed.
1. The `--virtual-merge`/`WERF_VIRTUAL_MERGE` flag and functionality are removed.
1. Nameless stapel images (`image: ~`) are no longer supported; give every image an explicit name.
1. Image names are validated at config load. A name consists of latin letters, digits, `_`, `.` and `-`, optionally split into slash-separated segments, and every segment must start and end with a letter or a digit. Hierarchical names such as `modules/controller` keep working; a name that is empty, contains whitespace, or has a dangling separator (`/api`, `api-`, `modules//controller` — typically a Go template that rendered to nothing) is now a config error instead of an obscure failure later in the build.
1. The stapel `import` directive's `image: NAME` field is renamed to `from: NAME`.
1. External `from:` references (stapel base image and `import.from`) must now include an explicit tag (`:TAG`) or digest (`@sha256:...`); an untagged reference is a hard config error instead of silently resolving to `:latest`.
1. `WERF_STAGED_DOCKERFILE_VERSION=v1` is no longer honored — staged Dockerfile always uses the v2 code path. This, together with unrelated digest-calculation fixes bundled into the same release, invalidates all pre-v3 stage caches (expect a one-time full rebuild after upgrading).
1. `werf-giterminism.yaml` schema validation is now strict: unknown/misspelled keys are a validation error instead of being silently ignored.
1. `werf.yaml` files with a trailing slash in an export/import `to:` path (other than `to: /`) now fail to build with an error instead of only logging a warning. Drop the trailing slash, e.g. `to: /usr/sbin/` → `to: /usr/sbin`.
1. Applying git patches in the legacy stapel builder no longer goes through `git apply`; a file modified by an earlier `install`/`beforeSetup`/`setup` command is silently overwritten instead of failing with a conflict error.
1. The `AllowMissedSecretKeyMode` v1.2 secret-key compatibility mode is removed. `bundle publish` itself still doesn't require a secret key by default (secret values are handled via a different, non-decrypting mechanism now), but the old compatibility flag/behavior is gone.
1. The `werf.io/base-image-id` image label and the corresponding `Info.ParentID` field are removed. Ancestor-tracking during cleanup now relies solely on `werf.io/parent-stage-id`.
1. The git commit ancestry check on git-stage reuse is removed together with the `WERF_DISABLE_GIT_COMMIT_ANCESTRY_CHECK` env var: a cached git stage is now reused regardless of whether its commit is an ancestor of the current one.
1. Command `werf bundle publish` now defaults `--helm-compatible-chart` to `true`: the published chart name in `Chart.yaml` is set to the last path component of the repo address, so `.Chart.Name` and template paths change. Pass `--helm-compatible-chart=false` to keep the name from your `Chart.yaml`.

Other changes:
1. `fromImage` (stapel image directive) and `import: - image:` are deprecated in favor of `from:`; they still work but emit a deprecation warning. Specifying both the old and new key together is a hard error.
1. `dependencies.image` is deprecated in favor of `dependencies.from`; both still work, but prefer `from` in new configs.
1. Hardcoded removal of `WERF_COMMIT_HASH`/`WERF_COMMIT_TIME_HUMAN`/`WERF_COMMIT_TIME_UNIX` from the imageSpec env-modification path is dropped. If your base image was built by a pre-fix werf version and still carries these vars, add them explicitly to `imageSpec.config.removeEnv`.
1. The `index <sha>..<sha>` line is dropped from git patch output. Only relevant if you parse werf's raw patch output externally.
1. The `werf synchronization` package/subsystem and public `synchronization.werf.io` dependency are gone; nothing to configure or migrate unless you ran a private synchronization server.
