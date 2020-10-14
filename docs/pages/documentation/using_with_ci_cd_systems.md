---
title: Using with CI/CD systems
permalink: documentation/using_with_ci_cd_systems.html
sidebar: documentation
---

## Intro

In this article we are covering the basics of using werf with any CI/CD system.

A more advanced article [generic CI/CD integration]({{ site.baseurl }}/documentation/advanced/ci_cd/generic_ci_cd_integration.html) exists, if you need such.

There is also a special support for GitLab CI/CD and GitHub Actions in werf and special command `werf ci-env` (this command is optional, but allows automatic configuration of all werf params described in this article in a common way). Following articles contain more details:
 - [GitLab CI/CD]({{ site.baseurl }}/documentation/advanced/ci_cd/gitlab_ci_cd.html);
 - [GitHub Actions]({{ site.baseurl }}/documentation/advanced/ci_cd/github_actions.html).

## Werf commands you will need

Werf provides following main commands as building blocks to construct your pipelines:

 - `werf converge`;
 - `werf dismiss`;
 - `werf cleanup`.

There are actually more werf commands which you may need eventually, such as:
 - `werf build` to just build needed images;
 - `werf run` to run unit tests in built images.

## Steps to embed werf into CI/CD system

Generally it is really easy to emded werf into any CI/CD system job with following steps:

 1. Setup CI/CD job access to the Kubernetes cluster and Docker Registry.
 2. Checkout target git commit of the project repo.
 3. Call `werf converge` command to deploy app, or `werf dismiss` command to terminate app, or `werf cleanup` to clean unused docker images from the Docker Registry.

### Prepare Kubernetes access

Werf need a [kube config](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/) to connect and interract with the Kubernetes cluster.

If you already have set up `kubectl` tool to work with your Kubernetes cluster, then `werf` tool will also work out of box, because werf respects the same kube config settings (`~/.kube/config` default config file and `KUBECONFIG` environment variable).

Otherwise you should setup kube config and pass it to the werf using `--kube-config` param (or `WERF_KUBE_CONFIG` environment variable), `--kube-context` param (or `WERF_KUBE_CONTEXT` environment variable) or `--kube-config-base64` param (or `WERF_KUBE_CONFIG_BASE64` environment variable) to pass base64 encoded kube config data directly into werf.

<div class="details">
<div id="details_link">
<a href="javascript:void(0)" class="details__summary">See example</a>
</div>
<div class="details__content" markdown="1">
Use custom kube config file `./my-kube-config`:

```shell
werf converge --kube-config ./my-kube-config
```

Pass base64 encoded config using environment variable:

```shell
export WERF_KUBE_CONFIG_BASE64="$(cat ./my-kube-config | base64 -w0)"
werf converge
```
</div>
</div>

### Prepare Docker Registry access

If your project use any custom images, which should be built specifically for the project (rather than using publicly available), then you need a private Docker Registry. In such case a Docker Registry instance should be available from the CI/CD runner host (to publish built images) and from within your Kubernetes cluster (to pull images).

If you already have set up `docker` tool with access for your Docker Registry from your host, then `werf` tool will also work with this Docker Registry out of the box, because werf respects the same docker config settings (`~/.docker/` default config directory or `DOCKER_CONFIG` environment variable).

<div class="details">
<div id="details_link">
<a href="javascript:void(0)" class="details__summary">Otherwise perform standard docker login procedure into Docker Registry.</a>
</div>
<div class="details__content" markdown="1">
```shell
docker login registry.mydomain.org/application -uUSER -pPASSWORD
```

If you don't want your runner host to be logged into Docker Registry all the time then perform docker login procedure in each CI/CD job. Usually it is important to create a temporal docker config for each CI/CD job (instead of using default `~/.docker/` config directory) to prevent multiple jobs from conflicting with each other when performing login from multiple jobs running at the same runner host at the same time.

```shell
# Job 1
export DOCKER_CONFIG=$(mktemp -d)
docker login registry.mydomain.org/application -uUSER -pPASSWORD
# Job 2
export DOCKER_CONFIG=$(mktemp -d)
docker login registry.mydomain.org/application -uUSER -pPASSWORD
```
</div>
</div>

### Setup destination environment for werf

