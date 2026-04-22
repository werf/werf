---
title: TypeScript-шаблоны
permalink: usage/deploy/typescript.html
---

> **Обратите внимание**: TypeScript-шаблоны — экспериментальная функция. Для включения установите переменную окружения `NELM_FEAT_TYPESCRIPT=true`.

## Обзор

Помимо [Helm-шаблонов]({{ "/usage/deploy/templates.html" | true_relative_url }}), werf поддерживает генерацию Kubernetes-манифестов с помощью TypeScript. Helm-шаблоны и TypeScript-шаблоны могут сосуществовать в одном чарте — результирующие манифесты объединяются в единый мульти-документный YAML.

TypeScript-шаблоны работают из коробки: развёртывание чарта с директорией `ts/` не требует дополнительных инструментов или настройки — werf автоматически скачивает рантайм Deno TypeScript и рендерит TypeScript-шаблоны.

### Зачем TypeScript

Язык шаблонов Helm хорошо работает для простых случаев, но с ростом сложности чарта становится сложным в поддержке: примитивный язык с большим количеством подводных камней, !ограниченные библиотеки!, проблемы с производительностью, сложная отладка, слабая поддержка в IDE и редакторах. TypeScript в werf решает эти проблемы, не усложняя процесс развёртывания.

### Возможности

