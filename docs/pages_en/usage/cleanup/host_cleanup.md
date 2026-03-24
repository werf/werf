---
title: Automatic host cleanup
permalink: usage/cleanup/host_cleanup.html
---

## Overview

Host cleanup removes irrelevant data and reduces cache size **automatically** as part of the basic werf command invocation and **for all projects** at once. If necessary, the cleanup can be performed manually using the [**werf host cleanup**]({{"reference/cli/werf_host_cleanup.html" | true_relative_url }} ) command.

## Changing build backend (Docker or Buildah) storage directory

The `--backend-storage-path` parameter (or the `WERF_BACKEND_STORAGE_PATH` environment variable) allows you to explicitly specify the backend storage directory in case werf fails to detect it automatically.

## Changing the space usage threshold and cleanup depth of build backend (Docker or Buildah) storage

The `--allowed-backend-storage-volume-usage` (`WERF_ALLOWED_BACKEND_STORAGE_VOLUME_USAGE`) parameter allows you to adjust the backend cleanup threshold.

- Plain numbers are treated as **percentages**. For example, `70` means cleanup starts when backend storage usage exceeds `70%`.
- Values with units are treated as an **absolute free space threshold**. For example, `10GB` means cleanup starts when free space in backend storage drops below `10GB`.

The `--allowed-backend-storage-volume-usage-margin` (`WERF_ALLOWED_BACKEND_STORAGE_VOLUME_USAGE_MARGIN`) parameter controls how far cleanup should continue after it has been triggered.

- In percentage mode, the margin is also a percentage. For example, `70` with margin `5` means cleanup starts above `70%` usage and continues until usage drops below `65%`.
- In bytes mode, the margin is also an absolute value. For example, `10GB` with margin `2GB` means cleanup starts below `10GB` free space and continues until free space reaches `12GB`.

If `--allowed-backend-storage-volume-usage` is set with units and `--allowed-backend-storage-volume-usage-margin` is not set explicitly, werf uses `0B` as the effective margin.

If both parameters are set explicitly, they must use the same format:

- `70` and `5` — valid;
- `10GB` and `2GB` — valid;
- `70` and `2GB` — invalid;
- `10GB` and `5` — invalid.

## Cleaning a build backend cache (Docker or Buildah)

When cleaning up the host, werf removes containers, images, and unused volumes. However, the strategy for cleaning the build cache of the build backend is left to the user.

For example, when using the Docker backend, the following approaches can be used:

- Set a cache size limit in the [builder configuration](https://docs.docker.com/build/cache/garbage-collection/#configuration).
- Configure periodic cleanup using `cron` and the `docker buildx prune --max-used-space=bytes` command.

Similarly, users can set up a custom cleanup strategy for Buildah.

## Changing the space usage threshold and cleanup depth of the local cache

The `--allowed-local-cache-volume-usage` (`WERF_ALLOWED_LOCAL_CACHE_VOLUME_USAGE`) parameter allows you to adjust the local cache cleanup threshold.

- Plain numbers are treated as **percentages**. For example, `70` means cleanup starts when local cache usage exceeds `70%`.
- Values with units are treated as an **absolute free space threshold**. For example, `10GB` means cleanup starts when free space in the local cache volume drops below `10GB`.

The `--allowed-local-cache-volume-usage-margin` (`WERF_ALLOWED_LOCAL_CACHE_VOLUME_USAGE_MARGIN`) parameter controls how far cleanup should continue after it has been triggered.

- In percentage mode, the margin is also a percentage. For example, `70` with margin `5` means cleanup starts above `70%` usage and continues until usage drops below `65%`.
- In bytes mode, the margin is also an absolute value. For example, `10GB` with margin `2GB` means cleanup starts below `10GB` free space and continues until free space reaches `12GB`.

If `--allowed-local-cache-volume-usage` is set with units and `--allowed-local-cache-volume-usage-margin` is not set explicitly, werf uses `0B` as the effective margin.

If both parameters are set explicitly, they must use the same format:

- `70` and `5` — valid;
- `10GB` and `2GB` — valid;
- `70` and `2GB` — invalid;
- `10GB` and `5` — invalid.

## Turning off automatic cleaning

The user can disable automatic cleanup of outdated host data using the `--disable-auto-host-cleanup` parameter (`WERF_DISABLE_AUTO_HOST_CLEANUP`). In this case, we recommend adding the `werf host cleanup` command to the list of cron jobs, e.g., as follows:

```shell
# /etc/cron.d/werf-host-cleanup
SHELL=/bin/bash
*/30 * * * * gitlab-runner source ~/.profile ; source $(trdl use werf 2 stable) ; werf host cleanup
```
