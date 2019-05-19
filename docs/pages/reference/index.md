---
title: Reference
sidebar: reference
permalink: reference/
---

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

* [Build command]({{ site.baseurl }}/reference/build/assembly_process.html#werf-build) pull base images from docker registry.
* [Publish commands]({{ site.baseurl }}/reference/registry/publish.html) used to create and update images in docker registry.
* [Cleaning commands]({{ site.baseurl }}/reference/registry/cleaning.html) used to delete images from docker registry.
* [Deploy commands]({{ site.baseurl }}/reference/deploy/deploy_to_kubernetes.html#werf-kube-deploy) need to access _images_ from docker registry and _stages_ which could also be stored in registry.

In this section, you may find information about:

* Docker registry authorization
* Image naming during tagging and publishing 
* Publish procedure
* Cleaning abilities
