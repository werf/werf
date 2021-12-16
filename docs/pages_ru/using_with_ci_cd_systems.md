---
title: Использование в CI/CD системах
permalink: using_with_ci_cd_systems.html
description: Основы использования werf в рамках различных систем CI/CD
---

## Вступление

В этой статье мы рассмотрим основы использования werf в рамках различных систем CI/CD.

Также доступна статья, в которой обсуждаются более продвинутые вопросы [общей интеграции в процесс CI/CD]({{ "advanced/ci_cd/generic_ci_cd_integration.html" | true_relative_url }}).

werf изначально поддерживает GitLab CI/CD и GitHub Actions, а также предлагает специальную команду `werf ci-env` (она необязательна и позволяет настраивать параметры werf, описываемые в этой статье, автоматически и единообразно). Дополнительные сведения можно найти в следующих документах:

 - [GitLab CI/CD]({{ "advanced/ci_cd/gitlab_ci_cd.html" | true_relative_url }});
 - [GitHub Actions]({{ "advanced/ci_cd/github_actions.html" | true_relative_url }}).

## Команды werf, которые понадобятся

Следующие базовые команды werf выступают строительными блоками для построения пайплайнов:

 - `werf converge`;
 - `werf dismiss`;
 - `werf cleanup`.

Также имеются некоторые специальные команды:

  - `werf build` позволяет собирать необходимые образы;
  - `werf run` позволяет проводить модульное тестирование собранных образов.

## Встраивание werf в систему CI/CD

Обычно процесс встраивания werf в задания CI/CD-систем прост и необременителен. Для этого достаточно выполнить следующие действия:

 1. Обеспечьте заданию CI/CD доступ к кластеру Kubernetes и container registry.
 2. Проведите checkout на нужный коммит git в репозитории проекта.
 3. Запустите команду `werf converge`, чтобы развернуть приложение; запустите `werf dismiss`, чтобы удалить его из Kubernetes; выполните `werf cleanup`, чтобы удалить неиспользуемые образы из container registry.

### Подключение к кластеру Kubernetes

werf подключается и работает с кластером Kubernetes через файлы [kubeconfig](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/).

Если `kubectl` уже установлен и сконфигурирован для работы с кластером Kubernetes, то `werf` также будет работать "из коробки", поскольку он использует те же самые kubeconfig-настройки (дефолтный файл настроек `~/.kube/config` и переменную среды `KUBECONFIG`).

В ином случае необходимо создать файл kubeconfig и передать его в werf с помощью флага `--kube-config` (или переменной среды `WERF_KUBE_CONFIG`), флага `--kube-context` (или переменной `WERF_KUBE_CONTEXT`) или флага `--kube-config-base64` (переменной `WERF_KUBE_CONFIG_BASE64`), чтобы передать закодированные base64 kubeconfig-настройки непосредственно в werf.

<div class="details">
<a href="javascript:void(0)" class="details__summary">Пример</a>
<div class="details__content" markdown="1">
Используем кастомный файл kubeconfig `./my-kube-config`:

```shell
werf converge --kube-config ./my-kube-config
```

Передаем закодированный в base64 конфиг с помощью переменной окружения:

```shell
export WERF_KUBE_CONFIG_BASE64="$(cat ./my-kube-config | base64 -w0)"
werf converge
```
</div>
</div>

### Настройка доступа задания CI/CD к container registry

Если используются непубличные образы, собранные специально для конкретного проекта, приходится использовать приватный container registry. В этом случае хост, на котором работает раннер CI/CD, должен иметь доступ к container registry, чтобы публиковать собранные образы. Кроме того, кластеру Kubernetes необходим доступ для скачивания этих образов.

Если клиент `docker` уже настроен на доступ к приватному container registry с вашего хоста, дополнительное конфигурирование werf не требуется, поскольку он использует те же настройки, что и docker (каталог с настройками по умолчанию `~/.docker/` или переменную среды `DOCKER_CONFIG`).

<div class="details">
<a href="javascript:void(0)" class="details__summary">В противном случае выполните стандартную процедуру входа в container registry.</a>
<div class="details__content" markdown="1">
```shell
docker login registry.mydomain.org/application -uUSER -pPASSWORD
```
  
Если постоянный логин раннер-хоста в container registry не требуется, выполните процедуру входа в каждом из заданий Ci/CD. Мы рекомендуем создавать временную конфигурацию docker для каждого CI/CD-задания (вместо использования директории по умолчанию `~/.docker/`), чтобы предотвратить конфликт разных заданий, выполняющихся на одном и том же раннере в одно и то же время.

```shell
# Job 1
export DOCKER_CONFIG=$(mktemp -d)
docker login registry.mydomain.org/application -uUSER -pPASSWORD
# Job 2
export DOCKER_CONFIG=$(mktemp -d)
docker login registry.mydomain.org/application -uUSER -pPASSWORD
```
</div>
</div>

### Настройка целевого окружения для werf

Обычно приложение развертывается в различные [окружения]({{ "advanced/ci_cd/ci_cd_workflow_basics.html#окружение" | true_relative_url }}) (`production`, `staging`, `testing`, и т.д.).

