---
title: Overview
permalink: documentation/index.html
sidebar: documentation
---

The documentation of Werf comprises ~100 articles. They include common use cases (getting started, deploying to Kubernetes, CI/CD integration and more), a comprehensive description of functions & architecture, CLI, and commands.

We recommend to start off with our **Guides** section:

- [Installation]({{ site.baseurl }}/documentation/guides/installation.html) describes Werf dependencies and different installation methods.
- [Getting started]({{ site.baseurl }}/documentation/guides/getting_started.html) describes how to use Werf with regular Dockerfiles. Learn how to integrate Werf into your project smoothly.
- [Deploying into Kubernetes]({{ site.baseurl }}/documentation/guides/deploy_into_kubernetes.html) is a basic example of application deployment.
- [Gitlab CI/CD integration]({{ site.baseurl }}/documentation/guides/gitlab_ci_cd_integration.html) is all about integrating with GitLab: build, publish, deploy and schedule registry cleanup.
- [Unsupported CI/CD integration]({{ site.baseurl }}/documentation/guides/unsupported_ci_cd_integration.html) is about how to plug werf into CI/CD system that is not [officially supported]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html).
- In the Advanced build section, you can learn more about our image description syntax to take advantage of incremental rebuilds based on git history, as well as other carefully crafted tools. We recommend starting with the [First application guide]({{ site.baseurl }}/documentation/guides/advanced_build/first_application.html).

The next step is the **Configuration** section.

To use Werf, an application should be configured in `werf.yaml` file.
This configuration includes:

1. Definition of the project meta information such as a project name (it will affect build, deploy, and other commands).
2. Definition of images to be built.

In the [Overview]({{ site.baseurl }}/documentation/configuration/introduction.html) article you can find information about:

* Structure and config sections.
* Organization approaches.
* Config processing steps.
* Supported Go templates functions.

Other section articles provide detailed information about [Dockerfile Image]({{ site.baseurl }}/documentation/configuration/dockerfile_image.html), [Stapel Image]({{ site.baseurl }}/documentation/configuration/stapel_image/naming.html) and [Stapel Artifact]({{ site.baseurl }}/documentation/configuration/stapel_artifact.html) directives and their features of usage.

The **Reference** section covers essential Werf processes:

* [Build process]({{ site.baseurl }}/documentation/reference/build_process.html).
* [Publish process]({{ site.baseurl }}/documentation/reference/publish_process.html).
* [Deploy process]({{ site.baseurl }}/documentation/reference/deploy_process/deploy_into_kubernetes.html).
* [Cleaning process]({{ site.baseurl }}/documentation/reference/cleaning_process.html).

Each article describes a cpecific process: process composition, available options and features.

Also, this section includes articles with base primitives and general tools:

* [Stages and images]({{ site.baseurl }}/documentation/reference/stages_and_images.html).
* [Registry authorization]({{ site.baseurl }}/documentation/reference/registry_authorization.html).
* [Development and debug]({{ site.baseurl }}/documentation/reference/development_and_debug/setup_minikube.html).
* [Toolbox]({{ site.baseurl }}/documentation/reference/toolbox/slug.html).

Werf is a CLI utility, so if you want to find a description of both basic commands required for the CI/CD process and service commands that provide advanced functionality — use the **CLI Commands** section.

The [**Development** section]({{ site.baseurl }}/documentation/development/stapel.html) contains service and maintenance manuals and other docs that help developers to understand how a specific werf subsystem works, how to maintain said subsystem in the actual state, how to write and build new code for the werf, etc.
