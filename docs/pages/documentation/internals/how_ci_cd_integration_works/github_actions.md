---
title: GitHub Actions
sidebar: documentation
permalink: documentation/internals/how_ci_cd_integration_works/github_actions.html
---

According to the description of [ci-env command]({{ "documentation/internals/how_ci_cd_integration_works/general_overview.html#what-is-ci-env" | true_relative_url }}), werf should define a set of `WERF_*` variables and perform some actions to integrate with the CI/CD system.

werf uses the following values for werf environment variables:

## WERF_REPO

The value of [`WERF_REPO`]({{ "documentation/internals/how_ci_cd_integration_works/general_overview.html#werf_repo" | true_relative_url }}) is derived from the [`GITHUB_REPOSITORY`](https://docs.github.com/en/free-pro-team@latest/actions/reference/environment-variables#default-environment-variables) GitHub Actions variable (**converted to lowercase**) and project name from `werf.yaml`: `docker.pkg.github.com/$GITHUB_REPOSITORY/<project-name>-werf`.

## WERF_ADD_ANNOTATION_PROJECT_GIT

The value of [`WERF_ADD_ANNOTATION_PROJECT_GIT`]({{ "documentation/internals/how_ci_cd_integration_works/general_overview.html#werf_add_annotation_project_git" | true_relative_url }}) is based on the [`GITHUB_REPOSITORY`](https://docs.github.com/en/free-pro-team@latest/actions/reference/environment-variables#default-environment-variables) GitHub Actions environment variable and composed as follows:

```
project.werf.io/git=https://github.com/$GITHUB_REPOSITORY
```

## WERF_ADD_ANNOTATION_CI_COMMIT

The value of [`WERF_ADD_ANNOTATION_CI_COMMIT`]({{ "documentation/internals/how_ci_cd_integration_works/general_overview.html#werf_add_annotation_ci_commit" | true_relative_url }}) is extracted from the [`GITHUB_SHA`](https://docs.github.com/en/free-pro-team@latest/actions/reference/environment-variables#default-environment-variables) GitHub Actions environment variable and composed as follows:

```
ci.werf.io/commit=$GITHUB_SHA
```

## WERF_ADD_ANNOTATION_GITHUB_ACTIONS_RUN_URL

The value of `WERF_ADD_ANNOTATION_GITHUB_ACTIONS_RUN_URL` is derived from the [`GITHUB_RUN_ID`](https://docs.github.com/en/free-pro-team@latest/actions/reference/environment-variables#default-environment-variables) GitHub Actions environment variable and composed in the following way:

```
github.ci.werf.io/workflow-run-url=https://github.com/$GITHUB_REPOSITORY/actions/runs/$GITHUB_RUN_ID
```

## Other variables

Other variables are configured in the regular way described in the [overview article]({{ "documentation/internals/how_ci_cd_integration_works/general_overview.html" | true_relative_url }}):
 * [`DOCKER_CONFIG`]({{ "documentation/internals/how_ci_cd_integration_works/general_overview.html#docker_config" | true_relative_url }});
 * [`WERF_LOG_COLOR_MODE`]({{ "documentation/internals/how_ci_cd_integration_works/general_overview.html#werf_log_color_mode" | true_relative_url }});
 * [`WERF_LOG_PROJECT_DIR`]({{ "documentation/internals/how_ci_cd_integration_works/general_overview.html#werf_log_project_dir" | true_relative_url }});
 * [`WERF_ENABLE_PROCESS_EXTERMINATOR`]({{ "documentation/internals/how_ci_cd_integration_works/general_overview.html#werf_enable_process_exterminator" | true_relative_url }});
 * [`WERF_LOG_TERMINAL_WIDTH`]({{ "documentation/internals/how_ci_cd_integration_works/general_overview.html#werf_log_terminal_width" | true_relative_url }}).

## How to use

You can turn on the integration with GitHub Actions by invoking the [`werf ci-env` command]({{ "documentation/reference/cli/werf_ci_env.html" | true_relative_url }}) with the required positional argument:

```shell
werf ci-env github
```
