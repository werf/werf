---
title: Using with CI/CD systems
permalink: documentation/using_with_ci_cd_systems.html
sidebar: documentation
---

Werf provides following main commands as building blocks to construct your pipelines:

 - `werf converge`;
 - `werf dismiss`;
 - `werf cleanup`.

Generally it is really easy to emded werf into any CI/CD system job with following steps:

 1. Setup CI/CD job access to the Docker Registry and Kubernetes cluster.
 2. Checkout target git commit of the project repo.
 3. Call `werf converge` command to deploy app, or `werf dismiss` command to terminate app, or `werf cleanup` to clean unused docker images from the Docker Registry.

There are actually more werf commands which you may need eventually, but these are basic commands.

## Setup Docker Registry access

If your project use any custom images, which should be built specifically for the project (rather than using publicly available), then you need a Docker Registry. In such case a Docker Registry instance should be available from the CI/CD runner host (to publish built images) and from within your Kubernetes cluster (to pull images).

Perform login into Docker Registry as a part of CI/CD job and specify `--repo` param (or `WERF_REPO` environment variable) for the `werf converge` command.

## Setup Kubernetes cluster access

Werf need a [kube config](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/) to connect and interract with the Kubernetes cluster.

Prepare kube config and pass it to the werf command using kube config file or as base64 encoded data.

## Setup environment param for werf

Typically an application should be deployed into different [environments]({{ site.baseurl }}/documentation/advanced/ci_cd/ci_cd_workflow_basics.html#environment) (such as `production`, `staging`, `testing`, etc.). Different groups of users interact with application instance of suitable environment (end users interact with `production` environment, testers interact with `testing` environments and so on).

Werf supports optional `--env` param (or environment variable `WERF_ENV`) which specifies the name of the used environment. This environment name will affect used [Kubernetes namespace]() and [Helm release name](). It is recommended to detect environment in the CI/CD job (for example using builtin environment variables of your CI/CD system) and set werf env param correspondingly.

## Checkout git commit and run werf

It is required to checkout target git commit of the project repo before running main werf commands (`werf converge`, `werf dismiss` or `werf cleanup`).

_NOTE: `werf converge` command behaviour is fully deterministic and transparent from the git repo standpoint. After converge is done your application is up and running in the state defined in the target git commit. Typically to rollback your application to the previous version you just need to run converge job on the corresponding previous commit (werf will correctly build and use the correct images for this commit)._

## Other configuration of werf for CI/CD

There is other stuff which is optional, but typically configured for werf in CI/CD:
 - custom annotations and labels for all deployed Kubernetes resources (project name, project git url, CI/CD converge job url, application version, git commit, git tag, git branch, etc.);
 - colorized log output and output width.

## Summary

In this article we have covered the basics of using werf with any CI/CD system.

[Generic CI/CD integration]({{ site.baseurl }}/documentation/advanced/ci_cd/generic_ci_cd_integration.html) article contains more details of setting up werf for any CI/CD system.

There is also special support for GitLab CI/CD and GitHub Actions in werf and special command `werf ci-env`. This command is optional, but allows automatic configuration of all werf params described in this article in a common way. Following articles contain more details:
 - [GitLab CI/CD]({{ site.baseurl }}/documentation/advanced/ci_cd/gitlab_ci_cd.html);
 - [GitHub Actions]({{ site.baseurl }}/documentation/advanced/ci_cd/github_actions.html).

## What's next?

It is recommended to continue with the [guides](https://ru.werf.io/applications_guide_ru) section to find a guide suitable for your project. These guides also contain detailed information about setting up CI/CD.
