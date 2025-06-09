---
title: Debug werf.yaml
permalink: /usage/project_configuration/debug_werf_yaml.html
---

## Обзор

Файл `werf.yaml` может быть параметризован с помощью Go templates, что делает его мощным, но сложным в отладке. Werf предоставляет несколько команд для просмотра конечной конфигурации проекта и построения зависимостей: `render`, `list`, `graph`. Подробнее про шаблонизатор можно ознакомится [тут]({{"usage/project_configuration/werf_yaml_template_engine.html" | true_relative_url }})

[Пример](https://werf.test.flant.com/getting_started/), который будет использоваться

## `werf config render`

Показывает итоговую конфигурацию `werf.yaml` после рендеринга шаблонов.

### Пример команды

```bash
werf config render
```
### Пример вывода:

```yaml
project: demo-app
configVersion: 1

---
image: backend
dockerfile: backend.Dockerfile

---
image: frontend
dockerfile: frontend.Dockerfile
```

### Когда полезно

* Проверка, что шаблоны Go корректно отрабатывают;
* Удостовериться, что значения переменных заданы как ожидалось;
* Отладка `dependencies` и других вложенных структур.

## `werf config list`

Выводит список всех образов, определённых в итоговом `werf.yaml`.

### Пример команды

```bash
werf config list
```

### Пример вывода:

```bash
backend
frontend
```

### Когда полезно

* Быстрая проверка всех образов, которые будут собираться;
* Диагностика генерации имён при использовании шаблонов.

## `werf config graph`

Строит граф зависимостей между образами.

### Пример команды

```bash
werf config graph
```

### Пример вывода:

```yaml
- image: backend
- image: frontend
```

### Когда полезно

* Проверка, что зависимости подключены корректно;
* Отладка сложных схем сборки с `dependencies`.

<!-- ## debug-templates flag

Включает режим отладки шаблонов, который позволяет просматривать отрендеренные шаблоны в файле werf.yaml (по умолчанию — значение переменной окружения $WERF_DEBUG_TEMPLATES или false).

### Пример команды

```bash
werf config render
```
-->
