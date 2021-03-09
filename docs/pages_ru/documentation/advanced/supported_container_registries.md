---
title: Поддерживаемые Docker Registry
permalink: documentation/advanced/supported_container_registries.html
author: Alexey Igrychev <alexey.igrychev@flant.com>
---

Некоторые категории команд работают с Docker registry, и требуют соответствующей авторизации:
* [Во время процесса сборки]({{ "documentation/internals/build_process.html" | true_relative_url }}) werf может делать pull образов из Docker registry.
* [Во время процесса очистки]({{ "documentation/advanced/cleanup.html" | true_relative_url }}) werf удаляет образы из Docker registry.
* [Во время процесса деплоя]({{ "/documentation/advanced/helm/deploy_process/steps.html" | true_relative_url }}) werf требует доступа к _образам_ в Docker registry и _стадиям_, которые также могут находиться в Docker registry.

## Поддерживаемые имплементации

|                 	                    | Build          	        | Cleanup                         	                                    |
| -------------------------------------	| :-----------------------:	| :-------------------------------------------------------------------:	|
| [_AWS ECR_](#aws-ecr)             	|         **ок**        	|                    **ок (с нативным API)**                   	        |
| [_Azure CR_](#azure-cr)            	|         **ок**        	|                    **ок (с нативным API)**                            |
| _Default_         	                |         **ок**        	|                            **ок**                            	        |
| [_Docker Hub_](#docker-hub)      	    |         **ок**        	|                    **ок (с нативным API)**                   	        |
| _GCR_             	                |         **ок**        	|                            **ок**                            	        |
| [_GitHub Packages_](#github-packages) |         **ок**        	| **ок (с нативным API и только в приватных GitHub репозиториях)** 	    |
| _GitLab Registry_ 	                |         **ок**        	|                            **ок**                            	        |
| _Harbor_          	                |         **ок**        	|                            **ок**                            	        |
| _JFrog Artifactory_         	        |         **ок**        	|                            **ок**                            	        |
| _Quay_                    	        |         **ок**        	|                            **ок**                            	        |

> В ближайшее время планируется добавить поддержку для Nexus Registry

Следующие имплементации полностью поддерживаются и от пользователя требуется только выполнить [авторизацию Docker](#авторизация-docker): 
* _Default_.
* _GCR_.
* _GitLab Registry_.
* _Harbor_.

_Azure CR_, _AWS ECR_, _Docker Hub_ и _GitHub Packages_ имплементации поддерживают Docker Registry API, но не полностью. Для перечисленных имплементаций необходимо использовать нативное API для удаления тегов. Поэтому при [очистке]({{ "documentation/advanced/cleanup.html" | true_relative_url }}) для werf может потребоваться дополнительные пользовательские данные.

## AWS ECR

Хранение образов в _AWS ECR_  не отличается от остальных имплементаций, но пользователь должен самостоятельно создать репозитории перед использованием werf. 

werf использует _AWS SDK_ для удаления тегов, поэтому перед использованием [команд очистки]({{ "documentation/advanced/cleanup.html" | true_relative_url }}) пользователь должен:
* [Установить _AWS CLI_ и выполнить конфигурацию](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html#cli-quick-configuration) (`aws configure`) или
* Определить `AWS_ACCESS_KEY_ID` и `AWS_SECRET_ACCESS_KEY` переменные окружения.
      
## Azure CR

Для того, чтобы werf начал удалять теги, пользователю необходимо выполнить следующие шаги:
* Установить [Azure CLI](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli?view=azure-cli-latest) (`az`).
* Выполнить авторизацию (`az login`).

> Для удаления тегов пользователь должен иметь одну из следующих ролей: `Owner`, `Contributor` или `AcrDelete` (подробнее [Azure CR roles and permissions](https://docs.microsoft.com/en-us/azure/container-registry/container-registry-roles)) 

## Docker Hub

Для чистки тегов в _Docker Hub_ репозитории werf использует _Docker Hub API_ и для работы требуются дополнительные параметры.

Пользователь должен определить token или username и password.

Используя следующий скрипт, пользователь может получить token самостоятельно:
```shell
HUB_USERNAME=USERNAME
HUB_PASSWORD=PASSWORD
HUB_TOKEN=$(curl -s -H "Content-Type: application/json" -X POST -d '{"username": "'${HUB_USERNAME}'", "password": "'${HUB_PASSWORD}'"}' https://hub.docker.com/v2/users/login/ | jq -r .token)
```

> В качестве токена нельзя использовать [personal access token](https://docs.docker.com/docker-hub/access-tokens/), т.к. удаление ресурсов возможно только при использовании основных учётных данных

Для того, чтобы задать параметры, следует использовать следующие опции или соответствующие переменные окружения:
* `--repo-docker-hub-token` или
* `--repo-docker-hub-username` и `--repo-docker-hub-password`.

## GitHub Packages

Для [удаления версий package приватного репозитория](https://help.github.com/en/packages/publishing-and-managing-packages/deleting-a-package) используется GraphQL. От пользователя требуется token с `read:packages`, `write:packages`, `delete:packages` и `repo` scopes.

> GitHub не поддерживает удаление версий package в публичных репозиториях 

Для того, чтобы задать параметры, следует использовать опцию `--repo-github-token` или соответствующую переменную окружения.
   
## Авторизация Docker

Все команды, требующие авторизации в Docker registry, не выполняют ее сами, а используют подготовленную _конфигурацию Docker_.

_Конфигурация Docker_ — это папка, в которой хранятся данные авторизации используемые для доступа в различные Docker registry и другие настройки Docker.
По умолчанию, werf использует стандартную для Docker папку конфигурации: `~/.docker`. Другую используемую папку конфигурации можно указать с помощью параметра `--docker-config`, либо с помощью переменных окружения `$DOCKER_CONFIG` или `$WERF_DOCKER_CONFIG`. Все параметры и опции в файле конфигурации стандартны для Docker, их список можно посмотреть с помощью команды `docker --config`.

Для подготовки конфигурации Docker вы можете использовать команду `docker login`, либо, если вы выполняете werf в рамках CI-системы, вызвать команду [werf ci-env]({{ "documentation/reference/cli/werf_ci_env.html" | true_relative_url }})  (более подробно о подключении werf к CI-системам читай в [соответствующем разделе]({{ "documentation/internals/how_ci_cd_integration_works/general_overview.html" | true_relative_url }})).

> Использование `docker login` при параллельном выполнении заданий в CI-системе может приводить к ошибкам выполнения заданий из-за работы с временными правами и состояния race condition (одно задание влияет на другое, переопределяя конфигурацию Docker). Поэтому, необходимо обеспечивать независимую конфигурацию Docker между заданиями, используя `docker --config` или `werf ci-env`
