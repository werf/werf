---
title: werf dimg flush
sidebar: not_used
permalink: not_used/cli/dimg_flush.html
---

### werf dimg flush local
Удалить все теги приложений проекта локально.

```
werf dimg flush local [options] [DIMG ...]
```

#### `--with-stages`
Соответствует вызову команды `werf dimg stages flush local`.

### werf dimg flush repo
Удалить все теги приложений проекта.
```
werf dimg flush repo [options] [DIMG ...] REPO
```

#### `--with-stages`
Соответствует вызову команды `werf dimg stages cleanup flush`.
