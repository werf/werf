---
title: Automatic host cleanup
permalink: usage/cleanup/host_cleanup.html
---

## Overview

Host cleanup removes irrelevant data and reduces cache size **automatically** as part of the basic werf command invocation and **for all projects** at once. If necessary, the cleanup can be performed manually using the [**werf host cleanup**]({{"reference/cli/werf_host_cleanup.html" | true_relative_url }}) command.

## Changing the Docker storage directory

The `--docker-server-storage-path` parameter (or the `WERF_DOCKER_SERVER_STORAGE_PATH` environment variable) allows you to explicitly specify the Docker storage directory in case werf fails to detect it automatically.

## Changing the space usage threshold and cleanup depth of the Docker storage

The `--allowed-docker-storage-volume-usage` (`WERF_ALLOWED_DOCKER_STORAGE_VOLUME_USAGE`) parameter allows you to adjust the volume usage threshold (the default is 70%). Reaching it will trigger a Docker storage cleanup.

The `--allowed-docker-storage-volume-usage-margin` (`WERF_ALLOWED_DOCKER_STORAGE_VOLUME_USAGE_MARGIN`) parameter allows you to set the extra cleanup margin relative to the Docker storage usage threshold (the default is 5%).

## Changing the space usage threshold and cleanup depth of the local cache

The `--allowed-local-cache-volume-usage` (`WERF_ALLOWED_LOCAL_CACHE_VOLUME_USAGE`) parameter allows you to adjust the threshold of space used on the volume at which the local cache cleanup is triggered (the default is 70%).

The `--allowed-docker-storage-volume-usage-margin` (`WERF_ALLOWED_LOCAL_CACHE_VOLUME_USAGE_MARGIN`) parameter allows to set the cleanup margin relative to the local cache usage threshold (the default is 5%).

## Turning off automatic cleaning

The user can disable automatic cleanup of outdated host data using the `--disable-auto-host-cleanup` parameter (`WERF_DISABLE_AUTO_HOST_CLEANUP`). In this case, we recommend adding the `werf host cleanup` command to the list of cron jobs, e.g., as follows:

```shell
# /etc/cron.d/werf-host-cleanup
SHELL=/bin/bash
*/30 * * * * gitlab-runner source ~/.profile ; source $(trdl use werf 1.2 stable) ; werf host cleanup
```
