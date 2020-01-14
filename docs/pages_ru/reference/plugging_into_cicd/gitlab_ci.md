---
title: GitLab CI
sidebar: documentation
permalink: documentation/reference/plugging_into_cicd/gitlab_ci.html
author: Timofey Kirillov <timofey.kirillov@flant.com>
---

Согласно общему описанию [использования ci-env переменных]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#что-такое-ci-env-переменные), для интеграции с системой CI/CD werf должен установить набор переменных `WERF_*` и выполнить некоторые действия.

Для получения данных, необходимых при интеграции с GitLab CI, werf использует переменные окружения CI-задания приведенные далее.

### WERF_IMAGES_REPO

Значение для установки переменной окружения [`WERF_IMAGES_REPO`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#werf_images_repo) формируется на основе переменной окружения GitLab [`CI_REGISTRY_IMAGE`](https://docs.gitlab.com/ee/ci/variables/).

### WERF_TAG_GIT_TAG

Значение для установки переменной окружения [`WERF_TAG_GIT_TAG`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#werf_tag_git_tag) формируется на основе переменной окружения GitLab [`CI_COMMIT_TAG`](https://docs.gitlab.com/ee/ci/variables/).

### WERF_TAG_GIT_BRANCH

Значение для установки переменной окружения [`WERF_TAG_GIT_BRANCH`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#werf_tag_git_branch) формируется на основе переменной окружения GitLab [`CI_COMMIT_REF_NAME`](https://docs.gitlab.com/ee/ci/variables/).

### WERF_ADD_ANNOTATION_PROJECT_GIT

Значение для установки переменной окружения [`WERF_ADD_ANNOTATION_PROJECT_GIT`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#werf_add_annotation_project_git) формируется на основе переменной окружения GitLab [`CI_PROJECT_URL`](https://docs.gitlab.com/ee/ci/variables/) следующим образом:

```
project.werf.io/git=$CI_PROJECT_URL
```

### WERF_ADD_ANNOTATION_CI_COMMIT

Значение для установки переменной окружения [`WERF_ADD_ANNOTATION_CI_COMMIT`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#werf_add_annotation_ci_commit) формируется на основе переменной окружения GitLab [`CI_COMMIT_SHA`](https://docs.gitlab.com/ee/ci/variables/) следующим образом:

```
ci.werf.io/commit=$CI_COMMIT_SHA
```

### WERF_ADD_ANNOTATION_GITLAB_CI_PIPELINE_URL

Значение для установки переменной окружения `WERF_ADD_ANNOTATION_GITLAB_CI_PIPELINE_URL` формируется на основе переменной окружения GitLab [`CI_PIPELINE_ID`](https://docs.gitlab.com/ee/ci/variables/) следующим образом:

```
gitlab.ci.werf.io/pipeline-url=$CI_PROJECT_URL/pipelines/$CI_PIPELINE_ID
```

### WERF_ADD_ANNOTATION_GITLAB_CI_JOB_URL

Значение для установки переменной окружения `WERF_ADD_ANNOTATION_GITLAB_CI_JOB_URL` формируется на основе переменной окружения GitLab [`CI_JOB_ID`](https://docs.gitlab.com/ee/ci/variables/) следующим образом:

```
gitlab.ci.werf.io/job-url=$CI_PROJECT_URL/-/jobs/$CI_JOB_ID
```

### WERF_ENV

В GitLab реализована [поддержка окружений](https://docs.gitlab.com/ce/ci/environments.html). werf определяет название текущего окружения из CI-задания GitLab.

Значение для установки переменной окружения [`WERF_ENV`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#werf_env) формируется на основе переменной окружения GitLab [`CI_ENVIRONMENT_SLUG`](https://docs.gitlab.com/ee/ci/variables/).

### Другие переменные

Значения остальных переменных окружения формируются стандартным способом, описанным в [соответствующей статье]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html):
 * [`DOCKER_CONFIG`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#docker_config);
 * [`WERF_GIT_TAG_STRATEGY_LIMIT`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#werf_git_tag_strategy_limit);
 * [`WERF_GIT_TAG_STRATEGY_EXPIRY_DAYS`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#werf_git_tag_strategy_expiry_days);
 * [`WERF_LOG_COLOR_MODE`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#werf_log_color_mode);
 * [`WERF_LOG_PROJECT_DIR`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#werf_log_project_dir);
 * [`WERF_ENABLE_PROCESS_EXTERMINATOR`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#werf_enable_process_exterminator);
 * [`WERF_LOG_TERMINAL_WIDTH`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#werf_log_terminal_width).

## Как использовать

Интеграция с GitLab CI включается указанием параметра `gitlab` в команде [`werf ci-env`]({{ site.baseurl }}/documentation/cli/toolbox/ci_env.html):

```shell
werf ci-env gitlab --tagging-strategy ...
```
