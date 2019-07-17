---
title: Documentation
permalink: documentation/index.html
sidebar: documentation
---

Documentation of werf comprises ~100 articles which include common use cases (getting started, deploy to Kubernetes, CI/CD integration and more), comprehensive description of its functions & architecture, as well as CLI, commands.


If you are a developer and want to learn **how to build applications using Werf** — [start here](guides/getting_started.html).

If you want to understand how to configure application and environment for **deploying to Kubernetes using Werf** — [start here](guides/deploy_into_kubernetes.html).

If you want to understand how to set **CI/CD process with Werf** — [start here](guides/gitlab_ci_cd_integration.html) to understand build, deployment and scheduled registry cleanup.

[Other guides](howto/) describe **advanced cases**, you shall discover on futher working with Werf.

Want to **install Werf**? [README.md in github repository](https://github.com/flant/werf/blob/master/README.md) will help you.

If you have already begun using Werf and trying to understand **common rules, available directives specification and werf architecture details**, read [the Reference section](reference/).

Werf is a CLI utility, so if you want to find description of both **basic commands needed to provide the CI/CD process and service commands that provide advanced functionality** — use [Command Line Interface section](cli/).

Can not **find something**? Try using google using search icon in menu.

## Config

To use Werf an application should be configured in `werf.yaml` file. 
This configuration includes:

1. Definition of project meta information such as project name, which will affect build, deploy and other commands.
2. Definition of the images to be built.

In **Config** article you can find information about:

* Structure and config sections
* Organization approaches
* Config processing steps
* Supported Go templates functions

## Build

We propose to divide the assembly process into steps, intermediate images (like layers in Docker), with clear functions and assignments. 
In Werf, such step is called ***stage*** and result image consists of a set of built *stages*. 
All stages are kept in a ***stages storage*** and defining build cache of application (not really cache but part of building context).

**Build** section covers questions about:

* Stages and images
* Building process
* Configuration image and artifact directives
* Tools for debugging build process

## Registry 

There are several categories of commands that work with docker registry. 
Thus, need to authorize in docker registry:

* [Build command]({{ site.baseurl }}/documentation/reference/build_process.html#build-command) pull base images from docker registry.
* [Publish commands]({{ site.baseurl }}/documentation/reference/publish_process.html) used to create and update images in docker registry.
* [Cleaning commands]({{ site.baseurl }}/documentation/reference/cleanup_process.html) used to delete images from docker registry.
* [Deploy commands]({{ site.baseurl }}/documentation/reference/deploy/deploy_to_kubernetes.html#deploy-command) need to access _images_ from docker registry and _stages_ which could also be stored in registry.

In this section, you may find information about:

* Docker registry authorization
* Image naming during tagging and publishing 
* Publish procedure
* Cleaning abilities
