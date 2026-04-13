---
title: TypeScript-шаблоны
permalink: usage/deploy/typescript.html
---

> Обратите внимание, что TypeScript-шаблоны — экспериментальная функция. Для включения установите переменную окружения `NELM_FEAT_TYPESCRIPT=true`.

## Особенности

Помимо стандартного способа описания Kubernetes-манифестов через [Helm-шаблоны]({{ "/usage/deploy/templates.html" | true_relative_url }}), werf поддерживает описание манифестов на TypeScript:

- типизация и автодополнение в IDE;
- подключение сторонних библиотек (например, [kubernetes-models](https://github.com/tommy351/kubernetes-models-ts) для строгой типизации Kubernetes-ресурсов);
- стандартный синтаксис — обычные функции, циклы и условия вместо конструкций шаблонизатора;
- тестирование — возможность покрывать логику генерации манифестов обычными тестами;
- безопасность — код выполняется в изолированной песочнице без доступа к сети, переменным окружения и файловой системе.

TypeScript-код получает тот же [корневой контекст]({{ "/usage/deploy/values.html" | true_relative_url }}), что и Helm-шаблоны (`Values`, `Release`, `Chart` и т. д.), и может сосуществовать с Helm-шаблонами в одном чарте.

## Быстрый старт

Инициализация TypeScript-файлов в существующем чарте:

```shell
werf chart ts init
```

В директории чарта появится поддиректория `ts/` с готовым примером. Отредактируйте сгенерированные файлы в `ts/src/` и запустите:

```shell
werf converge
```

Манифесты генерируются из директории `ts/` TypeScript-движком и из директории `templates/` Helm-шаблонизатором, после чего объединяются в один YAML-документ.

### Структура

```
.helm/
  templates/              # Helm-шаблоны (как обычно)
  ts/                     # TypeScript-шаблоны
    src/
      index.ts            # Точка входа
      helpers.ts          # Вспомогательные функции
      deployment.ts       # Генерация Deployment
      service.ts          # Генерация Service
    deno.json             # Конфигурация Deno и зависимости
    tsconfig.json         # Конфигурация TypeScript
  values.yaml
  Chart.yaml
```

werf ищет точку входа в порядке: `ts/src/index.ts`, затем `ts/src/index.js`. Если ни один из файлов не найден, TypeScript-рендеринг пропускается.

### Точка входа

```typescript
// .helm/ts/src/index.ts:
import { render, WerfRenderContext, RenderResult } from '@nelm/chart-ts-sdk';
import { newDeployment } from './deployment.ts';
import { newService } from './service.ts';

function generate($: WerfRenderContext): RenderResult {
  const manifests: object[] = [];

  manifests.push(newDeployment($));

  if ($.Values.service?.enabled !== false) {
    manifests.push(newService($));
  }

  return { manifests };
}

await render(generate);
```

Функция `generate` получает корневой контекст `$` типа `WerfRenderContext` и возвращает `RenderResult` с массивом манифестов — обычных JavaScript-объектов, которые будут сериализованы в YAML.

## Корневой контекст

TypeScript-код получает тот же контекст, что и Helm-шаблоны, через переменную `$` типа `WerfRenderContext`:

| Поле | Тип | Описание |
|------|-----|----------|
| `$.Values` | `WerfServiceValues` | Параметры чарта с типизированными сервисными значениями werf |
| `$.Release` | `Release` | Информация о релизе |
| `$.Chart` | `ChartMetadata` | Метаданные из Chart.yaml |
| `$.Capabilities` | `Capabilities` | Возможности кластера (API-версии, версия Kubernetes) |
| `$.Files` | `Record<string, Uint8Array>` | Файлы чарта (кроме `templates/` и `ts/`) |

Подробнее о параметрах и их формировании — в разделе [Параметризация шаблонов]({{ "/usage/deploy/values.html" | true_relative_url }}).

### Сервисные значения werf

При использовании `WerfRenderContext` в `$.Values.global.werf` доступны типизированные сервисные параметры:

```typescript
$.Values.global.werf.name       // имя проекта
$.Values.global.werf.version    // версия werf
$.Values.global.werf.repo       // адрес container registry
$.Values.global.werf.commit     // информация о коммите (hash, date)
$.Values.global.werf.images     // собранные образы с тегами и digest'ами
```

## Сторонние библиотеки

Для установки библиотек используйте `deno install` из директории `ts/` чарта:

```shell
cd .helm/ts
deno install npm:kubernetes-models
```

Зависимость будет добавлена в `deno.json` автоматически. Пример использования:

```typescript
// .helm/ts/src/deployment.ts:
import { Deployment } from 'kubernetes-models/apps/v1';

export function newDeployment($: WerfRenderContext): object {
  return new Deployment({
    metadata: { name: 'myapp' },
    spec: {
      replicas: $.Values.replicaCount ?? 1,
      selector: { matchLabels: { app: 'myapp' } },
      template: {
        metadata: { labels: { app: 'myapp' } },
        spec: {
          containers: [{ name: 'myapp', image: 'nginx:latest' }],
        },
      },
    },
  }).toJSON();
}
```

## Сборка и дистрибуция

При выполнении `werf converge`, `werf render`, `werf lint` и `werf plan` TypeScript-код собирается автоматически — бандл формируется в памяти и передаётся в Deno для исполнения.

Для явной сборки в JavaScript-бандл:

```shell
werf chart ts build
```

Сформированный бандл записывается в файл `ts/dist/bundle.js`. Обратите внимание, что из-за [гиттерминизма]({{ "/usage/project_configuration/giterminism.html" | true_relative_url }}) этот файл должен быть закоммичен в Git, либо следует использовать флаг `--dev`.

Если файл `ts/dist/bundle.js` уже существует, при рендеринге он используется как есть. Для принудительной пересборки из исходников:

```shell
werf render --ignore-bundle-js
```

### Бандлы

При публикации бандла TypeScript-бандл автоматически собирается и включается в пакет:

```shell
werf bundle publish --repo registry.example.org/myapp
```

При применении бандла TypeScript-рендеринг выполняется автоматически:

```shell
werf bundle apply --repo registry.example.org/myapp --env production
```

Рендеринг манифестов из бандла без применения:

```shell
werf bundle render --repo registry.example.org/myapp --env production
```

Предварительный просмотр изменений:

```shell
werf bundle plan --repo registry.example.org/myapp --env production
```

### Зависимые чарты

Зависимые чарты тоже могут содержать TypeScript-код — рендеринг выполняется рекурсивно. Пересборка бандла из исходников возможна только для локальных зависимых чартов. Если зависимый чарт загружен из репозитория, но содержит готовый `ts/dist/bundle.js` (например, собранный при публикации через `werf bundle publish`), рендеринг будет выполнен из него.

## Активация и окружение выполнения

Функция является экспериментальной и по умолчанию отключена:

```shell
export NELM_FEAT_TYPESCRIPT=true
```

Без этой переменной команды `werf chart ts init` и `werf chart ts build` завершатся с ошибкой, а TypeScript-рендеринг при `werf converge`, `werf render` и других командах будет пропущен без ошибки.

### Deno

TypeScript-код исполняется рантаймом [Deno](https://deno.com/). Бинарный файл Deno скачивается автоматически при первом использовании и кешируется локально. Если автоматическое скачивание недоступно (например, в закрытом контуре), укажите путь к предустановленному бинарнику:

```shell
werf render --deno-binary-path /usr/local/bin/deno
```

При первом запуске `werf render`, `werf converge` или другой команды, работающей с чартом, Deno создаёт файл `deno.lock` в директории `ts/`. Этот файл также обновляется при добавлении или изменении зависимостей. После инициализации чарта или изменений в `deno.json` рекомендуется выполнить `deno install` из директории `ts/`, чтобы установить зависимости, обновить `deno.lock` и обеспечить корректную работу IDE.

Из-за [гиттерминизма]({{ "/usage/project_configuration/giterminism.html" | true_relative_url }}) файл `deno.lock` необходимо закоммитить в Git, иначе последующие запуски werf завершатся с ошибкой. Альтернативно можно использовать флаг `--dev`.

### Песочница

TypeScript-код выполняется в изолированной песочнице: нет доступа к сети, переменным окружения, запуску процессов; файловая система ограничена файлами обмена данными между werf и Deno. Это гарантирует детерминированность и безопасность рендеринга.
