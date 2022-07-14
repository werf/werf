---
title: Steps
permalink: advanced/helm/deploy_process/steps.html
---

When running the `werf converge` command, werf starts the deployment process that includes the following steps:

1. Rendering chart templates into a single list of Kubernetes resource manifests and validating them.
2. Sequentially running `pre-install` or `pre-upgrade` [hooks]({{ "/advanced/helm/deploy_process/helm_hooks.html" | true_relative_url }}) sorted by their weight and tracking each hook to completion while printing its logs in the process.
3. Grouping non-hook Kubernetes resources by their [weight]({{ "/advanced/helm/deploy_process/deployment_order.html" | true_relative_url }}) and deploying each group sequentially according to its weight: creating/updating/deleting resources and tracking them until they are ready while printing logs in the process.
4. Running `post-install` or `post-upgrade` [hooks]({{ "/advanced/helm/deploy_process/helm_hooks.html" | true_relative_url }}) the same way as `pre-install` and `pre-upgrade` hooks.

## Resource tracking

werf uses the [kubedog library](https://github.com/werf/kubedog) to track resources. Tracking behavior can be configured for each resource by setting [resource annotations]({{ "/reference/deploy_annotations.html" | true_relative_url }}) in the chart templates.

## Multiple Kubernetes clusters

There are cases when separate Kubernetes clusters are required for a different environments. You can [configure access to multiple clusters](https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters) using kube contexts in a single kube config.

In that case, the `--kube-context=CONTEXT` deploy option should be set manually along with the environment.

## Subcharts

During deploy process werf will render, create and track all resources of all [subcharts]({{ "/advanced/helm/configuration/chart_dependencies.html" | true_relative_url }}).
