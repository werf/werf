---
title: Обзор
permalink: usage/project_configuration/overview.html
---

werf следует принципам подхода IaC (Infrastructure as Code) и стимулирует пользователя хранить конфигурацию доставки проекта вместе с кодом приложения в Git и осознанно использовать внешние зависимости. Отвечает за это механизм под названием [гитерминизм]({{ "usage/project_configuration/giterminism.html" | true_relative_url }}).

Типовая конфигурация проекта состоит из нескольких файлов:

- werf.yaml;
- одного или нескольких Dockerfile-файлов;
- Helm-чарта.

## werf.yaml

Главный конфигурационный файл проекта в werf. Его основное предназначение — связывание инструкций для сборки и развёртывания.

### Инструкции для сборки

Определяются для каждого компонента приложения. Могут быть представлены в двух форматах:
- Dockerfile-файлы, описывающие образы проекта.
- Stapel — альтернативный синтаксис для сборки.

> Подробнее по конфигурации сборки смотрите [в разделе «Сборка»]({{ "usage/build/overview.html" | true_relative_url }}).

### Инструкции развёртывания

Определяются для всего приложения (и всех окружений развёртывания) и должны быть представлены в виде Helm-чарта.

> Подробнее по конфигурации развёртывания смотрите [в разделе «Развёртывание»]({{ "usage/deploy/overview.html" | true_relative_url }}).

## Пример типовой конфигурации проекта

```yaml
# werf.yaml
project: app
configVersion: 1
---
image: backend
context: backend
dockerfile: Dockerfile
---
image: frontend
context: frontend
dockerfile: Dockerfile
```

```shell
$ tree -a
.
├── .helm
│   ├── templates
│   │   ├── NOTES.txt
│   │   ├── _helpers.tpl
│   │   ├── deployment.yaml
│   │   ├── hpa.yaml
│   │   ├── ingress.yaml
│   │   ├── service.yaml
│   │   └── serviceaccount.yaml
│   └── values.yaml
├── backend
│   ├── Dockerfile
│   └── ...
├── frontend
│   ├── Dockerfile
│   └── ...
└── werf.yaml
```
