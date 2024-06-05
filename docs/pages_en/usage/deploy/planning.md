---
title: Planning
permalink: usage/deploy/planning.html
---

## Show expected changes in cluster before deployment

To see how the resources in the cluster will change during the next deployment run `werf plan` command before running `werf converge`:

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

Now run `werf converge` and see that this is exactly what happens:

```shell
$ werf converge --repo example.org/mycompany/myapp
...

┌ Completed operations
│ Update resource: Deployment/myapp
│ Create resource: ConfigMap/mycm
└ Completed operations

Succeeded release "myapp" (namespace: "myapp")
```
