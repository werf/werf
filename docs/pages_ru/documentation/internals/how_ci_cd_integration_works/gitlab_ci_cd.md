---
title: GitLab CI/CD
sidebar: documentation
permalink: documentation/internals/how_ci_cd_integration_works/gitlab_ci_cd.html
---

Согласно общему описанию [использования ci-env переменных]({{ "documentation/internals/how_ci_cd_integration_works/general_overview.html#что-такое-ci-env-переменные" | true_relative_url: page.url }}), для интеграции с системой CI/CD werf должен установить набор переменных `WERF_*` и выполнить некоторые действия.

Для получения данных, необходимых при интеграции с GitLab CI, werf использует переменные окружения CI-задания приведенные далее.

## WERF_REPO

Значение для установки переменной окружения [`WERF_REPO`]({{ "documentation/internals/how_ci_cd_integration_works/general_overview.html#werf_repo" | true_relative_url: page.url }}) формируется на основе переменной окружения GitLab [`CI_REGISTRY_IMAGE`](https://docs.gitlab.com/ee/ci/variables/).

## WERF_REPO_IMPLEMENTATION

Значение `gitlab` для установки переменной окружения [`WERF_REPO_IMPLEMENTATION`]({{ "documentation/internals/how_ci_cd_integration_works/general_overview.html#werf_repo_implementation" | true_relative_url: page.url }}) проставляется вдобавок к [`WERF_REPO`](#werf_repo) при использовании встроенного GitLab Container Registry.

## WERF_ADD_ANNOTATION_PROJECT_GIT

Значение для установки переменной окружения [`WERF_ADD_ANNOTATION_PROJECT_GIT`]({{ "documentation/internals/how_ci_cd_integration_works/general_overview.html#werf_add_annotation_project_git" | true_relative_url: page.url }}) формируется на основе переменной окружения GitLab [`CI_PROJECT_URL`](https://docs.gitlab.com/ee/ci/variables/) следующим образом:

```
project.werf.io/git=$CI_PROJECT_URL
```

## WERF_ADD_ANNOTATION_CI_COMMIT

Значение для установки переменной окружения [`WERF_ADD_ANNOTATION_CI_COMMIT`]({{ "documentation/internals/how_ci_cd_integration_works/general_overview.html#werf_add_annotation_ci_commit" | true_relative_url: page.url }}) формируется на основе переменной окружения GitLab [`CI_COMMIT_SHA`](https://docs.gitlab.com/ee/ci/variables/) следующим образом:

```
ci.werf.io/commit=$CI_COMMIT_SHA
```

## WERF_ADD_ANNOTATION_GITLAB_CI_PIPELINE_URL

Значение для установки переменной окружения `WERF_ADD_ANNOTATION_GITLAB_CI_PIPELINE_URL` формируется на основе переменной окружения GitLab [`CI_PIPELINE_ID`](https://docs.gitlab.com/ee/ci/variables/) следующим образом:

```
gitlab.ci.werf.io/pipeline-url=$CI_PROJECT_URL/pipelines/$CI_PIPELINE_ID
```

## WERF_ADD_ANNOTATION_GITLAB_CI_JOB_URL

Значение для установки переменной окружения `WERF_ADD_ANNOTATION_GITLAB_CI_JOB_URL` формируется на основе переменной окружения GitLab [`CI_JOB_ID`](https://docs.gitlab.com/ee/ci/variables/) следующим образом:

```
gitlab.ci.werf.io/job-url=$CI_PROJECT_URL/-/jobs/$CI_JOB_ID
```

## WERF_ENV

В GitLab реализована [поддержка окружений](https://docs.gitlab.com/ce/ci/environments.html). werf определяет название текущего окружения из CI-задания GitLab.

Значение для установки переменной окружения [`WERF_ENV`]({{ "documentation/internals/how_ci_cd_integration_works/general_overview.html#werf_env" | true_relative_url: page.url }}) формируется на основе переменной окружения GitLab [`CI_ENVIRONMENT_SLUG`](https://docs.gitlab.com/ee/ci/variables/).

## Другие переменные

Значения остальных переменных окружения формируются стандартным способом, описанным в [соответствующей статье]({{ "documentation/internals/how_ci_cd_integration_works/general_overview.html" | true_relative_url: page.url }}):
 * [`DOCKER_CONFIG`]({{ "documentation/internals/how_ci_cd_integration_works/general_overview.html#docker_config" | true_relative_url: page.url }});
 * [`WERF_LOG_COLOR_MODE`]({{ "documentation/internals/how_ci_cd_integration_works/general_overview.html#werf_log_color_mode" | true_relative_url: page.url }});
 * [`WERF_LOG_PROJECT_DIR`]({{ "documentation/internals/how_ci_cd_integration_works/general_overview.html#werf_log_project_dir" | true_relative_url: page.url }});
 * [`WERF_ENABLE_PROCESS_EXTERMINATOR`]({{ "documentation/internals/how_ci_cd_integration_works/general_overview.html#werf_enable_process_exterminator" | true_relative_url: page.url }});
 * [`WERF_LOG_TERMINAL_WIDTH`]({{ "documentation/internals/how_ci_cd_integration_works/general_overview.html#werf_log_terminal_width" | true_relative_url: page.url }}).

## Как использовать

Интеграция с GitLab CI включается указанием параметра `gitlab` в команде [`werf ci-env`]({{ "documentation/reference/cli/werf_ci_env.html" | true_relative_url: page.url }}):

```shell
werf ci-env gitlab
```
