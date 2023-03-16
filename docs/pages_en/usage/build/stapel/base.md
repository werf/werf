---
title: Base image
permalink: usage/build/stapel/base.html
author: Alexey Igrychev <alexey.igrychev@flant.com>
directive_summary: base_image
---

Here's a minimal `werf.yaml`. It describes an _image_ named `example` that is based on a _base image_ named `alpine`:

```yaml
project: my-project
configVersion: 1
---
image: example
from: alpine
```

A _base image_ can be declared with `from`, `fromImage`, or `fromArtifact` directive.

## from, fromLatest

The `from` directive defines the name and tag of a _base image_. If no tag is specified, the tag defaults to `latest`.

```yaml
from: <image>[:<tag>]
```

By default, the assembly process does not depend on the actual _base image_ digest in the repository, it only depends on the value of the _from_ directive.
Thus, changing the _base image_ locally or in the repository will not affect the build as long as the _from_ stage exists in the _storage_.

Use the _fromLatest_ directive to ensure that the current _base image_ is used. If this directive is set, werf will check for the current digest of the base image in the container registry at each run.

Here is an example of how to use the fromLatest directive:

```yaml
fromLatest: true
```

> By default, giterminism does not allow the use of the `fromLatest` directive (you can read more about it [here]({{"usage/project_configuration/giterminism.html" | true_relative_url }}))

## fromImage and fromArtifact

In addition to the image from the repository, the _base image_ can also refer to an _image_ or an [_artifact_]({{ "usage/build/stapel/imports.html#what-is-an-artifact" | true_relative_url }}) defined in the same `werf.yaml`.

```yaml
fromImage: <image name>
fromArtifact: <artifact name>
```

If the _base image_ is specific to a particular application,
it makes sense to store its description together with _images_ and _artifacts_ which use it as opposed to storing the _base image_ in a container registry.

This method comes in handy if the stages of the existing _stage conveyor_ are not enough for building the image. Using the image described in the same `werf.yaml` as the base image, you can essentially build your own _stage conveyor_.

<a class="google-drawings" href="{{ "images/configuration/base_image2.svg" | true_relative_url }}" data-featherlight="image">
    <img src="{{ "images/configuration/base_image2.svg" | true_relative_url }}" alt="Conveyor with fromImage and fromArtifact stages">
</a>

## fromCacheVersion

As described above, normally the build process actively uses caching. During the build, a werf checks to see if the base image has changed. Depending on the directives used, the digest or the image name and tag are checked for changes. If the image is unchanged, the digest of the from stage is unchanged as well. If there is an image with this digest in the stages storage, it will be used during the build.

You can use the `fromCacheVersion` directive to manipulate the digest of the `from` stage (since the `fromCacheVersion` value is part of the stage digest) and thus force the rebuild of the image. If you modify the value specified in the `fromCacheVersion` directive, then regardless of whether the _base_ image (or its digest) is changed or stays the same, the digest of the `from` stage (as well as all subsequent stages) will change. This will result in rebuilding all the stages.

```yaml
fromCacheVersion: <arbitrary string>
```

## How the Stapel builder processes CMD and ENTRYPOINT

To build a stage, werf runs a container with the `CMD` and `ENTRYPOINT` service parameters and then substitutes them with the values of the [base image]({{"usage/build/stapel/base.html" | true_relative_url }}). If these values are not set in the base image, werf resets them as follows:

- `[]` for `CMD`;
- `[""]` for `ENTRYPOINT`.

On top of that, werf resets (using special empty values) `ENTRYPOINT` of the base image if the `CMD` parameter is specified in the configuration (`docker.CMD`).

Otherwise, the behavior of werf is similar to that of [Docker](https://docs.docker.com/engine/reference/builder/#understand-how-cmd-and-entrypoint-interact).
