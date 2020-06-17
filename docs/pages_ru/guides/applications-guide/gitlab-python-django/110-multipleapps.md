---
title: Несколько приложений в одном репозитории
sidebar: applications-guide
permalink: documentation/guides/applications-guide/template/110-multipleapps.html
layout: guide
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

Покажем это на примере, запустим рядом с нашим приложением простой websocket chat.

## Сборка приложений

Сборка приложения с несколькими образами описана в [статье](https://ru.werf.io/documentation/guides/advanced_build/multi_images.html). На ее основе покажем наш пример для нашего приложения.

Структура каталогов будет организована следующим образом


```
├── .helm
│   ├── templates
│   └── values.yaml
├── django
├── chat
└── werf.yaml
```



Сборка для chat приложения описана в файле werf.yaml как отдельный образ



```yaml
image: chat
from: node:10.16.3
git:
- add: /chat
  to: /app
ansible:
  beforeInstall:
  - name: install dependencies
    apt:
      name:
      - yarn
  install:
  - name: install node dependencies
    shell: yarn install
    args:
      chdir: /app
  setup:
  - name: build js app
    shell: yarn build
    args:
      chdir: /app
```


{{1. Добавляем кронджоб}}
{{2. Добавляем воркер/консюмер}}
{{3. Добавляем вторую приложуху на другом языке (например, это может быть webscoket’ы на nodejs; показать организацию helm, организацию werf.yaml, и ссылку на другую статью)}}

