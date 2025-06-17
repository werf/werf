---
title: Отладка `werf.yaml`
permalink: usage/project_configuration/debug_werf_yaml.html
---

## Проверка итоговой конфигурации `werf.yaml`

Команда `werf config render` отображает итоговое содержимое `werf.yaml` после рендеринга всех шаблонов и подстановки переменных окружения. 

```bash
$ werf config render
project: demo-app
configVersion: 1
---
image: backend
dockerfile: backend.Dockerfile
...
```

## Получение списка образов, описанных в `werf.yaml`

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