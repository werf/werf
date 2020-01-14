---
title: Unsupported CI/CD integration
sidebar: documentation
permalink: documentation/guides/unsupported_ci_cd_integration.html
author: Timofey Kirillov <timofey.kirillov@flant.com>
---

werf for now only supports [GitLab CI system]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/gitlab_ci.html). Support for top-10 popular CI systems [coming soon](https://github.com/flant/werf/issues/1682).

To use werf with any CI/CD system that does not supported yet user should perform procedures described in the [what is ci-env]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#what-is-ci-env) by own script.

The behaviour of `werf ci-env` command should be resembled (without actual using of this command) prior running any werf command in the begin of CI/CD job. This is accoumplished by some actions and defining environment variables from [the list of environment variables]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#complete-list-of-ci-env-params-and-customizing).

## Ci-env procedures

### Docker registry integration

According to [Docker registry integration]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#docker-registry-integration) procedure, variables to define:
 * [`DOCKER_CONFIG`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#docker_config);
 * [`WERF_IMAGES_REPO`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#werf_images_repo).

Create temporal docker config and define images repo:

```shell
TMP_DOCKER_CONFIG=$(mktemp -d)
DOCKER_CONFIG=${DOCKER_CONFIG:-$HOME/.docker}
[[ -d "$DOCKER_CONFIG" ]] && cp -a $DOCKER_CONFIG/. $TMP_DOCKER_CONFIG
export DOCKER_CONFIG=$TMP_DOCKER_CONFIG
export WERF_IMAGES_REPO=registry.company.com/project
```

### Git integration

According to [git integration]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#git-integration) procedure, variables to define:
 * [`WERF_TAG_GIT_TAG`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#werf_tag_git_tag);
 * [`WERF_TAG_GIT_BRANCH`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#werf_tag_git_branch).

### CI/CD pipelines integration

According to [CI/CD pipelines integration]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#ci-cd-pipelines-integration) procedure, variables to define:
 * [`WERF_ADD_ANNOTATION_PROJECT_GIT`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#werf_add_annotation_project_git);
 * [`WERF_ADD_ANNOTATION_CI_COMMIT`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#werf_add_annotation_ci_commit).

### CI/CD configuration integration

According to [CI/CD configuration integration]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#ci-cd-configuration-integration) procedure, variables to define:
 * [`WERF_ENV`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#werf_env).

### Configure modes of operation in CI/CD systems

According to [configure modes of operation in CI/CD systems]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#configure-modes-of-operation-in-ci-cd-systems) procedure, variables to define:

Variables to define:
 * [`WERF_GIT_TAG_STRATEGY_LIMIT`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#werf_git_tag_strategy_limit);
 * [`WERF_GIT_TAG_STRATEGY_EXPIRY_DAYS`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#werf_git_tag_strategy_expiry_days);
 * [`WERF_LOG_COLOR_MODE`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#werf_log_color_mode);
 * [`WERF_LOG_PROJECT_DIR`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#werf_log_project_dir);
 * [`WERF_ENABLE_PROCESS_EXTERMINATOR`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#werf_enable_process_exterminator);
 * [`WERF_LOG_TERMINAL_WIDTH`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#werf_log_terminal_width).

## Ci-env script

Copy following script and place into `werf-ci-env.sh` in the root of the project:

```shell
TMP_DOCKER_CONFIG=$(mktemp -d)
DOCKER_CONFIG=${DOCKER_CONFIG:-$HOME/.docker}
[[ -d "$DOCKER_CONFIG" ]] && cp -a $DOCKER_CONFIG/. $TMP_DOCKER_CONFIG
export DOCKER_CONFIG=$TMP_DOCKER_CONFIG
export WERF_IMAGES_REPO=registry.company.com/project

docker login -u USER -p PASSWORD $WERF_IMAGES_REPO

export WERF_TAG_GIT_TAG=GIT_TAG
export WERF_TAG_GIT_BRANCH=GIT_BRANCH
export WERF_ADD_ANNOTATION_PROJECT_GIT="project.werf.io/git=https://cicd.domain.com/project/x"
export WERF_ADD_ANNOTATION_CI_COMMIT="ci.werf.io/commit=b9a1ddd366aa6a20a0fd43fb6612f349d33465ff"
export WERF_ENV=ENV
export WERF_GIT_TAG_STRATEGY_LIMIT=10
export WERF_GIT_TAG_STRATEGY_EXPIRY_DAYS=30
export WERF_LOG_COLOR_MODE=on
export WERF_LOG_PROJECT_DIR=1
export WERF_ENABLE_PROCESS_EXTERMINATOR=1
export WERF_LOG_TERMINAL_WIDTH=95
```

This script needs to be customized to your CI/CD system: change `WERF_*` environment variables values to the real ones. Consult with the following pages to get an idea and examples of how to retrieve real values for werf variables:
 * [GitLab CI integration]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/gitlab_ci.html)

Copy following script and place into `werf-ci-env-cleanup.sh`:

```shell
rm -rf $TMP_DOCKER_CONFIG
```

`werf-ci-env.sh` should be called in the beginning of every CI/CD job prior running any werf commands.
`werf-ci-env-cleanup.sh` should be called in the end of every CI/CD job.
