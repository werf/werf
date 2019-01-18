---
title: Introduction
sidebar: how_to
permalink: /
---

Werf helps implement and support Continuous Integration and Continuous Delivery (CI/CD).

With werf, DevOps engineers generate and deploy images by linking together:

- application code (with Git support),
- infrastructure code (with Ansible or shell scripts), and
- platform as a service (Kubernetes).

Werf simplifies development of build scripts, reduces commit build time and automates deployment.
It is designed to make engineer's work fast end efficient.


## Features

* Reducing average build time.
* Sharing a common cache between builds.
* Running distributed builds with common registry.
* Reducing image size by detaching source data and build tools.
* Building images with Ansible or shell scripts.
* Building multiple images from one description.
* Advanced tools for debugging built images.
* Deploying to Kubernetes via [helm](https://helm.sh/), the Kubernetes package manager.
* Tools for cleaning both local and remote Docker registry caches.

## First Steps

* [Install werf and dependencies]({{ site.baseurl }}/how_to/installation.html).
* [Build your first application with werf]({{ site.baseurl }}/how_to/getting_started.html).