Typically an application should be deployed into different [environments]({{ site.baseurl }}/documentation/advanced/ci_cd/ci_cd_workflow_basics.html#environment) (such as `production`, `staging`, `testing`, etc.).

Werf supports an optional `--env` param (or environment variable `WERF_ENV`) which specifies the name of the used environment. This environment name will affect used [Kubernetes namespace]() and [Helm release name](). It is recommended to detect environment in the CI/CD job (for example using builtin environment variables of your CI/CD system) and set werf env param correspondingly.

<div class="details">
<div id="details_link">
<a href="javascript:void(0)" class="details__summary">See example</a>
</div>
<div class="details__content" markdown="1">
Specify `WERF_ENV` environment variable:

```shell
export WERF_ENV=production
werf converge
```

Or use cli param:

```shell
werf converge --env staging
```

Or take environment from variables built into your CI/CD system:

```shell
export WERF_ENV=$CI_ENVIRONMENT_NAME
werf converge
```
</div>
</div>

### Checkout git commit and run werf

Usually CI/CD system checkout current git commit completely automatically and you don't have to make anything. But some systems may not perform this step automatically. In such case it is required to checkout target git commit of the project repo before running main werf commands (`werf converge`, `werf dismiss` or `werf cleanup`).

_NOTE: `werf converge` command behaviour is fully deterministic and transparent from the git repo standpoint. After converge is done your application is up and running in the state defined in the target git commit. Typically to rollback your application to the previous version you just need to run converge job on the corresponding previous commit (werf will use correct images for this commit)._

### Other configuration of werf for CI/CD

There is other stuff which is optional, but typically configured for werf in CI/CD.

 1. Custom annotations and labels for all deployed Kubernetes resources:
   - application name;
   - application git url;
   - application version;
   - CI/CD converge job url;
   - git commit;
   - git tag;
   - git branch;
   - etc.

    Custom annotations and labels can be set with `--add-annotation`, `--add-label` options (or `WERF_ADD_ANNOTATION_<ARBITRARY_STRING>`, `WERF_ADD_LABEL_<ARBITRARY_STRING>` environment variables).

 2. Colorized log output mode, log output width and custom logging options which can be set with `--log-color-mode=on|off`, `--log-terminal-width=N`, `--log-project-dir` options (or `WERF_COLOR_MODE=on|off`, `WERF_LOG_TERMINAL_WIDTH=N`, `WERF_LOG_PROJECT_DIR=1` environment variables).

 3. Special werf process mode call "werf process exterminator", which detects cancelled CI/CD job and terminates werf process automatically. Not all CI/CD systems need this, but some systems does not send correct termination signal to spawned processes (for example GitLab CI/CD). In such case it is bettern to enable werf `--enable-process-exterminator` option (or `WERF_ENABLE_PROCESS_EXTERMINATOR=1` environment variable).

<br>

<div class="details">
<div id="details_link">
<a href="javascript:void(0)" class="details__summary">See example</a>
</div>
<div class="details__content" markdown="1">
Setup custom annotations, logging options and werf process exterminator mode with environment variables:

```shell
export WERF_ADD_ANNOTATION_APPLICATION_NAME="project.werf.io/name=myapp"
export WERF_ADD_ANNOTATION_APPLICATION_GIT_URL="project.werf.io/git-url=https://mygit.org/myapp"
export WERF_ADD_ANNOTATION_APPLICATION_VERSION="project.werf.io/version=v1.4.14"
export WERF_ADD_ANNOTATION_CI_CD_JOB_URL="ci-cd.werf.io/job-url=https://mygit.org/myapp/jobs/256385"
export WERF_ADD_ANNOTATION_CI_CD_GIT_COMMIT="ci-cd.werf.io/git-commit=a16ce6e8b680f4dcd3023c6675efe5df3f40bfa5"
export WERF_ADD_ANNOTATION_CI_CD_GIT_TAG="ci-cd.werf.io/git-tag=v1.4.14"
export WERF_ADD_ANNOTATION_CI_CD_GIT_BRANCH="ci-cd.werf.io/git-branch=master"

export WERF_LOG_COLOR_MODE=on
export WERF_LOG_TERMINAL_WIDTH=95
export WERF_LOG_PROJECT_DIR=1
export WERF_ENABLE_PROCESS_EXTERMINATOR=1
```
</div>
</div>

## What's next?

To control output during deploy check out [this section]({{ site.baseurl }}/documentation/reference/deploy_annotations.html).

You can also check out [CI/CD workflow basics article]({{ site.baseurl }}/documentation/advanced/ci_cd/ci_cd_workflow_basics.html), which describes how to setup your CI/CD workflows in a different ways.

Find a guide which is suitable for you project in the [guides section](https://ru.werf.io/applications_guide_ru). These guides also contain detailed information about setting up concrete CI/CD systems.
