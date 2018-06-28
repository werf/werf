---
title: Конфигурация выката
sidebar: doc_sidebar
permalink: templates_for_kube.html
folder: kube
---

Dapp при выкате вызывает [helm](https://helm.sh/), который использует чарт из папки `.helm` в корне проекта для конфигурации выката. Helm ищет YAML описания для объектов kubernetes в папке `templates` чарта и обрабатывает каждый файл с помощью механизма рендеринга шаблонов [GO](https://golang.org/). Dapp не обрабатывает чарты и шаблоны чарта сам, а использует для этого helm.

> Ссылки на дополнительную информацию:
- [Язык описания GO-шаблонов](https://godoc.org/text/template)
- [Справочник по Sprig](https://godoc.org/github.com/Masterminds/sprig) - библиотека, которую использует helm для рендеринга шаблонов GO
- [Дополнительные функции](https://docs.helm.sh/developing_charts/#chart-development-tips-and-tricks) добавленные в helm для шаблонов, такие как `include` и `required`

Создать структуру чарта можно с помощью команды `dapp kube chart create`, выполнив ее в корневой папке проекта. В результате выполнения будет создана папка `.helm` в которой будет создан чарт с примерами описания объектов kubernetes.

## Передача параметров

При выкате приложения, существует несколько способов передачи параметров:
- использование файлов `values.yaml`, `secret-values.yaml`. Оба варианта идентичны с точки зрения доступа, разница только в способе хранения - в файле `secret-values.yaml` значение переменных зашифрованы. Далее по тексту на эти различия не будет обращаться внимание;
- использование параметра --set в командах `dapp kube deploy` (см. раздел [Управление выкатом](deploy_for_kube.html));
- использование переменных окружения.

В файлах values.yaml и secret-values.yaml содержатся описание переменных, которые будут доступны в шаблонах. Например, есть такой values.yaml:

```yaml
replicas:
  production: 3
  staging: 2

db:
  host:
    production: prod-db.mycompany.int
    stage: stage-db
    _default: 192.168.6.3
  username:
    production: user-production
    stage: user-stage
  database:
    production: productiondb
    stage: stagedb
```

Тогда, обратиться к соответствующим переменным в шаблоне можно через конструкцию подобной - {% raw %}`{{ .Values.db.username.production }}`{% endraw %}.

Dapp устанавливает и использует ряд переменных, также доступных в шаблонах. Получить их значения можно с помощью команды `dapp kube value get VALUE_KEY`.

Например получить значения всех переменных можно следующим образом:
```
dapp kube value get .
```

## Особенности написания шаблонов чарта

Шаблоны чарта размещаются в директории `templates` в виде YAML-файлов.

Dapp предоставляет для использования следующие дополнительные шаблоны:
- `dapp_container_image`
- `dapp_container_env`

### Шаблон `dapp_container_image`

Пришел на смену используемому в старых версиях dapp шаблону `dimg`. Шаблон генерирует ключи `image` и `imagePullPolicy` для контейнера пода.

Особенностью шаблона является то, что `imagePullPolicy` генерируется в зависимости от значения `.Values.global.dapp.is_branch` - в случае использования тегов не будет ставиться `imagePullPolicy: Always`.

Шаблон может вернуть несколько строк, поэтому обязательно его использование с конструкцией indent.

Логика генерации ключа `imagePullPolicy`:
* Значение `.Values.global.dapp.is_branch=true` означает, что происходит деплой образа по логике "latest" для ветки.
    * В этом случае образ по соответствующему docker-тегу нужно обновлять через docker pull, даже если он уже существует, чтобы получить актуальную "latest" версию соответствущего тега.
    * В этом случае - `imagePullPolicy=Always`.
* Значение `.Values.global.dapp.is_branch=false` означает, что происходит деплой тега или отдельного коммита образа.
    * В этом случае образ по соответствующему docker-тегу не нужно обновлять через docker pull, если он уже существует.
    * В этом случае `imagePullPolicy` не указывается, что соответствует значению по умолчанию принятому в kubernetes на данный момент: `imagePullPolicy=IfNotPresent`.

Пример использования шаблона, при наличии нескольких dimg в dappfile:
* `tuple <dimg-name> . | include "dapp_container_image" | indent <N-spaces>`

Пример использования шаблона, при наличии только одного dimg без имени в dappfile:
* `tuple . | include "dapp_container_image" | indent <N-spaces>`
* `include "dapp_container_image" . | indent <N-spaces>`  (дополнительная упрощенная форма записи)

### Шаблон `dapp_container_env`

Позволяет оптимизировать работу выката, если образ не меняется. Генерирует блок с переменной окружения `DOCKER_IMAGE_ID` для контейнера пода, но только если `.Values.global.dapp.is_branch=true`, т.к. в этом случае образ по соответствующему docker-тегу мог обновится, но его имя осталось прежним. В переменной `DOCKER_IMAGE_ID` будет новый id docker образа, что заставит kubernetes обновить ресурс. Шаблон может вернуть несколько строк, поэтому обязательно его использование с indent.


Пример использования шаблона, при наличии нескольких dimg в dappfile:
* `tuple <dimg-name> . | include "dapp_container_env" | indent <N-spaces>`

Пример использования шаблона, при наличии только одного dimg без имени в dappfile:
* `tuple . | include "dapp_container_env" | indent <N-spaces>`
* `include "dapp_container_env" . | indent <N-spaces>`  (дополнительная упрощенная форма записи)

## Пример конфигурации
Пример описания конфигурации приложения, состоящего из контейнеров frontend, backend, db, демонстрирующий использование шаблонов dapp.

> В примере рассматривается только сутевая часть описания конфигурации. Для запуска приложения, в зависимости от конфигурации вашего кластера, дополнительно может потребоваться создание Ingress-ресурса, создание объекта Secret (если вы работаете с приватным репозиторием), создание объекта Service, обеспечение маршрутизации трафика и т.д.

Chart.yaml
```
name: example-dapp-deploy
version: 0.1.0
```

values.yaml
```
replicas:
  production: 3
  staging: 1
```

dappfile.yaml
```
dimg: "frontend"
from: "nginx"
---
dimg: "backend"
from: "alpine"
---
dimg: "db"
from: "mysql"
```

app.yaml
{% raw %}
```
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  name: {{ .Chart.Name }}-frontend
spec:
  replicas: {{ .Values.replicas.production }}
  template:
    spec:
      containers:
        - name: frontend
{{ tuple "frontend" . | include "dapp_container_image" | indent 10 }}
          env:
            - name: VAR1
              value: value
{{ tuple "frontend" . | include "dapp_container_env" | indent 12 }}
            - name: VAR2
              value: value
---
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  name: {{ .Chart.Name }}-backend
spec:
  template:
    spec:
      containers:
        - name: backend
{{ tuple "backend" . | include "dapp_container_image" | indent 10 }}
          env:
{{ tuple "backend" . | include "dapp_container_env" | indent 12 }}
---
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  name: {{ .Chart.Name }}-db
spec:
  template:
    spec:
      containers:
        - name: db
{{ tuple "db" . | include "dapp_container_image" | indent 10 }}
          env:
{{ tuple "db" . | include "dapp_container_env" | indent 12 }}
```
{% endraw %}

Результат:
```
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  name: example-dapp-deploy-frontend
spec:
  replicas: 3
  template:
    spec:
      containers:
        - name: frontend
          image: localhost:5000/example-dapp-deploy/frontend:latest
          imagePullPolicy: Always
          env:
            - name: VAR1
              value: value
            - name: DOCKER_IMAGE_ID
              value: sha256:7a126ea38f24d3ca98207d28414e4f6ae5ae30458539828a125d029dea8a93cb
            - name: VAR2
              value: value
---
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  name: example-dapp-deploy-backend
spec:
  template:
    spec:
      containers:
        - name: backend
          image: localhost:5000/example-dapp-deploy/backend:latest
          imagePullPolicy: Always
          env:
            - name: DOCKER_IMAGE_ID
              value: sha256:b325c0788c80efb7a08e6eb7c04e11e412a035c9d39a1430260d776421ea1a4a
---
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  name: example-dapp-deploy-db
spec:
  template:
    spec:
      containers:
        - name: db
          image: localhost:5000/example-dapp-deploy/db:latest
          imagePullPolicy: Always
          env:
            - name: DOCKER_IMAGE_ID
              value: sha256:da27cdffc4fcafaa4f6ced8b3bc1409191b9876b9b75c0e91ddffaceba5b497c

```

Упомянутое в конфигурации `.Chart.Name` — это значение ключа `name` из Chart.yaml, а значение `.Values.replicas.production` - взято из values.yaml.
