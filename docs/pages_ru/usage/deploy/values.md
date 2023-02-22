---
title: Параметризация шаблонов
permalink: usage/deploy/values.html
---

## Основы параметризации

Содержимое словаря `$.Values` можно использовать для параметризации шаблонов. Каждый чарт имеет свой словарь `$.Values`. Словарь формируется слиянием параметров, полученных из файлов параметров, опций командной строки и других источников.

Простой пример параметризации через `values.yaml`:

```yaml
# values.yaml:
myparam: myvalue
```

{% raw %}

```
# templates/example.yaml:
{{ $.Values.myparam }}
```

{% endraw %}

Результат:

```yaml
myvalue
```

Более сложный пример:

```yaml
# values.yaml:
myparams:
- value: original
```

{% raw %}

```
# templates/example.yaml:
{{ (index $.Values.myparams 0).value }}
```

{% endraw %}

```
werf render --set myparams[0].value=overriden
```

Результат:

```yaml
overriden
```

## Источники параметров и их приоритет

Словарь `$.Values` формируется объединением параметров из источников параметров в указанном порядке:

1. `values.yaml` текущего чарта.
2. `secret-values.yaml` текущего чарта (только в werf).
3. Словарь в `values.yaml` родительского чарта, у которого ключ — алиас или имя текущего чарта.
4. Словарь в `secret-values.yaml` родительского чарта (только в werf), у которого ключ — алиас или имя текущего чарта.
5. Файлы параметров из переменной `WERF_VALUES_*`.
6. Файлы параметров из опции `--values`.
7. Файлы секретных параметров из переменной `WERF_SECRET_VALUES_*`.
8. Файлы секретных параметров из опции `--secret-values`.
9. Параметры в set-файлах из переменной `WERF_SET_FILE_*`.
10. Параметры в set-файлах из опции `--set-file`.
11. Параметры из переменной `WERF_SET_STRING_*`.
12. Параметры из опции `--set-string`.
13. Параметры из переменной `WERF_SET_*`.
14. Параметры из опции `--set`.
15. Служебные параметры werf.
16. Параметры из директивы `export-values` родительского чарта (только в werf).
17. Параметры из директивы `import-values` дочерних чартов.

Правила объединения параметров:

* простые типы данных перезаписываются;

* списки перезаписываются;

* словари объединяются;

* при конфликтах параметры из источников выше по списку перезаписываются параметрами из источников ниже по списку.

## Параметризация чарта

Чарт можно параметризовать через его файл параметров:

```yaml
# values.yaml:
myparam: myvalue
```

{% raw %}

```
# templates/example.yaml:
{{ $.Values.myparam }}
```

{% endraw %}

Результат:

```
myvalue
```

Также добавить/переопределить параметры чарта можно и аргументами командной строки:

```shell
werf render --set myparam=overriden  # или WERF_SET_MYPARAM=myparam=overriden werf render
```

```shell
werf render --set-string myparam=overriden  # или WERF_SET_STRING_MYPARAM=myparam=overriden werf render
```

... или дополнительными файлами параметров:

```yaml
# .helm/values-production.yaml:
myparam: overriden
```

```shell
werf render --values .helm/values-production.yaml  # или WERF_VALUES_PROD=.helm/values-production.yaml werf render
```

... или файлом секретных параметров основного чарта (только в werf):

```yaml
# .helm/secret-values.yaml:
myparam: <encrypted>
```

```shell
werf render
```

... или дополнительными файлами секретных параметров основного чарта (только в werf):

```yaml
# .helm/secret-values-production.yaml:
myparam: <encrypted>
```

```shell
werf render --secret-values .helm/secret-values-production.yaml  # или WERF_SECRET_VALUES_PROD=.helm/secret-values-production.yaml werf render
```

... или set-файлами:

```
# myparam.txt:
overriden
```

```shell
werf render --set-file myparam=myparam.txt  # или WERF_SET_FILE_PROD=myparam=myparam.txt werf render
```

Результат везде тот же:

```
overriden
```

## Параметризация зависимых чартов

Зависимый чарт можно параметризовать как через его собственный файл параметров, так и через файл параметров родительского чарта.

К примеру, здесь параметры из словаря `mychild` в файле `values.yaml` чарта `myparent` перезаписывают параметры в файле `values.yaml` чарта `mychild`:

