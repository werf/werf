---
title: CI/CD flow overview
permalink: advanced/ci_cd/werf_with_argocd/ci_cd_flow_overview.html
---

In this article we will describe how to implement continuous delivery and continuous deployment for your project using werf with gitops model. Briefly it implies:
* usage of **werf** to build and publish release artifacts;
* usage of **argocd gitops tool** to deliver release artifacts into the kubernetes;
* gluing together werf and argocd with **bundles**.

## Keypoints and principles

In this section we would define keypoints and principles used to implement CI/CD process.

### Single source of truth

Project git repository is the single source of truth, which contains:
* application source code;
* instructions to build images;
* helm chart files.

### Release artifacts

1. Release artifacts published into container registry and consists of: 
   * built images;
   * helm chart.
2. Release artifact could be prepared and published for an arbitrary git commit.
3. Once published for a commit release artifact should be immutable and reproducible.
4. Dependening on CI/CD flow:
   * Release artifact could be published for each commit of the main branch, and then deployed into production.
   * Release artifact could be published only for git tags, which represent application releases.
5. Any conventional CI/CD system (like Gitlab CI/CD or GitHub Actions) could be used to prepare release artifacts.

### Deploy into Kubernetes with GitOps

There is a gitops operator deployed into the cluster which implements pull model to deploy project releases:
* Operator watches for new versions of release artifacts in the container registry.
* Operator rollout some release artifact version into kubernetes cluster.

Depending on CI/CD flow GitOps operator can rollout release artifact either manually (Continuous Delivery) or automatically (Continous Deployment). 

## Reference CI/CD flow

Reference CI/CD flow will be implemented using 2 main tools: werf and ArgoCD.

In the following scheme we have a flow blocks and specify which tool implement blocks of this flow.   

<a class="google-drawings" href="{{ "images/advanced/ci_cd/werf_with_argocd/ci-cd-flow.svg" | true_relative_url }}" data-featherlight="image">
<img src="{{ "images/advanced/ci_cd/werf_with_argocd/ci-cd-flow.svg" | true_relative_url }}">
</a>

### 1. Local dev

Flow starts with local development which is covered by the following werf abilities:
* All werf commands support `--dev` and `--follow` flags to simplify work with uncommitted changes and new changes introduced into the application source code.
* werf allows usage of the same build and deploy configuration to deploy an application either locally or into production.
* Using special environment (passed with the `--env ENV` param) project could have customizations of build and deploy configuration, which are used to deploy an application into the local Kubernetes cluster for local development purposes.

### 2. Commit

Commit stage implies that all werf build and deploy configuration and application source code should be strictly commited into the project git repository.

For example, when your CI/CD system checkout some commit and run some instructions these instructions could not accidentally or by intention generate some dynamic input configuration files for werf build and deploy configuration.

This is default werf behaviour which is controlled by the **giterminism**. Giterminism is a tool which allows users to define how reproducible project's build and deploy configuration is (the most strict rules enabled by default, and could be loosened by the `werf-giterminism.yaml` config).

### 3. Build and publish images

werf supports building and publishing images into the container registry using Dockerfile or stapel builder, enables distributed layers caching, and content based tagging, which optimizes images rebuilds.

werf supports building of images with the `werf build` command. More info [in the configuration article]({{ "advanced/ci_cd/werf_with_argocd/configure_ci_cd.html#build-and-publish-images" | true_relative_url }}).

### 4. Fast commit testing

It is essential part of CI/CD approach to perform so called fast testing of each commit in the git repository. Typically during this stage we run unit tests and linters.

Such tests would run in one-shot containers in built images in the common case. werf supports running such one-shot containers with the following commands:
* `werf run` — run one-shot command in container in the docker server.
* `werf kube-run` — run one-shot command container in the Kubernetes cluster. 

More info [in the configuration article]({{ "advanced/ci_cd/werf_with_argocd/configure_ci_cd.html#fast-commit-testing" | true_relative_url }}).

### 5. Prepare release artifacts

At this stage we should prepare [release artifact](#release-artifacts) using so called [werf bundles]({{ "advanced/bundles.html" | true_relative_url }}).

`werf bundle publish` command:
* ensures that all needed images are build;
* publishes helm chart files configured to use built images into the container registry — so-called bundle.

Published bundle can be later deployed into the kubernetes cluster by ArgoCD as a regular OCI helm chart.  

More info [in the configuration article]({{ "advanced/ci_cd/werf_with_argocd/configure_ci_cd.html#prepare-release-artifacts" | true_relative_url }}).

### 6. Acceptance testing

It is also essential part of CI/CD to perform so called acceptance testing. Such tests are typically long-running and may require full production-like environment to be available during tests.

Typically acceptance testing should be performed automatically after pull request merge into the main branch.

Pass of acceptance tests do not block Continuous Integration process, but in most cases should block Continuous Delivery of new releases. It means that it is allowed to merge new commits and PRs into main branch without running acceptance tests, but it is required to pass acceptance tests for a release artifact before releasing it into production. 

With werf we could define special test environment (with `--env ENV` param).
* In the test environment there may be special build and deploy configuration files.
* In the test environment there will be special helm post install `Job` resource.

werf supports running acceptance tests with the following commands:
* `werf converge` to install production-like environment into the Kubernetes and run test `Job` as a helm post install hook.
  * This command will need project git repository.
* `werf dismiss` to uninstall previously installed production-like test environment.
* `werf bundle apply` to install production-like environment into the Kubernetes and run test `Job` as a helm post install hook.
  * This command will not need the git repository, only bundle published into the container registry.

Alternatively acceptance tests could be run by the ArgoCD using published release artifact — werf bundle. In such case ArgoCD would install production-like environment into the Kubernetes and run test `Job` as a helm post install hook.

More info [in the configuration article]({{ "advanced/ci_cd/werf_with_argocd/configure_ci_cd.html#acceptance-testing" | true_relative_url }}).

### 7. Deploy into production

In this step ArgoCD should deploy published release artifact for target application commit into the Kubernetes cluster.

We are using werf bundle as release artifact, which could be consumed as a regular OCI helm chart by the ArgoCD.

ArgoCD could deploy target release artifact by manual user action.

ArgoCD Image Updater with the ["continuous deployment of OCI Helm chart type application" patch](https://github.com/argoproj-labs/argocd-image-updater/pull/405) could be used to enable automatic deploy to the latest version of published release artifact.

More info [in the configuration article]({{ "advanced/ci_cd/werf_with_argocd/configure_ci_cd.html#acceptance-testing" | true_relative_url }}).
