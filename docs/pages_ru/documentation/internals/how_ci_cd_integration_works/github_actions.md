---
title: GitHub Actions
sidebar: documentation
permalink: documentation/internals/how_ci_cd_integration_works/github_actions.html
---

Согласно общему описанию [использования ci-env переменных]({{ site.baseurl }}/documentation/internals/how_ci_cd_integration_works/general_overview.html#что-такое-ci-env-переменные), для интеграции с системой CI/CD werf должен установить набор переменных `WERF_*` и выполнить некоторые действия.

Для получения данных, необходимых при интеграции с GitHub Actions, werf использует переменные окружения CI-задания приведенные далее.

### WERF_REPO

Значение для установки переменной окружения [`WERF_REPO`]({{ site.baseurl }}/documentation/internals/how_ci_cd_integration_works/general_overview.html#werf_repo) формируется на основе строки **в нижнем регистре** из переменной окружения GitHub Actions [`GITHUB_REPOSITORY`](https://docs.github.com/en/free-pro-team@latest/actions/reference/environment-variables#default-environment-variables) и имени проекта из `werf.yaml`: `docker.pkg.github.com/$GITHUB_REPOSITORY/<project-name>-werf`.

### WERF_ADD_ANNOTATION_PROJECT_GIT

Значение для установки переменной окружения [`WERF_ADD_ANNOTATION_PROJECT_GIT`]({{ site.baseurl }}/documentation/internals/how_ci_cd_integration_works/general_overview.html#werf_add_annotation_project_git) формируется на основе переменной окружения GitHub Actions [`GITHUB_REPOSITORY`](https://docs.github.com/en/free-pro-team@latest/actions/reference/environment-variables#default-environment-variables) следующим образом:

```
project.werf.io/git=https://github.com/$GITHUB_REPOSITORY
```

### WERF_ADD_ANNOTATION_CI_COMMIT

Значение для установки переменной окружения [`WERF_ADD_ANNOTATION_CI_COMMIT`]({{ site.baseurl }}/documentation/internals/how_ci_cd_integration_works/general_overview.html#werf_add_annotation_ci_commit) формируется на основе переменной окружения GitHub Actions [`GITHUB_SHA`](https://docs.github.com/en/free-pro-team@latest/actions/reference/environment-variables#default-environment-variables) следующим образом:

```
ci.werf.io/commit=$GITHUB_SHA
```

### WERF_ADD_ANNOTATION_GITHUB_ACTIONS_RUN_URL

Значение для установки переменной окружения `WERF_ADD_ANNOTATION_GITHUB_CI_WORKFLOW_URL` формируется на основе переменной окружения GitHub Actions [`GITHUB_RUN_ID`](https://docs.github.com/en/free-pro-team@latest/actions/reference/environment-variables#default-environment-variables) следующим образом:

```
github.ci.werf.io/run-url=https://github.com/$GITHUB_REPOSITORY/actions/runs/$GITHUB_RUN_ID
```

### Другие переменные

Значения остальных переменных окружения формируются стандартным способом, описанным в [соответствующей статье]({{ site.baseurl }}/documentation/internals/how_ci_cd_integration_works/general_overview.html):
 * [`DOCKER_CONFIG`]({{ site.baseurl }}/documentation/internals/how_ci_cd_integration_works/general_overview.html#docker_config);
 * [`WERF_LOG_COLOR_MODE`]({{ site.baseurl }}/documentation/internals/how_ci_cd_integration_works/general_overview.html#werf_log_color_mode);
 * [`WERF_LOG_PROJECT_DIR`]({{ site.baseurl }}/documentation/internals/how_ci_cd_integration_works/general_overview.html#werf_log_project_dir);
 * [`WERF_ENABLE_PROCESS_EXTERMINATOR`]({{ site.baseurl }}/documentation/internals/how_ci_cd_integration_works/general_overview.html#werf_enable_process_exterminator);
 * [`WERF_LOG_TERMINAL_WIDTH`]({{ site.baseurl }}/documentation/internals/how_ci_cd_integration_works/general_overview.html#werf_log_terminal_width).

## Как использовать

Интеграция с GitHub Actions включается указанием параметра `github` в команде [`werf ci-env`]({{ site.baseurl }}/documentation/reference/cli/werf_ci_env.html):

```shell
werf ci-env github
```