В werf имеется опциональный параметр `--env` (или переменная среды `WERF_ENV`), с помощью которого можно задать имя используемого окружения. Оно влияет на название соответствующего [пространства имен Kubernetes]() и [название Helm-релиза](). Мы рекомендуем проверять имя окружения в процессе выполнения CI/CD-задания (например, с помощью встроенных переменных окружения вашей CI/CD-системы) и соответствующим образом устанавливать параметр `--env`.

<div class="details">
<a href="javascript:void(0)" class="details__summary">Пример</a>
<div class="details__content" markdown="1">

Задаем переменную среду `WERF_ENV`:

```shell
export WERF_ENV=production
werf converge
```

Используем флаг cli:

```shell
werf converge --env staging
```

Используем встроенные переменные среды CI/CD-системы:

```shell
export WERF_ENV=$CI_ENVIRONMENT_NAME
werf converge
```
</div>
</div>

### Checkout на коммит git и запуск werf

Обычно CI/CD-система производит checkout на текущий git-коммит полностью автоматически, и пользователю не приходится этим заниматься самостоятельно. Однако некоторые системы могут быть лишены подобной функциональности. В этом случае вы должны самостоятельно произвести checkout на целевой git-коммит в репозитории проекта перед запуском основных команд werf (`werf converge`, `werf dismiss`, `werf cleanup`).

Примечание: поведение команды werf converge полностью детерминировано и прозрачно с точки зрения репозитория git. После завершения ее работы приложение будет запущено. Оно будет соответствовать состоянию, определенному в целевом git-коммите. -

### Другие настройки werf, связанные с CI/CD

Существуют и другие настройки (как правило, они настраиваются для werf в CI/CD):

 1. Кастомные аннотации и лейблы для всех развернутых ресурсов Kubernetes
   - имя приложения;
   - URL-адрес приложения в Git;
   - версия приложения;
   - URL converge-задания в CI/CD;
   - git-коммит;
   - git-тег;
   - git-ветвь;
   - и т.д.

Кастомные аннотации и лейблы можно задавать с помощью опций -`-add-annotation`, `--add-label` (или переменных среды `WERF_ADD_ANNOTATION_<ARBITRARY_STRING>`, `WERF_ADD_LABEL_<ARBITRARY_STRING>`).

 2. Подсветка логов, ширина вывода и пользовательские параметры задаются с помощью флагов `--log-color-mode=on|off`, `--log-terminal-width=N`, `--log-project-dir` (а также с помощью соответствующих переменных среды `WERF_COLOR_MODE=on|off`, `WERF_LOG_TERMINAL_WIDTH=N`, `WERF_LOG_PROJECT_DIR=1`).

 3. Особый режим работы werf под названием "уничтожитель процессов". Он обнаруживает отмененное задание CI/CD и автоматически завершает процесс werf. Эта функциональность востребована не во всех системах, а только в тех, которые не умеют посылать правильный сигнал о завершении порожденным процессам (например, в GitLab CI/CD). В этом случае лучше включить опцию werf `--enable-process-exterminator` (или установить переменную окружения `WERF_ENABLE_PROCESS_EXTERMINATOR=1`).

<br>

<div class="details">
<a href="javascript:void(0)" class="details__summary">Пример</a>
<div class="details__content" markdown="1">
Давайте зададим кастомные аннотации, настроим логирование и включим режим "уничтожителя процессов" с помощью переменных среды:

```shell
export WERF_ADD_ANNOTATION_APPLICATION_NAME="project.werf.io/name=myapp"
export WERF_ADD_ANNOTATION_APPLICATION_GIT_URL="project.werf.io/git-url=https://mygit.org/myapp"
export WERF_ADD_ANNOTATION_APPLICATION_VERSION="project.werf.io/version=v1.4.14"
export WERF_ADD_ANNOTATION_CI_CD_JOB_URL="ci-cd.werf.io/job-url=https://mygit.org/myapp/jobs/256385"
export WERF_ADD_ANNOTATION_CI_CD_GIT_COMMIT="ci-cd.werf.io/git-commit=a16ce6e8b680f4dcd3023c6675efe5df3f40bfa5"
export WERF_ADD_ANNOTATION_CI_CD_GIT_TAG="ci-cd.werf.io/git-tag=v1.4.14"
export WERF_ADD_ANNOTATION_CI_CD_GIT_BRANCH="ci-cd.werf.io/git-branch=master"

export WERF_LOG_COLOR_MODE=on
export WERF_LOG_TERMINAL_WIDTH=95
export WERF_LOG_PROJECT_DIR=1
export WERF_ENABLE_PROCESS_EXTERMINATOR=1
```
</div>
</div>

## Что дальше?

В [этом разделе]({{ "reference/deploy_annotations.html" | true_relative_url }}) рассказывается, как управлять отслеживанием ресурсов в процессе развертывания.

Также рекомендуем ознакомиться со статьей ["Основы рабочего процесса CI/CD"]({{ "advanced/ci_cd/ci_cd_workflow_basics.html" | true_relative_url }}). В ней описываются различные способы настройки рабочих процессов CI/CD.

В разделе ["Руководства"](/guides.html) можно найти инструкцию, подходящую для вашего проекта. Эти руководства также содержат подробную информацию о настройке конкретных систем CI/CD.
