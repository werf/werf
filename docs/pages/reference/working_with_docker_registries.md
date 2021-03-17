---
title: Working with Docker registries
sidebar: documentation
permalink: reference/working_with_docker_registries.html
author: Alexey Igrychev <alexey.igrychev@flant.com>
---

There are several types of commands that are working with the Docker registries and require the appropriate authorization:

* [During the building process]({{ site.baseurl }}/reference/build_process.html), werf may pull base images from the Docker registry and pull/push _stages_ in distributed builds.
* [During the publishing process]({{ site.baseurl }}/reference/publish_process.html), werf creates and updates _images_ in the Docker registry.
* [During the cleaning process]({{ site.baseurl }}/reference/cleaning_process.html), werf deletes _images_ and _stages_ from the Docker registry.
* [During the deploying process]({{ site.baseurl }}/reference/deploy_process/deploy_into_kubernetes.html), werf requires access to the _images_ from the Docker registry and to the _stages_ that could also be stored in the Docker registry.

## Supported implementations

|                 	                    | Build and Publish 	    | Cleanup                         	                                    |
| -------------------------------------	| :-----------------------:	| :-------------------------------------------------------------------:	|
| [_AWS ECR_](#aws-ecr)             	|         **ok**        	|                    **ok (with native API)**                   	    |
| [_Azure CR_](#azure-cr)            	|         **ok**        	|                            **ok (with native API)**                   |
| _Default_         	                |         **ok**        	|                            **ok**                            	        |
| [_Docker Hub_](#docker-hub)      	    |         **ok**        	|                    **ok (with native API)**                   	    |
| _GCR_             	                |         **ok**        	|                            **ok**                            	        |
| [_GitHub Packages_](#github-packages) |         **ok**        	| **ok (with native API and only in private GitHub repositories)** 	    |
| _GitLab Registry_ 	                |         **ok**        	|                            **ok**                            	        |
| _Harbor_          	                |         **ok**        	|                            **ok**                            	        |
| _JFrog Artifactory_         	        |         **ok**        	|                            **ok**                            	        |
| [_Quay_](#quay)            	        |         **ok**        	|                            **ok**                            	        |

> In the nearest future we will add support for Nexus Registry

The following implementations are fully supported and do not require additional actions except [docker authorization](#docker-authorization):
* _Default_.
* _GCR_.
* _GitLab Registry_.
* _Harbor_.

There are two main issues for the rest:
1. _Azure CR_, _AWS ECR_, _Docker Hub_ and _GitHub Packages_ implementations provide Docker Registry API but do not implement the delete tag method and offer it with native API. 
Therefore, werf may require extra credentials for [cleanup commands]({{ site.baseurl }}/reference/cleaning_process.html).
2. Some implementations do not support nested repositories (_Docker Hub_, _GitHub Packages_ and _Quay_) or support, but the user should create repositories manually using UI or API (_AWS ECR_). Thus, _multirepo_ images repo mode might require specific use.

## How to store images

The _images repo_ and _images repo mode_ params define where and how to store images (read more about [image naming]({{ site.baseurl }}/reference/publish_process.html#naming-images)).

The _images repo_ can be **registry** or **repository** address.

The resulting name of a docker image depends on the _images repo mode_:
- `IMAGES_REPO:IMAGE_NAME-TAG` pattern for a **monorepo** mode;
- `IMAGES_REPO/IMAGE_NAME:TAG` pattern for a **multirepo** mode.

> The images are not stored in registries. Therefore, it is not acceptable to use **registry** as _images repo_ and **monorepo** mode. Also, the user should be aware and do not use **registry** when using nameless image (`image: ~`). 
>
> Do not forget to rename nameless image in werf.yaml and delete nameless managed image (`werf managed-images rm '~'`) if you want to use **registry** with **multirepo** _images repo mode_

Thus, the user has only 3 combinations of using _images repo_ and _images repo mode_.

|                	                    | registry + multirepo 	    | repository + monorepo 	    | repository + multirepo 	    |
|--------------------------------------	|:------------------------: |:----------------------------:	|:---------------------------:	|
| [_AWS ECR_](#aws-ecr)             	| **ok**                   	| **ok**                    	| **ok**                     	|
| [_Azure CR_](#azure-cr)           	| **ok**                   	| **ok**                    	| **ok**                     	|
| _Default_         	                | **ok**                   	| **ok**                    	| **ok**                     	|
| [_Docker Hub_](#docker-hub)      	    | **ok**                   	| **ok**                    	| **not supported**         	|
| _GCR_             	                | **ok**                   	| **ok**                    	| **ok**                     	|
| [_GitHub Packages_](#github-packages) | **ok**                   	| **ok**                    	| **not supported**         	|
| _GitLab Registry_ 	                | **ok**                   	| **ok**                    	| **ok**                     	|
| _Harbor_          	                | **ok**                   	| **ok**                    	| **ok**                     	|
| _JFrog Artifactory_                   | **ok**                   	| **ok**                    	| **ok**                     	|
| [_Quay_](#quay)            	        | **ok**                   	| **ok**                    	| **not supported**         	|

Most implementations support nested repositories and work with different _images repo mode_. By default, such implementations has **multirepo** _images repo mode_. The default _images repo mode_ value for rest implementations depends on specified _images repo_. 

## AWS ECR

### How to store images

Working with _AWS ECR_ is not different from the rest implementations. 
The user can use both _images repo modes_ but should manually create repositories before using werf.

### How to cleanup stages and images

werf deletes tags from _AWS ECR_ with _AWS SDK_. 
Therefore, before using [cleanup commands]({{ site.baseurl }}/reference/cleaning_process.html) the user should:
* [Set up _AWS CLI_ installation](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html#cli-quick-configuration) (`aws configure`) or 
* Define `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` environment variables.
      
## Azure CR

### How to cleanup stages and images

The user should do the next steps to enable tag deletion with werf: 
* Install [Azure CLI](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli?view=azure-cli-latest) (`az`).
* Sign in with relevant credentials (`az login`).

> The user must have one of the following roles: `Owner`, `Contributor` or `AcrDelete` (more about [Azure CR roles and permissions](https://docs.microsoft.com/en-us/azure/container-registry/container-registry-roles)) 

## Docker Hub

### How to store images

If you use one Docker Hub account per project and would like to store images separately you should specify account as _images repo_: 
```shell
<ACCOUNT> # e.g. library or index.docker.io/library
``` 

> Be aware that using nameless image with such an approach leads to failures  and issues that are difficult to debug.
>
> Do not forget to rename nameless image in _werf.yaml_ and delete nameless managed image (`werf managed-images rm '~'`) if you want to move to such an approach    

If **werf configuration has nameless image (`image: ~`)** or the user want to keep all images in the single repository it is necessary to use a certain repository as _images repo_:
```shell
<ACCOUNT>/<REPOSITORY> # e.g. library/alpine or index.docker.io/library/alpine
```

#### Example

The user has the following:

* Two images in _werf.yaml_: **frontend**, **backend**.
* Docker Hub account: **foo**.

There are two ways to store these images:
1. Images repo **foo** leads to `foo/frontend:tag` and `foo/backend:tag` tags. 
    ```shell
    werf build-and-publish -s=:local -i=foo --tag-custom=tag
    ```
2. Images repo **foo/project** leads to `foo/project:frontend-tag` and `foo/project:backend-tag` tags.
    ```shell
    werf build-and-publish -s=:local -i=foo/project --tag-custom=tag
    ```

### How to cleanup stages and images

To delete tags from _Docker Hub_ repository werf uses _Docker Hub API_ and requires extra user credentials.

The user should specify token or username and password. The following script can be used to get the token:

```shell
HUB_USERNAME=USERNAME
HUB_PASSWORD=PASSWORD
HUB_TOKEN=$(curl -s -H "Content-Type: application/json" -X POST -d '{"username": "'${HUB_USERNAME}'", "password": "'${HUB_PASSWORD}'"}' https://hub.docker.com/v2/users/login/ | jq -r .token)
```

> Be aware that access to the resources is forbidden with a [personal access token](https://docs.docker.com/docker-hub/access-tokens/)

To define credentials check options and related environments:
* For stages storage: `--stages-storage-repo-docker-hub-token` or `--stages-storage-repo-docker-hub-username` and `--stages-storage-repo-docker-hub-password`.
* For images repo: `--images-repo-docker-hub-token` or `--images-repo-docker-hub-username` and `--images-repo-docker-hub-password`.
* For both: `--repo-docker-hub-token` or `--repo-docker-hub-username` and `--repo-docker-hub-password`.

## GitHub Packages

### How to store images

If you want to keep each image in a separate package you should specify _images repo_ without package name:
```shell
docker.pkg.github.com/<ACCOUNT>/<PROJECT> # e.g. docker.pkg.github.com/flant/werf
```

> Be aware that using nameless image with such an approach leads to failures  and issues that are difficult to debug.
>
> Do not forget to rename nameless image in _werf.yaml_ and delete nameless managed image (`werf managed-images rm '~'`) if you want to move to such an approach    

If **werf configuration has nameless image (`image: ~`)** or all images should be stored together use certain package:
```shell
docker.pkg.github.com/<ACCOUNT>/<PROJECT>/<PACKAGE> # e.g. docker.pkg.github.com/flant/werf/image
```

#### Example

The user has the following:

* Two images in _werf.yaml_: **frontend**, **backend**.
* GitHub repository: **github.com/company/project**.

There are two ways to store these images:
1. Images repo **docker.pkg.github.com/company/project** leads to `docker.pkg.github.com/company/project/frontend:tag` and `docker.pkg.github.com/company/project/backend:tag` tags. 
    ```shell
    werf build-and-publish -s=:local -i=docker.pkg.github.com/company/project --tag-custom=tag
    ```
2. Images repo **docker.pkg.github.com/company/project/app** leads to `docker.pkg.github.com/company/project/app:frontend-tag` and `docker.pkg.github.com/company/project/app:backend-tag` tags.
    ```shell
    werf build-and-publish -s=:local -i=docker.pkg.github.com/company/project/app --tag-custom=tag
    ```

### How to cleanup stages and images

To [delete versions of a private package](https://help.github.com/en/packages/publishing-and-managing-packages/deleting-a-package) we are using GraphQL and need GitHub token with `read:packages`, `write:packages`, `delete:packages` and `repo` scopes.

> Be aware that GitHub only supports deleting in private repositories 

To define credentials check options and related environments:
* For stages storage: `--stages-storage-repo-github-token`.
* For images repo: `--images-repo-github-token`.
* For both: `--repo-github-token`.

## Quay

### How to store images

If you want to keep each image in a separate repository you should specify _images repo_ without repository name:
```shell
quay.io/<USER or ORGANIZATION> # e.g. quay.io/werf
```

> Be aware that using nameless image with such an approach leads to failures  and issues that are difficult to debug.
>
> Do not forget to rename nameless image in _werf.yaml_ and delete nameless managed image (`werf managed-images rm '~'`) if you want to move to such an approach 

If **werf configuration has nameless image (`image: ~`)** or all images should be stored together use certain repository:
```shell
quay.io/<USER or ORGANIZATION>/<REPOSITORY> # e.g. quay.io/werf/image
```

#### Example

The user has the following:

* Two images in _werf.yaml_: **frontend**, **backend**.
* quay.io organization: **quay.io/company**.

There are two ways to store these images:
1. Images repo **quay.io/company** leads to `quay.io/company/frontend:tag` and `quay.io/company/backend:tag` tags. 
    ```shell
    werf build-and-publish -s=:local -i=quay.io/company --tag-custom=tag
    ```
2. Images repo **quay.io/company/app** leads to `quay.io/company/app:frontend-tag` and `quay.io/company/app:backend-tag` tags.
    ```shell
    werf build-and-publish -s=:local -i=quay.io/company/app --tag-custom=tag
    ```
   
## Docker Authorization

werf commands do not perform authorization and use the predefined _docker config_ to work with the Docker registry.
_Docker config_ is a directory with the authorization data for registries and other settings.
By default, werf uses the same _docker config_ as the Docker utility: `~/.docker`.
The Docker config directory can be redefined by setting a `--docker-config` option, `$DOCKER_CONFIG`, or `$WERF_DOCKER_CONFIG` environment variables.
The option and variables are the same as the `docker --config` regular option.

To define the _docker config_, you can use `login` — the regular directive of a Docker client, or, if you are using a CI system, [ci-env command]({{ site.baseurl }}/cli/toolbox/ci_env.html) in werf ([learn more about how to plug werf into CI systems]({{ site.baseurl }}/reference/plugging_into_cicd/overview.html)).

> In the case of several CI jobs running simultaneously, executing `docker login` can lead to failed jobs because of a race condition and conflicting temporary credentials.
One job affects another job by overriding temporary credentials in the _Docker config_.
Therefore, the user should provide an individual _Docker config_ for each job via the `docker --config` or by using the `ci-env` command instead
