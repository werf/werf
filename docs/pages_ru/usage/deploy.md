---
title: Выкат
permalink: usage/deploy.html
---

# Helm

## Обзор

Начать пользоваться werf для выката, используя существующие [Helm](https://helm.sh) чарты, не составит никакого труда, т.к. они полностью совместимы с werf. Конфигурация описывается в формате аналогичном формату [Helm-чарта]({{ "advanced/helm/configuration/chart.html" | true_relative_url }}).

werf включает всю существующую функциональность Helm (он вкомпилен в werf) и свои дополнения:
- несколько настраиваемых режимов отслеживания выкатываемых ресурсов, в том числе обработка логов и событий;
- интеграция собираемых образов с [шаблонами]({{ "advanced/helm/configuration/templates.html" | true_relative_url }}) Helm-чартов;
- возможность простановки произвольных аннотаций и лейблов во все ресурсы, создаваемые в Kubernetes, глобально через опции утилиты werf;
- werf читает все конфигурационные файлы helm из git в соответствии с режимом [гитерминизма]({{ "advanced/giterminism.html" | true_relative_url }}), что позволяет создавать по-настоящему воспроизводимые pipeline'ы в CI/CD и на локальных машинах.
- и другие особенности, о которых пойдёт речь далее.

С учётом всех этих дополнений и способа реализации можно рассматривать werf как альтернативный или улучшенный helm-клиент, для деплоя стандартных helm-совместимых чартов.

Для работы с приложением в Kubernetes используются следующие основные команды:
- [converge]({{ "reference/cli/werf_converge.html" | true_relative_url }}) — для установки или обновления приложения в кластере, и
- [dismiss]({{ "reference/cli/werf_dismiss.html" | true_relative_url }}) — для удаления приложения из кластера.
- [bundle apply]({{ "reference/cli/werf_bundle_apply.html" | true_relative_url }}) — для выката приложения из опубликованного ранее [бандла]({{ "advanced/bundles.html" | true_relative_url }}).

Данная глава покрывает следующие разделы:
1. Конфигурация helm для деплоя вашего приложения в kubernetes с помощью werf: [раздел "конфигурация"]({{ "advanced/helm/configuration/chart.html" | true_relative_url }}).
2. Как werf реализует процесс деплоя: [раздел "процесс деплоя"]({{ "advanced/helm/deploy_process/steps.html" | true_relative_url }}).
3. Что такое релиз и как управлять выкаченными релизами своих приложений: [раздел "релизы"]({{ "advanced/helm/releases/release.html" | true_relative_url }})

## Конфигурация

### Chart

Чарт — набор конфигурационных файлов описывающих приложение. Файлы чарта находятся в папке `.helm`, в корневой папке проекта:

```
.helm/
  templates/
    <name>.yaml
    <name>.tpl
    <some_dir>/
      <name>.yaml
      <name>.tpl
  charts/
  secret/
  values.yaml
  secret-values.yaml
```

Чарт werf может содержать опциональный файл `.helm/Chart.yaml` с описанием чарта, который полностью совместим с [`Chart.yaml`](https://helm.sh/docs/topics/charts/) и может содержать примерно следующее:

```yaml
apiVersion: v2
name: mychart
version: 1.0.0
dependencies:
 - name: redis
   version: "12.7.4"
   repository: "https://charts.bitnami.com/bitnami" 
```

По умолчанию werf будет использовать [имя проекта]({{ "/reference/werf_yaml.html#имя-проекта" | true_relative_url }}) из `werf.yaml` в качестве имени чарта. Версия чарта по умолчанию: `1.0.0`. Это можно переопределить создав файл `.helm/Chart.yaml` с явным переопределением имени чарта или его версии:

```yaml
name: mychart
version: 2.4.6
```

`.helm/Chart.yaml` также требуется для определения [зависимостей чарта]({{ "/advanced/helm/configuration/chart_dependencies.html" | true_relative_url }}).

### Зависимости чартов

**Сабчарт** — это helm-чарт, который включён в текущий чарт в качестве зависимости. werf позволяет использование сабчартов тем же способом [как helm](https://helm.sh/docs/topics/charts/). Чарт может включать произвольное количество зависимых сабчартов. Использование проекта werf в качестве сабчарта в другом проекте werf на данный момент не поддерживается.

Сабчарты располагаются в директории `.helm/charts/SUBCHART_DIR`. Каждый сабчарт в директории `SUBCHART_DIR` сам по себе является чартом и имеет схожую файловую структуру (каждый сабчарт может в свою очередь также содержать сабчарт).

Во время процесса деплоя werf рендерит, создает и отслеживает все ресурсы всех сабчартов.

#### Включение сабчарта для проекта

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

#### Конфигурация зависимостей

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

#### Передача values в сабчарты

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

#### Передача динамических values из родительского чарта в сабчарты

Если вы хотите передать values, доступные только в родительском чарте, в сабчарты, то вам поможет директива `export-values`, которая имитирует (с небольшими отличиями) поведение [import-values](https://helm.sh/docs/topics/charts/#importing-child-values-via-dependencies), только вместо передачи values из сабчарта в родительский чарт она делает обратное: передает values в сабчарт из родительского чарта. Пример использования:

```yaml
# .helm/Chart.yaml
apiVersion: v2
dependencies:
  - name: backend
    version: 1.0.0
    export-values:
    - parent: werf.image.backend
      child: backend.image
```

Таким образом мы передадим в сабчарт всё, что доступно в `.Values.werf.image.backend` родительского чарта. В нашем случае это будет строка с репозиторием, именем и тегом образа `backend`, описанного в `werf.yaml`, который может выглядеть так: `example.org/backend/<имя_тега>`. В сабчарте эта строка станет доступна через `.Values.backend.image`:

{% raw %}
```yaml
# .helm/charts/backend/app.yaml
...
spec:
  template:
    spec:
      containers:
      - name: backend
        image: {{ .Values.backend.image }}  # Ожидаемый результат: `image: example.org/backend:<имя_тега>`
```
{% endraw %}

В отличие от YAML-якорей `export-values` будет работать с динамически выставляемыми values (сервисные данные werf), с переданными через командную строку values (`--set` и пр.) и с секретными values.

Также доступна альтернативная укороченная форма `export-values`, которая работает только для словарей (maps):

```yaml
    export-values:
    - "someMap"
```

Это будет эквивалентно следующей полной форме `export-values`:

```yaml
    export-values:
    - parent: exports.somemap
      child: .
```

Так в корень values сабчарта будут экспортированы все ключи, найденные в словаре `.Values.exports.somemap`.

#### Устаревшие файлы requirements.yaml и requirements.lock

Устаревший формат описания зависимостей через файлы `.helm/requirements.yaml` и `.helm/requirements.lock` тоже поддерживается werf. Однако рекомендуется переходить на `.helm/Chart.yaml` и `.helm/Chart.lock`.

### Giterminism

werf реализует режим гитерминизма при чтении конфигурации helm. Все конфигурационные файлы и дополнительные файлы (указанные через параметры) должны быть коммитнуты в текущем коммите git репозитория проекта, потому что werf будет читать конфигурацию напрямую из git. Эти файлы включают в себя:

1. Все конфигурационные файлы чарта:

   ```
   .helm/
     Chart.yaml
     Chart.lock
     templates/
       <name>.yaml
       <name>.tpl
       <some_dir>/
         <name>.yaml
         <name>.tpl
     charts/
     secret/
     values.yaml
     secret-values.yaml
   ```

2. Любые дополнительные [файлы со значениями (values)]({{ "/advanced/helm/configuration/values.html#обычные-пользовательские-данные" | true_relative_url }}), указанные пользователем с помощью опций `--values`, `--secret-values`, `--set-file`.

**ЗАМЕЧАНИЕ** werf поддерживает режим разработки, который позволяет работать с файлами проекта без создания ненужных промежуточных коммитов в процессе отладки и разработки. Режим активируется опцией `--dev` (к примеру, `werf converge --dev`).

#### Сабчарты и гитерминизм

Для корректного использования сабчартов необходимо добавить файл [`.helm/Chart.lock`](https://helm.sh/docs/helm/helm_dependency/) и коммитнуть его в git репозиторий. werf автоматически скачает все зависимости указанные в lock-файле и загрузит файлы сабчартов корректно. Для стандартного кейса использования рекомендуется добавить директорию `.helm/charts` в `.gitignore`.

Имеется возможность явно добавить файлы сабчартов в директорию `.helm/charts/` и коммитнуть содержимое этой директории в git репозиторий. В таком случае werf скачает все сабчарты указанные в `.helm/Chart.lock` и загрузит чарты существующие в директории `.helm/charts` и совместит их. Чарты указанные в `.helm/charts` будут иметь приоритет в данном случае и должны переопределять чарты указанные в `.helm/Chart.lock`.

Больше информации по сабчартам в отдельной статье: [зависимости чартов]({{ "/advanced/helm/configuration/chart_dependencies.html" | true_relative_url }}).

### Секреты

Для хранения в репозитории паролей, файлов сертификатов и т.п., рекомендуется использовать подсистему работы с секретами werf.

Идея заключается в том, что конфиденциальные данные должны храниться в репозитории вместе с приложением, и должны оставаться независимыми от какого-либо конкретного сервера.

werf поддерживает указание секретов следующими способами:
- отдельный [values-файл для секретов]({{ "/advanced/helm/configuration/values.html#пользовательские-секреты" | true_relative_url }}) (`.helm/secret-values.yaml` по умолчанию или любой файл из репозитория, указанный опцией `--secret-values`).
- секретные файлы — закодированные файлы в сыром виде без yaml, могут быть использованы в шаблонах.

#### Ключ шифрования

Для шифрования и дешифрования данных необходим ключ шифрования. Есть три места откуда werf может прочитать этот ключ:
* из переменной окружения `WERF_SECRET_KEY`
* из специального файла `.werf_secret_key`, находящегося в корневой папке проекта
* из файла `~/.werf/global_secret_key` (глобальный ключ)

> Ключ шифрования должен иметь **шестнадцатеричный дамп** длиной 16, 24, или 32 байта для выбора соответственно алгоритмов AES-128, AES-192, или AES-256. Команда [werf helm secret generate-secret-key]({{ "reference/cli/werf_helm_secret_generate_secret_key.html" | true_relative_url }}) возвращает ключ шифрования, подходящий для использования алгоритма AES-128.

Вы можете быстро сгенерировать ключ, используя команду [werf helm secret generate-secret-key]({{ "reference/cli/werf_helm_secret_generate_secret_key.html" | true_relative_url }}).

> **ВНИМАНИЕ! Не сохраняйте файл `.werf_secret_key` в git-репозитории. Если вы это сделаете, то потеряете весь смысл шифрования, т.к. любой пользователь с доступом к git-репозиторию, сможет получить ключ шифрования. Поэтому, файл `.werf_secret_key` должен находиться в исключениях, т.е. в файле `.gitignore`!**

#### Ротация ключа шифрования

werf поддерживает специальную процедуру смены ключа шифрования с помощью команды [`werf helm secret rotate-secret-key`]({{ "reference/cli/werf_helm_secret_rotate_secret_key.html" | true_relative_url }}).

#### Secret values

Файлы с секретными переменными предназначены для хранения секретных данных в виде — `ключ: секрет`. **По умолчанию** werf использует для этого файл `.helm/secret-values.yaml`, но пользователь может указать любое число подобных файлов с помощью параметров запуска.

Файл с секретными переменными может выглядеть следующим образом:
```yaml
mysql:
  host: 10005968c24e593b9821eadd5ea1801eb6c9535bd2ba0f9bcfbcd647fddede9da0bf6e13de83eb80ebe3cad4
  user: 100016edd63bb1523366dc5fd971a23edae3e59885153ecb5ed89c3d31150349a4ff786760c886e5c0293990
  password: 10000ef541683fab215132687a63074796b3892d68000a33a4a3ddc673c3f4de81990ca654fca0130f17
  db: 1000db50be293432129acb741de54209a33bf479ae2e0f53462b5053c30da7584e31a589f5206cfa4a8e249d20
```

Для управления файлами с секретными переменными используйте следующие команды:
- [`werf helm secret values edit`]({{ "reference/cli/werf_helm_secret_values_edit.html" | true_relative_url }})
- [`werf helm secret values encrypt`]({{ "reference/cli/werf_helm_secret_values_encrypt.html" | true_relative_url }})
- [`werf helm secret values decrypt`]({{ "reference/cli/werf_helm_secret_values_decrypt.html" | true_relative_url }})

###### Использование в шаблонах чарта

Значения секретных переменных расшифровываются в процессе деплоя и используются в Helm в качестве [дополнительных значений](https://helm.sh/docs/chart_template_guide/values_files/). Таким образом, использование секретов не отличается от использования данных в обычном случае:

{% raw %}
```yaml
...
env:
- name: MYSQL_USER
  value: {{ .Values.mysql.user }}
- name: MYSQL_PASSWORD
  value: {{ .Values.mysql.password }}
```
{% endraw %}

#### Секретные файлы

Помимо использования секретов в переменных, в шаблонах также используются файлы, которые нельзя хранить незашифрованными в репозитории. Для размещения таких файлов выделен каталог `.helm/secret`, в котором должны храниться файлы с зашифрованным содержимым.

Чтобы использовать файлы содержащие секретную информацию в шаблонах Helm, вы должны сохранить их в соответствующем виде в каталоге `.helm/secret`.

Для управления файлами, содержащими секретную информацию, используйте следующие команды:
- [`werf helm secret file edit`]({{ "reference/cli/werf_helm_secret_file_edit.html" | true_relative_url }})
- [`werf helm secret file encrypt`]({{ "reference/cli/werf_helm_secret_file_encrypt.html" | true_relative_url }})
- [`werf helm secret file decrypt`]({{ "reference/cli/werf_helm_secret_file_decrypt.html" | true_relative_url }})

> **ПРИМЕЧАНИЕ** werf расшифрует содержимое всех файлов в `.helm/secret` перед рендерингом шаблонов helm-чарта. Необходимо удостовериться что директория `.helm/secret` содержит корректные зашифрованные файлы.

###### Использование в шаблонах чарта

<!-- Move to reference -->

Функция `werf_secret_file` позволяет использовать расшифрованное содержимое секретного файла в шаблоне. Обязательный аргумент функции путь к секретному файлу, относительно папки `.helm/secret`.

Пример использования секрета `.helm/secret/backend-saml/tls.key` в шаблоне:

{% raw %}
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: myproject-backend-saml
type: kubernetes.io/tls
data:
  tls.crt: {{ werf_secret_file "backend-saml/stage/tls.crt" | b64enc }}
  tls.key: {{ werf_secret_file "backend-saml/stage/tls.key" | b64enc }}
```
{% endraw %}

### Шаблоны

Определения ресурсов Kubernetes располагаются в директории `.helm/templates`.

В этой папке находятся YAML-файлы `*.yaml`, каждый из которых описывает один или несколько ресурсов Kubernetes, разделенных тремя дефисами `---`, например:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mydeploy
  labels:
    service: mydeploy
spec:
  selector:
    matchLabels:
      service: mydeploy
  template:
    metadata:
      labels:
        service: mydeploy
    spec:
      containers:
      - name: main
        image: ubuntu:18.04
        command: [ "/bin/bash", "-c", "while true; do date ; sleep 1 ; done" ]
---
apiVersion: v1
kind: ConfigMap
  metadata:
    name: mycm
  data:
    node.conf: |
      port 6379
      loglevel notice
```

Каждый YAML-файл предварительно обрабатывается как [Go-шаблон](https://golang.org/pkg/text/template/#hdr-Actions).

Использование Go-шаблонов дает следующие возможности:

* генерация Kubernetes-ресурсов, а также их составляющих в зависимости от произвольных условий;
* передача [values]({{ "/advanced/helm/configuration/values.html" | true_relative_url }}) в шаблон в зависимости от окружения;
* выделение общих частей шаблона в блоки и их переиспользование в нескольких местах;
* и т.д.

В дополнении к основным функциям Go-шаблонов также могут быть использоваться [функции Sprig](https://masterminds.github.io/sprig/) и [дополнительные функции](https://helm.sh/docs/howto/charts_tips_and_tricks/), такие как `include` и `required`.

Пользователь также может размещать `*.tpl` файлы, которые не будут рендериться в объект Kubernetes. Эти файлы могут быть использованы для хранения Go-шаблонов. Все шаблоны из `*.tpl` файлов доступны для использования в `*.yaml` файлах.

#### Интеграция с собранными образами

werf предоставляет набор сервисных значений, которые содержат маппинг `.Values.werf.image`. В этом маппинге по имени образа из `werf.yaml` содержится полное имя docker-образа. Полное описание сервисных значений werf доступно [в статье про values]({{ "/advanced/helm/configuration/values.html" | true_relative_url }}).

Как использовать образ по имени `backend` описанный в `werf.yaml`:

{% raw %}
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend
spec:
  template:
    spec:
      containers:
      - image: {{ .Values.werf.image.backend }}
```
{% endraw %}

Если в имени образа содержится дефис (`-`), то запись должна быть такого вида: {% raw %}`image: '{{ index .Values.werf.image "IMAGE-NAME" }}'`{% endraw %}.

#### Встроенные шаблоны и параметры

{% raw %}
* `{{ .Chart.Name }}` — возвращает имя проекта, указанное в `werf.yaml` (ключ `project`).
* `{{ .Release.Name }}` — {% endraw %}возвращает [имя релиза]({{ "/advanced/helm/releases/release.html" | true_relative_url }}).{% raw %}
* `{{ .Files.Get }}` — функция для получения содержимого файла в шаблон, требует указания пути к файлу в качестве аргумента. Путь указывается относительно папки `.helm` (файлы вне папки `.helm` недоступны).
  {% endraw %}

###### Окружение

Текущее окружение werf можно использовать в шаблонах.

Например, вы можете использовать его для создания разных шаблонов для разных окружений:

{% raw %}
```
apiVersion: v1
kind: Secret
metadata:
  name: regsecret
type: kubernetes.io/dockerconfigjson
data:
{{ if eq .Values.werf.env "dev" }}
  .dockerconfigjson: UmVhbGx5IHJlYWxseSByZWVlZWVlZWVlZWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWxsbGxsbGxsbGxsbGxsbGxsbGxsbGxsbGxsbGxsbGx5eXl5eXl5eXl5eXl5eXl5eXl5eSBsbGxsbGxsbGxsbGxsbG9vb29vb29vb29vb29vb29vb29vb29vb29vb25ubm5ubm5ubm5ubm5ubm5ubm5ubm5ubmdnZ2dnZ2dnZ2dnZ2dnZ2dnZ2cgYXV0aCBrZXlzCg==
{{ else }}
  .dockerconfigjson: {{ .Values.dockerconfigjson }}
{{ end }}
```
{% endraw %}

Следует обратить внимание, что значение параметра `--env ENV` доступно не только в шаблонах helm, но и [в шаблонах конфигурации `werf.yaml`]({{ "/reference/werf_yaml_template_engine.html#env" | true_relative_url }}).

Больше информации про сервисные значения доступно [в статье про values]({{ "/advanced/helm/configuration/values.html" | true_relative_url }}).

### Values

Под данными (или **values**) понимается произвольный YAML, заполненный парами ключ-значение или массивами, которые можно использовать в [шаблонах]({{ "/advanced/helm/configuration/templates.html" | true_relative_url }}). Все данные передаваемые в chart можно условно разбить на следующие категории:

- Обычные пользовательские данные.
- Пользовательские секреты.
- Сервисные данные.

#### Обычные пользовательские данные

Для хранения обычных данных используйте файл чарта `.helm/values.yaml` (необязательно). Пример структуры:

```yaml
global:
  names:
  - alpha
  - beta
  - gamma
  mysql:
    staging:
      user: mysql-staging
    production:
      user: mysql-production
    _default:
      user: mysql-dev
      password: mysql-dev
```

Данные, размещенные внутри ключа `global`, будут доступны как в текущем чарте, так и во всех [вложенных чартах]({{ "/advanced/helm/configuration/chart_dependencies.html" | true_relative_url }}) (сабчарты, subcharts).

Данные, размещенные внутри произвольного ключа `SOMEKEY` будут доступны в текущем чарте и во [вложенном чарте]({{ "/advanced/helm/configuration/chart_dependencies.html" | true_relative_url }}) с именем `SOMEKEY`.

Файл `.helm/values.yaml` — файл по умолчанию для хранения данных. Данные также могут передаваться следующими способами:

* С помощью параметра `--values=PATH_TO_FILE` может быть указан отдельный файл с данными (может быть указано несколько параметров, по одному для каждого файла данных).
* С помощью параметров `--set key1.key2.key3.array[0]=one`, `--set key1.key2.key3.array[1]=two` могут быть указаны непосредственно пары ключ-значение (может быть указано несколько параметров, смотри также `--set-string key=forced_string_value`).

**ЗАМЕЧАНИЕ.** Все values-файлы, включая `.helm/values.yaml` и любые другие файлы, указанные с помощью опций `--values` — все должны быть коммитнуты в git репозиторий проекта. Больше информации см. [в статье про гитерминизм]({{ "advanced/helm/configuration/giterminism.html" | true_relative_url }}).

###### Параметры set

Имеется возможность переопределить значения values и передать новые values через параметры командной строки:

- `--set KEY=VALUE`;
- `--set-string KEY=VALUE`;
- `--set-file=PATH`;
- `--set-docker-config-json-value=true|false`.

**ЗАМЕЧАНИЕ.** Все файлы, указанные опцией `--set-file` должны быть коммитнуты в git репозиторий проекта. Больше информации см. [в статье про гитерминизм]({{ "advanced/helm/configuration/giterminism.html" | true_relative_url }}).

######### set-docker-config-json-value

При использовании параметра `--set-docker-config-json-value` werf выставит специальный value `.Values.dockerconfigjson` взяв текущий docker config из окружения где запущен werf (поддерживается переменная окружения `DOCKER_CONFIG`).

В данном значении `.Values.dockerconfigjson` будет содержимое конфига docker закодированное в base64 и пригодное для использования например при создании секретов для доступа к container registry:

{% raw %}
```
{{- if .Values.dockerconfigjson -}}
apiVersion: v1
kind: Secret
metadata:
  name: regsecret
type: kubernetes.io/dockerconfigjson
data:
  .dockerconfigjson: {{ .Values.dockerconfigjson }}
{{- end -}}
```
{% endraw %}

**ВАЖНО!** Конфигурация docker текущего окружения, где запущен werf, может содержать доступы к registry, созданные с помощью короткоживущих токенов (например `CI_JOB_TOKEN` в GitLab). В таком случае использование `.Values.dockerconfigjson` для `imagePullSecrets` недопустимо, т.к. registry перестанет быть доступным из кластера Kubernetes как только завершится срок действия токена.

#### Пользовательские секреты

Секреты, предназначенные для хранения конфиденциальных данных (паролей, сертификатов и других чувствительных к утечке данных), удобны для хранения прямо в репозитории проекта.

Для хранения секретов может использоваться дефолтный файл чарта `.helm/secret-values.yaml` (необязательно) или любое количество файлов с произвольным именем (`--secret-values`). Пример структуры:

```yaml
global:
  mysql:
    production:
      password: 100024fe29e45bf00665d3399f7545f4af63f09cc39790c239e16b1d597842161123
    staging:
      password: 100024fe29e45bf00665d3399f7545f4af63f09cc39790c239e16b1d597842161123
```

Каждое значение в файле секретов (например, `100024fe29e45bf00665d3399f7545f4af63f09cc39790c239e16b1d597842161123`), представляет собой зашифрованные с помощью werf данные. Структура хранения секретов, такая же как и при хранении обычных данных, например, в `values.yaml`. Читайте подробнее о [генерации секретов и работе с ними]({{ "/advanced/helm/configuration/secrets.html" | true_relative_url }}) в соответствующей статье.

Файл `.helm/secret-values.yaml` — файл для хранения данных секретов по умолчанию. Данные также могут передаваться с помощью параметра `--secret-values=PATH_TO_FILE`, с помощью которого может быть указан отдельный файл с данными секретов (может быть указано несколько параметров, по одному для каждого файла данных секретов).

**ЗАМЕЧАНИЕ.** Все secret-values-файлы, включая `.helm/secret-values.yaml` и любые другие файлы, указанные с помощью опций `--secret-values` — все должны быть коммитнуты в git репозиторий проекта. Больше информации см. [в статье про гитерминизм]({{ "advanced/helm/configuration/giterminism.html" | true_relative_url }}).

#### Сервисные данные

Сервисные данные генерируются werf автоматически для передачи дополнительной информации при рендеринге шаблонов чарта.

Пример структуры и значений сервисных данных werf:

```yaml
werf:
  name: myapp
  namespace: myapp-production
  env: production
  repo: registry.domain.com/apps/myapp
  image:
    assets: registry.domain.com/apps/myapp:a243949601ddc3d4133c4d5269ba23ed58cb8b18bf2b64047f35abd2-1598024377816
    rails: registry.domain.com/apps/myapp:e760e9311f938e3d92681e93da3a81e176aa7f7e684ee06d092ec199-1598269478292
  tag:
    assets: a243949601ddc3d4133c4d5269ba23ed58cb8b18bf2b64047f35abd2-1598024377816
    rails: e760e9311f938e3d92681e93da3a81e176aa7f7e684ee06d092ec199-1598269478292
  commit:
    date:
      human: 2022-01-21 18:51:39 +0300 +0300
      unix: 1642780299
    hash: 1b28e6843a963c5bdb3579f6fc93317cc028051c

global:
  werf:
    name: myapp
    version: v1.2.7
```

Существуют следующие сервисные значения:

- Имя проекта из файла конфигурации `werf.yaml`: `.Values.werf.name`.
- Используемая версия werf: `.Values.werf.version`.
- Развертывание будет осуществлено в namespace `.Values.werf.namespace`.
- Название окружения CI/CD системы, используемое во время деплоя: `.Values.werf.env`.
- Адрес container registry репозитория, используемый во время деплоя: `.Values.werf.repo`.
- Полное имя и тег Docker-образа для каждого описанного в файле конфигурации `werf.yaml` образа: `.Values.werf.image.NAME`. Больше информации про использование этих значений доступно [в статье про шаблоны]({{ "/advanced/helm/configuration/templates.html#интеграция-с-собранными-образами" | true_relative_url }}).
- Только теги собранных Docker-образов. Предназначены в первую очередь для использования совместно с `.Values.werf.repo`, для проброса полного имени и тега образов по-отдельности.
- Информация о коммите, с которого werf был запущен: `.Values.werf.commit.hash`, `.Values.werf.commit.date.human`, `.Values.werf.commit.date.unix`.

###### Сервисные данные в сабчартах

Если вы используете [сабчарты]({{ "/advanced/helm/configuration/chart_dependencies.html" | true_relative_url }}) и хотите использовать неглобальные сервисные данные (`.Values.werf`) в сабчарте, то эти сервисные данные потребуется явно экспортировать в сабчарт из родительского чарта:

```yaml
# .helm/Chart.yaml
apiVersion: v2
dependencies:
  - name: rails
    version: 1.0.0
    export-values:
    - parent: werf
      child: werf
```

Теперь сервисные данные, изначально доступные только на `.Values.werf` в родительском чарте, стали доступными по тому же пути (`.Values.werf`) и в сабчарте "rails". Обращайтесь к сервисным данным из сабчарта таким образом:

{% raw %}
```yaml
# .helm/charts/rails/app.yaml
...
spec:
  template:
    spec:
      containers:
      - name: rails
        image: {{ .Values.werf.image.rails }}  # Ожидаемый результат: `image: registry.domain.com/apps/myapp/rails:e760e931...`
```
{% endraw %}

Путь, по которому сервисные данные будут доступны после экспорта, можно изменить:

```yaml
    export-values:
    - parent: werf
      child: definitely.not.werf  # Сервисные данные станут доступными на `.Values.definitely.not.werf` в сабчарте.
```

Также можно экспортировать сервисные значения в сабчарт по отдельности:

```yaml
# .helm/Chart.yaml
apiVersion: v2
dependencies:
  - name: postgresql
    version: "10.9.4"
    repository: "https://charts.bitnami.com/bitnami"
    export-values:
    - parent: werf.repo
      child: image.repository
    - parent: werf.tag.my-postgresql
      child: image.tag
```

Больше информации про `export-values` можно найти [здесь]({{ "advanced/helm/configuration/chart_dependencies.html#передача-динамических-values-из-родительского-чарта-в-сабчарты" | true_relative_url }}).

#### Итоговое объединение данных

<!-- This section could be in internals -->

Во время процесса деплоя werf объединяет все данные, включая секреты и сервисные данные, в единую структуру, которая передается на вход этапа рендеринга шаблонов (смотри подробнее [как использовать данные в шаблонах](#использование-данных-в-шаблонах)). Данные объединяются в следующем порядке (более свежее значение переопределяет предыдущее):

1. Данные из файла `.helm/values.yaml`.
2. Данные из параметров запуска `--values=PATH_TO_FILE`, в порядке указания параметров.
3. Данные секретов из файла `.helm/secret-values.yaml`.
4. Данные секретов из параметров запуска `--secret-values=PATH_TO_FILE`, в порядке указания параметров.
5. Данные из параметров `--set*`, указанные при вызове утилиты werf.
6. Сервисные данные.

#### Использование данных в шаблонах

Для доступа к данным в шаблонах чарта используется следующий синтаксис:

{% raw %}
```yaml
{{ .Values.key.key.arraykey[INDEX].key }}
```
{% endraw %}

Объект `.Values` содержит [итоговый набор объединенных значений](#итоговое-объединение-данных).

## Процесс выката

### Аннотации и лейблы для ресурсов чарта

#### Автоматические аннотации

werf автоматически выставляет следующие встроенные аннотации всем ресурсам чарта в процессе деплоя:

* `"werf.io/version": FULL_WERF_VERSION` — версия werf, использованная в процессе запуска команды `werf converge`;
* `"project.werf.io/name": PROJECT_NAME` — имя проекта, указанное в файле конфигурации `werf.yaml`;
* `"project.werf.io/env": ENV` — имя окружения, указанное с помощью параметра `--env` или переменной окружения `WERF_ENV` (не обязательно, аннотация не устанавливается, если окружение не было указано при запуске).

При использовании команды `werf ci-env` перед выполнением команды `werf converge`, werf также автоматически устанавливает аннотации содержащие информацию из используемой системы CI/CD (например, GitLab CI).
Например, [`project.werf.io/git`]({{ "internals/how_ci_cd_integration_works/gitlab_ci_cd.html#werf_add_annotation_project_git" | true_relative_url }}), [`ci.werf.io/commit`]({{ "internals/how_ci_cd_integration_works/gitlab_ci_cd.html#werf_add_annotation_ci_commit" | true_relative_url }}), [`gitlab.ci.werf.io/pipeline-url`]({{ "internals/how_ci_cd_integration_works/gitlab_ci_cd.html#werf_add_annotation_gitlab_ci_pipeline_url" | true_relative_url }}) и [`gitlab.ci.werf.io/job-url`]({{ "internals/how_ci_cd_integration_works/gitlab_ci_cd.html#werf_add_annotation_gitlab_ci_job_url" | true_relative_url }}).

Для более подробной информации об интеграции werf с системами CI/CD читайте статьи по темам:

* [Общие сведения по работе с CI/CD системами]({{ "internals/how_ci_cd_integration_works/general_overview.html" | true_relative_url }});
* [Работа GitLab CI]({{ "internals/how_ci_cd_integration_works/gitlab_ci_cd.html" | true_relative_url }}).

#### Пользовательские аннотации и лейблы

Пользователь может устанавливать произвольные аннотации и лейблы используя CLI-параметры при деплое `--add-annotation annoName=annoValue` (может быть указан несколько раз) и `--add-label labelName=labelValue` (может быть указан несколько раз).

Например, для установки аннотаций и лейблов `commit-sha=9aeee03d607c1eed133166159fbea3bad5365c57`, `gitlab-user-email=vasya@myproject.com` всем ресурсам Kubernetes в чарте, можно использовать следующий вызов команды деплоя:

```shell
werf converge \
  --add-annotation "commit-sha=9aeee03d607c1eed133166159fbea3bad5365c57" \
  --add-label "commit-sha=9aeee03d607c1eed133166159fbea3bad5365c57" \
  --add-annotation "gitlab-user-email=vasya@myproject.com" \
  --add-label "gitlab-user-email=vasya@myproject.com" \
  --env dev \
  --repo REPO
```

### Внешние зависимости

Чтобы сделать один ресурс релиза зависимым от другого, можно изменить его [порядок развертывания]({{ "/advanced/helm/deploy_process/deployment_order.html" | true_relative_url }}). После этого зависимый ресурс будет развертываться только после успешного развертывания основного. Но что делать, когда ресурс релиза должен зависеть от ресурса, который не является частью данного релиза или даже не управляется werf (например, создан неким оператором)?

В этом случае можно воспользоваться аннотацией [`<name>.external-dependency.werf.io/resource`]({{ "/reference/deploy_annotations.html#external-dependency-resource" | true_relative_url }}). Ресурс с данной аннотацией не будет развернут, пока не будет создана и готова заданная внешняя зависимость.

Пример:
```yaml
kind: Deployment
metadata:
  name: app
  annotations:
    secret.external-dependency.werf.io/resource: secret/dynamic-vault-secret
```

В приведенном выше примере werf дождется создания `dynamic-vault-secret`, прежде чем приступить к развертыванию `app`. Мы исходим из предположения, что `dynamic-vault-secret` создается оператором из инстанса Vault и не управляется werf.

Давайте рассмотрим еще один пример:
```yaml
kind: Deployment
metadata:
  name: service1
  annotations:
    service2.external-dependency.werf.io/resource: deployment/service2
    service2.external-dependency.werf.io/namespace: service2-production
```

В данном случае werf дождется успешного развертывания `service2` в другом пространстве имен, прежде чем приступить к развертыванию `service1`. Обратите внимание, что `service2` может развертываться либо как часть другого релиза werf, либо управляться иными инструментами в рамках CI/CD (т.е. находиться вне контроля werf).

Полезные ссылки:
* [`<name>.external-dependency.werf.io/resource`]({{ "/reference/deploy_annotations.html#external-dependency-resource" | true_relative_url }})
* [`<name>.external-dependency.werf.io/namespace`]({{ "/reference/deploy_annotations.html#external-dependency-namespace" | true_relative_url }})

### Helm хуки

Helm-хуки — произвольные ресурсы Kubernetes, помеченные специальной аннотацией `helm.sh/hook`. Например:

```yaml
kind: Job
metadata:
  name: somejob
  annotations:
    "helm.sh/hook": pre-upgrade,pre-install
    "helm.sh/hook-weight": "1"
```

Существует много разных helm-хуков, влияющих на процесс деплоя. Вы уже читали [ранее]({{ "/advanced/helm/deploy_process/steps.html" | true_relative_url }}) про `pre|post-install|upgade` хуки, используемые в процессе деплоя. Эти хуки наиболее часто используются для выполнения таких задач, как миграция (в хуках `pre-upgrade`) или выполнении некоторых действий после деплоя. Полный список доступных хуков можно найти в соответствующей документации [Helm](https://helm.sh/docs/topics/charts_hooks/).

Хуки сортируются в порядке возрастания согласно значению аннотации `helm.sh/hook-weight` (хуки с одинаковым весом сортируются по имени в алфавитном порядке), после чего хуки последовательно создаются и выполняются. werf пересоздает ресурс Kubernetes для каждого хука, в случае когда ресурс уже существует в кластере. Созданные ресурсы-хуки не удаляются после выполнения, если не указано [специальной аннотации `"helm.sh/hook-delete-policy": hook-succeeded,hook-failed`](https://helm.sh/docs/topics/charts_hooks/).

### Порядок развертывания

Все ресурсы по умолчанию применяются (apply) и отслеживаются одновременно, поскольку изначально у них одинаковый вес ([`werf.io/weight: 0`]({{ "/reference/deploy_annotations.html#resource-weight" | true_relative_url }})) (при условии, что вы его не меняли). Можно задать порядок применения (apply) и отслеживания ресурсов, установив для них различные веса.

Перед фазой развертывания werf группирует ресурсы на основе их веса, затем выполняет apply для группы с наименьшим ассоциированным весом и ждет, пока ресурсы из этой группы будут готовы. После успешного развертывания всех ресурсов из этой группы, werf переходит к развертыванию ресурсов из группы со следующим наименьшим весом. Процесс продолжается до тех пор, пока не будут развернуты все ресурсы релиза.

> Обратите внимание, что "werf.io/weight" работает только для ресурсов, не относящихся к хукам. Для хуков следует использовать "helm.sh/hook-weight".

Давайте рассмотрим следующий пример:
```yaml
kind: Job
metadata:
  name: db-migration
---
kind: StatefulSet
metadata:
  name: postgres
---
kind: Deployment
metadata:
  name: app
---
kind: Service
metadata:
  name: app
```

Все ресурсы из примера выше будут развертываться одновременно, поскольку у них одинаковый вес по умолчанию (0). Но что если базу данных необходимо развернуть до выполнения миграций, а приложение, в свою очередь, должно запуститься только после их завершения? Что ж, у этой проблемы есть решение! Попробуйте сделать так:
```yaml
kind: StatefulSet
metadata:
  name: postgres
  annotations:
    werf.io/weight: "-2"
---
kind: Job
metadata:
  name: db-migration
  annotations:
    werf.io/weight: "-1"
---
kind: Deployment
metadata:
  name: app
---
kind: Service
metadata:
  name: app
```

В приведенном выше примере werf сначала развернет базу данных и дождется ее готовности, затем запустит миграции и дождется их завершения, после чего развернет приложение и связанный с ним сервис.

Полезные ссылки:
* [`werf.io/weight`]({{ "/reference/deploy_annotations.html#resource-weight" | true_relative_url }})

### Принятие ресурсов в релиз

По умолчанию Helm и werf позволяют управлять лишь теми ресурсами, которые были созданы самим Helm или werf в рамках релиза. При попытке выката чарта с манифестом ресурса, который уже существует в кластере и который был создан не с помощью Helm или werf, тогда возникнет следующая ошибка:

```
Error: helm upgrade have failed: UPGRADE FAILED: rendered manifests contain a resource that already exists. Unable to continue with update: KIND NAME in namespace NAMESPACE exists and cannot be imported into the current release: invalid ownership metadata; label validation error: missing key "app.kubernetes.io/managed-by": must be set to "Helm"; annotation validation error: missing key "meta.helm.sh/release-name": must be set to RELEASE_NAME; annotation validation error: missing key "meta.helm.sh/release-namespace": must be set to RELEASE_NAMESPACE
```

Данная ошибка предотвращает деструктивное поведение, когда некоторый уже существующий релиз случайно становится частью выкаченного релиза.

Однако если данное поведение является желаемым, то требуется отредактировать целевой ресурс в кластере и добавить в него следующие аннотации и лейблы:

```yaml
metadata:
  annotations:
    meta.helm.sh/release-name: "RELEASE_NAME"
    meta.helm.sh/release-namespace: "NAMESPACE"
  labels:
    app.kubernetes.io/managed-by: Helm
```

После следующего деплоя этот ресурс будет принят в новую ревизию релиза и соотвественно будет приведен к тому состоянию, которое описано в манифесте чарта.

### Шаги

Во время запуска команды `werf converge` werf запускает процесс деплоя, включающий следующие этапы:

1. Преобразование шаблонов чартов в единый список манифестов ресурсов Kubernetes и их проверка.
2. Последовательный запуск [хуков]({{ "/advanced/helm/deploy_process/helm_hooks.html" | true_relative_url }}) `pre-install` или `pre-upgrade`, отсортированных по весу, и контроль каждого хука до завершения его работы с выводом логов.
3. Группировка ресурсов Kubernetes, не относящихся к хукам, по их [весу]({{ "/advanced/helm/deploy_process/deployment_order.html" | true_relative_url }}) и последовательное развертывание каждой группы в соответствии с ее весом: создание/обновление/удаление ресурсов и отслеживание их до готовности с выводом логов в процессе.
4. Запуск [хуков]({{ "/advanced/helm/deploy_process/helm_hooks.html" | true_relative_url }}) `post-install` или `post-upgrade` по аналогии с хуками `pre-install` и `pre-upgrade`.

#### Отслеживание ресурсов

werf использует библиотеку [kubedog](https://github.com/werf/kubedog) для отслеживания ресурсов. Отслеживание можно настроить для каждого ресурса, указывая соответствующие [аннотации]({{ "/reference/deploy_annotations.html" | true_relative_url }}) в шаблонах чартов.

#### Работа с несколькими кластерами Kubernetes

В некоторых случаях, необходима работа с несколькими кластерами Kubernetes для разных окружений. Все что вам нужно, это настроить необходимые [контексты](https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters) kubectl для доступа к необходимым кластерам и использовать для werf параметр `--kube-context=CONTEXT`, совместно с указанием окружения.

#### Сабчарты

Во время процесса деплоя werf выполнит рендер, создаст все требуемые ресурсы указанные во всех [используемых сабчартах]({{ "/advanced/helm/configuration/chart_dependencies.html" | true_relative_url }}) и будет отслеживать каждый из этих ресурсов до состояния готовности.

## Релизы

### Управление релизами

Следующие основные werf команды создают и удаляют релизы:

 - [`werf converge`]({{ "reference/cli/werf_converge.html" | true_relative_url }}) создаёт новую версию релиза для проекта;
 - [`werf dismiss`]({{ "reference/cli/werf_dismiss.html" | true_relative_url }}) удаляет все существующие версии релизов для проекта.

werf предоставляет следующие команды низкого уровня для управления релизами:

 - [`werf helm list -A`]({{ "reference/cli/werf_helm_list.html" | true_relative_url }}) — выводит список всех релизов всех namespace'ов кластера;
 - [`werf helm get all RELEASE`]({{ "reference/cli/werf_helm_get_all.html" | true_relative_url }}) — для получения информации по указанному релизу, манифестов, хуков и values, записанных в версию релиза;
 - [`werf helm status RELEASE`]({{ "reference/cli/werf_helm_status.html" | true_relative_url }}) — для получения статуса последней версии указанного релиза;
 - [`werf helm history RELEASE`]({{ "reference/cli/werf_helm_history.html" | true_relative_url }}) — для получения списка версий указанного релиза.

### Релиз

В то время как чарт — набор конфигурационных файлов вашего приложения, релиз (**release**) — это объект времени выполнения, экземпляр вашего приложения, развернутого с помощью werf.

#### Хранение релизов

Информация о каждой версии релиза хранится в самом кластере Kubernetes. werf поддерживает сохранение в произвольном namespace в объектах Secret или ConfigMap.

По умолчанию, werf хранит информацию о релизах в объектах Secret в целевом namespace, куда происходит деплой приложения. Это полностью совместимо с конфигурацией по умолчанию по хранению релизов в [Helm 3](https://helm.sh), что полностью совместимо с конфигурацией [Helm 2](https://helm.sh) по умолчанию. Место хранения информации о релизах может быть указано при деплое с помощью параметров werf: `--helm-release-storage-namespace=NS` и `--helm-release-storage-type=configmap|secret`.

Для получения информации обо всех созданных релизах можно использовать команду [werf helm list]({{ "reference/cli/werf_helm_list.html" | true_relative_url }}), а для просмотра истории конкретного релиза [werf helm history]({{ "reference/cli/werf_helm_history.html" | true_relative_url }}).

###### Замечание о совместимости с Helm

werf полностью совместим с уже установленным Helm 2, т.к. хранение информации о релизах осуществляется одним и тем же образом, как и в Helm. Если вы используете в Helm специфичное место хранения информации о релизах, а не значение по умолчанию, то вам нужно указывать место хранения с помощью опций werf `--helm-release-storage-namespace` и `--helm-release-storage-type`.

Информация о релизах, созданных с помощью werf, может быть получена с помощью Helm, например, командами `helm list` и `helm get`. С помощью werf также можно обновлять релизы, развернутые ранее с помощью Helm.

Команда [`werf helm list -A`]({{ "reference/cli/werf_helm_list.html" | true_relative_url }}) выводит список релизов созданных werf или Helm 3. Релизы, созданные через werf могут свободно просматриваться через утилиту helm командами `helm list` или `helm get` и другими.

###### Совместимость с Helm 2

Существующие релизы helm 2 (созданные например через werf v1.1) могут быть конвертированы в helm 3 либо автоматически во время работы команды [`werf converge`]({{ "/reference/cli/werf_converge.html" | true_relative_url }}), либо с помощью команды [`werf helm migrate2to3`]({{ "/reference/cli/werf_helm_migrate2to3.html" | true_relative_url }}).

### Именование

#### Окружение

По умолчанию, werf предполагает, что каждый релиз должен относиться к какому-либо окружению, например, `staging`, `test` или `production`.

На основании окружения werf определяет:

 1. Имя релиза.
 2. Namespace в Kubernetes.

Передача имени окружения является обязательной для операции деплоя и должна быть выполнена либо с помощью параметра `--env` либо на основании данных используемой CI/CD системы (читай подробнее про [интеграцию c CI/CD системами]({{ "internals/how_ci_cd_integration_works/general_overview.html#интеграция-с-настройками-cicd" | true_relative_url }})) определиться автоматически.

#### Имя релиза

По умолчанию название релиза формируется по шаблону `[[project]]-[[env]]`. Где `[[ project ]]` — имя [проекта]({{ "reference/werf_yaml.html#имя-проекта" | true_relative_url }}), а `[[ env ]]` — имя [окружения](#окружение).

Например, для проекта с именем `symfony-demo` будет сформировано следующее имя релиза в зависимости от имени окружения:
* `symfony-demo-stage` для окружения `stage`;
* `symfony-demo-test` для окружения `test`;
* `symfony-demo-prod` для окружения `prod`.

Имя релиза может быть переопределено с помощью параметра `--release NAME` при деплое. В этом случае werf будет использовать указанное имя как есть, без каких либо преобразований и использования шаблонов.

Имя релиза также можно явно определить в файле конфигурации `werf.yaml`, установив параметр [`deploy.helmRelease`]({{ "/reference/werf_yaml.html#имя-релиза" | true_relative_url }}).

###### Слагификация имени релиза

Сформированное по шаблону имя Helm-релиза слагифицируется, в результате чего получается уникальное имя Helm-релиза.

Слагификация имени Helm-релиза включена по умолчанию, но может быть отключена указанием параметра [`deploy.helmReleaseSlug=false`]({{ "/reference/werf_yaml.html#имя-релиза" | true_relative_url }}) в файле конфигурации `werf.yaml`.

#### Namespace в Kubernetes

По умолчанию namespace, используемый в Kubernetes, формируется по шаблону `[[ project ]]-[[ env ]]`, где `[[ project ]]` — [имя проекта]({{ "reference/werf_yaml.html#имя-проекта" | true_relative_url }}), а `[[ env ]]` — имя [окружения](#окружение).

Например, для проекта с именем `symfony-demo` будет сформировано следующее имя namespace в Kubernetes, в зависимости от имени окружения:
* `symfony-demo-stage` для окружения `stage`;
* `symfony-demo-test` для окружения `test`;
* `symfony-demo-prod` для окружения `prod`.

Имя namespace в Kubernetes может быть переопределено с помощью параметра `--namespace NAMESPACE` при деплое. В этом случае werf будет использовать указанное имя как есть, без каких либо преобразований и использования шаблонов.

Имя namespace также можно явно определить в файле конфигурации `werf.yaml`, установив параметр [`deploy.namespace`]({{ "/reference/werf_yaml.html#namespace-в-kubernetes" | true_relative_url }}).

###### Слагификация namespace Kubernetes

Сформированное по шаблону имя namespace слагифицируется, чтобы удовлетворять требованиям к [DNS именам](https://www.ietf.org/rfc/rfc1035.txt), в результате чего получается уникальное имя namespace в Kubernetes.

Слагификация имени namespace включена по умолчанию, но может быть отключена указанием параметра [`deploy.namespaceSlug=false`]({{ "/reference/werf_yaml.html#namespace-в-kubernetes" | true_relative_url }}) в файле конфигурации `werf.yaml`.

