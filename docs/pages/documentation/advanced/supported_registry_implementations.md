---
title: Supported registry implementations
sidebar: documentation
permalink: documentation/advanced/supported_registry_implementations.html
author: Alexey Igrychev <alexey.igrychev@flant.com>
---

There are several types of commands that are working with the Docker registries and require the appropriate authorization:

* During the building process, werf may pull base images from the Docker registry and pull/push _stages_ in distributed builds.
* [During the cleaning process]({{ "documentation/advanced/cleanup.html" | true_relative_url }}), werf deletes _images_ and _stages_ from the Docker registry.
* [During the deploying process]({{ "/documentation/advanced/helm/deploy_process/steps.html" | true_relative_url }}), werf requires access to the _images_ from the Docker registry and to the _stages_ that could also be stored in the Docker registry.

**NOTE** You should specify your implementation by `--repo-implementation` CLI option.

## Supported implementations

|                 	                    | Build              	    | Cleanup                         	                                    |
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
| _Quay_                    	        |         **ok**        	|                            **ok**                            	        |

> In the nearest future we will add support for Nexus Registry

The following implementations are fully supported and do not require additional actions except [docker authorization](#docker-authorization):
* _Default_.
* _GCR_.
* _GitLab Registry_.
* _Harbor_.

_Azure CR_, _AWS ECR_, _Docker Hub_ and _GitHub Packages_ implementations provide Docker Registry API but do not implement the delete tag method and offer it with native API. 
Therefore, werf may require extra credentials for [cleanup commands]({{ "documentation/advanced/cleanup.html" | true_relative_url }}).

## AWS ECR

Working with _AWS ECR_ is not different from the rest implementations. 
The user can use both _images repo modes_ but should manually create repositories before using werf.

werf deletes tags from _AWS ECR_ with _AWS SDK_. 
Therefore, before using [cleanup commands]({{ "documentation/advanced/cleanup.html" | true_relative_url }}) the user should:
* [Set up _AWS CLI_ installation](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html#cli-quick-configuration) (`aws configure`) or 
* Define `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` environment variables.
      
## Azure CR

The user should do the next steps to enable tag deletion with werf: 
* Install [Azure CLI](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli?view=azure-cli-latest) (`az`).
* Sign in with relevant credentials (`az login`).

> The user must have one of the following roles: `Owner`, `Contributor` or `AcrDelete` (more about [Azure CR roles and permissions](https://docs.microsoft.com/en-us/azure/container-registry/container-registry-roles)) 

## Docker Hub

To delete tags from _Docker Hub_ repository werf uses _Docker Hub API_ and requires extra user credentials.

The user should specify token or username and password. The following script can be used to get the token:

```shell
HUB_USERNAME=USERNAME
HUB_PASSWORD=PASSWORD
HUB_TOKEN=$(curl -s -H "Content-Type: application/json" -X POST -d '{"username": "'${HUB_USERNAME}'", "password": "'${HUB_PASSWORD}'"}' https://hub.docker.com/v2/users/login/ | jq -r .token)
```

> Be aware that access to the resources is forbidden with a [personal access token](https://docs.docker.com/docker-hub/access-tokens/)

To define credentials check options and related environments:
* `--repo-docker-hub-token` or
* `--repo-docker-hub-username` and `--repo-docker-hub-password`.

## GitHub Packages

To [delete versions of a private package](https://help.github.com/en/packages/publishing-and-managing-packages/deleting-a-package) we are using GraphQL and need GitHub token with `read:packages`, `write:packages`, `delete:packages` and `repo` scopes.

> Be aware that GitHub only supports deleting in private repositories 

To define credentials check `--repo-github-token` option and related environment.

## Docker Authorization

werf commands do not perform authorization and use the predefined _docker config_ to work with the Docker registry.
_Docker config_ is a directory with the authorization data for registries and other settings.
By default, werf uses the same _docker config_ as the Docker utility: `~/.docker`.
The Docker config directory can be redefined by setting a `--docker-config` option, `$DOCKER_CONFIG`, or `$WERF_DOCKER_CONFIG` environment variables.
The option and variables are the same as the `docker --config` regular option.

To define the _docker config_, you can use `login` — the regular directive of a Docker client, or, if you are using a CI system, [ci-env command]({{ "documentation/reference/cli/werf_ci_env.html" | true_relative_url }}) in werf ([learn more about how to plug werf into CI systems]({{ "documentation/internals/how_ci_cd_integration_works/general_overview.html" | true_relative_url }})).

> In the case of several CI jobs running simultaneously, executing `docker login` can lead to failed jobs because of a race condition and conflicting temporary credentials.
One job affects another job by overriding temporary credentials in the _Docker config_.
Therefore, the user should provide an individual _Docker config_ for each job via the `docker --config` or by using the `ci-env` command instead
