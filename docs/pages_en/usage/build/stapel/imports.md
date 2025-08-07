---
title: Importing from images
permalink: usage/build/stapel/imports.html
author: Alexey Igrychev <alexey.igrychev@flant.com>
directive_summary: import
---

The size of the final image can grow dramatically due to the assembly tools and source files eating up space. These files are generally not needed in the final image.
To avoid this, the Docker community suggests installing tools, building, and removing irrelevant files in one step:

```Dockerfile
RUN “download-source && cmd && cmd2 && remove-source”
```

> You can do the same in werf — just specify the relevant instructions for some _user stage_. Below is an example of specifying the _shell assembly instructions_ for the _install stage_ (you can do so for the _ansible_ builder as well):
```yaml
shell:
  install:
  - "download-source"
  - "cmd"
  - "cmd2"
  - "remove-source"
```

However, this method does not support caching. Thus, a build toolkit will be installed all over again.

Another way is to use a multi-stage build, a feature supported in Docker since version 17.05:

```Dockerfile
FROM node:latest AS storefront
WORKDIR /app
COPY react-app .
RUN npm install
RUN npm run build

FROM maven:latest AS appserver
WORKDIR /app
COPY . .
RUN mvn package

FROM java:8-jdk-alpine
COPY --from=storefront /app/react-app/build/ /static
COPY --from=appserver /app/target/AtSea-0.0.1-SNAPSHOT.jar /app/AtSea.jar
```

The point of this approach is to describe several auxiliary images and selectively copy artifacts from one image to another, leaving all the unnecessary data outside the final image.

werf offers the same approach.

> Why doesn't werf use multi-stage assembly?
* Historically, _imports_ came much earlier than the Docker multi-stage mechanism, and
* werf allows for greater flexibility when working with auxiliary images.

## Import configuration

Importing _resources_ from the _images_ must be described in the `import` directive in the _destination image_ in the _image_ config section. `import` is an array of records, where each record must contain the following:

- `image: <image name>`: _source image_; the name of the image copy files from.
- `from: <image name>`: _source image_; the name of the image copy files from. Both imports from images of the current project and from external images in the format `image_name:tag` or `image_name@digest` are supported.
- `stage: <stage name>`: _source image stage_; the stage of the _source_image_ to copy files from.
- `add: <absolute path>`: _source path_; the absolute path to the file or directory in the _source image_ to copy from.
- `to: <absolute path>`: _destination path_; the absolute path in the _destination image_. If absent, the _destination path_ defaults to the  _source path_ (as specified by the `add` directive).
- `before: <install || setup>` or `after: <install || setup>`: _destination image stage_; the stage to import files. Currently, only _install_ and _setup_ stages are supported.

An example of the `import` directive:

```yaml
import:
  - from: application-assets
    add: /app/public/assets
    to: /var/www/site/assets
    after: install
  - from: node:20-alpine
    add: /usr/share/nginx/html
    to: /var/www/site/prebuilt
    after: setup
```

As with the _git mappings_ configuration, include and exclude file and directory masks are supported (`includePaths: []` and `excludePaths: []`, respectively). Masks must be specified relative to the source path (as in the `add` parameter).
You can also specify an owner and a group for the imported resources, `owner: <owner>` and `group: <group>`.
This behavior is similar to the one used when adding code from Git repositories, and you can read more about it in the [git directive section]({{ "usage/build/stapel/git.html" | true_relative_url }}).

> Note that the path of imported resources and the path specified in _git mappings_ must not overlap.
