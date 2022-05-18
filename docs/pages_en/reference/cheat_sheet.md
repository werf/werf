---
title: Cheat sheet
permalink: reference/cheat_sheet.html
toc: false
---

## Build and deploy with a single command

Build and deploy to production:

```
werf converge --repo ghcr.io/group/project --env production
```

Build and deploy to default environment, but use custom image tags:

```
werf converge --repo ghcr.io/group/project --use-custom-tag "%image%-$CI_JOB_ID"
```

## Pipeline building blocks

Following commands can be composed to make a pipeline suited for your needs.

Running most commands will trigger rebuild of missing images first. You can skip the rebuild with the `--skip-build` flag. Make sure required images are built beforehand with `werf build`.

### integrate with a CI system (GitLab and GitHub Workflows supported)

Set the defaults for werf commands and perform login into a container registry based on GitLab environment variables:

```
. $(werf ci-env gitlab --as-file) 
```
### build, tag, and publish images

Build images using container registry:

```
werf build --repo ghcr.io/group/project
```

Build images with custom tags in addition to content-based tags:

```
werf build --repo ghcr.io/group/project --add-custom-tag latest --add-custom-tag 1.2.1
```

Build and store final images in a separate registry that is deployed in the Kubernetes cluster:

```
werf build --repo ghcr.io/group/project --final-repo fast-in-cluster-registry.cluster/group/project
```

Build images using container registry (or local storage if needed) and export them into another container registry:

```
werf export --repo ghcr.io/group/project --tag ghcr.io/group/otherproject/%image%:latest
```

### run one-off tasks (unit-tests, lint, one-time jobs)

Run tests in built image `frontend_image` in Kubernetes Pod:

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

Run default command of built image in Kubernetes in a Pod with CPU requests set:

```
werf kube-run frontend_image --repo ghcr.io/group/project --overrides='{"spec":{"containers":[{"name": "%container_name%", "resources":{"requests":{"cpu":"100m"}}}]}}'
```

### run integration tests

Generally, you need a production-like environment (prepared with converge or bundle) and a container with a command to run some integration tests (e2e, acceptance, security, etc.). 

Run integration tests using converge:

```
werf converge --repo ghcr.io/group/project --env integration
```

Run integration tests using converge to prepare environment and kube-run to run one-off task:

```
werf converge --repo ghcr.io/group/project --env integration_infra
werf kube-run --repo ghcr.io/group/project --env integration -- npm run acceptance-tests
```

### prepare release artifact (optional)

Use werf bundles to prepare release artifacts that could be tested or deployed later (by the werf or Argo CD, or Helm), and store them into the container registry using a specified tag. 

Use helm OCI chart compatible semver tag:

```
werf bundle publish --repo ghcr.io/group/project --tag 1.0.0
```

Use arbitrary symbolic tag:

```
werf bundle publish --repo ghcr.io/group/project --tag latest
```

### deploy application

Build and deploy the application to production:

```
werf converge --skip-build --repo ghcr.io/group/project --env production
```

Build and deploy the application using a custom tag built on a previous step :

```
werf converge --skip-build --repo ghcr.io/group/project --use-custom-tag "%image%-$CI_JOB_ID"
```

Deploy previously published bundle with tag 1.0.0 to production:

```
werf bundle apply --repo ghcr.io/group/project --env production --tag 1.0.0
```

## Local development

Most commands have the `--dev` flag, usually this is what you want for local development. It allows running werf commands without `git add`ing them first. `--follow` flag allows restarting the command when files in the repository are changed.

Render and show manifests:

```
werf render --dev
```

Build image, but open interactive shell in a container with a failed stage on failure:

```
werf build --dev [--follow] --introspect-error
```

Build and run command in a Kubernetes Pod in a specified image:

```
werf kube-run --dev [--follow] --repo ghcr.io/group/project frontend -- npm lint
```

Open an interactive shell in a container of a Kubernetes Pod for specified image:

```
werf kube-run --dev --repo ghcr.io/group/project -it frontend -- bash
```

Build and deploy in a dev cluster, which might be a local one:

```
werf converge --dev [--follow] --repo ghcr.io/group/project
```

Build and deploy in a dev cluster, use stages from another secondary read-only registry to speed up the build:

```
werf converge --dev [--follow] --repo ghcr.io/group/project --secondary-repo ghcr.io/group/otherproject
```

Run "docker-compose up" command with forwarded image names:
```
werf compose up --dev [--follow]
```
