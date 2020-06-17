---
title: Генерируем и раздаем ассеты
sidebar: applications-guide
permalink: documentation/guides/applications-guide/template/040-assets.html
layout: guide
---

{% filesused title="Файлы, упомянутые в главе" %}
- .helm/templates/service.yaml
- .helm/templates/ingress.yaml
- werf.yaml
{% endfilesused %}




В какой-то момент в процессе разработки вам понадобятся ассеты (т.е. картинки, css, js).

Для генерации ассетов мы будем использовать webpack.

Интуитивно понятно, что на стадии сборки нам надо будет вызвать скрипт, который генерирует файлы, т.е. что-то надо будет дописать в `werf.yaml`. Однако, не только там — ведь какое-то приложение в production должно непосредственно отдавать статические файлы. Мы не будем отдавать файлики с помощью Express Nodejs. Хочется, чтобы статику раздавал nginx. А значит надо будет внести какие-то изменения и в helm чарты.

<a name="assets-scenario" />

## Сценарий сборки ассетов

Webpack собирает ассеты используя конфигурацию указанную в webpack.config.js

Запуск генерации с webpack вставляют сразу в package.json как скрипт:



```json
  "scripts": {
    "test": "echo \"Error: no test specified\" && exit 1",
    "build": "webpack --mode production",
    "watch": "webpack --mode development --watch",
    "start": "webpack-dev-server --mode development --open",
    "clear": "del-cli dist",
    "migrate": "node-pg-migrate"
  },
```

Обычно используется несколько режимов для удобной разработки, но конечный вариант который идёт в kubernetes (в любое из окружений) всегда сборка как в production. Остальная отладка производится только локально. Почему необходимо чтобы сборка всегда была одинаковой?

Потому что наш docker image между окружениями должен быть неизменяемым. Создано это для того чтобы мы всегда были уверены в нашей сборке, т.е. образ оттестированный на stage окружении должен попадать в production точно таким же. Для всего остального мира наше приложение должно быть чёрным ящиком, которое лишь может принимать параметры.

Если же вам необходимо иметь внутри вашего кода ссылки или любые други изменяемые между окружениями объекты, то сделать это можно несколькими способами:

1. Мы можем динамически в зависимости от окружение монтировать в контейнер с нашим приложением json с нужными параметрами. Для этого нам нужно создать объект configmap в .helm/templates.

**10-app-config.yaml**

```yaml
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ .Chart.Name }}-config
data:
  config.json: |-
   {
    "domain": "{{ pluck .Values.global.env .Values.domain | first }}",
    "loginUrl": "{{ pluck .Values.global.env .Values.loginUrl | first }}"
   }
```


А затем примонтировать к нашему приложению в то место где мы могли получать его по запросу от клиента: \
Код из 01-app.yaml:


```yaml
       volumeMounts:
          - name: app-config
            mountPath: /app/dist/config.json
            subPath: config.json
      volumes:
        - name: app-config
          configMap:
            name: {{ .Chart.Name }}-config
```
После того как конфиг окажется внутри вашего приложения вы сможете обращатся к нему как обычно.

2. И второй вариант перед запуском приложения получать конфиги из внешнего ресурса,например из [consul](https://www.consul.io/). Подробно не будем расписывать данный вариант, так как он достоин отдельной главы. Дадим лишь два совета: 

*   Можно запускать его перед запуском приложения добавив в `command: ["node","/app/src/js/server.js"] `его запуск через `&&` как в синтаксисе любого shell языка, прямо в манифесте описывающем приложение:
```yaml
command: 
- /usr/bin/bash
- -c
- --
- "consul kv get app/config/urls && node /app/src/js/server.js"]
```


*   Либо добавив его запуск в [init-container](https://kubernetes.io/docs/concepts/workloads/pods/init-containers/) и подключив между инит контейнером и основным контейнером общий [volume](https://kubernetes.io/docs/concepts/storage/volumes/), это означает что просто нужно смонтировать volume в оба контейнера. Например [emptyDir](https://kubernetes.io/docs/concepts/storage/volumes/#emptydir). Тем самым консул отработав в инит контейнере, сохранит конфиг в наш volume, а приложение из основного контейнера просто заберёт этот конфиг. 

<a name="assets-implementation" />

## Какие изменения необходимо внести

Генерация ассетов происходит в image `node` на стадии `setup`, так как данная стадия рекомендуется для настройки приложения

Для уменьшения нагрузки на процесс основного приложения которое обрабатыаем логику работы Nodejs приложения мы будем отдавать статические файлы через `nginx`
Мы запустим оба контейнера одним деплойментом, а запросы разделим на уровне сущности Ingress.

### Изменения в сборке

Для начала добавим стадию сборки в наш docker image в файл [werf.yaml](/werf-articles/gitlab-nodejs-files/examples/werf_2.yaml)

```yaml
  setup:
  - name: npm run build
    shell: npm run build
    args:
      chdir: /app
```
Далее пропишем отдельный образ с nginx и перенесём в него собранную статику из нашего основного образа:
```yaml
---
image: node_assets
from: nginx:stable-alpine
docker:
  EXPOSE: '80'
import:
- image: node
  add: /app/dist/assets
  to: /usr/share/nginx/html/assets
  after: setup
```
Вы можете заметить два абсолютно новых поля конфигурации.
`import:` - позволяет импортировать директории из других ранее собраных образов.
Внутри мы можем списком указать образа с директориями которые мы хотим забрать.
В нашем случае мы берем наш образ `image:node` и берем из него директорию `add: /app/dist/assets` в которую webpack положил сгенерированную статику, и копируем её в `/usr/share/nginx/html/assets` в образе с nginx, и что примечательно пишем стадию после которой нам нужно файлы импортировать `after:setup`. В нашем случае мы не описывали никаких стадий, потому мы указали выполнить импорт после последней стадии по-умолчанию.

### Изменения в деплое

При таком подходе изменим деплой нашего приложения добавив еще один контейнер в наш деплоймент с приложением.  Укажем livenessProbe и readinessProbe, которые будут проверять корректную работу контейнера в поде. preStop команда необходима для корректного завершение процесса nginx. В таком случае при новом выкате новой версии приложения будет корректное завершение всех активных сессий.

```yaml
      - name: assets
{{ tuple "node_assets" . | include "werf_container_image" | indent 8 }}
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

В описании сервиса - так же должен быть указан правильный порт

```yaml
  ports:
  - name: http
    port: 80
    protocol: TCP
```


### Изменения в роутинге

Поскольку у нас маршрутизация запросов происходит черех nginx контейнер а не на основе ingress ресурсов - нам необходимо только указать коректный порт для сервиса

```yaml
      paths:
      - path: /
        backend:
          serviceName: {{ .Chart.Name }}
          servicePort: 80
```

Если мы хотим разделять трафик на уровне ingress - нужно разделить запросы по path и портам

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
