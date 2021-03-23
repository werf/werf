---
title: Base image
permalink: advanced/building_images_with_stapel/base_image.html
author: Alexey Igrychev <alexey.igrychev@flant.com>
directive_summary: base_image
---

Here's a minimal `werf.yaml`. It describes a _image_ named `example` that is based on a _base image_ named `alpine`:

```yaml
project: my-project
configVersion: 1
---
image: example
from: alpine
```

_Base image_ can be declared with `from`, `fromImage` or `fromArtifact` directive.

## from, fromLatest

The `from` directive defines the name and tag of a _base image_. If absent, tag defaults to `latest`.

```yaml
from: <image>[:<tag>]
```

By default, the assembly process does not depend on actual _base image_ digest in the repository, only on _from_ directive value.
Thus, changing _base image_ locally or in the repository does not matter if _from_ stage is already exists in _storage_.

If you want always build the image with actual _base image_ you should use _fromLatest_ directive.
_fromLatest_ directive allows connecting the assembly process with the _base image_ digest getting from the repository.

```yaml
fromLatest: true
```

> By default, the use of the `fromLatest` directive is not allowed by giterminism (read more about it [here]({{ "advanced/giterminism.html" | true_relative_url }}))

## fromImage and fromArtifact

Besides using docker image from a repository, the _base image_ can refer to _image_ or [_artifact_]({{ "advanced/building_images_with_stapel/artifacts.html" | true_relative_url }}), that is described in the same `werf.yaml`.

```yaml
fromImage: <image name>
fromArtifact: <artifact name>
```

If a _base image_ is specific to a particular application,
it is reasonable to store its description with _images_ and _artifacts_ which are used it as opposed to storing the _base image_ in a container registry.

Also, this method can be useful if the stages of _stage conveyor_ are not enough for building the image. You can design your _stage conveyor_.

<a class="google-drawings" href="{{ "images/configuration/base_image2.png" | true_relative_url }}" data-featherlight="image">
    <img src="{{ "images/configuration/base_image2_preview.png" | true_relative_url }}" alt="Conveyor with fromImage and fromArtifact stages">
</a>

## fromCacheVersion

The `fromCacheVersion` directive allows managing image reassembly.

```yaml
fromCacheVersion: <arbitrary string>
```
