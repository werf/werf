Images are declared with _image_ directive: `image: <image name>`. 
The _image_ directive starts a description for building an application image.
The _image name_ is a string, similar to the image name in Docker:

```yaml
image: frontend
```

If _image_ only one in the config, it can be nameless:

```yaml
image: ~
```

In the config with multiple, **all images** must have names:

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

You can use _image name_ for most commands to execute ones for specific _image(s)_:
* [werf build \[IMAGE_NAME...\] \[options\]]({{ site.baseurl }}/documentation/reference/cli/werf_build.html)
* [werf run \[options\] \[IMAGE_NAME\] \[-- COMMAND ARG...\]]({{ site.baseurl }}/documentation/reference/cli/werf_run.html)
