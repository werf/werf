---
title: dapp dimg build
sidebar: reference
permalink: reference/cli/dapp_dimg_build.html
---

Собрать dimg-ы, удовлетворяющие хотя бы одному из **DIMG**-ов (по умолчанию *).

```
dapp dimg build [options] [DIMG ...]
```

### Опции
{% include_relative partial/option/ssh_option.md %}
{% include_relative partial/option/use_system_tar_option.md %}
{% include_relative partial/option/build_dir_option.md %}
{% include_relative partial/option/lock_option.md %}
{% include_relative partial/option/tmp_dir_prefix_option.md %}
{% include_relative partial/option/build_context_directory_option.md %}
{% include_relative partial/option/introspection_options.md %}
{% include_relative partial/option/common_options.md %}

### Примеры

#### Сборка в текущей директории
```bash
$ dapp dimg build
```

#### Сборка dimg-ей из соседней директории
```bash
$ dapp dimg build --dir ../project
```

#### Запуск вхолостую с выводом процесса сборки
```bash
$ dapp dimg build --dry-run
```

#### Выполнить сборку, а в случае ошибки, предоставить образ для тестирования
```bash
$ dapp dimg build --introspect-error
```

#### Выполнить сборку, а в случае ошибки, сохранить контекст сборки в директории 'context'
```bash
$ dapp dimg build --build-context-directory context
```
