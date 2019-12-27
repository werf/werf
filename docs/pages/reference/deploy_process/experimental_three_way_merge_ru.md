---
title: "Experimental: three way merge"
sidebar: documentation
permalink: ru/documentation/reference/deploy_process/experimental_three_way_merge.html
ref: documentation_reference_deploy_process_experimental_three_way_merge
lang: ru
author: Timofey Kirillov <timofey.kirillov@flant.com>
---

> Three way merge is under development now. This page contains implementation notes and other debug and development info about resources update method which werf currently uses.

Three way merge is a way of applying changes to Kubernetes resources which uses previous resource configuration, new resource configuration and current resource state to calculate and then apply a three-way-merge patch with new changes for for each resource. This method used by the `kubectl apply` command.

werf is fully compatible with the Helm and currently makes use of two-way-merge patches as Helm does. This method implies creating patches between previous chart resource configuration and a new one.

Two-way-merge method have a problem: chart resource configuration gets out of sync with the current resource state when user suddenly makes changes to the current configuration (via `kubectl edit` command for example).

## Migration to three-way-merge

werf currently migrating from two-way-merge to three-way-merge method. Migration consists of 3 steps named:

 1. Annotations mode.
 2. Compatibility mode.
 3. Full mode.

Currently annotations mode is active: three-way-merge repair patch only created and written to resource annotation and not applied during deploy.

### Annotations mode

In the annotations mode werf still applies two-way-merge patches. The following annotations are set into each chart resource by the werf when updating these resources:

 * `debug.werf.io/repair-patch`;
 * `debug.werf.io/repair-patch-errors`.

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

Also during deploy process a WARNING message will be written to the screen if repair-patch needs to be applied by administrator.

Annotation `debug.werf.io/repair-patch-errors` will contain errors occurred when creating a repair patch, deploy process in this case will not be interrupted.

### Compatibility mode

In the compatibility mode werf will apply three-way-merge patches only to the new resources created by the werf and use old two-way-merge patches for already existing resources.

### Full mode

In the full mode werf will apply three-way-merge patches to all resources.
