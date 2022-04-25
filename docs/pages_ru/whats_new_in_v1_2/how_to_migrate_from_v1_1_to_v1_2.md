---
title: Как мигрировать с v1.1 на v1.2
permalink: whats_new_in_v1_2/how_to_migrate_from_v1_1_to_v1_2.html
description: How to migrate your application from v1.1 to v1.2
sidebar: documentation
---

Чтобы мигрировать ваш проект с werf v1.1 на werf v1.2 следуйте приведенным ниже шагам.

**ВНИМАНИЕ!** Существуют такие конфигурации v1.1, которые используют директивы **`mount from build_dir`** и **`mount fromPath`** для кеширования внешних зависимостей приложения (зависимости, описываемые в таких файлах, как `Gemfile`, `package.json`, `go.mod` и т.п.). Такие директивы `mount` всё ещё поддерживаются в werf v1.2, однако не рекомендуются к использованию и по умолчанию выключены [режимом гитерминизма]({{ "/whats_new_in_v1_2/changelog.html#гитерминизм" | true_relative_url }}) (за исключением `mount from tmp_dir` — данная директива не ограничена к использованию), потому что использование этих директив может привести к ненадёжным, недетерминированным сборкам, зависимым от сборочных хостов. Для таких конфигураций планируется добавление новой возможности хранения такого кеша в Docker-образах. Данная фича планируется к добавлению в v1.2, но [**еще не реализована**](https://github.com/werf/werf/issues/3318).

Поэтому, если ваша конфигурация использует такие `mounts`, то предлагаются следующие варианты:
1. Продолжить использование версии v1.1, пока данная улучшенная возможность кеширования не [будет добавлена в werf v1.2](https://github.com/werf/werf/issues/3318).
2. Удалить использование таких `mounts` и подождать [добавления улучшенной возможности кеширования в werf v1.2](https://github.com/werf/werf/issues/3318).
3. Включить используемые `mounts` в werf v1.2 с помощью конфигурационного файла [`werf-giterminism.yaml`](https://werf.io/documentation/reference/werf_giterminism_yaml.html) (**не рекомендуется**).

## 1. Используйте стратегию тегирования content-based

[Стратегия тегирования content-based]({{ "/whats_new_in_v1_2/changelog.html#стратегия-тегирования" | true_relative_url }}) — это выбор по умолчанию и единственно доступный вариант, используемый в werf во время процесса деплоя.

 - Удалите параметр `--tagging-strategy` команды `werf ci-env`.
 - Удалите опции `--tag-custom`, `--tag-git-tag`, `--tag-git-branch`, и `--tag-by-stages-signature`.

 В случае, если требуется, чтобы определённый Docker-тег для собранного образа существовал в container registry для использования вне werf, тогда допустимо использование параметров `--report-path` и `--report-format` следующим образом:
 - `werf build/converge --report-path=images-report.json --repo REPO`;
 - `docker pull $(cat images-report.json | jq -r .Images.IMAGE_NAME_FROM_WERF_YAML.DockerImageName)`;
 - `docker tag $(cat images-report.json | jq -r .Images.IMAGE_NAME_FROM_WERF_YAML.DockerImageName) REPO:mytag`;
 - `docker push REPO:mytag`.

## 2. Используйте команду `converge` вместо команды `build-and-publish`

Теперь вместо `werf deploy` должна использоваться команда `werf converge`. Она используется сразу для сборки, публикации и деплоя приложения в Kubernetes и поддерживает те же параметры, что и `werf deploy`.

Изменения в командах (для справки):
 - Команда `werf build-and-publish` удалена.
 - Добавлена команда `werf build`, использование которой опционально.
 - `werf converge` самостоятельно собирает и публикует в container registry недостающие образы.
 - `werf build` также может быть использована на стадии `prebuild` в CI/CD pipeline вместо `werf build-and-publish`.

## 3. Измените Helm-шаблоны

 Вместо `werf_container_image` используйте `.Values.werf.image.IMAGE_NAME`:
 {% raw %}
 ```yaml
 spec:
   template:
     spec:
       containers:
         - name: main
           image: {{ .Values.werf.image.myimage }}
 ```
 {% endraw %}

> **ВАЖНО!** Убедитесь, что используете `apiVersion: v2`! Это необходимо, чтобы разрешить описание зависимостей непосредственно в `.helm/Chart.yaml` вместо `.helm/dependencies.yaml`.

Изменения в Helm-шаблонах (для справки):
 - Полностью удалена шаблонная функция  `werf_container_env`.
 - Необходимо использовать `.Values.werf.env` вместо `.Values.global.env`.
 - Необходимо использовать `.Values.werf.namespace` вместо `.Values.global.namespace`.
 - Необходимо использовать `"werf.io/replicas-on-creation": "NUM"` вместо `"werf.io/set-replicas-only-on-creation": "true"`.
   > **ВАЖНО!** `"NUM"` должно быть указано **строкой**, а не как число `NUM`, иначе аннотация [будет проигнорирована]({{ "/reference/deploy_annotations.html#replicas-on-creation" | true_relative_url }}).

   > **ВАЖНО!** При использовании данной аннотации необходимо удалить явное определение поля `spec.replicas`, [больше информации в changelog]({{ "/whats_new_in_v1_2/changelog.html#конфигурация" | true_relative_url }}).

## 4. Используйте `.helm/Chart.lock` для сабчартов

 В соответствии [с режимом гитерминизма]({{ "/whats_new_in_v1_2/changelog.html#гитерминизм" | true_relative_url  }}) werf не позволяет держать некоммитнутую директорию `.helm/charts/`. Для использования сабчартов необходимо определить `dependencies` в `.helm/Chart.yaml` следующим способом:

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

> **ВАЖНО!** Убедитесь, что используете `apiVersion: v2`! Это необходимо, чтобы разрешить описание зависимостей непосредственно в `.helm/Chart.yaml` вместо `.helm/dependencies.yaml`.

Далее выполните следующие шаги:

 1. Добавьте директорию `.helm/charts/` в `.gitignore`.
 2. Запустите `werf helm dependency update .helm` для создания файла `.helm/Chart.lock` и директории `.helm/charts/` в локальной файловой системе.
 3. Коммитните `.helm/Chart.lock` в Git-репозиторий проекта.

 werf автоматически скачает указанные сабчарты в кеш и загрузит файлы сабчартов во время команды `werf converge` (и других высокоуровневых командах, которые используют Helm-чарт). Больше информации доступно [в документации]({{ "advanced/helm/configuration/chart_dependencies.html" | true_relative_url }}).

## 5. Используйте очистку на основе истории Git

 Удалите опцию `--git-history-based-cleanup-v1.2` — теперь поведение `werf cleanup` по умолчанию совпадает с поведением v1.1 с данной опцией. Больше информации по изменению [в changelog]({{ "/whats_new_in_v1_2/changelog.html#очистка" | true_relative_url }}) и [в статье про cleanup]({{ "/advanced/cleanup.html" | true_relative_url }}).

## 6. Определите используемые переменные окружения в `werf-giterminism.yaml`

> **ЗАМЕЧАНИЕ!** На самом деле использование переменных окружения в `werf.yaml` не рекомендуется (за исключением редких случаев), больше информации [в статье]({{ "/advanced/giterminism.html" | true_relative_url }}).

Если в `werf.yaml` используются, например, переменные окружения `ENV_VAR1` и `ENV_VAR2`, то необходимо добавить следующее описание в `werf-giterminism.yaml`:

```yaml
# werf-giterminism.yaml
giterminismConfigVersion: 1
config:
  goTemplateRendering:
    allowEnvVariables:
      - ENV_VAR1
      - ENV_VAR2
```

## 7. Скоректируйте конфигурационный файл `werf.yaml`

 1. Замените относительные пути для включения дополнительных шаблонов с {% raw %}{{ include ".werf/templates/1.tmpl" . }}{% endraw %} на {% raw %}{{ include "templates/1.tmpl" . }}{% endraw %}.
 1. Переименуйте `fromImageArtifact` в `fromArtifact`.
    > **ЗАМЕЧАНИЕ!** Не обязательно, но желательно заменить `artifact` и `fromArtifact` на `image` и `fromImage` в данном случае. [Заметка про запрет использования `fromArtifact` в changelog]({{ "/whats_new_in_v1_2/changelog.html#werfyaml" | true_relative_url }}).

## 8. Исправьте определение нестандартной директории чарта `.helm` в `werf.yaml`

Вместо опции `--helm-chart-dir` используйте директиву `deploy.helmChartDir` файла `werf.yaml` следующим образом:

{% raw %}
```yaml
configVersion: 1
deploy:
  helmChartDir: .helm2
```
{% endraw %}

Следует рассмотреть другой вариант расположения конфигурации для вашего проекта: werf v1.2 поддерживает несколько `werf.yaml` в едином Git-репозитории.
  - Вместо переопределения нескольких разных `deploy.helmChartDir` в `werf.yaml` создаётся несколько `werf.yaml` в разных директориях проекта.
  - Каждая такая директория содержит свою директорию `.helm` соответственно.
  - werf необходимо запускать из каждой директории.
  - Все относительные пути, указанные в `werf.yaml`, будут рассчитаны относительно той директории, где лежит `werf.yaml`.
  - Абсолютные пути, указанные в директиве `git.add`, так и остаются абсолютными путями относительно корня Git-репозитория независимо от положения `werf.yaml` внутри этого репозитория.

## 9. Мигрируйте на Helm 3

werf v1.2 [осуществляет автоматическую миграцию существующего релиза с Helm 2 на Helm 3]({{ "/whats_new_in_v1_2/changelog.html#helm-3" | true_relative_url }}).

> Релиз Helm 2 должен иметь то же самое имя, как и вновь создаваемый релиз Helm 3.

Перед запуском процесса миграции werf проверит, что текущие шаблоны из `.helm/templates` рендерятся и валидируются корректно, и только после успешного рендеринга миграция будет начата.

> **ВАЖНО!** Как только проект переехал на Helm 3 – не существует обратного пути на werf v1.1 и Helm 2.
