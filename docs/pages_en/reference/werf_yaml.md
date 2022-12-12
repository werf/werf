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

## Deploy

### Helm chart dir

Specify custom directory to the helm chart of the project, for example `.deploy/chart`:

```yaml
deploy:
  helmChartDir: .deploy/chart
```

### Release name

werf allows to define a custom release name template, which [used during deploy process]({{ "/usage/deploy/releases/naming.html#release-name" | true_relative_url }}) to generate a release name:

```yaml
project: PROJECT_NAME
configVersion: 1
deploy:
  helmRelease: TEMPLATE
  helmReleaseSlug: false
```

`deploy.helmRelease` is a Go template with `[[` and `]]` delimiters. There are `[[ project ]]`, `[[ env ]]` and `[[ namespace ]]` functions support. Default: `[[ project ]]-[[ env ]]`. Template can be customized as follows:

{% raw %}
```yaml
deploy:
  helmRelease: >-
    [[ project ]]-[[ namespace ]]-[[ env ]]-{{ env "HELM_RELEASE_EXTRA" }}
```
{% endraw %}

**NOTE**. Usage of the `HELM_RELEASE_EXTRA` environment variable should be allowed explicitly in the [werf-giterminism.yaml]({{ "reference/werf_giterminism_yaml.html" | true_relative_url }}) configuration in that case.

`deploy.helmReleaseSlug` defines whether to apply or not [slug]({{ "/usage/deploy/releases/naming.html#slugging-the-release-name" | true_relative_url }}) to generated helm release name. Default: `true`.

### Kubernetes namespace

werf allows to define a custom Kubernetes namespace template, which [used during deploy process]({{ "/usage/deploy/releases/naming.html#kubernetes-namespace" | true_relative_url }}) to generate a Kubernetes Namespace:

```yaml
project: PROJECT_NAME
configVersion: 1
deploy:
  namespace: TEMPLATE
  namespaceSlug: true|false
```

`deploy.namespace` is a Go template with `[[` and `]]` delimiters. There are `[[ project ]]`, `[[ env ]]` functions support. Default: `[[ project ]]-[[ env ]]`. Template can be customized as follows:

{% raw %}
```yaml
deploy:
  namespace: >-
    [[ project ]]-[[ env ]]
```
{% endraw %}

`deploy.namespaceSlug` defines whether to apply or not [slug]({{ "/usage/deploy/releases/naming.html#slugging-kubernetes-namespace" | true_relative_url }}) to generated kubernetes namespace. Default: `true`.

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

## Image section

Images are declared with _image_ directive: `image: string`. 
The _image_ directive starts a description for building an application image.

```yaml
image: frontend
```

If there is only one _image_ in the config, it can be nameless:

```yaml
image: ~
```

In the config with multiple images, **all images** must have names:

```yaml
image: frontend
...
---
image: backend
...
```

An _image_ can have several names, set as a list in YAML syntax
(this usage is equal to describing similar images with different names):

```yaml
image: [main-front,main-back]
```

You will need an image name when setting up helm templates or running werf commands to refer to the specific image defined in the `werf.yaml`.

### Dockerfile builder

werf supports building images using Dockerfile. Building an image from Dockerfile is the easiest way to start using werf in an existing project.

`werf.yaml` below describes an unnamed image built from `Dockerfile` which reside in the root of the project dir:

```yaml
project: my-project
configVersion: 1
---
image: ~
dockerfile: Dockerfile
```

To build multiple named stages from a single Dockerfile:

```yaml
image: backend
dockerfile: Dockerfile
target: backend
---
image: frontend
dockerfile: Dockerfile
target: frontend
```

And also build multiple images from different Dockerfiles:

```yaml
image: backend
dockerfile: backend/Dockerfile
context: backend/
---
image: frontend
dockerfile: frontend/Dockerfile
context: frontend/
```

#### contextAddFiles

The build context consists of the files from a directory, defined by `context` directive (the project directory by default), from the current project git repository commit.

The `contextAddFiles` directive allows adding of arbitrary files or directories from the project directory to the build context.

```yaml
image: app
context: app
contextAddFiles:
 - file1
 - dir1/
 - dir2/file2.out
```

The configuration describes the build context that consists of the following files:

- `app/**/*`  from the current project git repository commit;
- Files `app/file1`, `app/dir2/file2.out` and directory `dir1` from the project directory.

The `contextAddFiles` files have a higher priority than the files from the current project git repository commit. When these files are crossing, the user will work with files from the project directory.

> By default, the use of the `contextAddFiles` directive is not allowed by giterminism (read more about it [here]({{ "/usage/project_configuration/giterminism.html#contextaddfiles" | true_relative_url }}))

### Stapel builder

Another alternative to building images with Dockerfiles is werf stapel builder, which is tightly integrated with Git and allows really fast incremental rebuilds on changes in the Git files.
