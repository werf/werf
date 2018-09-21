---
title: dapp dimg stages cleanup local
sidebar: reference
permalink: reference/cli/dapp_dimg_stages_cleanup_local.html
---

Удалить неактуальный локальный кэш приложений проекта.

```
dapp dimg stages cleanup local [options] [REPO]
```

### `--improper-repo-cache`
Удалить кэш, который не использовался при сборке приложений в репозитории **REPO**.

### `--improper-git-commit`
Удалить кэш, связанный с отсутствующими в git-репозиториях коммитами.

### `--improper-cache-version`
Удалить устаревший кэш приложений проекта.

### Примеры

#### Оставить только актуальный кэш, исходя из приложений в localhost:5000/test
```bash
$ dapp dimg stages cleanup local --improper-repo-cache localhost:5000/test
```

##### Удалить кэш, версия которого не совпадает с текущей
```bash
$ dapp dimg stages cleanup local --improper-cache-version localhost:5000/test
```

##### Почистить кэш после rebase в одном из связанных git-репозиториев
```bash
$ dapp dimg stages cleanup local --improper-git-commit localhost:5000/test
```

