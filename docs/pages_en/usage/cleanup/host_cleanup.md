---
title: Automatic host cleanup
permalink: usage/cleanup/host_cleanup.html
---

## Overview

Host cleanup removes irrelevant data and reduces cache size **automatically** as part of the basic werf command invocation and **for all projects** at once. If necessary, the cleanup can be performed manually using the [**werf host cleanup**]({{"reference/cli/werf_host_cleanup.html" | true_relative_url }}) command.

## Changing build backend (Docker or BuildKit) storage directory

The `--backend-storage-path` parameter (or the `WERF_BACKEND_STORAGE_PATH` environment variable) allows you to explicitly specify the backend storage directory in case werf fails to detect it automatically.

## Changing the space usage threshold and cleanup depth of build backend (Docker or BuildKit) storage

The `--allowed-backend-storage-volume-usage` (`WERF_ALLOWED_BACKEND_STORAGE_VOLUME_USAGE`) parameter allows you to adjust the volume usage threshold (the default is 70%). Reaching it will trigger a backend storage cleanup. You can specify the value as a percentage (e.g., `70`) or in absolute units (e.g., `50GB`, `100GiB`).

The `--allowed-backend-storage-volume-usage-margin` (`WERF_ALLOWED_BACKEND_STORAGE_VOLUME_USAGE_MARGIN`) parameter allows you to set the extra cleanup margin relative to the backend storage usage threshold (the default is 5%).

> **Note:** Options within the same group (e.g., usage and margin for backend storage) must use the same units. Mixing percentages and absolute units (e.g., `--allowed-backend-storage-volume-usage=100GB --allowed-backend-storage-volume-usage-margin=5`) is not allowed.

## Cleaning a build backend cache (Docker or BuildKit)

When cleaning up the host, werf removes containers, images, and unused volumes. However, the strategy for cleaning the build cache of the build backend is left to the user.

For example, when using the Docker backend, the following approaches can be used:

- Set a cache size limit in the [builder configuration](https://docs.docker.com/build/cache/garbage-collection/#configuration).
- Configure periodic cleanup using `cron` and the `docker buildx prune --max-used-space=bytes` command.

Similarly, users can set up a custom cleanup strategy for BuildKit by pruning the buildkitd build cache (e.g., `buildctl prune`) on a schedule.

## Changing the space usage threshold and cleanup depth of the local cache

The `--allowed-local-cache-volume-usage` (`WERF_ALLOWED_LOCAL_CACHE_VOLUME_USAGE`) parameter allows you to adjust the threshold of space used on the volume at which the local cache cleanup is triggered (the default is 70%). You can specify the value as a percentage (e.g., `70`) or in absolute units (e.g., `10GB`, `500MiB`).

The `--allowed-local-cache-volume-usage-margin` (`WERF_ALLOWED_LOCAL_CACHE_VOLUME_USAGE_MARGIN`) parameter allows to set the cleanup margin relative to the local cache usage threshold (the default is 5%).

> **Note:** Options within the same group (e.g., usage and margin for local cache) must use the same units. Mixing percentages and absolute units (e.g., `--allowed-local-cache-volume-usage=10GB --allowed-local-cache-volume-usage-margin=5`) is not allowed.

## Turning off automatic cleaning

The user can disable automatic cleanup of outdated host data using the `--disable-auto-host-cleanup` parameter (`WERF_DISABLE_AUTO_HOST_CLEANUP`). In this case, we recommend adding the `werf host cleanup` command to the list of cron jobs, e.g., as follows:

```shell
# /etc/cron.d/werf-host-cleanup
SHELL=/bin/bash
*/30 * * * * gitlab-runner source ~/.profile ; source $(trdl use werf 2 stable) ; werf host cleanup
```
