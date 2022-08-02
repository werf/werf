---
title: Generic CI/CD integration
permalink: advanced/ci_cd/generic_ci_cd_integration.html
---

Currently, the following CI systems are officially supported and fully tested to be used with werf:
 * [GitLab CI]({{ "advanced/ci_cd/gitlab_ci_cd.html" | true_relative_url }});
 * [GitHub Actions]({{ "advanced/ci_cd/github_actions.html" | true_relative_url }}).

Please refer to relevant guides if you're using one of them. This list will be extended with other CI systems. If you are particularly interested in any of them, please let us know via [this issue](https://github.com/werf/werf/issues/1617).

In general, to integrate werf with any CI/CD system you need to prepare a script following the guidelines from "[What is ci-env?]({{ "internals/how_ci_cd_integration_works/general_overview.html#what-is-ci-env" | true_relative_url }})". This script will be used instead of `werf ci-env` command. It should be executed in the beginning of your CI/CD job, prior to running any werf commands.

Below, we will outline the most important things to consider while you're creating such a script to integrate with your CI system.

## Ci-env procedures

### Container registry integration

According to [container registry integration]({{ "internals/how_ci_cd_integration_works/general_overview.html#container-registry-integration" | true_relative_url }}) procedure, variables to define:
 * [`DOCKER_CONFIG`]({{ "internals/how_ci_cd_integration_works/general_overview.html#docker_config" | true_relative_url }});
 * [`WERF_REPO`]({{ "internals/how_ci_cd_integration_works/general_overview.html#werf_repo" | true_relative_url }}).

Create temporal docker config and define repo:

```shell
TMP_DOCKER_CONFIG=$(mktemp -d)
DOCKER_CONFIG=${DOCKER_CONFIG:-$HOME/.docker}
[[ -d "$DOCKER_CONFIG" ]] && cp -a $DOCKER_CONFIG/. $TMP_DOCKER_CONFIG
export DOCKER_CONFIG=$TMP_DOCKER_CONFIG
export WERF_REPO=registry.company.com/project/werf
```

### CI/CD pipelines integration

According to [CI/CD pipelines integration]({{ "internals/how_ci_cd_integration_works/general_overview.html#cicd-pipelines-integration" | true_relative_url }}) procedure, variables to define:
 * [`WERF_ADD_ANNOTATION_PROJECT_GIT`]({{ "internals/how_ci_cd_integration_works/general_overview.html#werf_add_annotation_project_git" | true_relative_url }});
 * [`WERF_ADD_ANNOTATION_CI_COMMIT`]({{ "internals/how_ci_cd_integration_works/general_overview.html#werf_add_annotation_ci_commit" | true_relative_url }}).

### CI/CD configuration integration

According to [CI/CD configuration integration]({{ "internals/how_ci_cd_integration_works/general_overview.html#cicd-configuration-integration" | true_relative_url }}) procedure, variables to define:
 * [`WERF_ENV`]({{ "internals/how_ci_cd_integration_works/general_overview.html#werf_env" | true_relative_url }}).

### Configure modes of operation in CI/CD systems

According to [configure modes of operation in CI/CD systems]({{ "advanced/ci_cd/generic_ci_cd_integration.html#configure-modes-of-operation-in-cicd-systems" | true_relative_url }}) procedure, variables to define:

Variables to define:
 * [`WERF_LOG_COLOR_MODE`]({{ "internals/how_ci_cd_integration_works/general_overview.html#werf_log_color_mode" | true_relative_url }});
 * [`WERF_LOG_PROJECT_DIR`]({{ "internals/how_ci_cd_integration_works/general_overview.html#werf_log_project_dir" | true_relative_url }});
 * [`WERF_ENABLE_PROCESS_EXTERMINATOR`]({{ "internals/how_ci_cd_integration_works/general_overview.html#werf_enable_process_exterminator" | true_relative_url }});
 * [`WERF_LOG_TERMINAL_WIDTH`]({{ "internals/how_ci_cd_integration_works/general_overview.html#werf_log_terminal_width" | true_relative_url }}).

## Ci-env script

Create `werf-ci-env.sh` in the root directory of your project and make it look like this:

```shell
TMP_DOCKER_CONFIG=$(mktemp -d)
DOCKER_CONFIG=${DOCKER_CONFIG:-$HOME/.docker}
[[ -d "$DOCKER_CONFIG" ]] && cp -a $DOCKER_CONFIG/. $TMP_DOCKER_CONFIG
export DOCKER_CONFIG=$TMP_DOCKER_CONFIG
export WERF_REPO=registry.company.com/project/werf

docker login -u USER -p PASSWORD $WERF_REPO

export WERF_ADD_ANNOTATION_PROJECT_GIT="project.werf.io/git=https://cicd.domain.com/project/x"
export WERF_ADD_ANNOTATION_CI_COMMIT="ci.werf.io/commit=b9a1ddd366aa6a20a0fd43fb6612f349d33465ff"
export WERF_ENV=ENV
export WERF_LOG_COLOR_MODE=on
export WERF_LOG_PROJECT_DIR=1
export WERF_ENABLE_PROCESS_EXTERMINATOR=1
export WERF_LOG_TERMINAL_WIDTH=95
```

> This script needs to be customized to your CI/CD system: change `WERF_*` environment variables values to the real ones. To get an idea and examples of how you can get these real values, please have a look at our "[GitLab CI integration]({{ "internals/how_ci_cd_integration_works/gitlab_ci_cd.html" | true_relative_url }})" reference article.

Copy the following script and place into `werf-ci-env-cleanup.sh`:

```shell
rm -rf $TMP_DOCKER_CONFIG
```

`werf-ci-env.sh` should be called in the beginning of every CI/CD job, prior to running any werf commands.
`werf-ci-env-cleanup.sh` should be called in the end of every CI/CD job.
