---
title: werf dimg mrproper
sidebar: reference
permalink: reference/cli/dimg_mrproper.html
---

### werf dimg mrproper
Очистить docker в соответствии с переданным параметром. Команда не ограничивается каким-либо проектом, а влияет на образы **всех проектов**.

```
werf dimg mrproper [options]
```

#### `--all`
Удалить docker-образы и docker-контейнеры **всех проектов**, связанные с werf.

#### `--improper-dev-mode-cache`
Удалить docker-образы и docker-контейнеры всех проектов, собранные в dev-режиме и связанные с werf.

#### `--improper-cache-version-stages`
Удалить устаревший кэш приложений, т.е. кэш, версия которого не соответствует версии кэша запущенного werf, во всех проектах.

#### Примеры

##### Запустить очистку
```bash
$ werf dimg mrproper --all
```

##### Посмотреть, версия кэша каких образов устарела, какие команды могут быть выполнены:
```bash
$ werf dimg mrproper --improper-cache-version-stages --dry-run
mrproper
  proper cache
    docker rmi dimgstage-werf-test-project-services-stats:ba95ec8a00638ddac413a13e303715dd2c93b80295c832af440c04a46f3e8555 dimgstage-werf-test-project-services-stats:f53af70566ec23fb634800d159425da6e7e61937afa95e4ed8bf531f3503daa6
```
