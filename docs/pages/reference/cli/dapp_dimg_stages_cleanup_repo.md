---
title: dapp dimg stages cleanup repo
sidebar: reference
permalink: reference/cli/dapp_dimg_stages_cleanup_repo.html
---

Удалить неиспользуемый кэш приложений в репозитории **REPO**.

```
dapp dimg stages cleanup repo [options] REPO
```

### `--improper-repo-cache`
Удалить кэш, который не использовался при сборке приложений в репозитории **REPO**.

### `--improper-git-commit`
Удалить кэш, связанный с отсутствующими в git-репозиториях коммитами.

### `--improper-cache-version`
Удалить устаревший кэш приложений проекта.

### Примеры

#### Удалить неактуальный кэш в репозитории localhost:5000/test, где:

* Версия кэша не соответствует текущей.
* Не связан с приложениями в репозитории.
* Собран из коммитов, которые в данный момент отсутствуют в git-репозиториях проекта.

```bash
$ dapp dimg stages cleanup repo localhost:5000/test --improper-cache-version --improper-repo-cache --improper-git-commit
```
