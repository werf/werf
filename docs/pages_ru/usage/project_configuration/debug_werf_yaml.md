---
title: Отладка werf.yaml
permalink: usage/project_configuration/debug_werf_yaml.html
---

## Отображение результата шаблонизации

Команда `werf config render` отображает результат шаблонизации `werf.yaml` или конкретного образа/ов. 

```bash
$ werf config render
project: demo-app
configVersion: 1
---
image: backend
dockerfile: backend.Dockerfile
```

```bash
$ werf config render backend
image: backend
dockerfile: backend.Dockerfile
```

## Получить имена всех собираемых образов

Команда `werf config list` выводит список всех образов, определённых в итоговом `werf.yaml`. 

```bash
$ werf config list
backend
frontend
```

Флаг `--final-images-only` выведет список только конечных образов. Подробнее ознакомится с конечными и промежуточными образами можно [здесь]({{ "usage/build/images.html#использование-промежуточных-и-конечных-образов" | true_relative_url }})

## Анализ зависимостей между образами

Команда `werf config graph` строит граф зависимостей между образами (или конкретных образов).

```bash
$ werf config graph
- image: images/argocd
  dependsOn:
    dependencies:
    - images/argocd-source
- image: images/argocd-operator
  dependsOn:
    from: common/distroless
    import:
    - images/argocd-operator-artifact
- image: images/argocd-operator-artifact
- image: images/argocd-artifact
- image: images/argocd-source
  dependsOn:
    import:
    - images/argocd-artifact
- image: common/distroless-artifact
- image: common/distroless
  dependsOn:
    import:
    - common/distroless-artifact
- image: images-digests
  dependsOn:
    dependencies:
    - images/argocd-operator
    - images/argocd
    - common/distroless
- image: python-dependencies
- image: bundle
  dependsOn:
    import:
    - images-digests
    - python-dependencies
- image: release-channel-version-artifact
- image: release-channel-version
  dependsOn:
    import:
    - release-channel-version-artifact
```

```bash
$ werf config graph images-digests
- image: images-digests
  dependsOn:
    dependencies:
    - images/argocd-operator
    - images/argocd
    - common/distroless
```

{% include pages/ru/debug_template_flag.md.liquid %}
