---
title: Track kubernetes resources
sidebar: reference
permalink: reference/deploy/track_kubernetes_resources.html
author: Timofey Kirillov <timofey.kirillov@flant.com>
---

Werf has two types of tracking for kubernetes resources during deploy process:

1. Track [Helm Hooks](https://github.com/helm/helm/blob/master/docs/charts_hooks.md).
2. Track regular Helm Chart resources readiness.

Werf uses [kubedog](https://github.com/flant/kubedog) library to track resources under the hood.

## Track Helm Hooks

Werf can optionally track statuses and logs of Helm Hooks. Only Jobs Helm Hooks tracking is supported for now (other resources tracking already implemented and will be added soon).

To enable logs special annotation `"werf.io/track": "true"` should be specified in the resource template. For example:

```yaml
...
kind: Job
metadata:
  ...
  annotations:
    "helm.sh/hook": post-install,post-upgrade
    "helm.sh/hook-weight": "1"
    "werf.io/track": "true"
```

Logs of pods will be printed for a job until job terminated.

This feature is useful to print logs of migrations, which runs in helm hook jobs.

## Track regular Helm Chart resources readiness

Track statuses and logs of all chart resources after helm release installed or updated till every resource become ready to serve.

The following Kubernetes resources are supported:
* Pod;
* Deployment;
* StatefulSet;
* DaemonSet;
* Job.

Werf tracks resources kinds in the specified order one by one. Within the same resource kind resources are sorted in the order of definition.

This tracking procedure is the *default behavior* for any supported resource found in the chart. To turn it off for specific resource add special annotation `"werf.io/track": "false"` into the resource template. For example:

```yaml
...
kind: Deployment
metadata:
  ...
  annotations:
    "werf.io/track": "false"
```
