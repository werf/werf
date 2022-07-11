---
title: Steps
permalink: advanced/helm/deploy_process/steps.html
---

When running the `werf converge` command, werf starts the deployment process that includes the following steps:

1. Render chart templates into a single list of Kubernetes resource manifests and validate them.
2. Run `pre-install` or `pre-upgrade` [hooks]({{ "/advanced/helm/deploy_process/helm_hooks.html" | true_relative_url }}) sequentially, sorted by their weight, and track each hook until its termination while printing its logs in the process.
3. Group non-hook Kubernetes resources by their [weight]({{ "/advanced/helm/deploy_process/deployment_order.html" | true_relative_url }}) and deploy each group sequentially, sorted by their weight: create/update/delete resources and track these resources until they are ready while printing logs in the process.
4. Run `post-install` or `post-upgrade` [hooks]({{ "/advanced/helm/deploy_process/helm_hooks.html" | true_relative_url }}) the same way as with the `pre-install` and `pre-upgrade` hooks.

## Resource tracking

werf uses the [kubedog library](https://github.com/werf/kubedog) to track resources. Tracking behaviour can be configured for each resource using [resource annotations]({{ "/reference/deploy_annotations.html" | true_relative_url }}), which should be set in the chart templates.

## Multiple Kubernetes clusters

There are cases when separate Kubernetes clusters are required for a different environments. You can [configure access to multiple clusters](https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters) using kube contexts in a single kube config.

In that case, the `--kube-context=CONTEXT` deploy option should be set manually along with the environment.

## Subcharts

During deploy process werf will render, create and track all resources of all [subcharts]({{ "/advanced/helm/configuration/chart_dependencies.html" | true_relative_url }}).
