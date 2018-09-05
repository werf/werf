---
title: dapp dimg build-context
sidebar: reference
permalink: dimg_build_context.html
---


### dapp dimg build-context export
Экспортировать контекст, кэш приложений и директорию сборки.

```
dapp dimg build-context export [options] [DIMG ...]
```

#### `--build-context-directory DIR_PATH`
Определить директорию сохранения контекста.

#### Пример

##### Экспортировать контекст приложения 'frontend' из проекта 'project' в директорию 'context'

```bash
$ dapp dimg build-context export --dir ~/workspace/project --build-context-directory context frontend
```

### dapp dimg build-context import
Импортировать контекст, кэш приложений и директорию сборки.

```
dapp dimg build-context import [options]
```

#### `--build-context-directory DIR_PATH`
Определить директорию хранения контекста.

#### Пример

##### Импортировать контекст приложений проекта 'project' из директории 'context'

```bash
$ dapp dimg build-context import --dir ~/workspace/project --build-context-directory context
```
