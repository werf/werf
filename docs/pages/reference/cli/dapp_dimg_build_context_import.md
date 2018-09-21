---
title: dapp dimg build-context import
sidebar: reference
permalink: reference/cli/dapp_dimg_build_context_import.html
---

Импортировать контекст, кэш приложений и директорию сборки.

```
dapp dimg build-context import [options]
```

### Опции
{% include_relative partial/option/build_context_directory_option.md %}
{% include_relative partial/option/common_options.md %}

### Пример

#### Импортировать контекст приложений проекта 'project' из директории 'context'

```bash
$ dapp dimg build-context import --dir ~/workspace/project --build-context-directory context
```
