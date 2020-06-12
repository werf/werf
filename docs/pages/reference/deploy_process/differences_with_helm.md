---
title: Differences with Helm
sidebar: documentation
permalink: documentation/reference/deploy_process/differences_with_helm.html
author: Timofey Kirillov <timofey.kirillov@flant.com>
---

## Builtin Helm client and Tiller

Helm 2 uses server component called [Tiller](https://helm.sh/docs/glossary/#tiller). Tiller manages [releases]({{ site.baseurl }}/documentation/reference/deploy_process/deploy_into_kubernetes.html#release): performs create, update, delete and list releases operations.

Deploying with werf does not require a Tiller installed in Kubernetes cluster. Helm client and Tiller is fully embedded into werf, and Tiller is operating locally (without grpc network requests) within single werf process during execution of deploy commands.

Yet werf is [fully compatible]({{ site.baseurl }}/documentation/reference/deploy_process/deploy_into_kubernetes.html#helm-compatibility-notice) with already existing helm 2 installations.

This architecture gives werf following advantages.

### Ability to implement resources tracking properly

werf imports helm codebase and redefines key points of helm standard deploy process. These changes making it possible to [track resources with logs](#proper-tracking-of-deployed-resources) without complex architectural solutions such as streaming of logs over grpc when using helm client with Tiller server.

### Security advantage

Typically Tiller is deployed into cluster with admin permissions. With werf user can setup fine grained access to cluster for example per project namespace.

In contrast to Tiller, releases storage in werf is not a static configuration of cluster installation. Thus werf can store releases of each project right in the namespace of the project, so that neighbour projects does not have ability to see other releases if there is no access to these neighbour namespaces.

### Installation advantage

As helm client and Tiller is built into werf there is no need for helm client installed on the host nor Tiller installed in the cluster.

## Proper tracking of deployed resources

werf tracks all chart resources until each resource reaches ready state and prints logs and info about current resources status, resources status changes during deploy process and "what werf is waiting for?".

With pure helm user has only the ability to wait for resources using `--wait` flag, but helm does not provide any interactive info about "what is going on now?" when using this flag.

![Comparison between werf and Helm]({% asset deploy/werf_and_helm.gif @path %})

Also werf fails fast when there is an error occurred during deploy process. With pure helm and wait flag user should wait till timeout occurred when something went wrong. And failed deploys is not a rare case, so waiting for timeouts significantly slows down CI/CD experience.

And additionally default tracking and error response behaviour can be configured, see [deploy essentials for more info]({{ site.baseurl }}/documentation/reference/deploy_process/deploy_into_kubernetes.html#resource-tracking-configuration).

werf uses [kubedog resource tracking library](https://github.com/werf/kubedog) under the hood.

## Annotate and label chart resources

werf automatically appends annotations to all chart resources (including hooks), such as (for GitLab):
* `"project.werf.io/git": $CI_PROJECT_URL`;
* `"ci.werf.io/commit": $CI_COMMIT_SHA`;
* `"gitlab.ci.werf.io/pipeline-url":  $CI_PROJECT_URL/pipelines/$CI_PIPELINE_ID`;
* `"gitlab.ci.werf.io/job-url": $CI_PROJECT_URL/pipelines/$CI_JOB_ID`.

User can pass any additional annotations and labels to werf deploy invocation and werf will set these annotations to all resources of the chart.

This feature opens possibilities for grouping resources by project name, url, etc. for example in monitoring task.

See [deploy essentials for more info]({{ site.baseurl }}/documentation/reference/deploy_process/deploy_into_kubernetes.html#annotate-and-label-chart-resources).

Helm does not have this ability yet and will not have in the near future, [see issue](https://github.com/helm/helm/pull/2631).

## Parallel deploys are safe

werf uses locks system to prevent any parallel deploys or dismisses when working with a single release. Release is locked by name.

Helm does not use any locks, so parallel deploys may lead to unexpected results.

## Builtin secrets support

werf has a builtin tools and abilities to work with secret files and secret values to store sensitive data right in the project repo. See [deploy essentials]({{ site.baseurl }}/documentation/reference/deploy_process/deploy_into_kubernetes.html) and [working with secrets]({{ site.baseurl }}/documentation/reference/deploy_process/working_with_secrets.html) for more info.

## Integration with built images

werf extends helm with special templates functions such as `werf_container_image` and `werf_container_env` to generate images names for the images built with werf builder. See [deploy essentials for more info]({{ site.baseurl }}/documentation/reference/deploy_process/deploy_into_kubernetes.html).

Thus werf makes it super easy for the user to integrate custom built images into user templates: werf does the work of generating "the right" docker images names and tags, and werf makes sure that image will be pulled when it is necessary for Kubernetes to pull new image version.

Using pure helm user need to invent own system of passing actual images names using values.

## Chart generation

During command execution werf:

* Passes [service values]({{ site.baseurl }}/documentation/reference/deploy_process/deploy_into_kubernetes.html#service-values) into Helm.
* Decrypts secret values files (by default `secret-values.yaml`) and passes the result values into Helm. More details about working with secret values files in [a separate article]({{ site.baseurl }}/documentation/reference/deploy_process/working_with_secrets.html).
* Adds extra Go-template functions to Helm render:
  * [integration with built images functions]({{ site.baseurl }}/documentation/reference/deploy_process/deploy_into_kubernetes.html#integration-with-built-images): `werf_container_image`, `werf_container_env`;
  * [function to use decrypted secret file content]({{ site.baseurl }}/documentation/reference/deploy_process/deploy_into_kubernetes.html#werf_secret_file): `werf_secret_file`;
  * [environment variables functions](http://masterminds.github.io/sprig/os.html): `env`, `expandenv`.

### Chart.yaml is not required

Helm chart requires `Chart.yaml` file in the root of the chart which should define chart name and version.

werf generates Chart metadata based on project `werf.yaml` therefore `Chart.yaml` file is not needed. If user have created `Chart.yaml` by itself, then werf will ignore it and print a warning.

## Fixed helm chart path in the project

werf requires [chart]({{ site.baseurl }}/documentation/reference/deploy_process/deploy_into_kubernetes.html#chart) to be placed in the `.helm` directory in the same directory where `werf.yaml` config is placed.

Helm does not define a static place in the project where helm chart should be stored.

## Chart resource validation

Errors like non-existing fields in resource configuration of chart templates are not shown by standard helm. No feedback is given to the user.

werf writes all validation errors as WARNINGS and also writes these warnings into the resource annotation (so that all these warnings can easily be fetched by cli scripts from multiple clusters). See more info [here]({{ site.baseurl }}/documentation/reference/deploy_process/deploy_into_kubernetes.html#resources-manifests-validation).

## Three way merge patches and resources adoption

werf may use 3-way-merge patches and 2-way-merge patches (as Helm 2 does) with repair patches. There are different resource update methods and adoption, all of which are compatible with existing Helm 2 installations.

More info about resources update methods and patches is available [in the reference article]({{ site.baseurl }}/documentation/reference/deploy_process/resources_update_methods_and_adoption.html) and ["3-way merge in werf: deploying to Kubernetes via Helm “on steroids” medium article](https://medium.com/flant-com/3-way-merge-patches-helm-werf-beb7eccecdfe).

Also werf has an ability to adopt already existing resources in the working Kubernetes cluster into the helm release in an explicit manner (user should explicitly manually set an annotation to the desired resource and deploy a chart using werf). More info about resources adoption is available [in the article]({{ site.baseurl }}/documentation/reference/deploy_process/resources_update_methods_and_adoption.html#resources-adoption).
