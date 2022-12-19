---
title: Чарты
permalink: usage/deploy/charts.html
---

## Чарт

Чарт в werf это Helm-чарт с некоторыми дополнительными возможностями. А Helm-чарт это распространяемый пакет с Helm-шаблонами, values-файлами и некоторыми метаданными. Из чарта формируются конечные Kubernetes-манифесты для дальнейшего развертывания.

Типичный чарт выглядит следующим образом:

```
chartname/
  charts/                   
    dependent-chart/
      # ...
  templates/
    deployment.yaml  
    _helpers.tpl
    NOTES.txt
  crds/
    crd.yaml
  secret/                   # Только в werf
    some-secret-file
  values.yaml
  values.schema.json
  secret-values.yaml        # Только в werf
  Chart.yaml
  Chart.lock
  README.md
  LICENSE
  .helmignore
```

Подробнее:

- `charts/*` — зависимые чарты, чьи Helm-шаблоны/values-файлы используются для формирования манифестов наравне с Helm-шаблонами/values-файлами родительского чарта;

- `templates/*.yaml` — Helm-шаблоны, из которых формируются Kubernetes-манифесты;

- `templates/*.tpl` — файлы с Helm-шаблонами для использования в других Helm-шаблонах. Результат шаблонизации этих файлов игнорируется;

- `templates/NOTES.txt` — werf отображает содержимое этого файла в терминале в конце каждого удачного развертывания;