- Поддержка IDE — автокомплит, проверка типов, go-to-definition и рефакторинг в любом редакторе с поддержкой Deno/TypeScript (VS Code, JetBrains, Neovim и др.).
- Стандартный синтаксис — обычные функции, циклы и условия вместо неудобных конструкций шаблонизатора.
- Чистый TypeScript — директория `ts` является обычным Deno TypeScript-проектом и может рендериться без werf, с помощью одного лишь [рантайма Deno TypeScript](https://deno.com/).
- Большая экосистема — TypeScript один из самых популярных языков с обширной документацией, ресурсами сообщества и инструментарием.
- Возможность использовать практически любую стороннюю TypeScript/JavaScript-библиотеку, например [kubernetes-models](https://github.com/tommy351/kubernetes-models-ts), [cdk8s](https://cdk8s.io/) или любую другую из экосистем npm/Deno.
- Тестирование — тестируйте код с помощью привычных TypeScript-библиотек и инструментов.
- Никаких дополнительных требований к хосту — для развёртывания TypeScript-чарта достаточно только werf. Не нужно устанавливать Node, Deno, npm, модули npm или что-либо ещё. Мы берём это на себя — просто выполните `werf converge`.
- Изолированные окружения — модули npm по умолчанию бандлируются в чарт, а рантайм Deno может предоставляться хостовой системой, так что во время развёртывания не будет сетевых обращений, кроме обращений к самому Kubernetes.
- Безопасность — код выполняется в изолированной песочнице Deno без доступа к сети, переменным окружения и запуску процессов. Доступ к файловой системе ограничен чтением файлов чарта.

## Быстрый старт

Инициализация TypeScript-файлов в существующем чарте:

```shell
werf chart ts init
```

Команда создаст директорию `.helm/ts/` с готовым скелетом TypeScript-проекта и несколькими файлами с примерами ресурсов. Попробуйте отредактировать `ts/src/deployment.ts` — например, изменить количество реплик — и проверьте результат:

```shell
werf render --dev
```

Для развёртывания:

```shell
werf converge --dev
```

## Структура чарта

{% tree_file_viewer 'examples/ts/example-chart' default_file='ts/src/index.ts' expanded=true %}

## Разработка чарта с TypeScript-шаблонами

Установите [Deno](https://docs.deno.com/runtime/getting_started/installation/) и следуйте [руководству по настройке](https://docs.deno.com/runtime/getting_started/setup_your_environment/) для вашей IDE или редактора (VS Code, JetBrains, Neovim и др.).

Инициализируйте TypeScript-файлы в чарте, если они ещё не инициализированы:

```shell
werf chart ts init
```

Откройте директорию `ts/` в редакторе как обычный Deno/TypeScript-проект. Работайте с ним так же, как с любой TypeScript-кодовой базой — запускайте скрипты, пишите тесты, используйте отладчик. Deno предоставляет богатый набор инструментов для тестирования, линтинга, форматирования и многого другого. Подробнее см. [документацию Deno](https://docs.deno.com/runtime/).

Структуру кодовой базы можно организовать по своему усмотрению. Единственное требование — файл `ts/src/index.ts` должен существовать, и функция `render` из `@nelm/chart-ts-sdk` **обязательно** должна быть вызвана. Иначе рендеринг TypeScript не произойдёт.

Для отладки рендеринга шаблонов в окружении, максимально близком к тому, как werf запускает Deno, используйте задачу `dev` из `ts/deno.json`:

```shell
cd .helm/ts
deno task dev
```
TypeScript-движок вызовет функцию `render` из `ts/src/index.ts` с примером контекста из `ts/input.example.yaml`. Результирующий YAML будет выведен в консоль под сообщением `Rendered manifests:`.

Устанавливайте библиотеки с помощью `deno add`, например попробуйте установить [kubernetes-models](https://github.com/tommy351/kubernetes-models-ts) — библиотеку для строгой типизации Kubernetes-ресурсов:

```shell
deno add npm:kubernetes-models
```

Зависимость добавится в `deno.json` автоматически. Теперь можно импортировать и использовать её:

```typescript
// .helm/ts/src/deployment.ts:
import { Deployment } from 'kubernetes-models/apps/v1';

export function newDeployment($: WerfRenderContext): object {
  return new Deployment({
    metadata: { name: 'myapp' },
    spec: {
      // other fields
    },
  }).toJSON();
}
```

Чтобы убедиться, что всё работает с рантаймом Deno из werf, выполните:
```shell
werf lint --dev
```

```shell
werf render --dev
```

## Как развернуть чарт с TypeScript-шаблонами

Просто запустите `werf converge`: бинарный файл Deno будет скачан в кеш, TypeScript-шаблоны будут отрендерены и развёрнуты.

> **Обратите внимание**: Согласно [политикам гиттерминизма]({{ "/usage/project_configuration/giterminism.html" | true_relative_url }}), все изменённые файлы должны быть закоммичены.

## Развёртывание в изолированных окружениях

Для изолированных окружений, где Deno не может быть скачан автоматически:

1. Опубликуйте чарт:
   ```shell
   werf bundle publish --repo example.org/mycompany/myapp
   ```
   Все модули npm будут минифицированы и включены в бандл, так что чарт можно установить даже без доступа к интернету.

2. На целевой машине в изолированном окружении (без доступа к сети) скачайте Deno вручную и выполните:
   ```shell
   werf bundle apply --repo example.org/mycompany/myapp --deno-binary-path /usr/local/bin/deno
   ```
   Где `/usr/local/bin/deno` — путь к локальному бинарному файлу Deno. TypeScript-шаблоны будут отрендерены и развёрнуты с использованием предварительно скомпилированных файлов из пакета чарта.

## Обзор SDK API

TypeScript-движок использует пакет [@nelm/chart-ts-sdk](https://github.com/werf/nelm-chart-ts-sdk).

### Функции "render" и "generate"

`index.ts` обязан вызвать функцию `render()`. Функция `generate()`, которая непосредственно генерирует манифесты, должна быть передана в `render()` в качестве аргумента, например:

```typescript
// .helm/ts/src/index.ts:
await render(generate);
```

### Объект "WerfRenderContext"

Функция `generate` получает корневой контекст в переменной `$` типа `WerfRenderContext` — тот же контекст, что и в Helm-шаблонах:

| Поле | Тип | Описание |
|------|-----|----------|
| `$.Values` | `WerfServiceValues` | Параметры чарта + сервисные значения в `$.Values.global.werf` |
| `$.Release` | `Release` | Информация о релизе |
| `$.Chart` | `ChartMetadata` | Метаданные из Chart.yaml |
| `$.Capabilities` | `Capabilities` | Возможности кластера (API-версии, версия Kubernetes) |
| `$.Files` | `Record<string, Uint8Array>` | Исходные файлы чарта (кроме `templates/` и `ts/`) |

Пример контекста — в файле `ts/input.example.yaml`. Подробнее о параметрах и их формировании — в разделе [Параметризация шаблонов]({{ "/usage/deploy/values.html" | true_relative_url }}).

### Объект "RenderResult"

Функция `generate` возвращает `RenderResult` — объект с массивом `manifests`. Каждый элемент — обычный JavaScript-объект, представляющий Kubernetes-ресурс. Пример вывода:

```json
{
  "manifests": [
    {
      "apiVersion": "apps/v1",
      "kind": "Deployment",
      "metadata": { "name": "myapp" },
      "spec": { "..." }
    },
    {
      "apiVersion": "v1",
      "kind": "Service",
      "metadata": { "name": "myapp" },
      "spec": { "..." }
    }
  ]
}
```

Каждый объект сериализуется в YAML и включается в итоговый результат рендеринга.
