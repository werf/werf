---
title: Build process
permalink: usage/build/process.html
---

{% include pages/en/cr_login.md.liquid %}

## Tagging images

<!-- reference https://werf.io/documentation/v1.2/internals/stages_and_storage.html#stage-naming -->

The tagging of werf images is performed automatically as part of the build process. werf uses an optimal tagging scheme based on the contents of the image, thus preventing unnecessary rebuilds and application wait times during deployment.

<div class="details">
<a href="javascript:void(0)" class="details__summary">Tagging in details</a>
<div class="details__content" markdown="1">

By default, a **tag is some hash-based identifier** which includes the checksum of instructions and build context files. For example:

```
registry.example.org/group/project  d4bf3e71015d1e757a8481536eeabda98f51f1891d68b539cc50753a-1589714365467  7c834f0ff026  20 hours ago  66.7MB
registry.example.org/group/project  e6073b8f03231e122fa3b7d3294ff69a5060c332c4395e7d0b3231e3-1589714362300  2fc39536332d  20 hours ago  66.7MB
```

Such a tag will only change when the underlying data used to build an image changes. This way, one tag can be reused for different Git commits. werf:
- calculates these tags;
- atomically publishes the images based on these tags to the repository or locally;
- passes the tags to the Helm chart.

The images in the repository are named according to the following scheme: `CONTAINER_REGISTRY_REPO:DIGEST-TIMESTAMP_MILLISEC`. Here:

- `CONTAINER_REGISTRY_REPO` — repository defined by the `--repo` option;
- `DIGEST` — the checksum calculated for:
  - build instructions defined in the Dockerfile or `werf.yaml`;
  - build context files used in build instructions.
