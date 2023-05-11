---
title: Deployment scenarios
permalink: usage/deploy/deployment_scenarios.html
---

## Normal deployment

A normal deployment is carried out by the `werf converge` command. It builds images and deploys the application but requires the application's Git repository to run. Example:

```shell
werf converge --repo example.org/mycompany/myapp
```

You can separate the build and deployment steps as follows:

```shell
werf build --repo example.org/mycompany/myapp
```

```shell
werf converge --require-built-images --repo example.org/mycompany/myapp
```

## Deploying using custom image tags

By default, built images are tagged based on their contents. The tag becomes available in Values and allows those images to be used in templates during deployment. But if you want to use a different tag for the images, you can use the `--use-custom-tag` parameter, for example:

```shell
werf converge --use-custom-tag '%image%-v1.0.0' --repo example.org/mycompany/myapp
```

Running the command above will result in the images being assembled, tagged with `<name image>-v1.0.0`, and published. The tags of those images will then become available in Values. The final Kubernetes manifests will then be generated and applied based on those tags.

The tag name set by the `--use-custom-tag` parameter supports the `%image%`, `%image_slug%`, and `%image_safe_slug%` patterns to substitute the image name and `%image_content_based_tag%` to substitute the original content-based tag.

> Note that when you set a custom tag, an image with a content-based tag is published as well. Later on, when `werf cleanup` is invoked, the image with the content-based tag and the images with arbitrary tags are deleted together.

You can separate the assembly and deployment steps like this:

```shell
werf build --add-custom-tag '%image%-v1.0.0' --repo example.org/mycompany/myapp
```

```shell
werf converge --require-built-images --use-custom-tag '%image%-v1.0.0' --repo example.org/mycompany/myapp
```

## Deploying without access to the application's Git repository

To deploy the application with no access to the application's Git repository, follow these three steps:

1. Build images and publish them to the container registry.

2. Add the passed parameters and publish the main chart to the OCI repository. The chart contains references to the images published in the first step.

3. Apply the published bundle to the cluster.

The first two steps are carried out by the `werf bundle publish` command from the application's Git repository, for example:

```shell
werf bundle publish --tag latest --repo example.org/mycompany/myapp
```

The third step is carried out by the `werf bundle apply` command, but you don't have to be in the application's Git repository; for example:

```shell
werf bundle apply --tag latest --release myapp --namespace myapp-production --repo example.org/mycompany/myapp
```

You will end up with the same result as with `werf converge`.

You can separate the first and second steps as follows:

```shell
werf build --repo example.org/mycompany/myapp
```

```
werf bundle publish --require-built-images --tag latest --repo example.org/mycompany/myapp
```

## Deploying without access to an application's Git repository and container registry

Follow these five steps to deploy an application with no access to the application's Git repository and the container registry:

1. Build images and publish them to the container registry of the application.

2. Add the passed parameters and publish the main chart to the OCI repository. The chart contains pointers to the images published in the first step.

3. Export the bundle and its related images to a local archive.

4. Import the archived bundle and its images into the container registry accessible from the Kubernetes cluster used for deployment.

5. Apply the bundle published in the new container registry to the cluster.

The first two steps are handled by the `werf bundle publish` command running in the application's Git repository, for example:

```shell
werf bundle publish --tag latest --repo example.org/mycompany/myapp
```

