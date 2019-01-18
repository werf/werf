---
title: Naming
sidebar: reference
permalink: reference/build/naming.html
author: Alexey Igrychev <alexey.igrychev@flant.com>
---

Images are declared with _dimg_ directive: `dimg: <dimg name>`. You can use this name for most of the commands to execute ones for specific _dimgs_. Also, _dimg name_ uses when tagging and pushing built images into registry (read about it in separate [article]({{ site.baseurl }}/reference/registry/push.html)).

The _dimg_ directive starts a description for building an application image.
The _dimg name_ is a string, similar to the image name in Docker:

```yaml
dimg: frontend
```

A _dimg_ can be nameless, `dimg: ` or `dimg: ~`, but only if _dimg_ is single in config. In a config with multiple _dimgs_ there cannot be nameless _dimg_.

```yaml
dimg: frontend
...
---
dimg: backend
...
```

A _dimg_ can have several names, set as a list in YAML syntax.
This usage is equal to describing similar images with different names.

```yaml
dimg: [main-front,main-back]
```
