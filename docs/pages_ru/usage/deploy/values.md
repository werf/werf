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

Словарь `$.Values` формируется объединением параметров из источников параметров в указанном порядке, от низшего приоритета к высшему:

1. `values.yaml` текущего чарта.
2. `secret-values.yaml` текущего чарта (только в werf).
3. Словарь в `values.yaml` родительского чарта, у которого ключ — алиас или имя текущего чарта.
4. Словарь в `secret-values.yaml` родительского чарта (только в werf), у которого ключ — алиас или имя текущего чарта.
5. Файлы параметров из переменной `WERF_VALUES_*`.
6. Файлы параметров из опции `--values`.
7. Файлы секретных параметров из переменной `WERF_SECRET_VALUES_*` (только в werf).
8. Файлы секретных параметров из опции `--secret-values` (только в werf).
9. Параметры из переменной `WERF_SET_JSON_*`.
10. Параметры из опции `--set-json`.
11. Параметры из переменной `WERF_SET_*`.
12. Параметры из опции `--set`.
13. Параметры из переменной `WERF_SET_STRING_*`.
14. Параметры из опции `--set-string`.
15. Параметры в set-файлах из переменной `WERF_SET_FILE_*`.
16. Параметры в set-файлах из опции `--set-file`.
17. Параметры из переменной `WERF_SET_LITERAL_*`.
18. Параметры из опции `--set-literal`.
19. Служебные параметры werf.

Источники 5–19 передаются в основной чарт. Параметры зависимого чарта берутся из словаря, у которого ключ — алиас или имя зависимого чарта, поэтому `--set mychild.myparam=value` перезаписывает `myparam` из `values.yaml` и `secret-values.yaml` чарта `mychild` и из словаря `mychild` родительского чарта.

Правила объединения параметров:

* простые типы данных перезаписываются;

* списки перезаписываются;

* словари объединяются;

* при конфликтах параметры из источников выше по списку перезаписываются параметрами из источников ниже по списку;

* `null`, заданный опциями `--values`, `--secret-values` или `--set*`, убирает значение параметра с тем же ключом, полученное из источников выше по списку.

* `null` для всего словаря зависимого чарта, например `--set mychild=null`, убирает только параметры этого чарта, полученные из опций `--values`, `--secret-values` и `--set*`: у чарта остаются параметры из его `values.yaml` и `secret-values.yaml`, из словаря в `values.yaml` и `secret-values.yaml` родительского чарта и параметры, экспортированные в него из `values.yaml` родительского чарта, из опций `--set*` и из служебных параметров werf. Если в этот чарт экспортируется параметр из опций `--set*`, `null` не действует.

Директивы `export-values` и `import-values` копируют параметры между родительским чартом и зависимыми чартами (см. ниже) и не образуют отдельных источников: копирование выполняется внутри каждого из следующих источников, и скопированный параметр получает приоритет того источника, из которого он взят (от низшего приоритета к высшему):

1. `values.yaml` всех чартов и `secret-values.yaml` зависимых чартов.
2. `secret-values.yaml` основного чарта.
3. Файлы параметров (`WERF_VALUES_*` и `--values`).
4. Файлы секретных параметров (`WERF_SECRET_VALUES_*` и `--secret-values`).
5. Все переменные `WERF_SET_*` и опции `--set*`.
6. Служебные параметры werf.

Например, при `--set g.repo=A --set-string mychild.image.repository=B` и экспорте `g.repo` в `mychild.image.repository` результат — `A`: обе опции относятся к одному источнику, внутри которого побеждает экспортированный параметр.

* `export-values` копирует параметр родительского чарта в зависимый чарт. Внутри одного источника экспортированный параметр перезаписывает параметр зависимого чарта с тем же ключом. Между источниками побеждает параметр из источника ниже по списку. Например, `--set mychild.myparam=value` перезаписывает параметр, экспортированный в `mychild.myparam` из `values.yaml` родительского чарта, а служебный параметр werf, экспортированный в `mychild.image.repository`, перезаписывает `mychild.image.repository` из `--values`.

