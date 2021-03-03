---
title: Steps
permalink: documentation/advanced/helm/deploy_process/steps.html
---

When running the `werf converge` command, werf starts the deployment process that includes the following steps:

1. Rendering chart templates into a single list of Kubernetes resource manifests and validating them.
2. Running `pre-install` or `pre-upgrade` [hooks]({{ "/documentation/advanced/helm/deploy_process/helm_hooks.html" | true_relative_url }}) and tracking each hook until successful or failed termination; printing logs and other information in the process.
3. Applying changes to Kubernetes resources: creating new, deleting old, updating the existing.
4. Creating a new release version and saving the current state of resource manifests into this release’s data.
5. Tracking all release resources until the readiness state is reached; printing logs and other information in the process.
6. Running `post-install` or `post-upgrade` [hooks]({{ "/documentation/advanced/helm/deploy_process/helm_hooks.html" | true_relative_url }}) and tracking each hook until successful or failed termination; printing logs and other information in the process.

**NOTE** werf would delete all newly created resources immediately during the ongoing deploy process if this process fails at any of the steps described above!

When executing helm hooks at the step 2 and 6, werf would track these hooks resources until successful termination. Tracking [can be configured](#resource-tracking) for each hook resource.

## Resource tracking

On step 5, werf would track all release resources until each resource reaches the "ready" state. All resources are tracked simultaneously. During tracking, werf aggregates information obtained from all release resources into the single text output in real-time, and periodically prints the so-called status progress table.

werf displays logs of resource Pods until those pods reach the "ready" state. In the case of Job pods, logs are shown until Pods are terminated.

werf uses the [kubedog library](https://github.com/werf/kubedog) to track resources. Currently, tracking is implemented for Deployments, StatefulSets, DaemonSets, and Jobs. Support for tracking Service, Ingress, PVC, and other resources [will be added ](https://github.com/werf/werf/issues/1637).

Tracking behaviour can be configured for each resource using [resource annotations]({{ "/documentation/reference/deploy_annotations.html" | true_relative_url }}), which should be set in the chart templates.

## If the deploy failed

In the case of failure during the release process, werf would create a new release having the FAILED state. This state can then be inspected by the user to find the problem and solve it on the next deploy invocation.

## Multiple Kubernetes clusters

There are cases when separate Kubernetes clusters are required for a different environments. You can [configure access to multiple clusters](https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters) using kube contexts in a single kube config.

In that case, the `--kube-context=CONTEXT` deploy option should be set manually along with the environment.

## Subcharts

During deploy process werf will render, create and track all resources of all [subcharts]({{ "/documentation/advanced/helm/configuration/chart_dependencies.html" | true_relative_url }}).
