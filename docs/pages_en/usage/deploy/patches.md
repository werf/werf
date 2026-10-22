---
title: Resource patches
permalink: usage/deploy/patches.html
---

## Overview

werf has two kinds of patches, each for a different problem:
* Render patches change what is deployed, similar to Helm's `--post-renderer`. Use them to set fields a chart doesn't expose in its values, without forking it.
* Diff patches change only what werf compares with the cluster when deciding whether to update a resource. Use them to stop a resource from being updated because of fields that change on every render or are modified in the cluster.

To add patches, create `patches.yaml` in the root of the chart or pass a file with `--patches`; a patch itself is a [jq](https://jqlang.org/manual/) program.

## Changing resources without editing the chart (werf only)

To change a rendered resource without editing its template — for example, one of a dependent chart whose values don't expose the field — add a render patch to `patches.yaml` in the root of the chart.

Example:

```yaml
# .helm/patches.yaml:
renderPatches:
- match:
    kinds: [Deployment, StatefulSet]
    charts: [redis]
  patch: .spec.template.spec.priorityClassName = $Values.redisPriorityClass
```

```shell
werf render --set redisPriorityClass=high-priority
```

Output:

```yaml
# Source: myapp/charts/redis/templates/statefulset.yaml
apiVersion: apps/v1
kind: StatefulSet
# ...
spec:
  template:
    spec:
      priorityClassName: high-priority
```

Here, every Deployment and StatefulSet of the `redis` chart gets the priority class from the `redisPriorityClass` value of the main chart: `$Values` holds the values of the chart that ships `patches.yaml`, not of `redis`.

## Ignoring field changes that don't need a deploy (werf only)

To stop a resource from being updated only because a field differs on every render (timestamps, random values) or is changed in the cluster by a controller, add a diff patch that removes the field.

Example:

```yaml
# .helm/patches.yaml:
diffPatches:
- match:
    kinds: [Deployment]
  patch: del(.spec.template.metadata.annotations."example.com/deployed-at")
```

Here, a change of the `example.com/deployed-at` annotation alone no longer updates the Deployment and restarts its Pods. If other fields change too, the Deployment is updated, and the new annotation reaches the cluster along with them.

## Patch sources

By default, werf takes patches from `patches.yaml` of every chart. To apply patches in one environment only, or to patch a chart you don't own, pass a file of the same format with `--patches` (or `WERF_PATCHES_*`):

```shell
werf converge --patches .helm/patches-production.yaml
```

Patches are applied in order, each to the result of the previous one:
1. `patches.yaml` of dependent charts;
2. `patches.yaml` of the main chart;
3. `--patches` files.

The sources differ in:
* What they apply to: `patches.yaml` — only resources of its chart, including its dependent charts; `--patches` — all resources of the release.
* Whose data they see (render patches only): `patches.yaml` — `$Values`, `$Release`, `$Chart` and `$Capabilities` of its chart; `--patches` — of the main chart.

To ignore `patches.yaml` of all charts, pass `--no-default-patches`; it doesn't affect `--patches` files.

`werf rollback` and `werf dismiss` apply only diff patches, and take `patches.yaml` from the chart stored in the release.

## patches.yaml format

```
renderPatches:
- match:
    names: [<name or /regexp/>, ...]
    namespaces: [<namespace or /regexp/>, ...]
    groups: [<group or /regexp/>, ...]
    versions: [<version or /regexp/>, ...]
    kinds: [<kind or /regexp/>, ...]
    charts: [<chart name, alias, path or /regexp/>, ...]
    labels: {<key>: <value>, ...}
    annotations: {<key>: <value>, ...}
  patch: <jq program>
diffPatches:
- <same as renderPatches>
```

A patch applies to every resource that matches all specified `match` fields:
* a value in `/.../` is a regular expression that must match the whole value; other values are compared literally, case-insensitively for `kinds`, `groups` and `versions`;
* `labels` and `annotations` match if the resource has all of them with the same values;
* `charts` matches the alias of the chart the resource is rendered from, or its name when no alias is set, or its path from the main chart, e.g. `myapp/charts/redis`.

A patch receives the whole resource and must return exactly one object, so it can't add or remove resources.

Example:

```yaml
# .helm/patches.yaml:
renderPatches:
- match:
    kinds: [Deployment]
    names: [/backend-.*/]
    labels:
      app: backend
  patch: .metadata.labels.team = $Values.team
```

Here, only Deployments whose name starts with `backend-` and which carry the `app: backend` label get the `team` label.
