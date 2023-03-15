---
title: Images
permalink: usage/distribute/images.html
---

## Distribution of images 

The `werf export` command distributes the assembled werf images and adapts them for use by third-party users and/or software. This command builds and publishes the images in the container registry and removes all metadata that is not needed by third-party software. As a result, the images are no longer controlled by werf, so you can use any third-party tools to control their lifecycle.

> Images published with the `werf export` command will *never* be deleted with the `werf cleanup` command as opposed to images published in the usual way. Cleanup of exported images must be implemented by third-party tools.

## Image distribution

```shell
werf export \
    --repo example.org/myproject \
    --tag other.example.org/myproject/myapp:latest
```

As a result of running the above command, the image will be built and initially published with a content-based tag to the `example.org/myproject` container registry. The image will then be published to another container registry (`other.example.org/myproject`) as the `other.example.org/myproject/myapp:latest` final exported image.

You can specify the same repository in the `--tag` parameter as in `--repo`, thus using the same container registry for both the build and the exported image.

## Distributing multiple images

The `--tag` parameter supports the `%image%`, `%image_slug%`, and `%image_safe_slug%` patterns to substitute an image name from `werf.yaml` based on its contents, for example:

```shell
werf export \
    --repo example.org/mycompany/myproject \
    --tag example.org/mycompany/myproject/%image%:latest
```

## Distributing arbitrary images

You can select the images to be published using positional arguments and image names from `werf.yaml`, for example:

```shell
werf export backend frontend \
    --repo example.org/mycompany/myproject \
    --tag example.org/mycompany/myproject/%image%:latest
```

## Using a content-based tag to generate a tag

With the `%image_content_based_tag%` pattern, you can use the content-based tag in the `--tag` parameter, for example:

```shell
werf export \
    --repo example.org/mycompany/myproject \
    --tag example.org/mycompany/myproject/myapp:%image_content_based_tag%
```

## Adding extra labels

Using the `--add-label` parameter, you can add an arbitrary number of additional labels to the image(s) being exported, for example:

```shell
werf export \
    --repo example.org/mycompany/myproject \
    --tag registry.werf.io/werf/werf:latest \
    --add-label io.artifacthub.package.readme-url=https://raw.githubusercontent.com/werf/werf/main/README.md \
    --add-label org.opencontainers.image.created=2023-03-13T11:55:24Z \
    --add-label org.opencontainers.image.description="Official image to run werf in containers"
```