---
title: Описание манифестов в TypeScript
permalink: usage/deploy/typescript.html
---

> Обратите внимание, что функция TypeScript-рендеринга является экспериментальной. Для её включения установите переменную окружения `NELM_FEAT_TYPESCRIPT=true`.

## Особенности TypeScript-рендеринга

Помимо стандартного способа описания Kubernetes-манифестов через [Helm-шаблоны]({{ "/usage/deploy/templates.html" | true_relative_url }}), werf поддерживает описание манифестов на TypeScript. При использовании TypeScript становятся доступны:

- типизация — опечатка в имени параметра обнаруживается на этапе написания кода, а не при развертывании;
- поддержка IDE — автодополнение, навигация по коду и рефакторинг;
- стандартная отладка — сообщения об ошибках содержат стектрейсы и указывают на конкретное место в коде;
- стандартный синтаксис — обычные функции, циклы и условия вместо специфичных конструкций шаблонизатора.

TypeScript-код получает тот же контекст, что и Helm-шаблоны (`Values`, `Release`, `Chart` и т. д.), и может сосуществовать с ними в одном чарте.

## Простой пример

Инициализация TypeScript-структуры в существующем чарте:

```shell
werf chart ts init
```

Результат: в директории чарта появится поддиректория `ts/` с готовым примером — `Deployment` и `Service`.

Теперь можно отрендерить чарт:

```shell
werf render
```

TypeScript-манифесты будут сгенерированы и объединены с результатами Helm-шаблонов из `templates/`.

### Пример `Deployment` в Helm-шаблонах и TypeScript

В Helm-шаблонах:

{% raw %}

```yaml
# .helm/templates/deployment.yaml:
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ $.Release.Name }}-myapp
  labels:
    app.kubernetes.io/name: {{ $.Chart.Name }}
    app.kubernetes.io/instance: {{ $.Release.Name }}
spec:
  replicas: {{ $.Values.replicaCount | default 1 }}
  selector:
    matchLabels:
      app.kubernetes.io/name: {{ $.Chart.Name }}
      app.kubernetes.io/instance: {{ $.Release.Name }}
  template:
    metadata:
      labels:
        app.kubernetes.io/name: {{ $.Chart.Name }}
        app.kubernetes.io/instance: {{ $.Release.Name }}
    spec:
      containers:
      - name: {{ $.Release.Name }}-myapp
        image: {{ $.Values.image.repository }}:{{ $.Values.image.tag | default "latest" }}
        ports:
        - name: http
          containerPort: {{ $.Values.service.port | default 80 }}
```

{% endraw %}

В TypeScript:

```typescript
// .helm/ts/src/deployment.ts:
import type { RenderContext } from '@nelm/chart-ts-sdk';

export function newDeployment($: RenderContext): object {
  const name = `${$.Release.Name}-myapp`;

  return {
    apiVersion: 'apps/v1',
    kind: 'Deployment',
    metadata: {
      name,
      labels: {
        'app.kubernetes.io/name': $.Chart.Name,
        'app.kubernetes.io/instance': $.Release.Name,
      },
    },
    spec: {
      replicas: $.Values.replicaCount ?? 1,
      selector: {
        matchLabels: {
          'app.kubernetes.io/name': $.Chart.Name,
          'app.kubernetes.io/instance': $.Release.Name,
        },
      },
      template: {
        metadata: {
          labels: {
            'app.kubernetes.io/name': $.Chart.Name,
            'app.kubernetes.io/instance': $.Release.Name,
          },
        },
        spec: {
          containers: [
            {
              name,
              image: ($.Values.image?.repository ?? 'nginx') + ':' + ($.Values.image?.tag ?? 'latest'),
              ports: [{ name: 'http', containerPort: $.Values.service?.port ?? 80 }],
            },
          ],
        },
      },
    },
  };
}
```

Оба варианта формируют одинаковый манифест.

## Структура TypeScript-чарта

После `werf chart ts init` в чарте появляется директория `ts/`:

```
.helm/
  templates/              # Helm-шаблоны (как обычно)
    deployment.yaml
    _helpers.tpl
  ts/                     # TypeScript-манифесты
    src/
      index.ts            # Точка входа
      helpers.ts          # Вспомогательные функции
      deployment.ts       # Генерация Deployment
      service.ts          # Генерация Service
    deno.json             # Конфигурация Deno и зависимости
    tsconfig.json         # Конфигурация TypeScript
    input.example.yaml    # Пример контекста рендеринга для локальной разработки
  values.yaml
  Chart.yaml
```

### Точка входа

werf ищет точку входа в следующем порядке:

1. `ts/src/index.ts`
2. `ts/src/index.js`

Если ни один из файлов не найден, TypeScript-рендеринг для данного чарта пропускается.

### Файл `index.ts`

Файл `index.ts` содержит функцию рендеринга и вызов `runRender`:

