---
title: Naming
sidebar: reference
permalink: reference/build/naming.html
author: Alexey Igrychev <alexey.igrychev@flant.com>
---

Images are declared with _image_ directive: `image: <image name>`. You can use this name for most of the commands to execute ones for specific _images_. Also, _image name_ uses when tagging and pushing built images into registry (read about it in separate [article]({{ site.baseurl }}/reference/registry/push.html)).

The _image_ directive starts a description for building an application image.
The _image name_ is a string, similar to the image name in Docker:

```yaml
image: frontend
```

A _image_ can be nameless, `image: ` or `image: ~`, but only if _image_ is single in config. In a config with multiple _images_ there cannot be nameless _image_.

```yaml
image: frontend
...
---
image: backend
...
```

A _image_ can have several names, set as a list in YAML syntax.
This usage is equal to describing similar images with different names.

```yaml
image: [main-front,main-back]
```
