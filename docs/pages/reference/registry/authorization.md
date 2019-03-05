---
title: Authorization
sidebar: reference
permalink: reference/registry/authorization.html
author: Timofey Kirillov <timofey.kirillov@flant.com>
---

There are several categories of commands that work with docker registry, thus need to authorize in docker registry:

* [Build command]({{ site.baseurl }}/reference/build/assembly_process.html#werf-build) pull base images from docker registry.
* [Push commands]({{ site.baseurl }}/reference/registry/push.html) used to create and update images in docker registry.
* [Pull commands]({{ site.baseurl }}/reference/registry/pull.html) used to pull build cache from docker registry.
* [Cleaning commands]({{ site.baseurl }}/reference/registry/cleaning.html) used to delete images from docker registry.
* [Deploy commands]({{ site.baseurl }}/reference/deploy/deploy_to_kubernetes.html#werf-kube-deploy) used to read meta information about images from docker registry during deploy.

Werf uses the same procedure to authorize in the docker registry as the docker client does. Docker client command `docker login` is used to authorize in docker registry. This command creates new **docker config** directory (`~/.docker` by default) or updates existing one with credentials that can be used to make docker registry operations. The process of creating such config will be referred as **login procedure**.

However, unlike docker werf does not have separate `login` command to perform login procedure.

## How to perform a login procedure with werf

There are two options:

1. Prepare docker config externally.
2. Let werf transparently perform login procedure on each call to werf command, i.e. login by werf.

### Prepare docker config externally

Standard docker login command, for example, can be used to prepare [docker config](https://docs.docker.com/engine/reference/commandline/cli/#configuration-files).

Once the docker config has been created its path should be specified in the environment variable `WERF_DOCKER_CONFIG` for all werf commands. The variable value is the same as `docker --config` standard option value.

### Login by werf

With this method, werf command will create a *temporary* docker config and remove it after command termination on each command call. The standard docker config path will not be overridden.

Thus login procedure performed transparently on every werf command execution.

#### Autologin

By default without options werf will try to use username `gitlab-ci-token` and password from environment variable `CI_JOB_TOKEN`. Therefore gitlab token used to login into specified docker registry when werf command being run inside gitlab job. This is referred to as **autologin** procedure. When gitlab environment is not available werf will not perform autologin procedure.

[Gitlab container registry](https://docs.gitlab.com/ee/user/project/container_registry.html) should be enabled for this scheme.

When one of [google container registry](https://cloud.google.com/container-registry/) official address is specified werf will not perform autologin even if the gitlab environment is available.

To manually **disable** autologin procedure specify environment variable `WERF_IGNORE_CI_DOCKER_AUTOLOGIN=1`.

#### Autologin for build commands

For [build commands]({{ site.baseurl }}/cli/main/build.html) werf uses pull procedure for images specified as base (see [from directive]({{ site.baseurl }}/reference/build/base_image.html#from-and-fromcacheversion)).

Running in gitlab environment werf build command will perform autologin procedure for the gitlab container registry under these conditions:

* gitlab container registry is enabled and `CI_REGISTRY` variable with the docker registry address is available;
* image specified in `from` directive belongs to this gitlab container registry.

For example, given this build configuration:

```yaml
from: registry.myhost.com/sys/project/image:v2.1.0
```

Werf will perform autologin into `registry.myhost.com` gitlab container registry if `CI_REGISTRY` variable is set to `registry.myhost.com`.

#### Autologin for cleaning commands

Default job token `CI_JOB_TOKEN` of gitlab is not suitable to perform delete operations on [gitlab container registry](https://docs.gitlab.com/ee/user/project/container_registry.html). This token only allows read, create and update operations, but does not allow delete operations.

To work around this problem werf supports special environment variable `WERF_IMAGES_CLEANUP_PASSWORD`. This variable should contain a password for the user with enough permissions to delete images from the docker registry. It could contain *gitlab token* of regular gitlab user with such permissions. Werf will use the username `werf-cleanup` in this case.

For [cleaning commands]({{ site.baseurl }}/reference/registry/cleaning.html) werf autologin procedure use `WERF_IMAGES_CLEANUP_PASSWORD` instead of `CI_JOB_TOKEN` to access docker registry. This variable should be set up manually via [GitLab CI/CD Secret Variables](https://docs.gitlab.com/ee/ci/variables/#variables).

#### Specify username and password manually

Options `--registry-username` and `--registry-password` can be used for werf commands to perform login procedure into docker registry.

## Examples

Run werf command without login options. Autologin procedure is enabled in this case:

```bash
werf push --repo registry.myhost.com/web/backend
```

Run werf command with a manually specified user and password for docker registry:

```bash
werf push --repo registry.myhost.com/web/backend --registry-username=myuser --registry-password=mypassword
```

Run werf command with disabled autologin procedure:

```bash
WERF_IGNORE_CI_DOCKER_AUTOLOGIN=1 werf push --repo registry.myhost.com/web/backend
```

Run werf cleaning command with a special token for autologin procedure:

```bash
export WERF_IMAGES_CLEANUP_PASSWORD="A3XewXjfldf"
werf cleanup --repo ${CI_REGISTRY_IMAGE}
```

Run werf cleaning command with a manually specified user and password:

```bash
werf cleanup --repo registry.myhost.com/web/backend --registry-username=myuser --registry-password=mypassword
```

Run werf command with external docker config. Werf will not perform the login procedure in this case:

```bash
export WERF_DOCKER_CONFIG=./.docker
werf push --repo registry.myhost.com/web/backend
```
