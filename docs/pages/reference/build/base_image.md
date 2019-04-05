---
title: Base image
sidebar: reference
permalink: reference/build/base_image.html
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

## from and fromCacheVersion

The `from` directive sets the name and tag of a _base image_ to build _from_ stage. If absent, tag defaults to `latest`.

```yaml
from: <image>[:<tag>]
```

On _from_ stage, _from image_ is pulled from a repository and saved in the [_stages cache_]({{ site.baseurl }}/reference/build/stages.html). If _from_ stage is using cache image won't be pulled.

Pay attention, _from_ stage depends on a variable of _from_ directive, not on _from image_ digest or something else.

If you have, e.g., _from image_ `alpine:latest`, and need to rebuild _image_ with the actual latest alpine image you should set _fromCacheVersion_ (`fromCacheVersion: <arbitrary string>`) and delete existent _from image_ from local docker or manually pull it.

If you want always build image with actual latest _from image_ you should automate this process outside of werf and use a particular environment variable for `fromCacheVersion` (e.g., the variable with _from image_ digest).

{% raw %}
```yaml
from: "alpine:latest"
fromCacheVersion: {{ env "FROM_IMAGE_DIGEST" }}
```
{% endraw %}

## fromImage and fromImageArtifact

Besides using docker image from a repository, _base image_ can refer to _image_ or [_artifact_]({{ site.baseurl }}/reference/build/artifact.html), described in the same `werf.yaml`.

```yaml
fromImage: <image name>
fromImageArtifact: <artifact name>
```

If a _base image_ is specific to a particular application,
it is reasonable to store its description with _images_ and _artifacts_ which are used it as opposed to storing the _base image_ in a registry.

Also, this method can be helpful if the stages of _stage conveyor_ are not enough for building an image. You can design your _stage conveyor_.

<a class="google-drawings" href="https://docs.google.com/drawings/d/e/2PACX-1vTmQBPjB6p_LUpwiae09d_Jp0JoS6koTTbCwKXfBBAYne9KCOx2CvcM6DuD9pnopdeHF--LPpxJJFhB/pub?w=1629&amp;h=1435" data-featherlight="image">
<img src="https://docs.google.com/drawings/d/e/2PACX-1vTmQBPjB6p_LUpwiae09d_Jp0JoS6koTTbCwKXfBBAYne9KCOx2CvcM6DuD9pnopdeHF--LPpxJJFhB/pub?w=850&amp;h=673">
</a>
