---
title: Importing from images and artifacts
permalink: advanced/building_images_with_stapel/import_directive.html
author: Alexey Igrychev <alexey.igrychev@flant.com>
directive_summary: import
---

The size of the image can be increased several times due to the assembly tools and source files, while the user does not need them.
To solve such problems, Docker community suggests doing the installation of tools, the assembly, and removal in one step.

```
RUN “download-source && cmd && cmd2 && remove-source”
```

> To obtain a similar effect when using werf, it is sufficient describing the instructions in one _user stage_. For example, _shell assembly instructions_ for the _install stage_ (similarly for _ansible_):
```yaml
shell:
  install:
  - "download-source"
  - "cmd"
  - "cmd2"
  - "remove-source"
```

However, with this method, it is not possible using a cache, so toolkit installation runs each time.

Another solution is using multi-stage builds, which are supported starting with Docker 17.05.

```
FROM node:latest AS storefront
WORKDIR /usr/src/atsea/app/react-app
COPY react-app .
RUN npm install
RUN npm run build

FROM maven:latest AS appserver
WORKDIR /usr/src/atsea
COPY pom.xml .
RUN mvn -B -f pom.xml -s /usr/share/maven/ref/settings-docker.xml dependency:resolve
COPY . .
RUN mvn -B -s /usr/share/maven/ref/settings-docker.xml package -DskipTests

FROM java:8-jdk-alpine
RUN adduser -Dh /home/gordon gordon
WORKDIR /static
COPY --from=storefront /usr/src/atsea/app/react-app/build/ .
WORKDIR /app
COPY --from=appserver /usr/src/atsea/target/AtSea-0.0.1-SNAPSHOT.jar .
ENTRYPOINT ["java", "-jar", "/app/AtSea-0.0.1-SNAPSHOT.jar"]
CMD ["--spring.profiles.active=postgres"]
```

The meaning of such an approach is as follows, describe several auxiliary images and selectively copy artifacts from one image to another leaving behind everything you do not want in the result image.

We suggest the same, but using [_images_]({{ "reference/werf_yaml.html#image-section" | true_relative_url }}) and [_artifacts_]({{ "reference/werf_yaml.html#image-section" | true_relative_url }}).

> Why is werf not using multi-stage?
* Historically, _imports_ appeared much earlier than Docker multi-stage, and
* werf gives more flexibility working with auxiliary images

Importing _resources_ from _images_ and _artifacts_ should be described in `import` directive in _destination image_ config section ([_image_]({{ "reference/werf_yaml.html#image-section" | true_relative_url }}) or [_artifact_]({{ "reference/werf_yaml.html#image-section" | true_relative_url }}). `import` is an array of records. Each record should contain the following:

- `image: <image name>` or `artifact: <artifact name>`: _source image_, image name from which you want to copy files.
- `stage: <stage name>`: _source image stage_, particular stage of _source_image_ from which you want to copy files.
- `add: <absolute path>`: _source path_, absolute file or folder path in _source image_ for copying.
- `to: <absolute path>`: _destination path_, absolute path in _destination image_. In case of absence, _destination path_ equals _source path_ (from `add` directive).
- `before: <install || setup>` or `after: <install || setup>`: _destination image stage_, stage for importing files. At present, only _install_ and _setup_ stages are supported.

```yaml
import:
- artifact: application-assets
  add: /app/public/assets
  to: /var/www/site/assets
  after: install
- image: frontend
  add: /app/assets
  after: setup
```

As in the case of adding _git mappings_, masks are supported for including, `include_paths: []`, and excluding files, `exclude_paths: []`, from the specified path.
You can also define the rights for the imported resources, `owner: <owner>` and `group: <group>`.
Read more about these in the [git directive article]({{ "advanced/building_images_with_stapel/git_directive.html" | true_relative_url }}).

> Import paths and _git mappings_ must not overlap with each other

Information about _using artifacts_ available in [separate article]({{ "advanced/building_images_with_stapel/artifacts.html" | true_relative_url }}).
