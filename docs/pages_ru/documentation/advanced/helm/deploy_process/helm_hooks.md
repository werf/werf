---
title: Helm хуки
sidebar: documentation
permalink: documentation/advanced/helm/deploy_process/helm_hooks.html
---

Helm-хуки — произвольные ресурсы Kubernetes, помеченные специальной аннотацией `helm.sh/hook`. Например:

```yaml
kind: Job
metadata:
  name: somejob
  annotations:
    "helm.sh/hook": pre-upgrade,pre-install
    "helm.sh/hook-weight": "1"
```

Существует много разных helm-хуков, влияющих на процесс деплоя. Вы уже читали [ранее]({{ "/documentation/advanced/helm/deploy_process/steps.html" | true_relative_url }}) про `pre|post-install|upgade` хуки, используемые в процессе деплоя. Эти хуки наиболее часто используются для выполнения таких задач, как миграция (в хуках `pre-upgrade`) или выполнении некоторых действий после деплоя. Полный список доступных хуков можно найти в соответствующей документации [Helm](https://helm.sh/docs/topics/charts_hooks/).

Хуки сортируются в порядке возрастания согласно значению аннотации `helm.sh/hook-weight` (хуки с одинаковым весом сортируются по имени в алфавитном порядке), после чего хуки последовательно создаются и выполняются. werf пересоздает ресурс Kubernetes для каждого хука, в случае когда ресурс уже существует в кластере. Созданные хуки ресурсов не удаляются после выполнения, если не указано [специальной аннотации `"helm.sh/hook-delete-policy": hook-succeeded,hook-failed`](https://helm.sh/docs/topics/charts_hooks/).
