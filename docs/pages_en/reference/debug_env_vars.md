---
title: Debug environment variables
permalink: reference/debug_env_vars.html
---

werf provides a set of environment variables that enable additional debug output. They are intended for troubleshooting: their output format is not part of the stable interface and may change between releases.

## Log levels

| Variable | Value | Description |
|----------|-------|-------------|
| `WERF_LOG_DEBUG` (alias: `WERF_DEBUG`) | boolean | Same as `--log-debug`: enable the debug log level |
| `WERF_LOG_VERBOSE` (alias: `WERF_VERBOSE`) | boolean | Same as `--log-verbose`: enable the verbose log level |

## Topic-specific debug channels

The variables below enable narrow debug channels. Channels that write to the debug log level additionally require `--log-debug` (or `WERF_LOG_DEBUG=1`) to be visible — setting the topic variable alone is not enough. Unless stated otherwise, a variable is enabled by setting it to `1`.

### Build

| Variable | Description |
|----------|-------------|
| `WERF_DEBUG_STAGES_STORAGE` | Stages storage internals: storage method call traces, manifest cache operations, per-tag stage listings, selected stage description dumps. Requires `--log-debug` |
| `WERF_DEBUG_CONVEYOR_PHASES` | Build conveyor phase wrappers (BeforeImages, OnImageStage, AfterImages, etc.). Requires `--log-debug` |
| `WERF_DEBUG_IMAGE_SPEC` | YAML snapshots of the image spec configuration (source image info, prepared and pushed config). Requires `--log-debug` |
| `WERF_DEBUG_IMPORT_SERVER` | Rsync import server traces. Requires `--log-debug` |
| `WERF_DEBUG_USER_STAGE_CHECKSUM` | User stage checksum calculation details. Requires `--log-debug` |
| `WERF_DEBUG_IMPORT_SOURCE_CHECKSUM` | Import source checksum calculation details. Requires `--log-debug` |
| `WERF_DEBUG_DOCKERFILE_STAGE_DEPENDENCIES` | Controls two channels: Dockerfile stage dependency calculation and per-match build context paths during context checksum calculation. Requires `--log-debug` |
| `WERF_CONTAINER_RUNTIME_DEBUG` | Container backend internals: container path removal, buildah container mount/unmount, per-file checksum calculation. Requires `--log-debug` |
| `WERF_BUILDAH_DEBUG` | Buildah backend debug output |
| `WERF_DEBUG_DOCKER` | Docker CLI debug mode |
| `WERF_DEBUG_DOCKER_RUN_COMMAND` | Print full `docker run` commands |

### Git

| Variable | Description |
|----------|-------------|
| `WERF_DEBUG_TRUE_GIT` | Git command execution tracing. Requires `--log-debug` |
| `WERF_TRUE_GIT_DEBUG_ARCHIVE` | Per-file entries when creating git archives. Requires `--log-debug` |
| `WERF_TRUE_GIT_DEBUG_PATCH` | Git patch creation debug output |
| `WERF_TRUE_GIT_DEBUG_PATCH_PARSER` | Git patch parser debug output |
| `WERF_DEBUG_LS_TREE_PROCESS` | Git ls-tree traversal and per-file checksum entries. Requires `--log-debug` |
| `WERF_DEBUG_GIT_STATUS` | Git status result details. Requires `--log-debug` |
| `WERF_DEBUG_GITERMINISM_MANAGER` | Giterminism manager file reader operations. Requires `--log-debug` |

### Container registry

| Variable | Description |
|----------|-------------|
| `WERF_DOCKER_REGISTRY_DEBUG` | Registry API call tracing and push/pull progress ticks. Requires `--log-debug` for the progress ticks |
| `WERF_DEBUG_DOCKER_REGISTRY_API` | Debug logs of the underlying go-containerregistry library |

### Configuration and templates

| Variable | Description |
|----------|-------------|
| `WERF_DEBUG_TEMPLATES` | boolean; same as `--debug-templates`: debug mode for Go templates |
| `WERF_DEBUG_SECRET_VALUES` | Debug output of secret values decoding. Requires `--log-debug` |

### Deploy

| Variable | Description |
|----------|-------------|
| `WERF_NELM_TRACE` | boolean; enable trace-level logging of the Nelm deployment engine |
| `WERF_HELM_V3_EXTRA_ANNOTATIONS_AND_LABELS_DEBUG` | Debug output of the extra annotations and labels post-renderer |
| `WERF_SHOW_VERBOSE_DIFFS` | boolean, default `true`; same as `--show-verbose-diffs` in `werf plan`: show verbose diff lines |
| `WERF_SHOW_VERBOSE_CRD_DIFFS` | boolean, default `false`; same as `--show-verbose-crd-diffs` in `werf plan`: show verbose CRD diff lines |

### Ansible builder

| Variable | Description |
|----------|-------------|
| `WERF_DEBUG_ANSIBLE_ARGS` | String with extra arguments passed to `ansible-playbook` |
| `WERF_DEBUG_ANSIBLE_LIVE_PY_PATH` | Path to a file overriding the built-in `live.py` ansible callback |
| `WERF_DEBUG_ANSIBLE_WERF_PY_PATH` | Path to a file overriding the built-in `werf.py` ansible callback |

### Process diagnostics

| Variable | Description |
|----------|-------------|
| `WERF_PRINT_STACK_TRACES` | Periodically print goroutine stack traces |
| `WERF_PRINT_STACK_TRACES_PERIOD` | Integer number of seconds between stack trace prints (default `5`); only takes effect with `WERF_PRINT_STACK_TRACES=1` |
