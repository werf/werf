---
title: Generic CI/CD systems integration
sidebar: documentation
permalink: guides/generic_ci_cd_integration.html
author: Timofey Kirillov <timofey.kirillov@flant.com>
---

Currently, the following CI systems are officialy supported and fully tested to be used with werf:
 * [GitLab CI]({{ site.baseurl }}/guides/gitlab_ci_cd_integration.html);
 * [GitHub Actions]({{ site.baseurl }}/guides/github_ci_cd_integration.html).

Please refer to relevant guides if you're using one of them. This list will be extended with other CI systems. If you are particularly interested in any of them, please let us know via [this issue](https://github.com/werf/werf/issues/1617).

In general, to integrate werf with any CI/CD system you need to prepare a script following the guidelines from "[What is ci-env?]({{ site.baseurl }}/reference/plugging_into_cicd/overview.html#what-is-ci-env)". This script will be used instead of `werf ci-env` command. It should be executed in the beginning of your CI/CD job, prior to running any werf commands.

Below, we will outline the most important things to consider while you're creating such a script to integrate with your CI system.

## Ci-env procedures

### Docker registry integration

According to [Docker registry integration]({{ site.baseurl }}/reference/plugging_into_cicd/overview.html#docker-registry-integration) procedure, variables to define:
 * [`DOCKER_CONFIG`]({{ site.baseurl }}/reference/plugging_into_cicd/overview.html#docker_config);
 * [`WERF_IMAGES_REPO`]({{ site.baseurl }}/reference/plugging_into_cicd/overview.html#werf_images_repo).

Create temporal docker config and define images repo:

```shell
TMP_DOCKER_CONFIG=$(mktemp -d)
DOCKER_CONFIG=${DOCKER_CONFIG:-$HOME/.docker}
[[ -d "$DOCKER_CONFIG" ]] && cp -a $DOCKER_CONFIG/. $TMP_DOCKER_CONFIG
export DOCKER_CONFIG=$TMP_DOCKER_CONFIG
export WERF_IMAGES_REPO=registry.company.com/project
```

### Git integration

According to [git integration]({{ site.baseurl }}/reference/plugging_into_cicd/overview.html#git-integration) procedure, variables to define:
 * [`WERF_TAG_GIT_TAG`]({{ site.baseurl }}/reference/plugging_into_cicd/overview.html#werf_tag_git_tag);
 * [`WERF_TAG_GIT_BRANCH`]({{ site.baseurl }}/reference/plugging_into_cicd/overview.html#werf_tag_git_branch).

### CI/CD pipelines integration

According to [CI/CD pipelines integration]({{ site.baseurl }}/reference/plugging_into_cicd/overview.html#cicd-pipelines-integration) procedure, variables to define:
 * [`WERF_ADD_ANNOTATION_PROJECT_GIT`]({{ site.baseurl }}/reference/plugging_into_cicd/overview.html#werf_add_annotation_project_git);
 * [`WERF_ADD_ANNOTATION_CI_COMMIT`]({{ site.baseurl }}/reference/plugging_into_cicd/overview.html#werf_add_annotation_ci_commit).

### CI/CD configuration integration

According to [CI/CD configuration integration]({{ site.baseurl }}/reference/plugging_into_cicd/overview.html#cicd-configuration-integration) procedure, variables to define:
 * [`WERF_ENV`]({{ site.baseurl }}/reference/plugging_into_cicd/overview.html#werf_env).

### Configure modes of operation in CI/CD systems

According to [configure modes of operation in CI/CD systems]({{ site.baseurl }}/reference/plugging_into_cicd/overview.html#configure-modes-of-operation-in-cicd-systems) procedure, variables to define:

Variables to define:
 * [`WERF_GIT_TAG_STRATEGY_LIMIT`]({{ site.baseurl }}/reference/plugging_into_cicd/overview.html#werf_git_tag_strategy_limit);
 * [`WERF_GIT_TAG_STRATEGY_EXPIRY_DAYS`]({{ site.baseurl }}/reference/plugging_into_cicd/overview.html#werf_git_tag_strategy_expiry_days);
 * [`WERF_LOG_COLOR_MODE`]({{ site.baseurl }}/reference/plugging_into_cicd/overview.html#werf_log_color_mode);
 * [`WERF_LOG_PROJECT_DIR`]({{ site.baseurl }}/reference/plugging_into_cicd/overview.html#werf_log_project_dir);
 * [`WERF_ENABLE_PROCESS_EXTERMINATOR`]({{ site.baseurl }}/reference/plugging_into_cicd/overview.html#werf_enable_process_exterminator);
 * [`WERF_LOG_TERMINAL_WIDTH`]({{ site.baseurl }}/reference/plugging_into_cicd/overview.html#werf_log_terminal_width).

## Ci-env script

Create `werf-ci-env.sh` in the root directory of your project and make it look like this:

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

> This script needs to be customized to your CI/CD system: change `WERF_*` environment variables values to the real ones. To get an idea and examples of how you can get these real values, please have a look at our "[GitLab CI integration]({{ site.baseurl }}/reference/plugging_into_cicd/gitlab_ci.html)" reference article.

Copy the following script and place into `werf-ci-env-cleanup.sh`:

```shell
rm -rf $TMP_DOCKER_CONFIG
```

`werf-ci-env.sh` should be called in the beginning of every CI/CD job, prior to running any werf commands.
`werf-ci-env-cleanup.sh` should be called in the end of every CI/CD job.
