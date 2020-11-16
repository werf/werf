---
title: Base image
sidebar: documentation
permalink: documentation/advanced/building_images_with_stapel/base_image.html
author: Alexey Igrychev <alexey.igrychev@flant.com>
summary: |
  <a class="google-drawings" href="../../../images/configuration/base_image1.png" data-featherlight="image">
      <img src="../../../images/configuration/base_image1_preview.png" alt="Base image">
  </a>

  <div class="language-yaml highlighter-rouge"><div class="highlight"><pre class="highlight"><code><span class="na">from</span><span class="pi">:</span> <span class="s">&lt;image[:&lt;tag&gt;]&gt;</span>
  <span class="na">fromLatest</span><span class="pi">:</span> <span class="s">&lt;bool&gt;</span>
  <span class="na">fromCacheVersion</span><span class="pi">:</span> <span class="s">&lt;arbitrary string&gt;</span>
  <span class="na">fromImage</span><span class="pi">:</span> <span class="s">&lt;image name&gt;</span>
  <span class="na">fromArtifact</span><span class="pi">:</span> <span class="s">&lt;artifact name&gt;</span>
  </code></pre></div>
  </div>
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

> Pay attention, werf uses actual _base image_ digest in stage digest if _fromLatest_ is specified. Thus, the usage of this directive might break the reproducibility of previous builds. If the base image is changed in the registry, all previously built stages become not usable.
>
> * Previous pipeline jobs (e.g. deploy) cannot be retried without the image rebuild after changing base image in the registry.
> * If base image is modified unexpectedly it might lead to the inexplicably failed pipeline. For instance, the modification occurs after successful build and the following jobs will be failed due to changing of stages digests alongside base image digest.
>
> If you want to use this directive, you need to disable determinism mode with `--disable-determinism` option.
>
> **We do not recommend using the actual base image such way**. Use a particular unchangeable tag or periodically change [fromCacheVersion](#fromcacheversion) value to provide controllable and predictable lifecycle of software       

## fromImage and fromArtifact

Besides using docker image from a repository, the _base image_ can refer to _image_ or [_artifact_]({{ "documentation/advanced/building_images_with_stapel/artifacts.html" | relative_url }}), that is described in the same `werf.yaml`.

```yaml
fromImage: <image name>
fromArtifact: <artifact name>
```

If a _base image_ is specific to a particular application,
it is reasonable to store its description with _images_ and _artifacts_ which are used it as opposed to storing the _base image_ in a Docker registry.

Also, this method can be useful if the stages of _stage conveyor_ are not enough for building the image. You can design your _stage conveyor_.

<a class="google-drawings" href="../../../images/configuration/base_image2.png" data-featherlight="image">
<img src="../../../images/configuration/base_image2_preview.png" alt="Conveyor with fromImage and fromArtifact stages">
</a>

## fromCacheVersion

The `fromCacheVersion` directive allows to manage image reassembly.

```yaml
fromCacheVersion: <arbitrary string>
```