* `import-values` копирует словарь зависимого чарта в родительский чарт. Внутри одного источника собственный параметр родительского чарта перезаписывает импортированный параметр с тем же ключом. Между источниками побеждает параметр из источника ниже по списку. Например, словарь, импортированный из `--set mychild.mydict.myparam=value`, перезаписывает `mydict.myparam` из `values.yaml` родительского чарта.

Параметр, который нельзя скопировать, пропускается с предупреждением: значение, не являющееся словарём, экспортированное в корень зависимого чарта, или значение, не являющееся словарём или `null`, импортированное из зависимого чарта. Импортированный `null` копируется как `null`.

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

Для передачи параметров зависимого чарта в родительский можно использовать  директиву `import-values` в родительском чарте. Импортировать можно только словари (или `null`), а собственные параметры родительского чарта перезаписывают импортированные:

```yaml
# Chart.yaml:
name: myparent
dependencies:
- name: mychild
  import-values:
  - child: mydict
    parent: mydict
```

```yaml
# values.yaml:
mydict:
  myparam1: original
```

```yaml
# charts/mychild/values.yaml:
mydict:
  myparam1: imported
  myparam2: imported
```

{% raw %}

```
# templates/example.yaml:
{{ $.Values.mydict.myparam1 }} {{ $.Values.mydict.myparam2 }}
```

{% endraw %}

Результат:

```
original imported
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

### Совместимость зашифрованных данных между версиями werf

Начиная с werf v3.3.0 (Nelm v2.0.0-alpha.1), секреты шифруются с помощью AES-GCM, а каждый параметр в файле секретных параметров сохраняет свой YAML-тип: `port: 8080` остаётся числом, `enabled: true` — булевым значением. Секреты, зашифрованные более ранними версиями, читаются как и раньше.

Более ранние версии werf и Nelm не могут прочитать секреты, зашифрованные новыми версиями, и завершаются ошибкой `decode file ...`. Новые версии записывают файлы секретных параметров и секретные файлы в новом формате командами `werf helm secret values edit`, `werf helm secret ... encrypt` и `werf helm secret rotate-secret-key` и всегда используют его для зашифрованных планов, сохранённых через `werf plan --save-plan`. Все werf и Nelm, которые читают такие данные, включая CI-задания и `werf bundle apply` бандлов, которые их содержат, тоже нужно обновить.

Параметры, зашифрованные более ранними версиями, не хранят свой тип, поэтому всегда читаются как строки: `port: 8080` превращается в `port: "8080"`, из-за чего может не пройти проверка `values.schema.json` или измениться логика шаблонов. Перешифровка, например через `werf helm secret rotate-secret-key`, не помогает: тип уже потерян. Восстановить тип можно только введя параметр заново: откройте файл командой `werf helm secret values edit` и уберите кавычки, например замените `"8080"` на `8080`. При сохранении файла все его параметры перешифровываются в новом формате с теми типами, которые они имеют в редакторе, поэтому исправьте все такие параметры за одно редактирование. При загрузке файлов секретных параметров werf предупреждает о параметрах в старом формате, которые выглядят как числа или булевы значения.

### Секретные параметры в релизе

werf сохраняет в релизе параметры из опций `--values`, `--secret-values` и `--set*` вместе с параметрами, которые `export-values` копирует из них, причём параметры из `--secret-values` сохраняются в расшифрованном виде. Их показывает `werf release get --print-values`, их может прочитать любой, у кого есть доступ к Secret релиза в кластере, а изменение одних только `--secret-values` создаёт новую ревизию релиза. Релизы, развёрнутые с `--secret-values` более ранней версией werf, получают одну новую ревизию при первом развёртывании после обновления. Параметры из `secret-values.yaml` чартов среди этих параметров не сохраняются, но если в `Chart.yaml` чарта зависимые чарты перечислены в `dependencies`, сохранённый в релизе чарт содержит их в расшифрованном виде, и `werf release get --print-values` их тоже показывает.

`werf bundle publish` сохраняет файлы `--secret-values` в `secret-values.yaml` бандла в зашифрованном виде, поэтому `werf bundle apply` не сохраняет их в релизе среди параметров опций; как и любой `secret-values.yaml`, они сохраняются в расшифрованном виде в чарте в релизе, если в бандле зависимые чарты перечислены в `dependencies`.

## Информация о собранных образах (только в werf)

werf хранит подробную информацию о собранных образах в `$.Values.global.werf.images`, доступную во всех чартах (основном и зависимых):

```yaml
global:
  werf:
    images:
      backend:
        # Адрес container registry:
        registry: example.org
        # Пространство имён репозитория:
        namespace: apps
        # Короткое имя образа:
        name: myapp
        # Тег образа:
        tag: a243949601ddc3d4133c4d5269ba23ed58cb8b18bf2b64047f35abd2-1598024377816
        # Дайджест образа:
        digest: sha256:ece15c4a8e7e04dd0afa25b85a51e3fba35eb31e0e1489a2f3e916da77e4e865
        # Тег и дайджест вместе:
        tag_digest: "a243949...816@sha256:ece15c4a..."
        # Полное имя образа без тега (registry + repository):
        image: example.org/apps/myapp
        # Полный путь репозитория (namespace + name):
        repository: apps/myapp
        # Полная ссылка на образ с дайджестом:
        ref: "example.org/apps/myapp:a243949...816@sha256:ece15c4a..."
        # Полная ссылка на образ с тегом:
        ref_tag: "example.org/apps/myapp:a243949...816"
        # Путь репозитория с тегом и дайджестом:
        repository_ref: "apps/myapp:a243949...816@sha256:ece15c4a..."
        # Путь репозитория с тегом:
        repository_tag: "apps/myapp:a243949...816"
        # Короткое имя с тегом и дайджестом:
        name_ref: "myapp:a243949...816@sha256:ece15c4a..."
        # Короткое имя с тегом:
        name_tag: "myapp:a243949...816"
