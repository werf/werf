---
title: Overview
permalink: documentation/index.html
sidebar: documentation
---

The documentation for werf consists of over a hundred articles. They include common use cases (getting started, deploying to Kubernetes, CI/CD integration, and more), a thorough description of functions & architecture, CLI, and commands.

We recommend starting with our **Guides** section:

- [Installation]({{ site.baseurl }}/documentation/guides/installation.html) describes werf dependencies and various installation methods.
- [Getting started]({{ site.baseurl }}/documentation/guides/getting_started.html) describes how to use werf with regular Dockerfiles. Learn how to integrate werf into your project smoothly and effortlessly.
- [Deploying to Kubernetes]({{ site.baseurl }}/documentation/guides/deploy_into_kubernetes.html) is a basic example of deploying an application.
- [Gitlab CI/CD integration]({{ site.baseurl }}/documentation/guides/gitlab_ci_cd_integration.html) has all the necessary information about werf integration with GitLab: build, publish, deploy, and schedule Docker registry cleanup.
- [Unsupported CI/CD integration]({{ site.baseurl }}/documentation/guides/unsupported_ci_cd_integration.html) is about how to use werf with the CI/CD system that is not [officially supported]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html).
- In the Advanced build section, you can learn more about our image description syntax to take advantage of incremental rebuilds based on git history, as well as about other carefully crafted tools. We recommend starting with the [First application guide]({{ site.baseurl }}/documentation/guides/advanced_build/first_application.html).

The next step is the **Configuration** section.

To use werf, an application should be configured in `werf.yaml` file.

The configuration includes:

1. Definition of the project meta information such as a project name (it will affect build, deploy, and other commands).
2. Definition of images to be built.

In the [Overview]({{ site.baseurl }}/documentation/configuration/introduction.html) article you can find information about:

* Structure and config sections.
* Organization approaches.
* Config processing steps.
* Supported Go templates functions.

Other section articles provide detailed information about [Dockerfile Image]({{ site.baseurl }}/documentation/configuration/dockerfile_image.html), [Stapel Image]({{ site.baseurl }}/documentation/configuration/stapel_image/naming.html) and [Stapel Artifact]({{ site.baseurl }}/documentation/configuration/stapel_artifact.html) directives and their features.

The **Reference** section covers essential werf processes:

* [Building process]({{ site.baseurl }}/documentation/reference/build_process.html).
* [Publishing process]({{ site.baseurl }}/documentation/reference/publish_process.html).
* [Deploying process]({{ site.baseurl }}/documentation/reference/deploy_process/deploy_into_kubernetes.html).
* [Cleaning process]({{ site.baseurl }}/documentation/reference/cleaning_process.html).

Each article describes specific process: its composition, available options and features.

Also, this section includes articles with base primitives and general tools:

* [Stages and images]({{ site.baseurl }}/documentation/reference/stages_and_images.html).
* [Registry authorization]({{ site.baseurl }}/documentation/reference/registry_authorization.html).
* [Development and debug]({{ site.baseurl }}/documentation/reference/development_and_debug/setup_minikube.html).
* [Toolbox]({{ site.baseurl }}/documentation/reference/toolbox/slug.html).

Since werf is a CLI utility, you may find a thorough description of basic commands required for the CI/CD process as well as service commands that provide advanced functionality in the **CLI Commands** section.

The [**Development** section]({{ site.baseurl }}/documentation/development/stapel.html) contains service and maintenance manuals and other docs that help developers to understand how a specific werf subsystem works, how to maintain said subsystem in the actual state, how to write and build new code for the werf, etc.
