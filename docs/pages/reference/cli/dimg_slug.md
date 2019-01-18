---
title: werf slug
sidebar: reference
permalink: reference/cli/dimg_slug.html
---


### werf slug
"Слагифицирует" строку (применяется алгоритм slug), добавляет хэш и выводит результат. Слагификация применяется например при указании тегов `--tag`, `--tag-slug` при запуске команд `werf dimg bp`, `werf dimg push`, `werf dimg push`, `werf kube deploy`  и других.

```
werf slug STRING
```

#### Пример

```bash
$ werf slug 'Длинный, mixed.language tag with sP3c!AL chars'
dlinnyj-mixed-language-tag-with-sp3cal-chars-ae959974
```
