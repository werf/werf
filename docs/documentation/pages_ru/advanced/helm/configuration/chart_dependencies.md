---
title: Зависимости чартов
permalink: advanced/helm/configuration/chart_dependencies.html
---

**Сабчарт** — это helm-чарт, который включён в текущий чарт в качестве зависимости. werf позволяет использование сабчартов тем же способом [как helm](https://helm.sh/docs/topics/charts/). Чарт может включать произвольное количество зависимых сабчартов. Использование проекта werf в качестве сабчарта в другом проекте werf на данный момент не поддерживается.

Сабчарты располагаются в директории `.helm/charts/SUBCHART_DIR`. Каждый сабчарт в директории `SUBCHART_DIR` сам по себе является чартом и имеет схожую файловую структуру (каждый сабчарт может в свою очередь также содержать сабчарт).

Во время процесса деплоя werf рендерит, создает и отслеживает все ресурсы всех сабчартов.

## Включение сабчарта для проекта

 1. Определим зависимость `redis` для нашего werf чарта с помощью файла `.helm/Chart.yaml`:

     ```yaml
     # .helm/Chart.yaml
     apiVersion: v2
     dependencies:
       - name: redis
         version: "12.7.4"
         repository: "https://charts.bitnami.com/bitnami"
     ```

     **ЗАМЕЧАНИЕ.** Необязательно определять полный `Chart.yaml` с именем и версией как для стандартного helm-чарта. werf генерирует имя чарта и версию на основе директивы `project` из файла `werf.yaml`. См. больше информации [в статье про чарты]({{ "/advanced/helm/configuration/chart.html" | true_relative_url }}).

 2. Далее требуется сгенерировать `.helm/Chart.lock` с помощью команды `werf helm dependency update`.

     ```shell
     werf helm dependency update .helm
     ```

     Данная команда создаст `.helm/Chart.lock` и скачает все зависимости в директорию `.helm/charts`.

 3. Файл `.helm/Chart.lock` следует коммитнуть в git репозиторий, а директорию `.helm/charts` можно добавить в `.gitignore`.

Позднее, во время процесса деплоя (командой [`werf converge`]({{ "/reference/cli/werf_converge.html" | true_relative_url }}) или [`werf bundle apply`]({{ "/reference/cli/werf_bundle_apply.html" | true_relative_url }})) или рендеринга шаблонов (командой [`werf render`]({{ "/reference/cli/werf_render.html" | true_relative_url }})), werf автоматически скачает все зависимости указанные в lock-файле `.helm/Chart.lock`.

**ЗАМЕЧАНИЕ.** Файл `.helm/Chart.lock` должен быть коммитнут в git репозиторий, больше информации [в статье про гитерминизм]({{ "/advanced/helm/configuration/giterminism.html#сабчарты-и-гитерминизм" | true_relative_url }}).

## Конфигурация зависимостей

<!-- Move to reference -->

Опишем формат зависимостей в файле `.helm/Chart.yaml`.

 - `name` — имя чарта, которое должно совпадать с именем (параметр `name`) в файле Chart.yaml соответствующего чарта — зависимости.
 - `version` — версия чарта согласно схеме семантического версионирования, либо диапазон версий.
 - `repository` — URL **репозитория чартов**. Helm ожидает, что добавив `/index.yaml` к URL, он получит список чартов репозитория. Значение `repository` может быть псевдонимом, который в этом случае должен начинаться с префикса `alias:` или `@`.

Файл `.helm/Chart.lock` содержит точные версии прямых зависимостей, версии зависимостей прямых зависимостей и т.д.

Для работы с файлом зависимостей существуют команды `werf helm dependency`, которые упрощают синхронизацию между желаемыми зависимостями и фактическими зависимостями, указанными в папке чарта:
* [werf helm dependency list]({{ "reference/cli/werf_helm_dependency_list.html" | true_relative_url }}) — проверка зависимостей и их статуса.
* [werf helm dependency update]({{ "reference/cli/werf_helm_dependency_update.html" | true_relative_url }}) — обновление папки `.helm/charts` согласно содержимому файла `.helm/Chart.yaml`.
* [werf helm dependency build]({{ "reference/cli/werf_helm_dependency_build.html" | true_relative_url }}) — обновление `.helm/charts` согласно содержимому файла `.helm/Chart.lock`.

Все репозитории чартов, используемые в `.helm/Chart.yaml`, должны быть настроены в системе. Для работы с репозиториями чартов можно использовать команды `werf helm repo`:
* [werf helm repo add]({{ "reference/cli/werf_helm_repo_add.html" | true_relative_url }}) — добавление репозитория чартов.
* [werf helm repo index]({{ "reference/cli/werf_helm_repo_index.html" | true_relative_url }}).
* [werf helm repo list]({{ "reference/cli/werf_helm_repo_list.html" | true_relative_url }}) — вывод списка существующих репозиториев чартов.
* [werf helm repo remove]({{ "reference/cli/werf_helm_repo_remove.html" | true_relative_url }}) — удаление репозитория чартов.
* [werf helm repo update]({{ "reference/cli/werf_helm_repo_update.html" | true_relative_url }}) — обновление локального индекса репозиториев чартов.

werf совместим с настройками Helm, поэтому по умолчанию команды `werf helm dependency` и `werf helm repo` используют настройки из папки конфигурации Helm в домашней папке пользователя, — `~/.helm`. Вы можете указать другую папку с помощью параметра `--helm-home`. Если у вас нет папки `~/.helm` в домашней папке, либо вы хотите создать другую, то вы можете использовать команду `werf helm repo init` для инициализации необходимых настроек и конфигурации репозитория чартов по умолчанию.

## Сабчарты и их конфигурация

Чтобы передать данные из родительского чарта в сабчарт `mysubchart` необходимо определить следующие [values]({{ "/advanced/helm/configuration/values.html" | true_relative_url }}) в родительском чарте:

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

## Устаревшие файлы requirements.yaml и requirements.lock

Устаревший формат описания зависимостей через файлы `.helm/requirements.yaml` и `.helm/requirements.lock` тоже поддерживается werf. Однако рекомендуется переходить на `.helm/Chart.yaml` и `.helm/Chart.lock`.