The third step is handled by the `werf bundle copy` command (no need to be in the application's Git repository), for example:

```shell
werf bundle copy --from example.org/mycompany/myapp:latest --to archive:myapp-latest.tar.gz
```

The local `myapp-latest.tar.gz` archive can now easily be pushed to the container registry used for deployment to the Kubernetes cluster, and again the `werf bundle copy` command comes into play, for example:

```shell
werf bundle copy --from archive:myapp-latest.tar.gz --to registry.internal/mycompany/myapp:latest
```

As a result, the chart and its associated images will be published to the new container registry accessible from the Kubernetes cluster. All that remains is to deploy the published bundle to the cluster using the `werf bundle apply` command like this:

```shell
werf bundle apply --tag latest --release myapp --namespace myapp-production --repo registry.internal/mycompany/myapp
```

This step no longer requires access to either the application's Git repository or its original container registry. The end result of deploying the bundle will be the same as when using `werf converge'.

You can separate the first and second step, if necessary:

```shell
werf build --repo example.org/mycompany/myapp
```

```
werf bundle publish --require-built-images --tag latest --repo example.org/mycompany/myapp
```

## Deploying with a third-party tool

To apply the final application manifests with a tool other than werf (kubectl, Helm, etc.) follow the steps below:

1. Build images and publish them to the container registry.

2. Render the final manifests.

3. Deploy those manifests to a cluster using a third-party tool.

The first two steps are carried out by the `werf render` command from the application's Git repository:

```shell
werf render --output manifests.yaml --repo example.org/mycompany/myapp
```

You can now pass the rendered manifests to a third-party tool for deployment:

```shell
kubectl apply -f manifests.yaml
```

> Note that some special features of werf, like the ability to reorder resource deployments based on their weight (using the `werf.io/weight` annotation), most likely won't work when the manifests are applied by a third-party tool.

You can separate the first and second steps as follows:

```shell
werf build --repo example.org/mycompany/myapp
```

```
werf render --require-built-images --output manifests.yaml --repo example.org/mycompany/myapp
```

## Deploying with a third-party tool without access to the application's Git repository

To deploy the application using some third-party tool (kubectl, Helm, etc.), and there's no access to the application's Git repository, follow these three steps:

1. Build images and publish them to the container registry.

2. Add the passed parameters and publish the main chart to the OCI repository. The chart contains references to the images published in the first step.

3. Render the final manifests using the bundle.

4. Deploy those final manifests to a cluster using a third-party tool.

The first two steps are carried out by the `werf bundle publish` command from the application's Git repository:

```shell
werf bundle publish --tag latest --repo example.org/mycompany/myapp
```

The third step is carried out by the `werf bundle render` command, but this time, you don't have to be in the application's Git repository; for example:

```shell
werf bundle render --output manifests.yaml --tag latest --release myapp --namespace myapp-production --repo example.org/mycompany/myapp
```

You can now pass the rendered manifests to a third-party tool for deployment, e.g.:

```shell
kubectl apply -f manifests.yaml
```

> Note that some special features of werf, like the ability to reorder resource deployments based on their weight (using the `werf.io/weight` annotation), most likely won't work when the manifests are applied by a third-party tool.

You can separate the first and second steps as follows:

```shell
werf build --repo example.org/mycompany/myapp
```

```
werf bundle publish --require-built-images --tag latest --repo example.org/mycompany/myapp
```

## Saving a deployment report

The `werf converge` and `werf bundle apply` commands come with the `-save-deploy-report` parameter. You can use it to save a report about the deployment to a file. The report contains the release name, Namespace, deployment status, and some other data. Here is a usage example:

```shell
werf converge --save-deploy-report
```

Running the command above will create a `.werf-deploy-report.json` file containing information about the latest release once the deployment is complete.

The custom path to the deployment report can be set with the `--deploy-report-path` parameter.

## Deleting a deployed application

You can delete a deployed application using the `werf dismiss` command run from the application's Git repository, for example:

```shell
werf dismiss --env staging
```

You can explicitly specify the release name and Namespace if there is no access to the application's Git repository:

```shell
werf dismiss --release myapp-staging --namespace myapp-staging
```

... or you can use a previous deployment report which contains the release name and Namespace. You can enable saving this report by using the `-save-deploy-report` flag of the `werf converge` or `werf bundle apply` commands.

```shell
werf converge --save-deploy-report
cp .werf-deploy-report.json /anywhere
cd /anywhere
werf dismiss --use-deploy-report
```

The custom path to the deployment report can be set with the `--deploy-report-path` parameter.
