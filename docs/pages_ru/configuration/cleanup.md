---
title: Очистка
sidebar: documentation
permalink: documentation/configuration/cleanup.html
author: Alexey Igrychev <alexey.igrychev@flant.com>
---

## Конфигурация политик очистки

Доступно два типа политик, для git-веток и git-тегов:

```yaml
cleanup:
  policies:
  - branch: <string|REGEXP>
    refsToKeepImagesIn:
      last: <uint>
      modifiedIn: <duration string>
      operator: <And|Or>
    imageDepthToKeep: <uint>
  - tag: <string|REGEXP>
    refsToKeepImagesIn:
      last: <uint>
      modifiedIn: <duration string>
      operator: <And|Or>
```

Каждая политика описывает набор references и, в случае с git-ветками, максимальное количество образов, которое сохраняется при очистке. Для git-тегов искомый образ всегда один. 

При описании набора references пользователь может указывать определённое имя или специфичную группу, используя [синтаксис регулярных выражений golang](https://golang.org/pkg/regexp/syntax/#hdr-Syntax):

```yaml
- tag: v1.1.1
- tag: /^v.*$/
- branch: master
- branch: /^master|production$/
```

> При очистке используются origin remote references для git-веток, но при написании конфигурации `origin/` опускается 

Для использования актуального множества references при очистке, основываясь на времени создания git-тега или активности в git-ветке, можно воспользоваться группой параметров `refsToKeepImagesIn`:

```yaml
- branch: /^features\/.*/
  refsToKeepImagesIn:
    last: 10
    modifiedIn: 168h
    operator: And
  imageDepthToKeep: 5
``` 

В примере описывается выборка из не более чем 10 последних веток с префиксом `features/` в имени, в которых была какая-либо активность за последнюю неделю. Для каждой ветки из этого набора по условию необходимо оставить по 5 образов.

- Параметр `last` позволяет выбирать последние `n` references из множества, определённого в `branch`/`tag`.
- Параметр `modifiedIn` (синтаксис `duration string` доступен в [документации](https://golang.org/pkg/time/#ParseDuration)) позволяет выбирать git-теги, которые были созданы в указанный период, или git-ветки с активностью в рамках периода. Также для определённого множества `branch`/`tag`.
- Параметр `operator: And|Or` определяет какие references будут результатом политики, те которые удовлетворяют оба условия или любое из них (`Or` по умолчанию).

При описании группы политик необходимо идти от общего к частному. Другими словами, `imageDepthToKeep` для конкретного reference будет соответствовать последней политике, под которую он подпадает:

```yaml
- branch: /.*/
  imageDepthToKeep: 1
- branch: master
  imageDepthToKeep: 5
```

В данном случае, для reference master справедливы обе политики и при сканировании ветки `imageDepthToKeep` будет равен 5.

> Так как результат сканирование git истории напрямую зависит от актуальности локального репозитория, не забудьте синхонизировать его с удалённым репозиторием вручную или воспользуйтесь опцией `--git-history-synchronization`, которая по умолчанию включена при запуске в CI системах

## Политики по умолчанию

В случае, если пользователь не использует политики очистки, применяются политики по умолчанию, соответствующие следующей конфигурации:

```yaml
cleanup:
  policies:
  - tag: /.*/
    refsToKeepImagesIn:
      last: 10
  - branch: /.*/
    refsToKeepImagesIn:
      last: 10
      modifiedIn: 168h
      operator: And
    imageDepthToKeep: 2
  - branch: /^(master|staging|production)$/
    imageDepthToKeep: 10
``` 

Разберём каждую политику по отдельности:

1. Сохранять образ для 10-ти последних тегов (по дате создания).
2. Сохранять по два образа для не более 10-ти веток с активностью за последнюю неделю. 
3. Сохранять по 10 образов для веток master, staging и production. 
