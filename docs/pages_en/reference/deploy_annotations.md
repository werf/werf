---
title: Deploy annotations
permalink: reference/deploy_annotations.html
description: Description of annotations which control werf tracking of resources during deploy process
toc: false
---

This article contains description of annotations which control werf resource operations and tracking of resources during deploy process. Annotations should be configured in the chart templates.

 - [`werf.io/weight`](#resource-weight) — defines the weight of the resource, which will affect the order in which the resources are deployed.
 - [`werf.io/deploy-dependency-ANY_NAME`](#resource-dependencies) — define a dependency for the resource, which will affect the order in which the resources are deployed.
 - [`<any-name>.external-dependency.werf.io/resource`](#external-dependency-resource) — wait for specified external dependency to be up and running, and only then proceed to deploy the annotated resource.
 - [`<any-name>.external-dependency.werf.io/namespace`](#external-dependency-namespace) — specify the namespace for the external dependency.
 - [`werf.io/ownership`](#resource-ownership) — defines how resource deletions are handled and how release annotations are managed.
 - [`werf.io/deploy-on`](#conditional-resource-deployment) — defines when to render the resource for the deployment and on which stages should it be deployed.
 - [`werf.io/delete-policy`](#resource-delete-policy) — defines how resource deletions should be handled during resource deployment.
 - [`werf.io/replicas-on-creation`](#replicas-on-creation) — defines number of replicas that should be set only when creating resource initially (useful for HPA).
 - [`werf.io/track-termination-mode`](#track-termination-mode) — defines a condition when werf should stop tracking of the resource.
 - [`werf.io/fail-mode`](#fail-mode) — defines how werf will handle a resource failure condition which occurred after failures threshold has been reached for the resource during deploy process.
 - [`werf.io/failures-allowed-per-replica`](#failures-allowed-per-replica) — defines a threshold of failures after which resource will be considered as failed and werf will handle this situation using [fail mode](#fail-mode).
 - [`werf.io/ignore-readiness-probe-fails-for-CONTAINER_NAME`](#ignore-readiness-probe-failures-for-container) — override automatically calculated ignore period during which readiness probe failures will not mark resource as failed.
 - [`werf.io/no-activity-timeout`](#no-activity-timeout) — change inactivity period after which the resource will be marked as failed.
 - [`werf.io/log-regex`](#log-regex) — specifies a template for werf to show only those log lines of the resource that fit the specified regex template.
 - [`werf.io/log-regex-for-CONTAINER_NAME`](#log-regex-for-container) — specifies a template for werf to show only those log lines of the resource container that fit the specified regex template.
 - [`werf.io/log-regex-skip`](#log-regex-skip) — specifies a template for werf to skip log lines of the resource that match the specified regex template.
 - [`werf.io/skip-logs`](#skip-logs) — completely disable logs printing for the resource.
 - [`werf.io/skip-logs-for-containers`](#skip-logs-for-containers) — disable logs of specified containers of the resource.
 - [`werf.io/show-logs-only-for-number-of-replicas`](#show-logs-only-for-number-of-replicas) — enable logging only for the specified number of replicas of the resource.
 - [`werf.io/show-logs-only-for-containers`](#show-logs-only-for-containers) — enable logging only for specified containers of the resource.
 - [`werf.io/show-service-messages`](#show-service-messages) — enable additional logging of Kubernetes related service messages for resource.
 - [`werf.io/sensitive`](#mark-resource-as-sensitive) — mark the resource as sensitive, so werf will not show diffs for this resource in `werf plan`.
 - [`werf.io/sensitive-paths`](#mark-fields-of-a-resource-as-sensitive) — mark the fields of a resource as sensitive, so werf will not show diffs for this resource in `werf plan`.

More info about chart templates and other stuff is available in the [helm chapter]({{ "usage/deploy/overview.html" | true_relative_url }}).

## Resource weight

`werf.io/weight: "NUM"`

Example: \
`werf.io/weight: "10"` \
`werf.io/weight: "-10"`

- Can be a positive number, a negative number, or a zero. 
- Value is passed as a string. If omitted, the `weight` is set to 0 by default. 
- Works for non-Hook resources only. For Hooks, use `helm.sh/hook-weight`, which works almost the same.

This parameter sets the weight of the resources, defining the order in which they are deployed. First, werf groups resources according to their weight and then sequentially deploys them, starting with the group with the lowest weight. In this case, werf will not proceed to deploy the next batch of resources until the previous group has been successfully deployed.

More info: [deployment order]({{ "/usage/deploy/deployment_order.html" | true_relative_url }})

## Resource dependencies

`werf.io/deploy-dependency-ANY_NAME: state=STATE[,name=NAME][,namespace=NAMESPACE][,kind=KIND][,group=GROUP][,version=VERSION]`

Example: \
`werf.io/deploy-dependency-db: state=ready,kind=StatefulSet,name=postgres` \
`werf.io/deploy-dependency-app: state=present,kind=Deployment,group=apps,version=v1,name=app,namespace=app`

Required parameters:
- `state`: `ready` or `present`. If `present`, then wait until resource is created/updated, if `ready`, then wait until resource is created/updated and ready.

At least one of these parameters must be specified:
- `name`: name of a resource to depend on.
- `namespace`: namespace of a resource to depend on.
- `kind`: kind of a resource to depend on.
- `group`: api group of a resource to depend on.
- `version`: api version of a resource to depend on.

More info: [deployment order]({{ "/usage/deploy/deployment_order.html" | true_relative_url }})

## External dependency resource

`<any-name>.external-dependency.werf.io/resource: type[.version.group]/name`

Example: \
`secret.external-dependency.werf.io/resource: secret/config` \
`someapp.external-dependency.werf.io/resource: deployments.v1.apps/app`

Sets the external dependency for the resource. The annotated resource won't be deployed until the external dependency has been created and ready.

More info: [deployment order]({{ "/usage/deploy/deployment_order.html" | true_relative_url }})

## External dependency namespace

`<any-name>.external-dependency.werf.io/namespace: name`

Sets the namespace for the external dependency specified by the [external dependency resource](#external-dependency-resource) annotation. The `<any-name>` prefix must be the same as in the annotation of the external dependency resource.

More info: [deployment order]({{ "/usage/deploy/deployment_order.html" | true_relative_url }})

## Resource ownership

`werf.io/ownership: release|anyone`

Defines how resource deletions are handled and how release annotations are managed. Allowed values are:
 * `release`: the resource is deleted if removed from the chart or when the release is uninstalled, and release annotations of the resource are applied/validated during deploy.
 * `anyone`: the opposite of `release`, the resource is never deleted on uninstall or when removed from the chart, and release annotations are not applied/validated during deploy.

General resources have `release` ownership by default, while hooks and CRDs from the `crds` directory have `anyone` ownership.

## Conditional resource deployment

`werf.io/deploy-on: pre-install|install|post-install|pre-upgrade|upgrade|post-upgrade|pre-rollback|rollback|post-rollback|pre-delete|delete|post-delete`

Define when to render the resource for the deployment and on which stages should it be deployed. Available values are:
 * `pre-install`, `install`, `post-install` — render only on release installation
 * `pre-upgrade`, `upgrade`, `post-upgrade` — render only on release upgrade
 * `pre-rollback`, `rollback`, `post-rollback` — render only on release rollback
 * `pre-delete`, `delete`, `post-delete` — render only on release deletion

The default value for general resources is `install,upgrade,rollback`, while for hooks it is populated from `helm.sh/hook`.

## Resource delete policy

`werf.io/delete-policy: before-creation|before-creation-if-immutable|succeeded|failed`

The `werf.io/delete-policy` annotation controls resource deletions during its deployment and is inspired by `helm.sh/hook-delete-policy`. Allowed values:
* `before-creation`: the resource is always recreated
* `before-creation-if-immutable`: the resource is recreated only if we got the `field is immutable` error trying to update the resource
* `succeeded`: the resource is deleted after its readiness check succeeds
* `failed`: the resource is deleted if its readiness check fails

By default, general resources have no delete policy, while hooks have values from `helm.sh/hook-delete-policy` mapped to `werf.io/delete-policy`.

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

## Ignore readiness probe failures for container

`"werf.io/ignore-readiness-probe-fails-for-CONTAINER_NAME": "TIME"`

This annotation allows to override the result of automatically calculated period during which readiness probe failures
will be ignored and won't fail the rollout. After this period the first failed readiness probe fails the rollout. By
default, the ignore period is automatically calculated based on readiness probe configuration. Note that
if `failureThreshold: 1` specified in the probe configuration then the first received failed probe will fail the
rollout, regardless of ignore period.

The value format is specified [here](https://pkg.go.dev/time#ParseDuration).

Example:
`"werf.io/ignore-readiness-probe-fails-for-backend": "20s"`

## No activity timeout

`werf.io/no-activity-timeout: "TIME"`

Default: `4m`

Example: \
`werf.io/no-activity-timeout: "8m30s"` \
`werf.io/no-activity-timeout: "90s"`

If no new events or resource updates received in `TIME` then the resource will be marked as failed.

The value format is specified [here](https://pkg.go.dev/time#ParseDuration).

## Log regex

`"werf.io/log-regex": RE2_REGEX`

Defines a [Re2 regex](https://github.com/google/re2/wiki/Syntax) template that applies to all logs of all containers of all Pods owned by a resource with this annotation. werf would show only those log lines that fit the specified regex template. By default, werf shows all log lines.

## Log regex for container

`"werf.io/log-regex-for-CONTAINER_NAME": RE2_REGEX`

Defines a [Re2 regex](https://github.com/google/re2/wiki/Syntax) template that applies to logs of `CONTAINER_NAME` containers in all Pods owned by a resource with this annotation. werf would show only those log lines that fit the specified regex template. By default, werf shows all log lines.

## Log regex skip

`"werf.io/log-regex-skip": RE2_REGEX`

Defines a [Re2 regex](https://github.com/google/re2/wiki/Syntax) template that applies to all logs of all containers of all Pods owned by a resource with this annotation. werf would skip log lines that match the specified regex template. By default, werf shows all log lines.

## Skip logs

`"werf.io/skip-logs": "true"|"false"`

Set to `"true"` to turn off printing logs of all containers of all Pods owned by a resource with this annotation. This annotation is disabled by default.

<img src="https://raw.githubusercontent.com/werf/demos/master/deploy/werf-new-track-modes-2.gif" />

## Skip logs for containers

`"werf.io/skip-logs-for-containers": CONTAINER_NAME1,CONTAINER_NAME2,CONTAINER_NAME3...`

The comma-separated list of containers in all Pods owned by a resource with this annotation. werf would turn off log output for those containers.

## Show logs only for number of replicas

`"werf.io/show-logs-only-for-number-of-replicas": "NUMBER"`

Print logs only for the specified number of replicas during resource tracking. We print logs only for a single replica by default to avoid excessive log output and to optimize resource usage.

## Show logs only for containers

`"werf.io/show-logs-only-for-containers": CONTAINER_NAME1,CONTAINER_NAME2,CONTAINER_NAME3...`

The comma-separated list of containers in all Pods owned by a resource with this annotation. werf would show logs for these containers. Logs of containers that are not included in this list will not be printed. By default, werf displays logs of all containers of all Pods of a resource.

## Show service messages

`"werf.io/show-service-messages": "true"|"false"`

Set to `"true"` to enable additional real-time debugging info (including Kubernetes events) for a resource during tracking. By default, werf would show these service messages only if the resource has failed the entire deploy process.

<img src="https://raw.githubusercontent.com/werf/demos/master/deploy/werf-new-track-modes-1.gif" />

## Mark fields of a resource as sensitive

`"werf.io/sensitive-paths": "JSONPath,JSONPath,..."`

Example: \
`"werf.io/sensitive-paths": "$.spec.template.spec.containers[*].env[*].value,$.data.*"`

Don't show diffs for resource fields that match specified JSONPath expressions. Overrides the behavior of `werf.io/sensitive`.

## Mark resource as sensitive

`"werf.io/sensitive": "true"|"false"`

Set to `"true"` to mark the resource as sensitive, so werf will not show diffs for this resource in `werf plan`. By default, werf shows diffs for all resources, except Secrets.
