---
title: Поддерживаемые container registries
permalink: advanced/supported_container_registries.html
author: Alexey Igrychev <alexey.igrychev@flant.com>
---

По умолчанию werf использует [_Docker Registry API_](https://docs.docker.com/registry/spec/api/) и для большинства container registries от пользователя никаких дополнительных действий не требуется.

Тем не менее container registry может поддерживать _Docker Registry API_ не полностью и/или реализовывать часть функций в нативном API. В таком случае, при работе с container registry от пользователя могут потребоваться дополнительные действия и данные.

Используя заданный адрес репозитория (опция `--repo`) werf пытается автоматически определить container registry. Пользователь может явно указать container registry, используя опцию `--repo-container-registry` или переменную окружения `WERF_REPO_CONTAINER_REGISTRY`.

|                                           | Сборка | Бандлы                | Очистка                                        |
| -------------------------------------     | :----: | :-------------------: | :--------------------------------------------: |
| _AWS ECR_                                 | **ок** |         **ок**        |            [***ок**](#aws-ecr)                 |
| _Azure CR_                                | **ок** |         **ок**        |            [***ок**](#azure-cr)                |
| _Default_                                 | **ок** |         **ок**        |                   **ок**                       |
| _Docker Hub_                              | **ок** | **не поддерживается** |           [***ок**](#docker-hub)               |
| _GCR_                                     | **ок** |         **ок**        |                   **ок**                       |
| _GitHub Packages_ (docker.pkg.github.com) | **ок** | **не поддерживается** | [***ок**](#github-packages-dockerpkggithubcom) |
| _GitHub Packages_ (ghcr.io)               | **ок** |         **ок**        |            **не поддерживается**               |
| _GitLab Registry_                         | **ок** |         **ок**        |         [***ок**](#gitlab-registry)            |
| _Harbor_                                  | **ок** |         **ок**        |                   **ок**                       |
| _JFrog Artifactory_                       | **ок** |         **ок**        |                   **ок**                       |
| _Nexus_                                   | **ок** |   **не проверялся**   |                   **ок**                       |
| _Quay_                                    | **ок** | **не поддерживается** |                   **ок**                       |
| _Yandex Container Registry_               | **ок** |   **не проверялся**   |                   **ок**                       |

## Авторизация

При работе с container registry команды werf используют данные хранящиеся в _конфигурации Docker_. По умолчанию используется пользовательская директория `~/.docker` или альтернативный путь, заданный переменной окружения `WERF_DOCKER_CONFIG` или `DOCKER_CONFIG`.

> _Конфигурация Docker_ — это директория, в которой хранятся данные авторизации, используемые для доступа в различные container registries, а также настройки Docker

Для авторизации можно использовать команду `docker login`, `oras login` или нативные решения container registry.

В рамках CI-заданий рекомендуется использовать команду [werf ci-env]({{ "reference/cli/werf_ci_env.html" | true_relative_url }}), которая создаёт временную _конфигурацию Docker_ на базе пользовательской, а также выполняет авторизацию (при использовании container registry CI), используя переменные окружения CI. Подробнее о подключении werf к CI-системам в [соответствующем разделе]({{ "internals/how_ci_cd_integration_works/general_overview.html" | true_relative_url }}).

> Использование общей _конфигурации Docker_ при параллельном выполнении заданий в CI-системе может приводить к падениям из-за временных доступов и состояния race condition (одно задание влияет на другое, переопределяя доступы в общей _конфигурации Docker_), поэтому для каждого CI-задания мы рекомендуем создавать собственную _конфигурацию Docker_ (команда `werf ci-env` делает это по умолчанию)

## Бандлы

Для работы с [бандлами]({{ "advanced/bundles.html" | true_relative_url }}) достаточно поддержки [спецификации формата образов Open Container Initiative (OCI)](https://github.com/opencontainers/image-spec) в container registry.

## Очистка

По умолчанию при удалении тегов werf использует _Docker Registry API_ и от пользователя требуется только авторизация с использованием доступов с достаточным набором прав. 

Если удаление посредством _Docker Registry API_ не поддерживается и оно реализуется в нативном API, то от пользователя потребуется выполнить специфичные для container registry действия.

### AWS ECR

При удалении тегов werf использует _AWS SDK_, поэтому перед очисткой container registry необходимо выполнить **одно из** следующих действий:

- [Установить _AWS CLI_ и выполнить конфигурацию](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html#cli-quick-configuration) (`aws configure`) или
- Определить `AWS_ACCESS_KEY_ID` и `AWS_SECRET_ACCESS_KEY` переменные окружения.
      
### Azure CR

При удалении тегов werf использует _Azure CLI_, поэтому перед очисткой container registry необходимо выполнить следующие действия: 

- Установить [Azure CLI](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli?view=azure-cli-latest) (`az`).
- Выполнить авторизацию (`az login`).

> Пользователю необходимо иметь одну из следующих ролей: `Owner`, `Contributor` или `AcrDelete` (подробнее [Azure CR roles and permissions](https://docs.microsoft.com/en-us/azure/container-registry/container-registry-roles)) 

### Docker Hub

При удалении тегов werf использует _Docker Hub API_, поэтому при очистке container registry необходимо определить _token_ или _username_ и _password_.

Для получения _token_ можно использовать следующий скрипт:

```shell
HUB_USERNAME=username
HUB_PASSWORD=password
HUB_TOKEN=$(curl -s -H "Content-Type: application/json" -X POST -d '{"username": "'${HUB_USERNAME}'", "password": "'${HUB_PASSWORD}'"}' https://hub.docker.com/v2/users/login/ | jq -r .token)
```

> В качестве _token_ нельзя использовать [personal access token](https://docs.docker.com/docker-hub/access-tokens/), т.к. удаление ресурсов возможно только при использовании основных учётных данных пользователя

Для того, чтобы задать параметры, следует использовать следующие опции или соответствующие им переменные окружения:
- `--repo-docker-hub-token` или
- `--repo-docker-hub-username` и `--repo-docker-hub-password`.

### GitHub Packages (docker.pkg.github.com)

При удалении тегов werf использует _GitHub GraphQL API_, поэтому при очистке container registry необходимо определить _token_ с `read:packages`, `write:packages`, `delete:packages` и `repo` scopes.

> GitHub [не поддерживает](https://help.github.com/en/packages/publishing-and-managing-packages/deleting-a-package) удаление версий package в публичных репозиториях

Для того, чтобы задать параметры, следует использовать опцию `--repo-github-token` или соответствующую переменную окружения.

### GitLab Registry

При удалении тегов werf использует _GitLab Container Registry API_ или _Docker Registry API_ в зависимости от версии GitLab.

> Для удаления тега прав временного токена CI-задания (`$CI_JOB_TOKEN`) недостаточно, поэтому пользователю необходимо создать специальный токен в разделе Access Token (в секции Scope необходимо выбрать `api`) и выполнить авторизацию с ним 
