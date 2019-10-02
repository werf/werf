---
title: Overview
permalink: documentation/index.html
sidebar: documentation
---

Documentation of Werf comprises ~100 articles which include common use cases (getting started, deploy to Kubernetes, CI/CD integration and more), comprehensive description of its functions & architecture, as well as CLI, commands.

We recommend to start discovering from our **Guides** section:

- [Installation]({{ site.baseurl }}/documentation/guides/installation.html) describes Werf dependencies and different installation methods.
- [Getting started]({{ site.baseurl }}/documentation/guides/getting_started.html) helps to start using Werf with regular Dockerfile. Take your project and put into Werf easily just now.
- [Deploying into Kubernetes]({{ site.baseurl }}/documentation/guides/deploy_into_kubernetes.html) is a short example of application deployment.
- [Gitlab CI/CD integration]({{ site.baseurl }}/documentation/guides/gitlab_ci_cd_integration.html) is all about integration with GitLab: build, publish, deployment and scheduled registry cleanup.
- [Unsupported CI/CD integration]({{ site.baseurl }}/documentation/guides/unsupported_ci_cd_integration.html) is about plugging werf into CI/CD system that is not [officially supported]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html).
- Advanced build section is about our image description syntax to take advantage of incremental rebuilds based on git history and other carefully crafted tools. Recommend to start reading from [First application guide]({{ site.baseurl }}/documentation/guides/advanced_build/first_application.html).

The next step is **Configuration** section.

To use Werf an application should be configured in `werf.yaml` file.
This configuration includes:

1. Definition of project meta information such as project name, which will affect build, deploy and other commands.
2. Definition of the images to be built.

In the [Overview]({{ site.baseurl }}/documentation/configuration/introduction.html) article you can find information about:

* Structure and config sections.
* Organization approaches.
* Config processing steps.
* Supported Go templates functions.

Other section articles give detailed information about [Dockerfile Image]({{ site.baseurl }}/documentation/configuration/dockerfile_image.html), [Stapel Image]({{ site.baseurl }}/documentation/configuration/stapel_image/naming.html) and [Stapel Artifact]({{ site.baseurl }}/documentation/configuration/stapel_artifact.html) directives and their features of usage.

**Reference** section is dedicated to Werf main processes:

* [Build process]({{ site.baseurl }}/documentation/reference/build_process.html).
* [Publish process]({{ site.baseurl }}/documentation/reference/publish_process.html).
* [Deploy process]({{ site.baseurl }}/documentation/reference/deploy_process/deploy_into_kubernetes.html).
* [Cleaning process]({{ site.baseurl }}/documentation/reference/cleaning_process.html).

Each article describes a certain process: process composition, available options and features.

Also, this section includes articles with base primitives and general tools:

* [Stages and images]({{ site.baseurl }}/documentation/reference/stages_and_images.html).
* [Registry authorization]({{ site.baseurl }}/documentation/reference/registry_authorization.html).
* [Development and debug]({{ site.baseurl }}/documentation/reference/development_and_debug/setup_minikube.html).
* [Toolbox]({{ site.baseurl }}/documentation/reference/toolbox/slug.html).

Werf is a CLI utility, so if you want to find a description of both basic commands needed to provide the CI/CD process and service commands that provide advanced functionality — use **CLI Commands** section.

[**Development** section]({{ site.baseurl }}/documentation/development/stapel.html) contains service and maintenance manuals and other docs which help developers to understand how some werf subsystem works, how to maintain some subsystem in the actual state, how to write and build new code for the werf, etc.