```yaml
# Chart.yaml:
name: myparent
dependencies:
- name: mychild
```

```yaml
# values.yaml:
mychild:
  myparam: overriden
```

```yaml
# charts/mychild/values.yaml:
myparam: original
```

{% raw %}

```
# charts/mychild/templates/example.yaml:
{{ $.Values.myparam }}
```

{% endraw %}

Результат:

```
overriden
```

Обратите внимание, что словарь, находящийся в `values.yaml` родительского чарта и содержащий параметры для зависимого чарта, должен иметь в качестве имени `alias` (если есть) или `name` зависимого чарта.

Также добавить/переопределить параметры зависимого чарта можно и аргументами командной строки:

```shell
werf render --set mychild.myparam=overriden  # или WERF_SET_MYPARAM=mychild.myparam=overriden werf render
```

```shell
werf render --set-string mychild.myparam=overriden  # или WERF_SET_STRING_MYPARAM=mychild.myparam=overriden werf render
```

... или дополнительными файлами параметров:

```yaml
# .helm/values-production.yaml:
mychild:
  myparam: overriden
```

```shell
werf render --values .helm/values-production.yaml  # или WERF_VALUES_PROD=.helm/values-production.yaml werf render
```

... или файлом секретных параметров основного чарта (только в werf):

```yaml
# .helm/secret-values.yaml:
mychild:
  myparam: <encrypted>
```

```shell
werf render
```

... или дополнительными файлами секретных параметров основного чарта (только в werf):

```yaml
# .helm/secret-values-production.yaml:
mychild:
  myparam: <encrypted>
```

```shell
werf render --secret-values .helm/secret-values-production.yaml  # или WERF_SECRET_VALUES_PROD=.helm/secret-values-production.yaml werf render
```

... или set-файлами:

```
# mychild-myparam.txt:
overriden
```

```shell
werf render --set-file mychild.myparam=mychild-myparam.txt  # или WERF_SET_FILE_PROD=mychild.myparam=mychild-myparam.txt werf render
```

... или директивой `export-values` (только в werf):

```yaml
# Chart.yaml:
name: myparent
dependencies:
- name: mychild
  export-values:
  - parent: myparam
    child: myparam
```

```yaml
# values.yaml:
myparam: overriden
```

```shell
werf render
```

Результат везде тот же:

```
overriden
```

## Использование параметров зависимого чарта в родительском

Для передачи параметров зависимого чарта в родительский можно использовать  директиву `import-values` в родительском чарте:

```yaml
# Chart.yaml:
name: myparent
dependencies:
- name: mychild
  import-values:
  - child: myparam
    parent: myparam
```

```yaml
# values.yaml:
myparam: original
```

```yaml
# charts/mychild/values.yaml:
myparam: overriden
```

{% raw %}

```
# templates/example.yaml:
{{ $.Values.myparam }}
```

{% endraw %}

Результат:

```
overriden
```

## Глобальные параметры

Параметры чарта доступны только в этом же чарте (и ограниченно доступны в зависимых от него). Один из простых способов получить доступ к параметрам одного чарта в других подключенных чартах — использование глобальных параметров.

**Глобальный параметр имеет глобальную область видимости** — параметр, объявленный в родительском, дочернем или другом подключенном чарте становится доступен *во всех подключенных чартах* по одному и тому же пути:

```yaml
# Chart.yaml:
name: myparent
dependencies:
- name: mychild1
- name: mychild2
```

```yaml
# charts/mychild1/values.yaml:
global:
  myparam: myvalue
```

{% raw %}

```
# templates/example.yaml:
myparent: {{ $.Values.global.myparam }}
```

{% endraw %}

{% raw %}

```
# charts/mychild1/templates/example.yaml:
mychild1: {{ $.Values.global.myparam }}
```

{% endraw %}

{% raw %}

```
# charts/mychild2/templates/example.yaml:
mychild2: {{ $.Values.global.myparam }}
```

{% endraw %}

Результат:

```yaml
myparent: myvalue
---
mychild1: myvalue
---
mychild2: myvalue
```

## Секретные параметры (только в werf)

Для хранения секретных параметров можно использовать файлы секретных параметров, хранящиеся в зашифрованном виде в Git-репозитории.

По умолчанию werf пытается найти файл `.helm/secret-values.yaml`, содержащий зашифрованные параметры, и при нахождении файла расшифровывает его и объединяет расшифрованные параметры с остальными:

