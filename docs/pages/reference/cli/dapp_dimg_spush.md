---
title: dapp dimg spush
sidebar: reference
permalink: reference/cli/dapp_dimg_spush.html
---

Выкатить собранный dimg в docker registry, в формате **REPO**:**TAG**.

```
dapp dimg spush [options] [DIMG] REPO
```

Опции - такие же как у [**dapp dimg push**](#dapp-dimg-push).

`spush` - сокращение от `simple push`, и отличие `spush` от `dapp dimg push` в том, что при использовании `spush` пушится только один образ, и формат его имени в registry - простой (**REPO**:**TAG**). Т.о., если в dappfile описано несколько образов, то указание образа при вызове `dapp dimg spush` является необходимым.

### Примеры

#### Выкатить собранный dimg **app** в репозиторий test и тегом latest
```bash
$ dapp dimg spush app localhost:5000/test
```

#### Выкатить собранный dimg с произвольными тегами
```bash
$ dapp dimg spush app localhost:5000/test --tag 1 --tag test
```

#### Посмотреть, какие образы могут быть добавлены в репозиторий
```bash
$ dapp dimg spush app localhost:5000/test --tag-commit --tag-branch --dry-run
localhost:5000/test:2c622c16c39d4938dcdf7f5c08f7ed4efa8384c4
localhost:5000/test:master
```
