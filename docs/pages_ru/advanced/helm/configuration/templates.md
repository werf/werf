---
title: Шаблоны
permalink: advanced/helm/configuration/templates.html
---

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

## Интеграция с собранными образами

Ресурсы kubernetes требуют полное имя docker образа, включая Docker-репозиторий и Docker-тег, чтобы использовать Docker-образы в шаблонах чарта. Но как указать данные образа из файла конфигурации `werf.yaml` учитывая то, что полное имя Docker-образа зависит от указанного Docker-репозитория?

werf предоставляет набор сервисных значений, которые содержат маппинг `.Values.werf.image`. В этом маппинге по имени образа из `werf.yaml` содержится полное имя docker-образа. Полное описание сервисных значений werf доступно [в статье про values]({{ "/advanced/helm/configuration/values.html" | true_relative_url }}).

### .Values.werf.image

```
map[string]string

SHORT_IMAGE_NAME => FULL_DOCKER_IMAGE_NAME
```

Данный маппинг содержит соответствие имени образа из `werf.yaml` полному имени образа, которое может быть использовано в ключе `image` спецификации Pod'а.

##### Примеры

Чтобы получить полное имя docker-образа для образа `backend` из `werf.yaml`, используется:

```
.Values.werf.image.backend
```

Чтобы получить полное имя docker-образа для образа `nginx-assets` из `werf.yaml`, используется:

```
index .Values.werf.image "nginx-assets"
```

Как использовать образ по имени `backend` описанный в `werf.yaml`:

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
        image: {{ .Values.werf.image.backend }}
```
{% endraw %}

## Встроенные шаблоны и параметры

{% raw %}
 * `{{ .Chart.Name }}` — возвращает имя проекта, указанное в `werf.yaml` (ключ `project`).
 * `{{ .Release.Name }}` — {% endraw %}возвращает [имя релиза]({{ "/advanced/helm/releases/release.html" | true_relative_url }}).{% raw %}
 * `{{ .Files.Get }}` — функция для получения содержимого файла в шаблон, требует указания пути к файлу в качестве аргумента. Путь указывается относительно папки `.helm` (файлы вне папки `.helm` недоступны).
{% endraw %}

### Окружение

Основные команды werf (такие как [`werf converge`]({{ "reference/cli/werf_converge.html" | true_relative_url }}), [`werf build`]({{ "reference/cli/werf_build.html" | true_relative_url }}), [`werf render`]({{ "reference/cli/werf_render.html" | true_relative_url }}) и проч.) поддерживают параметр `--env ENV`.

Данный параметр может влиять на [имя релиза]({{ "/advanced/helm/releases/naming.html" | true_relative_url }}) и [kubernetes namespace]({{ "/advanced/helm/releases/naming.html" | true_relative_url }}). Также этот параметр можно получить в шаблонах следующей конструкцией:

{% raw %}
```
{{ .Values.werf.env }}
```
{% endraw %}

Допускается использование данного параметра для корректировки описания шаблонов ресурсов в зависимости от используемого окружения. Например:

{% raw %}
```
{{ if eq .Values.werf.env "dev" }}
apiVersion: v1
kind: Secret
metadata:
  name: regsecret
type: kubernetes.io/dockerconfigjson
data:
  .dockerconfigjson: UmVhbGx5IHJlYWxseSByZWVlZWVlZWVlZWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWxsbGxsbGxsbGxsbGxsbGxsbGxsbGxsbGxsbGxsbGx5eXl5eXl5eXl5eXl5eXl5eXl5eSBsbGxsbGxsbGxsbGxsbG9vb29vb29vb29vb29vb29vb29vb29vb29vb25ubm5ubm5ubm5ubm5ubm5ubm5ubm5ubmdnZ2dnZ2dnZ2dnZ2dnZ2dnZ2cgYXV0aCBrZXlzCg==
---
{{ else }}
apiVersion: v1
kind: Secret
metadata:
  name: regsecret
type: kubernetes.io/dockerconfigjson
data:
  .dockerconfigjson: {{ .Values.dockerconfigjson }}
---
{{ end }}
```
{% endraw %}

Следует обратить внимание, что значение параметра `--env ENV` доступно не только в шаблонах helm, но и [в шаблонах конфигурации `werf.yaml`]({{ "/reference/werf_yaml_template_engine.html#env" | true_relative_url }}).

Больше информации про сервисные значения доступно [в статье про values]({{ "/advanced/helm/configuration/values.html" | true_relative_url }}).
