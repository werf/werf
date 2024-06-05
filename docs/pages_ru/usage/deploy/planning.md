---
title: Планирование
permalink: usage/deploy/planning.html
---

## Показать ожидаемые изменения в кластере перед развертыванием

Чтобы увидеть, как изменятся ресурсы в кластере во время следующего развертывания, выполните команду `werf plan` перед выполнением `werf converge`:

```shell
$ werf plan --repo example.org/mycompany/myapp

┌ Update Deployment/myapp
│     annotations:
│ +     myanno: myval
│     ...
│           resources:
│             limits:
│ +             cpu: 100m
│ -             memory: 100Mi
│ +             memory: 200Mi
└ Update Deployment/myapp

┌ Create ConfigMap/mycm
| + apiVersion: v1
| + kind: ConfigMap
| + metadata:
| +   name: mycm
| + data:
| +   mykey: myval
└ Create ConfigMap/mycm

Planned changes summary for release "myapp" (namespace: "myapp"):
- create: 1 resource(s)
- update: 1 resource(s)
```

Теперь выполните `werf converge` и убедитесь, что это именно то, что произойдет:

```shell
$ werf converge --repo example.org/mycompany/myapp
...

┌ Completed operations
│ Update resource: Deployment/myapp
│ Create resource: ConfigMap/mycm
└ Completed operations

Succeeded release "myapp" (namespace: "myapp")
```
