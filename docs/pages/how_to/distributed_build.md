---
title: Distributed build with GitLab
sidebar: how_to
permalink: how_to/distributed_build.html
author: Artem Kladov <artem.kladov@flant.com>
---

## Task Overview

Werf can store [_stages cache_](https://flant.github.io/werf/reference/build/stages.html#what-is-a-stage) in a Docker registry. It gives an opportunity to use a distributed build in the following cases:
* **Dynamic resources.** You can setup a dynamic allocation of resources for building by using cloud providers. These resources can be temporary, so in build time a stages cache may be absent.
* **More than one build nodes.** If you have more than one nodes (hosts) for building images, you need to synchronize stages cache before building an image.

Distributed build assumes the following steps:
* pull existing stages cache from the Docker registry;
* build images;
* push built images and stages cache to the Docker registry.

In this tutorial, we setup the CI/CD process using GitLab CI and two persistent build nodes with a distributed build.

> The same approach described in this tutorial is also appliable when there are arbitrary number of dynamically allocated build nodes, instead of fixed number of persistent build nodes.

## Requirements

* Kubernetes cluster
* Server with GitLab above 10.x version (or account on [SaaS GitLab](https://gitlab.com/));
* Docker registry (GitLab embedded or somewhere else);
* A GitLab project with an application, you can successfully build with werf (e.g., PHP [Symfony application](https://github.com/symfony/demo)).
* Two hosts for building (**build nodes**) with:
  * GitLab runner installed and activated for the project with the `build` tag;
  * werf installed (same versions on both nodes);
* Host for deploying (**deploy node**) with:
    * GitLab runner installed and activated for the project with the `deploy` tag;
    * werf installed;
    * helm installed and initialized;
    * kubectl CLI tool configured to communicate with the Kubernetes cluster;

## Infrastructure

![]({{ site.baseurl }}/images/howto_distributed_build1.png)

> We don't recommend to run werf in docker for now as it can give an unexpected result.

To set up the CI/CD process, you need to describe build, deploy and cleanup stages. For all of these stages, a GitLab runner with shell executor is needed to run werf.

Build nodes need access to the application git repository and to the docker registry, while the deployment node additionally needs access to the kubernetes cluster.

## Pipeline

You have working build and deployment runners in your project, but there are no `.gitlab-ci.yml` file for GitLab pipeline. Let's compose it step by step (the complete `.gitlab-ci.yml` file will be given below).

When GitLab starts a job, it sets a list of [environments](https://docs.gitlab.com/ee/ci/variables/README.html), and we will use some of them.

### Base setup

Setup `APP_NAMESPACE` variable to use as the name of kubernetes namespace for deploying application.

```yaml
variables:
  APP_NAMESPACE: ${CI_PROJECT_NAME}-${CI_ENVIRONMENT_SLUG}
```

Define stages for building, deploying and two stages for cleaning app.

```yaml
stages:
- build
- deploy
- cleanup_registry
- cleanup_builder
```

### Build stage

```yaml
Build:
  stage: build
  script:
  ## The line below uses for debugging
  - werf --version; pwd; set -x
  ## Pull stages caсhe
  - werf dimg stages pull ${CI_REGISTRY_IMAGE}
  ## build dimg's than push it with the stages cache
  - werf dimg bp --with-stages ${CI_REGISTRY_IMAGE} --tag-ci
  ## Specify to use runner with the build tag.
  tags:
  - build
  ## Cleanup will use schedules, and it is not necessary to rebuild images on running cleanup jobs.
  ## Therefore we need to specify
  except:
  - schedules
```

It is better to use bp command instead of separate steps with the build and push commands.

Pay attention to the two important moments:
* There is a `stages pull` step, before build, on which werf download latest stages cache.
* There is a build and push stage with the [\-\-with-stages]({{ site.baseurl }}/reference/registry/push.html#werf-bp) option.

Two main steps allow distributing cache work.

There are two runners with the `build` tag in your project, so build jobs runs on a random runner from both.

> It is essential to use `--tag-ci`, `--tag-branch` or `--tag-commit` options here (in bp or push commands) otherwise cleanup won't work.

### Deploy stage

Our `deploy` stage is simple, and consists of the following environments:
* Test environment.
* Production environment.

You can read about more complicated deploy [here]({{ site.baseurl }}/how_to/gitlab_ci_cd_integration.html).

First of all, we define a template for the deploy jobs — this decreases the size of the `.gitlab-ci.yml` and makes it more readable. We will use this template in every deploy stage further.

Add the following lines to `.gitlab-ci.yml` file:

```yaml
.base_deploy: &base_deploy
  stage: deploy
  script:
  ## create k8s namespace we will use if it doesn't exist.
  - kubectl get ns ${APP_NAMESPACE} || kubectl create namespace ${APP_NAMESPACE}
  ## If your application use private registry, you have to:
  ## 1. create appropriate secret with name registrysecret in namespace kube-system of your k8s cluster
  ## 2. uncomment following lines:
  ##- kubectl get secret registrysecret -n kube-system -o json |
  ##                  jq ".metadata.namespace = \"${APP_NAMESPACE}\"|
  ##                  del(.metadata.annotations,.metadata.creationTimestamp,.metadata.resourceVersion,.metadata.selfLink,.metadata.uid)" |
  ##                  kubectl apply -f -
  - werf --version; pwd; set -x
  ## Next command makes deploy and will be discussed further
  - werf kube deploy
      --tag-ci
      --namespace ${APP_NAMESPACE}
      --set "global.env=${CI_ENVIRONMENT_SLUG}"
      --set "global.ci_url=$(echo ${CI_ENVIRONMENT_URL} | cut -d / -f 3)"
      ${CI_REGISTRY_IMAGE}
  ## It is important that the deploy stage depends on the build stage. If the build stage fails, deploy stage should not start.
  dependencies:
  - build
  ## We need to use deploy runner, because werf needs to interact with the kubectl
  tags:
  - deploy
```

Pay attention to `werf kube deploy` command. It is the main step in deploying the application and note that:
* it is important to use `--tag-ci`, `--tag-branch` or `--tag-commit` options here (in `werf kube deploy`) otherwise you can't use werf templates `werf_container_image` and `werf_container_env` (see more [about deploy]({{ site.baseurl }}/reference/deploy/templates.html));
* we've used the `APP_NAMESPACE` variable, defined at the top of the `.gitlab-ci.yml` file (it is not one of the GitLab [environments](https://docs.gitlab.com/ee/ci/variables/README.html));
* we've passed the `global.env` parameter, which contains the name of the environment. You can access it in helm templates as `.Values.global.env` in Go-template's blocks, to configure deployment of your application according to the environment;
* we've passed the `global.ci_url` parameter, which contains an URL of the environment. You can use it in your helm templates, e.g. to configure ingress.

#### Test

As you may remember, we need to deploy to the stage environment only code from the master branch, and we allow automatic deploy.

Add the following lines to `.gitlab-ci.yml` file:

```yaml
Deploy to Stage:
  <<: *base_deploy
  environment:
    name: stage
    url: http://${CI_PROJECT_NAME}-${CI_COMMIT_REF_SLUG}.kube.DOMAIN
  only:
  - master
  except:
  - schedules
```

We use `base_deploy` template and define unique to stage environment variables such as — environment name and URL. Because of this approach, we get a small and readable job description and as a result — more compact and readable `.gitlab-ci.yml`.

#### Production

The production environment is the last and important environment as it is public accessible! We only deploy tags on production environment and only by manual action (maybe at night?).

Add the following lines to `.gitlab-ci.yml` file:

```yaml
Deploy to Production:
  <<: *base_deploy
  environment:
    name: production
    url: http://www.company.my
  only:
  - tags
  when: manual
  except:
  - schedules
```

Pay attention to `environment.url` — as we deploy the application to production (to public access), we definitely have a static domain for it. We write it here and also use in helm templates as `.Values.global.ci_url` (see definition of `base_deploy` template earlier).

### Cleanup stages

Werf has an efficient cleanup functionality which can help you to avoid overflow registry and disk space on _build nodes_. You can read more about werf cleanup functionality [here]({{ site.baseurl }}/reference/registry/cleaning.html).

In the results of werf works, we have images in a registry and a build cache. Build cache exists only on build node and to the registry werf push only built images.

There are two stages — `cleanup_registry` and `cleanup_builder`, in the `.gitlab-ci.yml` file for the cleanup process. Every stage has only one job in it and order of stage definition (see `stages` list in the top of the `.gitlab-ci.yml` file) is essential.

The first step in the cleanup process is to clean registry from unused images (built from stale or deleted branches and so on — see more [about werf cleanup]({{ site.baseurl }}/reference/registry/cleaning.html)). This work will be done on the `cleanup_registry` stage. On this stage, werf connect to the registry and to the kubernetes cluster. That is why at this stage we need to use deploy runner. From kubernetes cluster, werf gets info about images are currently used by pods.

The second step — is to clean up cache on the build node **after** registry has been cleaned. The important word is — after, because werf will use info from the registry to clean up build cache, and if you haven't cleaned registry you won't get an efficiently cleaned build cache. That's why is important that the `cleanup_builder` stage starts after the `cleanup_registry` stage.

Add the following lines to `.gitlab-ci.yml` file:

```yaml
Cleanup registry:
  stage: cleanup_registry
  script:
  - werf --version; pwd; set -x
  - werf dimg cleanup repo ${CI_REGISTRY_IMAGE}
  only:
  - schedules
  tags:
  - deploy

Cleanup builder:
  stage: cleanup_builder
  script:
  - werf --version; pwd; set -x
  - werf dimg stages cleanup local
      --improper-cache-version
      --improper-git-commit
      --improper-repo-cache
      ${CI_REGISTRY_IMAGE}
  only:
  - schedules
  tags:
  - build
```

To use cleanup, you should create `Personal Access Token` with necessary rights and put it into the `WERF_CLEANUP_REGISTRY_PASSWORD` environment variable. You can simply put this variable in GitLab variables of your project. To do this, go to your project in GitLab Web interface, then open `Settings` —> `CI/CD` and expand `Variables`. Then you can create a new variable with a key `WERF_CLEANUP_REGISTRY_PASSWORD` and a value consisting of `Personal Access Token`.

> Note: `WERF_CLEANUP_REGISTRY_PASSWORD` environment variable is used by werf only for deleting images in the registry when running `werf dimg cleanup repo` command. In the other cases, werf uses `CI_JOB_TOKEN`.

For the demo project create `Personal Access Token` for your account. To do this, in GitLab go to your settings, then open `Access Token` section. Fill token name, make check in Scope on `api` and click `Create personal access token` — you'll get the `Personal Access Token`.

Both stages start only by schedules. You can define schedule in `CI/CD` —> `Schedules` section of your project in GitLab Web interface. Push `New schedule` button, fill description, define cron pattern, leave the master branch in target branch (because it doesn't affect on cleanup), check on Active (if it's not checked) and save pipeline schedule. That's all!

## Building the application

To build the application create the following `.gitlab-ci.yml` **in the root folder** of the repository:

<details markdown="1">
<summary>The complete <b><i>.gitlab-ci.yml</i></b> file...</summary>

{% raw %}
```yaml
variables:
  APP_NAMESPACE: ${CI_PROJECT_NAME}-${CI_ENVIRONMENT_SLUG}

stages:
- build
- deploy
- cleanup_registry
- cleanup_builder

Build:
  stage: build
  script:
  ## The line below uses for debugging
  - werf --version; pwd; set -x
  ## Pull stages caсhe
  - werf dimg stages pull ${CI_REGISTRY_IMAGE}
  ## build dimg's than push it with the stages cache
  - werf dimg bp --with-stages ${CI_REGISTRY_IMAGE} --tag-ci
  ## Specify to use runner with the build tag.
  tags:
  - build
  ## Cleanup will use schedules, and it is not necessary to rebuild images on running cleanup jobs.
  ## Therefore we need to specify
  except:
  - schedules

.base_deploy: &base_deploy
  stage: deploy
  script:
  ## create k8s namespace we will use if it doesn't exist.
  - kubectl get ns ${APP_NAMESPACE} || kubectl create namespace ${APP_NAMESPACE}
  ## If your application use private registry, you have to:
  ## 1. create appropriate secret with name registrysecret in namespace kube-system of your k8s cluster
  ## 2. uncomment following lines:
  ##- kubectl get secret registrysecret -n kube-system -o json |
  ##                  jq ".metadata.namespace = \"${APP_NAMESPACE}\"|
  ##                  del(.metadata.annotations,.metadata.creationTimestamp,.metadata.resourceVersion,.metadata.selfLink,.metadata.uid)" |
  ##                  kubectl apply -f -
  - werf --version; pwd; set -x
  ## Next command makes deploy and will be discussed further
  - werf kube deploy
      --tag-ci
      --namespace ${APP_NAMESPACE}
      --set "global.env=${CI_ENVIRONMENT_SLUG}"
      --set "global.ci_url=$(echo ${CI_ENVIRONMENT_URL} | cut -d / -f 3)"
      ${CI_REGISTRY_IMAGE}
  ## It is important that the deploy stage depends on the build stage. If the build stage fails, deploy stage should not start.
  dependencies:
  - build
  ## We need to use deploy runner, because werf needs to interact with the kubectl
  tags:
  - deploy

Review:
  <<: *base_deploy
  environment:
    name: review/${CI_COMMIT_REF_SLUG}
    ## Of course, you need to change domain suffix (`kube.DOMAIN`) of the url if you want to use it in you helm templates.
    url: http://${CI_PROJECT_NAME}-${CI_COMMIT_REF_SLUG}.kube.DOMAIN
    on_stop: Stop review
  only:
  - branches
  except:
  - master
  - schedules

Stop review:
  stage: deploy
  script:
  - werf --version; pwd; set -x
  - werf kube dismiss --namespace ${APP_NAMESPACE} --with-namespace
  environment:
    name: review/${CI_COMMIT_REF_SLUG}
    action: stop
  tags:
  - deploy
  only:
  - branches
  except:
  - master
  - schedules
  when: manual

Deploy to Stage:
  <<: *base_deploy
  environment:
    name: stage
    url: http://${CI_PROJECT_NAME}-${CI_COMMIT_REF_SLUG}.kube.DOMAIN
  only:
  - master
  except:
  - schedules

Deploy to Production:
  <<: *base_deploy
  environment:
    name: production
    url: http://www.company.my
  only:
  - tags
  when: manual
  except:
  - schedules

Cleanup registry:
  stage: cleanup_registry
  script:
  - werf --version; pwd; set -x
  - werf dimg cleanup repo ${CI_REGISTRY_IMAGE}
  only:
  - schedules
  tags:
  - deploy

Cleanup builder:
  stage: cleanup_builder
  script:
  - werf --version; pwd; set -x
  - werf dimg stages cleanup local
      --improper-cache-version
      --improper-git-commit
      --improper-repo-cache
      ${CI_REGISTRY_IMAGE}
  only:
  - schedules
  tags:
  - build
```
{% endraw %}

</details>

With the added `.gitlab-ci.yml`, you can build the application on any runner with such werf advantages as using stages cache.

When `werf dimg stages pull <registry_address>` executed, werf looks into config, calculates signatures of stages and pulls only last available in a Docker registry image, according to stages conveyor.

> If you need to pull images of every stage you can use `--all` option with `werf dimg stages pull` command.

Make changes in the config and push it. Retry build stage several times and compare job logs. You will see that werf uses cache and build only stage with changes and stages after that stage, according to the stage conveyor.

## Conclusions

Werf can use a distributed cache and can work with:
* **Dynamic resources.** You can start build node on-demand, pull stages cache and build n image.
* **More than one build nodes.** You can have more than one build nodes in your environment.

The only steps you need are:
* pull existing stages cache from the Docker registry with the `werf dimg stages pull REPO` command before building;
* build images and push it with stages cache to the Docker registry with the `werf dimg bp --with-stages REPO` command.
