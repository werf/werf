---
title: Интеграция с неподдерживаемыми системами CI/CD
sidebar: documentation
permalink: documentation/guides/unsupported_ci_cd_integration.html
author: Timofey Kirillov <timofey.kirillov@flant.com>
---

В настоящий момент Werf поддерживает только работу с [GitLab CI]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/gitlab_ci.html), но мы планируем в [скором будущем](https://github.com/flant/werf/issues/1682) обеспечить поддержку top-10 популярных CI систем.

Чтобы использовать Werf с любой CI/CD системой которая пока не поддерживается, необходимо выполнить шаги описанные [здесь]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#what-is-ci-env), с помощью собственного скрипта. Запуск такого скрипта нужно производить вместо вызова команды `werf ci-env`, но поведение скрипта должно быть похожим на результат выполнения команды `werf ci-env`. Запуск скрипта также должен осуществляться перед выполнением любых других команд werf, вначале задания CI/CD.

Необходимого результата можно добиться выполнив в скрипте некоторые действия и определив ряд переменных окружения из [списка]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#complete-list-of-ci-env-params-and-customizing) (далее рассмотрим эти шаги подробнее).

## Настройка CI-окружения

### Интеграция с Docker Registry

Согласно процедуре [интеграции с Docker Registry]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#интеграция-с-docker-registry), необходимо определить следующие переменные:
 * [`DOCKER_CONFIG`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#docker_config);
 * [`WERF_IMAGES_REPO`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#werf_images_repo).

Также необходимо создать папку для временного файла конфигурации docker в рамках выполнения CI-задания. Пример:

```bash
mkdir .docker
export DOCKER_CONFIG=$(pwd)/.docker
export WERF_IMAGES_REPO=DOCKER_REGISTRY_REPO
```

### Интеграция с Git

Согласно [описанным шагам]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#git-integration) по интеграции с git, необходимо определить следующие переменные:
 * [`WERF_TAG_GIT_TAG`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#werf_tag_git_tag);
 * [`WERF_TAG_GIT_BRANCH`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#werf_tag_git_branch).

### Интеграция с CI/CD pipeline'ами

Согласно [описанным шагам]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#ci-cd-pipelines-integration) по интеграции с CI/CD pipeline, необходимо определить следующие переменные:
 * [`WERF_ADD_ANNOTATION_PROJECT_GIT`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#werf_add_annotation_project_git);
 * [`WERF_ADD_ANNOTATION_CI_COMMIT`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#werf_add_annotation_ci_commit).


### Интеграция с CI/CD процессами

Согласно [описанным шагам]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#ci-cd-configuration-integration) по интеграции с CI/CD процессами, необходимо определить следующие переменные:
 * [`WERF_ENV`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#werf_env).

### Общая интеграция с CI/CD системами

Согласно [описанным шагам]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#configure-modes-of-operation-in-ci-cd-systems) по общей интеграции с CI/CD системами, необходимо определить следующие переменные:
 * [`WERF_GIT_TAG_STRATEGY_LIMIT`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#werf_git_tag_strategy_limit);
 * [`WERF_GIT_TAG_STRATEGY_EXPIRY_DAYS`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#werf_git_tag_strategy_expiry_days);
 * [`WERF_LOG_COLOR_MODE`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#werf_log_color_mode);
 * [`WERF_LOG_PROJECT_DIR`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#werf_log_project_dir);
 * [`WERF_ENABLE_PROCESS_EXTERMINATOR`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#werf_enable_process_exterminator);
 * [`WERF_LOG_TERMINAL_WIDTH`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#werf_log_terminal_width).

## Пример скрипта настройки CI-окружения

В корневой папке проекта создайте bash-скрипт `werf-ci-env.sh` следующего содержания:

```bash
mkdir .docker
export DOCKER_CONFIG=$(pwd)/.docker
export WERF_IMAGES_REPO=DOCKER_REGISTRY_REPO
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

Исправьте скрипт для работы с вашей CI/CD системой: измените присваиваемые значения переменных `WERF_*` согласно вашему случаю. Будет полезно ознакомиться и ориентироваться на статью по [интеграции с GitLab CI]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/gitlab_ci.html), чтобы понимать, какие значения для переменных вам стоит брать.

Создайте также скрипт `werf-ci-env-cleanup.sh`, следующего содержания:

```bash
rm -rf .docker
```

Скрипт `werf-ci-env.sh` должен вызываться в начале каждого CI-задания, до вызова любой команды werf.
Скрипт `werf-ci-env-cleanup.sh` должен вызываться в конце каждого CI-задания.
