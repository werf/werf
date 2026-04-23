---
title: Planning
permalink: usage/deploy/planning.html
---

## Review ongoing changes before deployment

To see how the resources in the cluster will change during the next deployment run `werf plan`:

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

## Deploying with a two-stage workflow

The planning feature is quite handy, but on its own, it does not guarantee that the exact same deployment plan will be used during a subsequent `werf converge`, as the cluster state may change after the `werf plan` run. To ensure that the exact planned changes will be applied, you can utilize a two-stage deployment workflow:

1. **Plan:** Generate, review, and save a plan artifact using the --save-plan flag:
```
werf plan --save-plan=plan.gz
```
2. **Deploy:** Perform `werf converge` using the pre-generated reviewed plan:
````
werf converge --use-plan=plan.gz
```

The plan artifact is a gzip-compressed JSON file that can be reviewed with `werf plan --show-plan plan.gz` command. If you used `--secret-key` or `WERF_SECRET_KEY` environment variable during planning, the plan artifact will be encrypted.
