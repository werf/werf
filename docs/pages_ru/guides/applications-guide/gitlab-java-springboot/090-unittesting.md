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


Java - компилируемый язык, значит в случае проблем в коде приложение с большой вероятностью просто не соберется. Тем не менее хорошо бы получать информацию о проблеме в коде не дожидаясь пока упадет сборка.
Чтобы этого избежать пробежимся по коду линтером, а затем запустим unit-тесты.
Для запуска линта воспользуемся [maven checkstyle plugin](https://maven.apache.org/plugins/maven-checkstyle-plugin/usage.html). Запускать можно его несколькими способами - либо вынести на отдельную стадию в gitlab-ci - перед сборкой или вызывать только при merge request. Например:

``` yaml
test: 
  stage: test
  script: mvn checkstyle:checkstyle
  only:
  - merge_requests
```

Так же можно добавить этот плагин в pom.xml в секцию build (подробно описано в [документации](https://maven.apache.org/plugins/maven-checkstyle-plugin/usage.html) или можно посмотреть на готовый [pom.xml](воттут)) и тогда checkstyle будет выполняться до самой сборки при выполнении `mvn package`. Воспользуемся как раз этим способом для нашего примера. Стоит отметить, что в нашем случае используется [google_checks.xml](https://github.com/checkstyle/checkstyle/blob/master/src/main/resources/google_checks.xml) для описания checkstyle и мы запускаем их на стадии validate - до компиляции.

Для unit-тестирования воспользуемся инструментом предлагаемым по умолчанию - junit. Если вы пользовались start.spring.io - то он уже включен в pom.xml автоматичекси, если нет, то нужно его там прописать.
Запускаются тесты при выполнении `mvn package`.

Подробно посмотреть как это работает можно у нас в этом [репозитории](gitlab-java-springboot-files/04-demo-tests/)

<div>
    <a href="110-multipleapps.html" class="nav-btn">Далее: Несколько приложений в одном репозитории</a>
</div>
