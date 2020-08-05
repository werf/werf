---
title: Как использовать гайд
sidebar: applications_guide
guide_code: gitlab_nodejs
permalink: documentation/guides/applications_guide/gitlab_nodejs/000_task.html
layout: guide
toc: false
---

Этот гайд расскажет, как NodeJs разработчику развернуть своё приложение в Kubernetes с помощью утилиты Werf.

![](/images/applications-guide/navigation.png)

Обязательны к прочтению главы "Подготовка к работе" и "Базовые настройки" — в них будут разобраны вопросы настройки окружения и основы работы с Werf, сборки и деплоя приложения в production. Однако, чтобы построить серьёзное приложение понадобится чуть больше навыков, раскрытых в других главах.

## Работа с исходными кодами

Для прохождения гайда предоставляется много исходного кода: как самого приложения, которое будет переносится в Kubernetes, так и кода инфраструктуры, связанного с каждой главой. В тексте будут контрольные точки, где вы можете сверить состояние своих исходников с образцом.

Мы рекомендуем сперва пройти гайд с предложенным приложением, разобравшись в механиках сборки и деплоя, и только затем — пробовать перенести в Kubernetes свой код.

## Условные обозначения

В начале каждой главы мы показываем, **какие файлы будут затронуты**:

{% filesused title="Файлы, упомянутые в главе" %}
- just/an/example.yaml
- of/files.yaml
- like.yml
- many.yml
- others.yml
{% endfilesused %}

Для вещей, выходящих за рампки повествования, но полезных для саморазвития, предусмотрены **схлопытвающиеся блоки**, например:

{% offtopic title="Нажми сюда чтобы узнать больше" %}

Это просто пример блока, который может раскрываться. Здесь, внутри, будет дополнительная информация для самых любознательных и желающих разобраться в матчасти.

{% endofftopic %}

В коде можно регулярно встретить **блоки с кодом**. Обратите внимание, что они **почти всегда отображают только часть файла**. Куда вставлять этот кусок текста — объясняется в тексте гайда, а также вы можете нажать на ссылку (в приведённом ниже примере - `deployment.yaml`) и перейти к github с полным исходным кодом файла. Пропущенный текст обозначается с помощью `<...>`:

{% snippetcut name="deployment.yaml" url="#" %}
{% raw %}
```yaml
git:
  <...>
  stageDependencies:
    install:
    - package.json
```
{% endraw %}
{% endsnippetcut %}

Мы против бездумного копирования файлов, но для того, чтобы вам было проще разобраться — в некоторых точках гайда есть **milestone-ы с полным кодом файлов на текущий момент**. Таким образом вы можете свериться с образцом:

<div class="tabs">
  <a href="javascript:void(0)" class="tabs__btn active" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'example_1')">example_1.yaml</a>
  <a href="javascript:void(0)" class="tabs__btn" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'example_2')">example_2.yaml</a>
</div>

<div id="example_1" class="tabs__content active">
{% snippetcut name="example_1.yaml" url="#" limited="true" %}
{% raw %}
```yaml
      containers:
      - name: example_1
        command:
         - java
         - -jar
         - /app/target/demo-1.0.jar $JAVA_OPT
      containers:
      - name: example_1
        command:
         - java
         - -jar
         - /app/target/demo-1.0.jar $JAVA_OPT
               containers:
      - name: example_1
        command:
         - java
         - -jar
         - /app/target/demo-1.0.jar $JAVA_OPT
      containers:
      - name: example_1
        command:
         - java
         - -jar
         - /app/target/demo-1.0.jar $JAVA_OPT
      containers:
      - name: example_1
        command:
         - java
         - -jar
         - /app/target/demo-1.0.jar $JAVA_OPT
               containers:
      - name: example_1
        command:
         - java
         - -jar
         - /app/target/demo-1.0.jar $JAVA_OPT
{{ tuple "basicapp" . | include "werf_container_image" | indent 8 }}
```
{% endraw %}
{% endsnippetcut %}
</div>
<div id="example_2" class="tabs__content">
{% snippetcut name="example_2.yaml" url="#" limited="true" %}
{% raw %}
```yaml
      containers:
      - name: example_2
        command:
         - java
         - -jar
         - /app/target/demo-1.0.jar $JAVA_OPT
      containers:
      - name: example_2
        command:
         - java
         - -jar
         - /app/target/demo-1.0.jar $JAVA_OPT
{{ tuple "basicapp" . | include "werf_container_image" | indent 8 }}
```
{% endraw %}
{% endsnippetcut %}
</div>

<div>
    <a href="010_preparing.html" class="nav-btn">Далее: Подготовка к работе</a>
</div>
