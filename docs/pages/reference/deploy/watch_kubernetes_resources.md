---
title: Watch kubernetes resources
sidebar: reference
permalink: reference/deploy/watch_kubernetes_resources.html
author: Timofey Kirillov <timofey.kirillov@flant.com>
---

Dapp has ability of watching statuses and logs of kubernetes resources during deploy process.

## Watch hook logs

Dapp can print helm hook logs ([more about helm hooks](https://github.com/helm/helm/blob/master/docs/charts_hooks.md)).

To enable logs special annotation `"dapp/watch": "true"` should be specified in the resource template. For example:

```yaml
...
kind: Job
metadata:
  ...
  annotations:
    "helm.sh/hook": post-install,post-upgrade
    "helm.sh/hook-weight": "1"
    "dapp/watch": "true"
```

Logs of pods will be printed for a job, until job terminated.

**NOTICE** For now logs watch is supported only for **Job** resource kind.

## Watch release status

After helm release has been installed or updated dapp will watch all Deployments that existed in the chart.

**Status watch procedure** consists of:

1. Periodical polling of the resource status and printing info about resource readiness.
2. Periodical printing of the resource logs and events.

The procedure is ended when resource readiness condition is reached.

This is **default behaviour** for any Deployment found in chart. To turn it off for specific deployment add special annotation `"dapp/watch": "false"` in the resource template. For example:

```yaml
...
kind: Deployment
metadata:
  ...
  annotations:
    "dapp/watch": "false"
```

**NOTICE** For now status watch is supported only for **Deployment** resource kind. ReplicaSet, DaemonSet and other resource will be supported soon.