```typescript
// .helm/ts/src/index.ts:
import { RenderContext, RenderResult, runRender } from '@nelm/chart-ts-sdk';
import { newDeployment } from './deployment.ts';
import { newService } from './service.ts';

function render($: RenderContext): RenderResult {
  const manifests: object[] = [];

  manifests.push(newDeployment($));

  if ($.Values.service?.enabled !== false) {
    manifests.push(newService($));
  }

  return { manifests };
}

await runRender(render);
```

Функция `render` получает контекст рендеринга `$` и возвращает объект с массивом `manifests` — Kubernetes-манифестов в виде обычных JavaScript-объектов. Каждый объект будет сериализован в YAML.

## Контекст рендеринга

TypeScript-код получает тот же контекст, что и Helm-шаблоны. Контекст передаётся в функцию `render` как типизированный объект `RenderContext`:

| Поле | Тип | Описание | Аналог в Helm-шаблонах |
|------|-----|----------|---------------------|
| `$.Values` | `Record<string, any>` | Параметры чарта из values.yaml и всех переопределений | `$.Values` |
| `$.Release` | `Release` | Информация о релизе | `$.Release` |
| `$.Chart` | `ChartMetadata` | Метаданные из Chart.yaml | `$.Chart` |
| `$.Capabilities` | `Capabilities` | Возможности кластера (API-версии, версия Kubernetes) | `$.Capabilities` |
| `$.Files` | `Record<string, Uint8Array>` | Дополнительные файлы чарта (не из `templates/` и не из `ts/`) | `$.Files` |

### Поле `Values`

Словарь `$.Values` формируется так же, как и для Helm-шаблонов: из `values.yaml`, `secret-values.yaml`, опций `--set`, `--values` и других [источников параметров]({{ "/usage/deploy/values.html" | true_relative_url }}). Все механизмы параметризации работают одинаково.

### Поле `Release`

```typescript
$.Release.Name        // имя релиза
$.Release.Namespace   // Namespace релиза
$.Release.Revision    // номер ревизии
$.Release.IsInstall   // true, если это первая установка
$.Release.IsUpgrade   // true, если это обновление
```

### Поле `Chart`

```typescript
$.Chart.Name          // имя чарта
$.Chart.Version       // версия чарта
$.Chart.AppVersion    // версия приложения
$.Chart.Description   // описание
$.Chart.Keywords      // ключевые слова
$.Chart.Home          // домашняя страница
$.Chart.Sources       // ссылки на исходный код
```

### Поле `Capabilities`

```typescript
$.Capabilities.KubeVersion.Version  // например, "v1.28.0"
$.Capabilities.KubeVersion.Major    // "1"
$.Capabilities.KubeVersion.Minor    // "28"
$.Capabilities.APIVersions          // список доступных API-версий
```

## Написание манифестов

### Условное создание ресурсов

В Helm-шаблонах условное создание ресурса требует обернуть весь файл в блок `if`:

{% raw %}

```
# templates/service.yaml:
{{ if $.Values.service.enabled }}
apiVersion: v1
kind: Service
# ...
{{ end }}
```

{% endraw %}

В TypeScript это обычный `if` в функции `render`:

```typescript
function render($: RenderContext): RenderResult {
  const manifests: object[] = [];

  manifests.push(newDeployment($));

  if ($.Values.service?.enabled) {
    manifests.push(newService($));
  }

  return { manifests };
}
```

### Циклы и трансформации данных

В Helm-шаблонах итерация по данным выполняется через `range`:

{% raw %}

```yaml
# templates/configmaps.yaml:
{{ range $name, $data := $.Values.configmaps }}
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ $name }}
data:
  {{ $data | toYaml | nindent 2 }}
{{ end }}
```

{% endraw %}

В TypeScript — стандартными средствами языка:

```typescript
function render($: RenderContext): RenderResult {
  const manifests: object[] = [];

  for (const [name, data] of Object.entries($.Values.configmaps ?? {})) {
    manifests.push({
      apiVersion: 'v1',
      kind: 'ConfigMap',
      metadata: { name },
      data,
    });
  }

  return { manifests };
}
```

### Переиспользование кода

В Helm-шаблонах для переиспользования используются именованные шаблоны в файлах `_*.tpl`:

{% raw %}

```
# templates/_helpers.tpl:
{{ define "myapp.labels" }}
app.kubernetes.io/name: {{ $.Chart.Name }}
app.kubernetes.io/instance: {{ $.Release.Name }}
{{ end }}
```

{% endraw %}

В TypeScript — обычные функции и модули:

```typescript
// ts/src/helpers.ts:
import type { RenderContext } from '@nelm/chart-ts-sdk';

export function getLabels($: RenderContext): Record<string, string> {
  return {
    'app.kubernetes.io/name': $.Chart.Name,
    'app.kubernetes.io/instance': $.Release.Name,
  };
}
```

```typescript
// ts/src/deployment.ts:
import { getLabels } from './helpers.ts';

export function newDeployment($: RenderContext): object {
  return {
    // ...
    metadata: {
      labels: getLabels($),
    },
    // ...
  };
}
```

### Сторонние библиотеки

