---
title: Очистка
sidebar: documentation
permalink: configuration/cleanup.html
author: Alexey Igrychev <alexey.igrychev@flant.com>
summary: |
  <div class="language-yaml highlighter-rouge"><div class="highlight"><pre class="highlight"><code><span class="na">cleanup</span><span class="pi">:</span>
    <span class="na">keepPolicies</span><span class="pi">:</span>
    <span class="pi">-</span> <span class="na">references</span><span class="pi">:</span>
        <span class="na">branch</span><span class="pi">:</span> <span class="s">&lt;string|/REGEXP/&gt;</span>
        <span class="na">tag</span><span class="pi">:</span> <span class="s">&lt;string|/REGEXP/&gt;</span>
        <span class="na">limit</span><span class="pi">:</span>
          <span class="na">last</span><span class="pi">:</span> <span class="s">&lt;int&gt;</span>
          <span class="na">in</span><span class="pi">:</span> <span class="s">&lt;duration string&gt;</span>
          <span class="na">operator</span><span class="pi">:</span> <span class="s">&lt;And|Or&gt;</span>
      <span class="na">imagesPerReference</span><span class="pi">:</span>
        <span class="na">last</span><span class="pi">:</span> <span class="s">&lt;int&gt;</span>
        <span class="na">in</span><span class="pi">:</span> <span class="s">&lt;duration string&gt;</span>
        <span class="na">operator</span><span class="pi">:</span> <span class="s">&lt;And|Or&gt;</span>
  </code></pre></div></div>  
---

## Конфигурация политик очистки

Конфигурация очистки состоит из набора политик, `keepPolicies`, по которым выполняется выборка значимых образов на основе истории git. Таким образом, в результате [очистки]({{ site.baseurl }}/reference/cleaning_process.html#алгоритм-работы-очистки-по-истории-git) __неудовлетворяющие политикам образы удаляются__.

Каждая политика состоит из двух частей: 
- `references` определяет множество references, git-тегов или git-веток, которые будут использоваться при сканировании.
- `imagesPerReference` определяет лимит искомых образов для каждого reference из множества.

Любая политика должна быть связана с множеством git-тегов (`tag: <string|/REGEXP/>`) либо git-веток (`branch: <string|/REGEXP/>`). Можно указать определённое имя reference или задать специфичную группу, используя [синтаксис регулярных выражений golang](https://golang.org/pkg/regexp/syntax/#hdr-Syntax).

```yaml
tag: v1.1.1
tag: /^v.*$/
branch: master
branch: /^(master|production)$/
```

> При сканировании описанный набор git-веток будет искаться среди origin remote references, но при написании конфигурации префикс `origin/` в названии веток опускается  

Заданное множество references можно лимитировать, основываясь на времени создания git-тега или активности в git-ветке. Группа параметров `limit` позволяет писать гибкие и эффективные политики под различные workflow.

```yaml
- references:
    branch: /^features\/.*/
    limit:
      last: 10
      in: 168h
      operator: And
``` 

В примере описывается выборка из не более чем 10 последних веток с префиксом `features/` в имени, в которых была какая-либо активность за последнюю неделю.

- Параметр `last: <int>` позволяет выбирать последние `n` references из определённого в `branch`/`tag` множества.
- Параметр `in: <duration string>` (синтаксис доступен в [документации](https://golang.org/pkg/time/#ParseDuration)) позволяет выбирать git-теги, которые были созданы в указанный период, или git-ветки с активностью в рамках периода. Также для определённого множества `branch`/`tag`.
- Параметр `operator: <And|Or>` определяет какие references будут результатом политики, те которые удовлетворяют оба условия или любое из них (`And` по умолчанию).

По умолчанию при сканировании reference количество искомых образов не ограничено, но поведение может настраиваться группой параметров `imagesPerReference`:

```yaml
imagesPerReference:
  last: <int>
  in: <duration string>
  operator: <And|Or>
```

- Параметр `last: <int>` определяет количество искомых образов для каждого reference. По умолчанию количество не ограничено (`-1`).
- Параметр `in: <duration string>` (синтаксис доступен в [документации](https://golang.org/pkg/time/#ParseDuration)) определяет период, в рамках которого необходимо выполнять поиск образов.
- Параметр `operator: <And|Or>` определяет какие образы сохранятся после применения политики, те которые удовлетворяют оба условия или любое из них (`And` по умолчанию)

> Для git-тегов проверяется только HEAD-коммит и значение `last` >1 не имеет никакого смысла, является невалидным

При описании группы политик необходимо идти от общего к частному. Другими словами, `imagesPerReference` для конкретного reference будет соответствовать последней политике, под которую он подпадает:

```yaml
- references:
    branch: /.*/
  imagesPerReference:
    last: 1
- references:
    branch: master
  imagesPerReference:
    last: 5
```

В данном случае, для reference _master_ справедливы обе политики и при сканировании ветки `last` будет равен 5.

## Политики по умолчанию

В случае, если в `werf.yaml` отсутствуют пользовательские политики очистки, используются политики по умолчанию, соответствующие следующей конфигурации:

```yaml
cleanup:
  keepPolicies:
  - references:
      tag: /.*/
      limit:
        last: 10
  - references:
      branch: /.*/
      limit:
        last: 10
        in: 168h
        operator: And
    imagesPerReference:
      last: 2
      in: 168h
      operator: And
  - references:  
      branch: /^(master|staging|production)$/
    imagesPerReference:
      last: 10
``` 

Разберём каждую политику по отдельности:

1. Сохранять образ для 10-ти последних тегов (по дате создания).
2. Сохранять по не более чем два образа, опубликованных за последнюю неделю, для не более 10-ти веток с активностью за последнюю неделю. 
3. Сохранять по 10 образов для веток master, staging и production. 
