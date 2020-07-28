---
title: Юнит-тесты и Линтеры
sidebar: applications_guide
permalink: documentation/guides/applications_guide/gitlab_java_springboot/090_unittesting.html
layout: guide
toc: false
---

{% filesused title="Файлы, упомянутые в главе" %}
- .gitlab-ci.yml
{% endfilesused %}

В этой главе мы настроим в нашем базовом приложении выполнение тестов/линтеров. Запуск тестов и линтеров - это отдельная стадия в pipelinе Gitlab CI для выполнения которых могут быть нужны определенные условия.

Java - компилируемый язык, значит в случае проблем в коде приложение с большой вероятностью просто не соберется. 
Есть два способа реализовать проверки lint:

* либо в отдельной стадии, описанной в `.gitlab-ci.yaml`
* либо как часть процесса сборки

Какой конкретно выбрать зависит от объёма зависимостей необходимых для запуска линтера. Если таких зависимостей мало, то удобно вынести тесты в отдельную стадию, которую запускать с помощью docker run. А если зависимостей много или они тяжелые, то лучше откорректировать стадию сборки, а конкретно файл `pom.xml`.

## Сборка отдельной стадией

Тем не менее хорошо бы получать информацию о проблеме в коде не дожидаясь пока упадет сборка, для чего воспользуемся  [maven checkstyle plugin](https://maven.apache.org/plugins/maven-checkstyle-plugin/usage.html).

Нам нужно добавить эту зависимость в наше приложение (в файл `pom.xml`) и прописать выполнение задания отдельной стадией на нашем gitlab runner командной [werf run](https://ru.werf.io/documentation/cli/main/run.html).

{% snippetcut name=".gitlab-ci.yaml" url="#" %}
{% raw %}
```yaml
test:
  script:
    - docker run --rm -v "$(pwd):/app" -w /app maven:3-jdk-8 mvn checkstyle:checkstyle
  stage: test
  tags:
    - werf
```
{% endraw %}
{% endsnippetcut %}

Созданную стадию нужно добавить в список стадий

{% snippetcut name=".gitlab-ci.yaml" url="#" %}
{% raw %}
```yaml
stages:
  - test
  - build
  - deploy
```
{% endraw %}
{% endsnippetcut %}

Обратите внимание, что процесс будет выполняться на runner-е, внутри собранного контейнера, но без доступа к базе данных и каким-либо ресурсам Kubernetes-кластера.

{% offtopic title="А если нужно больше?" %}
Если нужен доступ к ресурсам кластера или база данных — это уже не линтер: нужно собирать отдельный образ и прописывать сложный сценарий деплоя объектов Kubernetes. Эти кейсы выходят за рамки нашего гайда для начинающих.
{% endofftopic %}


## Тестирование внутри стадии сборки

Так же можно добавить этот плагин в `pom.xml` в секцию `build` (подробно описано в [документации](https://maven.apache.org/plugins/maven-checkstyle-plugin/usage.html)) и тогда checkstyle будет выполняться до самой сборки при выполнении `mvn package`. Стоит отметить, что в нашем случае используется [google_checks.xml](https://github.com/checkstyle/checkstyle/blob/master/src/main/resources/google_checks.xml) для описания checkstyle. Мы запускаем их на стадии `validate` - до компиляции.

<div>
    <a href="110_multipleapps.html" class="nav-btn">Далее: Несколько приложений в одном репозитории</a>
</div>