- `TIMESTAMP_MILLISEC` — timestamp that is added while [saving a layer to the container registry](#layer-by-layer-image-caching) after the stage has been built.

The assembly algorithm also ensures that the image with such a tag is unique and that tag will never be overwritten by an image with different content.

**Tagging intermediate layers**

The automatically generated tag described above is used for both the final images which the user runs and for intermediate layers stored in the container registry. Any layer found in the repository can either be used as an intermediate layer to build a new layer based on it, or as a final image.

</div>
</div>

### Retrieving tags

You can retrieve image tags using the `--save-build-report` option for `werf build`, `werf converge`, and other commands:

```shell
# By default, the JSON format is used.
werf build --save-build-report --repo REPO

# The envfile format is also supported.
werf converge --save-build-report --build-report-path .werf-build-report.env --repo REPO

# For the render command, the final tags are only available with the --repo parameter.
werf render --save-build-report --repo REPO
```

> **NOTE:** Retrieving tags beforehand without first invoking the build process is currently impossible. You can only retrieve tags from the images you've already built.

### Adding custom tags

The user can add any number of custom tags using the `--add-custom-tag` option:

```shell
werf build --repo REPO --add-custom-tag main

# You can add some alias tags.
werf build --repo REPO --add-custom-tag main --add-custom-tag latest --add-custom-tag prerelease
```

The tag template may include the following parameters:

- `%image%`, `%image_slug%` or `%image_safe_slug%` to use an image name defined in `werf.yaml` (mandatory when building multiple images);
- `%image_content_based_tag%` to use the content-based werf tag.

```shell
werf build --repo REPO --add-custom-tag "%image%-latest"
```

> **NOTE:** When you use the options listed above, werf still creates the **additional alias tags** that reference the automatic hash tags. It is not possible to completely disable auto-tagging.

## Layer-by-layer image caching

<!-- reference https://werf.io/documentation/v1.2/internals/stages_and_storage.html#storage -->

Layer-by-layer image caching is essential part of the werf build process. werf saves and reuses the build cache in the container registry and synchronizes parallel builders.

<div class="details">
<a href="javascript:void(0)" class="details__summary">How assembly works</a>
<div class="details__content" markdown="1">

Building in werf deviates from the standard Docker paradigm of separating the `build` and `push` stages. It consists of a single `build` stage, which combines both building and publishing layers.

A standard approach for building and publishing images and layers via Docker might look like this:
1. Downloading the build cache from the container registry (optional).
2. Local building of all the intermediate image layers using the local layer cache.
3. Publishing the built image.
4. Publishing the local build cache to the container registry (optional).

The image building algorithm in werf is different:
1. If the next layer to be built is already present in the container registry, it will not be built or downloaded.
2. If the next layer to be built is not in the container registry, the previous layer is downloaded (the base layer for building the current one).
3. The new layer is built on the local machine and published to the container registry.
4. At publishing time, werf automatically resolves conflicts between builders from different hosts that try to publish the same layer. This ensures that only one layer is published, and all other builders are required to reuse that layer. ([The built-in sync service](#synchronizing-builders) makes this possible).
5. The process continues until all the layers of the image are built.

The algorithm of stage selection in werf works as follows:
1. werf calculates the [stage digest](#tagging-images).
2. Then it selects all the stages matching the digest, since several stages in the repository may be tied to a single digest.
3. For the Stapel builder, if the current stage involves Git (a Git archive stage, a custom stage with Git patches, or a `git latest patch` stage), then only those stages associated with commits that are ancestral to the current commit are selected. Thus, commits from neighboring branches will be discarded.
4. Then the oldest `TIMESTAMP_MILLISEC` is selected.

If you run a build with storing images in the repository, werf will first check if the required stages exist in the local repository and copy the suitable stages from there, so that no rebuilding of those stages is necessary.

</div>
</div>

> **NOTE:** It is assumed that the image repository for the project will not be deleted or cleaned by third-party tools, otherwise it will have negative consequences for users of a werf-based CI/CD ([see image cleanup]({{"usage/cleanup/cr_cleanup.html" | true_relative_url }})).

### Dockerfile

By default, Dockerfile images are cached by a single image in the container registry.

To enable layered caching of Dockerfile instructions in the container registry, use the `staged` directive in werf.yaml:

```yaml
# werf.yaml
image: example
dockerfile: ./Dockerfile
staged: true
```

<div class="details">
<a href="javascript:void(0)" class="details__summary">**NOTE**: The staged Dockerfile caching feature is currently alpha</a>
<div class="details__content" markdown="1">

There are several generations of the staged dockerfile builder. You can switch between them using the `WERF_STAGED_DOCKERFILE_VERSION={v1|v2}` variable. Note that changing the version of a staged dockerfile can cause images to be rebuilt.

* `v1` is used by default.
* `v2` version enables a dedicated `FROM` layer for caching the base image specified in the `FROM` instruction.

`v2` version compatibility may be broken in future releases.

</div>
</div>

### Stapel

Stapel images are cached layer-by-layer in the container registry by default and do not require any configuration.

## Parallelism and image assembly order

<!-- reference: https://werf.io/documentation/v1.2/internals/build_process.html#parallel-build -->

All the images described in `werf.yaml` are built in parallel on the same build host. If there are dependencies between the images, the build is split into stages, with each stage containing a set of independent images that can be built in parallel.

> When Dockerfile stages are used, the parallelism of their assembly is also determined based on the dependency tree. On top of that, if different images use a Dockerfile stage declared in `werf.yaml`, werf will make sure that this common stage is built only once, without any redundant rebuilds.

The parallel assembly in werf is regulated by two parameters: `--parallel` and `--parallel-tasks-limit`. By default, the parallel build is enabled and no more than 5 images can be built at a time.

Let's look at the following example:

```Dockerfile
# backend/Dockerfile
FROM node as backend
WORKDIR /app
COPY package*.json /app/
RUN npm ci
COPY . .
CMD ["node", "server.js"]
```

```Dockerfile
# frontend/Dockerfile

FROM ruby as application
WORKDIR /app
COPY Gemfile* /app
RUN bundle install
COPY . .
RUN bundle exec rake assets:precompile
CMD ["rails", "server", "-b", "0.0.0.0"]

FROM nginx as assets
WORKDIR /usr/share/nginx/html
COPY configs/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=application /app/public/assets .
COPY --from=application /app/vendor .
ENTRYPOINT ["nginx", "-g", "daemon off;"]
```

```yaml
image: backend
dockerfile: Dockerfile
context: backend
---
image: frontend
dockerfile: Dockerfile
context: frontend
target: application
---
image: frontend-assets
dockerfile: Dockerfile
context: frontend
target: assets
```

There are 3 images: `backend`, `frontend` and `frontend-assets`. The `frontend-assets` image depends on `frontend` because it imports compiled assets from `frontend`.

In this case, werf will compose the following sets to build:

```shell
┌ Concurrent builds plan (no more than 5 images at the same time)
│ Set #0:
│ - ⛵ image backend
│ - ⛵ image frontend
│
│ Set #1:
│ - ⛵ frontend-assets
└ Concurrent builds plan (no more than 5 images at the same time)
```

## Using container registry

In werf, the container registry is used not only to store the final images, but also to store the build cache and service data required for werf (e.g., metadata for cleaning the container registry based on Git history). The container registry is set by the `--repo` parameter:

```shell
werf converge --repo registry.mycompany.org/project
```

There are a number of additional repositories on top of the main repository:

- `--final-repo` to store the final images in a dedicated repository;
- `--secondary-repo` to use the repository in `read-only` mode (e.g. to use a container registry CI that you cannot push into, but you can reuse the build cache);
- `--cache-repo` to set the repository containing the build cache alongside the builders.

> **Caution!** For werf to operate properly, the container registry must be persistent, and cleaning should only be done with the `werf cleanup` special command.

### Extra repository for final images

If necessary, the so-called **final** repositories can be used to exclusively store the final images.

```shell
werf build --repo registry.mycompany.org/project --final-repo final-registry.mycompany.org/project-final
```

Final repositories reduce image retrieval time and network load by bringing the container registry closer to the Kubernetes cluster on which the application is being deployed. Final repositories can also be used in the same container registry as the main repository (`--repo`), if necessary.

### Extra repository for quick access to the build cache

You can specify one or more so-called **caching** repositories using the `--cache-repo` parameter.

```shell
# An extra caching repository on the local network.
werf build --repo registry.mycompany.org/project --cache-repo localhost:5000/project
```

A caching repository can help reduce build cache loading times. However, for this to work, download speeds from a caching repository must be significantly higher than those from the main repository. This is usually achieved by hosting a container registry on the local network, but it is not mandatory.

Caching repositories have higher priority than the main repository when the build cache is retrieved. When caching repositories are used, the build cache remains stored in the main repository as well.

You can clean up a caching repository by deleting it entirely without any risks.

## Synchronizing builders

<!-- reference https://werf.io/documentation/v1.2/advanced/synchronization.html -->

To ensure consistency among parallel builders and to guarantee the reproducibility of images and intermediate layers, werf handles the synchronization of the builders. By default, the public synchronization service at [https://synchronization.werf.io/](https://synchronization.werf.io/) is used and no extra user interaction is required.

<div class="details">
<a href="javascript:void(0)" class="details__summary">How the synchronization service works</a>
<div class="details__content" markdown="1">

The synchronization service is a werf component that is designed to coordinate multiple werf processes. It acts as a _lock manager_. The locks are required to correctly publish new images to the container registry and to implement the build algorithm described in ["Layer-by-layer image caching"](#layer-by-layer-image-caching).

The data sent to the sync service are anonymized and are hash sums of the tags published in the container registry.

A synchronization service can be:
1. An HTTP synchronization server implemented in the `werf synchronization` command.
2. The ConfigMap resource in a Kubernetes cluster. The mechanism used is the [lockgate](https://github.com/werf/lockgate) library, which implements distributed locks by storing annotations in the selected resource.
3. Local file locks provided by the operating system.

</div>
</div>

### Using your own synchronization service

#### HTTP server

The synchronization server can be run with the `werf synchronization` command. In the example below, port 55581 (the default one) is used:

```shell
werf synchronization --host 0.0.0.0 --port 55581
```

— This server only supports HTTP mode. To use HTTPS, you have to configure additional SSL termination by third-party tools (e.g., via the Kubernetes Ingress).

Then, for all werf commands that use the `--repo` parameter, the `--synchronization=http[s]://DOMAIN` parameter must be specified as well, for example:

```shell
werf build --repo registry.mydomain.org/repo --synchronization https://synchronization.domain.org
werf converge --repo registry.mydomain.org/repo --synchronization https://synchronization.domain.org
```

#### Dedicated Kubernetes resource

You only have to specify a running Kubernetes cluster and choose the namespace where the ConfigMap/werf service will reside. Its annotations will be used for distributed locking.

Then, for all werf commands that use the `--repo` parameter, the `--synchronization=kubernetes://NAMESPACE[:CONTEXT][@(base64:CONFIG_DATA)|CONFIG_PATH]` parameter must be specified as well, for example:

```shell
# The regular ~/.kube/config or KUBECONFIG is used.
werf build --repo registry.mydomain.org/repo --synchronization kubernetes://mynamespace
werf converge --repo registry.mydomain.org/repo --synchronization kubernetes://mynamespace

# Here, the base64-encoded contents of kubeconfig are explicitly specified.
werf build --repo registry.mydomain.org/repo --synchronization kubernetes://mynamespace@base64:YXBpVmVyc2lvbjogdjEKa2luZDogQ29uZmlnCnByZWZlcmVuY2VzOiB7fQoKY2x1c3RlcnM6Ci0gY2x1c3RlcjoKICBuYW1lOiBkZXZlbG9wbWVudAotIGNsdXN0ZXI6CiAgbmFtZTogc2NyYXRjaAoKdXNlcnM6Ci0gbmFtZTogZGV2ZWxvcGVyCi0gbmFtZTogZXhwZXJpbWVudGVyCgpjb250ZXh0czoKLSBjb250ZXh0OgogIG5hbWU6IGRldi1mcm9udGVuZAotIGNvbnRleHQ6CiAgbmFtZTogZGV2LXN0b3JhZ2UKLSBjb250ZXh0OgogIG5hbWU6IGV4cC1zY3JhdGNoCg==

# The mycontext context is used in the /etc/kubeconfig config.
werf build --repo registry.mydomain.org/repo --synchronization kubernetes://mynamespace:mycontext@/etc/kubeconfig
```

> **NOTE:** This method is poorly suited when the project is delivered to different Kubernetes clusters from the same Git repository due to the difficulties of setting it up correctly. In this case, the same cluster address and resource must be specified for all werf commands even if the deployment occurs to different environments to ensure data consistency in the container registry. Therefore, it is recommended to run a dedicated shared synchronization service for this case to avoid the risk of incorrect configuration.

#### Local synchronization

Local synchronization is enabled by the `--synchronization=:local` option. The local _lock manager_ uses file locks provided by the operating system.

```shell
werf build --repo registry.mydomain.org/repo --synchronization :local
werf converge --repo registry.mydomain.org/repo --synchronization :local
```

> **NOTE:** This method is only suitable if all werf runs are triggered by the same runner in your CI/CD system.

## Multi-platform builds

Multi-platform builds use the cross-platform instruction execution mechanics provided by the [Linux kernel](https://en.wikipedia.org/wiki/Binfmt_misc) and the QEMU emulator. [List of supported architectures](https://www.qemu.org/docs/master/about/emulation.html). Refer to the [Installation]({{ "index.html" | true_relative_url }}) section for more information on how to configure the host system to do cross-platform builds.

The table below summarizes support of multi-platform building for different configuration syntaxes, building modes, and build backends:

|                         | buildah        | docker-server      |
|---                      |---             |---                 |
| **Dockerfile**          | full support   | full support       |
| **staged Dockerfile**   | full support   | no support         |
| **stapel**              | full support   | linux/amd64 only   |
