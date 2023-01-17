---
title: Бандлы
permalink: usage/distribute/bundles.html
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

**Примечание**. На данный момент не поддерживается, но в дальнейшем планируется добавить оператор бандлов для kubernetes, для полностью автоматизированного выката в kubernetes, а также автовыкат новой версии по маске semver.

## Публикация бандлов

Публикация бандла осуществляется с помощью команды [werf-bundle-publish]({{ "/reference/cli/werf_bundle_apply.html" | true_relative_url }}). Команда принимает параметры по аналогии с [werf-converge]({{ "/reference/cli/werf_converge.html" | true_relative_url }}) и **требует запуска в git проекта**.

Во время работы `werf bundle publish`:

- будут собраны недостающие образы в container registry (также как и в [werf-converge]({{ "/reference/cli/werf_converge.html" | true_relative_url }}));
- будут опубликован [helm chart]({{ "/usage/deploy/charts.html" | true_relative_url }}) проекта в container registry;
- werf запомнит в опубликованном бандле переданные параметры [values для helm chart]({{ "/usage/deploy/values.html" | true_relative_url }});
- werf запомнит в опубликованном бандле переданные параметры аннотаций и лейблов для ресурсов.

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

При публикации бандла по уже существующему тегу — он будет переопубликован с обновлениями в container registry. Данная возможность может быть использована для организации автоматических обновлений приложения при наличии новой версии в container registry.

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

## Поддерживаемые container registries

Для работы с бандлами достаточно поддержки [спецификации формата образов Open Container Initiative (OCI)](https://github.com/opencontainers/image-spec) в container registry.

|                             |                       |
| --------------------------- |:---------------------:|
| _AWS ECR_                   | **ок**                |
| _Azure CR_                  | **ок**                |
| _Default_                   | **ок**                |
| _Docker Hub_                | **ок**                |
| _GCR_                       | **ок**                |
| _GitHub Packages_           | **ок**                |
| _GitLab Registry_           | **ок**                |
| _Harbor_                    | **ок**                |
| _JFrog Artifactory_         | **ок**                |
| _Nexus_                     | **не проверялся**     |
| _Quay_                      | **не поддерживается** |
| _Yandex Container Registry_ | **ок**                |
| _Selectel CRaaS_            | **не проверялся**     |
