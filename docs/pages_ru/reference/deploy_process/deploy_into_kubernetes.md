---
title: Деплой в Kubernetes
sidebar: documentation
permalink: documentation/reference/deploy_process/deploy_into_kubernetes.html
author: Timofey Kirillov <timofey.kirillov@flant.com>
---

werf дает совместимую альтернативу [Helm 2](https://helm.sh), но предлагает улучшенный процесс деплоя.

Для работы Kubernetes у werf есть 2 основные команды: [deploy]({{ site.baseurl }}/documentation/cli/main/deploy.html) — для установки или обновления приложения в кластере, и [dismiss]({{ site.baseurl }}/documentation/cli/main/dismiss.html) — для удаления приложения из кластера.

В werf есть несколько настраиваемых режимов отслеживания развернутых ресурсов, в том числе с возможностью отслеживания журналов и событий. Образы, собранные werf легко интегрируются в [шаблоны](#шаблоны) helm-чартов. werf может устанавливать аннотации и метки с произвольной информацией всем разворачиваемым в Kubernetes ресурсам проекта.

Конфигурация описывается в формате аналогичном формату [Helm-чарта](#чарт).

## Чарт

Чарт — набор конфигурационных файлов описывающих приложение. Файлы чарта находятся в папке `.helm`, в корневой папке проекта:

```
.helm/
  templates/
    <name>.yaml
    <name>.tpl
  charts/
  secret/
  values.yaml
  secret-values.yaml
```

### Шаблоны

Шаблоны находятся в папке `.helm/templates`.

В этой папке находятся YAML-файлы `*.yaml`, каждый из который описывает один или несколько ресурсов Kubernetes, разделенных тремя дефисами `---`, например:

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
 * генерирование разных спецификаций объекта Kubernetes в зависимости от какого-либо условия;
 * передача [данных](#данные) в шаблон, в зависимости от окружения;
 * выделение общих частей шаблона в блоки и их переиспользование в нескольких местах;
 * и т.д..

В шаблонах чарта также могут быть использованы [функции Sprig](https://masterminds.github.io/sprig/) и [дополнительные функции](https://docs.helm.sh/developing_charts/#chart-development-tips-and-tricks), такие как `include` и `required`.

Пользователь также может размещать `*.tpl` файлы, которые не будут рендериться в объект Kubernetes. Эти файлы могут быть использованы для хранения произвольных Go-шаблонов и выражений. Все шаблоны и выражения из `*.tpl` файлов доступны для использования в `*.yaml` файлах.

#### Интеграция с собранными образами

Чтобы использовать Docker-образы в шаблонах чарта, необходимо указать полное имя Docker-образа, включая Docker-репозиторий и Docker-тег. Но как указать данные образа из файла конфигурации `werf.yaml` учитывая то, что полное имя Docker-образа зависит от выбранной стратегии тегирования и указанного Docker-репозитория?

Второй вопрос, — как использовать параметр [`imagePullPolicy`](https://kubernetes.io/docs/concepts/containers/images/#updating-images) вместе с образом из  `werf.yaml`: указывать `imagePullPolicy: Always`? А как добиться скачивания образа только когда это действительно необходимо?

Для ответа на эти вопросы в werf есть две функции: [`werf_container_image`](#werf_container_image) и [`werf_container_env`](#werf_container_env). Пользователь может использовать эти функции в шаблонах чарта для корректного и безопасного указания образов описанных в файле конфигурации `werf.yaml`.

##### werf_container_image

Данная функция генерирует ключи `image` и `imagePullPolicy` со значениями, необходимыми для соответствующего контейнера пода.

Особенность функции в том, что значение `imagePullPolicy` формируется исходя из значения `.Values.global.werf.is_branch`. Если не используется тег, то функция возвращает `imagePullPolicy: Always`, иначе (если используется тег) — ключ `imagePullPolicy` не возвращается. В результате, образ будет всегда скачиваться если он был собран для git-ветки, т.к. у Docker-образа с тем же именем мог измениться ID.

Функция может возвращать несколько строк, поэтому она должна использоваться совместно с конструкцией `indent`.

Логика генерации ключа `imagePullPolicy`:
* Значение `.Values.global.werf.is_branch=true` подразумевает, что развертывается образ для git-ветки, с расчетом на использование самого свежего образа.
  * В этом случае, образ с соответствующим тегом должен быть принудительно скачан, даже если он уже есть в локальном хранилище Docker-образов. Это необходимо, чтобы получить самый "свежий" образ, соответствующий образу с таким Docker-тегом.
  * В этом случае – `imagePullPolicy=Always`.
* Значение `.Values.global.werf.is_branch=false` подразумевает, что развертывается образ для git-тега или конкретного git-коммита.
  * В этом случае, образ для соответствующего Docker-тега можно не обновлять, если он уже находится в локальном хранилище Docker-образов.
  * В этом случае, `imagePullPolicy` не устанавливается, т.е. итоговое значение у объекта в кластере будет соответствовать значению по умолчанию — `imagePullPolicy=IfNotPresent`.

> Образы, протегированные с использованием пользовательской стратегии тегирования (`--tag-custom`) обрабатываются аналогично образам протегированным стратегией тегирования *git-branch* (`--tag-git-branch`)

Пример использования функции в случае нескольких описанных в файле конфигурации `werf.yaml` образов:
* `tuple <image-name> . | werf_container_image | indent <N-spaces>`

Пример использования функции в случае описанного в файле конфигурации `werf.yaml` безымянного образа:
* `tuple . | werf_container_image | indent <N-spaces>`
* `werf_container_image . | indent <N-spaces>` (дополнительный упрощенный формат использования)

##### werf_container_env

Позволяет упростить процесс релиза, в случае если образ остается неизменным. Возвращает блок с переменной окружения `DOCKER_IMAGE_ID` контейнера пода. Значение переменной будет установлено только если `.Values.global.werf.is_branch=true`, т.к. в этом случае Docker-образ для соответствующего имени и тега может быть обновлен, а имя и тег останутся неизменными. Значение переменной `DOCKER_IMAGE_ID` содержит новый ID Docker-образа, что вынуждает Kubernetes обновить объект.

Функция может возвращать несколько строк, поэтому она должна использоваться совместно с конструкцией `indent`.

> Образы, протегированные с использованием пользовательской стратегии тегирования (`--tag-custom`) обрабатываются аналогично образам протегированным стратегией тегирования *git-branch* (`--tag-git-branch`)

Пример использования функции в случае нескольких описанных в файле конфигурации `werf.yaml` образов:
* `tuple <image-name> . | werf_container_env | indent <N-spaces>`

Пример использования функции в случае описанного в файле конфигурации `werf.yaml` безымянного образа:
* `tuple . | werf_container_env | indent <N-spaces>`
* `werf_container_env . | indent <N-spaces>` (дополнительный упрощенный формат использования)

##### Примеры

Пример использования образа `backend`, описанного в `werf.yaml`:

{% raw %}
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend
  labels:
    service: backend
spec:
  selector:
    matchLabels:
      service: backend
  template:
    metadata:
      labels:
        service: backend
    spec:
      containers:
      - name: main
        command: [ ... ]
{{ tuple "backend" . | werf_container_image | indent 8 }}
        env:
{{ tuple "backend" . | werf_container_env | indent 8 }}
```
{% endraw %}

Пример использования безымянного образа, описанного в `werf.yaml`:

{% raw %}
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend
  labels:
    service: backend
spec:
  selector:
    matchLabels:
      service: backend
  template:
    metadata:
      labels:
        service: backend
    spec:
      containers:
      - name: main
        command: [ ... ]
{{ werf_container_image . | indent 8 }}
        env:
{{ werf_container_env . | indent 8 }}
```
{% endraw %}

#### Файлы секретов

Файлы секретов удобны для хранения непосредственно в репозитории проекта конфиденциальных данных, таких как сертификаты и закрытые ключи.

Файлы секретов размещаются в папке `.helm/secret`, где пользователь может создать произвольную структуру файлов. Читайте подробнее о том, как шифровать файлы в соответствующей [статье]({{ site.baseurl }}/documentation/reference/deploy_process/working_with_secrets.html#шифрование-файлов-секретов).

##### werf_secret_file

`werf_secret_file` — это функция используемая в шаблонах чартов, предназначенная для удобной работы с секретами, — она возвращает содержимое файла секрета.
Обычно она используется при формировании манифестов секретов в Kubernetes (`Kind: Secret`).
Функции в качестве аргумента необходимо передать путь к файлу относительно папки `.helm/secret`.

Пример использования расшифрованного содержимого файлов `.helm/secret/backend-saml/stage/tls.key` и `.helm/secret/backend-saml/stage/tls.crt` в шаблоне:

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

Обратите внимание, что `backend-saml/stage/` — произвольная структура файлов, и пользователь может либо размещать все файлы в одной папке `.helm/secret`, либо создавать структуру со своему усмотрению.

#### Встроенные шаблоны и параметры

{% raw %}
 * `{{ .Chart.Name }}` — возвращает имя проекта, указанное в `werf.yaml` (ключ `project`).
 * `{{ .Release.Name }}` — возвращает [имя релиза](#релиз).
 * `{{ .Files.Get }}` — функция для получения содержимого файла в шаблон, требует указания пути к файлу в качестве аргумента. Путь указывается относительно папки `.helm` (файлы вне папки `.helm` недоступны).
{% endraw %}

### Данные

Под данными понимается произвольная YAML-карта, заполненная парами ключ-значение или массивами, которые можно использовать в [шаблонах](#шаблоны).

werf позволяет использовать следующие типы данных:

 * Обычные пользовательские данные
 * Пользовательские секреты
 * Сервисные данные

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

Данные, размещенные внутри ключа `global`, будут доступны как в текущем чарте, так и во всех [вложенных чартах]({{ site.baseurl }}/documentation/reference/deploy_process/working_with_chart_dependencies.html) (сабчарты, subcharts).

Данные, размещенные внутри произвольного ключа `SOMEKEY` будут доступны в текущем чарте и во [вложенном чарте]({{ site.baseurl }}/documentation/reference/deploy_process/working_with_chart_dependencies.html) с именем `SOMEKEY`.

Файл `.helm/values.yaml` — файл по умолчанию для хранения данных. Данные также могут передаваться следующими способами:

 * С помощью параметра `--values=PATH_TO_FILE` может быть указан отдельный файл с данными (может быть указано несколько параметров, по одному для каждого файла данных).
 * С помощью параметров `--set key1.key2.key3.array[0]=one`, `--set key1.key2.key3.array[1]=two` могут быть указаны непосредственно пары ключ-значение (может быть указано несколько параметров, смотри также `--set-string key=forced_string_value`).

#### Пользовательские секреты

Секреты, предназначенные для хранения конфиденциальных данных, удобны для хранения прямо в репозитории проекта паролей, сертификатов и других чувствительных к утечке данных.

Для хранения данных секретов, используйте файл чарта `.helm/secret-values.yaml` (необязательно). Пример структуры:

```yaml
global:
  mysql:
    production:
      password: 100024fe29e45bf00665d3399f7545f4af63f09cc39790c239e16b1d597842161123
    staging:
      password: 100024fe29e45bf00665d3399f7545f4af63f09cc39790c239e16b1d597842161123
```

Каждое значение в файле секретов похожее на это — `100024fe29e45bf00665d3399f7545f4af63f09cc39790c239e16b1d597842161123`, представляет собой какие-то зашифрованные с помощью werf данные. Структура хранения секретов, такая-же как и при хранении обычных данных, например, в `values.yaml`. Читайте подробнее о [генерации секретов и работе с ними]({{ site.baseurl }}/documentation/reference/deploy_process/working_with_secrets.html#шифрование-секретов) в соответствующей статье.

Файл `.helm/secret-values.yaml` — файл по умолчанию для хранения данных секретов. Данные также могут передаваться с помощью параметра `--secret-values=PATH_TO_FILE`, с помощью которого может быть указан отдельный файл с данными секретов (может быть указано несколько параметров, по одному для каждого файла данных секретов).

#### Сервисные данные

Сервисные данные генерируются werf автоматически для передачи дополнительной информации при рендеринге шаблонов чарта.

Пример структуры и значений сервисных данных werf:

```yaml
global:
  env: stage
  namespace: myapp-stage
  werf:
    ci:
      branch: mybranch
      is_branch: true
      is_tag: false
      tag: '"-"'
    docker_tag: mybranch
    image:
      assets:
        docker_image: registry.domain.com/apps/myapp/assets:mybranch
        docker_image_id: sha256:ddaec322ee2c622aa0591177062a81009d9e52785be6915c5a37e822c2019755
        docker_image_digest: sha256:81009d9e52785be6915c5a37e822c2019755ddaec322ee2c622aa0591177062a
      rails:
        docker_image: registry.domain.com/apps/myapp/rails:mybranch
        docker_image_id: sha256:646c56c828beaf26e67e84a46bcdb6ab555c6bce8ebeb066b79a9075d0e87f50
        docker_image_digest: sha256:555c6bce8ebeb066b79a9075d0e87f50646c56c828beaf26e67e84a46bcdb6ab
    is_nameless_image: false
    name: myapp
    repo: registry.domain.com/apps/myapp
```

Существуют следующие сервисные данные:
 * Название окружения CI/CD системы, используемое во время деплоя: `.Values.global.env`.
 * Namespace Kubernetes используемый во время деплоя: `.Values.global.namespace`.
 * Имя используемой git-ветки или git-тега: `.Values.global.werf.ci.is_branch`, `.Values.global.werf.ci.branch`, `.Values.global.werf.ci.is_tag`, `.Values.global.werf.ci.tag`.
 * `.Values.global.ci.ref` — содержит либо название git-ветки либо название git-тега.
 * Полное имя Docker-образа и его ID, для каждого описанного в файле конфигурации `werf.yaml` образа: `.Values.global.werf.image.IMAGE_NAME.docker_image` и `.Values.global.werf.image.IMAGE_NAME.docker_image_id` и `.Values.global.werf.image.IMAGE_NAME.docker_image_digest`.
 * `.Values.global.werf.is_nameless_image` — устанавливается если в файле конфигурации `werf.yaml` описан безымянный образ.
 * Имя проекта из файла конфигурации `werf.yaml`: `.Values.global.werf.name`.
 * Docker-тег, используемый при деплое образа, описанного в файле конфигурации `werf.yaml` (соответственно выбранной стратегии тегирования): `.Values.global.werf.docker_tag`.
 * Docker-репозиторий образа используемый при деплое: `.Values.global.werf.repo`.

#### Итоговое объединение данных

Во время процесса деплоя werf объединяет все данные, включая секреты и сервисные данные, в единую структуру, которая передается на вход этапа рендеринга шаблонов (смотри подробнее [как использовать данные в шаблонах](#использование-данных-в-шаблонах)). Данные объединяются в следующем порядке приоритета (последующее значение переопределяет предыдущее):

 1. Данные из файла `.helm/values.yaml`.
 2. Данные из параметров запуска `--values=PATH_TO_FILE`, в порядке указания параметров.
 3. Данные секретов из файла `.helm/secret-values.yaml`.
 4. Данные секретов из параметров запуска `--secret-values=PATH_TO_FILE`, в порядке указания параметров.
 5. Сервисные данные.

### Использование данных в шаблонах

Для доступа к данным в шаблонах чарта используется следующий синтаксис:

{% raw %}
```yaml
{{ .Values.key.key.arraykey[INDEX].key }}
```
{% endraw %}

Объект `.Values` содержит [итоговый набор объединенных значений](#итоговое-объединение-данных).

## Релиз

В то время как чарт — набор конфигурационных файлов вашего приложения, релиз (release) — это объект времени выполнения, экземпляр вашего приложения, развернутого с помощью werf.

У каждого релиза есть одно имя и несколько версий. При каждом деплое с помощью werf создается новая версия релиза.

### Хранение релизов

Информация о каждой версии релиза хранится в самом кластере Kubernetes. werf может хранить ее в объектах ConfigMap или Secret, в любых namespace.

По умолчанию, werf хранит информацию о релизах в объектах ConfigMap в namespace `kube-system`, что полностью совместимо с конфигурацией [Helm 2](https://helm.sh) по умолчанию. Место хранения информации о релизах может быть указано при деплое с помощью параметров werf: `--helm-release-storage-namespace=NS` и `--helm-release-storage-type=configmap|secret`.

Для получения информации обо всех созданных релизах, нужно использовать команду: `kubectl -n kube-system get cm`. Имена объектов ConfigMap, содержащих информацию о релизах, имеют следующий шаблон имени — `RELEASE_NAME.RELEASE_VERSION`. Наибольший номер `RELEASE_VERSION` соответствует последней развернутой версии. В ConfigMap, содержащих информацию о релизах, также есть метки (labels) по которым можно получить информацию о статусе релиза:

```yaml
kind: ConfigMap
metadata:
  ...
  labels:
    MODIFIED_AT: "1562938540"
    NAME: werfio-test
    OWNER: TILLER
    STATUS: DEPLOYED
    VERSION: "165"
```

**ЗАМЕЧАНИЕ:** Изменение статуса релиза в метках ConfigMap не повлияет на реальный статус релиза, так как метки содержат информацию только для справочных целей и поиска/фильтрации объектов. Реальное состояние релиза хранится в ключе `data` ConfigMap.

#### Замечание о совместимости с Helm

werf полностью совместим с уже установленным Helm 2, т.к. хранение информации о релизах осуществляется одинаковым образом как и в Helm, и в одном и том же месте. Если вы используете в Helm специфичное место хранения информации о релизах, а не значение по умолчанию, то вам нужно указывать место хранения с помощью опций werf `--helm-release-storage-namespace` и `--helm-release-storage-type`.

Информация о релизах, созданных с помощью werf, может быть получена с помощью Helm, например, командами `helm list` и `helm get`. С помощью werf также можно обновлять релизы, развернутые ранее с помощью Helm.

Более того, вы можете работать в одном кластере Kubernetes одновременно и с werf и с Helm 2.

### Окружение

По умолчанию, werf предполагает что каждый релиз должен относиться к какому-либо окружению, например, `staging`, `test` или `production`.

На основании окружения werf определяет:

 1. Имя релиза.
 2. Namespace в Kubernetes.

Передача имени окружения является обязательной для операции деплоя, и должна быть выполнена либо с помощью параметра `--env` werf либо автоматически определяться на основании данных используемой CI/CD системы (читай подробнее про [интеграцию c CI/CD системами]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html#интеграция-с-настройками-ci-cd)).

### Имя релиза

По умолчанию название релиза формируется по шаблону `[[project]]-[[env]]`. Где `[[ project ]]` — имя [проекта]({{ site.baseurl }}/documentation/configuration/introduction.html#имя-проекта), а `[[ env ]]` — имя [окружения](#окружение).

Например, для проекта с именем `symfony-demo` будет сформировано следующее имя релиза в зависимости от имени окружения:
* `symfony-demo-stage` для окружения `stage`;
* `symfony-demo-test` для окружения `test`;
* `symfony-demo-prod` для окружения `prod`.

Имя релиза может быть переопределено с помощью параметра `--release NAME` при деплое. В этом случае werf будет использовать указанное имя как есть, без каких либо преобразований и использования шаблонов.

Имя релиза также можно явно определить в файле конфигурации `werf.yaml`, установив параметр [`deploy.helmRelease`]({{ site.baseurl }}/documentation/configuration/deploy_into_kubernetes.html#имя-релиза).

#### Слагификация имени релиза

Сформированное по шаблону имя Helm-релиза [слагифицируется]({{ site.baseurl }}/documentation/reference/toolbox/slug.html#базовый-алгоритм), в результате чего получается уникальное имя Helm-релиза.

Слагификация имени Helm-релиза включена по умолчанию, но может быть отключена указанием параметра [`deploy.helmReleaseSlug=false`]({{ site.baseurl }}/documentation/configuration/deploy_into_kubernetes.html#имя-релиза) в файле конфигурации `werf.yaml`.

### Kubernetes namespace

По умолчанию namespace, используемый в Kubernetes, формируется по шаблону `[[ project ]]-[[ env ]]`, где `[[ project ]]` — [имя проекта]({{ site.baseurl }}/documentation/configuration/introduction.html#meta-configuration-doc), а `[[ env ]]` — имя [окружения](#окружения).

Например, для проекта с именем `symfony-demo` будет сформировано следующее имя namespace в Kubernetes, в зависимости от имени окружения:
* `symfony-demo-stage` для окружения `stage`;
* `symfony-demo-test` для окружения `test`;
* `symfony-demo-prod` для окружения `prod`.

Имя namespace в Kubernetes может быть переопределено с помощью параметра `--namespace NAMESPACE` при деплое. В этом случае werf будет использовать указанное имя как есть, без каких либо преобразований и использования шаблонов.

Имя namespace также можно явно определить в файле конфигурации `werf.yaml`, установив параметр [`deploy.namespace`]({{ site.baseurl }}/documentation/configuration/deploy_into_kubernetes.html#kubernetes-namespace).

#### Слагификация namespace Kubernetes

Сформированное по шаблону имя namespace [слагифицируется]({{ site.baseurl }}/documentation/reference/toolbox/slug.html#базовый-алгоритм) чтобы удовлетворять требованиям к [DNS именам](https://www.ietf.org/rfc/rfc1035.txt), в результате чего получается уникальное имя namespace в Kubernetes.

Слагификация имени namespace включена по умолчанию, но может быть отключена указанием параметра [`deploy.namespaceSlug=false`]({{ site.baseurl }}/documentation/configuration/deploy_into_kubernetes.html#kubernetes-namespace) в файле конфигурации `werf.yaml`.

## Процесс деплоя

Во время запуска команды `werf deploy` werf запускает процесс деплоя, включающий следующие этапы:

 1. Рендеринг шаблонов чартов в единый список манифестов объектов Kubernetes и их проверка.
 2. Запуск [хуков](#helm-hooks) `pre-install` или `pre-upgrade`, отслеживание их работы вплоть до успешного или неуспешного завершения, вывод логов и другой информации.
 3. Применение изменений к ресурсам Kubernetes: создание новых, удаление старых, обновление существующих.
 4. Создание новых версий релизов и сохранение состояния манифестов ресурсов в данные этого релиза.
 5. Отслеживание всех ресурсов релиза (для тех, у кого есть пробы, — до готовности readiness-проб), вывод их логов и другой информации.
 6. Запуск [хуков](#helm-hooks) `post-install` или `post-upgrade`, отслеживание их работы вплоть до успешного или неуспешного завершения, вывод логов и другой информации.

**ЗАМЕЧАНИЕ:** werf удалит все созданные им при деплое ресурсы сразу же во время процесса деплоя, если он завершится неудачей на любом из указанных выше этапе!

Во время выполнения Helm-хуков на шагах 2 и 6 werf будет отслеживать ресурсы хуков до их успешного завершения. Отслеживание может быть [настроено](#настройка-отслеживания-ресурсов) для каждого из хуков ресурсов.

На шаге 5, werf будет отслеживать ресурсы релиза то их перехода в статус Ready. Все ресурсы отслеживаются одновременно, результат отслеживания всех ресурсов релиза выводится комбинированно с периодическим выводом т.н. таблицы прогресса. Отслеживание может быть [настроено](#настройка-отслеживания-ресурсов) для каждого ресурса.

werf отслеживает и выводит логи подов Kubernetes только до перехода их в статус "Ready", но не включая задания (ресурсы с `Kind: Job`). Для подов заданий, логи выводятся до момента завершения работы соответствущих подов.

С точки зрения реализации, для отслеживания ресурсов используется библиотека [kubedog](https://github.com/flant/kubedog).
В настоящий момент отслеживание ресурсов поддерживается для следующих типов: Deployment, StatefulSet, DaemonSet и Job.
В [ближайшее время](https://github.com/flant/werf/issues/1637) планируется реализация поддержки отслеживания ресурсов с типом Service, Ingress, PVC и других.

### Методы применения изменений

werf пытается применять трехсторонний метод обновления ресурсов Kubernetes, т.к. это наилучший вариант применения изменений. Существуют, также, и другие методы обновления ресурсов.

Читай более подробнее об этом в соответствующих статьях:
 - [Методы обновления ресурсов и применения изменений]({{ site.baseurl }}/documentation/reference/deploy_process/resources_update_methods_and_adoption.html);
 - [Сравнение методов обновления ресурсов с Helm]({{ site.baseurl }}/documentation/reference/deploy_process/differences_with_helm.html#трехстороннее-слияние-и-применение-изменений);
 - [Статья на Хабр "3-way merge в werf: деплой в Kubernetes с Helm «на стероидах»"](https://habr.com/ru/company/flant/blog/476646/).

### Если деплой завершился неудачно

В режиме двухстороннего слияния (2-way-merge), в случае ошибки во время деплоя, werf создает новый релиз со статусом `FAILED`. Далее, этот релиз может быть проанализирован пользователем для поиска и устранения проблем при следующем деплое.

При следующем деплое, werf выполнит откат релиза до последней успешной версии. Во время отката все ресурсы релиза будут восстановлены до состояния последней успешной версии, которое было до применением новых изменений в манифестах ресурсов.

В режиме трехстороннего слияния (3-way-merge) откат релиза не требуется, т.к. работает другой механизм, читай про него подробнее в соответствующей [статье](https://{{site.baseurl}}/documentation/reference/deploy_process/resources_update_methods_and_adoption.html).

### Helm-хуки

Helm-хуки — произвольный ресурс Kubernetes, помеченный специальной аннотацией `helm.sh/hook`. Например:

```yaml
kind: Job
metadata:
  name: somejob
  annotations:
    "helm.sh/hook": pre-upgrade,pre-install
    "helm.sh/hook-weight": "1"
```

Существует много разных helm-хуков, влияющих на процесс деплоя. Вы уже читали [выше](#процесс-деплоя) про `pre|post-install|upgade` хуки, используемые в процессе деплоя. Эти хуки наиболее часто используются для выполнения таких задач, как миграция (в хуках `pre-upgrade`) или выполнении некоторых действий после деплоя. Полный список доступных хуков можно найти в соответствующей документации [Helm](https://helm.sh/docs/topics/charts_hooks/).

Хуки сортируются в порядке возрастания согласно значению аннотации `helm.sh/hook-weight` (хуки с одинаковым весом сортируются по имени в алфавитном порядке), после чего хуки последовательно создаются и выполняются. werf пересоздает ресурс Kubernetes для каждого хука, в случае когда ресурс уже существует в кластере. Созданные хуки ресурсов не удаляются после выполнения.

### Настройка отслеживания ресурсов

Отслеживание ресурсов может быть настроено для каждого ресурса с помощью его аннотаций:

 * [`werf.io/track-termination-mode`](#track-termination-mode);
 * [`werf.io/fail-mode`](#fail-mode);
 * [`werf.io/failures-allowed-per-replica`](#failures-allowed-per-replica);
 * [`werf.io/log-regex`](#log-regex);
 * [`werf.io/log-regex-for-CONTAINER_NAME`](#log-regex-for-container);
 * [`werf.io/skip-logs`](#skip-logs);
 * [`werf.io/skip-logs-for-containers`](#skip-logs-for-containers);
 * [`werf.io/show-logs-only-for-containers`](#show-logs-only-for-containers);
 * [`werf.io/show-service-messages`](#show-service-messages).

Все приведенные аннотации могут использоваться совместно в одном ресурсе.

**СОВЕТ** Используйте аннотации — `"werf.io/track-termination-mode": NonBlocking` и `"werf.io/fail-mode": IgnoreAndContinueDeployProcess`, когда описываете в релизе объект Job, который должен быть запущен в фоне и который не влияет на процесс деплоя.

**СОВЕТ** Используйте аннотацию `"werf.io/track-termination-mode": NonBlocking`, когда описываете в релизе объект StatefulSet с ручной стратегией выката (параметр `OnDelete`), и не хотите блокировать весь процесс деплоя из-за этого объекта, дожидаясь его обновления.

#### Примеры использования аннотаций

<div class="tabs">
  <a href="javascript:void(0)" class="tabs__btn active" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'show-service-messages')">show-service-messages</a>
  <a href="javascript:void(0)" class="tabs__btn" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'skip-logs')">skip-logs</a>
  <a href="javascript:void(0)" class="tabs__btn" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'track-termination-mode')">NonBlocking track-termination-mode</a>
</div>

<div id="show-service-messages" class="tabs__content active">
  <img src="https://raw.githubusercontent.com/flant/werf-demos/master/deploy/werf-new-track-modes-1.gif" />
</div>
<div id="skip-logs" class="tabs__content">
  <img src="https://raw.githubusercontent.com/flant/werf-demos/master/deploy/werf-new-track-modes-2.gif" />
</div>

<div id="track-termination-mode" class="tabs__content">
  <img src="https://raw.githubusercontent.com/flant/werf-demos/master/deploy/werf-new-track-modes-3.gif" />
</div>

#### Track termination mode

`"werf.io/track-termination-mode": WaitUntilResourceReady|NonBlocking`

 * `WaitUntilResourceReady` (по умолчанию) — весь процесс деплоя будет отслеживать и ожидать готовности ресурса с данной аннотацией. Т.к. данный режим включен по умолчанию, то, по умолчанию, процесс деплоя ждет готовности всех ресурсов.
 * `NonBlocking` — ресурс с данной аннотацией отслеживается только пока есть другие ресурсы, готовности которых ожидает процесс деплоя.

#### Fail mode

`"werf.io/fail-mode": FailWholeDeployProcessImmediately|HopeUntilEndOfDeployProcess|IgnoreAndContinueDeployProcess`

 * `FailWholeDeployProcessImmediately` (по умолчанию) — в случае ошибки при деплое ресурса с данной аннотацией, весь процесс деплоя будет завершен с ошибкой.
 * `HopeUntilEndOfDeployProcess` — в случае ошибки при деплое ресурса с данной аннотацией его отслеживание будет продолжаться, пока есть другие ресурсы, готовности которых ожидает процесс деплоя, либо все оставшиеся ресурсы имеют такую-же аннотацию. Если с ошибкой остался только этот ресурс или несколько ресурсов с такой-же аннотацией, то в случае сохранения ошибки весь процесс деплоя завершается с ошибкой.
 * `IgnoreAndContinueDeployProcess` — ошибка при деплое ресурса с данной аннотацией не влияет на весь процесс деплоя.

#### Failures allowed per replica

`"werf.io/failures-allowed-per-replica": DIGIT`

По умолчанию, при отслеживании статуса ресурса допускается срабатывание ошибки 1 раз, прежде чем весь процесс деплоя считается ошибочным. Этот параметр влияет на поведение настройки [Fail mode](#fail-mode): определяет порог срабатывания, после которого начинает работать режим реакции на ошибки.

#### Log regex

`"werf.io/log-regex": RE2_REGEX`

Определяет [Re2 regex](https://github.com/google/re2/wiki/Syntax) шаблон, применяемый ко всем логам всех контейнеров всех подов ресурса с этой аннотацией. werf будет выводить только те строки лога, которые удовлетворяют regex-шаблону. По умолчанию werf выводит все строки лога.

#### Log regex for container

`"werf.io/log-regex-for-CONTAINER_NAME": RE2_REGEX`

Определяет [Re2 regex](https://github.com/google/re2/wiki/Syntax) шаблон, применяемый к логам контейнера с именем `CONTAINER_NAME` всех подов с данной аннотацией. werf будет выводить только те строки лога, которые удовлетворяют regex-шаблону. По умолчанию werf выводит все строки лога.

#### Skip logs

`"werf.io/skip-logs": true|false`

Если установлена в `true`, то логи всех контейнеров пода с данной аннотацией не выводятся при отслеживании. Отключено по умолчанию.

#### Skip logs for containers

`"werf.io/skip-logs-for-containers": CONTAINER_NAME1,CONTAINER_NAME2,CONTAINER_NAME3...`

Список (через запятую) контейнеров пода с данной аннотацией, для которых логи не выводятся при отслеживании.

#### Show logs only for containers

`"werf.io/show-logs-only-for-containers": CONTAINER_NAME1,CONTAINER_NAME2,CONTAINER_NAME3...`

Список (через запятую) контейнеров пода с данной аннотацией, для которых выводятся логи при отслеживании. Для контейнеров, чьи имена отсутствуют в списке, логи не выводятся. По умолчанию выводятся логи для всех контейнеров всех подов ресурса.

#### Show service messages

`"werf.io/show-service-messages": true|false`

Если установлена в `true`, то при отслеживании для ресурсов будет выводиться дополнительная отладочная информация, такая как события Kubernetes. По умолчанию, werf выводит такую отладочную информацию только в случае если ошибка ресурса приводит к ошибке всего процесса деплоя.

### Аннотации и метки ресурсов чарта

#### Автоматические аннотации

werf автоматически устанавливает следующие встроенные аннотации всем ресурсам чарта в процессе деплоя:

 * `"werf.io/version": FULL_WERF_VERSION` — версия werf, использованная в процессе запуска команды `werf deploy`;
 * `"project.werf.io/name": PROJECT_NAME` — имя проекта, указанное в файле конфигурации `werf.yaml`;
 * `"project.werf.io/env": ENV` — имя окружения, указанное с помощью параметра `--env` или переменной окружения `WERF_ENV` (не обязательно, аннотация не устанавливается если окружение не было указано при запуске).

При использовании команды `werf ci-env` перед выполнением команды `werf deploy`, werf также автоматически устанавливает аннотации содержащие информацию из используемой системы CI/CD (например, GitLab CI).
Например, [`project.werf.io/git`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/gitlab_ci.html#werf_add_annotation_project_git), [`ci.werf.io/commit`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/gitlab_ci.html#werf_add_annotation_ci_commit), [`gitlab.ci.werf.io/pipeline-url`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/gitlab_ci.html#werf_add_annotation_gitlab_ci_pipeline_url) и [`gitlab.ci.werf.io/job-url`]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/gitlab_ci.html#werf_add_annotation_gitlab_ci_job_url).

Для более подробной информации об интеграции werf с системами CI/CD читайте статьи по темам:

 * [Общие сведения по работе с CI/CD системами]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html);
 * [Работа GitLab CI]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/gitlab_ci.html).

#### Пользовательские аннотации и метки

Пользователь может устанавливать произвольные аннотации и метки используя CLI-параметры при деплое `--add-annotation annoName=annoValue` (может быть указан несколько раз) и `--add-label labelName=labelValue` (может быть указан несколько раз).

Например, для установки аннотаций и меток `commit-sha=9aeee03d607c1eed133166159fbea3bad5365c57`, `gitlab-user-email=vasya@myproject.com` всем ресурсам Kubernetes в чарте, можно использовать  следующий вызов команды деплоя:

```shell
werf deploy \
  --add-annotation "commit-sha=9aeee03d607c1eed133166159fbea3bad5365c57" \
  --add-label "commit-sha=9aeee03d607c1eed133166159fbea3bad5365c57" \
  --add-annotation "gitlab-user-email=vasya@myproject.com" \
  --add-label "gitlab-user-email=vasya@myproject.com" \
  --env dev \
  --images-repo :minikube \
  --stages-storage :local
```

### Проверка манифестов ресурсов

Если манифест ресурса в чарте содержит логические или синтаксические ошибки, то werf выведет соответствующее предупреждение во время процесса деплоя. Также, все ошибки проверки манифеста заносятся в аннотацию `debug.werf.io/validation-messages`. Такие ошибки обычно не влияют прямо на процесс деплоя и его статус выполнения, т.к. apiserver Kubernetes может принимать манифесты содержащие опечатки или ошибки, не выдавая какого-либо предупреждения.

Например, допустим имеем следующую опечатку в шаблоне чарта (`envs` вместо `env`, и `redinessProbe` вместо `readinessProbe`):

```
containers:
- name: main
  command: [ "/bin/bash", "-c", "while true; do date ; sleep 1 ; done" ]
  image: ubuntu:18.04
  redinessProbe:
    tcpSocket:
      port: 8080
    initialDelaySeconds: 5
    periodSeconds: 10
envs:
- name: MYVAR
  value: myvalue
```

Результат проверки манифеста будет примерно следующим:

```
│   WARNING ### Following problems detected during deploy process ###
│   WARNING Validation of target data failed: deployment/mydeploy1: [ValidationError(Deployment.spec.template.spec.containers[0]): unknown field               ↵
│ "redinessProbe" in io.k8s.api.core.v1.Container, ValidationError(Deployment.spec.template.spec): unknown field "envs" in io.k8s.api.core.v1.PodSpec]
```

В результате ресурс будет иметь аннотацию `debug.werf.io/validation-messages` следующего содержания:

```
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    debug.werf.io/validation-messages: 'Validation of target data failed: deployment/mydeploy1:
      [ValidationError(Deployment.spec.template.spec.containers[0]): unknown field
      "redinessProbe" in io.k8s.api.core.v1.Container, ValidationError(Deployment.spec.template.spec):
      unknown field "envs" in io.k8s.api.core.v1.PodSpec]'
...
```

## Работа с несколькими кластерами Kubernetes

В некоторых случаях, необходима работа с несколькими кластерами Kubernetes для разных окружений. Все что вам нужно, это настроить необходимые [контексты](https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters) kubectl для доступа к необходимым кластерам, и использовать для werf при деплое параметр `--kube-context=CONTEXT`, совместно с указанием окружения.
