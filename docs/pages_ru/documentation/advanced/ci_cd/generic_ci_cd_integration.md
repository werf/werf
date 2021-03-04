---
title: Интеграция с CI/CD-системами
permalink: documentation/advanced/ci_cd/generic_ci_cd_integration.html
---

В настоящий момент официально поддерживается и полностью протестирована работа werf со следующими CI-системами:
 * [GitLab CI/CD]({{ "documentation/advanced/ci_cd/gitlab_ci_cd.html" | true_relative_url }});
 * [GitHub Actions]({{ "documentation/advanced/ci_cd/github_actions.html" | true_relative_url }}).

Если вы используете одну из них — обращайтесь к соответствующим руководствам. В дальнейшем этот список будет пополняться и другими CI-решениями. Если вы заинтересованы в поддержке конкретной системы — пожалуйста, дайте нам знать в [этом issue](https://github.com/werf/werf/issues/1617).

В общем случае, для интеграции werf с CI/CD-системой требуется написать собственный скрипт в соответствии с инструкциями из документации «[Что такое переменные ci-env?]({{ "documentation/internals/how_ci_cd_integration_works/general_overview.html#что-такое-ci-env-переменные" | true_relative_url }})». Получившийся скрипт необходимо использовать вместо вызова команды `werf ci-env`. Так же, как и команда, скрипт должен выполняться перед использованием команд werf в начале задания CI/CD.

В этом руководстве мы пошагово пройдём по основным моментам, которые необходимо учесть при создании такого скрипта интеграции с CI.

## Настройка CI-окружения

### Интеграция с Docker registry

Согласно процедуре [интеграции с Docker registry]({{ "documentation/internals/how_ci_cd_integration_works/general_overview.html#интеграция-с-docker-registry" | true_relative_url }}) необходимо определить следующие переменные:
 * [`DOCKER_CONFIG`]({{ "documentation/internals/how_ci_cd_integration_works/general_overview.html#docker_config" | true_relative_url }});
 * [`WERF_REPO`]({{ "documentation/internals/how_ci_cd_integration_works/general_overview.html#werf_repo" | true_relative_url }}).

Создадим временную папку для конфигураций docker на базе существующей и определим Docker registry для публикации собранных образов:

```shell
TMP_DOCKER_CONFIG=$(mktemp -d)
DOCKER_CONFIG=${DOCKER_CONFIG:-$HOME/.docker}
[[ -d "$DOCKER_CONFIG" ]] && cp -a $DOCKER_CONFIG/. $TMP_DOCKER_CONFIG
export DOCKER_CONFIG=$TMP_DOCKER_CONFIG
export WERF_REPO=registry.company.com/project/werf
```

### Интеграция с CI/CD pipeline'ами

Согласно [описанным шагам]({{ "documentation/internals/how_ci_cd_integration_works/general_overview.html#интеграция-с-настройками-cicd-pipeline" | true_relative_url }}) по интеграции с CI/CD pipeline необходимо определить следующие переменные:
 * [`WERF_ADD_ANNOTATION_PROJECT_GIT`]({{ "documentation/internals/how_ci_cd_integration_works/general_overview.html#werf_add_annotation_project_git" | true_relative_url }});
 * [`WERF_ADD_ANNOTATION_CI_COMMIT`]({{ "documentation/internals/how_ci_cd_integration_works/general_overview.html#werf_add_annotation_ci_commit" | true_relative_url }}).

### Интеграция с CI/CD процессами

Согласно [описанным шагам]({{ "documentation/internals/how_ci_cd_integration_works/general_overview.html#интеграция-с-настройками-cicd" | true_relative_url }}) по интеграции с CI/CD процессами необходимо определить следующие переменные:
 * [`WERF_ENV`]({{ "documentation/internals/how_ci_cd_integration_works/general_overview.html#werf_env" | true_relative_url }}).

### Общая интеграция с CI/CD системами

Согласно [описанным шагам]({{ "documentation/internals/how_ci_cd_integration_works/general_overview.html#настройка-режима-работы-в-cicd-системе" | true_relative_url }}) по общей интеграции с CI/CD системами необходимо определить следующие переменные:
 * [`WERF_LOG_COLOR_MODE`]({{ "documentation/internals/how_ci_cd_integration_works/general_overview.html#werf_log_color_mode" | true_relative_url }});
 * [`WERF_LOG_PROJECT_DIR`]({{ "documentation/internals/how_ci_cd_integration_works/general_overview.html#werf_log_project_dir" | true_relative_url }});
 * [`WERF_ENABLE_PROCESS_EXTERMINATOR`]({{ "documentation/internals/how_ci_cd_integration_works/general_overview.html#werf_enable_process_exterminator" | true_relative_url }});
 * [`WERF_LOG_TERMINAL_WIDTH`]({{ "documentation/internals/how_ci_cd_integration_works/general_overview.html#werf_log_terminal_width" | true_relative_url }}).

## Пример скрипта настройки CI-окружения

В корневой папке проекта создадим bash-скрипт `werf-ci-env.sh` со следующим содержанием:

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

> Исправьте скрипт для работы в своей CI/CD-системе: измените присваиваемые значения переменных `WERF_*` согласно конкретному случаю. Будет полезно ознакомиться и ориентироваться на статью по [интеграции с GitLab CI]({{ "documentation/internals/how_ci_cd_integration_works/gitlab_ci_cd.html" | true_relative_url }}), чтобы понимать, какие значения для переменных стоит использовать.

Также создадим скрипт `werf-ci-env-cleanup.sh` со следующим содержимым:

```shell
rm -rf $TMP_DOCKER_CONFIG
```

Скрипт `werf-ci-env.sh` должен вызываться в начале каждого CI-задания, до вызова любой команды werf.
Скрипт `werf-ci-env-cleanup.sh` должен вызываться в конце каждого CI-задания.
