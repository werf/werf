---
title: Changelog
permalink: whats_new_in_v1_2/changelog.html
description: Description of key differences since v1.1
sidebar: documentation
---

Данная статья содержит полный список изменений с версии v1.1. Если требуется инструкция по переводу проекта с v1.1, тогда следует обратиться [к статье про миграцию с v1.1 на v1.2]({{ "/whats_new_in_v1_2/how_to_migrate_from_v1_1_to_v1_2.html" | true_relative_url }}).

## Гитерминизм

werf вводит так называемый режим гитерминизма. Название происходит от слов git и детерминизм, что означает "детерминированный git'ом".

Все конфигурационные файлы werf, конфигурация helm и файлы приложения werf будет читать из текущего коммита git репозитория проекта. Существует также [режим разработки](#follow-и-dev) для упрощения локальной разработки конфигурации werf и локальной разработки приложения с использованием werf.

Больше информации о режиме доступно [на странице документации]({{ "/advanced/giterminism.html" | true_relative_url }}).

Для конфигурации гитерминизма в werf также был добавлен новый конфигурационный файл [`werf-giterminism.yaml`]({{ "/reference/werf_giterminism_yaml.html" | true_relative_url }}).

## Локальная разработка

### Follow и dev

Все высокоуровневые команды: `werf converge`, `werf run`, `werf bundle publish`, `werf render` и `werf build` — имеют 2 основных флага `--follow` и `--dev` для локальной разработки.

По умолчанию каждая из этих команд читает все требуемые файлы из текущего коммита git репозиторию проекта. С флагом `--dev` команда будет также использовать tracked/modified файлы из рабочей директории git проекта (использование untracked файлов ограничено).

Также существует и третий флаг `--dev-mode simple|strict` — режим работы флага `--dev`:
 - в режиме `simple` werf читает все tracked/modified файлы рабочей git директории проекта (наличие untracked файлов может вызвать ошибку, если werf требуется один из untracked файлов);
 - в режиме `strict` werf читает только те файлы, которые были добавлены в stage для коммита (файлы из git индекса рабочей директории проекта, которые были добавлены пользователем командой `git add`);
 - по умолчанию используется режим `simple`, когда указан флаг `--dev`.

Команда с флагом `--follow` будет работать в цикле, в котором:
 - новый коммит в рабочую директорию git проекта вызовет перезапуск команды — по умолчанию;
 - изменения в git-индексе (файлы добавленные через команду `git add`) рабочей директории git вызовут перезапуск команды — когда флаг `--follow` совмещён с флагом `--dev` (вне зависимости от используемого режима `--dev-mode`: `strict` или `simple`).

Внутри werf создаёт специальные коммиты для modified/tracked и staged файлов из git-индекса в ветке `werf-dev-<commmit>`. Кеш в режиме разработки будет привязан к этим временным коммитам и тем самым отделён от основного сборочного кеша (когда не указано флага `--dev`).

### Поддержка composer

werf поддерживает основные команды composer. При запуске этих команд werf автоматом установит специальные переменные окружения, в которых доступны полные имена docker-образов для образов, описанных в `werf.yaml`:

```
WERF_<FORMATTED_WERF_IMAGE_NAME>_DOCKER_IMAGE_NAME=application:45f03bdd90c844eb2e61e7e01dae491588d2bdadbd195881b25be9b0-1613371915351
```

Например, если имеется следующий `werf.yaml`:

```yaml
# werf.yaml
project: myproj
configVersion: 1
---
image: myimage
from: alpine
```

Полное имя docker-образа для `myimage` может быть использовано в `docker-compose.yml` следующим образом:

```yaml
version: '3'
services:
  web:
    image: "${WERF_MYIMAGE_DOCKER_IMAGE_NAME}"
```

werf предоставляет следующие команды compose:
 - [`werf compose config`]({{ "/reference/cli/werf_compose_config.html" | true_relative_url }}) — запуск команды `docker-compose config` с прокинутыми именами образов из `werf.yaml`;
 - [`werf compose down`]({{ "/reference/cli/werf_compose_down.html" | true_relative_url }}) — запуск команды `docker-compose down` с прокинутыми именами образов из `werf.yaml`;
 - [`werf compose up`]({{ "/reference/cli/werf_compose_up.html" | true_relative_url }}) — запуск команды `docker-compose up` с прокинутыми именами образов из `werf.yaml`;

## Новый функционал бандлов

 - Бандлы позволяют разделить процесс создания нового релиза кода приложения и процесс деплоя этого релиза в кластер kubernetes.
 - Бандлы хранятся в container registry.
 - Работа с бандлами предполагает 2 основных шага: 1) публикация бандла для приложения из рабочей директории git проекта в container registry; 2) деплой ранее опубликованного бандла из container registry в кластер kubernetes.
 - Больше информации [в документации]({{ "/advanced/bundles.html" | true_relative_url }}).
 