- `crds/*.yaml` — [Custom Resource Definitions](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/#customresourcedefinitions), которые развертываются до развертывания манифестов в `templates`;

- `secret/*` — (только в werf) зашифрованные файлы, расшифрованное содержимое которых можно подставлять в Helm-шаблоны;

- `values.yaml` — файлы с декларативной конфигурацией для использования в Helm-шаблонах. Конфигурация в них может переопределяться переменными окружения, аргументами командной строки или другими values-файлами;

- `values.schema.json` — JSON-схема для валидации `values.yaml`;

- `secret-values.yaml` — (только в werf) зашифрованный файл с декларативной конфигурацией, аналогичный `values.yaml`. Его расшифрованное содержимое объединяется с `values.yaml` во время формирования манифестов;

- `Chart.yaml` — основная конфигурация и метаданные чарта;

- `Chart.lock` — lock-файл, защищающий от нежелательного изменения/обновления зависимых чартов;

- `README.md` — документация чарта;

- `LICENSE` — лицензия чарта;

- `.helmignore` — список файлов в директории чарта, которые не нужно включать в чарт при его публикации.

### Основной чарт

При запуске команд вроде `werf converge` или `werf render` werf по умолчанию использует чарт, лежащий в директории `<корень Git-репозитория>/.helm`. Этот чарт называют основным чартом.

Директорию с основным чартом можно изменить директивой `deploy.helmChartDir` файла `werf.yaml`.

Если в основном чарте нет файла `Chart.yaml`, то werf использует имя проекта из `werf.yaml` в качестве имени чарта и версию чарта `1.0.0`.

Если вы хотите использовать чарт из OCI/HTTP-репозитория вместо локального чарта или разворачивать несколько чартов сразу, то просто укажите интересующие вас чарты как зависимые для основного чарта. Используйте для этого директиву `dependencies` файла `Chart.yaml`.

### Конфигурация чарта

Основной конфигурационный файл чарта — `Chart.yaml`. В нём указываются имя, версия и другие параметры чарта, а также настраиваются зависимые чарты, например:

```yaml
# Chart.yaml:
apiVersion: v2
name: mychart
version: 1.0.0-anything-here

# Опционально:
type: application
kubeVersion: "~1.20.0"
dependencies:
  - name: nginx
    version: 1.2.3
    repository: https://example.com/charts

# Опциональные информационные директивы:
appVersion: "1.0"
deprecated: false
icon: https://example.org/mychart-icon.svg
description: This is My Chart
home: https://example.org
sources:
  - https://github.com/my/chart
keywords: 
  - apps
annotations:
  anyAdditionalInfo: here
maintainters:
  - name: John Doe
    email: john@example.org
    url: https://john.example.org
```

Обязательные директивы:

* `apiVersion` — формат `Chart.yaml`, либо `v2` (рекомендуется), либо `v1`;

* `name` — имя чарта;

* `version` — версия чарта согласно ["Семантическое версионирование 2.0"](https://semver.org/spec/v2.0.0.html);

Опциональные директивы:

* `type` — тип чарта;

* `kubeVersion` — совместимые версии Kubernetes;

* `dependencies` — конфигурация зависимых чартов;

Опциональные информационные директивы:

* `appVersion` — версия приложения, устанавливаемого чартом, в произвольном формате;

* `deprecated` — является ли чарт устаревшим, `false` (по умолчанию) или `true`;

* `icon` — URL к иконке чарта;

* `description` — описание чарта;

* `home` — веб-сайт чарта;

* `sources` — репозитории с содержимым чарта;

* `keywords` — ключевые слова, связанные с чартом;

* `annotations` — дополнительная произвольная информация о чарте;

* `maintainers`  — список разработчиков чарта;

### Тип чарта

Тип чарта указывается в директиве `type` файла `Chart.yaml`. Допустимые типы:

* `application` — обычный чарт, без ограничений;

* `library` — чарт, содержащий только шаблоны, и не формирующий никаких манифестов сам по себе. Такой чарт нельзя установить, только использовать как зависимый.

### Совместимые версии Kubernetes

Указать с какими версиями Kubernetes совместим данный чарт можно директивой `kubeVersion`, например:

```yaml
# Chart.yaml:
name: mychart
kubeVersion: "1.20.0"   # Совместимо с 1.20.0
kubeVersion: "~1.20.3"  # Совместимо со всеми 1.20.x, но минимум 1.20.3
kubeVersion: "~1"       # Совместимо со всеми 1.x.x
kubeVersion: ">= 1.20.3 < 2.0.0 && != 1.20.4"
```

## Зависимости от других чартов

Чарт может зависеть от других чартов. В таком случае манифесты сформируются и для родительского и для зависимых чартов, после чего объединятся вместе для дальнейшего развертывания. Зависимые чарты настраиваются директивой `dependencies` файла `Chart.yaml`, например:

```yaml
# Chart.yaml:
dependencies:
- name: backend
  version: "~1.2.3"
  repository: https://example.com/charts
```

Не рекомендуется указывать зависимые чарты в `requirements.yaml` вместо `Chart.yaml`. Файл `requirements.yaml` устарел.

### Name, alias

`name` — оригинальное имя зависимого чарта, указанное разработчиком чарта.

Если нужно подключить несколько чартов с одинаковым `name` или подключить один и тот же чарт несколько раз, то используйте директиву `alias`, чтобы поменять имена подключаемых чартов, например:

```yaml
# Chart.yaml:
dependencies:
- name: backend
  alias: main-backend
- name: backend
  alias: secondary-backend
```

### Version

`version` — ограничение подходящих версий зависимого чарта, например:

```yaml
# Chart.yaml:
dependencies:
- name: backend
  version: "1.2.3"   # Использовать 1.2.3
  version: "~1.2.3"  # Использовать последнюю 1.2.x, но минимум 1.2.3
  version: "~1"      # Использовать последнюю 1.x.x
  version: "^1.2.3"  # Использовать последнюю 1.x.x, но минимум 1.2.3
  version: ">= 1.2.3 < 2.0.0 && != 1.2.4"
```

Полную справку по синтаксису ограничения версии можно найти [здесь](https://github.com/Masterminds/semver#checking-version-constraints).

### Repository

`repository` — путь к источнику чартов, в котором можно найти указанный зависимый чарт, например:

```yaml
# Chart.yaml:
dependencies:  
- name: mychart
  # OCI-репозиторий (рекомендуется):
  repository: oci://example.org/myrepo
  # HTTP-репозиторий:
  repository: https://example.org/myrepo
  # Короткое имя репозитория (если репозиторий добавлен через
  # `werf helm repo add shortreponame`):
  repository: alias:myrepo
  # Короткое имя репозитория (альтернативный синтаксис)
  repository: "@myrepo"
  # Локальный чарт из произвольного места:
  repository: file://../mychart
  # Для локального чарта из "charts/" репозиторий не указывается:
  # repository:
```

### Condition, tags

По умолчанию все зависимые чарты включены. Для произвольного включения/выключения чартов можно использовать директиву `condition`. Пример:

```yaml
# Chart.yaml:
dependencies:
- name: backend 
  # Чарт будет включен, если .Values.backend.enabled == true:
  condition: backend.enabled  
```

Также можно использовать директиву `tags` для включения/выключения целых групп чартов сразу. Пример:

```yaml
# Chart.yaml:
dependencies:
# Включить backend, если .Values.tags.app == true:
- name: backend
  tags: ["app"]
# Включить frontend, если .Values.tags.app == true:
- name: frontend
  tags: ["app"]
# Включить database, если .Values.tags.database == true:
- name: database
  tags: ["database"]
```

### Export-values, import-values

`export-values` — (только в werf) автоматический проброс указанных Values из родительского чарта в зависимый. Пример:

```yaml
# Chart.yaml:
dependencies:
- name: backend
  export-values:
  # Автоматически подставлять .Values.werf в .Values.backend.werf:
  - parent: werf
    child: werf
  # Автоматически подставлять .Values.exports.myMap.* в .Values.backend.*:
  - parent: exports.myMap
    child: .
  # Более короткая форма предыдущего экспорта:  
  - myMap
```

`import-values` — автоматический проброс указанных Values из зависимого чарта в родительский, т. е. в обратном направлении. Пример:

```yaml
# Chart.yaml:
dependencies:
- name: backend
  import-values:
  # Автоматически подставлять .Values.backend.from в .Values.to:
  - child: from
    parent: to
  # Автоматически подставлять .Values.backend.exports.myMap.* в .Values.*:
  - child: exports.myMap
    parent: .
  # Более короткая форма предыдущего импорта:  
  - myMap
```

### Добавление/обновление зависимых чартов

Рекомендуемый способ добавления/обновления зависимых чартов:

1. Добавить/обновить конфигурацию зависимого чарта в `Chart.yaml`;

2. При использовании приватного OCI или HTTP-репозитория c зависимым чартом добавьте OCI или HTTP-репозиторий вручную с `werf helm repo add`, указав нужные опции для доступа к ним;

3. Вызовите `werf helm dependency update`, который обновит `Chart.lock`;

4. Закоммитите обновлённые `Chart.yaml` и `Chart.lock` в Git.

Также рекомендуется добавить `.helm/charts/**.tgz` в `.gitignore`.

## Публикация

Рекомендуемый способ публикации чарта — публикация бандла (который по существу и является чартом) в OCI-репозиторий:

1. Разместите чарт в `.helm`;

2. Если ещё нет `werf.yaml`, то создайте его:
   
   ```yaml
   # werf.yaml:
   project: mychart
   configVersion: 1
   ```

3. Опубликуйте содержимое `.helm` как чарт `example.org/charts/mychart:v1.0.0` в виде OCI-образа:
   
   ```shell
   werf bundle publish --repo example.org/charts --tag v1.0.0
   ```

### Публикация нескольких чартов из одного Git-репозитория

Разместите `.helm` с содержимым чарта и соответствующий ему `werf.yaml` в отдельную директорию для каждого чарта:

```
chart1/
  .helm/
  werf.yaml
chart2/
  .helm/
  werf.yaml
```

Теперь опубликуйте каждый чарт по отдельности:

```shell
cd chart1
werf bundle publish --repo example.org/charts --tag v1.0.0

cd ../chart2
werf bundle publish --repo example.org/charts --tag v1.0.0
```

### .helmignore

Файл `.helmignore`, находящийся в корне чарта, может содержать фильтры по именам файлов, при соответствии которым файлы *не будут добавляться* в чарт при публикации. Формат правил такой же, как и в [.gitignore](https://git-scm.com/docs/gitignore), за исключением:

- `**` не поддерживается;

- `!` в начале строки не поддерживается;

- `.helmignore` не исключает сам себя по умолчанию.
