---
title: dapp dimg build-context export
sidebar: reference
permalink: reference/cli/dapp_dimg_build_context_export.html
---

Экспортировать контекст, кэш приложений и директорию сборки.

```
dapp dimg build-context export [options] [DIMG ...]
```

### Опции
{% include_relative partial/option/build_context_directory_option.md %}
{% include_relative partial/option/common_options.md %}

### Пример

#### Экспортировать контекст приложения 'frontend' из проекта 'project' в директорию 'context'

```bash
$ dapp dimg build-context export --dir ~/workspace/project --build-context-directory context frontend
```
