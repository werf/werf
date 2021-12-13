---
title: Using with CI/CD systems
permalink: using_with_ci_cd_systems.html
description: Learn the essentials of using werf in any CI/CD system
---

## Intro

In this article we are covering the basics of using werf with any CI/CD system.

Also, there is an article discussing the more advanced topic of [generic CI/CD integration]({{ "advanced/ci_cd/generic_ci_cd_integration.html" | true_relative_url }}).

werf supports any CI/CD system out-of-the-box. Furthermore, there is enhanced support for GitLab CI/CD and GitHub Actions in the form of the `werf ci-env` shortcut command that simplifies the configuration of these two systems even more (this command allows you to configure all werf parameters discussed in this article universally and automatically). You may find additional details in the following documents:

 - [GitLab CI/CD]({{ "advanced/ci_cd/gitlab_ci_cd.html" | true_relative_url }});
 - [GitHub Actions]({{ "advanced/ci_cd/github_actions.html" | true_relative_url }}).

## werf commands you will need

werf provides the following basic commands that serve as building blocks to construct your pipelines:

 - `werf converge`;
 - `werf dismiss`;
 - `werf cleanup`.

Also, there are some more specific werf commands, such as:

 - `werf build` to build the images;
 - `werf run` for unit testing of built images.

## Embedding werf into a CI/CD system

werf is designed to be easily embedded into any CI/CD system. The process is simple and intuitive - just follow the steps below:

 1. Configure a CI/CD job to access the Kubernetes cluster and the container registry.
 2. Checkout the target git commit in the project repository.
 3. Call the `werf converge` command to deploy an app, or `werf dismiss` command to terminate the app, or `werf cleanup` to clean up unused docker images from the container registry.

### Connecting to the Kubernetes cluster

werf requires [kubeconfig](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/) in order to connect to the Kubernetes cluster and interact with it.

If you have already installed and set up a `kubectl` tool to work with your Kubernetes cluster, then `werf` would also work out-of-the-box since it uses the same kubeconfig settings (the default `~/.kube/config` configuration file and the `KUBECONFIG` environment variable).

Otherwise, you should create the kubeconfig file and pass it to the werf by setting the `--kube-config` flag (or `WERF_KUBE_CONFIG` environment variable), `--kube-context` flag (or `WERF_KUBE_CONTEXT` environment variable) or `--kube-config-base64` flag (or `WERF_KUBE_CONFIG_BASE64` environment variable) to pass the base64-encoded kubeconfig data directly into werf.

<div class="details">
<a href="javascript:void(0)" class="details__summary">Example</a>
<div class="details__content" markdown="1">
Use a custom kubeconfig file `./my-kube-config`:

```shell
werf converge --kube-config ./my-kube-config
```

Pass the base64-encoded config using the environment variable:

```shell
export WERF_KUBE_CONFIG_BASE64="$(cat ./my-kube-config | base64 -w0)"
werf converge
```
</div>
</div>

### Configure a CI/CD job to access the container registry

You probably need a private container registry if your project involves any custom images (rather than publicly available) that have to be built for the project specifically. In such a case, the container registry instance should be accessible from the CI/CD runner host (to publish newly built images) and from within your Kubernetes cluster (to pull those images).

If you have configured the `docker` tool to access the private container registry from your host, then `werf` would work with this container registry right out-of-the-box, since it uses the same docker config settings (the default `~/.docker/` config directory or the `DOCKER_CONFIG` environment variable).

<div class="details">
<a href="javascript:void(0)" class="details__summary">Otherwise, perform the standard docker login procedure into your container registry.</a>
<div class="details__content" markdown="1">
```shell
docker login registry.mydomain.org/application -uUSER -pPASSWORD
```

If you don't want your runner host to be logged into the container registry all the time, perform the docker login procedure in each CI/CD job individually. It is recommended to create a temporary docker config for each CI/CD job (instead of using the default `~/.docker/` config directory) to prevent a conflict of different jobs running on the same runner host while logging in at the same time.

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

### Configure the destination environment for werf

Typically, an application is deployed into different [environments]({{ "advanced/ci_cd/ci_cd_workflow_basics.html#environment" | true_relative_url }}) (`production`, `staging`, `testing`, etc.).

werf supports the optional `--env` param (or the `WERF_ENV` environment variable) that specifies the name of the environment in use. This environment name affects the [Kubernetes namespace]({{ "/advanced/helm/releases/naming.html#kubernetes-namespace" | true_relative_url }}) and the [Helm release name]({{ "/advanced/helm/releases/naming.html#release-name" | true_relative_url }}). It is recommended to find out the name of the environment as part of the CI/CD job (for example, using built-in environment variables of your CI/CD system) and set the werf `--env` parameter accordingly.

<div class="details">
<a href="javascript:void(0)" class="details__summary">Example</a>
<div class="details__content" markdown="1">
Specify the `WERF_ENV` environment variable:

```shell
export WERF_ENV=production
werf converge
```

Or use the cli parameter:

```shell
werf converge --env staging
```

Or make use of built-in variables of your CI/CD system:

```shell
export WERF_ENV=$CI_ENVIRONMENT_NAME
werf converge
```
</div>
</div>

### Checkout the git commit and run werf

Usually, a CI/CD system checks out to the current git commit in a fully automatic manner, and you don’t have to do anything. But some systems may not perform this step automatically. If this is the case, then you have to check out the target git commit in the project repository before running main werf commands (`werf converge`, `werf dismiss`, or `werf cleanup`).

Read more about the converge process in the [introduction](/introduction.html#what-is-converge).

### Other werf CI/CD settings

There are other optional settings (they are typically configured for werf in CI/CD):

 1. Custom annotations and labels for all deployed Kubernetes resources:
   - application name;
   - application git url;
   - application version;
   - CI/CD converge job url;
   - git commit;
   - git tag;
   - git branch;
   - etc.

    Custom annotations and labels can be set via `--add-annotation`, `--add-label` options (or `WERF_ADD_ANNOTATION_<ARBITRARY_STRING>`, `WERF_ADD_LABEL_<ARBITRARY_STRING>` environment variables).

 2. Colorized log output mode, log output width and custom logging options can be set via `--log-color-mode=on|off`, `--log-terminal-width=N`, `--log-project-dir` options (or `WERF_COLOR_MODE=on|off`, `WERF_LOG_TERMINAL_WIDTH=N`, `WERF_LOG_PROJECT_DIR=1` environment variables).

 3. Special werf process mode called "werf process exterminator". It detects the canceled CI/CD job and terminates the werf process automatically. Not all CI/CD systems require this, but some systems do not send the correct termination signal to spawned processes (such as GitLab CI/CD). In this case, it is better to enable the `--enable-process-exterminator` option (or the `WERF_ENABLE_PROCESS_EXTERMINATOR=1` environment variable).

<br>

<div class="details">
<a href="javascript:void(0)" class="details__summary">Example</a>
<div class="details__content" markdown="1">
Let's set up custom annotations, logging options, and the process exterminator mode using environment variables:

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

[This section]({{ "reference/deploy_annotations.html" | true_relative_url }}) shows you how to control output during the deploy process.

You can also check out the [CI/CD workflow basics article]({{ "advanced/ci_cd/ci_cd_workflow_basics.html" | true_relative_url }}) that describes setting up your CI/CD workflows in a variety of ways.

You may find a guide suitable for you project in the [guides section](/guides.html). These guides also contain detailed information about setting up specific CI/CD systems.
