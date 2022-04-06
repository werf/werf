---
title: Configure CI/CD
permalink: advanced/ci_cd/werf_with_argocd/configure_ci_cd.html
---

This chapter discusses setting up the individual stages that make up the [reference CI/CD flow]({{ "advanced/ci_cd/werf_with_argocd/ci_cd_flow_overview.html#reference-cicd-flow" | true_relative_url }}):

**NOTE** The container registry must be set up beforehand and the same `CONTAINER_REGISTRY` parameter must be used in all werf commands. 

## Building and publishing images

You can build and publish the images described in `werf.yaml` using the following command:

```shell
werf build --repo CONTAINER_REGISTRY
```

## Fast commit testing

Fast tests (unit tests or linters) can be performed using the following werf command (it starts a one-shot container and runs the specified `TEST_COMMAND` command):

```shell
werf kube-run IMAGE_NAME -- TEST_COMMAND
```

The command above will run the `TEST_COMMAND` in the container in the Kubernetes cluster. Note that the `IMAGE_NAME` must be defined in the `werf.yaml` file — it can be an image built specifically for such tests or a regular application image.

## Preparing release artifacts

At this step, you need to publish the werf bundle to the container registry. The `werf bundle publish` command will automatically make sure that all the necessary images are built and published to the container registry and will publish the bundle as an OCI Helm chart with the specified SEMVER tag:

```shell
werf bundle publish --repo CONTAINER_REGISTRY --tag SEMVER
```

The use of SEMVER tags with OCI Helm charts is mandatory at this point (static tags like `main` / `latest` are not yet supported). Depending on the CI/CD model used:
* If the application is released under SEMVER Git tags, it is recommended to publish the bundle under the same SEMVER tag;
* If the application is deployed to production from the main branch (either automatically from HEAD commit or manually from some other commit), it is recommended to publish the bundle under an artificial tag like `0.0.${CI_PIPELINE_ID}`. In this case, `CI_PIPELINE_ID` is some built-in CI/CD system variable, which is a monotonically increasing number that identifies the entire pipeline of CI/CD steps for each Git commit.

## Acceptance testing

The acceptance testing step requires deploying the application to a temporary production-like environment, running the test job, and waiting for the results. This task can be accomplished in a number of alternative ways.

Typically, acceptance testing should run automatically for new commits to the main branch. In addition, failed acceptance tests usually prevent the deployment of release artifacts to production.

Thus, you need to configure the CI/CD step to run the test command on new commits to the main integration branch.
* The process can then be optimized so that tests are not run on every commit to the branch, but only on the latest commit at the moment;
* It may be necessary to implements a queue to run only one testing environment (namespace and release) at a time, or run each commit in a separate unique environment.

The following sections will describe how to do this.

### Running with werf converge

If the special test `Job` is defined as a Helm `post-install` hook in the `test` environment, run the following command in the **project's Git worktree**:

```shell
werf converge --repo CONTAINER_REGISTRY --env test
```

This command will start testing by deploying the `PROJECT_NAME-test` release in a production-like Kubernetes environment in the `PROJECT_NAME-test` namespace and run the `post-install` test Job (or Jobs). Then the converge command will automatically collect the test Job (or Jobs) output and terminate the entire process if the test fails.

Use the dismiss command to terminate the test environment:

```shell
werf dismiss --env test
```

### Running with release artifact and werf

In this case, the `werf bundle apply` command deploys the previously published release artifact. 

If you have defined a special test `Job` running in the `test` environment as a `post-install` Helm hook, run the following command:

```shell
werf bundle apply --repo CONTAINER_REGISTRY --env test --namespace TEST_NAMESPACE --release TEST_RELEASE
```

Note that this command (unlike werf converge) does not require you to specify the Git worktree of the project, or its output will be the similar to the one of [werf converge](#running-with-werf-converge).

The `TEST_NAMESPACE` and `TEST_RELEASE` variables must be specified explicitly and cleaned up after use:

```shell
werf helm uninstall TEST_RELEASE -n TEST_NAMESPACE
kubectl delete ns TEST_NAMESPACE
```

### Running with release artifact and ArgoCD

First, you need to create a special ArgoCD Application CRD:

```yaml
spec:
  destination:
    namespace: TEST_NAMESPACE
  source:
    chart: PROJECT_NAME
    repoURL: REGISTRY
    targetRevision: SEMVER
```

Execute the following command in your CI/CD to run the tests:

```
argocd app sync TEST_APPNAME --revision SEMVER
```

— where `SEMVER` is the regular version or the artificial tag like `0.0.${CI_PIPELINE_ID}` (see [preparing release artifacts](#preparing-release-artifacts)).

## Deploying to production

### Configuring ArgoCD Application

When creating the ArgoCD Application CRD, note the following fields:

```yaml
spec:
  destination:
    namespace: TEST_NAMESPACE
  source:
    chart: PROJECT_NAME
    repoURL: REGISTRY
    targetRevision: SEMVER
```

**NOTE** The bundle must be published using the `werf bundle publish --repo REGISTRY/PROJECT_NAME --tag SEMVER` format to match the parameters specified in the CRD. `REGISTRY/PROJECT_NAME` is the repository address in the container registry that contains the published werf bundle as an OCI Helm chart.

Set `targetRevision` to some initial bootstrap version of your application. We will cover updating `targetRevision` when the new release comes out in the following sections.

### Deploying by manual action

Add the command below to your CI/CD system. It will be invoked every time the user wants to manually deploy the application to production:

```shell
argocd app sync APPNAME --revision SEMVER
```

Note that the `SEMVER` revision should be generated in the same way as described in [preparing release artifacts](#preparing-release-artifacts). 

### Continuous Deployment

Continuos deployment means that the latest release artifact for, e.g., the `main` branch will be automatically deployed to the Kubernetes production cluster.

To configure continuous deployment of release artifacts, add the following annotations to the ArgoCD Application CRD:

```
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  annotations:
    argocd-image-updater.argoproj.io/chart-version: ~ 0.0
    argocd-image-updater.argoproj.io/pull-secret: pullsecret:NAMESPACE/SECRET
```

The value of `argocd-image-updater.argoproj.io/chart-version="~ 0.0"` means that the operator must automatically roll out the chart to the latest patch version in the SEMVER range of `0.0.*`.

You must also create a `NAMESPACE/SECRET` secret (it is recommended to use the `type: kubernetes.io/dockerconfigjson` secret) to access the `CONTAINER_REGISTRY` with the published [release artifact](#preparing-release-artifacts).
