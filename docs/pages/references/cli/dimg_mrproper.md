---
title: dapp dimg mrproper
sidebar: reference
permalink: dimg_mrproper.html
---

### dapp dimg mrproper
Очистить docker в соответствии с переданным параметром. Команда не ограничивается каким-либо проектом, а влияет на образы **всех проектов**.

```
dapp dimg mrproper [options]
```

#### `--all`
Удалить docker-образы и docker-контейнеры **всех проектов**, связанные с dapp.

#### `--improper-dev-mode-cache`
Удалить docker-образы и docker-контейнеры всех проектов, собранные в dev-режиме и связанные с dapp.

#### `--improper-cache-version-stages`
Удалить устаревший кэш приложений, т.е. кэш, версия которого не соответствует версии кэша запущенного dapp, во всех проектах.

#### Примеры

##### Запустить очистку
```bash
$ dapp dimg mrproper --all
```

##### Посмотреть, версия кэша каких образов устарела, какие команды могут быть выполнены:
```bash
$ dapp dimg mrproper --improper-cache-version-stages --dry-run
mrproper
  proper cache
    docker rmi dimgstage-dapp-test-project-services-stats:ba95ec8a00638ddac413a13e303715dd2c93b80295c832af440c04a46f3e8555 dimgstage-dapp-test-project-services-stats:f53af70566ec23fb634800d159425da6e7e61937afa95e4ed8bf531f3503daa6
```
