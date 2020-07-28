---
title: Динамические окружения
sidebar: applications_guide
permalink: documentation/guides/applications_guide/gitlab_rails/120_dynamicenvs.html
layout: guide
toc: false
---

{% filesused title="Файлы, упомянутые в главе" %}
- .gitlab-ci.yaml
{% endfilesused %}

В этой главе мы добавим к нашему базовому приложению возможность выкатываться на множество окружений, организуя так называемые "feature-ветки" или "review-окружения".

Не редко необходимо разрабатывать и тестировать сразу несколько feature для вашего приложения, и нет понимания как это делать, если у вас всего два окружения. Разработчику или тестеру приходится дожидаться своей очереди на контуре staging и затем проводить необходимые манипуляции с кодом (тестирование, отладка, демонстрация функционала). Таким образом разработка сильно замедляется. Динамические окружения позволяют развернуть и погасить окружение под каждую ветку в любой момент, разблокируя процессы тестирования.

Для того, чтобы организовать динамические окружения, необходимо, чтобы:

* Домен, на котором разворачивается приложение, конфигурировался в объекте Ingress на основании значения из `.gitlab-ci.yaml` (мы сделали это в главе [Базовое приложение](020_basic.html))
* В `.gitlab-ci.yaml` было прописано создание и удаление Review-окружений.

Пропишем в`.gitlab-ci.yml` создание Review-окружений:

{% snippetcut name=".gitlab-ci.yaml" url="#" %}
{% raw %}
```yaml
Deploy to Review:
  extends: .base_deploy
  stage: deploy
  environment:
    name: review/${CI_COMMIT_REF_SLUG}
    url: http://${CI_COMMIT_REF_SLUG}.mydomain.io
    on_stop: Stop Review
  only:
    - /^feature-*/
  when: manual
```
{% endraw %}
{% endsnippetcut %}

Создаваемые окружения нужно не забывать отключать — в противном случае ресуры в кластере закончатся. Мы добавили зависимость `on_stop: Stop Review` простыми словами означающую, что мы будем останавливать наше окружение стадией `Stop Review`. Сама стадия описывается следующим образом:

{% snippetcut name=".gitlab-ci.yaml" url="#" %}
{% raw %}
```yaml
Stop Review:
  stage: deploy
  variables:
    GIT_STRATEGY: none
  script:
    - werf dismiss --env $CI_ENVIRONMENT_SLUG --namespace ${CI_ENVIRONMENT_SLUG} --with-namespace
  when: manual
  environment:
    name: review/${CI_COMMIT_REF_SLUG}
    action: stop
  only:
    - /^feature-*/
```
{% endraw %}
{% endsnippetcut %}

Удаление helm-релиза (и, как следствие, всех выкаченных им объектов), осуществляется командой [`werf dismiss`](https://ru.werf.io/documentation/cli/main/dismiss.html). Мы указываем название окружения ([`--env $CI_ENVIRONMENT_SLUG`](https://docs.gitlab.com/ee/ci/environments/#environment-variables-and-runner))

С помощью атрибута `GIT_STRATEGY: none` мы говорит нашему ранеру, что check out нашего кода не требуется, чем немного ускоряем операцию и разгружаем раннер.

При таком CI-процессе мы можем выкатывать каждую ветку `feature/*` в отдельный namespace с изолированной базой данных, накатом необходимых миграций и например проводить тесты для данного окружения.
