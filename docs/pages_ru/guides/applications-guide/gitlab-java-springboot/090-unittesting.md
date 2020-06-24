---
title: Юнит-тесты и Линтеры
sidebar: applications-guide
permalink: documentation/guides/applications-guide/gitlab-java-springboot/090-unittesting.html
layout: guide
toc: false
---

{% filesused title="Файлы, упомянутые в главе" %}
- .gitlab-ci.yml
{% endfilesused %}

В этой главе мы настроим в нашем базовом приложении выполнение тестов/линтеров. Запуск тестов и линтеров - это отдельная стадия в pipelinе Gitlab CI для выполнения которых могут быть нужны определенные условия.

Java - компилируемый язык, значит в случае проблем в коде приложение с большой вероятностью просто не соберется. Тем не менее хорошо бы получать информацию о проблеме в коде не дожидаясь пока упадет сборка, для чего воспользуемся  [maven checkstyle plugin](https://maven.apache.org/plugins/maven-checkstyle-plugin/usage.html).

Нам нужно добавить эту зависимость в наше приложение (в файл ____________) и прописать выполнение задания отдельной стадией на нашем gitlab runner командной [werf run](https://ru.werf.io/documentation/cli/main/run.html).

{% snippetcut name=".gitlab-ci.yaml" url="#" %}
```yaml
test:
  script:
    - werf run ____________ -- mvn checkstyle:checkstyle
  stage: test
  tags:
    - werf
  needs: ["Build"]
```
{% endsnippetcut %}

Созданную стадию нужно добавить в список стадий

{% snippetcut name=".gitlab-ci.yaml" url="#" %}
```yaml
stages:
  - build
  - test
  - deploy
```
{% endsnippetcut %}

Обратите внимание, что процесс будет выполняться на runner-е, внутри собранного контейнера, но без доступа к базе данных и каким-либо ресурсам kubernetes-кластера.

{% offtopic title="А если нужно больше?" %}
Если нужен доступ к ресурсам кластера или база данных — это уже не линтер: нужно собирать отдельный образ и прописывать сложный сценарий деплоя объектов kubernetes. Эти кейсы выходят за рамки нашего гайда для начинающих.
{% endofftopic %}

TODO: вот этот текст надо нарезать и встроить в общий шаблон страницы

Так же можно добавить этот плагин в pom.xml в секцию build (подробно описано в [документации](https://maven.apache.org/plugins/maven-checkstyle-plugin/usage.html) или можно посмотреть на готовый [pom.xml](воттут)) и тогда checkstyle будет выполняться до самой сборки при выполнении `mvn package`. Воспользуемся как раз этим способом для нашего примера. Стоит отметить, что в нашем случае используется [google_checks.xml](https://github.com/checkstyle/checkstyle/blob/master/src/main/resources/google_checks.xml) для описания checkstyle и мы запускаем их на стадии validate - до компиляции.

Для unit-тестирования воспользуемся инструментом предлагаемым по умолчанию - junit. Если вы пользовались start.spring.io - то он уже включен в pom.xml автоматичекси, если нет, то нужно его там прописать.
Запускаются тесты при выполнении `mvn package`.

Подробно посмотреть как это работает можно у нас в этом [репозитории](gitlab-java-springboot-files/04-demo-tests/)

<div>
    <a href="110-multipleapps.html" class="nav-btn">Далее: Несколько приложений в одном репозитории</a>
</div>
