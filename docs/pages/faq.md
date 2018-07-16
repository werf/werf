---
title: Dapp Frequently Asked Questions
sidebar: faq
permalink: faq.html
---

## General

**Q: Can i use different dapp,docker version for build and for deploy?**

A: In some case - you can and it will work, but please try to avoid this and use latest dapp version.


## Building


**Q: How to use `dimg_group` in YAML dappfiles?**

A: You can't use `dimg_group` in `dappfile.y[a]ml`.

`dappfile.y[a]ml` don't allow dimg nesting. All of dimg's and artifact's have to be described one by one and have to be separated by `---`.

E.g.:
```
dimg: container1
from: alpine
---
dimg: container2
from: alpine
```


**Q: How to specify ssh keys?**

A: Use `--ssh-key=path-to-id-rsa` option. E.g. `dapp dimg build --ssh-key=path-to-id-rsa`.


**Q: How to specify stageDependency to all files in all subdirectories?**

A: You can use `**/*` mask.

**Q: How to convert `COPY . /var/app/` instruction from Dockerfile to dappfile?**

A: To add files from local git repository you can use following:
```
git:
- add: /
  to: /var/app
```

**Q: I'v added files from git repository, but i can't access it.**

A: You can't access files on stage `before_install` because dapp clone repos right before runs stage install, and therefore you can access files on any stage starting from `install` (e.g install, beforeSetup, setup)


**Q: How can i build only one image?**

A: You can use `dapp dimg build <DIMG_NAME>`

If you have three dimg's in dappfile.yaml, e.g:
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

you can use `dapp dimg build container2` to build only container2.


**Q: Can i set environment variable to use it in all build stages?**

A: We recommend you to build an image which building instructions depends on you code but not on environment in build dime. In other words you better build one image, which you can run in any environment, and this image have to change its logic when it runs rely on environment variables. If building stage will depends on such black box like changing environments you can get an unexpected behavior of dapp builder and unexpected results.

Environment variables which have been set in docker.* section will be added by builder on last stage - `docker_instructions` - and will not be accessible on other build stages. This behavior not affect variables declared in docker.from image - they will be accessible on all stages.

Also you can use ANSIBLE_ARGS env when you use ansible builder. E.g. you can `export ANSIBLE_ARGS=-vvv` and get verbose ansible output.


**Q: How can i find image name after build?**

A: Use `dapp dimg tag <DIMG_NAME>` to tag your image.

```
$ dapp dimg tag helo-wrld
testing_dapp
  testing_dapp: calculating stages signatures         [RUNNING]
  testing_dapp: calculating stages signatures              [OK] 0.5 sec
  custom
    helo-wrld/testing_dapp:latest                   [EXPORTING]
    helo-wrld/testing_dapp:latest                          [OK] 2.11 sec
Running time 2.64 seconds

$ dapp dimg tag helo-wrld --tag-plain test1
testing_dapp
  testing_dapp: calculating stages signatures         [RUNNING]
  testing_dapp: calculating stages signatures              [OK] 0.39 sec
  custom
    helo-wrld/testing_dapp:test1                    [EXPORTING]
    helo-wrld/testing_dapp:test1                           [OK] 2.34 sec
Running time 2.77 seconds
```

**Q: I use dapp 0.7, alpine image and get error `standard_init_linux.go:178: exec user process caused "no such file or directory"` on build**

A: Dapp 0.7 doesn't support `alpine` image, so please use latest dapp version.



**Q: Can i use different images for artifact and for dimg?**

A: Yes, you can.



**Q: Can i push image to private repository**

A: Yes, you can use `--registry-username` and `--registry-password` options.

In general for authorisatoin in registry dapp use:
* `--registry-username` and `--registry-password` options if specified.
* CI_JOB_TOKEN (in CI environment, e.g. GitLab);
* docker config of user


**Q: How can i push to GCP?**

A: To push to GPC you can complete the following steps:
{% raw %}
```
dapp dimg build
dapp dimg tag --tag-ci <REPO>
docker login <REPO>
docker push $(docker images <REPO> --format "{{.Repository}}:{{.Tag}}")
dapp dimg flush local
```
{% endraw %}
Dapp support push to public and private repositories, but it not works with some platforms (such as GCP).


## Deploying

**Q: How to debug `Error: render error in...`?**

A: You can use `dapp kube render` to render templates and `dapp kube lint` to validate that it follows the conventions and requirements of the Helm chart standard.

**Q: How to resolve `ErrImagePull` after deploy?**

A: It's not a dapp problem. Most likly there is no access to your private repository and if so, you can read about how to add a registry secret in kubernetes documentation [here...](https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/).

**Q: How to change .Release.Name?**

A: Use DAPP_HELM_RELEASE_NAME environment.


**Q: Can i use several tags simultaneously?**

A: Yes.

```
dapp dimg push --tag custom1 --tag custom2 --tag-ci --tag-branch --tag-commit
```


**Q: How to deploy a several application with different names?**

A: You can pass a variable, e.g. `dapp kube deploy --set global.ciProjectName=$CI_PROJECT_NAME ...` and use it in deployment template as {% raw %}`{{ .Values.global.ciProjectName }}`{% endraw %}.
