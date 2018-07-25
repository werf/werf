---
title: Dapp Frequently Asked Questions
sidebar: faq
permalink: faq.html
---

## General

[Q: Can I use different dapp/docker version for build and for deploy?](#general-1){:id="general-1"}

In some case, you can and it will work, but please try to avoid this and use latest dapp version.


## Building

[Q: How to use **dimg_group** in YAML dappfiles?](#building-1){:id="building-1"}

You can't use `dimg_group` in `dappfile.y[a]ml`.

`dappfile.y[a]ml` don't allow dimg nesting. All of dimgs and artifacts have to be described one by one and have to be separated by `---`.

```
dimg: container1
from: alpine
---
dimg: container2
from: alpine
```


[Q: How to specify ssh keys?](#building-2){:id="building-2"}

Use `--ssh-key=path-to-id-rsa` option. E.g. `dapp dimg build --ssh-key=path-to-id-rsa`.


[Q: How to specify stageDependency to all files in all subdirectories?](#building-3){:id="building-3"}

You can use `**/*` mask.


[Q: How to convert **COPY . /var/app** instruction from Dockerfile to dappfile?](#building-4){:id="building-4"}

To add files from local git repository you can use the following:

```
git:
- add: /
  to: /var/app
```


[Q: I've added files from a git repository, but I can't access it](#building-5){:id="building-5"}

You can't access files on stage `before_install` because dapp clone repositories on stage `g_a_archive`, and therefore you can access files on any stage after (e.g `install`, `before_setup`, `setup`, `build_artifact`).


[Q: How can I build only one image?](#building-6){:id="building-6"}

You can use `dapp dimg build <DIMG_NAME>`.

If you have three dimgs in dappfile.yaml, e.g:

```
dimg: container1
from: alpine
---
dimg: container2
from: alpine
---
dimg: container3
from: alpine
```

You can use `dapp dimg build container2` to build the only `container2`.


[Q: Can I set an environment variable to use it in all build stages?](#building-7){:id="building-7"}

We recommend to build an image which building instructions depend on your code but not on an environment in build time. In other words, you better build one image, which you can run in any environment, and this image has to change its logic when it runs rely on environment variables. If building stage will depend on such black box like changing environments you can get an unexpected behavior of dapp builder and unexpected results.

Environment variables which have been set in `docker` dappfile section will be added by a builder on the last dimg stage, `docker_instructions`, and will not be accessible on other build stages.

Also, you can use `ANSIBLE_ARGS` env when you use ansible builder. E.g. you can `export ANSIBLE_ARGS=-vvv` and get verbose ansible output.


[Q: How can I find image name after build?](#building-8){:id="building-8"}

Use `dapp dimg tag <DIMG_NAME>` to tag your image.

```
$ dapp dimg tag hello-world
testing_dapp
  testing_dapp: calculating stages signatures         [RUNNING]
  testing_dapp: calculating stages signatures              [OK] 0.5 sec
  custom
    helo-wrld/testing_dapp:latest                   [EXPORTING]
    helo-wrld/testing_dapp:latest                          [OK] 2.11 sec
Running time 2.64 seconds

$ dapp dimg tag hello-world --tag-plain test1
testing_dapp
  testing_dapp: calculating stages signatures         [RUNNING]
  testing_dapp: calculating stages signatures              [OK] 0.39 sec
  custom
    helo-wrld/testing_dapp:test1                    [EXPORTING]
    helo-wrld/testing_dapp:test1                           [OK] 2.34 sec
Running time 2.77 seconds
```

[Q: I use dapp 0.7, alpine image and get error **standard_init_linux.go:178: exec user process caused "no such file or directory"** on build](#building-9){:id="building-9"}

Dapp 0.7 doesn't support `alpine` image, so please use latest dapp version.


[Q: Can I use different images for artifact and for dimg?](#building-10){:id="building-10"}

Yes, you can.


[Q: Can I push image to private repository?](#building-11){:id="building-11"}

Yes, you can use `--registry-username` and `--registry-password` options.

In general for authorization in registry dapp use:
* `--registry-username` and `--registry-password` options if specified.
* `CI_JOB_TOKEN` (in CI environment, e.g. GitLab).
* Docker config of a user running dapp, `~/.docker/config.json`.


[Q: How can I push to GCP?](#building-12){:id="building-12"}

To push to GCP you can use the following workaround:

{% raw %}
```
dapp dimg build
dapp dimg tag --tag-ci <REPO>
docker login <REPO>
docker push $(docker images <REPO> --format "{{.Repository}}:{{.Tag}}")
dapp dimg flush local
```
{% endraw %}

Dapp support push to public and private repositories, but it doesn't work with some platforms such as GCP.


## Deploying

[Q: How to debug **Error: render error in...**](#deploying-1){:id="deploying-1"}

You can use `dapp kube render` to render templates and `dapp kube lint` to validate that it follows the conventions and requirements of the Helm chart standard.


[Q: How to resolve **ErrImagePull** after deploy?](#deploying-2){:id="deploying-2"}

It's not a dapp problem. Most likely there is no access to your private repository and if so, you can read about how to add a registry secret in kubernetes documentation [here...](https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/).


[Q: How to change helm release name?](#deploying-3){:id="deploying-3"}

Use DAPP_HELM_RELEASE_NAME environment.


[Q: Can I use several tags at the same time?](#deploying-4){:id="deploying-4"}

Yes.

```
dapp dimg push --tag custom1 --tag custom2 --tag-build-id --tag-ci --tag-branch --tag-commit
```

[Q: How to deploy several applications with different names?](#deploying-5){:id="deploying-5"}

You can pass a variable, e.g. `dapp kube deploy --set global.ciProjectName=$CI_PROJECT_NAME ...` and use it in deployment template as {% raw %}`{{ .Values.global.ciProjectName }}`{% endraw %}.
