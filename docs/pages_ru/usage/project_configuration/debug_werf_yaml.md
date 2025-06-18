---
title: Отладка werf.yaml
permalink: usage/project_configuration/debug_werf_yaml.html
---

## Отображение результата шаблонизации `werf.yaml`

Команда `werf config render` отображает результат шаблонизации `werf.yaml`. 

```bash
$ werf config render
project: demo-app
configVersion: 1
---
image: backend
dockerfile: backend.Dockerfile
...
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

Команда `werf config graph` строит граф зависимостей между образами.

```bash
$ werf config graph
- image: backend
- image: frontend
```

{% include pages/ru/debug_template_flag.md.liquid %}