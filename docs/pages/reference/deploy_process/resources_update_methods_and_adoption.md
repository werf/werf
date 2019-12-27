---
title: Resources update methods and adoption
sidebar: documentation
permalink: documentation/reference/deploy_process/resources_update_methods_and_adoption.html
author: Timofey Kirillov <timofey.kirillov@flant.com>
---

> Three way merge is under development now. This page contains implementation notes and other debug and development info about resources update method which werf currently uses

Three way merge is a way of applying changes to Kubernetes resources which uses previous resource configuration, new resource configuration and current resource state to calculate and then apply a three-way-merge patch with new changes for for each resource. This method used by the `kubectl apply` command.

werf is fully compatible with the Helm 2, which makes use of two-way-merge patches. This method implies creating patches between previous chart resource configuration and a new one.

Two-way-merge method have a problem: chart resource configuration gets out of sync with the current resource state when user suddenly makes changes to the current resource state (via `kubectl edit` command for example). For example if you have `replicas=5` in the chart template and set `replicas=3` in the live resource state (by `kubectl edit` command), then subsequent werf deploy calls will not set replicas to 5 as expected, because two-way-merge patches does not take into account current live resource state.

## Resources update methods

werf currently migrating from two-way-merge to three-way-merge method. Following methods of resources update are avaiable:

 1. Two-way-merge and repair patches.
 2. Three-way-merge patches.

### Two-way-merge and repair patches

With this method werf still applies two-way-merge patches. The following annotations are set into each chart resource by the werf when updating these resources:
 * `debug.werf.io/repair-patch`;
 * `debug.werf.io/repair-patch-errors`;
 * `debug.werf.io/repair-messages`.

Annotation `debug.werf.io/repair-patch` contains calculated patch json between current resource state and new chart resource configuration.

This patch can be manually applied to repair current resource state to be in sync with the desired chart resource configuration.

For example, if resource contains the following repair patch:

```yaml
metadata:
  annotations:
    "debug.werf.io/repair-patch": '{"data":{"node.conf":"PROPER CONTENT"}}'
```

Then user can repair resource state with the following command:

```shell
kubectl -n mynamespace patch cm/myconfigmap '{"data":{"node.conf":"PROPER CONTENT"}}'
```

Also during deploy process a WARNING message will be written to the screen if repair-patch needs to be applied by an administrator.

Annotation `debug.werf.io/repair-patch-errors` will contain errors occurred when creating a repair patch, deploy process in this case will not be interrupted.

Annotation `debug.werf.io/repair-messages` will contain warnings written to the deploy log along with debug developer messages.

### Three-way-merge patches

With this method werf will apply three-way-merge patches which take into accout resource version from the previous release, resource version from current chart template and live resource version from cluster to calculate a patch. This patch should transform resource state to match chart template configuration.

## Selecting resources update method

To select resources update method werf has an option called `WERF_THREE_WAY_MERGE_MODE` (or `--three-way-merge-mode` cli option), which can have one of the following values:

 - "disabled" — do not use three-way-merge patches during updates neither for already existing nor new releases, use two-way-merge and repair patches method for all releases;
 - "enabled" — always use three-way-merge patches method during updates for already existing and new releases;
 - "onlyNewReleases" — new releases created since that mode is active will use three-way-merge patches method during updates, while already existing releases continue to use old helm two-way-merge and repair patches method.

Default resource update method before 01.12.2019 is `disabled`.

After 01.12.2019 default mode will be `onlyNewReleases`.

After 15.12.2019 default mode will be `enabled`.

After 01.03.2020 there will be no `WERF_THREE_WAY_MERGE_MODE` option to disable three-way-merge, werf will always use three-way-merge patches method.

Default update method could be redefined with `WERF_THREE_WAY_MERGE_MODE` option.

Already existing release can be switched to use three-way-merge patches method by setting `WERF_THREE_WAY_MERGE_MODE=enabled`.

The release that already uses three-way-merge patches method can be switched back to two-way-merge and repair patches by setting `WERF_THREE_WAY_MERGE_MODE=disabled`.

## Deal with HPA

[The Horizontal Pod Autoscaler](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/) automatically scales the number of pods in a replication controller, deployment or replica set.

