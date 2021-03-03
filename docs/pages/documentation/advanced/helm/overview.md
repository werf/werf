---
title: Overview
permalink: documentation/advanced/helm/overview.html
---

You can easily start using werf for deploying your projects via the existing [Helm](https://helm.sh) charts since they are fully compatible with werf. The configuration has a format similar to that of [Helm charts]({{ "documentation/advanced/helm/configuration/chart.html" | true_relative_url }}).

werf includes all the existing Helm functionality (the latter is built into werf) as well as some customizations:

 - werf has several configurable modes for tracking the resources being deployed, including processing logs and events;
 - you can integrate images built by werf into Helm chart's [templates]({{ "documentation/advanced/helm/configuration/templates.html" | true_relative_url }});
 - you can assign arbitrary annotations and labels to all resources being deployed to Kubernetes globally with cli options;
 - werf reads all helm configuration from the git due to [giterminism]({{ "documentation/advanced/giterminism.html" | true_relative_url }}) which enables really reproducible deploys; 
 - also, werf has some more features which will be described further.

With all of these features werf can be considered as an alternative or better helm-client to deploy standard helm-compatible charts.

Briefly, the following commands are used to deal with an application in the Kubernetes cluster:
- [converge]({{ "documentation/reference/cli/werf_converge.html" | true_relative_url }}) — to release an application;
- [dismiss]({{ "documentation/reference/cli/werf_dismiss.html" | true_relative_url }}) — to delete an application from the cluster.
- [bundle apply]({{ "documentation/reference/cli/werf_bundle_apply.html" | true_relative_url }}) — to release an application [bundle]({{ "documentation/advanced/bundles.html" | true_relative_url }}).

This chapter covers following sections:
 1. Configuration of helm to deploy your application into kubernetes with werf: [configuration section]({{ "documentation/advanced/helm/configuration/chart.html" | true_relative_url }}).
 2. How werf runs a deploy process: [deploy process section]({{ "documentation/advanced/helm/deploy_process/steps.html" | true_relative_url }}).
 3. What is release and how to manage deployed releases of your applications: [releases section]({{ "documentation/advanced/helm/releases/release.html" | true_relative_url }}).
