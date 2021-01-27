---
title: Работа с зависимостями чарта
sidebar: documentation
permalink: documentation/advanced/helm/working_with_chart_dependencies.html
author: Alexey Igrychev <alexey.igrychev@flant.com>
---

Helm-чарт может содержать произвольное число зависимостей — **сабчартов**.

Сабчарты размещаются в папке `.helm/charts/SUBCHART_DIR`. Каждый сабчарт в папке `SUBCHART_DIR` представляет собой отдельный чарт с похожей структурой файлов (каждый сабчарт может в свою очередь также содержать сабчарт).

Во время процесса деплоя werf рендерит, создает и отслеживает все ресурсы всех сабчартов.

## Управление сабчартами через файл requirements.yaml

Зачастую, в разработке применяется подход хранения зависимостей в одном файле, что позволяет проще управлять зависимостями из одной точки. Этот же подход справедлив и для управления зависимостями сабчартов.

YAML-файл `requirements.yaml` — это файл описания зависимостей, в котором разработчики могут указывать как зависимости чарта, так и его версию, а также путь размещения. Например, следующий файл описания зависимостей определяет две зависимости:

```yaml
# requirements.yaml
dependencies:
- name: nginx
  version: "1.2.3"
  repository: "https://example.com/charts"
- name: sysdig
  version: "1.4.12"
  repository: "@stable"
```

* `name` — имя чарта, которое должно совпадать с именем (параметр `name`) в файле Chart.yaml соответствующего чарта — зависимости.
* `version` — версия чарта согласно схеме семантического версионирования, либо диапазон версий.
* `repository` — URL **репозитория чартов**. Helm ожидает, что добавив `/index.yaml` к URL, он получит список чартов репозитория. Значение `repository` может быть псевдонимом, который в этом случае должен начинаться с префикса `alias:` или `@`.

Файл `requirements.lock` содержит точные версии прямых зависимостей, версии зависимостей прямых зависимостей и т.д.

Для работы с файлом зависимостей существуют команды `werf helm dependency`, которые упрощают синхронизацию между желаемыми зависимостями и фактическими зависимостями, указанными в папке чарта:
* [werf helm dependency list]({{ "documentation/reference/cli/werf_helm_dependency_list.html" | true_relative_url }}) — проверка зависимостей и их статуса.
* [werf helm dependency update]({{ "documentation/reference/cli/werf_helm_dependency_update.html" | true_relative_url }}) — обновление папки `/charts` согласно содержимому файла `requirements.yaml`.
* [werf helm dependency build]({{ "documentation/reference/cli/werf_helm_dependency_build.html" | true_relative_url }}) — обновление `/charts` согласно содержимому файла `requirements.lock`.

Все репозитории чартов, используемые в `requirements.yaml`, должны быть настроены в системе. Для работы с репозиториями чартов можно использовать команды `werf helm repo`:
* [werf helm repo add]({{ "documentation/reference/cli/werf_helm_repo_add.html" | true_relative_url }}) — добавление репозитория чартов.
* [werf helm repo index]({{ "documentation/reference/cli/werf_helm_repo_index.html" | true_relative_url }}).
* [werf helm repo list]({{ "documentation/reference/cli/werf_helm_repo_list.html" | true_relative_url }}) — вывод списка существующих репозиториев чартов.
* [werf helm repo remove]({{ "documentation/reference/cli/werf_helm_repo_remove.html" | true_relative_url }}) — удаление репозитория чартов.
* [werf helm repo update]({{ "documentation/reference/cli/werf_helm_repo_update.html" | true_relative_url }}) — обновление локального индекса репозиториев чартов.

werf совместим с настройками Helm, поэтому по умолчанию команды `werf helm dependency` и `werf helm repo` используют настройки из папки конфигурации Helm в домашней папке пользователя, — `~/.helm`. Вы можете указать другую папку с помощью параметра `--helm-home`. Если у вас нет папки `~/.helm` в домашней папке, либо вы хотите создать другую, то вы можете использовать команду `werf helm repo init` для инициализации необходимых настроек и конфигурации репозитория чартов по умолчанию.

## Сабчарты и данные

Чтобы передать данные из родительского чарта в сабчарт `mysubchart` необходимо определить следующие данные в родительском чарте:

```yaml
mysubchart:
  key1:
    key2:
    - key3: value
```

В сабчарте `mysubchart` эти данные можно использовать с помощью обращения к соответствующим параметрам без указания ключа `mysubchart`:

{% raw %}
```yaml
{{ .Values.key1.key2[0].key3 }}
```
{% endraw %}

Данные, определенные глобально в ключе верхнего уровня `global`, также доступны в сабчартах:

```yaml
global:
  database:
    mysql:
      user: user
      password: password
```

Обращаться к ним необходимо как обычно:

{% raw %}
```yaml
{{ .Values.global.database.mysql.user }}
```
{% endraw %}

В сабчарте `mysubchart` будут доступны только данные ключей `mysubchart` и `global`.

**ЗАМЕЧАНИЕ** Файлы `secret-values.yaml` сабчартов не будут использоваться во время процесса деплоя, несмотря на то, что данные секретов из главного чарта и данные переданные через параметр `--secret-values` будут доступны через массив `.Values` как обычно.
