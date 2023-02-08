---
title: Обзор
permalink: usage/distribute/overview.html
---

Обычный цикл доставки приложений с werf выглядит как сборка образов, их публикация и последующее развертывание чартов, для чего бывает достаточно одного вызова команды `werf converge`. Но иногда возникает необходимость разделить *дистрибуцию* артефактов (образы, чарты) и их *развертывание*, либо даже реализовать развертывание артефактов вовсе без werf, а с использованием стороннего ПО.

В этом разделе рассматриваются способы дистрибуции образов и бандлов (чартов и связанных с ними образов) для их дальнейшего развертывания с werf или без него. Инструкции развертывания опубликованных артефактов можно найти в разделе «Развертывание».

## Пример дистрибуции образа

Для дистрибуции единственного образа, собираемого через Dockerfile, который потом будет развернут сторонним ПО, достаточно двух файлов и одной команды `werf export`, запущенной в Git-репозитории приложения:

```yaml
# werf.yaml:
project: myproject
configVersion: 1
---
image: myapp
dockerfile: Dockerfile
```

```dockerfile
# Dockerfile:
FROM node

WORKDIR /app
COPY . .
RUN npm ci

CMD ["node", "server.js"]
```

```shell
werf export myapp --repo example.org/myproject --tag other.example.org/myproject/myapp:latest
```

Результат: опубликован образ приложения `other.example.org/myproject/myapp:latest`, готовый для развертывания сторонним ПО.

## Пример дистрибуции бандла

Для дистрибуции бандла для дальнейшего его развертывания с werf, подключения его как зависимого чарта или развертывания бандла как чарта сторонним ПО в простейшем случае достаточно трёх файлов и одной команды `werf bundle publish`, запущенной в Git-репозитории приложения:

```yaml
# werf.yaml:
project: mybundle
configVersion: 1
---
image: myapp
dockerfile: Dockerfile
```

```dockerfile
# Dockerfile:
FROM node

WORKDIR /app
COPY . .
RUN npm ci

CMD ["node", "server.js"]
```

{% raw %}

```
# .helm/templates/myapp.yaml:
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
spec:
  selector:
    matchLabels:
      app: myapp
  template:
    metadata:
      labels:
        app: myapp
    spec:
      containers:
      - image: {{ $.Values.werf.image.myapp }}
```

{% endraw %}

```shell
werf bundle publish --repo example.org/bundles/mybundle
```

Результат: опубликован чарт `example.org/bundles/mybundle:latest` и связанный с ним собранный образ.
