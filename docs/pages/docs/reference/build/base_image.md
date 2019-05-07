---
title: Base image
sidebar: reference
permalink: docs/reference/build/base_image.html
author: Alexey Igrychev <alexey.igrychev@flant.com>
summary: |
  <a class="google-drawings" href="https://docs.google.com/drawings/d/e/2PACX-1vReDSY8s7mMtxuxwDTwtPLFYjEXePaoIB-XbEZcunJGNEHrLbrb9aFxyOoj_WeQe0XKQVhq7RWnG3Eq/pub?w=2031&amp;h=144" data-featherlight="image">
      <img src="https://docs.google.com/drawings/d/e/2PACX-1vReDSY8s7mMtxuxwDTwtPLFYjEXePaoIB-XbEZcunJGNEHrLbrb9aFxyOoj_WeQe0XKQVhq7RWnG3Eq/pub?w=1016&amp;h=72">
  </a>

  <div class="language-yaml highlighter-rouge"><pre class="highlight"><code><span class="s">from</span><span class="pi">:</span> <span class="s">&lt;image[:&lt;tag&gt;]&gt;</span>
  <span class="s">fromLatest</span><span class="pi">:</span> <span class="s">&lt;false || true&gt;</span>
  <span class="s">fromCacheVersion</span><span class="pi">:</span> <span class="s">&lt;arbitrary string&gt;</span>
  <span class="s">fromImage</span><span class="pi">:</span> <span class="s">&lt;image name&gt;</span>
  <span class="s">fromImageArtifact</span><span class="pi">:</span> <span class="s">&lt;artifact name&gt;</span></code></pre>
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

_Base image_ can be declared with `from`, `fromImage` or `fromImageArtifact` directive.

## from, fromLatest

The `from` directive defines the name and tag of a _base image_. If absent, tag defaults to `latest`.

```yaml
from: <image>[:<tag>]
```

By default, the assembly process does not depend on actual _base image_ digest in the repository, only on _from_ directive value. 
Thus, changing _base image_ locally or in the repository does not matter if _from_ stage is already exists in _stages storage_.

If you want always build the image with actual _base image_ you should use _fromLatest_ directive. 
_fromLatest_ directive allows connecting the assembly process with the _base image_ digest getting from the repository.
```yaml
fromLatest: true
```

> Pay attention, werf uses actual _base image_ digest as extra _from_ stage dependency if _fromLatest_ is true. 
Therefore, using this directive implies not reproducible signatures:
after changing _base image_ in repository, all previously built stages, also like related images, become not usable. 
The problem might occur:
- between jobs of one pipeline (e.g. build and deploy) or 
- when you rerun the previous job (e.g. deploy)

## fromImage and fromImageArtifact

Besides using docker image from a repository, the _base image_ can refer to _image_ or [_artifact_]({{ site.baseurl }}/reference/build/artifact.html), that is described in the same `werf.yaml`.

```yaml
fromImage: <image name>
fromImageArtifact: <artifact name>
```

If a _base image_ is specific to a particular application,
it is reasonable to store its description with _images_ and _artifacts_ which are used it as opposed to storing the _base image_ in a registry.

Also, this method can be useful if the stages of _stage conveyor_ are not enough for building the image. You can design your _stage conveyor_.

<a class="google-drawings" href="https://docs.google.com/drawings/d/e/2PACX-1vTmQBPjB6p_LUpwiae09d_Jp0JoS6koTTbCwKXfBBAYne9KCOx2CvcM6DuD9pnopdeHF--LPpxJJFhB/pub?w=1629&amp;h=1435" data-featherlight="image">
<img src="https://docs.google.com/drawings/d/e/2PACX-1vTmQBPjB6p_LUpwiae09d_Jp0JoS6koTTbCwKXfBBAYne9KCOx2CvcM6DuD9pnopdeHF--LPpxJJFhB/pub?w=850&amp;h=673">
</a>

## fromCacheVersion

The `fromCacheVersion` directive allows to manage image reassembly.

```yaml
fromCacheVersion: <arbitrary string>
```
