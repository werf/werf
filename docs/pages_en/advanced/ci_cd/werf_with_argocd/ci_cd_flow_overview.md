---
title: CI/CD flow overview
permalink: advanced/ci_cd/werf_with_argocd/ci_cd_flow_overview.html
---

In this article, we will describe how to implement continuous delivery and continuous deployment for your project using werf and the GitOps model. The following topics will be discussed:
* using **werf** to build and publish release artifacts;
* using the **ArgoCD GitOps tool** to deliver release artifacts to Kubernetes;
* gluing together werf and ArgoCD using **bundles**.

## Key points and principles

This section outlines the key points and principles needed to implement the CI/CD process.

### Single source of truth

The Git repository of the project is considered a single source of truth. It contains:
* application source code;
* instructions for building images;
* Helm chart files.

### Release artifacts

1. Release artifacts are published to the container registry and include: 
   * built images;
   * Helm charts.
2. A release artifact can be prepared and published for an arbitrary git commit.
3. There are two requirements for published artifacts: they must be immutable and reproducible.
4. Dependening on CI/CD configuration, release artifacts can be published:
   * for each commit in the main branch, followed by deployment to production;
   * for Git tags that match the application's releases.
5. You can use any common CI/CD system (such as Gitlab CI/CD or GitHub Actions) to prepare release artifacts.

### Deploying to Kubernetes with GitOps

There is a GitOps operator deployed into the cluster that implements a pull model for deploying project releases. It:
* keeps track of new release artifact versions in the container registry;
* rolls out a particular version of the artifact to the Kubernetes cluster.

Depending on the CI/CD flow, the GitOps operator can deploy the release artifact either manually (Continuous Delivery) or automatically (Continous Deployment). 

## Reference CI/CD flow

The CI/CD reference flow will be implemented using two main tools: werf and ArgoCD.

The diagram below shows the blocks representing the stages of the cycle, as well as the tools for implementing them.   

<a class="google-drawings" href="{{ "images/advanced/ci_cd/werf_with_argocd/ci-cd-flow.svg" | true_relative_url }}" data-featherlight="image">
<img src="{{ "images/advanced/ci_cd/werf_with_argocd/ci-cd-flow.svg" | true_relative_url }}">
</a>

### 1. Local development

The flow starts with local development, for which werf offers the following functionality:
* All werf commands support the `--dev` and `--follow` flags to simplify handling uncommitted changes and additions made to the application source code.
* werf allows you to use the same build and deploy configuration to deploy an application either locally or in production.
* Special environments (passed by the `--env ENV` parameter) allow you to customize build and deployment configurations to deploy an application to a local Kubernetes cluster for local development.

### 2. Committing

The committing stage implies that the entire werf building and deployment configuration, as well as the application source code, must be committed to the project's Git repository.

For example, when your CI/CD system checks out some commit and run some instructions, these instructions should not accidentally or intentionally generate any dynamic input configuration files for the werf build and deploy configuration.

This is default werf behaviour which is controlled by the **giterminism**. Giterminism allows the users to set the degree to which a project's build and deploy configurations are reproducible (the strictest rules are enabled by default; you can loosen them by making changes to the `werf-giterminism.yaml` configuration file).

### 3. Building and publishing images

werf supports building images using Dockerfile or Stapel and then publishing them to the container registry. It provides distributed layer caching and content-based tagging, preventing unnecessary reassembly of images.

The images are built with the `werf build` command. More information can be found [in the configuration article]({{ "advanced/ci_cd/werf_with_argocd/configure_ci_cd.html#building-and-publishing-images" | true_relative_url }}).

### 4. Fast commit testing

Fast testing of each commit in the Git repository is an essential part of the CI/CD approach. Typically, unit tests and linters are run at this stage.

Usually, the tests are run in one-shot containers with built images. werf offers the following commands to run such one-shot containers:
* `werf run` — run the one-shot command container on the Docker server;
* `werf kube-run` — run the one-shot command container in the Kubernetes cluster. 

More information can be found [in the configuration article]({{ "advanced/ci_cd/werf_with_argocd/configure_ci_cd.html#fast-commit-testing" | true_relative_url }}).

### 5. Preparing release artifacts

At this stage, you need to prepare a [release artifact](#release-artifacts) using the so-called [werf bundles]({{ "advanced/bundles.html" | true_relative_url }}).

The `werf bundle publish` command:
* ensures that all the necessary images are build;
* publishes the Helm chart files, configured to use the built images, to the container registry — the so-called bundle.

The published bundle can then be deployed to a Kubernetes cluster using ArgoCD as a regular OCI Helm chart.  

More information can be found [in the configuration article]({{ "advanced/ci_cd/werf_with_argocd/configure_ci_cd.html#preparing-release-artifacts" | true_relative_url }}).

### 6. Acceptance testing

Acceptance testing is an integral part of CI/CD. Such tests usually take a long time to run, and may require a production-like environment to be available during testing.

Typically, acceptance testing is performed automatically after a Pull Request is merged into the main branch.

Acceptance testing of new releases is not necessary for Continuous Integration, but in most cases it is necessary for Continuous Delivery of new releases. In other words, you can add new commits and PRs to the main branch without conducting acceptance tests, but the release artifact must successfully pass acceptance tests before it can be deployed to production. 

werf allows you to set a special test environment  (with `--env ENV` param).
* In the test environment, there may be special build and deploy configuration files;
* The test environment also has a special Helm post-install `Job` resource.

werf offers the following commands to run acceptance tests:
* `werf converge` to deploy a production-like environment to the Kubernetes cluster and run the test `Job` as a Helm post-install hook;
  * This command requires access to the project's Git repository.
* `werf dismiss` to uninstall the previously deployed production-like test environment;
* `werf bundle apply` to deploy a production-like environment to Kubernetes and run the test `Job` as a Helm post-install hook.
  * This command does not require access to the Git repository, it only requires access to the bundle published in the container registry.

Acceptance tests can also be run by ArgoCD using a published release artifact, the werf bundle. In this case, ArgoCD will deploy a production-like environment to Kubernetes and run the test `Job` as a Helm post-install hook.

More information can be found [in the configuration article]({{ "advanced/ci_cd/werf_with_argocd/configure_ci_cd.html#acceptance-testing" | true_relative_url }}).

### 7. Deploying into production

In this step, ArgoCD must deploy the published release artifact for the target application commit to the Kubernetes cluster.

Here, the werf bundle serves as the release artifact. ArgoCD treats it as a regular OCI Helm chart.

ArgoCD can deploy the target release artifact after a manual user command.

ArgoCD Image Updater with the ["continuous deployment of OCI Helm chart type application" patch](https://github.com/argoproj-labs/argocd-image-updater/pull/405) can automatically deploy the latest version of a published release artifact.

More information can be found [in the configuration article]({{ "advanced/ci_cd/werf_with_argocd/configure_ci_cd.html#acceptance-testing" | true_relative_url }}).
