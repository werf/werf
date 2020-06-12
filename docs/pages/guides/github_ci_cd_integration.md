---
title: GitHub CI/CD integration
sidebar: documentation
permalink: documentation/guides/github_ci_cd_integration.html
author: Sergey Lazarev <sergey.lazarev@flant.com>, Alexey Igrychev <alexey.igrychev@flant.com>
---

## Task overview

In this article, we explore the various options for configuring CI/CD using GitHub CI/CD and werf.

The workflow in the repository (a set of GitHub workflow configurations) will be based on the following jobs:

* `converge` — the job to build, publish, and deploy an application for one of the cluster tiers;
* `dismiss` — the job to delete an application (it is used for review environments only);
* `cleanup` — the job to clean up the stages storage and the Docker registry.

Jobs are based on comprehensive GitHub Actions, [werf/actions](https://github.com/werf/actions), that include all the necessary steps for setting up a specific environment and running required werf commands. Below, we will discuss examples of their use (detailed information is available in the related repository).

The specific set of tiers in the Kubernetes cluster depends on many various factors. In this article, we will examine various options for setting up environments for the following tiers:

[The production tier]({{ site.baseurl }}/documentation/reference/ci_cd_workflows_overview.html#production).
[The staging tier]({{ site.baseurl }}/documentation/reference/ci_cd_workflows_overview.html#staging).
[The review tier]({{ site.baseurl }}/documentation/reference/ci_cd_workflows_overview.html#review).

The detailed description of various jobs and different options for their implementation are available below. We will start with general terms and then move on to the particularities. You may find the [complete set of various configurations for ready-made workflows](#complete-set-of-configurations-for-ready-made-workflows) at the end of this article.

Regardless of the workflow in question, all configuration versions are subject to the following rules:

* Building and publishing are integral parts of the deployment process.
* [*Deploying/deleting review environments*](#setting-up-a-review-environment):
  * You can only *roll out* to the review environment as part of the Pull Request (PR).
  * Review environments are deleted automatically when the PR is closed.
*  The [*cleanup*](#cleaning-up-images) process runs once a day according to the master schedule.

We provide various configuration options for deploying to review, staging, and production environments. Each option for the staging and production environments is complemented by several ways to roll back the release in production.

> You can learn more about implementing the CI/CD approach using werf and constructing the custom workflow in the [introductory article]({{ site.baseurl }}/documentation/reference/ci_cd_workflows_overview.html).

## Requirements
* The Kubernetes cluster.
* The project on GitHub.
* An application you are going to build and deploy with werf.
* A good understanding of GitHub Actions basics.

> Examples in this article are based on virtual machines provided by GitHub running on Linux (`runs-on: ubuntu-latest`). At the same time, all examples are also valid for pre-installed self-hosted runners based on any OS

## Building, publishing images, and deploying an application

First of all, you need to define a template — the general part of the deployment process suitable for any tier. It will allow you to focus on the rules of the deployment process and the workflows suggested. 

{% include /guides/github_ci_cd_integration/converge_base.md %}

> This job can be split into two independent jobs. However, in our case (when building and publishing are performed jointly with deploying as opposed to invoking for every commit), this is redundant and would worsen the readability of the configuration as well as execution time.
>
{% include /guides/github_ci_cd_integration/build_and_publish_note.md %}

First of all, you have to perform a `Checkout code` step — add the source code of an application. It is the initial step of a job. When using the werf builder (as you know, the incremental building is its notable feature), it is not enough to have a so-called `shallow clone` with a single commit that the action `actions/checkout@v2` creates when used with no parameters specified. 
werf generates stages on the basis of the git history. So, if there is no history, then each build would run without previously built images. Therefore, it is essential to use the `fetch-depth: 0` parameter to access the entire history when building, publishing (`werf build-and-publish`), deploying (`werf deploy`), and running (`werf run`). In other words, for all commands that stages use.

{% raw %}
```yaml
- name: Checkout code
  uses: actions/checkout@v2
  with:
    fetch-depth: 0
```
{% endraw %}

The action `werf/actions/converge` combines all the necessary steps, sets up an environment, and invokes the appropriate command.

{% raw %}
```yaml
- name: Converge
  uses: werf/actions/converge@master
  with:
    env: ANY_ENV_NAME
    kube-config-base64-data: ${{ secrets.KUBE_CONFIG_BASE64_DATA }}
  env:
    WERF_SET_ENV_URL: "global.env_url=ANY_ENV_URL"
```
{% endraw %}

The `kubectl` interface is already pre-installed on GitHub virtual machines, so the user only needs to:

* decide on the [kubeconfig](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/) configuration;
* create a secret `KUBE_CONFIG_BASE64_DATA` variable containing the contents of the kubeconfig file (`cat ~/.kube/config | base64`). To do this, go to the Settings/Secrets tab of the project on GitHub.
* forward the `secrets.KUBE_CONFIG_BASE64_DATA` secret variable to the target werf action:

{% raw %}
```yaml
- name: Converge
  uses: werf/actions/converge@master
  with:
    kube-config-base64-data: ${{ secrets.KUBE_CONFIG_BASE64_DATA }}
```
{% endraw %}

Job configuring is quite simple, so we prefer to focus on what it lacks — an explicit authorization in the Docker registry, and calling the `docker login`.

In the simplest case, if an [integrated Github Packages-like Docker registry](https://help.github.com/en/packages/using-github-packages-with-your-projects-ecosystem/configuring-docker-for-use-with-github-packages) is used, then the authorization is performed automatically when the `werf ci-env` command is invoked. This command is run with several required arguments such as GitHub environment variables, the [`GITHUB_TOKEN` secret](https://help.github.com/en/actions/configuring-and-managing-workflows/authenticating-with-the-github_token#about-the-github_token-secret) (you have to explicitly declare it), and the name of the user (`GITHUB_ACTOR`) who started the workflow process.

If there is a need to perform authorization using custom credentials or in an external Docker registry, then you have to use a ready-made action tailored to your implementation (or just run `docker login` before using werf).

Now, let us explore other parameters used in this step:

{% raw %}
```yaml
- name: Converge
  uses: werf/actions/converge@master
  with:
    env: ANY_ENV_NAME
  env:
    WERF_SET_ENV_URL: "global.env_url=ANY_ENV_URL"
```
{% endraw %}

You must define an environment for each tier. In our case, it is defined by:

* the name (`ANY_ENV_NAME`), and
* the URL (`ANY_ENV_URL`).

In order to configure the application for using in different tiers, you can use Go templates and the `.Values.global.env` variable. This is analogous to setting the `env` parameter of the action.

> The address of an environment is optional. In this article, it is used solely for illustrative purposes and to demonstrate the usage of werf when working with actions (all werf options can be set via environment variables)

You can also use the environment address — the URL for accessing the application deployed to the tier — in helm templates (for example, for configuring Ingress resources). It is passed via the `global.env_url` parameter. You can find out the URL via the `WERF_SET_ENV_URL` environment variable (it is analogous to running werf with the `--set` (`WERF_SET_<ANY_NAME>`) argument).

> To encrypt variable values using werf, you need to add the `encryption key` to the `WERF_SECRET_KEY` variable in the project's Settings/Secrets and the secret to the `env` section

Below, we will explore some popular strategies and practices that may serve as a basis for building your processes in GitHub Actions.

### Setting up a review environment

As it was said before, the review environment is a dynamic tier. Therefore, this environment should also have to be able to clean up resources in addition to deploying them.

Let us consider some basic GitHub workflow files that would serve as a basis for all the suggested options.

First, let's analyze the `.github\workflows\review_deployment.yml` file.

{% include /guides/github_ci_cd_integration/review_base.md %}

It lacks the triggering condition since it depends on the chosen approach to setting up environments.

This job is identical to the basic configuration except for the `Define environment url` step. At this step, a unique URL is generated. It will be used for accessing our application after the deployment (provided there is an appropriate helm templates structure).

{% raw %}
```yaml
- name: Define environment url
  run: |
    pr_id=${{ github.event.number }}
    github_repository_id=$(echo ${GITHUB_REPOSITORY} | sed -r s/[^a-zA-Z0-9]+/-/g | sed -r s/^-+\|-+$//g | tr A-Z a-z)
    echo ::set-env name=WERF_SET_ENV_URL::global.env_url=http://${github_repository_id}-${pr_id}.kube.DOMAIN
```
{% endraw %}

Now, let's analyze `.github\workflows\review_deployment_dismiss.yml`.

{% include /guides/github_ci_cd_integration/review_dismiss_base.md %}

This GitHub workflow will be run when closing the PR.

```yaml
on:
  pull_request:
    types: [closed]
```

The review release is deleted at the `Dismiss` step: werf deletes the helm release and the namespace in Kubernetes with all its contents ([werf dismiss]({{ site.baseurl }}/documentation/cli/main/dismiss.html)).

Now, let us explore the main strategies to deploy the review environment.

> We do not limit you to the options offered, but quite the opposite: we recommend combining them and creating a workflow configuration that suits your team's needs

#### 1. Manually

> This option implements the approach described in the [Deploy to review using a pull request at the click of a button]({{ site.baseurl }}/documentation/reference/ci_cd_workflows_overview.html#deploy-to-review-using-a-pull-request-at-the-click-of-a-button) section

In this approach, the user deploys and deletes the environment by assigning the appropriate label (`review_start` or `review_stop`) in the PR.

It is the most simplistic approach that can be useful when deployments are rare, and the review environment is not used during development. In essence, this approach is used for testing before accepting a PR.

{% include /guides/github_ci_cd_integration/review_1.md %}

In this case, both GitHub workflows expect the label to be assigned to the PR.

```yaml
on:
  pull_request:
    types: [labeled]
```
    
If the `review_start` or `review_stop` label is assigned, then the jobs of the corresponding workflow are run. Otherwise, if any other label is assigned, the workflow is started, but no jobs are run, and the workflow is marked as `skipped`. Using filtering by status, you can track activity in the review environment.

The `Label taking off` step removes the label that initiates the workflow. It serves as an indicator of processing a user's request to deploy and stop the review environment (and as a nice bonus, it allows us to track the history of changes and deployments via the PR's log).

{% raw %}
```yaml
labels:
  name: Label taking off 
  runs-on: ubuntu-latest
  if: github.event.label.name == 'review_stop'
  steps:

    - name: Take off label
      uses: actions/github-script@v1
      with:
        script: "github.issues.removeLabel({...context.issue, name: github.event.label.name})"
```
{% endraw %}

#### 2. Automatically using a branch name

> This option implements the approach described in the [Automatically deploy to review from a branch using a pattern]({{ site.baseurl }}/documentation/reference/ci_cd_workflows_overview.html#automatically-deploy-to-review-from-a-branch-using-a-pattern) section.

In the configuration below, the code is automatically released with every commit in the PR (if the name of the git branch contains the `review` prefix).

{% include /guides/github_ci_cd_integration/review_2.md %}

The deployment is triggered when there is a new commit to the branch, or a PR is opened (re-opened), which corresponds to the default set of types for the `pull_request` event:

```yaml
// equal conditions

on:
  pull_request:
    types:
      - opened
      - reopened
      - synchronize

on:
  pull_request:
```

#### 3. Semi-automatic mode using a label (recommended)

> This option implements the approach described in the [Automatically deploy to review using a pull request; manual triggering]({{ site.baseurl }}/documentation/reference/ci_cd_workflows_overview.html#automatically-deploy-to-review-using-a-pull-request-manual-triggering)

Semi-automatic mode with a label is a comprehensive solution that combines the previous two options.

By assigning a specific label (`review` in the example below), the user activates the automatic deployment to review environments for each commit. When a label is removed, the process of deleting the review release is triggered automatically.

{% include /guides/github_ci_cd_integration/review_3.md %}

The deployment process is triggered by a commit to the branch or by assigning/removing a label in the PR. All these actions correspond to the following set of pull_request events:

```yaml
pull_request:
  types:
    - labeled
    - unlabeled
    - synchronize
```

### Various scenarios for composing staging and production environments

The scenarios described below are the most effective combinations of rules for deploying to staging and production environments.

In our case, these environments are the most important ones. Thus, the names of the scenarios correspond to the names of ready-made workflows presented at the end of the article.

#### 1. Fast and Furious (recommended)

> This option implements the approaches described in the [Automatically deploy to production from master]({{ site.baseurl }}/documentation/reference/ci_cd_workflows_overview.html#automatically-deploy-to-production-from-master) and [Deploy to production-like using a pull request at the click of a button]({{ site.baseurl }}/documentation/reference/ci_cd_workflows_overview.html#deploy-to-production-like-using-a-pull-request-at-the-click-of-a-button) sections 

The code is automatically deployed to **production** in response to any changes in master. At the same time, you can deploy an application to **staging** by clicking the button in the PR.

{% include /guides/github_ci_cd_integration/production_staging_1.md %}

Options for rolling back changes in production:

- by [revert](https://git-scm.com/docs/git-revert)-ing changes in master (**recommended**);
- by deploying a stable PR.

#### 2. Push the button (*)

> This option implements the approaches described in the [Deploy to production from master at the click of a button]({{ site.baseurl }}/documentation/reference/ci_cd_workflows_overview.html#deploy-to-production-from-master-at-the-click-of-a-button) and [Automatically deploy to staging from master]({{ site.baseurl }}/documentation/reference/ci_cd_workflows_overview.html#automatically-deploy-to-staging-from-master) sections

{% include /guides/github_ci_cd_integration/not_recommended_approach_en.md %}

Deploying to **production** is triggered by clicking the button associated with the commit in master, and rolling out to **staging** is performed automatically in response to changes in master.

{% include /guides/github_ci_cd_integration/production_staging_2.md %}

We have the following condition for releasing to production:

```yaml
on:
  repository_dispatch:
    types: [production_deployment]
```

You can trigger this workflow via the following request:

```shell
curl \
  --location --request POST 'https://api.github.com/repos/<company>/<project>/dispatches' \
  --header 'Content-Type: application/json' \
  --header 'Accept: application/vnd.github.everest-preview+json' \
  --header "Authorization: token $GITHUB_TOKEN" \
  --data-raw '{
    "event_type": "production_deployment",
    "client_payload": {}
  }'
```
To use this approach, you can add the script to the project repository, use the postman platform, a browser plugin, etc.

Options for rolling back changes in production: by rolling out a stable PR and triggering the deployment via the repository_dispatch webhook.

#### 3. Tag everything (*)

> This option implements the approaches described in the [Automatically deploy to production using a tag]({{ site.baseurl }}/documentation/reference/ci_cd_workflows_overview.html#automatically-deploy-to-production-using-a-tag) and [Deploy to staging from master at the click of a button]({{ site.baseurl }}/documentation/reference/ci_cd_workflows_overview.html#deploy-to-staging-from-master-at-the-click-of-a-button) sections

{% include /guides/github_ci_cd_integration/not_recommended_approach_en.md %}

The rollout to **production** is triggered when the tag is assigned; deploying to **staging** is performed manually on master.

{% include /guides/github_ci_cd_integration/production_staging_3.md %}

```yaml
on:
  repository_dispatch:
    types: [staging_deployment]
```
You can trigger this workflow using the following request:

```shell
curl \
  --location --request POST 'https://api.github.com/repos/<company>/<project>/dispatches' \
  --header 'Content-Type: application/json' \
  --header 'Accept: application/vnd.github.everest-preview+json' \
  --header "Authorization: token $GITHUB_TOKEN" \
  --data-raw '{
    "event_type": "staging_deployment",
    "client_payload": {}
  }'
```

To use this approach, you can add the script to the project repository, use the postman platform, a browser plugin, etc.

Options for rolling back changes in production: by assigning a new tag to the old commit (not recommended).

#### 4. Branch, branch, branch!

> This option implements the approaches described in the [Automatically deploy to production from a branch]({{ site.baseurl }}/documentation/reference/ci_cd_workflows_overview.html#automatically-deploy-to-production-from-a-branch) and [Automatically deploy to production-like from a branch]({{ site.baseurl }}/documentation/reference/ci_cd_workflows_overview.html#automatically-deploy-to-production-like-from-a-branch) sections

The code is deployed to **production** automatically; rolling out to **staging** is performed in response to changes in the master branch.

{% include /guides/github_ci_cd_integration/production_staging_4.md %}

Options for rolling back changes in production:

- by [revert](https://git-scm.com/docs/git-revert)-ing changes in the production branch;
- by [revert](https://git-scm.com/docs/git-revert)-ing changes to master and performing a fast-forward merge to the production branch;
- by deleting a commit from the production branch and then performing a push-force.

## Cleaning up Images

{% include /guides/github_ci_cd_integration/cleanup_base.md %}

First of all, you need to perform a code checkout - add the source code of an application. It is the initial step of a job.

Most cleaning policies in werf are based on git primitives (commit, branch, and tag), so by using the action `actions/checkout@v2` without additional parameters, you risk accidentally deleting images. We recommend you to use the following steps for cleanup to go smoothly:

```yaml
- name: Checkout code
  uses: actions/checkout@v2
  
- name: Fetch all history for all tags and branches
  run: git fetch --prune --unshallow
```
  
werf has an efficient built-in cleanup mechanism to avoid overflowing the Docker registry and the disk space on the building node with outdated and unused images. You can learn more about the werf's cleanup functionality [here]({{ site.baseurl }}/documentation/reference/cleaning_process.html).

## Complete set of configurations for ready-made workflows

<div class="tabs" style="display: grid">
  <a href="javascript:void(0)" class="tabs__btn active" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'complete_github_ci_1')">№1 Fast and Furious (recommended)</a>
  <a href="javascript:void(0)" class="tabs__btn" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'complete_github_ci_2')">№2 Push the Button (*)</a>
  <a href="javascript:void(0)" class="tabs__btn" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'complete_github_ci_3')">№3 Tag everything (*)</a>
  <a href="javascript:void(0)" class="tabs__btn" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'complete_github_ci_4')">№4 Branch, branch, branch!</a>
</div>

<div id="complete_github_ci_1" class="tabs__content no_toc_section active" markdown="1">

Workflow details
{:.no_toc}

> You can learn more about this workflow in the [article]({{ site.baseurl }}/documentation/reference/ci_cd_workflows_overview.html#1-fast-and-furious)

* Deploying to the review tier via the strategy [No. 3 Semi-automatic mode using a label (recommended)](#3-semi-automatic-mode-using-a-label-recommended).
* Deploying to staging and production tiers via the strategy [No. 1 Fast and Furious (recommended)](#1-fast-and-furious-recommended).
* [Cleaning up stages](#cleaning-up-images) runs once a day according to the schedule.

### Related configuration files
{:.no_toc}

{% include /guides/github_ci_cd_integration/workflow_1.md %}

</div>

<div id="complete_github_ci_2" class="tabs__content no_toc_section" markdown="1">

Workflow details
{:.no_toc}

> You can learn more about this workflow in the [article]({{ site.baseurl }}/documentation/reference/ci_cd_workflows_overview.html#2-push-the-button)

{% include /guides/github_ci_cd_integration/not_recommended_approach_en.md %}

* Deploying to the review tier using the strategy [No. 1 Manually](#1-manually).
* Deploying to staging and production tiers is carried out according to the strategy [No. 2 Push the Button](#2-push-the-button-).
* [Cleaning up stages](#cleaning-up-images) runs once a day according to the schedule.

Related configuration files
{:.no_toc}

{% include /guides/github_ci_cd_integration/workflow_2.md %}

</div>

<div id="complete_github_ci_3" class="tabs__content no_toc_section" markdown="1">

### Workflow components
{:.no_toc}

> You can learn more about this workflow in the related [article]({{ site.baseurl }}/documentation/reference/ci_cd_workflows_overview.html#3-tag-everything)

{% include /guides/github_ci_cd_integration/not_recommended_approach_en.md %}

* Deploying to the review tier using the strategy [No. 1 Manually](#1-manually).
* Deploying to staging and production tiers is carried out according to the strategy [No. 3 Tag everything](#3-tag-everything-).
* [Cleaning up stages](#cleaning-up-images) runs once a day according to the schedule.

### Related configuration files
{:.no_toc}

{% include /guides/github_ci_cd_integration/workflow_3.md %}

</div>

<div id="complete_github_ci_4" class="tabs__content no_toc_section" markdown="1">

### Workflow details
{:.no_toc}

> You can learn more about this workflow in the related [article]({{ site.baseurl }}/documentation/reference/ci_cd_workflows_overview.html#4-branch-branch-branch)

* Deploying to the review tier using the strategy [No. 2 Automatically using a branch name](#2-automatically-using-a-branch-name.
* Deploying to staging and production tiers is carried out according to the strategy [No. 4 Branch, branch, branch!](#4-branch-branch-branch).
* [Cleaning up stages](#cleaning-up-images) runs once a day according to the schedule.

### Related configuration files
{:.no_toc}

{% include /guides/github_ci_cd_integration/workflow_4.md %}

</div>
