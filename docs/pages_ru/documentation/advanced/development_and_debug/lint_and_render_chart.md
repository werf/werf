---
title: Рендеринг и проверка конфигурации
sidebar: documentation
permalink: documentation/advanced/development_and_debug/lint_and_render_chart.html
---

Во время разработки [шаблонов helm-чартов]({{ "documentation/advanced/helm/basics.html#шаблоны" | true_relative_url: page.url }}) зачастую полезно выполнять проверку корректности синтаксиса перед выполнением процесса деплоя.

werf содержит два инструмента для выполнения этой задачи:

 1. [Рендеринг шаблонов](#рендеринг)
 2. [Линтер шаблонов](#линтер)

## Рендеринг

Во время рендеринга шаблонов werf возвращает содержимое всех manifest-файлов шаблона выполняя, в том числе и все Go-шаблоны.

Для получения отрендеренного manifest-файла шаблона необходимо использовать команду [werf helm template]({{ "documentation/reference/cli/werf_helm_template.html" | true_relative_url: page.url }}). Этой команде можно передавать все те-же параметры что и команде [`werf converge`]({{ "documentation/reference/cli/werf_converge.html" | true_relative_url: page.url }}), в том числе передавать дополнительные [переменные]({{ "documentation/advanced/helm/basics.html#данные" | true_relative_url: page.url }}), адрес репозитория Docker-образов и другие параметры.

Рендеринг помогает при отладке проблем деплоя связанных с ошибками в шаблонах, YAML-формате, описании объектов Kubernetes и т.д..

## Линтер

Линтер выполняет проверку [чарта]({{ "documentation/advanced/helm/basics.html#чарт" | true_relative_url: page.url }}) на различные проблемы, например:
 * Ошибки в Go-шаблонах;
 * Ошибки в YAML-синтаксисе;
 * Ошибки в синтаксисе объектов Kubernetes: не корректный тип объекта, отсутствующие параметры, поля, и т.д.;
 * Логические ошибки в описании объектов Kubernetes ([скоро](https://github.com/werf/werf/issues/1187)): отсутствующие label у ресурсов, ошибочные имена у связанных ресурсов, проверка apiVersion объекта на корректность, и т.д.;
 * Возможные проблемы безопасности ([скоро](https://github.com/werf/werf/issues/1317)).

Для запуска линтера необходимо выполнить команду [`werf helm lint`]({{ "documentation/reference/cli/werf_helm_lint.html" | true_relative_url: page.url }}). Ее можно выполнять как локально, так и в рамках pipeline CI/CD систем в качестве автоматического теста чарта на ошибки.
Этой команде можно передавать все те-же параметры что и команде [`werf converge`]({{ "documentation/reference/cli/werf_converge.html" | true_relative_url: page.url }}), в том числе  передавать дополнительные [переменные]({{ "documentation/advanced/helm/basics.html#данные" | true_relative_url: page.url }}), адрес репозитория Docker-образов и другие параметры.
