---
title: Бандлы
permalink: advanced/bundles.html
change_canonical: true
---

Стандартный flow использования werf предполагает запуск команды [werf-converge]({{ "/reference/cli/werf_converge.html" | true_relative_url }}) для выката новой версии приложения в kubernetes. Во время работы [werf-converge]({{ "/reference/cli/werf_converge.html" | true_relative_url }}):
 - будут собраны недостающие образы в container registry;
 - состояние приложения в kubernetes будет актуализировано.

werf позволяет разделить процесс выпуска новой версии приложения готовой к выкату и сам процесс выката в kubernetes с помощью так называемых **бандлов**.

**Бандл** — это коллекция образов и конфигурационных файлов деплоя некоторой версии приложения, опубликованная в container registry и готовая к выкату в kubernetes. Выкат приложения с помощью бандлов предполагает 2 шага:
  1. Публикация бандла.
  2. Выкат бандла в kubernetes.

Бандлы позволяют организовать гибридный подход push-pull к выкату приложения. Push в данном случае — это про публикацию бандла, pull — про выкат приложения в kubernetes.

**Примечание**. На данный момент не поддерживается, но в дальнейшем планируется добавить оператор бандлов для kubernetes, для полностью автоматизированного выката в kubernetes, а также [автовыкат новой версии по маске semver](#выкат-по-маске-semver).

## Публикация бандлов

Публикация бандла осуществляется с помощью команды [werf-bundle-publish]({{ "/reference/cli/werf_bundle_apply.html" | true_relative_url }}). Команда принимает параметры по аналогии с [werf-converge]({{ "/reference/cli/werf_converge.html" | true_relative_url }}) и **требует запуска в git проекта**.

Во время работы `werf bundle publish`:
 - будут собраны недостающие образы в container registry (также как и в [werf-converge]({{ "/reference/cli/werf_converge.html" | true_relative_url }}));
 - будут опубликован [helm chart]({{ "/advanced/helm/configuration/chart.html" | true_relative_url }}) проекта в container registry;
 - werf запомнит в опубликованном бандле переданные параметры [values для helm chart]({{ "/advanced/helm/configuration/values.html" | true_relative_url }});
 - werf запомнит в опубликованном бандле переданные параметры [аннотаций и лейблов для ресурсов]({{ "/advanced/helm/deploy_process/annotating_and_labeling.html" | true_relative_url }})

### Структура bundle

Перед публикацией имеется следующая структура чарта в обычном werf-проекте:

```
.helm/
  Chart.yaml
  LICENSE
  README.md
  values.yaml
  values.schema.json
  templates/
  crds/
  charts/
  files/
```

По умолчанию в бандл будут запакованы только указанные выше файлы и директории.

`.helm/files` — это общепринятая директория для хранения конфигурационных файлов и шаблонов, которые используются в таких директивах как `.Files.Get` и `.Files.Glob` для вставки контента файлов в шаблоны ресурсов.
`.helm/values.yaml` будет совмещён с внутренними сервисными значениями werf и с другими нестандартными values указанными параметрами при вызове команды публикации бандла, совмещенные values будут записаны в `values.yaml` файл внутри бандла. Использование стандартного `.helm/values.yaml` можно отключить опцией `--disable-default-values` для команды `werf bundle publish` (в этом случае в бандле всё равно будет опубликован `values.yaml` файл, но он не будет содержать values из `.helm/values.yaml`).

Структура опубликованного чарта в бандле выглядит следующим образом:

```
Chart.yaml
LICENSE
README.md
values.yaml
values.schema.json
templates/
crds/
charts/
files/
extra_annotations.json
extra_labels.json
```

Обращаем внимание на файлы `extra_annotations.json` и `extra_labels.json` — это сервисные файлы, хранящие аннотации и лейблы зашитые в бандл, которые будут добавлены во все ресурсы релиза при выкате бандла через команду `werf bundle apply`.

Чтобы скачаь и проинспектировать структуру опубликованного бандла используется следующая команда: `werf bundle copy --from REPO:TAG --to archive:PATH_TO_ARCHIVE.tar.gz`.

### Версионирование при публикации

При публикации бандла пользователь может выбрать версию с помощью параметра `--tag`. По умолчанию бандл публикуется с тегом `latest`.

При публикации бандла по уже существующему тегу — он будет переопубликован с обновлениями в container registry. Данная возможность может быть использована для организации автоматических обновлений приложения при наличии новой версии в container registry. Подробнее см. [автообновление бандлов](#автообновление-бандлов).

## Выкат бандлов

Бандл, опубликованный в container registry, можно выкатить в kubernetes с помощью werf одним из следующих путей:
 1. С помощью команды `werf bundle apply` напрямую из container registry.
 2. Из другого проекта werf в качестве зависимого helm-чарта.

### Выкат с помощью werf bundle apply

Выкат опубликованного бандла осуществляется с помощью команды [`werf bundle apply`]({{ "/reference/cli/werf_bundle_apply.html" | true_relative_url }}). Команда **не требует доступа к git** (бандл содержит все необходимые файлы и образы для выката приложения), и принимает параметры по аналогии с [`werf converge`]({{ "/reference/cli/werf_converge.html" | true_relative_url }}). Для `werf bundle apply` необходимо явно указывать имя создаваемого Helm-релиза (`--release`) и namespace, в который будет совершён выкат (`--namespace`).

Переданные [values для helm chart]({{ "/advanced/helm/configuration/values.html" | true_relative_url }}), [аннотаций и лейблов для ресурсов]({{ "/advanced/helm/deploy_process/annotating_and_labeling.html" | true_relative_url }}) будут объединены с теми values, аннотациями и лейблами, которые были указаны при публикации бандлов и зашиты в него.

### Выкат в качестве зависимого helm-чарта

Настраиваем зависимость в целевом werf-проекте в файле `.helm/Chart.yaml`:

```yaml
apiVersion: v2
dependencies:
- name: project
  repository: "oci://ghcr.io/group"
  version: 1.4.x
```

Обновляем зависимости проекта (данная команда актуализирует файл `.helm/Chart.lock`):

```
werf helm dependency update .helm
```

Выкатываем проект с зависимостями:

```
werf converge --repo ghcr.io/group/otherproject
```

### Версионирование при выкате

При выкате пользователь может выбрать версию выкатываемого бандла с помощью параметра `--tag`. По умолчанию выкатывается бандл по тегу `latest`.

При выкате бандла werf проверит наличие обновления в container registry для указанного тега (или `latest`) и обновит приложение в кластере до последней версии. Данная возможность может быть использована для организации автоматических обновлений приложения при наличии новой версии в container registry. Подробнее см. [автообновление бандлов](#автообновление-бандлов).

## Другие команды для работы с бандлами

### Рендеринг манифестов бандла

Команда [`werf bundle render`]({{ "/reference/cli/werf_bundle_render.html" | true_relative_url }}) позволяет отрендерить Kubernetes-манифесты в окончательном виде, что может использоваться для дальнейшего деплоя с другим ПО (например, с ArgoCD) или для отладки.

Команда **не требует запуска в git проекта**. Требуется указать одну из двух опций:
1. `--repo`, чтобы отрендерить бандл из удалённого OCI-репозитория.
2. или `--bundle-dir`, чтобы отрендерить бандл, находящийся в указанной директории в распакованном виде.

Если собираетесь деплоить получившиеся манифесты, то следует указать значения для опций `--release` и `--namespace`, иначе в качестве release и namespace будут использованы заглушки.

Опцию `--output` можно использовать для сохранения результата в произвольный файл.

### Экспорт бандла в директорию

Команда [`werf bundle export`]({{ "/reference/cli/werf_bundle_export.html" | true_relative_url }}) позволяет создать директорию чарта в том виде, в котором он будет опубликован в container registry.

Команда работает по аналогии с командой [`werf bundle publish`]({{ "/reference/cli/werf_bundle_publish.html" | true_relative_url }}) имеет те же параметры и **требует запуска в git проекта**.

Данную команду можно использовать для того, чтобы узнать, из чего состоит публикуемый бандл, а также для дебага.

### Скачивание опубликованного бандла

Команда [`werf bundle download`]({{ "/reference/cli/werf_bundle_download.html" | true_relative_url }}) позволяет скачать ранее опубликованный бандл в директорию.

Команда работает по аналогии с командой [`werf bundle apply`]({{ "/reference/cli/werf_bundle_apply.html" | true_relative_url }}) имеет те же параметры и **не требует доступа к git**.

Данную команду можно использовать для того, чтобы узнать, из чего состоит публикуемый бандл, а также для дебага.

## Примеры использования

Опубликуем бандл для приложения по определённой версии, запускаем в git директории проекта:

```
git checkout v3.5.0
werf bundle publish --repo registry.mydomain.io/project --tag v3.5.0 --add-label "project-version=v3.5.0"
```

Опубликуем бандл для приложения по статичному тегу `main`, запускаем в git директории проекта:

```
git checkout main
werf bundle publish --repo registry.mydomain.io/project --tag main --set "myvalue.x=150"
```

Выкатим бандл опубликованный по тегу `v3.4.9`, запускаем на произвольном хосте с доступом в kubernetes и container registry:

```
werf bundle apply --repo registry.mydomain.io/project --tag v3.4.9 --env production --release myproject --namespace myproject --set "sentry.url=sentry.mydomain.io"
```

## Автообновление бандлов

werf поддерживает автоматический перевыкат бандла, опубликованного по статичному тегу вроде `latest` или `main`. Если публикация бандла настроена на выкат такого тега, то при каждом новом вызове публикации его версия по этому тегу в container registry будет обновлена.

Со стороны выката этого бандла при каждом вызове команды выката будет использоваться последняя актуальная версия по указанному тегу.

### Выкат по маске semver

На данный момент не поддерживается, [но планируется добавить](https://github.com/werf/werf/issues/3169) возможность выката бандла по маске semver:

```
werf bundle apply --repo registry.mydomain.io/project --tag-mask v3.5.* --release myproject --namespace myproject
```

Данный вызов выката бандла должен проверить наличие версий по указанной маске `v3.5.*`, выбрать последнюю версию в рамках `v3.5` и выкатить её.

## Поддерживаемые container registries

Для работы с бандлами достаточно поддержки [спецификации формата образов Open Container Initiative (OCI)](https://github.com/opencontainers/image-spec) в container registry.

Подробнее про поддерживаемые container registries можно прочитать в отдельной [статье]({{ "/advanced/supported_container_registries.html" | true_relative_url }}).
