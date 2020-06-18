---
title: Генерируем и раздаем ассеты
sidebar: applications-guide
permalink: documentation/guides/applications-guide/template/040-assets.html
layout: guide
toc: false
---

{% filesused title="Файлы, упомянутые в главе" %}
- .helm/templates/service.yaml
- .helm/templates/ingress.yaml
- werf.yaml
{% endfilesused %}

В какой-то момент в процессе развития вашего базового приложения вам понадобятся ассеты (т.е. картинки, css, js).

Для того, чтобы обработать ассеты мы воспользуемся ____________. Он соединяет и сжимает ассеты JavaScript и CSS, добавляет возможность писать ассеты на других языках и с использованием препроцессоров, таких как CoffeeScript, Sass и ____________.

Для генерации ассетов мы будем использовать команду `____________`.

Интуитивно понятно, что на одной из стадии сборки нам надо будет вызвать скрипт, который генерирует файлы, т.е. что-то надо будет дописать в `werf.yaml`. Однако, не только там — ведь какое-то приложение в production должно непосредственно отдавать статические файлы. Мы не будем отдавать файлы с помощью ____________ — хочется, чтобы статику раздавал nginx. А значит надо будет внести какие-то изменения и в helm чарт.

Реализовать раздачу сгенерированных ассетов можно сделать двумя способами:

* Добавить в собираемый образ с ____________ ещё и nginx, а потом этот образ запускать уже двумя разными способами: один раз для раздачи статики, второй — для работы ____________-приложения
* Сделать два отдельных образа: в одном только nginx и сгенерированные ассеты, во втором — ____________-приложение 

{% offtopic title="Как правильно сделать выбор?" %}
TODO: дать небольшое рассуждение как правильно делать выбор. Что надо опираться на то, что будет и как часто перекатываться. И может быть постараться разобрать два разных кейса.
{% endofftopic %}

Мы сделаем два отдельных образа.

## Подготовка к внесению изменений

Перед тем, как вносить изменения — **необходимо убедиться, что в собранных ассетах нет привязки к конкретному окружению**. То есть в собранных не должно быть логинов, паролей, доменов и тому подобного. В момент сборки ____________ не должен подключаться к базе данных, использовать user-generated контент и тому подобное.

## Изменения в сборке

Для ассетов мы соберём отдельный образ с nginx и ассетами. Для этого нужно собрать образ с nginx и забросить туда предварительно собранные с помощью [механизма артефактов](https://ru.werf.io/documentation/configuration/stapel_artifact.html) ассеты.

{% offtopic title="Что за артефакты?" %}
[Артефакт](https://ru.werf.io/documentation/configuration/stapel_artifact.html) — это специальный образ, используемый в других артефактах или отдельных образах, описанных в конфигурации. Артефакт предназначен преимущественно для отделения ресурсов инструментов сборки от процесса сборки образа приложения. Примерами таких ресурсов могут быть — программное обеспечение или данные, которые необходимы для сборки, но не нужны для запуска приложения, и т.п.

TODO: наглядно и просто на простом примере показать, как данные из артефакта забрасываем в образ 

TODO: кратко рассказать, как отлаживать разработку артефактов: что можно просто переименовать директиву artifact в image и получить образ, который можно уже отладить.
{% endofftopic %}

Начнём с создания артефакта: установим необходимые пакеты и выполним сборку ассетов. Генерация ассетов должна происходить в артефакте на стадии `setup`.

{% snippetcut name="werf.yaml" url="template-files/examples/example_1/werf.yaml#L21" %}
{% raw %}
```yaml
artifact: build
from: ____________
ansible:
  ____________
  ____________
  ____________
```
{% endraw %}
{% endsnippetcut %}

Теперь, когда артефакт собран, соберём образ с nginx:

{% snippetcut name="werf.yaml" url="template-files/examples/example_1/werf.yaml#L21" %}
{% raw %}
```yaml
image: assets
from: nginx:alpine
ansible:
  beforeInstall:
  - name: Add nginx config
    copy:
      content: |
{{ .Files.Get ".werf/nginx.conf" | indent 8 }}
      dest: /etc/nginx/nginx.conf
```
{% endraw %}
{% endsnippetcut %}

И пропишем в нём импорт из артефакта под названием `build`.

{% snippetcut name="werf.yaml" url="template-files/examples/example_2/werf.yaml#21" %}
```yaml
import:
- artifact: build
  add: /app/public
  to: /www
  after: setup
```
{% endsnippetcut %}

## Изменения в деплое и роутинге

Внутри Deployment сделаем два контейнера: один с `nginx`, который будет раздавать статические файлы, второй — с ____________ приложением. Запросы сперва будут приходить на nginx, а тот будет перенаправлять запрос приложению, если не найдётся статических файлов.

 Обязательно укажем `livenessProbe` и `readinessProbe`, которые будут проверять корректную работу контейнера в Pod-е, а также `preStop` команду для корректного завершение процесса nginx, чтобы при выкате новой версии приложения корректно завершались активные сессии.

{% snippetcut name=".helm/templates/deployment.yaml" url="template-files/examples/example_2/.helm/templates/deployment.yaml#L33" %}
{% raw %}
```yaml
      - name: assets
{{ tuple "assets" . | include "werf_container_image" | indent 8 }}
        lifecycle:
          preStop:
            exec:
              command: ["/usr/sbin/nginx", "-s", "quit"]
        livenessProbe:
          httpGet:
            path: /healthz
            port: 80
            scheme: HTTP
        readinessProbe:
          httpGet:
            path: /healthz
            port: 80
            scheme: HTTP
        ports:
        - containerPort: 80
          name: http
          protocol: TCP
```
{% endraw %}
{% endsnippetcut %}

В описании Service так же должен быть указан правильный порт:

{% snippetcut name=".helm/templates/service.yaml" url="template-files/examples/example_2/.helm/templates/service.yaml#L9" %}
```yaml
  ports:
  - name: http
    port: 80
    protocol: TCP
```
{% endsnippetcut %}

Также необходимо отправить запросы на правильный порт, чтобы они попадали на nginx.

{% snippetcut name=".helm/templates/ingress.yaml" url="template-files/examples/example_2/.helm/templates/ingress.yaml" %}
```yaml
      paths:
      - path: /
        backend:
          serviceName: {{ .Chart.Name }}
          servicePort: 80
```
{% endsnippetcut %}

{% offtopic title="А можно ли разделять трафик на уровне ingress?" %}

В некоторых случаях нужно разделить трафик на уровне ingress. В таком случае можно разделить запросы по path и портам в объекте Ingress:

{% snippetcut name=".helm/templates/ingress.yaml" url="template-files/examples/example_2/.helm/templates/ingress.yaml#L9" %}
{% raw %}
```yaml
      paths:
      - path: /
        backend:
          serviceName: {{ .Chart.Name }}
          servicePort: 3000
      - path: /assets
        backend:
          serviceName: {{ .Chart.Name }}
          servicePort: 80
```
{% endraw %}
{% endsnippetcut %}

TODO: в идеале — описать, что за случаи такие

{% endofftopic %}

<div>
    <a href="050-files.html" class="nav-btn">Далее: Работа с файлами</a>
</div>

