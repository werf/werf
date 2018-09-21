---
title: dapp dimg cleanup
sidebar: reference
permalink: reference/cli/dapp_dimg_cleanup.html
---
Убраться в системе после некорректного завершения работы dapp, удалить нетегированные docker-образы и docker-контейнеры.

```
dapp dimg cleanup [options]
```

### Примеры

#### Запустить
```bash
$ dapp dimg cleanup
```

#### Посмотреть, какие команды могут быть выполнены
```bash
$ dapp dimg cleanup --dry-run
backend
  docker rm -f dd4ec7v33
  docker rmi dimgstage-dapp-test-project:07758b3ec8aec701a01 dimgstage-dapp-test-project:ec701a0107758b3ec8a
```