## Переработка основного командного интерфейса и поведения команд

### Переработка интерфейса

 - Единая команда `werf converge` для сборки и публикации требуемых образов в container registry и деплоя приложения в kubernetes.
 - Single `werf converge` command to build and publish needed images and deploy application into the kubernetes.
     - Вызов команды с опцией `werf converge --skip-build` эмулирует поведение ранее существующей команды `werf deploy`.
         - werf упадёт с ошибкой если требуемые образы не будут найдены в container registry, так же как падал с ошибкой `werf deploy`.
 - Удалены команды `werf stages *`, `werf images *` и `werf host project *`.
 - Более нет команды `werf publish`, потому что команда `werf build` с параметром `--repo` загрузит образы и все стадии, из которых они состоят, в container registry автоматически.
 - Более нет флагов тегирования: `--tag-by-stages-signature`, `--tag-git-branch`, `--tag-git-commit`, `--tag-git-tag` и `--tag-custom`, werf всегда использует поведение ранее включаемое флагом `--tag-by-stages-signature`.
     - Принудительное использование произвольных тегов [пока не поддерживается](https://github.com/werf/werf/issues/2869) в v1.2.
     - Команда `werf ci-env` принимает скрытый параметр `--tagging-strategy` по соображениям совместимости для следующего варианта вызова: `werf ci-env --tagging-strategy` без опции `--as-file`.
         - Использование данного флага должно быть удалено из всех вызовов команды `werf ci-env`.
         - Указанный параметр не влияет на поведение werf, поведение соответствующее ранее существовавшему флагу `--tag-by-stages-signature` будет использовано в любом случае.

### Поведение команд werf-build и werf-converge

 - По умолчанию команда `werf build` не требует никаких аргументов и будет собирать образы используя локальный docker server в качестве хранилища.
 - При запуске `werf build` с флагом `--repo registry.mydomain.org/project` werf будет искать локально собранные образы и если найдёт, то загрузит их в указанный repo (лишней пересборки не будет).
 - Команда `werf converge` требует параметр `--repo` для работы, и, так же как `werf build`, автоматически загрузит в repo локально существующие образы.

### Поведение команд очистки

- Команды `werf cleanup/purge` используются только для чистки сontainer registry.
- Команда `werf host purge --project=<project-name>` может использоваться для удаления образов проекта из локального Docker (ранее для этого можно было использовать команду `werf purge`).

### Автоматическая сборка образов в командах верхнего уровня

 - Команды `werf converge`, `werf run`, `werf bundle publish` и `werf render` автоматически соберут недостающие образы, которые не существуют в container registry.
 - Команда `werf render` будет собирать образы только при передаче опционального параметра `--repo`.

### Хранение образов в CI/CD системах

В качестве параметра `--repo` предполагается использование container registry привязанного к проекту.

 - Команда `werf ci-env` для GitLab CI/CD, например, сгенерирует следующий параметр: `WERF_REPO=$CI_REGISTRY_IMAGE`.
 - Указанный репозиторий `--repo` будет использован для хранения как собранных промежуточных стадий образа, так и финального образа (внутри эти слои переиспользуются и публикация промежуточных стадий не создаёт накладных расходов).

## werf всегда хранит стадии в container registry

 - Команда `werf converge` всегда хранит образы и стадии, из которых они состоят, в container registry.
 - Единый параметр `--repo` используется для указания хранилища образов и стадий этих образов (ранее в v1.1 было 2 параметра `--stages-storage` и `--images-repo`).
 - Благодаря использованию content-based тегирования финальный образ совпадает с последней сборочной стадией этого образа.
     - Т.к. образы состоят из набора слоёв, все эти слои всё равно хранятся в container registry.
         - Поэтому хранение промежуточных стадий в container registry вместе с image не создаёт накладных расходов.
         - Это делает процесс сборки образов в werf независимым от сборочного хоста, потому что в процессе сборки промежуточные стадии будут переиспользованы из container registry в качестве сборочного кеша.

## Изменения сигнатур

 - **Сигнатуры переименованы в дайджесты** (signature => digest).
 - Все дайджесты уже собранных образов изменились.
 - Образы собранные через сборщик как Dockerfile, так и stapel, будут пересобраны.
     - Другими словами сборочный кеш из v1.1 более не валиден.

## Сборщик stapel

 - Удалена возможность кеширования каждой сборочной инструкции из `werf.yaml` по отдельным слоям с помощью директивы `asLayers`.

## Опциональный параметр env

 - Для команд `werf converge`, `werf render` and `werf bundle publish` существует параметр `--env`.
 - `--env` [влияет на имя helm релиза и kubernetes namespace]({{ "/advanced/helm/releases/naming.html" | true_relative_url }}) также как в v1.1.
     - При указании параметра `--env` имя [helm релиза]({{ "/advanced/helm/releases/release.html" | true_relative_url }}) будет сгенерировано по шаблону `[[ project ]]-[[ env ]]` и namespace в kubernetes будет сгенерирован по такому же шаблону `[[ project ]]-[[ env ]]` — так же как в версии v1.1.
     - Когда параметр `--env` не указан, werf будет использовать шаблон `[[ project ]]` для генерации имени [helm релиза]({{ "/advanced/helm/releases/release.html" | true_relative_url }}) и такой же шаблон `[[ project ]]` для генерации namespace в kubernetes.

## Новая команда werf-render

 - Команда [`werf render`]({{ "/reference/cli/werf_render.html" | true_relative_url }}) автоматически соберёт недостающие образы в container registry.
 - Данная команда позволяет воспроизвести шаблоны в точно таком же виде, как они будут сгенерированы в команде [`werf converge`]({{ "/reference/cli/werf_converge.html" | true_relative_url }}).
 - Команда [`werf render`]({{ "/reference/cli/werf_render.html" | true_relative_url }}) работает [в режиме гитерминизма](#гитерминизм).

## Helm

### Helm 3

 - Helm 3 используется по умолчанию, и это единственная версия helm доступная для использования в v1.2.
 - Уже существующие релизы helm 3 будут смигрированы на helm 3 автоматически в команде `werf converge` при условии, что имя релиза helm 2 совпадает с именем нового релиза helm 3.
     - **ПРЕДУПРЕЖДЕНИЕ.** Как только релиз helm 2 конвертирован в helm 3 пути назад нет.
     - Прежде чем мигрировать релиз helm 2 в helm 3, команда `werf converge` проверит, что текущие шаблоны корректно рендерятся и валидируются.

### Конфигурация

 - `.Values.werf.image.IMAGE_NAME` вместо шаблона `werf_image`.
 - Удалена функция шаблонов `werf_container_env`.
 - Загрузка всех конфигурационных файлов чарта происходит в режиме [гитерминизма](#гитерминизм).
     - Данный режим используется только для высокоуровневых команд вроде `werf converge` и `werf render`.
     - Низкоуровневые helm-команды `werf helm *` работают в обычном режиме и загружают файлы из локальной файловой системы.
 - Окружение, переданное опцией `--env`, доступно для использования через `.Values.werf.env`.
 - Исправлен подход к использованию файла `.helm/Chart.yaml`. Файл `.helm/Chart.yaml` опционален, однако werf будет его использовать, если он существует следующим образом:
     - Файл `.helm/Chart.yaml` берётся из git репозитория проекта, если он существует.
     - Поле `metadata.name` перезаписывается именем проекта из `werf.yaml`.
     - Поле `metadata.version` устанавливается в `1.0.0`, если оно явно не определено.
 - Добавлено [сервисное значение `.Values.werf.version`]({{ "/advanced/helm/configuration/values.html#сервисные-данные" | true_relative_url }}) с версией утилиты werf, которая используется.
 - Поддерживается установка начального количества реплик, когда активен режим HPA для Deployment и других типов ресурсов. Необходимо установить аннотацию `"werf.io/replicas-on-creation": NUM` и убрать явное определение `spec.replicas` в шаблонах в таком случае.
     - `spec.replicas` переопределяет `werf.io/replicas-on-creation`.
     - Данная аннотация особенно полезна при использовании HPA, [в данной статье описано почему]({{ "/reference/deploy_annotations.html#replicas-on-creation" | true_relative_url }}).

### Работа с сабчартами

 - Лок-файл с зависимостями `.helm/Chart.lock` должен быть коммитнут в git репозиторий проекта.
 - werf автоматически скачает все зависимости определённые в лок-файле.
 - Обычно директория `.helm/charts/` должна быть добавлена в `.gitignore`.
 - Однако werf позволяет хранить зависимые чарты и в директории `.helm/charts`.
     - Данные чарты перезапишут автоматически загруженные по `.helm/Chart.lock` чарты.
 - См. больше информации [по зависимым чартам]({{ "/advanced/helm/configuration/chart_dependencies.html" | true_relative_url }}) и [гитерминизму]({{ "/advanced/helm/configuration/giterminism.html#сабчарты-и-гитерминизм" | true_relative_url }}).

### Удалена команда werf-helm-deploy-chart

Вместо команды `werf helm deploy-chart` можно использовать команды `werf helm install` и `werf helm upgrade`.

### Лучшая интеграция с командами werf-helm-*

 - Команда [`werf helm template`]({{ "/reference/cli/werf_helm_template.html" | true_relative_url }}) может быть использована для рендера локальных чартов даже при использовании секретов (функции шаблонов `werf_secret_file`).
 - Файлы секретов полностью поддерживаются командами `werf helm *`.
 - [Дополнительные аннотации и лейблы]({{ "/advanced/helm/deploy_process/annotating_and_labeling.html" | true_relative_url }}) полностью поддерживаются командами `werf helm *` (опции `--add-annotation` и `--add-label`).

### Изменён формат сервисных значений 

Удалено [сервисное значение]({{ "/advanced/helm/configuration/values.html#сервисные-данные" | true_relative_url }}) `.Values.global.werf.image`, вместо него используется `.Values.werf.image`.

### Процесс деплоя

 - Добавлено полноценное ожидание удаления ресурсов перед выходом в командах `werf dismiss` и `werf helm uninstall`.
 - Wait until all release resources terminated in the `werf dismiss` command before exiting.

## werf.yaml

 - Директива `fromImageArtifact` переименована в `fromArtifact`.
     - **ПРЕДУПРЕЖДЕНИЕ.** Директива `fromImageArtifact/fromArtifact` не рекомендована к использованию и будет удалена в v1.3 из-за неочевидной работы механизма наложения патчей в получившейся конфигурации. Рекомендуется уже сейчас перейти на `image` и `fromImage` вместо `artifact` и `fromArtifact`.
 - Директивы `herebyIAdmitThatFromLatestMightBreakReproducibility` и `herebyIAdmitThatBranchMightBreakReproducibility` удалены. Теперь при использовании директив `fromLatest` и `git.branch` необходимо ослаблять правила [гитерминизма]({{ "/advanced/giterminism.html" | true_relative_url }}), используя соответствующие директивы в [werf-giterminism.yaml]({{ "/reference/werf_giterminism_yaml.html" | true_relative_url }}).
 - Удалена опция `--helm-chart-dir`, директория helm-чарта определяется в `werf.yaml`:

    ```yaml
    configVersion: 1
    deploy:
      helmChartDir: .helm
    ```

 - Опции `--allow-git-shallow-clone`, `--git-unshallow` и `--git-history-synchronization` превращены в директивы `werf.yaml`, добавлена [новая мета-секция `gitWorktree`]({{ "/reference/werf_yaml.html#git-worktree" | true_relative_url }}). Выключено использование этих опций в команде `werf ci-env`: unshallow git clone происходит всегда.
 - Использование sprig v3 вместо v2: [http://masterminds.github.io/sprig/](http://masterminds.github.io/sprig/).
 - Новая страница документации, описывающая [движок шаблонов `werf.yaml`]({{ "/reference/werf_yaml_template_engine.html" | true_relative_url }}).
 - Пофикшено именование пользовательских шаблонов. Имя шаблона — это относительный путь в директории `.werf`. {% raw %}{{ include ".werf/templates/1.tmpl" . }} => {{ include "templates/1.tmpl" . }}{% endraw %}.
 - Определение безымянного образа в `werf.yaml` с помощью `image: ~` не рекомендовано к использованию и будет удалено в v1.3. Лучше всегда выставлять явные имена образам, даже если образ один.
 - Функция `env "ENVIRONMENT_VARIABLE_NAME"` теперь требует наличия переменной `ENVIRONMENT_VARIABLE_NAME` (переменная может содержать пустое значение, но она должна быть явно определена).
 - Новая функция `required` даёт разработчикам возможность определить значение, которое требуется для рендера конфига. Если это значение пустое, то рендер сообщит об ошибке с помощью текста указанного в директиве `required`:

    {% raw %}
    ```
    {{ required "A valid <anything> value required!" <anything> }}
    ```
    {% endraw %}

## Стратегия тегирования

 - Доступна только стратегия тегирования на основе контента.
 - Возможность экспорта собранных образов в произвольный container registry будет добавлено позже.
 - Принудительное использование произвольных тегов для собранных образов будет добавлено позже.

## Очистка

 - Очистка на основе истории git по умолчанию.
     - В v1.1 такой режим работы включался опцией `--git-history-based-cleanup-v1.2` (в v1.2 опция удалена).
 - Полностью удалены алгоритмы очистки на основе различных стратегий тегирования.
 - Опция `--keep-stages-built-within-last-n-hours` для сохранения образов, которые были собраны в указанный период до текущего момента (по умолчанию 2 часа).
 - Улучшения алгоритма очистки на основе истории git:
     - Политика сохранения образов по директиве `imagesPerReference.last` учитывает, что может существовать несколько стадий образа, основанных на одном и том же коммите. Эти образы сохраняются и засчитываются как один для директивы `imagesPerReference.last`. Другими словами: `imagesPerReference.last` — это про количество коммитом, для одного коммита может существовать несколько образов.

## Кеширование импортов из artifact/image по контрольной сумме

 - Представим, что имеется артефакт, который собирает некоторые файлы.
 - Данный артефакт имеет stage-dependency в `werf.yaml` для пересборки этих файлов, когда исходные зависимые файлы поменялись.
 - При очередном изменении исходных файлов вызывает пересборку артефакта.
     - Т.к. изменился stage-dependency, то последняя стадия артефакта будет иметь другой дайджест.
     - Однако после пересборки артефакта фактическая контрольная сумма файлов, которые были собраны не поменялась — получились точно такие же файлы.
         - В версии v1.1 werf всё равно выполнит переимпорт этих файлов в целевой образ.
             - Как последствие целевой образ будет пересобран и перевыкачен.
         - В версии v1.2 werf проверит контрольную сумму файлов, которые импортятся и перевыполнит импорт только при изменении этой контрольной суммы.

## Поддержка первичного и вторичного хранилища образов

 - Автоматическая загрузка собранных локально образов в указанный container registry в `--repo`.
 - Использование стадий из read-only вторичного хранилища указанного опциями `--secondary-repo` (опция может быть указана несколько раз).
     - Подходящие стадии из вторичного хранилища `--secondary-repo` будут скопированы в первичное хранилище `--repo`.

## Built images report format changes

Команды `werf converge`, `werf build`, `werf run`, `werf bundle publish` и `werf render` имеют опции `--report-path` и `--report-format`. `--report-path` включает генерацию в следующем формате:

```shell
$ werf build --report-path images.json --report-format json

{
  "Images": {
    "result": {
      "WerfImageName": "result",
      "DockerRepo": "quickstart-application",
      "DockerTag": "32e88a6a19a425c9254374ee2899b365876de31ac7d6857b523696a1-1613371915843",
      "DockerImageID": "sha256:fa6445196bc8ed44e4d2842eeb068aab4e627112d504334f2e56d235993ba4f0",
      "DockerImageName": "quickstart-application:32e88a6a19a425c9254374ee2899b365876de31ac7d6857b523696a1-1613371915843"
    },
    "vote": {
      "WerfImageName": "vote",
      "DockerRepo": "quickstart-application",
      "DockerTag": "45f03bdd90c844eb2e61e7e01dae491588d2bdadbd195881b25be9b0-1613371915351",
      "DockerImageID": "sha256:a6845bbc7912e45c601b0291170f9f503722efceb9e3cc98a5701ea4d26b017e",
      "DockerImageName": "quickstart-application:45f03bdd90c844eb2e61e7e01dae491588d2bdadbd195881b25be9b0-1613371915351"
    },
    "worker": {
      "WerfImageName": "worker",
      "DockerRepo": "quickstart-application",
      "DockerTag": "1b16118e7d5c67aa3c61fc0f8d49b3eccf8f72810f01c33a40290418-1613371916044",
      "DockerImageID": "sha256:546af94bd73dc20a7ab1f49562f42c547ee388fb75cf480eae9fde02b48ad6ad",
      "DockerImageName": "quickstart-application:1b16118e7d5c67aa3c61fc0f8d49b3eccf8f72810f01c33a40290418-1613371916044"
    }
  }
}
```

или

```shell
$ werf build --report-path images.sh --report-format envfile

WERF_RESULT_DOCKER_IMAGE_NAME=quickstart-application:32e88a6a19a425c9254374ee2899b365876de31ac7d6857b523696a1-1613371915843
WERF_WORKER_DOCKER_IMAGE_NAME=quickstart-application:1b16118e7d5c67aa3c61fc0f8d49b3eccf8f72810f01c33a40290418-1613371916044
WERF_VOTE_DOCKER_IMAGE_NAME=quickstart-application:45f03bdd90c844eb2e61e7e01dae491588d2bdadbd195881b25be9b0-1613371915351
```

## Поддержка нескольких werf.yaml в одном git репозитории проекта

Пользователь может создать несколько `werf.yaml` (и `.helm` соответственно) в одном git репозитории проекта.

 - Все относительные пути в `werf.yaml` будут рассчитаны относительно текущей рабочей директории процесса werf или параметра `--dir`.
 - Обычно пользователь запускает werf прямо из директории, в которой расположен `werf.yaml`.
 - Также существует опция `--config` для передачи кастомного `werf.yaml`, все относительные пути в таком случае также будут рассчитаны относительно текущей рабочей директории процесса werf или параметра `--dir`.
 - Добавлен параметр `--git-work-tree` (и переменная `WERF_GIT_WORK_TREE`) для указания директории, которая содержит `.git` для случая, когда автодетектор этой директории не смог её найти, или когда мы хотим явно указать другую рабочую директорию git проекта.
     - Например когда werf запускается из сабмодуля пользователь может захотеть использовать корневой git репозиторий проекта вместо git репозитория сабмодуля.
         - Чтобы сборочный кеш был связан с коммитами в главном репозитории.

 - All relative paths specified in the werf.yaml will be calculated relatively to the werf process cwd or `--dir` param.
 - Typically, user will run werf from the subdirectory where werf.yaml reside.
 - There is also `--config` option to pass custom werf.yaml config, all relative paths will also be calculated relatively to the werf process cwd or `--dir` param.
 - Added `--git-work-tree` param (or `WERF_GIT_WORK_TREE` variable) to specify a directory that contains `.git` in the case when autodetect failed, or we want to use specific work tree.
     - For example when running werf from the submodule of the project we may want to use root repo worktree instead of submodule's work tree.

## Разное и внутренности

 - Добавлен кеш git-архивов и git-патчей в `~/.werf/local_cache/` по аналогии с уже существующем кешом git-worktree. Патчи и архивы более не создаются в `/tmp/` (и `--tmp-dir` соответственно).
 - Удалена блокировка `stages_and_images` во время сборочного процесса.
     - `stages_and_images` — это вспомогательная блокировка, которая ранее предотвращала одновременный запуск `werf cleanup` и `werf build`.
     - В текущей версии данная блокировка может быть проигнорирована без последствий.
     - Данная блокировка создаёт излишнюю нагрузку на сервер синхронизации, потому что она активна во время всего процесса сборки.
 - Также удалена неиспользуемая более блокировка `image` (legacy из v1.1).
 - Также переименованы блокировки связанные с использованием kubernetes в качестве lock-storage (сделано для большей корректности, несовместимое с v1.1 изменение).

## Переработка документации

 - Вся документация разделена на секции по глубине погружения:
     - главная страница;
     - введение;
     - быстрый старт;
     - гайды/руководства;
     - справочник;
     - документация продвинутого уровня;
     - внутренности.
 - Новая секция с руководствами.
