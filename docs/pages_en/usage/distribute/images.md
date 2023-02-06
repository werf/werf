---
title: Images
permalink: usage/distribute/images.html
---

## Distributing images to deploy by third-party software

The distribution of werf images for third-party software is carried out with the `werf export` command. This command will build and publish the images to the container registry as well as remove all metadata not needed for third-party software. This way, the command will completely exclude the images from werf's control so that third-party tools can manage their lifecycle.

Example:

```shell
werf export --repo example.org/myproject --tag other.example.org/myproject/myapp:latest
```

As a result of running the above command, the image will be built and initially published with a content-based tag to the `example.org/myproject` container registry. The image will then be published to another container registry (`other.example.org/myproject`) as the `other.example.org/myproject/myapp:latest` final exported image.

You can specify the same repository in the `--tag` parameter as in `--repo`, thus using the same container registry for both the build and the exported image.

You can also use the `%image%`, `%image_slug%` and `%image_safe_slug%` patterns to substitute an image name and `%image_content_based_tag%` to substitute an image tag based on the image content, for example:

```shell
werf export --repo example.org/mycompany/myproject --tag example.org/mycompany/myproject/%image%:%image_content_based_tag%
```

> The images published with the `werf export` command will *never* be deleted by the `werf cleanup` command, unlike images published in the usual way. You can use third-party tools to clean up exported images.