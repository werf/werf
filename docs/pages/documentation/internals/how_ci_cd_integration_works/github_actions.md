---
title: GitHub Actions
sidebar: documentation
permalink: documentation/internals/how_ci_cd_integration_works/github_actions.html
---

According to the description of [ci-env command]({{ site.baseurl }}/documentation/internals/how_ci_cd_integration_works/general_overview.html#what-is-ci-env), werf should define a set of `WERF_*` variables and perform some actions to integrate with the CI/CD system.

werf uses the following values for werf environment variables:

### WERF_REPO

The value of [`WERF_REPO`]({{ site.baseurl }}/documentation/internals/how_ci_cd_integration_works/general_overview.html#werf_repo) is derived from the [`GITHUB_REPOSITORY`](https://docs.github.com/en/free-pro-team@latest/actions/reference/environment-variables#default-environment-variables) GitHub Actions variable (**converted to lowercase**) and project name from `werf.yaml`: `docker.pkg.github.com/$GITHUB_REPOSITORY/<project-name>-werf`.

### WERF_ADD_ANNOTATION_PROJECT_GIT

The value of [`WERF_ADD_ANNOTATION_PROJECT_GIT`]({{ site.baseurl }}/documentation/internals/how_ci_cd_integration_works/general_overview.html#werf_add_annotation_project_git) is based on the [`GITHUB_REPOSITORY`](https://docs.github.com/en/free-pro-team@latest/actions/reference/environment-variables#default-environment-variables) GitHub Actions environment variable and composed as follows:

```
project.werf.io/git=https://github.com/$GITHUB_REPOSITORY
```

### WERF_ADD_ANNOTATION_CI_COMMIT

The value of [`WERF_ADD_ANNOTATION_CI_COMMIT`]({{ site.baseurl }}/documentation/internals/how_ci_cd_integration_works/general_overview.html#werf_add_annotation_ci_commit) is extracted from the [`GITHUB_SHA`](https://docs.github.com/en/free-pro-team@latest/actions/reference/environment-variables#default-environment-variables) GitHub Actions environment variable and composed as follows:

```
ci.werf.io/commit=$GITHUB_SHA
```

### WERF_ADD_ANNOTATION_GITHUB_ACTIONS_RUN_URL

The value of `WERF_ADD_ANNOTATION_GITHUB_ACTIONS_RUN_URL` is derived from the [`GITHUB_RUN_ID`](https://docs.github.com/en/free-pro-team@latest/actions/reference/environment-variables#default-environment-variables) GitHub Actions environment variable and composed in the following way:

```
github.ci.werf.io/run-url=https://github.com/$GITHUB_REPOSITORY/actions/runs/$GITHUB_RUN_ID
```

### Other variables

Other variables are configured in the regular way described in the [overview article]({{ site.baseurl }}/documentation/internals/how_ci_cd_integration_works/general_overview.html):
 * [`DOCKER_CONFIG`]({{ site.baseurl }}/documentation/internals/how_ci_cd_integration_works/general_overview.html#docker_config);
 * [`WERF_LOG_COLOR_MODE`]({{ site.baseurl }}/documentation/internals/how_ci_cd_integration_works/general_overview.html#werf_log_color_mode);
 * [`WERF_LOG_PROJECT_DIR`]({{ site.baseurl }}/documentation/internals/how_ci_cd_integration_works/general_overview.html#werf_log_project_dir);
 * [`WERF_ENABLE_PROCESS_EXTERMINATOR`]({{ site.baseurl }}/documentation/internals/how_ci_cd_integration_works/general_overview.html#werf_enable_process_exterminator);
 * [`WERF_LOG_TERMINAL_WIDTH`]({{ site.baseurl }}/documentation/internals/how_ci_cd_integration_works/general_overview.html#werf_log_terminal_width).

## How to use

You can turn on the integration with GitHub Actions by invoking the [`werf ci-env` command]({{ site.baseurl }}/documentation/reference/cli/werf_ci_env.html) with the required positional argument:

```shell
werf ci-env github
```
