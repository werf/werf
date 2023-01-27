---
title: Deployment scenarios
permalink: usage/deploy/deployment_scenarios.html
---

## Normal deployment

A normal deployment is carried out by the `werf converge` command. It builds images and deploys the application but requires the application's Git repository to run. Example:

```shell
werf converge --repo <repository>
```

You can separate the build and deployment steps as follows:

```shell
werf build --repo <repository>
```

```shell
werf converge --skip-build --repo <repository>
```

## Deploying without access to the application's Git repository

To deploy the application with no access to the application's Git repository, follow these three steps:

1. Build images and publish them to the container registry.

2. Publish the main chart and the parameters passed to it as a *bundle* in the OCI repository. The bundle contains references to the images published in the first step.

3. Apply the published bundle to the cluster.

The first two steps are carried out by the `werf bundle publish` command from the application's Git repository, for example:

```shell
werf bundle publish --tag latest --repo <repository>
```

The third step is carried out by the `werf bundle apply` command, but you don't have to be in the application's Git repository; for example:

```shell
werf bundle apply --tag latest --release myapp --namespace myapp-production --repo <репозиторий>
```

You will end up with the same result as with `werf converge`.

You can separate the first and second steps as follows:

```shell
werf build --repo <repository>
```

```
werf bundle publish --skip-build --tag latest --repo <repository>
```

## Deploying with a third-party tool

To apply the final application manifests with a tool other than werf (kubectl, Helm, etc.) follow the steps below:

1. Build images and publish them to the container registry.

2. Render the final manifests.

3. Deploy those manifests to a cluster using a third-party tool.

The first two steps are carried out by the `werf render` command from the application's Git repository:

```shell
werf render --output manifests.yaml --repo <repository>
```

You can now pass the rendered manifests to a third-party tool for deployment:

```shell
kubectl apply -f manifests.yaml
```

> Note that some special features of werf, like the ability to reorder resource deployments based on their weight (using the `werf.io/weight` annotation), most likely won't work when the manifests are applied by a third-party tool.

You can separate the first and second steps as follows:

```shell
werf build --repo <repository>
```

```
werf render --skip-build --output manifests.yaml --repo <repository>
```

## Deploying with a third-party tool without access to the application's Git repository

To deploy the application using some third-party tool (kubectl, Helm, etc.), and there's no access to the application's Git repository, follow these three steps:

1. Build images and publish them to the container registry.

2. Publish the main chart and the parameters passed to it as a *bundle* in the OCI repository. The bundle contains references to the images published in the first step.

3. Render the final manifests using the bundle.

4. Deploy those final manifests to a cluster using a third-party tool.

The first two steps are carried out by the `werf bundle publish` command from the application's Git repository:

```shell
werf bundle publish --tag latest --repo <repository>
```

The third step is carried out by the `werf bundle render` command, but this time, you don't have to be in the application's Git repository; for example:

```shell
werf bundle render --output manifests.yaml --tag latest --release myapp --namespace myapp-production --repo <repository>
```

You can now pass the rendered manifests to a third-party tool for deployment, e.g.:

```shell
kubectl apply -f manifests.yaml
```

> Note that some special features of werf, like the ability to reorder resource deployments based on their weight (using the `werf.io/weight` annotation), most likely won't work when the manifests are applied by a third-party tool.

You can separate the first and second steps as follows:

```shell
werf build --repo <repository>
```

```
werf bundle publish --skip-build --tag latest --repo <repository>
```

## Deleting a deployed application

You can delete a deployed application using the `werf dismiss` command run from the application's Git repository, for example:

```shell
werf dismiss --env staging
```
