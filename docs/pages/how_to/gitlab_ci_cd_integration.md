---
title: Gitlab CI/CD integration
sidebar: how_to
permalink: how_to/gitlab_ci_cd_integration.html
author: Artem Kladov <artem.kladov@flant.com>
---

## Task Overview

Setup CI/CD process using GitLab CI and Werf.

## Requirements

To begin, you'll need the following:
* Kubernetes cluster and the kubectl CLI tool configured to communicate with the cluster;
* Server with GitLab above 10.x version (or account on [SaaS GitLab](https://gitlab.com/));
* Docker registry (GitLab embedded or somewhere else);
* An application you can successfully build and deploy with werf;

## Infrastructure

![scheme]({{ site.baseurl }}/images/howto_gitlabci_scheme.png)

* Kubernetes cluster
* GitLab with docker registry enabled
* Werf node (node for build and deployment)

We don't recommend to run werf in docker as it can give an unexpected result.

In order to set up the CI/CD process, you need to describe build, deploy and cleanup stages. For all of these stages, a GitLab runner with `shell` executor is needed to run `werf`.

Werf use `~/.werf/` folder to store build cache and other files. Werf assumes this folder is preserved for all pipelines. That is why for build processes we don't recommend you to use environments which are don't preserve a GitLab runner's state (files between runs), e.g. some cloud environments.

The deployment process needs access to the cluster through kubectl and you can simply use master node as Werf node. But in production case you better setup werf on a separate node.

This HowTo assumes you setup werf on the master kubernetes node.

Eventually, Werf node needs access:
- to the application git repository
- to the docker registry
- to the kubernetes cluster.

You need to set up GitLab runner with `werf` tags.

### Setup runner

On the master kubernetes node, you need to install and set up GitLab runner. Follow these steps:

1. Create GitLab project and push project code into it.
1. Get a registration token for the runner: in your GitLab project open `Settings` —> `CI/CD`, expand `Runners` and find the token in the section `Setup a specific Runner manually`.
1. [Install gitlab-runner](https://docs.gitlab.com/runner/install/linux-manually.html).
1. Register the `gitlab-runner`.

    [Use these steps](https://docs.gitlab.com/runner/register/index.html) to register runner, but:
    * enter the `werf` tag for runner;
    * enter executor `shell` for runner;
1. Add the `gitlab-runner` user into the `docker` group.

   ```bash
   sudo usermod -Ga docker gitlab-runner
   ```

1. Install [Werf dependencies]({{ site.baseurl }}/how_to/installation.html#install-dependencies) including Helm.
1. Install [`Multiwerf`](https://github.com/flant/multiwerf) under the `gitlab-runner` user:

   ```
   sudo su gitlab-runner
   mkdir -p ~/bin
   cd ~/bin
   curl -L https://raw.githubusercontent.com/flant/multiwerf/master/get.sh | bash
   ```

1. Copy kubectl config to the home folder of the `gitlab-runner` user.
   ```bash
   mkdir -p /home/gitlab-runner/.kube &&
   sudo cp -i /etc/kubernetes/admin.conf /home/gitlab-runner/.kube/config &&
   sudo chown -R gitlab-runner:gitlab-runner /home/gitlab-runner/.kube
   ```

> If you install Werf not on the master kubernetes node, you must [install](https://kubernetes.io/docs/setup/independent/install-kubeadm/#installing-docker) Docker and setup kubectl.

## Pipeline

When you have working runner, you are ready to set up GitLab pipeline.

When GitLab starts a job, it sets a list of [environments](https://docs.gitlab.com/ee/ci/variables/README.html), and we will use some of them.

Create a `.gitlab-ci.yml` file in the project's root directory and add the following lines:

```yaml
stages:
  - build
  - deploy
  - cleanup
```

We've defined the following stages:
* `build` — stage for building application images;
* `deploy` — stage for deploying built images on environments such as — stage, test, review, production or any other;
* `cleanup` — stage for cleaning up registry and build cache.

### Build stage

Add the following lines to the `.gitlab-ci.yml` file:

```yaml
Build:
  stage: build
  script:
    - type multiwerf && source <(multiwerf use 1.0 beta)
    - type werf && source <(werf ci-env gitlab --tagging-strategy tag-or-branch --verbose)
    - werf build-and-publish --stages-storage :local
  tags:
    - werf
  except:
    - schedules
```

Cleanup will use schedule, and it is not necessary to rebuild images on running cleanup job. Therefore we need to specify `except: schedules` in the build stage above.

For registry authorization on push/pull operations werf use `CI_JOB_TOKEN` GitLab environment (see more about [GitLab CI job permissions model](https://docs.gitlab.com/ee/user/project/new_ci_build_permissions_model.html)), and this is the most recommended way you to use (see more about [werf registry authorization]({{ site.baseurl }}/reference/registry/authorization.html)). In a simple case, when you use GitLab with enabled container registry in it, you needn't do anything for authorization.

If you want that werf won't use `CI_JOB_TOKEN` or you don't use GitLab's Container registry (e.g. `Google Container Registry`) please read more [here]({{ site.baseurl }}/reference/registry/authorization.html).

### Deploy stage

List of environments in `deploy` stage depends on your needs, but usually it includes:
* Review environment. It is a dynamic (or so-called temporary) environment for taking the first look at the result of development by developers. This environment will be deleted (stopped) after branch deletion in the repository or in case of manual environment stop.
* Test environment. The test environment may be used by developer when he ready to show his work. On the test environment, ones can manually deploy the application from any branches or tags.
* Stage environment. The stage environment may be used for taking final tests of the application, and deploy can proceed automatically after merging into master branch (but this is not a rule).
* Production environment. Production environment is the final environment in the pipeline. On production environment should be deployed the only production-ready version of the application. We assume that on production environment can be deployed only tags, and only after a manual action.

Of course, parts of CI/CD process described above is not a rule, and you can have your own.

First of all, we define a template for the deploy jobs — this will decrease the size of the `.gitlab-ci.yml` and will make it more readable. We will use this template in every deploy stage further.

Add the following lines to `.gitlab-ci.yml` file:

```yaml
.base_deploy: &base_deploy
  stage: deploy
  script:
    - type multiwerf && source <(multiwerf use 1.0 beta)
    - type werf && source <(werf ci-env gitlab --tagging-strategy tag-or-branch --verbose)
    ## Next command makes deploy and will be discussed further
    - werf deploy --stages-storage :local
        --set "global.ci_url=$(echo ${CI_ENVIRONMENT_URL} | cut -d / -f 3)"
  ## It is important that the deploy stage depends on the build stage. If the build stage fails, deploy stage should not start.
  dependencies:
    - Build
  tags:
    - werf
```

Pay attention to `werf deploy` command. It is the main step in deploying the application and note that we've passed the `global.ci_url` parameter, which will contain an URL of the environment. You can use it in your `helm` templates e.g. to configure ingress.

To configure deployment of your application according to the environment you can use `.Values.global.env` variable in Go-template's blocks of your `helm` templates - werf sets `global.env` parameter according to the CI environment.

#### Review

As it was said earlier, review environment — is a dynamic (or temporary, or developer) environment for taking the first look at the result of development by developers.

Add the following lines to `.gitlab-ci.yml` file:

```yaml
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
    - type multiwerf && source <(multiwerf use 1.0 beta)
    - type werf && source <(werf ci-env gitlab --tagging-strategy tag-or-branch --verbose)
    - werf dismiss --with-namespace
  environment:
    name: review/${CI_COMMIT_REF_SLUG}
    action: stop
  tags:
    - werf
  only:
    - branches
  except:
    - master
    - schedules
  when: manual
```

We've defined two jobs:
1. Review.
    In this job, we set a name of the environment based on CI_COMMIT_REF_SLUG GitLab variable. For every branch, GitLab will create a unique environment.

    The `url` parameter of the job you can use in you helm templates to set up e.g. ingress.
2. Stop review.
    In this job, werf will delete helm release and delete kubernetes namespace itself (see more about [werf dismiss]({{ site.baseurl }}/cli/main/dismiss.html)). This job will be available for the manual run and also it will run by GitLab in case of e.g branch deletion.

Review jobs needn't run on pushes to git master branch, because review environment is for developers.

As a result, you can stop this environment manually after deploying the application on it, or GitLab will stop this environment when you merge your branch into master with source branch deletion enabled.

#### Test

We've decided to don't give an example of a test environment in this howto because test environment is very similar to stage environment (see below). Try to describe test environment by yourself and we hope you will get a pleasure.

#### Stage

As you may remember, we need to deploy to the stage environment only code from the master branch and we allow automatic deploy.

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

We use `base_deploy` template and define only unique to stage environment variables such as — environment name and URL. Because of this approach, we get a small and readable job description and as a result — more compact and readable `.gitlab-ci.yml`.

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

Pay attention to `environment.url` — as we deploy the application to production (to public access), we definitely have a static domain for it. We simply write it here and also use in helm templates as `.Values.global.ci_url` (see definition of `base_deploy` template earlier).

### Cleanup stage

Werf has an efficient cleanup functionality which can help you to avoid overflow registry and disk space on build nodes. You can read more about werf cleanup functionality [here]({{ site.baseurl }}/reference/registry/cleaning.html).

In the results of werf works, we have images in a registry and a build cache. Build cache exists only on build node and to the registry werf push only built images.

There is a `cleanup` stage in the `.gitlab-ci.yml` file for the cleanup process.

To use cleanup you should create `Personal Access Token` with necessary rights and login to the docker registy with it before cleanup.

For demo project simply create `Personal Access Token` for your account. To do this, in GitLab go to your settings, then open `Access Token` section. Fill token name, make check in Scope on `api` and click `Create personal access token` — you'll get the `Personal Access Token`.

To put `Personal Access Token` in a GitLab variable, go to your project in GitLab Web interface, then open `Settings` —> `CI/CD` and expand `Variables`. Then, create a new variable with a key `WERF_IMAGES_CLEANUP_PASSWORD` and a value consisting `Personal Access Token`. Make this variable protected for security reason.

Add the following lines to `.gitlab-ci.yml` file:


```yaml
Cleanup:
  stage: cleanup
  script:
    - type multiwerf && source <(multiwerf use 1.0 beta)
    - type werf && source <(werf ci-env gitlab --tagging-strategy tag-or-branch --verbose)
    - docker login -u nobody -p ${WERF_IMAGES_CLEANUP_PASSWORD} ${WERF_IMAGES_REPO}
    - werf cleanup --stages-storage :local
  only:
    - schedules
  tags:
    - werf
```

Cleanup stage will start only by schedule. You can define schedule in `CI/CD` —> `Schedules` section of your project in GitLab Web interface. Push `New schedule` button, fill description, define cron pattern, leave the master branch in target branch (because it doesn't affect on cleanup), check on Active (if it's not checked) and save pipeline schedule. That's all!

> How it works:
   - `werf ci-env` creates a temporary docker configuration file for each GitLab job, sets the path to that file in the DOCKER_CONFIG variable, and exports it;
   - `docker login` uses a temporary configuration file from DOCKER_CONFIG environment for registry authentication;
   - `werf cleanup` also uses a temporary configuration file from DOCKER_CONFIG.


## Complete `.gitlab-ci.yml` file

```yaml
stages:
  - build
  - deploy
  - cleanup

Build:
  stage: build
  script:
    - type multiwerf && source <(multiwerf use 1.0 beta)
    - type werf && source <(werf ci-env gitlab --tagging-strategy tag-or-branch --verbose)
    - werf build-and-publish --stages-storage :local
  tags:
    - werf
  except:
    - schedules

.base_deploy: &base_deploy
  stage: deploy
  script:
    - type multiwerf && source <(multiwerf use 1.0 beta)
    - type werf && source <(werf ci-env gitlab --tagging-strategy tag-or-branch --verbose)
    ## Next command makes deploy and will be discussed further
    - werf deploy --stages-storage :local
        --set "global.ci_url=$(echo ${CI_ENVIRONMENT_URL} | cut -d / -f 3)"
  ## It is important that the deploy stage depends on the build stage. If the build stage fails, deploy stage should not start.
  dependencies:
    - Build
  tags:
    - werf

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
    - type multiwerf && source <(multiwerf use 1.0 beta)
    - type werf && source <(werf ci-env gitlab --tagging-strategy tag-or-branch --verbose)
    - werf dismiss --with-namespace
  environment:
    name: review/${CI_COMMIT_REF_SLUG}
    action: stop
  tags:
    - werf
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

Cleanup:
  stage: cleanup
  script:
    - type multiwerf && source <(multiwerf use 1.0 beta)
    - type werf && source <(werf ci-env gitlab --tagging-strategy tag-or-branch --verbose)
    - docker login -u nobody -p ${WERF_IMAGES_CLEANUP_PASSWORD} ${WERF_IMAGES_REPO}
    - werf cleanup --stages-storage :local
  only:
    - schedules
  tags:
    - werf
```
