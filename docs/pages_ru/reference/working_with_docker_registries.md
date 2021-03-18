---
title: Работа с Docker registries
sidebar: documentation
permalink: reference/working_with_docker_registries.html
author: Alexey Igrychev <alexey.igrychev@flant.com>
---

Некоторые категории команд работают с Docker registry, и требуют соответствующей авторизации:
* [Во время процесса сборки]({{ site.baseurl }}/reference/build_process.html) werf может делать pull образов из Docker registry.
* [Во время процесса публикации]({{ site.baseurl }}/reference/publish_process.html) werf создает и обновляет образы в Docker registry.
* [Во время процесса очистки]({{ site.baseurl }}/reference/cleaning_process.html) werf удаляет образы из Docker registry.
* [Во время процесса деплоя]({{ site.baseurl }}/reference/deploy_process/deploy_into_kubernetes.html) werf требует доступа к _образам_ в Docker registry и _стадиям_, которые также могут находиться в Docker registry.

## Поддерживаемые имплементации

|                 	                    | Build и Publish 	        | Cleanup                         	                                    |
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
| [_Quay_](#quay)            	        |         **ок**        	|                            **ок**                            	        |

> В ближайшее время планируется добавить поддержку для Nexus Registry

Следующие имплементации полностью поддерживаются и от пользователя требуется только выполнить [авторизацию Docker](#авторизация-docker): 
* _Default_.
* _GCR_.
* _GitLab Registry_.
* _Harbor_.

_Azure CR_, _AWS ECR_, _Docker Hub_ и _GitHub Packages_ имплементации поддерживают Docker Registry API, но не полностью. Для перечисленных имплементаций необходимо использовать нативное API для удаления тегов. Поэтому при [очистке]({{ site.baseurl }}/reference/cleaning_process.html) для werf может потребоваться дополнительные пользовательские данные.

Часть имплементаций не поддерживает вложенные репозитории (_Docker Hub_, _GitHub Packages_ и _Quay_) или как в случае с _AWS ECR_ пользователь должен создать репозитории вручную, используя UI или API, перед использованием werf.

## Как хранить образы

Параметры _images repo_ и _images repo mode_ определяют где и как хранятся образы в репозитории (подробнее в разделе [именование образов]({{ site.baseurl }}/reference/publish_process.html#именование-образов)).

Параметр _images repo_ может быть как адресом **registry**, так и **repository**.  

Конечное имя Docker-образа зависит от _images repo mode_:
- `IMAGES_REPO:IMAGE_NAME-TAG` шаблон для **monorepo** мода;
- `IMAGES_REPO/IMAGE_NAME:TAG` шаблон для **multirepo** мода.

> Docker-образы не могут храниться в registry. Поэтому комбинация **registry** и **monorepo** не имеет никакого смысла. Также, по той же причине, пользователь должен быть внимательным и не использовать  **registry** с безымянным образов (`image: ~`). 
>
> При переходе на использование **registry** с **multirepo** важно не забыть переименовать безымяный образ в _werf.yaml_ и удалить связанный managed image (`werf managed-images rm '~'`)

Таким образом, остаётся три возможных комбинации использования _images repo_ и _images repo mode_.

|                	                    | registry + multirepo 	    | repository + monorepo 	    | repository + multirepo 	    |
|--------------------------------------	|:------------------------: |:----------------------------:	|:---------------------------:	|
| [_AWS ECR_](#aws-ecr)             	| **ок**                   	| **ок**                    	| **ок**                     	|
| [_Azure CR_](#azure-cr)           	| **ок**                   	| **ок**                    	| **ок**                     	|
| _Default_         	                | **ок**                   	| **ок**                    	| **ок**                     	|
| [_Docker Hub_](#docker-hub)      	    | **ок**                   	| **ок**                    	| **не поддерживается**         |
| _GCR_             	                | **ок**                   	| **ок**                    	| **ок**                     	|
| [_GitHub Packages_](#github-packages) | **ок**                   	| **ок**                    	| **не поддерживается**         |
| _GitLab Registry_ 	                | **ок**                   	| **ок**                    	| **ок**                     	|
| _Harbor_          	                | **ок**                   	| **ок**                    	| **ок**                     	|
| _JFrog Artifactory_          	        | **ок**                   	| **ок**                    	| **ок**                     	|
| [_Quay_](#quay)            	        | **ок**                   	| **ок**                    	| **не поддерживается**         |

Большинство имплементаций поддерживают вложенные репозитории и по умолчанию используют **multirepo** _images repo mode_. 
Для остальных имплементаций значение по умолчанию зависит от указанного _images repo_.   

## AWS ECR

### Как хранить образы

Хранение образов в _AWS ECR_  не отличается от остальных имплементаций, но пользователь должен самостоятельно создать репозитории перед использованием werf. 

### Как чистить stages и images

werf использует _AWS SDK_ для удаления тегов, поэтому перед использованием [команд очистки]({{ site.baseurl }}/reference/cleaning_process.html) пользователь должен:
* [Установить _AWS CLI_ и выполнить конфигурацию](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html#cli-quick-configuration) (`aws configure`) или
* Определить `AWS_ACCESS_KEY_ID` и `AWS_SECRET_ACCESS_KEY` переменные окружения.
      
## Azure CR

### Как чистить stages и images

Для того, чтобы werf начал удалять теги, пользователю необходимо выполнить следующие шаги:
* Установить [Azure CLI](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli?view=azure-cli-latest) (`az`).
* Выполнить авторизацию (`az login`).

> Для удаления тегов пользователь должен иметь одну из следующих ролей: `Owner`, `Contributor` или `AcrDelete` (подробнее [Azure CR roles and permissions](https://docs.microsoft.com/en-us/azure/container-registry/container-registry-roles)) 

## Docker Hub

### Как хранить образы

Если используется один Docker Hub аккаунт для всего проекта и есть необходимость хранения образов в отдельных репозиториях, то в качестве _images repo_ необходимо указывать аккаунт:
```shell
<ACCOUNT> # к примеру, library или index.docker.io/library
``` 

> Будьте бдительны и не используйте безымянные образы в таком случае, т.к. это может приводить к сложнодиагностируемым ошибкам. 
>
> При переходе на такой подход пользователь должен переименовать безымянный образ в _werf.yaml_ и удалить соответствующий managed image в stages storage (`werf managed-images rm '~'`), если имеется    

Если в werf конфигурации используется безымянный образ (`image: ~`) или пользователь хочет хранить все образы в одном репозитории, то необходимо использовать конкретный репозиторий в качестве параметра для _images repo_: 
```shell
<ACCOUNT>/<REPOSITORY> # к примеру, library/alpine или index.docker.io/library/alpine
```

#### Пример

Пользователь имеет следующее:
* Два образа в _werf.yaml_: **frontend**, **backend**.
* Docker Hub аккаунт: **foo**.

Есть два возможных способа хранения конечных образов:
1. Параметр _images repo_ **foo** приведёт к следующим тегам: `foo/frontend:tag` и `foo/backend:tag`. 
    ```shell
    werf build-and-publish -s=:local -i=foo --tag-custom=tag
    ```
2. Параметр _images repo_ **foo/project** приведёт к следующим тегам: `foo/project:frontend-tag` и `foo/project:backend-tag`.
    ```shell
    werf build-and-publish -s=:local -i=foo/project --tag-custom=tag
    ```

### Как чистить stages и images

Для чистки тегов в _Docker Hub_ репозитории werf использует _Docker Hub API_ и для работы требуются дополнительные параметры.

Пользователь должен определить token или username и password.

Используя следующий скрипт, пользователь может получить token самостоятельно:
```shell
HUB_USERNAME=USERNAME
HUB_PASSWORD=PASSWORD
HUB_TOKEN=$(curl -s -H "Content-Type: application/json" -X POST -d '{"username": "'${HUB_USERNAME}'", "password": "'${HUB_PASSWORD}'"}' https://hub.docker.com/v2/users/login/ | jq -r .token)
```

> В качестве токена нельзя использовать [personal access token](https://docs.docker.com/docker-hub/access-tokens/), т.к. удаление ресурсов возможно только при использовании основных учётных данных

Для того, чтобы задать параметры, следует использовать следующие опции и соответствующие переменные окружения:
* Для stages storage: `--stages-storage-repo-docker-hub-token` или `--stages-storage-repo-docker-hub-username` и `--stages-storage-repo-docker-hub-password`.
* Для images repo: `--images-repo-docker-hub-token` или `--images-repo-docker-hub-username` и `--images-repo-docker-hub-password`.
* И то и другое: `--repo-docker-hub-token` или `--repo-docker-hub-username` и `--repo-docker-hub-password`.

## GitHub Packages

### Как хранить образы

Если есть необходимость хранения образов в отдельных packages, то пользователь не должен указывать package в _images repo_: 
```shell
docker.pkg.github.com/<ACCOUNT>/<PROJECT> # к примеру, docker.pkg.github.com/flant/werf
```

> Будьте бдительны и не используйте безымянные образы в таком случае, т.к. это может приводить к сложнодиагностируемым ошибкам. 
>
> При переходе на такой подход, пользователь должен переименовать безымянный образ в _werf.yaml_ и удалить соответствующий managed image в stages storage (`werf managed-images rm '~'`), если имеется    

Если в werf конфигурации используется безымянный образ (`image: ~`) или пользователь хочет хранить все образы в одном package, то необходимо использовать конкретный package в качестве параметра для _images repo_: 
```shell
docker.pkg.github.com/<ACCOUNT>/<PROJECT>/<PACKAGE> # к примеру, docker.pkg.github.com/flant/werf/image
```

#### Пример

Пользователь имеет следующее:
* Два образа в _werf.yaml_: **frontend**, **backend**.
* GitHub репозиторий: **github.com/company/project**.

Есть два возможных способа хранения конечных образов:
1. Параметр _images repo_ **docker.pkg.github.com/company/project** приведёт к следующим тегам: `docker.pkg.github.com/company/project/frontend:tag` и `docker.pkg.github.com/company/project/backend:tag`. 
    ```shell
    werf build-and-publish -s=:local -i=docker.pkg.github.com/company/project --tag-custom=tag
    ```
2. Параметр _images repo_ **docker.pkg.github.com/company/project/app**  приведёт к следующим тегам: `docker.pkg.github.com/company/project/app:frontend-tag` и `docker.pkg.github.com/company/project/app:backend-tag`.
    ```shell
    werf build-and-publish -s=:local -i=docker.pkg.github.com/company/project/app --tag-custom=tag
    ```

### Как чистить stages и images

Для [удаления версий package приватного репозитория](https://help.github.com/en/packages/publishing-and-managing-packages/deleting-a-package) используется GraphQL. От пользователя требуется token с `read:packages`, `write:packages`, `delete:packages` и `repo` scopes.

> GitHub не поддерживает удание версий package в публичных репозиториях 

Для того, чтобы задать параметры, следует использовать следующие опции и соответствующие переменные окружения:
* Для stages storage: `--stages-storage-repo-github-token`.
* Для images repo: `--images-repo-github-token`.
* И то и другое: `--repo-github-token`.

## Quay

### Как хранить образы

Если есть необходимость хранения образов в отдельных репозиториях, то не нужно указывать репозиторий _images repo_: 
```shell
quay.io/<USER or ORGANIZATION> # к примеру, quay.io/werf
```

> Будьте бдительны и не используйте безымянные образы в таком случае, т.к. это может приводить к сложнодиагностируемым ошибкам. 
>
> При переходе на такой подход, пользователь должен переименовать безымянный образ в _werf.yaml_ и удалить соответствующий managed image в stages storage (`werf managed-images rm '~'`), если имеется    

Если в werf конфигурации используется безымянный образ (`image: ~`) или пользователь хочет хранить все образы в одном репозиторий, то необходимо использовать конкретный репозиторий в качестве параметра для _images repo_: 
```shell
quay.io/<USER or ORGANIZATION>/<REPOSITORY> # к примеру, quay.io/werf/image
```

#### Пример

Пользователь имеет следующее:
* Два образа в _werf.yaml_: **frontend**, **backend**.
* Организацию company в quay.io: **quay.io/company**.

Есть два возможных способа хранения конечных образов:
1. Параметр _images repo_ **quay.io/company** приведёт к следующим тегам: `quay.io/company/frontend:tag` и `quay.io/company/backend:tag`. 
    ```shell
    werf build-and-publish -s=:local -i=quay.io/company --tag-custom=tag
    ```
2. Параметр _images repo_ **quay.io/company/app** приведёт к следующим тегам: `quay.io/company/app:frontend-tag` и `quay.io/company/app:backend-tag`.
    ```shell
    werf build-and-publish -s=:local -i=quay.io/company/app --tag-custom=tag
    ```
   
## Авторизация Docker

Все команды, требующие авторизации в Docker registry, не выполняют ее сами, а используют подготовленную _конфигурацию Docker_.

_Конфигурация Docker_ — это папка, в которой хранятся данные авторизации используемые для доступа вразличные Docker registry и другие настройки Docker.
По умолчанию, werf использует стандартную для Docker папку конфигурации: `~/.docker`. Другую используемую папку конфигурации можно указать с помощью параметра `--docker-config`, либо с помощью переменных окружения `$DOCKER_CONFIG` или `$WERF_DOCKER_CONFIG`. Все параметры и опции в файле конфигурации стандартны для Docker, их список можно посмотреть с помощью команды `docker --config`.

Для подготовки конфигурации Docker вы можете использовать команду `docker login`, либо, если вы выполняете werf в рамках CI-системы, вызвать команду [werf ci-env]({{ site.baseurl }}/cli/toolbox/ci_env.html)  (более подробно о подключении werf к CI-системам читай в [соответствующем разделе]({{ site.baseurl }}/reference/plugging_into_cicd/overview.html)).

> Использование `docker login` при параллельном выполнении заданий в CI-системе может приводить к ошибкам выполнения заданий из-за работы с временными правами и состояния race condition (одно задание влияет на другое, переопределяя конфигурацию Docker). Поэтому, необходимо обеспечивать независимую конфигурацию Docker между заданиями, используя `docker --config` или `werf ci-env`
