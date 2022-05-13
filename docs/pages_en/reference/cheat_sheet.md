---
title: Cheat sheet
permalink: reference/cheat_sheet.html
toc: false
---

## Building and deploying with a single command

Build an image and deploy it to production:

```
werf converge --repo ghcr.io/group/project --env production
```

Build an image and deploy it to the default environment, but use custom image tags:

```
werf converge --repo ghcr.io/group/project --use-custom-tag "%image%-$CI_JOB_ID"
```

## Pipeline building blocks

You can create a pipeline tailored to your needs using the commands below.

Running most of the commands will first cause the missing images to be rebuilt. You can skip the rebuild by using the `--skip-build` flag. Make sure that the necessary images are built beforehand using the `werf build` command.

### Integrating with a CI system (GitLab and GitHub-based workflows are currently supported)

Set default values for werf commands and log in to the container registry based on GitLab environment variables:

```
. $(werf ci-env gitlab --as-file) 
```
### Building, tagging, and publishing images

Build images using the container registry:

```
werf build --repo ghcr.io/group/project
```

Build images and attach custom tags to them in addition to content-based tags:

```
werf build --repo ghcr.io/group/project --add-custom-tag latest --add-custom-tag 1.2.1
```

Build and store the final images in a separate registry deployed in the Kubernetes cluster:

```
werf build --repo ghcr.io/group/project --final-repo fast-in-cluster-registry.cluster/group/project
```

Build images using the container registry (or local storage if needed) and export them to another container registry:

```
werf export --repo ghcr.io/group/project --tag ghcr.io/group/otherproject/%image%:latest
```

### Running one-off tasks (unit-tests, lints, one-time jobs)

Use the following command to run tests in the previously built `frontend_image` in the Kubernetes Pod:

```
werf kube-run frontend_image --repo ghcr.io/group/project -- npm test
```

Run tests in Pod, but copy file with secret env vars into container before running the command:
```
werf kube-run frontend_image --repo ghcr.io/group/project --copy-to ".env:/app/.env" -- npm run e2e-tests
```

Run tests in Pod and get the coverage report:
```
werf kube-run frontend_image --repo ghcr.io/group/project --copy-from "/app/report:." -- go test -coverprofile report ./...
```

The command below executes the default command of the built image in the Kubernetes Pod with the CPU requests set:

```
werf kube-run frontend_image --repo ghcr.io/group/project --overrides='{"spec":{"containers":[{"name": "%container_name%", "resources":{"requests":{"cpu":"100m"}}}]}}'
```

### Running integration tests

Generally, to run some integration tests (e2e, acceptance, security, etc.) you will need a production environment (you can prepare it with converge or bundle) and a container with the appropriate command. 

Running integration tests using converge:

```
werf converge --repo ghcr.io/group/project --env integration
```

Running integration tests using converge to prepare the environment and kube-run to run a one-off task:

```
werf converge --repo ghcr.io/group/project --env integration_infra
werf kube-run --repo ghcr.io/group/project --env integration -- npm run acceptance-tests
```

### Preparing a release artifact (optional)

Use werf bundles to prepare release artifacts that can be tested or deployed later (using werf, Argo CD, or Helm), and save them to the container registry using the specified tag. 

Using a semver tag that is compatible with the Helm OCI chart:

```
werf bundle publish --repo ghcr.io/group/project --tag 1.0.0
```

Using an arbitrary symbolic tag:

```
werf bundle publish --repo ghcr.io/group/project --tag latest
```

### Deploying the application

Building and deploying the application to production:

```
werf converge --skip-build --repo ghcr.io/group/project --env production
```

Deploying the application you built in the previous step and attaching a custom tag to it:

```
werf converge --skip-build --repo ghcr.io/group/project --use-custom-tag "%image%-$CI_JOB_ID"
```

Deploying the previously published bundle with the 1.0.0 tag to production:

```
werf bundle apply --repo ghcr.io/group/project --env production --tag 1.0.0
```

## Local development

Most commands have the `--dev` flag, which is usually what you need for local development. It allows you to run werf commands without first `git add`ing them. The `--follow` flag allows you to restart the command when files in the repository change.

Rendering and showing manifests:

```
werf render --dev
```

Building an image and starting the interactive shell in a container with a failed stage in case of failure:

```
werf build --dev [--follow] --introspect-error
```

Building an image, running it in a Kubernetes Pod, and executing the command in it:

```
werf kube-run --dev [--follow] --repo ghcr.io/group/project frontend -- npm lint
```

Starting an interactive shell in a container in a Kubernetes Pod for the specified image:

```
werf kube-run --dev --repo ghcr.io/group/project -it frontend -- bash
```

Building an image and deploying it to a dev cluster (can be local):

```
werf converge --dev [--follow] --repo ghcr.io/group/project
```

Building an image and deploying it to a dev cluster; using stages from the secondary read-only registry to speed up the build:

```
werf converge --dev [--follow] --repo ghcr.io/group/project --secondary-repo ghcr.io/group/otherproject
```

Running the "docker-compose up" command with the forwarded image names:
```
werf compose up --dev [--follow]
```
