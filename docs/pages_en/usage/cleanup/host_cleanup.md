---
title: Automatic host cleanup
permalink: usage/cleanup/host_cleanup.html
---

The [**werf host cleanup**]({{ "reference/cli/werf_host_cleanup.html" | true_relative_url }}) command cleans up old, unused, outdated data and reduces the cache size for all projects on the host. It uses the space occupied and user settings as a guide.

## Principle of operation

The algorithm automatically decides which data to delete. It consists of the following steps:

- Evaluating the space used on the volume where the local docker server data are located;
- If the space used exceeds the threshold, werf calculates the amount of space that needs to be freed to get the percent of occupied space back below the threshold (with an extra 5%). By default, the threshold is 70% of the volume space, and you can configure it using the relevant parameter.
- Next, the algorithm proceeds to delete the least recently used (LRU) data until the percent of occupied space goes back below the threshold. By default, the threshold is 70% and an extra 5% as a reserve (you can configure it using the relevant parameter).

What data can be deleted:
- Git archives in the local werf cache: `~/.werf/local_cache/git_archives`.
- Git patches in the local werf cache: `~/.werf/local_cache/git_patches`.
- Git repositories in the local werf cache: `~/.werf/local_cache/git_repos`.
- Git worktree in the local werf cache: `~/.werf/local_cache/git_worktrees`.
- All docker images that were built by version v1.2 and exist on the local docker server.
- Docker images that were built by version v1.1 and are stored in `--stages-storage=REPO`.
  - The algorithm cannot delete images created by version v1.1 and stored in `--stages-storage=:local` since this is the primary storage that keeps stages that can be used in production and other environments.

Note that the algorithm of the `werf host cleanup` command separately processes the volume where the local werf cache is stored (`~/.werf/local_cache`) and the volume where the local docker server data are stored (usually at `/var/lib/docker`). If werf cannot find the directory where the data of the local docker server are stored, you can specify the appropriate path explicitly via the `--docker-server-storage-path=/var/lib/docker` parameter (or via the `WERF_DOCKER_SERVER_STORAGE_PATH` environment variable).

## Turning off automatic cleaning

The user can disable auto-cleaning of outdated host data using the `--disable-auto-host-cleanup` parameter (or the respective `WERF_DISABLE_AUTO_HOST_CLEANUP` environment variable). In this case, we recommend adding the `werf host cleanup` command to the list of cron jobs, e.g., as follows:

```shell
# /etc/cron.d/werf-host-cleanup
SHELL=/bin/bash
*/30 * * * * gitlab-runner source ~/.profile ; source $(trdl use werf 1.2 stable) ; werf host cleanup
```
