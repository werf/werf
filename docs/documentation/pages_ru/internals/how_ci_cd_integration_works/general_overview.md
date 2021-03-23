---
title: Общие сведения
permalink: internals/how_ci_cd_integration_works/general_overview.html
---

werf представляет собой новое поколение CI/CD-инструментов, объединяющих в одном инструменте, который можно легко подключить к любой существующей системе CI/CD, сборку, деплой и очистку. Это возможно потому, что werf следует установленным концепциям таких систем.

werf интегрируется с CI/CD системами используя команду `werf ci-env`. Эта команда осуществляет сбор информации об окружении, в котором выполняется werf в рамках процесса CI/CD системы, и настраивает соответствующие параметры своей работы, устанавливая переменные окружения. Далее, такие переменные окружения будут упоминаться как *ci-env* или *ci-env переменные*. Т.е. задача команды `werf ci-env` — получить информацию от CI/CD системы и настроить переменные окружения (ci-env переменные) для дальнейшего запуска других команд werf.

![ci-env]({{ assets["plugging_into_cicd.svg"].digest_path | true_relative_url }})

Далее мы рассмотрим подробнее:
 * [Что такое ci-env переменные](#что-такое-ci-env-переменные) — какую информацию werf собирает в CI/CD системе и зачем;
 * [Как работают ci-env переменные](#как-работают-ci-env-переменные) — как нужно использовать команду `werf ci-env` и как она влияет на другие команды werf;
 * [Полный список ci-env переменных и возможности по их настройке](#полный-список-ci-env-переменных) — рассмотрим все ci-env переменные, а также то, как их можно настраивать для вашего особого случая.

## Что такое ci-env переменные

С помощью команды *werf ci-env*, werf получает информацию от CI/CD системы и устанавливает режимы дальнейшей работы для достижения следующих задач:
 * Интеграция с container registry;
 * Интеграция с настройками CI/CD pipeline;
 * Интеграция с настройками CI/CD;
 * Настройка режима работы в CI/CD системе.

### Интеграция с container registry

Обычно при работе задания в CI/CD системе, она устанавливает для каждого задания:
 1. Адрес container registry.
 2. Данные авторизации для доступа к container registry.

Команда `werf ci-env` должна обеспечить для последующих команд авторизацию в обнаруженном container registry с использованием также обнаруженных данных авторизации. Читайте подробнее [об авторизации в container registry](#авторизация-в-container-registry) ниже. В результате должна быть установлена переменная окружения  [`DOCKER_CONFIG=PATH_TO_TMP_CONFIG`](#docker_config).

Адрес container registry, передаваемый CI/CD системой, будет нужен в качестве основы для параметра `--repo` при выполнении некоторых команд. Поэтому, в результате выполнения `werf ci-env` должна быть установлена переменная окружения [`WERF_REPO=CONTAINER_REGISTRY_REPO`](#werf_repo). Это позволит запускать werf в рамках задания CI/CD системы (например, для публикации, деплоя или очистки) не передавая явно значение container registry через параметр `--repo`, а позволит брать его из переменной окружения (`$WERF_REPO`).

Практически все встроенные container registry в CI/CD системах имеют свои особенности и поэтому werf должен знать с каким registry он работает (подробнее об особенностях поддерживаемых container registry [здесь]({{ "advanced/supported_container_registries.html" | true_relative_url }})). Если адрес репозитория однозначно определяет container registry, то дополнительно от пользователя ничего не требуется. В противном случае container registry задаётся опцией `--repo-container-registry` (`$WERF_REPO_CONTAINER_REGISTRY`). Команда `werf ci-env` проставляет значение для встроенных container registry при необходимости.

### Интеграция с настройками CI/CD pipeline

При деплое werf может добавлять любую информацию к ресурсам Kubernetes, используя их аннотации и лейблы. Обычно CI/CD системы в окружении CI-задания предоставляют такую информацию, как ссылка на страницу проекта, ссылка на само CI-задание, pipeline id и другую информацию.

Команда `werf ci-env` получает из переменных окружения CI-задания ссылку на проект в CI/CD системе и настраивает окружение таким образом, чтобы при дальнейшем деплое с помощью werf в рамках этого CI-задания, эта информация была встроена в каждый ресурс Kubernetes.

Команда `werf ci-env` получает из переменных окружения CI-задания ссылку на проект в CI/CD системе и настраивает окружение таким образом, чтобы эта информация была добавлена в аннотации каждого ресурса Kubernetes, разворачиваемого с помощью werf при деплое в рамках текущего CI-задания.

Имя аннотации зависит от используемой CI/CD системы и формируется следующим образом: `"project.werf.io/CI_CD_SYSTEM_NAME-url": URL`.

В результате выполнения команды `werf ci-env` должна быть установлена переменная окружения [`WERF_ADD_ANNOTATION_PROJECT_GIT="project.werf.io/git": URL`](#werf_add_annotation_project_git).

werf автоматически добавляет и другие аннотации, а также позволяет добавлять *пользовательские лейблы и аннотации*, читай подробнее об этом в соответствующем разделе по [деплою]({{ "/advanced/helm/deploy_process/annotating_and_labeling.html" | true_relative_url }}).

### Интеграция с настройками CI/CD

В CI/CD системах используется понятие *окружения* (*environment*). Окружение может логически объединять хосты, параметры доступа, информацию о подключении к кластеру Kubernetes, параметры CI-заданий (с помощью переменных окружения например) и другую информацию. Часто используемые окружения, например, это: *development*, *staging*, *testing*, *production*, а также динамические окружения — *review environments*. werf также использует понятие *окружение* в процессе [деплоя]({{ "/advanced/helm/configuration/templates.html#окружение" | true_relative_url }}).

Команда `werf ci-env` анализирует текущее окружение CI-задания и дополняет его таким образом, чтобы последующие выполняемые в рамках CI-задания команды werf автоматически получили название соответствующего окружения. Если CI/CD система предоставляет, в том числе и слагифицированную версию имени окружения, то выбирается именно она.

В результате выполнения команды `werf ci-env` должна быть установлена переменная окружения [`WERF_ENV=ENV`](#werf_env).

### Настройка режима работы в CI/CD системе

Если CI/CD система поддерживает вывод текста разных цветов, то команда `werf ci-env` должна устанавливать переменную окружения [`WERF_LOG_COLOR_MODE=on`](#werf_log_color_mode).

По умолчанию, werf не выводит, например, в момент сборки папку проекта, где запущен процесс сборки.
Команда `werv ci-env` устанавливает переменную окружения [`WERF_LOG_PROJECT_DIR=1`](#werf_log_project_dir), в результате чего, при запуске команд werf будет выводиться текущая папка проекта.
Такое поведение удобно для отладки, т.к. позволяет концентрировать больше информации о работе CI-задания в одном месте — логе CI-задания.

Команда `werf ci-env` включает т.н. уничтожитель процессов (process exterminator), который выключен по умолчанию. Некоторые CI/CD системы прекращают работу запущенных в рамках CI-задания процессов посылая сигнал `SIGKILL`. Это происходит, например, когда пользователь нажимает кнопку отмены задания. В этом случае процессы порожденные в рамках выполнения CI-задания будут продолжать работать до их завершения. При включенном уничтожителе процессов, werf постоянно отслеживает статус родительского процесса, и в случае его завершения, также прерывает свою работу. Такое поведение определяется установкой переменной окружения  [`WERF_ENABLE_PROCESS_EXTERMINATOR=1`](#werf_enable_process_exterminator).

Команда `werf ci-env` включает ширину логирования в 100 символов, подобранную экспериментально, как удовлетворяющую большинству случаев работы с использованием современных экранов.
Такое поведение определяется установкой переменной окружения [`WERF_LOG_TERMINAL_WIDTH=100`](#werf_log_terminal_width)

## Как работают ci-env переменные

Команда `werf ci-env` возвращает список переменных окружения, необходимый для работы дальнейших команд werf. Читай подробнее об этом [далее](#передача-cli-параметров-через-переменные-окружения).

Команда [`werf ci-env`]({{ "reference/cli/werf_ci_env.html" | true_relative_url }}) должна вызываться в самом начале CI-задания CI/CD системы, перед любыми другими командами werf.

**ЗАМЕЧАНИЕ** Команда `werf ci-env` возвращает bash-скрипт, команды которого экспортируют необходимые [переменные окружения werf](#передача-cli-параметров-через-переменные-окружения). Поэтому, применение команды `werf ci-env` подразумевает использование ее вывода в bash-команде `source`. Например:

```shell
source $(werf ci-env gitlab --verbose --as-file)
werf converge
```

Использование такой конструкции также выводит экспортируемые значения в терминал.

Пример вывода команды `werf ci-env` (без направления вывода в `source`):

```shell
### DOCKER CONFIG
echo '### DOCKER CONFIG'
export DOCKER_CONFIG="/tmp/werf-docker-config-204033515"
echo 'export DOCKER_CONFIG="/tmp/werf-docker-config-204033515"'

### REPO
echo
echo '### REPO'
export WERF_REPO="registry.domain.com/project/werf"
echo 'export WERF_REPO="registry.domain.com/project/werf"'

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
 * `--env=staging` аналогичен использованию `WERF_ENV=staging`;
 * `--repo=myregistry.myhost.com/project/werf` аналогичен использованию `WERF_REPO=myregistry.myhost.com/project/werf`;
 * ... и.т.д.

Исключением из этого правила являются параметры `--add-label` и `--add-annotation`, которые могут быть указаны несколько раз. Чтобы указать эти параметры используя переменные окружения, нужно использовать следующий шаблон: `WERF_ADD_ANNOTATION_<ARBITRARY_VARIABLE_NAME_SUFFIX>="annoName1=annoValue1"`. Например:

```shell
export WERF_ADD_ANNOTATION_MYANNOTATION_1="annoName1=annoValue1"
export WERF_ADD_ANNOTATION_MYANNOTATION_2="annoName2=annoValue2"
export WERF_ADD_LABEL_MYLABEL_1="labelName1=labelValue1"
export WERF_ADD_LABEL_MYLABEL_2="labelName2=labelValue2"
```

Обратите внимание, что суффиксы `_MYANNOTATION_1`, `_MYANNOTATION_2`, `_MYLABEL_1`, `_MYLABEL_2` не влияют на имена или значения аннотаций и меток, а необходимы только для получения уникальных переменных окружения.

### Авторизация в container registry

Как уже упоминалось в разделе [интеграции с container registry](#интеграция-с-container-registry), werf выполняет автоматическую авторизацию в обнаруженном container registry.

Выполнение команды `werf ci-env` всегда приводит к созданию временного [файла конфигурации Docker](https://docs.docker.com/engine/reference/commandline/cli/#configuration-files).

Временный файл конфигурации Docker необходим для возможности параллельной работы CI-заданий, чтобы они не мешали друг другу. К тому-же, использование временного файла конфигурации является более безопасным вариантом, чем авторизация в container registry используя общий системный файл конфигурации (`~/.docker`).

Если определена переменная окружения `DOCKER_CONFIG` либо присутствует папка `~/.docker`, werf *копирует* оттуда данные в новый временный файл конфигурации. Поэтому, уже существующие или выполненные непосредственно перед вызовом `werf ci-env` авторизации в container registry (*docker login*) будут активны и доступны во временном файле конфигурации Docker.

После создания нового файла конфигурации, werf выполняет дополнительную авторизацию в обнаруженном из окружения CI-задания container registry.

Как результат процедуры [интеграции с container registry](#интеграция-с-container-registry), команда `werf ci-env` экспортирует в переменной окружения `DOCKER_CONFIG` путь к временному файлу конфигурации Docker.

Этот файл конфигурации Docker будет использован впоследствии любой выполняемой командой werf, а также может быть использован командой `docker login` для выполнения дополнительных авторизаций, исходя из ваших нужд.

### Полный список ci-env переменных

Выполнение команды `werf ci-env` приводит к выводу списка переменных окружения для их последующего экспорта.

Чтобы настроить эти переменные по особенному, исходя из вашей ситуации, вы можете определить необходимую переменную (с префиксом имени `WERF_`) до выполнения команды `werf ci-env`. После запуска `werf ci-env`, werf обнаружит установленную переменную окружения и не будет переопределять ее значение, а также будет использовать ее при необходимости как есть. Вы можете добиться этого как непосредственно экспортируя переменные в shell-сессии перед вызовом `werf ci-env`, так и устанавливая переменные окружения в настройках проекта в CI/CD системе.

Командой [`werf ci-env`]({{ "reference/cli/werf_ci_env.html" | true_relative_url }}) выводятся перечисленные далее переменные окружения.

#### DOCKER_CONFIG

Путь к новому временному файлу конфигурации Docker сгенерированный командой [`werf ci-env`]({{ "reference/cli/werf_ci_env.html" | true_relative_url }}).

#### WERF_REPO

Согласно процедуре [интеграции с container registry](#интеграция-с-container-registry), команда [`werf ci-env`]({{ "reference/cli/werf_ci_env.html" | true_relative_url }}) определяет container registry и устанавливает параметр `--repo`, используя переменную окружения `WERF_REPO`.

#### WERF_REPO_CONTAINER_REGISTRY

Согласно процедуре [интеграции с container registry](#интеграция-с-container-registry), команда [`werf ci-env`]({{ "reference/cli/werf_ci_env.html" | true_relative_url }}) устанавливает параметр `--repo-container-registry`, используя переменную окружения `WERF_REPO_CONTAINER_REGISTRY`, вдобавок к [`WERF_REPO`](#werf_repo).

#### WERF_ENV

Согласно процедуре [интеграции с настройками CI/CD](#интеграция-с-настройками-cicd), команда [`werf ci-env`]({{ "reference/cli/werf_ci_env.html" | true_relative_url }}) определяет имя окружения и устанавливает параметр `--env`, используя переменную окружения `WERF_ENV`.

#### WERF_ADD_ANNOTATION_PROJECT_GIT

Согласно процедуре [интеграции с настройками CI/CD pipeline](#интеграция-с-настройками-cicd-pipeline), команда [`werf ci-env`]({{ "reference/cli/werf_ci_env.html" | true_relative_url }}) определяет web-адрес проекта и устанавливает параметр `--add-annotation`, используя переменную окружения `WERF_ADD_ANNOTATION_PROJECT_GIT`.

#### WERF_ADD_ANNOTATION_CI_COMMIT

Согласно процедуре [интеграции с настройками CI/CD pipeline](#интеграция-с-настройками-cicd-pipeline), команда [`werf ci-env`]({{ "reference/cli/werf_ci_env.html" | true_relative_url }}) определяет хэш git-коммита и устанавливает параметр `--add-annotation`, используя переменную окружения `WERF_ADD_ANNOTATION_CI_COMMIT`.

#### WERF_LOG_COLOR_MODE

Согласно процедуре [настройки режима работы в CI/CD системе](#настройка-режима-работы-в-cicd-системе), команда [`werf ci-env`]({{ "reference/cli/werf_ci_env.html" | true_relative_url }}) устанавливает параметр `--log-color-mode`, используя переменную окружения `WERF_LOG_COLOR_MODE`.

#### WERF_LOG_PROJECT_DIR

Согласно процедуре [настройки режима работы в CI/CD системе](#настройка-режима-работы-в-cicd-системе), команда [`werf ci-env`]({{ "reference/cli/werf_ci_env.html" | true_relative_url }}) устанавливает параметр `--log-project-dir`, используя переменную окружения `WERF_LOG_PROJECT_DIR`.

#### WERF_ENABLE_PROCESS_EXTERMINATOR

Согласно процедуре [настройки режима работы в CI/CD системе](#настройка-режима-работы-в-cicd-системе), команда [`werf ci-env`]({{ "reference/cli/werf_ci_env.html" | true_relative_url }}) устанавливает параметр `--enable-process-exterminator`, используя переменную окружения `WERF_ENABLE_PROCESS_EXTERMINATOR`.

#### WERF_LOG_TERMINAL_WIDTH

Согласно процедуре [настройки режима работы в CI/CD системе](#настройка-режима-работы-в-cicd-системе), команда [`werf ci-env`]({{ "reference/cli/werf_ci_env.html" | true_relative_url }}) устанавливает параметр `--log-terminal-width`, используя переменную окружения `WERF_LOG_TERMINAL_WIDTH`.

## Что почитать дальше

 * [Как работает интеграция с GitLab CI]({{ "internals/how_ci_cd_integration_works/gitlab_ci_cd.html" | true_relative_url }}).
 * [Как работает интеграция с GitHub Actions]({{ "internals/how_ci_cd_integration_works/github_actions.html" | true_relative_url }}).
 * [Общая инструкция, как использовать werf с CI/CD-системами]({{ "advanced/ci_cd/generic_ci_cd_integration.html" | true_relative_url }}).
