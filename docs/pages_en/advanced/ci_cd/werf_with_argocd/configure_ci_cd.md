---
title: Configure CI/CD
permalink: advanced/ci_cd/werf_with_argocd/configure_ci_cd.html
---

In this chapter we will describe how to setup following CI/CD steps according to the [reference CI/CD flow overview]({{ "advanced/ci_cd/werf_with_argocd/ci_cd_flow_overview.html#reference-cicd-flow" | true_relative_url }}):

**NOTE** It is important to setup container registry and use single `CONTAINER_REGISTRY` param in all werf commands. 

## Build and publish images

Call following command to build and publish images described in the `werf.yaml`:

```shell
werf build --repo CONTAINER_REGISTRY
```

## Fast commit testing

To run fast tests like unit testing or linting we call werf command, which runs one-shot container with the specified command:

```shell
werf kube-run IMAGE_NAME -- TEST_COMMAND
```

This command will run specified command in the container using Kubernetes cluster. `IMAGE_NAME` should be defined in the `werf.yaml` — this may be an image built specifically for such tests or regular application image.

## Prepare release artifacts

At this point we need to publish werf bundle into the container registry. `werf bundle publish` command will automatically ensure that all needed images are built and published into container registry and will publish bundle as OCI helm chart by the specified semver tag:

```shell
werf bundle publish --repo CONTAINER_REGISTRY --tag SEMVER
```

It is current requirement of helm OCI charts to use semver tags (static tags like `main` / `latest` are not allowed at the moment). Depending on your CI/CD model:
* If application is already released under semver git tags, then we propose to publish bundle under the same semver tag.
* If application deployed into production from the main branch (either automatically from HEAD commit or from some commit by user manual action), then we propose to publish bundle under fake tag like `0.0.${CI_PIPELINE_ID}`. `CI_PIPELINE_ID` is some builtin variable of your CI/CD system, which is monotonically increasing number that identifies whole pipeline of CI/CD steps for each git commit.

## Acceptance testing

During acceptance testing step we need to rollout our application into the temporal production-like environment, then run test job and wait for results. This task may be performed in multiple alternative ways.

Typically acceptance testing should run automatically for new commits merged into the main branch. Also typically failed acceptance tests should prevent release artifact from being deployed into production.

So you need to setup CI/CD step which would run test command on new commits of the main integration branch.
* This could be optimized later to run tests not on every commit of the branch, but only on latest at the moment.
* Possibly you will need a queue to run only one test acceptance environment (namespace and release) at a time, or run each commit in the separate unique release.

In the following sections we would describe how to run such test

### Run with werf converge

In the case we have defined special test `Job` as a helm `post-install` hook in the `test` environment, we need to run following command in the **project git worktree**:

```shell
werf converge --repo CONTAINER_REGISTRY --env test
```

This command will start test by deploying a production-like application into the Kubernetes namespace `PROJECT_NAME-test` release `PROJECT_NAME-test` and run `post-install` Job (or Jobs) with tests. Converge command will automatically get test Job (or Jobs) output and fail whole process in the case of tests failure.

Use dismiss command to terminate test environment:

```shell
werf dismiss --env test
```

### Run with release artifact and werf

This way we use `werf bundle apply` command to rollout previously published release artifact. 

In the case we have defined special test `Job` as a helm `post-install` hook in the `test` environment, we need to run following command:

```shell
werf bundle apply --repo CONTAINER_REGISTRY --env test --namespace TEST_NAMESPACE --release TEST_RELEASE
```

Note that this command does not need project git worktree (as werf converge does), otherwise it will give the same output and effects as [run with werf converge](#run-with-werf-converge).

`TEST_NAMESPACE` and `TEST_RELEASE` should be specified explicitly and cleaned up afterwards:

```shell
werf helm uninstall TEST_RELEASE -n TEST_NAMESPACE
kubectl delete ns TEST_NAMESPACE
```

### Run with release artifact and ArgoCD

This way we need to configure special ArgoCD Application CRD:

```yaml
spec:
  destination:
    namespace: TEST_NAMESPACE
  source:
    chart: PROJECT_NAME
    repoURL: REGISTRY
    targetRevision: SEMVER
```

Perform following command in your CI/CD to start such tests:

```
argocd app sync TEST_APPNAME --revision SEMVER
```

— where `SEMVER` could be regular version of fake tag like `0.0.${CI_PIPELINE_ID}` (see [preparing release artifacts](#prepare-release-artifacts)).

## Deploy into production

### Configure ArgoCD Application

When configuring ArgoCD Application CRD the most notable fields are:

```yaml
spec:
  destination:
    namespace: TEST_NAMESPACE
  source:
    chart: PROJECT_NAME
    repoURL: REGISTRY
    targetRevision: SEMVER
```

**NOTE** Bundle should be published using `werf bundle publish --repo REGISTRY/PROJECT_NAME --tag SEMVER` format to confirm with the above setting. `REGISTRY/PROJECT_NAME` is the address of repository in the container registry, which contains published werf bundle as helm OCI chart.

Set `targetRevision` to some initial bootstrap version of your application. In the following sections we will describe how to update `targetRevision` when a new release arrives.

### Deploy by manual action

To deploy your application into production in a manual maner, setup the following command in your CI/CD system and trigger this command on manual user action:

```shell
argocd app sync APPNAME --revision SEMVER
```

Revision `SEMVER` should be formed the same way as in the [prepare release artifact step](#prepare-release-artifacts). 

### Continuous Deployment

Continuos deployment means that latest release artifact for the `main` branch for example will be automatically deployed into production Kubernetes cluster.

To setup continuous deployment of release artifacts, configure ArgoCD Application CRD with the following annotations:

```
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  annotations:
    argocd-image-updater.argoproj.io/chart-version: ~ 0.0
    argocd-image-updater.argoproj.io/pull-secret: pullsecret:NAMESPACE/SECRET
```

The value of `argocd-image-updater.argoproj.io/chart-version="~ 0.0"` means that operator should automatically rollout chart to the latest patch version within semver range `0.0.*`.

We need to create secret `NAMESPACE/SECRET` (it is recommended to use `type: kubernetes.io/dockerconfigjson` secret) to access `CONTAINER_REGISTRY` which contains published [release artifact](#prepare-release-artifacts). 
