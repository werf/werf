---
title: dapp dimg push
sidebar: reference
permalink: reference/cli/dapp_dimg_push.html
---

Выкатить собранные dimg-ы в docker registry, в формате **REPO**/**ИМЯ DIMG**:**TAG**. В случае, если **ИМЯ DIMG** отсутствует, формат следующий **REPO**:**TAG**.

```
dapp dimg push [options] [DIMG ...] REPO
```

{% include_relative partial/option/tag_options.md %}

#### `--with-stages`
Также выкатить кэш стадий.

{% include_relative partial/option/registry_login_options.md %}
{% include_relative partial/option/lock_option.md %}
{% include_relative partial/option/common_options.md %}

### Примеры

#### Выкатить все dimg-ы в репозиторий localhost:5000/test и тегом latest
```bash
$ dapp dimg push localhost:5000/test
```

#### Посмотреть, какие образы могут быть добавлены в репозиторий
```bash
$ dapp dimg push localhost:5000/test --tag yellow --tag-branch --dry-run
backend
  localhost:5000/test:backend-yellow
  localhost:5000/test:backend-master
frontend
  localhost:5000/test:frontend-yellow
  localhost:5000/test:frontend-0.2
```
