---
title: Юнит-тесты и Линтеры
sidebar: applications-guide
permalink: documentation/guides/applications-guide/gitlab-nodejs/090-unittesting.html
layout: guide
toc: false
---

{% filesused title="Файлы, упомянутые в главе" %}
- .gitlab-ci.yml
{% endfilesused %}


В качестве примера того каким образом строить в нашем CI запуск юнит тестов, мы продемонстрируем запуск ESLint.

ESLint - это линтер для языка программирования JavaScript, написанный на Node.js.

Он чрезвычайно полезен, потому что JavaScript, будучи интерпретируемым языком, не имеет этапа компиляции и многие ошибки могут быть обнаружены только во время выполнения.

Мы точно также добавляем пакет с нашим линтером в package.json и создаем к нему конфигурационный файл .eslintrc.json

```
node
├── migrations
├── src
├── package.json
├── .eslintrc.json
...
```

Для того чтобы запустить наш линтер мы добавляем наш .gitlab-ci.yml отдельную стадию:

```yaml
Run_Tests:
  stage: test
  script:
    - werf run --stages-storage :local node -- npm run pretest
  tags:
    - werf
  needs: ["Build"]
```

и не забываем добавить её в список стадий:

```yaml
stages:
  - build
  - test
  - deploy
```

В данном случае мы после сборки нашего docker image просто запускаем его командой [werf run](https://ru.werf.io/documentation/cli/main/run.html).

При таком запуске наш kubernetes кластер не задействован.

Полная конфигурация линтера доступна в примерах, тут мы описали лишь концепцию в примерах.

<div>
    <a href="110-multipleapps.html" class="nav-btn">Далее: Несколько приложений в одном репозитории</a>
</div>
