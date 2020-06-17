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

Как было уже описано в главе про сборку ассетов - можно использовать один репозиторий для сборки нескольких приложений - у нас это были бек на Java и фронт, представляющий собой собранную webpack-ом статику в контейнере с nginx.
Подробно такая сборка описана в отдельной [статье](https://ru.werf.io/documentation/guides/advanced_build/multi_images.html).
Вкратце напомню о чем шла речь.
У нас есть основное приложение - Java и ассеты, собираемые webpack-ом. 
Все что связано с генерацией ассетов лежит в assets. Мы добавили отдельный artifact (на основе nodejs) и image (на основе nginx:alpine). 
В artifact добавляем в /app нашу папку assets, там все собираем, как описано было ранее и отдаем результат в image с nginx.

```yaml
git:
  - add: /assets
    to: /app
    excludePaths:

    stageDependencies:
      install:
      - package.json
      - webpack.config.js
      setup:
      - "**/*"
```

[werf.yaml](gitlab-java-springboot-files/05-demo-complete/werf.yaml:45-53)

Такая организация репозитория удобна, если нужно выкатывать приложение целиком, или команда разработки не большая. Мы уже [рассказывали](https://www.youtube.com/watch?v=g9cgppj0gKQ) в каких случаях это правильный путь.

