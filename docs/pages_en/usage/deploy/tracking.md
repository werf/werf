---
title: Resource tracking
permalink: usage/deploy/tracking.html
---

## Tracking resource status

The resource deployment process consists of two steps: applying resources to the cluster and *tracking the state* of those resources. werf implements advanced resource state tracking (werf only) via the [kubedog](https://github.com/werf/kubedog) library.

Resource tracking is enabled by default for all supported resource types, namely:

* all release resources;

* some of the resources indirectly created by the release resources;

* off-release resources referenced in the `<name>.external-dependency.werf.io/resource` annotations.

Deployment, StatefulSet, DaemonSet, Job, and Flagger Canary resources utilize *dedicated* state trackers that not only determine accurately whether the resource was deployed successfully or unsuccessfully, but also track the status of child resources, such as Pods created by Deployment.

For other resources that do not possess *dedicated* state trackers, a *universal* tracker is used. It *predicts* the success of the resource deployment based on the information available in the cluster about the resource. In rare cases, when the universal tracker is wrong in its assumptions, tracking for that resource can be disabled.

### Changing the criteria for categorizing a resource deployment as failed (werf only)

By default, werf aborts the deployment and flags it as a failure if more than two errors occur when deploying one of the resources.

You can change the maximum number of deployment errors for a resource using the `werf.io/failures-allowed-per-replica` annotation as follows:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  annotations:
    werf.io/failures-allowed-per-replica: "5"
```

For resources annotated with `werf.io/fail-mode: HopeUntilEndOfDeployProcess`, deployment errors will only be accounted for after all other resources have been successfully deployed.

Similarly, a resource annotated with `werf.io/track-termination-mode: NonBlocking` will only be tracked until all other resources are deployed, after which that resource will automatically be seen as deployed, even if it is not.

The universal tracker treats *lack of activity* for a resource for 4 minutes as a deployment error. You can change this timeout using the `werf.io/no-activity-timeout` annotation, for example:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  annotations:
    werf.io/no-activity-timeout: 10m
```

### Disabling state tracking and ignoring resource errors (werf only)

To disable resource state tracking and ignore resource deployment errors, annotate the resource with `werf.io/fail-mode: IgnoreAndContinueDeployProcess` and `werf.io/track-termination-mode: NonBlocking` as follows:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  annotations:
    werf.io/fail-mode: IgnoreAndContinueDeployProcess
    werf.io/track-termination-mode: NonBlocking
```

## Displaying container logs (werf only)

Thanks to the [kubedog](https://github.com/werf/kubedog) library, werf can automatically display logs of the containers created as part of Deployment, StatefulSet, DaemonSet, and Job objects.

You can disable the log display for the resource via the `werf.io/skip-logs: "true"` annotation, for example:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  annotations:
    werf.io/skip-logs: "true"
```

Alternatively, you can explicitly list the containers whose logs should be displayed using the `werf.io/show-logs-only-for-containers` annotation; in this case, the logs of all other containers will be hidden, for example:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  annotations:
    werf.io/show-logs-only-for-containers: "backend,frontend"
```

... or the other way around — list the containers whose logs *shouldn't* be displayed in the `werf.io/skip-logs-for-containers` annotation; the logs of all other containers will be displayed, for example:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  annotations:
    werf.io/skip-logs-for-containers: "sidecar"
```

The `werf.io/log-regex` annotation allows you to filter and display only those log lines that match a regular expression, for example:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  annotations:
    werf.io/log-regex: ".*ERROR.*"
```

The `werf.io/log-regex-for-<container name>` annotation allows you to filter the log lines based on a regular expression for a specific container, rather than for all containers at once, for example:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  annotations:
    werf.io/log-regex-for-backend: ".*ERROR.*"
```

## Displaying Events resources (werf only)

Thanks to the [kubedog](https://github.com/werf/kubedog) library, werf can display Events of tracked resources if the resource has the `werf.io/show-service-messages: "true"` annotation, for example:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  annotations:
    werf.io/show-service-messages: "true"
```
