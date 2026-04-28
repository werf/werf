---
title: Планирование
permalink: usage/deploy/planning.html
---

## Показать запланированные изменения перед развертыванием

Чтобы увидеть, как изменятся ресурсы в кластере во время следующего развертывания, выполните команду `werf plan`:

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

## Выполнение деплоя в две стадии

Функция планирования Werf весьма удобна, но сама по себе она не гарантирует, что при последующем выполнении `werf converge` будет использован тот же самый план развертывания, так как состояние кластера может измениться после запуска `werf plan`. Чтобы гарантировать применение именно запланированы изменений, вы можете выполнить деплой в две стации:

1. **Plan:** Сгенерируйте, проверьте и сохраните артефакт плана с помощью флага --save-plan:
```
werf plan --save-plan=plan.gz
```
2. **Deploy:** Выполните `werf converge`, используя ранее созданный и проверенный план:
```
werf converge --use-plan=plan.gz
```

Артефакт плана представляет собой JSON-файл, сжатый с помощью gzip, который можно просмотреть командой `werf plan --show-plan plan.gz`. Если при планировании использовался флаг `--secret-key` или переменная окружения `$WERF_SECRET_KEY`, артефакт плана будет зашифрован.
