---
title: Overview
permalink: index.html
sidebar: documentation
---

The documentation for werf consists of over a hundred articles. They include common use cases (getting started, deploying to Kubernetes, CI/CD integration, and more), a thorough description of functions & architecture, CLI, and commands.

We recommend starting with our **Guides** section:

- [Installation]({{ site.baseurl }}/guides/installation.html) describes werf dependencies and various installation methods.
- [Getting started]({{ site.baseurl }}/guides/getting_started.html) describes how to use werf with regular Dockerfiles. Learn how to integrate werf into your project smoothly and effortlessly.
- [Deploying to Kubernetes]({{ site.baseurl }}/guides/deploy_into_kubernetes.html) is a basic example of deploying an application.
- [Generic CI/CD integration]({{ site.baseurl }}/guides/generic_ci_cd_integration.html) is a generalized howto on using werf with any CI/CD system.
- [GitLab CI integration]({{ site.baseurl }}/guides/gitlab_ci_cd_integration.html) has all the necessary information about werf integration with GitLab CI: build, publish, deploy, and schedule Docker registry cleanup.
- [GitHub Actions integration]({{ site.baseurl }}/guides/github_ci_cd_integration.html) has all the necessary information about werf integration with GitHub Actions: build, publish, deploy, and clean up.
- In the Advanced build section, you can learn more about our image description syntax to take advantage of incremental rebuilds based on git history, as well as about other carefully crafted tools. We recommend starting with the [First application guide]({{ site.baseurl }}/guides/advanced_build/first_application.html).

The next step is the **Configuration** section.

To use werf, an application should be configured in `werf.yaml` file.

The configuration includes:

1. Definition of the project meta information such as a project name (it will affect build, deploy, and other commands).
2. Definition of images to be built.

In the [Overview]({{ site.baseurl }}/configuration/introduction.html) article you can find information about:

* Structure and config sections.
* Organization approaches.
* Config processing steps.
* Supported Go templates functions.

Other section articles provide detailed information about [Dockerfile Image]({{ site.baseurl }}/configuration/dockerfile_image.html), [Stapel Image]({{ site.baseurl }}/configuration/stapel_image/naming.html) and [Stapel Artifact]({{ site.baseurl }}/configuration/stapel_artifact.html) directives and their features.

The **Reference** section covers essential werf processes:

* [Building process]({{ site.baseurl }}/reference/build_process.html).
* [Publishing process]({{ site.baseurl }}/reference/publish_process.html).
* [Deploying process]({{ site.baseurl }}/reference/deploy_process/deploy_into_kubernetes.html).
* [Cleaning process]({{ site.baseurl }}/reference/cleaning_process.html).

Each article describes specific process: its composition, available options and features.

Also, this section includes articles with base primitives and general tools:

* [Stages and images]({{ site.baseurl }}/reference/stages_and_images.html).
* [Working with Docker registries]({{ site.baseurl }}/reference/working_with_docker_registries.html).
* [Development and debug]({{ site.baseurl }}/reference/development_and_debug/setup_minikube.html).
* [Toolbox]({{ site.baseurl }}/reference/toolbox/slug.html).

Since werf is a CLI utility, you may find a thorough description of basic commands required for the CI/CD process as well as service commands that provide advanced functionality in the **CLI Commands** section.

The [**Development** section]({{ site.baseurl }}/development/stapel.html) contains service and maintenance manuals and other docs that help developers to understand how a specific werf subsystem works, how to maintain said subsystem in the actual state, how to write and build new code for the werf, etc.
