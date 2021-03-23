---
title: Deploy annotations
permalink: reference/deploy_annotations.html
description: Description of annotations which control werf tracking of resources during deploy process
toc: false
---

This article contains description of annotations which control werf resource operations and tracking of resources during deploy process. Annotations should be configured in the chart templates.

 - [`werf.io/replicas-on-creation`](#replicas-on-creation) — defines number of replicas that should be set only when creating resource initially (useful for HPA).
 - [`werf.io/track-termination-mode`](#track-termination-mode) — defines a condition when werf should stop tracking of the resource.
 - [`werf.io/fail-mode`](#fail-mode) — defines how werf will handle a resource failure condition which occurred after failures threshold has been reached for the resource during deploy process.
 - [`werf.io/failures-allowed-per-replica`](#failures-allowed-per-replica) — defines a threshold of failures after which resource will be considered as failed and werf will handle this situation using [fail mode](#fail-mode).
 - [`werf.io/log-regex`](#log-regex) — specifies a template for werf to show only those log lines of the resource that fit the specified regex template.
 - [`werf.io/log-regex-for-CONTAINER_NAME`](#log-regex-for-container) — specifies a template for werf to show only those log lines of the resource container that fit the specified regex template.
 - [`werf.io/skip-logs`](#skip-logs) — completely disable logs printing for the resource.
 - [`werf.io/skip-logs-for-containers`](#skip-logs-for-containers) — disable logs of specified containers of the resource.
 - [`werf.io/show-logs-only-for-containers`](#show-logs-only-for-containers) — enable logging only for specified containers of the resource.
 - [`werf.io/show-service-messages`](#show-service-messages) — enable additional logging of Kubernetes related service messages for resource.

More info about chart templates and other stuff is available in the [helm chapter]({{ "advanced/helm/overview.html" | true_relative_url }}).

## Replicas on creation

When HPA is active usage of default `spec.replicas` leads to harmful and tricky behaviour, because each time werf chart is being converged through CI/CD process, resource replicas will be reset to the static `spec.replicas` value in the chart templates, even if this value will was already changed in runtime by the HPA.

One recommended solution is to completely remove `spec.replicas` from the templates. But if you need to set replicas only when initially creating resource then you need `"werf.io/replicas-on-creation"` annotation.

`"werf.io/replicas-on-creation": "NUM"`

Defines a number of replicas which will be set for the resource on initial creation.

**NOTE** `"NUM"` should be specified as string, because annotations does not support anything but strings, any type other than string will be ignored.

## Track termination mode

`"werf.io/track-termination-mode": WaitUntilResourceReady|NonBlocking`

Defines a condition when werf should stop tracking of the resource:
 * `WaitUntilResourceReady` (default) —  the entire deployment process would monitor and wait for the readiness of the resource having this annotation. Since this mode is enabled by default, the deployment process would wait for all resources to be ready.
 * `NonBlocking` — the resource is tracked only if there are other resources that are not yet ready.

<img src="https://raw.githubusercontent.com/werf/demos/master/deploy/werf-new-track-modes-3.gif" />

**TIP** Use `"werf.io/track-termination-mode": NonBlocking` annotation when you need to describe a StatefulSet object with the `OnDelete` manual update strategy, and you don't want to block the entire deploy process while waiting for the StatefulSet update.

**TIP** Use `"werf.io/track-termination-mode": NonBlocking` and `"werf.io/fail-mode": IgnoreAndContinueDeployProcess` when you need to define a Job in the release that runs in the background and does not affect the deploy process.

## Fail mode

`"werf.io/fail-mode": FailWholeDeployProcessImmediately|HopeUntilEndOfDeployProcess|IgnoreAndContinueDeployProcess`

Defines how werf will handle a resource failure condition which occurred after failures threshold has been reached for the resource during deploy process:
 * `FailWholeDeployProcessImmediately` (default) — the entire deploy process will fail with an error if an error occurs for some resource.
 * `HopeUntilEndOfDeployProcess` — when an error occurred for the resource, set this resource into the "hope" mode, and continue tracking other resources. If all remained resources are ready or in the "hope" mode, transit the resource back to "normal" and fail the whole deploy process if an error for this resource occurs once again.
 * `IgnoreAndContinueDeployProcess` — resource errors do not affect the deployment process.

## Failures allowed per replica

`"werf.io/failures-allowed-per-replica": "NUMBER"`

By default, one error per replica is allowed before considering the whole deployment process unsuccessful. This setting defines a threshold of failures after which resource will be considered as failed and werf will handle this situation using [fail mode](#fail-mode).

## Log regex

`"werf.io/log-regex": RE2_REGEX`

Defines a [Re2 regex](https://github.com/google/re2/wiki/Syntax) template that applies to all logs of all containers of all Pods owned by a resource with this annotation. werf would show only those log lines that fit the specified regex template. By default, werf shows all log lines.

## Log regex for container

`"werf.io/log-regex-for-CONTAINER_NAME": RE2_REGEX`

Defines a [Re2 regex](https://github.com/google/re2/wiki/Syntax) template that applies to logs of `CONTAINER_NAME` containers in all Pods owned by a resource with this annotation. werf would show only those log lines that fit the specified regex template. By default, werf shows all log lines.

## Skip logs

`"werf.io/skip-logs": "true"|"false"`

Set to `"true"` to turn off printing logs of all containers of all Pods owned by a resource with this annotation. This annotation is disabled by default.

<img src="https://raw.githubusercontent.com/werf/demos/master/deploy/werf-new-track-modes-2.gif" />

## Skip logs for containers

`"werf.io/skip-logs-for-containers": CONTAINER_NAME1,CONTAINER_NAME2,CONTAINER_NAME3...`

The comma-separated list of containers in all Pods owned by a resource with this annotation. werf would turn off log output for those containers.

## Show logs only for containers

`"werf.io/show-logs-only-for-containers": CONTAINER_NAME1,CONTAINER_NAME2,CONTAINER_NAME3...`

The comma-separated list of containers in all Pods owned by a resource with this annotation. werf would show logs for these containers. Logs of containers that are not included in this list will not be printed. By default, werf displays logs of all containers of all Pods of a resource.

## Show service messages

`"werf.io/show-service-messages": "true"|"false"`

Set to `"true"` to enable additional real-time debugging info (including Kubernetes events) for a resource during tracking. By default, werf would show these service messages only if the resource has failed the entire deploy process.

<img src="https://raw.githubusercontent.com/werf/demos/master/deploy/werf-new-track-modes-1.gif" />
