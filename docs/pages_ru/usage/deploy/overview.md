---
title: Обзор
permalink: usage/deploy/overview.html
---

Начать пользоваться werf для выката, используя существующие [Helm](https://helm.sh) чарты, не составит никакого труда, т.к. они полностью совместимы с werf. Конфигурация описывается в формате аналогичном формату [Helm-чарта]({{ "usage/deploy/configuration/chart.html" | true_relative_url }}).

werf включает всю существующую функциональность Helm (он вкомпилен в werf) и свои дополнения:
 - несколько настраиваемых режимов отслеживания выкатываемых ресурсов, в том числе обработка логов и событий;
 - интеграция собираемых образов с [шаблонами]({{ "usage/deploy/configuration/templates.html" | true_relative_url }}) Helm-чартов;
 - возможность простановки произвольных аннотаций и лейблов во все ресурсы, создаваемые в Kubernetes, глобально через опции утилиты werf;
 - werf читает все конфигурационные файлы helm из git в соответствии с режимом [гитерминизма]({{ "usage/project_configuration/giterminism.html" | true_relative_url }}), что позволяет создавать по-настоящему воспроизводимые pipeline'ы в CI/CD и на локальных машинах.
 - и другие особенности, о которых пойдёт речь далее.

С учётом всех этих дополнений и способа реализации можно рассматривать werf как альтернативный или улучшенный helm-клиент, для деплоя стандартных helm-совместимых чартов.

Для работы с приложением в Kubernetes используются следующие основные команды:
 - [converge]({{ "reference/cli/werf_converge.html" | true_relative_url }}) — для установки или обновления приложения в кластере, и
 - [dismiss]({{ "reference/cli/werf_dismiss.html" | true_relative_url }}) — для удаления приложения из кластера.
 - [bundle apply]({{ "reference/cli/werf_bundle_apply.html" | true_relative_url }}) — для выката приложения из опубликованного ранее [бандла]({{ "usage/deploy/bundles.html" | true_relative_url }}).

Данная глава покрывает следующие разделы:
 1. Конфигурация helm для деплоя вашего приложения в kubernetes с помощью werf: [раздел "конфигурация"]({{ "usage/deploy/configuration/chart.html" | true_relative_url }}).
 2. Как werf реализует процесс деплоя: [раздел "процесс деплоя"]({{ "usage/deploy/deploy_process/steps.html" | true_relative_url }}).
 3. Что такое релиз и как управлять выкаченными релизами своих приложений: [раздел "релизы"]({{ "usage/deploy/releases/release.html" | true_relative_url }})
