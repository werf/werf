---
title: Cheat sheet
permalink: resources/cheat_sheet.html
toc: false
---

## Building and deploying with a single command

Build images and deploy application to production:

```shell
werf converge --repo ghcr.io/group/project --env production
```

Build images and deploy application to the default environment, but use custom image tags:

```shell
werf converge --repo ghcr.io/group/project --use-custom-tag "%image%-$CI_JOB_ID"
```

## Pipeline building blocks

You can create a pipeline tailored to your needs using the commands below.

Running most commands will first check for images in the specified repo. Build instructions will run for missing images. For some scenarios, e.g. running tests in CI system, it is more convenient to use `werf build` to build images first and then strictly use the same images on other steps with flag `--require-built-images`. The step command will exit with an error in case of missing images.

### Integrating with a CI system (GitLab and GitHub-based workflows are currently supported)

Set default values for werf commands and log in to the container registry based on GitLab environment variables:

```shell
. $(werf ci-env gitlab --as-file) 
```

### Building, tagging, and publishing images

Build images using the container registry:

```shell
werf build --repo ghcr.io/group/project
```

Build images and attach custom tags to them in addition to content-based tags:

```shell
werf build --repo ghcr.io/group/project --add-custom-tag latest --add-custom-tag 1.2.1
```

Build and store the final images in a separate registry deployed in the Kubernetes cluster:

```shell
werf build --repo ghcr.io/group/project --final-repo fast-in-cluster-registry.cluster/group/project
```

Build images using the container registry (or local storage if needed) and export them to another container registry:

```shell
werf export --repo ghcr.io/group/project --tag ghcr.io/group/otherproject/%image%:latest
```

### Running one-off tasks (unit-tests, lints, one-time jobs)

Use the following command to run the tests in the previously built `frontend_image` in the Kubernetes Pod:

```shell
werf kube-run frontend_image --repo ghcr.io/group/project -- npm test
```

Run the tests in a Pod, but copy the file with the secret env variables to a container before executing the command:

```shell
werf kube-run frontend_image --repo ghcr.io/group/project --copy-to ".env:/app/.env" -- npm run e2e-tests
```

Run the tests in a Pod and get the coverage report:

```shell
werf kube-run frontend_image --repo ghcr.io/group/project --copy-from "/app/report:." -- go test -coverprofile report ./...
```

The command below executes the default command of the built image in the Kubernetes Pod with the CPU requests set:

```shell
werf kube-run frontend_image --repo ghcr.io/group/project --overrides='{"spec":{"containers":[{"name": "%container_name%", "resources":{"requests":{"cpu":"100m"}}}]}}'
```

### Running integration tests

Generally, to run some integration tests (e2e, acceptance, security, etc.) you will need a production environment (you can prepare it with converge or bundle) and a container with the appropriate command. 

Running integration tests using converge:

```shell
werf converge --repo ghcr.io/group/project --env integration
```

Running integration tests using converge to prepare the environment and kube-run to run a one-off task:

```shell
werf converge --repo ghcr.io/group/project --env integration_infra
werf kube-run --repo ghcr.io/group/project --env integration -- npm run acceptance-tests
```

### Preparing a release artifact (optional)

Use werf bundles to prepare release artifacts that can be tested or deployed later (using werf, Argo CD, or Helm), and save them to the container registry using the specified tag. 

Using a semver tag that is compatible with the Helm OCI chart:

```shell
werf bundle publish --repo ghcr.io/group/project --tag 1.0.0
```

Using an arbitrary symbolic tag:

```shell
werf bundle publish --repo ghcr.io/group/project --tag latest
```

### Deploying the application

Building and deploying the application to production:

```shell
werf converge --repo ghcr.io/group/project --env production
```

Deploying the application you built in the previous step and using a custom tag:

```shell
werf converge --require-built-images --repo ghcr.io/group/project --use-custom-tag "%image%-$CI_JOB_ID"
```

Deploying the previously published bundle with the 1.0.0 tag to production:

```shell
werf bundle apply --repo ghcr.io/group/project --env production --tag 1.0.0
```

### Cleaning up a container registry

> The procedure must run on schedule. Otherwise, the number of images and werf metadata can significantly increase the size of the registry and the time it takes to complete operations

Perform a secure cleanup procedure for outdated images and werf metadata from the container registry, taking into account the user's cleanup policies and images running in the K8s cluster:

```shell
werf cleanup --repo ghcr.io/group/project
```

## Local development

Most commands have the `--dev` flag, which is usually what you need for local development. It allows you to run werf commands without first `git add`ing them. The `--follow` flag allows you to restart the command when files in the repository change.

Rendering and showing manifests:

```shell
werf render --dev
```

Building an image and starting the interactive shell in a container with a failed stage in case of failure:

```shell
werf build --dev [--follow] --introspect-error
```

Building an image, running it in a Kubernetes Pod, and executing the command in it:

```shell
werf kube-run --dev [--follow] --repo ghcr.io/group/project frontend -- npm lint
```

Starting an interactive shell in a container in a Kubernetes Pod for the specified image:

```shell
werf kube-run --dev --repo ghcr.io/group/project -it frontend -- bash
```

Building an image and deploying it to a dev cluster (can be local):

```shell
werf converge --dev [--follow] --repo ghcr.io/group/project
```

Building an image and deploying it to a dev cluster; using stages from the secondary read-only registry to speed up the build:

```shell
werf converge --dev [--follow] --repo ghcr.io/group/project --secondary-repo ghcr.io/group/otherproject
```

Running the "docker-compose up" command with the forwarded image names:

```shell
werf compose up --dev [--follow]
```
