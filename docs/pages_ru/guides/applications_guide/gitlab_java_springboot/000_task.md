---
title: Как использовать гайд
sidebar: applications_guide
guide_code: gitlab_java_springboot
permalink: documentation/guides/applications_guide/gitlab_java_springboot/000_task.html
layout: guide
toc: false
---

Этот гайд расскажет, как Java-spring разработчику развернуть своё приложение в Kubernetes с помощью утилиты Werf.

![](/images/applications-guide/navigation.png)

Обязательны к прочтению главы "Подготовка к работе" и "Базовые настройки" — в них будут разобраны вопросы настройки окружения и основы работы с Werf, сборки и деплоя приложения в production. Однако, чтобы построить серьёзное приложение понадобится чуть больше навыков, раскрытых в других главах.

## Работа с исходными кодами

Для прохождения гайда предоставляется много исходного кода: как самого приложения, которое будет переносится в Kubernetes, так и кода инфраструктуры, связанного с каждой главой. В тексте будут контрольные точки, где вы можете сверить состояние своих исходников с образцом.

Мы рекомендуем сперва пройти гайд с предложенным приложением, разобравшись в механиках сборки и деплоя, и только затем — пробовать перенести в Kubernetes свой код.

## Условные обозначения

В начале каждой главы мы показываем, **какие файлы будут затронуты**:

{% filesused title="Файлы, упомянутые в главе" %}
- .helm/templates/deployment.yaml
- .helm/templates/ingress.yaml
- .helm/templates/service.yaml
- .helm/values.yaml
- .helm/secret-values.yaml
{% endfilesused %}

Для вещей, выходящих за рампки повествования, но полезных для саморазвития, предусмотрены **схлопытвающиеся блоки**, например:

{% offtopic title="Что делать, если вы не работали с Helm?" %}

Мы не будем вдаваться в подробности [разработки yaml манифестов с помощью Helm для Kubernetes](https://habr.com/ru/company/flant/blog/423239/). Осветим лишь отдельные её части, которые касаются данного приложения и werf в целом. Если у вас есть вопросы о том как именно описываются объекты Kubernetes, советуем посетить страницы документации по Kubernetes с его [концептами](https://kubernetes.io/ru/docs/concepts/) и страницы документации по разработке [шаблонов](https://helm.sh/docs/chart_template_guide/) в Helm.
{% endofftopic %}

В коде можно регулярно встретить **блоки с кодом**. Обратите внимание, что они **почти всегда отображают только часть файла**. Куда вставлять этот кусок текста — объясняется в тексте гайда, а также вы можете нажать на ссылку (в приведённом ниже примере - `deployment.yaml`) и перейти к github с полным исходным кодом файла.

{% snippetcut name="deployment.yaml" url="#" %}
{% raw %}
```yaml
      containers:
      - name: basicapp
        command:
         - java
         - -jar
         - /app/target/demo-1.0.jar $JAVA_OPT
{{ tuple "basicapp" . | include "werf_container_image" | indent 8 }}
```
{% endraw %}
{% endsnippetcut %}

В конце будет **блок с полными файлами**, с содержимым которых вы работали в данном разделе. Это готовые файлы, которые помогут вам сверить полученные знания.

## Готовые файлы

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
