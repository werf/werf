---
title: Юнит-тесты и Линтеры
sidebar: applications-guide
permalink: documentation/guides/applications-guide/gitlab-python-django/090-unittesting.html
layout: guide
toc: false
---

{% filesused title="Файлы, упомянутые в главе" %}
- .gitlab-ci.yml
{% endfilesused %}

Запуск тестов и линтеров - это отдельные стадии в pipelinе для выполнения которых могут быть нужны определенные условия.

В django есть встроенный механизм для их запуска, это стандартный `unittest`. Что-бы ими воспользоваться, необходимо  собрать образ приложения и запустить выполнение задания отдельной стадией на нашем gitlab runner командной [werf run](https://ru.werf.io/documentation/cli/main/run.html).


```
Unittests:
  script:
    - werf run django -- python manage.py test
```


При таком запуске наш kubernetes кластер не задействован.

Если нам нужно проверить приложение линтером, но данные зависимости не нужны в итоговом образе - нам необходимо собрать отдельный образ. Данный пример будет в репозитории с примерами а тут мы его не будем описывать.

<div>
    <a href="110-multipleapps.html" class="nav-btn">Далее: Несколько приложений в одном репозитории</a>
</div>