```yaml
# .helm/values.yaml:
plainParam: plainValue
```

```yaml
# .helm/secret-values.yaml:
secretParam: 1000625c4f1d874f0ab853bf1db4e438ad6f054526e5dcf4fc8c10e551174904e6d0
```

{% raw %}

```
{{ $.Values.plainParam }}
{{ $.Values.secretParam }}
```

{% endraw %}

Результат:

```
plainValue
secretValue
```

### Работа с файлами секретных параметров

Порядок работы с файлами секретных параметров:

1. Возьмите существующий секретный ключ или создайте новый командой `werf helm secret generate-secret-key`.

2. Сохраните секретный ключ в переменную окружения `WERF_SECRET_KEY`, либо в файлы `<корень Git-репозитория>/.werf_secret_key` или `<домашняя директория>/.werf/global_secret_key`.

3. Командой `werf helm secret values edit .helm/secret-values.yaml` откройте файл секретных параметров и добавьте/измените в нём расшифрованные параметры.

4. Сохраните файл — файл зашифруется и сохранится в зашифрованном виде.

5. Закоммитите в Git добавленный/изменённый файл `.helm/secret-values.yaml`;

6. При дальнейших вызовах werf секретный ключ должен быть установлен в вышеупомянутых переменной окружения или файлах, иначе файл секретных параметров не сможет быть расшифрован.

> Имеющий доступ к секретному ключу может расшифровать содержимое файла секретных параметров, поэтому **держите секретный ключ в безопасном месте**!

При использовании файла `<корень Git-репозитория>/.werf_secret_key` обязательно добавьте его в `.gitignore`, чтобы случайно не сохранить его в Git-репозитории.

Многие команды werf можно запускать и без указания секретного ключа благодаря опции `--ignore-secret-key`, но в таком случае параметры будут доступны для использования не в расшифрованной форме, а в зашифрованной.

### Дополнительные файлы секретных параметров

В дополнение к файлу `.helm/secret-values.yaml` можно создавать и использовать дополнительные секретные файлы:

```yaml
# .helm/secret-values-production.yaml:
secret: 1000625c4f1d874f0ab853bf1db4e438ad6f054526e5dcf4fc8c10e551174904e6d0
```

```shell
werf --secret-values .helm/secret-values-production.yaml
```

## Информация о собранных образах (только в werf)

werf хранит информацию о собранных образах в параметрах `$.Values.werf` основного чарта:

```yaml
werf:
  image:
    # Полный путь к собранному Docker-образу для werf-образа "backend":
    backend: example.org/apps/myapp:a243949601ddc3d4133c4d5269ba23ed58cb8b18bf2b64047f35abd2-1598024377816
  # Адрес container registry для собранных образов:
  repo: example.org/apps/myapp
  tag:
    # Тег собранного Docker-образа для werf-образа "backend":
    backend: a243949601ddc3d4133c4d5269ba23ed58cb8b18bf2b64047f35abd2-1598024377816
```

Пример использования:

{% raw %}

```
image: {{ $.Values.werf.image.backend }}
```

{% endraw %}

Результат:

```yaml
image: example.org/apps/myapp:a243949601ddc3d4133c4d5269ba23ed58cb8b18bf2b64047f35abd2-1598024377816
```

Для использования `$.Values.werf` в зависимых чартах воспользуйтесь директивой `export-values` (только в werf):

```yaml
# .helm/Chart.yaml:
dependencies:
- name: backend
  export-values:
  - parent: werf
    child: werf
```

{% raw %}

```
# .helm/charts/backend/templates/example.yaml:
image: {{ $.Values.werf.image.backend }}
```

{% endraw %}

Результат:

```yaml
image: example.org/apps/myapp:a243949601ddc3d4133c4d5269ba23ed58cb8b18bf2b64047f35abd2-1598024377816
```

## Информация о релизе

werf хранит информацию о релизе в свойствах объекта `$.Release`:

```yaml
# Устанавливается ли релиз в первый раз:
IsInstall: true
# Обновляется ли уже существующий релиз:
IsUpgrade: false
# Имя релиза:
Name: myapp-production
# Имя Kubernetes Namespace:
Namespace: myapp-production
# Номер ревизии релиза:
Revision: 1
```