В TypeScript-чартах можно использовать сторонние библиотеки. Для установки используйте команду `deno install` из директории `ts/` чарта — например, для подключения [kubernetes-models](https://github.com/tommy351/kubernetes-models-ts) со строгой типизацией Kubernetes-ресурсов:

```shell
cd .helm/ts
deno install npm:kubernetes-models
```

Зависимость будет добавлена в секцию `imports` файла `deno.json` автоматически.

```typescript
// .helm/ts/src/deployment.ts:
import { Deployment } from 'kubernetes-models/apps/v1';

export function newDeployment($: RenderContext): object {
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
  });
}
```

При использовании таких библиотек IDE будет подсказывать допустимые поля ресурсов, их типы и обязательность — ошибки в структуре манифестов обнаруживаются ещё до рендеринга.

## Совместная работа с Helm-шаблонами

TypeScript-манифесты и Helm-шаблоны могут сосуществовать в одном чарте. Результаты обоих способов рендеринга объединяются перед развертыванием.

Это позволяет мигрировать постепенно:

1. Добавьте TypeScript-структуру в существующий чарт: `werf chart ts init`.
2. Перенесите один ресурс из `templates/` в `ts/src/`.
3. Убедитесь, что `werf render` выдаёт ожидаемый результат.
4. Повторите для остальных ресурсов.

> Обратите внимание, что TypeScript-манифесты и Helm-шаблоны являются независимыми источниками ресурсов. Именованные шаблоны (`define`/`include`) из `templates/_*.tpl` недоступны в TypeScript-коде, и наоборот.

### Зависимые чарты

Зависимые чарты тоже могут содержать TypeScript-код. Рендеринг выполняется рекурсивно: для каждого зависимого чарта, имеющего точку входа (`ts/src/index.ts` или `ts/src/index.js`), будет выполнен TypeScript-рендеринг.

Пересборка бандла из исходников возможна только для локальных зависимых чартов. Если зависимый чарт загружен из репозитория, но уже содержит готовый `ts/dist/bundle.js` (например, собранный при публикации через `werf bundle publish`), рендеринг будет выполнен из него.

Для зависимых чартов действуют те же правила, что и для Helm-шаблонов: параметры (`Values`) ограничиваются областью видимости зависимого чарта, а информация о релизе, возможностях кластера и рантайме наследуется от родительского чарта.

## Сборка и дистрибуция

### Автоматическая сборка

При выполнении команд `werf converge`, `werf render`, `werf lint` и `werf plan` TypeScript-код собирается автоматически. Явный вызов команды сборки не требуется — бандл формируется в памяти и передаётся в Deno для исполнения.

### Явная сборка

Для явной сборки TypeScript-кода в JavaScript-бандл используйте команду:

```shell
werf chart ts build .helm
```

Сформированный бандл записывается в файл `ts/dist/bundle.js`. Это может быть полезно для отладки или для подготовки чарта к публикации вручную. Обратите внимание, что из-за [гиттерминизма]({{ "/usage/project_configuration/giterminism.html" | true_relative_url }}) сгенерированный файл должен быть закоммичен в Git, либо следует использовать флаг `--dev`.

По умолчанию, если файл `ts/dist/bundle.js` уже существует, при рендеринге он будет использован как есть. Чтобы принудительно пересобрать бандл из исходников, используйте флаг `--ignore-bundle-js`:

```shell
werf render --ignore-bundle-js
```

### Публикация и применение бандлов

При публикации бандла TypeScript-бандл автоматически собирается и включается в пакет:

```shell
werf bundle publish --repo registry.example.org/myapp
```

При применении бандла TypeScript-рендеринг также выполняется автоматически:

```shell
werf bundle apply --repo registry.example.org/myapp --env production
```

Предварительный просмотр изменений с `werf plan` поддерживает TypeScript-манифесты так же, как и Helm-шаблоны:

```shell
werf plan --repo registry.example.org/myapp --env production
```

## Активация и окружение выполнения

### Включение функции

Функция TypeScript-рендеринга является экспериментальной и по умолчанию отключена. Для включения установите переменную окружения:

```shell
export NELM_FEAT_TYPESCRIPT=true
```

Без этой переменной:
- команды `werf chart ts init` и `werf chart ts build` завершатся с ошибкой;
- TypeScript-рендеринг при `werf converge`, `werf render` и других командах будет пропущен без ошибки.

### Deno

TypeScript-код исполняется рантаймом [Deno](https://deno.com/). Бинарный файл Deno скачивается автоматически при первом использовании и кешируется локально.

Если автоматическое скачивание недоступно (например, в закрытом контуре), укажите путь к предустановленному бинарнику:

```shell
werf render --deno-binary-path /usr/local/bin/deno
```

### Ограничения песочницы

TypeScript-код выполняется в изолированной песочнице со строгими ограничениями:

- **нет доступа к сети** — нельзя выполнять HTTP-запросы или открывать сокеты;
- **нет доступа к переменным окружения** — `Deno.env` недоступен;
- **нет запуска процессов** — `Deno.run` недоступен;
- **файловая система ограничена** — доступ только к файлам обмена данными между werf и Deno.

Это гарантирует, что рендеринг манифестов остаётся детерминированным и безопасным.