When HPA is enabled and user has set `spec.replicas` to static number in chart templates, then repair patch may detect a change in the live state of the object `spec.replicas` field, because this field has been changed by the autoscaler. To disable such repair patches user may either delete `spec.replicas` from chart configuration or define [`werf.io/set-replicas-only-on-creation` annotation](#werf-io-set-replicas-only-on-creation).

#### `werf.io/set-replicas-only-on-creation`

```yaml
"werf.io/set-replicas-only-on-creation": "true"
```

This annotation tells werf that resource `spec.replicas` field value from chart template should only be used when initializing a new resource.

werf will ignore `spec.replicas` field changes on subsequent resource updates.

This annotation should be turned on when [HPA](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/) is enabled.

## Deal with VPA

[The Vertical Pod Autoscaler](https://cloud.google.com/kubernetes-engine/docs/concepts/verticalpodautoscaler) can recommend values for CPU and memory requests, or it can automatically update values for CPU and memory requests.

When VPA is enabled and user has set `resources` to some static settings in chart templates, then repair patch may detect a change in the live state of the obje ct `resources` field, because this field has been changed by the autoscaler. To disable such repair patches user may either delete `resources` settings from chart configuration or define [`werf.io/set-resources-only-on-creation` annotation](#werf-io-set-resources-only-on-creation).

#### `werf.io/set-resources-only-on-creation`

```yaml
"werf.io/set-resources-only-on-creation": "true"
```

This annotation tells werf that resource `resources` field value from chart template should only be used when initializing a new resource.

werf will ignore `resources` field changes on subsequent resource updates.

This annotation should be turned on when [VPA](https://cloud.google.com/kubernetes-engine/docs/concepts/verticalpodautoscaler) is enabled.

## Resources adoption

There are 2 cases when resource defined in the chart already exists in the cluster.

### Installing new release

When installing a new release a resource can already exists in the cluster. In this case werf will exit with an error. A resource that already exists will not be marked as adopted into the release. New release will be installed in the FAILED state and deleting this failed release will not cause werf to delete not adopted during release installation resources.

### Updating release

User can define a new resource in the chart templates of already existing release. In this case werf will also exit with an error. A resource that already exists will not be marked as adopted into the release.

### How to adopt resource

To allow adoption of already existing resource into werf release set `werf.io/allow-adoption-by-release=RELEASENAME` annotation to the resource manifest in the chart and run werf deploy process. During adoption werf will generate a three-way-merge patch to bring existing resource to the state that is defined in the chart.

NOTE that `WERF_THREE_WAY_MERGE_MODE` setting does not affect resources adoption, three-way-merge patch will be used anyway during adoption process.

#### Three way merge patches and adoption

Three way merge patches are always used when adopting already existing resource into the release (`WERF_THREE_WAY_MERGE_MODE` setting does not affect resources adopter).

**NOTE** After adoption live resource manifest may not fully match resource manifest in the chart. In the cases when additional fields are defined in the live resource — these fields will not be deleted and stay in the live resource version.

For example, let's say we have an already existing Deployment in the cluster with the following spec:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mydeploy1
  annotations:
    extra: data
  labels:
    service: mydeploy1
spec:
  replicas: 1
  selector:
    matchLabels:
      service: mydeploy1
  template:
    metadata:
      labels:
        mylabel: myvalue
        service: mydeploy1
    spec:
      containers:
      - name: main
        command: [ "/bin/bash", "-c", "while true; do date ; sleep 1 ; done" ]
        image: ubuntu:18.04
      - name: additional
        command: [ "/bin/bash", "-c", "while true; do date ; sleep 1 ; done" ]
        image: ubuntu:18.04
```

And following Deployment is defined in the chart templates:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mydeploy1
  labels:
    service: mydeploy1
spec:
  replicas: 2
  selector:
    matchLabels:
      service: mydeploy1
  template:
    metadata:
      labels:
        service: mydeploy1
    spec:
      containers:
      - name: main
        command: [ "/bin/bash", "-c", "while true; do date ; sleep 1 ; done" ]
        image: ubuntu:18.04
```

After adoption werf will change `replicas=1` to `replicas=2`. Extra data `metadata.annotations.extra=data`, `spec.template.metadata.labels.mylable=myvalue` and `spec.template.spec.containers[]` container named `additional` will stay in the final live resource version, because these fields are not defined in the chart template. User should remove such fields manually if it is needed.
