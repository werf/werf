---
title: Несколько приложений в одном репозитории
sidebar: applications-guide
permalink: documentation/guides/applications-guide/gitlab-nodejs/110-multipleapps.html
layout: guide
toc: false
---

{% filesused title="Файлы, упомянутые в главе" %}
- .helm/templates/deployment-frontend.yaml
- .helm/templates/deployment-backend.yaml
- .helm/templates/service-backend.yaml
- .helm/templates/ingress.yaml
- werf.yaml
{% endfilesused %}


Если в одном репозитории находятся несколько приложений например для backend и frontend необходимо использовать сборку приложения с несколькими образами.

Мы рассказывали [https://www.youtube.com/watch?v=g9cgppj0gKQ](https://www.youtube.com/watch?v=g9cgppj0gKQ) о том, почему и в каких ситуациях это — хороший путь для микросервисов.

Покажем это на примере нашего приложения на node которое отправляет в rabbitmq имена тех кто зашёл в наш чат, а приложение на Python будет работать в качестве бота который приветсвует всех зашедших.

## Сборка приложений

Сборка приложения с несколькими образами описана в [статье](https://ru.werf.io/documentation/guides/advanced_build/multi_images.html). На ее основе покажем наш пример для нашего приложения.

Структура каталогов будет организована следующим образом

```
├── .helm
│   ├── templates
│   └── values.yaml
├── node
├── python
└── werf.yaml
```
Мы должны переместить весь наш исходный код Nodejs в отдельную директорию `node` и
далее собирать её как обычно, но заменив в `werf.yaml` строки клонирования кода с корня `- add: /` на нашу директорию `- add: /node`.
Сборка приложения для python структурно практически не отличается от сборки описанной в Hello World, за исключением того что импорт из git будет из директории python.

Пример конечного [werf.yaml](/werf-articles/gitlab-nodejs-files/examples/werf_3.yaml)

Для запуска подготовленных приложений отдельными деплойментами, необходимо создать еще один файл, который будет описывать запуск бота - [bot-deployment.yaml](/werf-articles/gitlab-nodejs-files/examples/python-deployment.yaml).

<details><summary>bot-deployment.yaml</summary>
<p>

```yaml
{{ $rmq_user := pluck .Values.global.env .Values.app.rabbitmq.user | first | default .Values.app.rabbitmq.user._default }}
{{ $rmq_pass := pluck .Values.global.env .Values.app.rabbitmq.password | first | default .Values.app.rabbitmq.password._default }}
{{ $rmq_host := pluck .Values.global.env .Values.app.rabbitmq.host | first | default .Values.app.rabbitmq.host._default }}
{{ $rmq_port := pluck .Values.global.env .Values.app.rabbitmq.port | first | default .Values.app.rabbitmq.port._default }}
{{ $rmq_vhost := pluck .Values.global.env .Values.app.rabbitmq.vhost | first | default .Values.app.rabbitmq.vhost._default }}
{{ $db_user := pluck .Values.global.env .Values.app.postgresql.user | first | default .Values.app.postgresql.user._default }}
{{ $db_pass := pluck .Values.global.env .Values.app.postgresql.password | first | default .Values.app.postgresql.password._default }}
{{ $db_host := pluck .Values.global.env .Values.app.postgresql.host | first | default .Values.app.postgresql.host._default }}
{{ $db_port := pluck .Values.global.env .Values.app.postgresql.port | first | default .Values.app.postgresql.port._default }}
{{ $db_name := pluck .Values.global.env .Values.app.postgresql.db | first | default .Values.app.postgresql.db._default }}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .Chart.Name }}-email-consumer
spec:
  revisionHistoryLimit: 3
  strategy:
    type: RollingUpdate
  replicas: 1
  selector:
    matchLabels:
      app: {{ $.Chart.Name }}-email-consumer
  template:
    metadata:
      labels:
        app: {{ $.Chart.Name }}-email-consumer
    spec:
      imagePullSecrets:
      - name: "registrysecret"
      containers:
      - name: {{ $.Chart.Name }}
{{ tuple "python" . | include "werf_container_image" | indent 8 }}
        workingDir: /app
        command: ["python","consumer.py"]
        ports:
        - containerPort: 5000
          protocol: TCP
        env:
        - name: AMQP_URI
          value: {{ printf "amqp://%s:%s@%s:%s/%s" $rmq_user $rmq_pass $rmq_host ($rmq_port|toString) $rmq_vhost }}
        - name: DATABASE_URL
          value: {{ printf "postgresql+psycorg2://%s:%s@%s:%s/%s" $db_user $db_pass $db_host ($db_port|toString) $db_name }}
{{ tuple "python" . | include "werf_container_env" | indent 8 }}
```

</p>
</details>

Обратите внимание что мы отделили переменные от деплоймента, мы сделали это специально, чтобы строка подключения к базам в переменных манифеста не была огромной. Потому мы выделили переменные и описали их вверху манифеста, а затем просто подставили в нужные нам строки.

Таким образом мы смогли собрать и запустить несколько приложений написанных на разных языках которые находятся в одном репозитории.

Если в вашей команды фуллстэки и/или она не очень большая и хочется видеть и выкатывать приложение целиком, может быть полезно разместить приложения на нескольких языках в одном репозитории.

К слову, мы рассказывали [https://www.youtube.com/watch?v=g9cgppj0gKQ](https://www.youtube.com/watch?v=g9cgppj0gKQ) о том, почему и в каких ситуациях это — хороший путь для микросервисов.


<div>
    <a href="120-dynamicenvs.html" class="nav-btn">Далее: Динамические окружения</a>
</div>
