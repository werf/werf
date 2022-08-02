---
title: Шаги
permalink: advanced/helm/deploy_process/steps.html
---

Во время запуска команды `werf converge` werf запускает процесс деплоя, включающий следующие этапы:

 1. Преобразование шаблонов чартов в единый список манифестов ресурсов Kubernetes и их проверка.
 2. Последовательный запуск [хуков]({{ "/advanced/helm/deploy_process/helm_hooks.html" | true_relative_url }}) `pre-install` или `pre-upgrade`, отсортированных по весу, и контроль каждого хука до завершения его работы с выводом логов.
 3. Группировка ресурсов Kubernetes, не относящихся к хукам, по их [весу]({{ "/advanced/helm/deploy_process/deployment_order.html" | true_relative_url }}) и последовательное развертывание каждой группы в соответствии с ее весом: создание/обновление/удаление ресурсов и отслеживание их до готовности с выводом логов в процессе.
 4. Запуск [хуков]({{ "/advanced/helm/deploy_process/helm_hooks.html" | true_relative_url }}) `post-install` или `post-upgrade` по аналогии с хуками `pre-install` и `pre-upgrade`.

## Отслеживание ресурсов

werf использует библиотеку [kubedog](https://github.com/werf/kubedog) для отслеживания ресурсов. Отслеживание можно настроить для каждого ресурса, указывая соответствующие [аннотации]({{ "/reference/deploy_annotations.html" | true_relative_url }}) в шаблонах чартов.

## Работа с несколькими кластерами Kubernetes

В некоторых случаях, необходима работа с несколькими кластерами Kubernetes для разных окружений. Все что вам нужно, это настроить необходимые [контексты](https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters) kubectl для доступа к необходимым кластерам и использовать для werf параметр `--kube-context=CONTEXT`, совместно с указанием окружения.

## Сабчарты

Во время процесса деплоя werf выполнит рендер, создаст все требуемые ресурсы указанные во всех [используемых сабчартах]({{ "/advanced/helm/configuration/chart_dependencies.html" | true_relative_url }}) и будет отслеживать каждый из этих ресурсов до состояния готовности.
