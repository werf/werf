---
title: Introduction
sidebar: how_to
permalink: /
---

Werf helps implement and support Continuous Integration and Continuous Delivery (CI/CD).

It helps DevOps engineers generate and deploy images by linking together:

- application code (with Git support),
- infrastructure code (with Ansible or shell scripts), and
- platform as a service (Kubernetes).

Werf simplifies development of build scripts, reduces commit build time and automates deployment.
It is designed to make engineer's work fast end efficient.

## Features

* Comlete application lifecycle management: build and cleanup images, deploy application into kubernetes.
* Reducing average build time.
* Building images with Ansible and shell scripts.
* Building multiple images from one description.
* Sharing a common cache between builds.
* Reducing image size by detaching source data and build tools.
* Running distributed builds with common registry.
* Advanced tools for debugging built images.
* Tools for cleaning both local and remote Docker registry caches.
* Deploying to Kubernetes via [helm](https://helm.sh/), the Kubernetes package manager.

## First Steps

* [Install Werf and dependencies]({{ site.baseurl }}/how_to/installation.html).
* [Build your first application with Werf]({{ site.baseurl }}/how_to/getting_started.html).
