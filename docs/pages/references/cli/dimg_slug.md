---
title: Превращение строки в уникальный ключ
sidebar: reference
permalink: dimg_slug.html
folder: command
---


### dapp slug
"Слагифицирует" строку (применяется алгоритм slug), добавляет хэш и выводит результат. Слагификация применяется например при указании тегов `--tag`, `--tag-slug` при запуске команд `dapp dimg bp`, `dapp dimg push`, `dapp dimg push`, `dapp kube deploy`  и других.

```
dapp slug STRING
```

#### Пример

```bash
$ $ dapp slug 'Длинный, mixed.language tag with sP3c!AL chars'
dlinnyj-mixed-language-tag-with-sp3cal-chars-ae959974
```