```

Этот структурированный формат позволяет использовать отдельные компоненты образа, что удобно при работе с внешними чартами, ожидающими значения образов в структурированном формате:

{% raw %}

```
image:
  registry: {{ $.Values.global.werf.images.backend.registry }}
  repository: {{ $.Values.global.werf.images.backend.repository }}
  tag: {{ $.Values.global.werf.images.backend.tag }}
```

{% endraw %}

Поскольку `$.Values.global.werf` имеет глобальную область видимости, `images` доступен во всех зависимых чартах. Также можно использовать `export-values` для передачи отдельных полей образа в формате, который ожидает внешний чарт:

```yaml
# .helm/Chart.yaml:
dependencies:
- name: postgres-operator
  export-values:
  - parent: global.werf.images.pg.registry
    child: image.registry
  - parent: global.werf.images.pg.repository
    child: image.repository
  - parent: global.werf.images.pg.tag
    child: image.tag
```

Чарт `postgres-operator` получит:

```yaml
image:
  registry: example.org
  repository: apps/pg
  tag: a243949601ddc3d4133c4d5269ba23ed58cb8b18bf2b64047f35abd2-1598024377816
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

... и в `$.Values.global.werf`, доступном во всех чартах (только в werf):

```yaml
global:
  werf:
    # Имя werf-проекта:
    name: myapp
    # Окружение:
    env: production
    # Kubernetes Namespace:
    namespace: myapp-production
```

Для обратной совместимости те же значения также доступны в `$.Values.werf`.

Пример использования:

{% raw %}

```
{{ $.Release.Namespace }}
{{ $.Values.global.werf.env }}
```

{% endraw %}

Результат:

```
myapp-production
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

werf хранит информацию о Git-коммите, на котором он был запущен, в `$.Values.global.werf.commit`, доступном во всех чартах:

```yaml
global:
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

Для обратной совместимости те же значения также доступны в `$.Values.werf.commit`.

Пример использования:

{% raw %}

```
{{ $.Values.global.werf.commit.hash }}
```

{% endraw %}

Результат:

```
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
