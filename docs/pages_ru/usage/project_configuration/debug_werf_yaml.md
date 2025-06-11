---
title: Отладка werf.yaml
permalink: /usage/project_configuration/debug_werf_yaml.html
---

## Обзор

Файл `werf.yaml` может быть параметризован с помощью Go templates. Werf предоставляет несколько команд для просмотра конечной конфигурации проекта и построения зависимостей: `render`, `list`, `graph`. Подробнее про шаблонизатор можно ознакомится [тут]({{"usage/project_configuration/werf_yaml_template_engine.html" | true_relative_url }})

[Пример](https://werf.io/getting_started/#build-and-deploy-with-werf), который будет использоваться

## `werf config render`
Показывает итоговую конфигурацию `werf.yaml` после рендеринга шаблонов, например:
```bash
$ werf config render
project: demo-app
configVersion: 1
---
image: backend
dockerfile: backend.Dockerfile
...
```

## `werf config list`

Выводит список всех образов, определённых в итоговом `werf.yaml`.

```bash
$ werf config list
backend
frontend
```

## `werf config graph`

Строит граф зависимостей между образами.

```bash
$ werf config graph
- image: backend
- image: frontend
```

{% include pages/ru/debug_template_flag.md.liquid %}