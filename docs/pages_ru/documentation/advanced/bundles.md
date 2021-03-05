---
title: Bundles
permalink: documentation/advanced/bundles.html
change_canonical: true
---

Стандартный flow использования werf предполагает запуск команды [werf-converge]({{ "/documentation/reference/cli/werf_converge.html" | true_relative_url }}) для выката новой версии приложения в kubernetes. Во время работы [werf-converge]({{ "/documentation/reference/cli/werf_converge.html" | true_relative_url }}):
 - будут собраны недостающие образы в container registry;
 - состояние приложения в kubernetes будет актуализировано.

Werf позволяет разделить процесс выпуска новой версии приложения готовой к выкату и сам процесс выката в kubernetes с помощью **bundles**.

**Bundle** — это коллекция образов и конфигурационных файлов деплоя некоторой версии приложения, опубликованная в container registry и готовая к выкату в kubernetes. Выкат приложения с помощью bundles предполагает 2 шага:
  1. Публикация bundle.
  2. Выкат bundle в kubernetes.

Bundles позволяют организовать гибридный подход push-pull к выкату приложения. Push в данном случае — это про публикацию bundle, pull — про выкат приложения в kubernetes.

**Примечание**. На данный момент не поддерживается, но в дальнейшем планируется добавить bundles оператор для kubernetes, для полностью автоматизированного выката в kubernetes, а также [автовыкат новой версии по маске semver](#выкат-по-маске-semver).

## Публикация bundles

Публикация bundle осуществляется с помощью команды [werf-bundle-publish]({{ "/documentation/reference/cli/werf_bundle_apply.html" | true_relative_url }}). Команда принимает параметры по аналогии с [werf-converge]({{ "/documentation/reference/cli/werf_converge.html" | true_relative_url }}) и **требует запуска в git проекта**.

Во время работы werf-bundle-publish:
 - будут собраны недостающие образы в container registry (также как и в [werf-converge]({{ "/documentation/reference/cli/werf_converge.html" | true_relative_url }}));
 - будут опубликован [helm chart]({{ "/documentation/advanced/helm/configuration/chart.html" | true_relative_url }}) проекта в container registry;
 - werf запомнит в опубликованном bundle переданные параметры [values для helm chart]({{ "/documentation/advanced/helm/configuration/values.html" | true_relative_url }});
 - werf запомнит в опубликованном bundle переданные параметры [аннотаций и лейблов для ресурсов]({{ "/documentation/advanced/helm/deploy_process/annotating_and_labeling.html" | true_relative_url }})

### Версионирование при публикации

При публикации bundle пользователь может выбрать версию публикуемого bundle с помощью параметра `--tag`. По умолчанию bundle публикуется с тегом `latest`.

При публикации bundle по уже существующему тегу — bundle будет переопубликован с обновлениями в container registry. Данная возможность может быть использована для организации автоматических обновлений приложения при наличии новой версии в container registry. Подробнее см. [автообновление bundles](#автообновление-bundles).

## Выкат bundles

Bundle, опубликованный в container registry, можно выкатить в kubernetes с помощью werf.

Выкат опубликованного bundle осуществляется с помощью команды [werf-bundle-apply]({{ "/documentation/reference/cli/werf_bundle_apply.html" | true_relative_url }}). Команда **не требует доступа к git** (bundle содержит все необходимые файлы и образы для выката приложения), и принимает параметры по аналогии с [werf-converge]({{ "/documentation/reference/cli/werf_converge.html" | true_relative_url }}).

Переданные [values для helm chart]({{ "/documentation/advanced/helm/configuration/values.html" | true_relative_url }}), [аннотаций и лейблов для ресурсов]({{ "/documentation/advanced/helm/deploy_process/annotating_and_labeling.html" | true_relative_url }}) будут объединены с теми values, аннотациями и лейблами, которые были указаны при публикации bundle и зашиты в него.

### Версионирование при выкате

При выкате bundle пользователь может выбрать версию выкатываемого bundle с помощью параметра `--tag`. По умолчанию выкатывается bundle по тегу `latest`.

При выкате bundle werf проверит наличие обновлёния в container registry для указанного тега (или `latest`) и обновит приложение в кластере до последней версии bundle. Данная возможность может быть использована для организации автоматических обновлений приложения при наличии новой версии в container registry. Подробнее см. [автообновление bundles](#автообновление-bundles).

## Другие команды для работы с bundles

### Экспорт bundle в директорию

Команда [werf-bundle-export]({{ "/documentation/reference/cli/werf_bundle_export.html" | true_relative_url }}) позволяет создать директорию чарта в том виде, в котором он будет опубликован в container registry.

Команда работает по аналогии с командой [werf-bundle-publish]({{ "/documentation/reference/cli/werf_bundle_publish.html" | true_relative_url }}) имеет те же параметры и **требует запуска в git проекта**.

Данную команду можно использовать, чтобы узнать из чего состоит публикуемый bundle и для дебага.

### Скачивание опубликованного bundle

Команда [werf-bundle-download]({{ "/documentation/reference/cli/werf_bundle_download.html" | true_relative_url }}) позволяет скачать ранее опубликованный bundle в директорию.

Команда работает по аналогии с командой [werf-bundle-apply]({{ "/documentation/reference/cli/werf_bundle_apply.html" | true_relative_url }}) имеет те же параметры и **не требует доступа к git**.

Данную команду можно использовать, чтобы узнать из чего состоит опубликованный bundle и для дебага.

## Примеры использования

Опубликуем bundle для приложения по определённой версии, запускаем в git директории проекта:

```
git checkout v3.5.0
werf bundle publish --repo registry.mydomain.io/project --tag v3.5.0 --add-label "project-version=v3.5.0"
```

Опубликуем bundle для приложения по статичному тегу `main`, запускаем в git директории проекта:

```
git checkout main
werf bundle publish --repo registry.mydomain.io/project --tag main --set "myvalue.x=150"
```

Выкатим bundle опубликованный по тегу `v3.4.9`, запускаем на произвольном хосте с доступом в kubernetes и container registry:

```
werf bundle apply --repo registry.mydomain.io/project --tag v3.4.9 --env production --set "sentry.url=sentry.mydomain.io"
```

## Автообновление bundles

Werf поддерживает автоматический перевыкат bundle, опубликованного по статичному тегу вроде `latest` или `main`. Если публикация bundle настроена на выкат такого тега, то при каждом новом вызове публикации версия bundle по этому тегу в container regsitry будет обновлена.

Со стороны выката этого bundle при каждом вызове команды выката bundle будет использоваться последняя актуальная версия по указанному тегу.

### Выкат по маске semver

На данный момент не поддерживается, [но планируется добавить](https://github.com/werf/werf/issues/3169) возможность выката bundle по маске semver:

```
werf bundle apply --repo registry.mydomain.io/project --tag-mask v3.5.*
```

Данный вызов выката bundle должен проверить наличие версий по указанной маске `v3.5.*`, выбрать последнюю версию в рамках `v3.5` и выкатить её.

## Поддерживаемые container registry

| Container registry | Поддержка в bundles  |
| ------------------ | -------------------- |
| docker registry    | поддерживается |
| aws ecr            | поддерживается |
| azure cr           | поддерживается |
| docker hub         | [не поддерживается](https://github.com/werf/werf/issues/3184) |
| gcr                | поддерживается |
| gitlab registry    | поддерживается |
| github packages    | [поддерживается только для новой имплементации](https://github.com/werf/werf/issues/3188) |
| harbor             | поддерживается | 
| quay               | [не поддерживается](https://github.com/werf/werf/issues/3182) |
