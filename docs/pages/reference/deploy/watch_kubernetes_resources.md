---
title: Watch kubernetes resources
sidebar: reference
permalink: reference/deploy/watch_kubernetes_resources.html
author: Timofey Kirillov <timofey.kirillov@flant.com>
---

Werf can watch statuses and logs of kubernetes resources during the deploy process.

## Watch hook logs

Werf can print helm hook logs ([more about helm hooks](https://github.com/helm/helm/blob/master/docs/charts_hooks.md)).

To enable logs special annotation `"werf/watch": "true"` should be specified in the resource template. For example:

```yaml
...
kind: Job
metadata:
  ...
  annotations:
    "helm.sh/hook": post-install,post-upgrade
    "helm.sh/hook-weight": "1"
    "werf/watch": "true"
```

Logs of pods will be printed for a job until job terminated.

**NOTICE** For now logs watch is supported only for **Job** resource kind.

## Watch release status

After helm, a release has been installed, or updated werf will watch all Deployments that existed in the chart.

**Status watch procedure** consists of:

1. Periodical polling of the resource status and printing info about resource readiness.
2. Periodical printing of the resource logs and events.

The procedure is ended when resource readiness condition is reached.

This is **default behavior** for any Deployment found in the chart. To turn it off for specific deployment add special annotation `"werf/watch": "false"` in the resource template. For example:

```yaml
...
kind: Deployment
metadata:
  ...
  annotations:
    "werf/watch": "false"
```

**NOTICE** For now status watch is supported only for **Deployment** resource kind. ReplicaSet, DaemonSet, and other resources will be supported soon.
