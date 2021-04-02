---
title: Как мигрировать с v1.1 на v1.2
permalink: whats_new_in_v1_2/how_to_migrate_from_v1_1_to_v1_2.html
description: How to migrate your application from v1.1 to v1.2
sidebar: documentation
---

Просто следуйте приведённым шагам, чтобы перевести ваш проект с v1.1 на v1.2.

## 0. Ограничение использования mounts

**ВАЖНО.** Существуют такие конфигурации v11, которые используют директивы **`mount from build_dir`** и **`mount fromPath`** для кеширования внешних зависимостей приложения (зависимости описываемые в таких файлах как Gemfile, package.json, go.mod и т.п.). Такие директивы `mount` всё ещё поддерживаются в werf v1.2, однако не рекомендуется к использованию и по умолчанию выключены [режимом гитерминизма]({{ "/whats_new_in_v1_2/changelog.html#гитерминизм" | true_relative_url }}) (за исключением `mount from tmp_dir` — данная директива не ограничена к использованию), потому что использование этих директив может привести к ненадёжным, недетерминированным сборкам, зависимым от сборочных хостов. Для таких конфигураций планируется добавление новой возможности хранения такого кеша в docker образах. Данная фича планируется к добавлению в v1.2, но [**еще не реализована**](https://github.com/werf/werf/issues/3318). Поэтому, если ваша конфигурация использует такие mounts, предлагаются следующие варианты:
 1. Продолжить использование версии v1.1, пока данная улучшенная возможность кеширования не [будет добавлена в werf v1.2](https://github.com/werf/werf/issues/3318).
 2. Удалить использование таких mounts и подождать [добавления улучшенной возможности кеширования в werf v1.2](https://github.com/werf/werf/issues/3318).
 3. Включить используемые mounts в werf v1.2 с помощью конфигурационного файла [`werf-giterminism.yaml`](https://werf.io/documentation/reference/werf_giterminism_yaml.html) (**не рекомендуется**).

## 1. Использовать стратегию тегирования content-based

[Стратегия тегирования content-based]({{ "/whats_new_in_v1_2/changelog.html#стратегия-тегирования" | true_relative_url }}) — это выбор по умолчанию и единственно доступный вариант, используемый в werf во время процесса деплоя.

 - Необходимо удалить параметр `--tagging-strategy` команды `werf ci-env`.
 - Удалены опции `--tag-custom`, `--tag-git-tag`, `--tag-git-branch`, и `--tag-by-stages-signature`.
 - В случае если требуется чтобы определённый docker tag для собранного образа существовал в container registry для использования вне werf, тогда допустимо использование параметров `--report-path` и `--report-format` следующим образом:
     - `werf build/converge --report-path=images-report.json --repo REPO`;
     - `docker pull $(cat images-report.json | jq -r .Images.IMAGE_NAME_FROM_WERF_YAML.DockerImageName)`;
     - `docker tag $(cat images-report.json | jq -r .Images.IMAGE_NAME_FROM_WERF_YAML.DockerImageName) REPO:mytag`;
     - `docker push REPO:mytag`.

## 2. Использовать команду converge, заменить команду build-and-publish

 - Команда `werf converge` должна быть использована вместо `werf deploy` сразу для сборки, публикации и деплоя приложения в kubernetes.
     - Команда поддерживает те же параметры, что и `werf deploy`.
 - Команда `werf build-and-publish` была удалена.
     - Добавлена команда `werf build`, использование которой опционально.
         - `werf converge` самостоятельно собирает и публикует в container registry недостающие образы.
         - `werf build` же может быть использована на стадии `prebuild` в CI/CD pipeline вместо `werf build-and-publish`.

 - Use `werf converge` command instead of `werf deploy` to build and publish images, and deploy into kubernetes.
     - All params are the same as for `werf deploy`.
 - `werf build-and-publish` command has been removed, there is only `werf build` command, usage of which is optional:
     - `werf converge` will build and publish needed images by itself, if not built already.
     - You may use `werf build` command in "prebuild" CI/CD pipeline stage instead of `werf build-and-publish` command.

## 3. Изменить helm шаблоны

 - Вместо `werf_container_image` используется `.Values.werf.image.IMAGE_NAME`:

    {% raw %}
    ```
    spec:
      template:
        spec:
          containers:
            - name: main
              image: {{ .Values.werf.image.myimage }}
    ```
    {% endraw %}

 - Полностью удалена шаблонная функция  `werf_container_env`.
 - Необходимо использовать `.Values.werf.env` вместо `.Values.global.env`.
 - Необходимо использовать `"werf.io/replicas-on-creation": "NUM"` вместо `"werf.io/set-replicas-only-on-creation": "true"`.
 - Use `"werf.io/replicas-on-creation": "NUM"` annotation instead of `"werf.io/set-replicas-only-on-creation": "true"`.
     - **ВАЖНО.** `"NUM"` должно быть указано **строкой**, а не как число `NUM`, иначе аннотация [будет проигнорирована]({{ "/reference/deploy_annotations.html#replicas-on-creation" | true_relative_url }}).
     - **ВАЖНО.** При использовании данной аннотации необходимо удалить явное определение поля `spec.replicas`, [больше информации в changelog]({{ "/whats_new_in_v1_2/changelog.html#конфигурация" | true_relative_url }}).

## 4. Использовать .helm/Chart.lock для сабчартов     

 - В соответствии с [режимом гитерминизма]({{ "/whats_new_in_v1_2/changelog.html#гитерминизм" | true_relative_url  }}) werf не позволяет держать некоммитнутую директорию `.helm/charts/`.
 - Для использования сабчартов необходимо определить dependencies в `.helm/Charts.yaml` следующим способом:

    {% raw %}
    ```yaml
    # .helm/Chart.yaml
    apiVersion: v2
    dependencies:
     - name: redis
       version: "12.7.4"
       repository: "https://charts.bitnami.com/bitnami"
    ```
    {% endraw %}

     - **ВАЖНО.** Следует обратить внимание на `apiVersion: v2` для использования описания зависимостей прямо в `.helm/Chart.yaml` вместо `.helm/dependencies.yaml`.

 - Директория `.helm/charts/` добавляется в `.gitignore`.
 - Далее следует запустить `werf helm dependency update .helm` для создания файла `.helm/Chart.lock` и директории `.helm/charts/` в локальной файловой системе.
 - `.helm/Chart.lock` надо коммитнуть в git репозиторий проекта.
 - werf автоматически скачает указанные сабчарты в кеш и загрузит файлы сабчартов во время команды `werf converge` (и других высокоуровневых командах, которые используют helm чарт).
 - Больше информации доступно [в документации]({{ "advanced/helm/configuration/chart_dependencies.html" | true_relative_url }}).

## 5. Использовать очистку на основе истории git

 - Опцию `--git-history-based-cleanup-v1.2` необходимо удалить, теперь поведение `werf cleanup` по умолчанию совпадает с поведением v1.1 с данной опцией.
 - Больше информации по изменению [в changelog]({{ "/whats_new_in_v1_2/changelog.html#очистка" | true_relative_url }}) и [в статье про cleanup]({{ "/advanced/cleanup.html" | true_relative_url }}).

## 6. Определить используемые переменные окружения в werf-giterminism.yaml

 - **ЗАМЕЧАНИЕ.** Вообще говоря использование переменные окружения в `werf.yaml` не рекомендуется (за исключением редких случаев), [больше информации в статье]({{ "/advanced/giterminism.html" | true_relative_url }}).
 - Если в `werf.yaml` используются например переменные окружения `ENV_VAR1` и `ENV_VAR2`, то необходимо добавить следующее описание в `werf-giterminism.yaml`:

    {% raw %}
    ```
    # werf-giterminism.yaml
    giterminismConfigVersion: 1
    config:
      goTemplateRendering:
        allowEnvVariables:
          - ENV_VAR1
          - ENV_VAR2
    ```
    {% endraw %}

## 7. Скоректировать конфигурационный файл werf.yaml

 - Относительные пути для включения дополнительных шаблонов меняются с {% raw %}{{ include ".werf/templates/1.tmpl" . }}{% endraw %} на {% raw %}{{ include "templates/1.tmpl" . }}{% endraw %}.
 - Переименовать `fromImageArtifact` в `fromArtifact`.
     - **ЗАМЕЧАНИЕ.** Не обязательно, но желательно заменить `artifact` и `fromArtifact` на `image` и `fromImage` в данном случае. [Заметка про запрет использования `fromArtifact` в changelog]({{ "/whats_new_in_v1_2/changelog.html#werfyaml" | true_relative_url }}).

## 8. Исправить определение нестандартной директории чарта .helm в werf.yaml

 - Вместо опции `--helm-chart-dir` используем директиву `deploy.helmChartDir` файла `werf.yaml` следующим образом:

    {% raw %}
    ```
    configVersion: 1
    deploy:
      helmChartDir: .helm2
    ```
    {% endraw %}

 - Следует рассмотреть другой вариант расположения конфигурации для вашего проекта: werf v1.2 поддерживает несколько werf.yaml в едином git репозитории.
     - Вместо переопределения нескольких разных `deploy.helmChartDir` в `werf.yaml` создаётся несколько `werf.yaml` в разных директориях проекта
     - Каждая такая директория содержит свою директорию `.helm` соответственно.
     - werf запускаем из каждой директории.
     - Все относительные пути указанные в `werf.yaml` будут рассчитаны относительно той директории, где лежит `werf.yaml`.
     - Абсолютные пути указанные в директиве `git.add` так и остаются абсолютными путями относительно корня git репозитория независимо от положения `werf.yaml` внутри этого репозитория.

## 9. Мигрировать на helm 3

 - werf v1.2 [осуществляет автоматическую миграцию существующего релиза с helm 2 на helm 3]({{ "/whats_new_in_v1_2/changelog.html#helm-3" | true_relative_url }}).
     - Релиз helm 2 должен иметь то же самое имя, как и вновь создаваемый релиз helm 3.
 - Перед запуском процесса миграции werf проверит что текущие шаблоны из `.helm/templates` рендерятся и валидируются корректно и только после успешного рендеринга миграция будет начата.
 - **ВАЖНО.** Как только проект переехал на helm 3 не существует обратного пути на werf v1.1 и helm 2.
