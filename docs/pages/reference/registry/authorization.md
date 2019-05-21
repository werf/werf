---
title: Authorization
sidebar: reference
permalink: reference/registry/authorization.html
author: Timofey Kirillov <timofey.kirillov@flant.com>
---

There are several categories of commands that work with docker registry. 
Thus, need to authorize in docker registry:

* [Build command]({{ site.baseurl }}/reference/build/assembly_process.html#werf-build) pull base images from docker registry.
* [Publish commands]({{ site.baseurl }}/reference/registry/publish.html) used to create and update images in docker registry.
* [Cleaning commands]({{ site.baseurl }}/reference/registry/cleaning.html) used to delete images from docker registry.
* [Deploy commands]({{ site.baseurl }}/reference/deploy/deploy_to_kubernetes.html#werf-kube-deploy) need to access _images_ from docker registry and _stages_ which could also be stored in registry.

These commands do not perform authorization and use defined _docker config_ working with docker registry.
By default, werf uses the same _docker config_ as docker utility does (defined in `$DOCKER_CONFIG`).
Docker config can be redefined by `--docker-config` option or `$WERF_DOCKER_CONFIG`. 
The variable value is the same as `docker --config` standard option value.   

## How to perform a login procedure with werf

To prepare _docker config_ you can use standart docker client command `login` or, if you are using CI system, [ci-env command]({{ site.baseurl }}/cli/toolbox/ci_env.html) in werf. 

> Using `docker login` in parallel CI jobs can lead to failed jobs because of a race condition and temporary credentials.
One job affects another job overriding temporary credentials in _docker config_. 
Thus, a user should provide independent _docker configs_ between jobs, `docker --config`, or use `ci-env` command 

`ci-env` command generates werf environment variables for a specified CI system. 
The command creates a temporary _docker config_ based on current user _docker config_. 
Then runs `docker login` command with special credentials for specified CI system (e.g., `$CI_REGISTRY_IMAGE` and `$CI_JOB_TOKEN` for GitLab) and overrides `$DOCKER_CONFIG` with the temporary _docker config_ path. 
So, this configuration will be used by werf and docker in current terminal session.
