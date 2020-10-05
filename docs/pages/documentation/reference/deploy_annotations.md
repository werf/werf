---
title: Deploy annotations
permalink: documentation/reference/deploy_annotations.html
sidebar: documentation
---

This article contains description of annotations which control werf tracking of the resources during deploy process. All annotations is supposed to be set directly in the chart templates of the resources.

More info about chart templates and other stuff is available in the [deploy basics article.]({{ site.baseurl }}/documentation/deploy/basics.html)

**TIP** Use `"werf.io/track-termination-mode": NonBlocking` and `"werf.io/fail-mode": IgnoreAndContinueDeployProcess` when you need to define a Job in the release that runs in the background and does not affect the deploy process.

**TIP** Use `"werf.io/track-termination-mode": NonBlocking` annotation when you need to describe a StatefulSet object with the `OnDelete` manual update strategy, and you don't want to block the entire deploy process while waiting for the StatefulSet update.

See also [demo examples]({{ site.baseurl }}/documentation/deploy/basics.html#configuring-resource-tracking) of using annotations.

## Track termination mode

`"werf.io/track-termination-mode": WaitUntilResourceReady|NonBlocking`

 * `WaitUntilResourceReady` (default) —  the entire deployment process would monitor and wait for the readiness of the resource having this annotation. Since this mode is enabled by default, the deployment process would wait for all resources to be ready.
 * `NonBlocking` — the resource is tracked only if there are other resources that are not yet ready.

## Fail mode

`"werf.io/fail-mode": FailWholeDeployProcessImmediately|HopeUntilEndOfDeployProcess|IgnoreAndContinueDeployProcess`

 * `FailWholeDeployProcessImmediately` (default) — the entire deploy process will fail with an error if an error occurs for some resource.
 * `HopeUntilEndOfDeployProcess` — when an error occurred for the resource, set this resource into the "hope" mode, and continue tracking other resources. If all remained resources are ready or in the "hope" mode, transit the resource back to "normal" and fail the whole deploy process if an error for this resource occurs once again.
 * `IgnoreAndContinueDeployProcess` — resource errors do not affect the deployment process.

## Failures allowed per replica

`"werf.io/failures-allowed-per-replica": "NUMBER"`

By default, one error per replica is allowed before considering the whole deployment process unsuccessful. This setting is related to the [fail mode](#fail-mode): it defines a threshold, after which the fail mode comes into play.

## Log regex

`"werf.io/log-regex": RE2_REGEX`

Defines a [Re2 regex](https://github.com/google/re2/wiki/Syntax) template that applies to all logs of all containers of all Pods owned by a resource with this annotation. werf would show only those log lines that fit the specified regex template. By default, werf shows all log lines.

## Log regex for container

`"werf.io/log-regex-for-CONTAINER_NAME": RE2_REGEX`

Defines a [Re2 regex](https://github.com/google/re2/wiki/Syntax) template that applies to logs of `CONTAINER_NAME` containers in all Pods owned by a resource with this annotation. werf would show only those log lines that fit the specified regex template. By default, werf shows all log lines.

## Skip logs

`"werf.io/skip-logs": "true"|"false"`

Set to `"true"` to turn off printing logs of all containers of all Pods owned by a resource with this annotation. This annotation is disabled by default.

## Skip logs for containers

`"werf.io/skip-logs-for-containers": CONTAINER_NAME1,CONTAINER_NAME2,CONTAINER_NAME3...`

The comma-separated list of containers in all Pods owned by a resource with this annotation. werf would turn off log output for those containers.

## Show logs only for containers

`"werf.io/show-logs-only-for-containers": CONTAINER_NAME1,CONTAINER_NAME2,CONTAINER_NAME3...`

The comma-separated list of containers in all Pods owned by a resource with this annotation. werf would show logs for these containers. Logs of containers that are not included in this list will not be printed. By default, werf displays logs of all containers of all Pods of a resource.

## Show service messages

`"werf.io/show-service-messages": "true"|"false"`

Set to `"true"` to enable additional real-time debugging info (including Kubernetes events) for a resource during tracking. By default, werf would show these service messages only if the resource has failed the entire deploy process.