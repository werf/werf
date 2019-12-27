---
title: Общие сведения
sidebar: documentation
permalink: documentation/reference/plugging_into_cicd/overview.html
author: Timofey Kirillov <timofey.kirillov@flant.com>
---

werf представляет собой новое поколение CI/CD-инструментов, объединяющих в одном инструменте, который можно легко подключить к любой существующей системе CI/CD,  сборку, деплой и очистку. Это возможно потому, что werf следует установленным концепциям таких систем.

werf интегрируется с CI/CD системами используя команду `werf ci-env`. Эта команда осуществляет сбор информации об окружении, в котором выполняется werf в рамках процесса CI/CD системы, и настраивает соответствующие параметры своей работы, устанавливая переменные окружения. Далее, такие переменные окружения будут упоминаться как *ci-env* или *ci-env переменные*. Т.е. задача команды `werf ci-env` — получить информацию от CI/CD системы и настроить переменные окружения (ci-env переменные) для дальнейшего запуска других команд werf.

![ci-env]({% asset plugging_into_cicd.svg @path %})

Далее мы рассмотрим подробнее:
 * [Что такое ci-env переменные](#что-такое-ci-env-переменные) — какую информацию werf собирает в CI/CD системе и зачем;
 * [Режимы тегирования при использовании ci-env переменных](#режимы-тегирования-при-использовании ci-env-переменных) — о различиях в доступных режимах тегирования [образов]({{ site.baseurl }}/documentation/reference/stages_and_images.html#images);
 * [Как работают ci-env переменные](#как-работают-ci-env-переменные) — как нужно использовать команду `werf ci-env` и как она влияет на другие команды werf;
 * [Полный список ci-env переменных и возможности по их настройке](#полный-список-ci-env-переменных) — рассмотрим все ci-env переменные, а также то, как их можно настраивать для вашего особого случая.

## Что такое ci-env переменные

С помощью команды *werf ci-env*, werf получает информацию от CI/CD системы и устанавливает режимы дальнейшей работы для достижения следующих задач:
 * Интеграция с Docker registry;
 * Интеграция с Git;
 * Интеграция с настройками CI/CD pipeline;
 * Интеграция с настройками CI/CD;
 * Настройка режима работы в CI/CD системе.

### Интеграция с Docker registry

Обычно при работе задания в CI/CD системе, она устанавливает для каждого задания:
 1. Адрес Docker registry.
 2. Данные авторизации для доступа к Docker registry.

Команда `werf ci-env` должна обеспечить для последующих команд авторизацию в обнаруженном Docker registry с использованием также обнаруженных данных авторизации. Читайте подробнее [об авторизации в Docker registry](#авторизация-в-Docker registry) ниже. В результате должна быть установлена переменная окружения  [`DOCKER_CONFIG=PATH_TO_TMP_CONFIG`](#docker_config).

Адрес Docker registry, передаваемый CI/CD системой, будет нужен в качестве значения параметра `--images-repo` при выполнении некоторых команд. Поэтому, в результате выполнения `werf ci-env` должна быть установлена переменная окружения [`WERF_IMAGES_REPO=DOCKER_REGISTRY_REPO`](#werf_images_repo). Это позволит запускать werf в рамках задания CI/CD системы (например, для публикации, деплоя или очистки) не передавая явно значение Docker registry через параметр `--images-repo`, а позволит брать его из переменной окружения (`$WERF_IMAGES_REPO`)

### Интеграция с Git

Обычно в CI/CD системах работащих с git, запуск каждого задания выполняется в состоянии git-репозитория переключенного на конкретный коммит (иначе — detached commit state). Необходимые значения коммита, тега или ветки передаются в CI-задание через переменные окружения.

Команда `werf ci-env` определяет из переменных окружения CI-задания данные коммита, тега или ветки, и настраивает окружение для дальнейшей работы werf таким образом, чтобы werf смог тегировать [образы]({{ site.baseurl }}/documentation/reference/stages_and_images.html#images), описанные в файле конфигурации `werf.yaml`, с учетом выбранной схемы тегирования. Читай подробнее о [схемах тегирования](#режимы-тегирования-при-использовании-ci-env-переменных) ниже.

В результате выполнения команды `werf ci-env` должна быть установлена переменная окружения режима тегирования, [`WERF_TAG_GIT_TAG=GIT_TAG`](#werf_tag_git_tag) или [`WERF_TAG_GIT_BRANCH=GIT_BRANCH`](#werf_tag_git_branch).

### Интеграция с настройками CI/CD pipeline

При деплое werf может добавлять любую информацию к ресурсам Kubernetes, используя их аннотации и метки. Обычно CI/CD системы в окружении CI-задания предоставляют такую информацию, как ссылка на страницу проекта, ссылка на само CI-задание, pipeline id и другую информацию.

Команда `werf ci-env` получает из переменных окружения CI-задания ссылку на проект в CI/CD системе и настраивает окружение таким образом, чтобы при дальнейшем деплое с помощью werf в рамках этого CI-задания, эта информация была встроена в каждый ресурс Kubernetes.

Команда `werf ci-env` получает из переменных окружения CI-задания ссылку на проект в CI/CD системе и настраивает окружение таким образом, чтобы эта информация была добавлена в аннотации каждого ресурса Kubernetes, разворачиваемого с помощью werf при деплое в рамках текущего CI-задания.

Имя аннотации зависит от используемой CI/CD системы и формируется следующим образом: `"project.werf.io/CI_CD_SYSTEM_NAME-url": URL`.

В результате выполнения команды `werf ci-env` должна быть установлена переменная окружения [`WERF_ADD_ANNOTATION_PROJECT_GIT="project.werf.io/git": URL`](#werf_add_annotation_project_git).

werf автоматически добавляет и другие аннотации, а также позволяет добавлять *пользовательские метки и аннотации*, читай подробнее об этом в соответствующем разделе по [деплою]({{ site.baseurl }}/documentation/reference/deploy_process/deploy_into_kubernetes.html#annotate-and-label-chart-resources).

### Интеграция с настройками CI/CD

В CI/CD системах используется понятие *окружения* (*environment*). Окружение может логически объединять хосты, параметры доступа, информацию о подключении к кластеру Kubernetes, параметры CI-заданий (с помощью переменных окружения например) и другую информацию. Часто используемые окружения, например, это: *development*, *staging*, *testing*, *production*, а также динамические окружения — *review environments*. werf также использует понятие *окружение* в процессе [деплоя]({{ site.baseurl }}/documentation/reference/deploy_process/deploy_into_kubernetes.html#environment).

Команда `werf ci-env` анализирует текущее окружение CI-задания и дополняет его таким образом, чтобы последующие выполняемые в рамках CI-задания команды werf автоматически получили название соответствующего окружения. Если CI/CD система предоставляет в том числе и слагифицированную версию имени окружения, то выбирается именно она.

В результате выполнения команды `werf ci-env` должна быть установлена переменная окружения [`WERF_ENV=ENV`](#werf_env).

### Настройка режима работы в CI/CD системе

Команда `werf ci-env` настраивает [политики очистки]({{ site.baseurl }}/documentation/reference/cleaning_process.html#политики-очистки) следующим образом:
 * Хранить не более чем 10 образов, собранных для git-тегов. Такое поведение определяется установкой переменной окружения [`WERF_GIT_TAG_STRATEGY_LIMIT=10`](#werf_git_tag_strategy_limit);
 * Хранить образы собранные для git-тегов не более чем 30 дней. Такое поведение определяется установкой переменной окружения [`WERF_GIT_TAG_STRATEGY_EXPIRY_DAYS=30`](#werf_git_tag_strategy_expiry_days).

Если CI/CD система поддерживает вывод текста разных цветов, то команда `werf ci-env` должна устанавливать переменную окружения [`WERF_LOG_COLOR_MODE=on`](#werf_log_color_mode).

По умолчанию, werf не выводит, например, в момент сборки папку проекта, где запущен процесс сборки. 
Команда `werv ci-env` устанавливает переменную окружения [`WERF_LOG_PROJECT_DIR=1`](#werf_log_project_dir), в результате чего, при запуске команд werf будет выводиться текущая папка проекта. 
Такое поведение удобно для отладки, т.к. позволяет концентрировать больше информации о работе CI-задания в одном месте — логе CI-задания.

Команда `werf ci-env` включает т.н. уничтожитель процессов (process exterminator), который выключен по умолчанию. Некоторые CI/CD системы прекращают работу запущенных в рамках CI-задания процессов посылая сигнал `SIGKILL`. Это происходит, например, когда пользователь нажимает кнопку отмены задания. В этом случае, процессы порожденные в рамках выполнения CI-задания будут продолжать работать до их завершения. При включенном уничтожителе процессов, werf постоянно отслеживает статус родительского процесса, и в случае его завершения, также прерывает свою работу. Такое поведение определяется установкой переменной окружения  [`WERF_ENABLE_PROCESS_EXTERMINATOR=1`](#werf_enable_process_exterminator).

Команда `werf ci-env` включает ширину логирования в 100 символов, подобранную экспериментально, как удовлетворяющую большинству случаев работы с использованием современных экранов.
Такое поведение определяется установкой переменной окружения [`WERF_LOG_TERMINAL_WIDTH=100`](#werf_log_terminal_width)

## Режимы тегирования при использовании ci-env переменных

Режим тегирования определяет то, как [образ]({{ site.baseurl }}/documentation/reference/stages_and_images.html#образы), описанный в файле конфигурации `werf.yaml` и собранный werf, будет тегироваться в процессе его [публикации]({{ site.baseurl }}/documentation/reference/publish_process.html).

Единственный применяемый в настоящий момент режим — *tag-or-branch*.

В [ближайшее время](https://github.com/flant/werf/issues/1184) будет доступна поддержка режима *--tag-content*, позволяющего формировать имя тега на основании непосредственно содержимого, безотносительно названия ветки или тега.

### tag-or-branch

В этом режиме, для тегирования [образа]({{ site.baseurl }}/documentation/reference/stages_and_images.html#images), определенного в файле конфигурации `werf.yaml`, будет использовано имя git-тега или git-ветки.

Образ, относящийся к соответствующему git-тегу или git-ветке будет храниться в Docker registry в соответствии с [политиками очистки]({{ site.baseurl }}/documentation/reference/cleaning_process.html#политики-очистки).

Этот режим автоматически определяет, какой параметр, `--tag-git-tag` или `--tag-git-branch`, нужно использовать в дальнейшем при [публикации]({{ site.baseurl }}/documentation/reference/publish_process.html#именование-образов) и [деплое]({{ site.baseurl }}/documentation/cli/main/deploy.html). В результате устанавливаются соответствующие переменные окружения, которые используются в дальнейшем в процессах публикации и деплоя.

Чтобы выбрать данный режим, необходимо указать параметр `--tagging-strategy=tag-or-branch` при вызове команды [`werf ci-env`]({{ site.baseurl }}/documentation/cli/toolbox/ci_env.html).

> Если имя испольуемого git-тега или git-ветки не удовлетворяет regex-шаблону `^[\w][\w.-]*$` либо содержит более чем 128 символов, werf применяет к имени [слагификацию]({{ site.baseurl }}/documentation/reference/toolbox/slug.html).
  <br />
  <br />
  Например:
  - имя ветки `developer-feature` допустимо, и не будет изменено;
  - имя ветки `developer/feature` не допустимо и в результате имя тжга будет — `developer-feature-6e0628fc`.

## Как работают ci-env переменные

Команда `werf ci-env` возвращает список переменных окружения, необходимый для работы дальнейших команд werf. Читай подробнее об этом [далее](#передача-cli-параметров-через-переменные-окружения).

Команда [`werf ci-env`]({{ site.baseurl }}/documentation/cli/toolbox/ci_env.html) должна вызываться в самом начале CI-задания CI/CD системы, перед любыми другими командами werf.

**ЗАМЕЧАНИЕ** Команда `werf ci-env` возвращает bash-скрипт, команды которого экпортируют необходимые [переменные окружения werf](#передача-cli-параметров-через-переменные-окружения). Поэтому, применение команды `werf ci-env` подразумевает использование ее вывода в bash-команде `source`. Например:

```shell
source <(werf ci-env gitlab --tagging-strategy tag-or-branch --verbose)
werf build-and-publish --stages-storage :local
```

Использование такой конструкции также выводит экспортируемые значения в терминал.

Пример вывода команды `werf ci-env` (без направления вывода в `source`):

```shell
### DOCKER CONFIG
echo '### DOCKER CONFIG'
export DOCKER_CONFIG="/tmp/werf-docker-config-204033515"
echo 'export DOCKER_CONFIG="/tmp/werf-docker-config-204033515"'

### IMAGES REPO
echo
echo '### IMAGES REPO'
export WERF_IMAGES_REPO="registry.domain.com/project/x"
echo 'export WERF_IMAGES_REPO="registry.domain.com/project/x"'

### TAGGING
echo
echo '### TAGGING'
export WERF_TAG_GIT_TAG="v1.0.3"
echo 'export WERF_TAG_GIT_TAG="v1.0.3"'

### DEPLOY
echo
echo '### DEPLOY'
export WERF_ENV="dev"
echo 'export WERF_ENV="dev"'
export WERF_ADD_ANNOTATION_PROJECT_GIT="project.werf.io/git=https://gitlab.domain.com/project/x"
echo 'export WERF_ADD_ANNOTATION_PROJECT_GIT="project.werf.io/git=https://gitlab.domain.com/project/x"'
export WERF_ADD_ANNOTATION_CI_COMMIT="ci.werf.io/commit=b9a1ddd366aa6a20a0fd43fb6612f349d33465ff"
echo 'export WERF_ADD_ANNOTATION_CI_COMMIT="ci.werf.io/commit=b9a1ddd366aa6a20a0fd43fb6612f349d33465ff"'
export WERF_ADD_ANNOTATION_GITLAB_CI_PIPELINE_URL="gitlab.ci.werf.io/pipeline-url=https://gitlab.domain.com/project/x/pipelines/43107"
echo 'export WERF_ADD_ANNOTATION_GITLAB_CI_PIPELINE_URL="gitlab.ci.werf.io/pipeline-url=https://gitlab.domain.com/project/x/pipelines/43107"'
export WERF_ADD_ANNOTATION_GITLAB_CI_JOB_URL="gitlab.ci.werf.io/job-url=https://gitlab.domain.com/project/x/-/jobs/110681"
echo 'export WERF_ADD_ANNOTATION_GITLAB_CI_JOB_URL="gitlab.ci.werf.io/job-url=https://gitlab.domain.com/project/x/-/jobs/110681"'

### IMAGE CLEANUP POLICIES
echo
echo '### IMAGE CLEANUP POLICIES'
export WERF_GIT_TAG_STRATEGY_LIMIT="10"
echo 'export WERF_GIT_TAG_STRATEGY_LIMIT="10"'
export WERF_GIT_TAG_STRATEGY_EXPIRY_DAYS="30"
echo 'export WERF_GIT_TAG_STRATEGY_EXPIRY_DAYS="30"'

### OTHER
echo
echo '### OTHER'
export WERF_LOG_COLOR_MODE="on"
echo 'export WERF_LOG_COLOR_MODE="on"'
export WERF_LOG_PROJECT_DIR="1"
echo 'export WERF_LOG_PROJECT_DIR="1"'
export WERF_ENABLE_PROCESS_EXTERMINATOR="1"
echo 'export WERF_ENABLE_PROCESS_EXTERMINATOR="1"'
export WERF_LOG_TERMINAL_WIDTH="95"
echo 'export WERF_LOG_TERMINAL_WIDTH="95"'
```

### Передача CLI-параметров через переменные окружения

Любой параметр любой команды werf (кроме команды `ci-env`) который может быть передан через CLI, также может быть установлен с помощью переменных окружения.

Существует общее правило соответствия параметра CLI переменной окружения: переменная окружения имеет префикс `WERF_`, состоит из заглавных букв и символа подчеркивания `_` вместо дефиса `-`.

Примеры:
 * `--tag-git-tag=mytag` аналогичен использованию `WERF_TAG_GIT_TAG=mytag`;
 * `--env=staging` аналогичен использованию `WERF_ENV=staging`;
 * `--images-repo=myregistry.myhost.com/project/x` аналогичен использованию `WERF_IMAGES_REPO=myregistry.myhost.com/project/x`;
 * ... и.т.д.

Исключением из этого правила являются параметры `--add-label` и `--add-annotation`, которые могут быть указаны несколько раз. Чтобы указать эти параметры используя переменные окружения нужно использовать следующий шаблон: `WERF_ADD_ANNOTATION_<ARBITRARY_VARIABLE_NAME_SUFFIX>="annoName1=annoValue1"`. Например:

```shell
export WERF_ADD_ANNOTATION_MYANNOTATION_1="annoName1=annoValue1"
export WERF_ADD_ANNOTATION_MYANNOTATION_2="annoName2=annoValue2"
export WERF_ADD_LABEL_MYLABEL_1="labelName1=labelValue1"
export WERF_ADD_LABEL_MYLABEL_2="labelName2=labelValue2"
```

Обратите внимание, что суффиксы `_MYANNOTATION_1`, `_MYANNOTATION_2`, `_MYLABEL_1`, `_MYLABEL_2` не влияют на имена или значения аннотаций и меток, а необходимы только для получения уникальных переменных окружения.

### Авторизация в Docker registry

Как уже упоминалось в разделе [интеграции с Docker registry](#интеграция-с-docker-registry), werf выполняет автоматическую авторизацию в обнаруженном Docker registry.

Выполнение команды `werf ci-env` всегда приводит к созданию временного [файла конфигурации Docker](https://docs.docker.com/engine/reference/commandline/cli/#configuration-files).

Временный файл конфигурации Docker необходим для возможности параллельной работы CI-заданий, чтобы они не мешали друг другу. К тому-же, использование временного файла конфигурации является более безопасным вариантом, чем регистрация в Docker registry используя общий системный файл конфигурации (`~/.docker`).

Если определена переменная окружения `DOCKER_CONFIG` либо присутствует папка `~/.docker`, werf *копирует* оттуда данные в новый временный файл конфигурации. Поэтому, уже существующие или выполненные непосредственно перед вызовом `werf ci-env` регистрации в Docker registry (*docker login*) будут активны и доступны во временном файле конфигурации Docker.

После создания нового файла конфигурации, werf выполняет дополнительную регистрацию в обнаруженном из окружения CI-задания Docker registry.

Как результат процедуры [интеграции с Docker registry](#интеграция-с-docker-registry), команда `werf ci-env` экспортирует в переменной окружения `DOCKER_CONFIG` путь к временному файлу конфигурации Docker.

Этот файл конфигурации Docker будет использован впоследствии любой выполняемой командой werf, а также может быть использован командой `docker login` для выполнения дополнительных регистраций исходя из ваших нужд.

### Полный список ci-env переменных

Выполнение команды `werf ci-env` приводит к выводу списка переменных окружения для их последующего экпорта.

Чтобы настроить эти переменные по особенному, исходя из вашей ситуации, вы можете определить необходимую переменную (с префиксом имени `WERF_`) до выполнения команды `werf ci-env`. После запуска `werf ci-env`, werf обнаружит установленную переменную окружения и не будет переопределять ее значение, а также будет использовать ее при необходимости как есть. Вы можете добиться этого как непосредственно экспортируя переменные в shell-сессии перед вызовом `werf ci-env`, так и устанавливая переменные окружения в настройках проекта в CI/CD системе.

Командой [`werf ci-env`]({{ site.baseurl }}/documentation/cli/toolbox/ci_env.html) выводятся перечисленные далее переменные окружения.

#### DOCKER_CONFIG

Путь к новому временному файлу конфигурации Docker сгенерированный командой [`werf ci-env`]({{ site.baseurl }}/documentation/cli/toolbox/ci_env.html).

#### WERF_IMAGES_REPO

Согласно процедуре [интеграции с Docker registry](#интеграция-с-docker-registry), команда [`werf ci-env`]({{ site.baseurl }}/documentation/cli/toolbox/ci_env.html) определяет Docker registry и устанавливает параметр `--images-repo`, используя переменную окружения `WERF_IMAGES_REPO`.

#### WERF_TAG_GIT_TAG

Согласно процедуре [интеграции с Git](#интеграция-с-git), команда [`werf ci-env`]({{ site.baseurl }}/documentation/cli/toolbox/ci_env.html) определяет, запущено ли задание для git-тега, и в случае положительного результата устанавливает параметр `--tag-git-tag`, используя переменную окружения `WERF_TAG_GIT_TAG`.

#### WERF_TAG_GIT_BRANCH

Согласно процедуре [интеграции с Git](#интеграция-с-git), команда [`werf ci-env`]({{ site.baseurl }}/documentation/cli/toolbox/ci_env.html) определяет, запущено ли задание для git-ветки, и в случае положительного результата устанавливает параметр `--tag-git-branch`, используя переменную окружения `WERF_TAG_GIT_BRANCH`.

#### WERF_ENV

Согласно процедуре [интеграции с настройками CI/CD](#интеграция-с-настройками-ci-cd), команда [`werf ci-env`]({{ site.baseurl }}/documentation/cli/toolbox/ci_env.html) определяет имя окружения и устанавливает параметр `--env`, используя переменную окружения `WERF_ENV`.

#### WERF_ADD_ANNOTATION_PROJECT_GIT

Согласно процедуре [интеграции с настройками CI/CD pipeline](#интеграция-с-настройками-ci-cd-pipeline), команда [`werf ci-env`]({{ site.baseurl }}/documentation/cli/toolbox/ci_env.html) определяет web-адрес проекта и устанавливает параметр `--add-annotation`, используя переменную окружения `WERF_ADD_ANNOTATION_PROJECT_GIT`.

#### WERF_ADD_ANNOTATION_CI_COMMIT

Согласно процедуре [интеграции с настройками CI/CD pipeline](#интеграция-с-настройками-ci-cd-pipeline), команда [`werf ci-env`]({{ site.baseurl }}/documentation/cli/toolbox/ci_env.html) определяет хэш git-коммита и устанавливает параметр `--add-annotation`, используя переменную окружения `WERF_ADD_ANNOTATION_CI_COMMIT`.

#### WERF_GIT_TAG_STRATEGY_LIMIT

Согласно процедуре [настройки режима работы в CI/CD системе](#настройка-режима-работы-в-ci-cd-системе), команда [`werf ci-env`]({{ site.baseurl }}/documentation/cli/toolbox/ci_env.html) устанавливает параметр `--git-tag-strategy-limit`, используя переменную окружения `WERF_GIT_TAG_STRATEGY_LIMIT`.

#### WERF_GIT_TAG_STRATEGY_EXPIRY_DAYS

Согласно процедуре [настройки режима работы в CI/CD системе](#настройка-режима-работы-в-ci-cd-системе), команда [`werf ci-env`]({{ site.baseurl }}/documentation/cli/toolbox/ci_env.html) устанавливает параметр `--git-tag-strategy-expiry-days`, используя переменную окружения `WERF_GIT_TAG_STRATEGY_EXPIRY_DAYS`.

#### WERF_LOG_COLOR_MODE

Согласно процедуре [настройки режима работы в CI/CD системе](#настройка-режима-работы-в-ci-cd-системе), команда [`werf ci-env`]({{ site.baseurl }}/documentation/cli/toolbox/ci_env.html) устанавливает параметр `--log-color-mode`, используя переменную окружения `WERF_LOG_COLOR_MODE`.

#### WERF_LOG_PROJECT_DIR

Согласно процедуре [настройки режима работы в CI/CD системе](#настройка-режима-работы-в-ci-cd-системе), команда [`werf ci-env`]({{ site.baseurl }}/documentation/cli/toolbox/ci_env.html) устанавливает параметр `--log-project-dir`, используя переменную окружения `WERF_LOG_PROJECT_DIR`.

#### WERF_ENABLE_PROCESS_EXTERMINATOR

Согласно процедуре [настройки режима работы в CI/CD системе](#настройка-режима-работы-в-ci-cd-системе), команда [`werf ci-env`]({{ site.baseurl }}/documentation/cli/toolbox/ci_env.html) устанавливает параметр `--enable-process-exterminator`, используя переменную окружения `WERF_ENABLE_PROCESS_EXTERMINATOR`.

#### WERF_LOG_TERMINAL_WIDTH

Согласно процедуре [настройки режима работы в CI/CD системе](#настройка-режима-работы-в-ci-cd-системе), команда [`werf ci-env`]({{ site.baseurl }}/documentation/cli/toolbox/ci_env.html) устанавливает параметр `--log-terminal-width`, используя переменную окружения `WERF_LOG_TERMINAL_WIDTH`.

## Что почитать дальше

 * [Как работает интеграция с GitLab CI]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/gitlab_ci.html).
 * [Как использовать werf с неподдерживаемыми системами CI/CD]({{ site.baseurl }}/documentation/guides/unsupported_ci_cd_integration.html).
