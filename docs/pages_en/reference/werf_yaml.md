---
title: werf.yaml
permalink: reference/werf_yaml.html
description: werf.yaml config
toc: false
---

{% include reference/werf_yaml/table.html %}

## Project name

`project` defines unique project name of your application. Project name affects build cache image names, Kubernetes Namespace, Helm release name and other derived names. This is a required field of meta configuration.

Project name should be unique within group of projects that shares build hosts and deployed into the same Kubernetes cluster (i.e. unique across all groups within the same gitlab). Project name must be maximum 50 chars, only lowercase alphabetic chars, digits and dashes are allowed.

### Warning on changing project name

**WARNING**. You should never change project name, once it has been set up, unless you know what you are doing. Changing project name leads to issues:
1. Invalidation of build cache. New images must be built. Old images must be cleaned up from local host and container registry manually.
2. Creation of completely new Helm release. So if you already had deployed your application, then changed project name and deployed it again, there will be created another instance of the same application.

werf cannot automatically resolve project name change. Described issues must be resolved manually in such case.

## Git worktree

werf stapel builder needs a full git history of the project to perform in the most efficient way. Based on this the default behaviour of the werf is to fetch full history for current git clone worktree when needed. This means werf will automatically convert shallow clone to the full one and download all latest branches and tags from origin during cleanup process. 

Default behaviour described by the following settings:

```yaml
gitWorktree:
  forceShallowClone: false
  allowUnshallow: true
  allowFetchOriginBranchesAndTags: true
```

For example to disable automatic unshallow of git clone use following settings:

```yaml
gitWorktree:
  forceShallowClone: true
  allowUnshallow: false
```