... и в параметрах `$.Values.werf` основного чарта (только в werf):

```yaml
werf:
  # Имя werf-проекта:
  name: myapp
  # Окружение:
  env: production
```

Пример использования:

{% raw %}

```
{{ $.Release.Namespace }}
{{ $.Values.werf.env }}
```

{% endraw %}

Результат:

```
myapp-production
production
```

Для использования `$.Values.werf` в зависимых чартах воспользуйтесь директивой `export-values` (только в werf):

```yaml
# .helm/Chart.yaml:
dependencies:
- name: backend
  export-values:
  - parent: werf
    child: werf
```

{% raw %}

```
# .helm/charts/backend/templates/example.yaml:
{{ $.Values.werf.env }}
```

{% endraw %}

Результат:

```yaml
production
```

## Информация о чарте

werf хранит информацию о текущем чарте в объекте `$.Chart`:

```yaml
# Является ли чарт основным:
IsRoot: true

# Содержимое Chart.yaml:
Name: mychart
Version: 1.0.0
Type: library
KubeVersion: "~1.20.3"
AppVersion: "1.0"
Deprecated: false
Icon: https://example.org/mychart-icon.svg
Description: This is My Chart
Home: https://example.org
Sources:
  - https://github.com/my/chart
Keywords:
  - apps
Annotations:
  anyAdditionalInfo: here
Dependencies:
- Name: redis
  Condition: redis.enabled
```

Пример использования:

{% raw %}

```
{{ $.Chart.Name }}
```

{% endraw %}

Результат:

```
mychart
```

## Информация о шаблоне

werf хранит информацию о текущем шаблоне в свойствах объекта `$.Template`:

```yaml
# Относительный путь к директории templates чарта:
BasePath: mychart/templates
# Относительный путь к текущему файлу шаблона:
Name: mychart/templates/example.yaml
```

Пример использования:

{% raw %}

```
{{ $.Template.Name }}
```

{% endraw %}

Результат:

```
mychart/templates/example.yaml
```

## Информация о Git-коммите (только в werf)

werf хранит информацию о Git-коммите, на котором он был запущен, в параметрах `$.Values.werf.commit` основного чарта:

```yaml
werf:
  commit:
    date:
      # Дата Git-коммита, на котором был запущен werf (человекочитаемая форма):
      human: 2022-01-21 18:51:39 +0300 +0300
      # Дата Git-коммита, на котором был запущен werf (Unix time):
      unix: 1642780299
    # Хэш Git-коммита, на котором был запущен werf:
    hash: 1b28e6843a963c5bdb3579f6fc93317cc028051c
```

Пример использования:

{% raw %}

```
{{ $.Values.werf.commit.hash }}
```

{% endraw %}

Результат:

```
1b28e6843a963c5bdb3579f6fc93317cc028051c
```

Для использования `$.Values.werf.commit` в зависимых чартах воспользуйтесь директивой `export-values` (только в werf):

```yaml
# .helm/Chart.yaml:
dependencies:
- name: backend
  export-values:
  - parent: werf
    child: werf
```

{% raw %}

```
# .helm/charts/backend/templates/example.yaml:
{{ $.Values.werf.commit.hash }}
```

{% endraw %}

Результат:

```yaml
1b28e6843a963c5bdb3579f6fc93317cc028051c
```

## Информация о возможностях кластера Kubernetes

werf предоставляет информацию о возможностях кластера Kubernetes, в который werf стал бы применять Kubernetes-манифесты, через свойства объекта `$.Capabilities`:

```yaml
KubeVersion:
  # Полная версия кластера Kubernetes:
  Version: v1.20.0
  # Мажорная версия кластера Kubernetes:
  Major: "1"
  # Минорная версия кластера Kubernetes:
  Minor: "20"
# API, поддерживаемые кластером Kubernetes:
APIVersions:
- apps/v1
- batch/v1
- # ...
```

... и методы объекта `$.Capabilities`:

* `APIVersions.Has <arg>` — поддерживается ли кластером Kubernetes указанное аргументом API (например, `apps/v1`) или ресурс (например, `apps/v1/Deployment`).

Пример использования:

{% raw %}

```
{{ $.Capabilities.KubeVersion.Version }}
{{ $.Capabilities.APIVersions.Has "apps/v1" }}
```

{% endraw %}

Результат:

```
v1.20.0
true
```
