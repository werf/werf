---
title: Управление релизами
permalink: documentation/advanced/helm/releases/manage_releases.html
---

Следующие основные werf команды создают и удаляют релизы:

 - [`werf converge`]({{ "documentation/reference/cli/werf_converge.html" | true_relative_url }}) создаёт новую версию релиза для проекта;
 - [`werf dismiss`]({{ "documentation/reference/cli/werf_dismiss.html" | true_relative_url }}) удаляет все существующие версии релизов для проекта.

werf предоставляет следующие команды низкого уровня для управления релизами:

 - [`werf helm list -A`]({{ "documentation/reference/cli/werf_helm_list.html" | true_relative_url }}) — выводит список всех релизов всех namespace'ов кластера;
 - [`werf helm get all RELEASE`]({{ "documentation/reference/cli/werf_helm_get_all.html" | true_relative_url }}) — для получения информации по указанному релизу, манифестов, хуков и values, записанных в версию релиза;
 - [`werf helm status RELEASE`]({{ "documentation/reference/cli/werf_helm_status.html" | true_relative_url }}) — для получения статуса последней версии указанного релиза;
 - [`werf helm history RELEASE`]({{ "documentation/reference/cli/werf_helm_history.html" | true_relative_url }}) — для получения списка версий указанного релиза.
