---
title: Supported container registries
permalink: advanced/supported_container_registries.html
author: Alexey Igrychev <alexey.igrychev@flant.com>
---

By default, werf uses the [_Docker Registry API_](https://docs.docker.com/registry/spec/api/). For most container registry implementations, no other actions are required from the user.

However, the container registry may have limited support for the _Docker Registry API_ and/or implement some functions in the native API. In this case, when using the container registry, the user may need to perform some additional actions and provide the supplementary data.

werf tries to automatically detect the type of container registry using the repository address provided (via the `--repo` option). The user can explicitly specify the container registry using the `--repo-container-registry` option or via the `WERF_REPO_CONTAINER_REGISTRY` environment variable.

|                                           | Build  | Bundles           | Cleanup                                             |
| -------------------------------------     | :----: | :---------------: | :-------------------------------------------------: |
| _AWS ECR_                                 | **ok** |         **ok**    |       [***ok**](#aws-ecr)                           |
| _Azure CR_                                | **ok** |         **ok**    |       [***ok**](#azure-cr)                          |
| _Default_                                 | **ok** |         **ok**    |         **ok**                                      |
| _Docker Hub_                              | **ok** | **not supported** |       [***ok**](#docker-hub)                        |
| _GCR_                                     | **ok** |         **ok**    |         **ok**                                      |
| _GitHub Packages_ (docker.pkg.github.com) | **ok** | **not supported** |      [***ok**](#github-packages-dockerpkggithubcom) |
| _GitHub Packages_ (ghcr.io)               | **ok** |         **ok**    |   **not supported**                                 |
| _GitLab Registry_                         | **ok** |         **ok**    |       [***ok**](#gitlab-registry)                   |
| _Harbor_                                  | **ok** |         **ok**    |         **ok**                                      |
| _JFrog Artifactory_                       | **ok** |         **ok**    |         **ok**                                      |
| _Nexus_                                   | **ok** |   **not tested**  |         **ok**                                      |
| _Quay_                                    | **ok** | **not supported** |         **ok**                                      |
| _Yandex Container Registry_               | **ok** |   **not tested**  |         **ok**                                      |

## Authorization

When interacting with the container registry, werf commands use the information available in the Docker configuration. By default, they use the `~/.docker` directory or an alternative path set by the `WERF_DOCKER_CONFIG` (or `DOCKER_CONFIG`) environment variable.

> The _Docker configuration_ is a directory where authorization data (and Docker settings) are stored (these data are used for accessing various container registries)

For authorization, you can use `docker login` / `oras login` commands as well as solutions provided by container registries.

For CI jobs, we recommend using the [werf ci-env]({{ "/reference/cli/werf_ci_env.html" | true_relative_url }}) command. It creates a temporary _Docker configuration_ based on a user configuration and performs authorization (if the container registry's CI is used) using CI environment variables. See the [relevant section]({{ "/internals/how_ci_cd_integration_works/general_overview.html" | true_relative_url }}) of the documentation for more information about integrating werf with CI systems.

> The usage of a shared _Docker configuration_ when running parallel jobs in a CI system may lead to a job failure because of a race condition and conflicting temporary credentials (one job interferes with another by overriding temporary credentials in the shared _Docker configuration_). That is why we recommend creating a dedicated _Docker configuration_ for each CI job (the `werf ci-env` command does this by default)

## Bundles

To use [bundles]({{ "advanced/bundles.html" | true_relative_url }}), the container registry must support the [OCI Image Format Specification](https://github.com/opencontainers/image-spec).

## Clean up

By default, werf uses the _Docker Registry API_ for deleting tags. The user must be authenticated and have a sufficient set of permissions.

If the _Docker Registry API_ isn't supported and tags are deleted using the native API, then some additional container registry-specific actions are required on the user's part.

### AWS ECR

werf deletes tags using the _AWS SDK_. Therefore, before performing a cleanup, the user must do **one of** the following:

- [Install and configure the AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html#cli-quick-configuration) (`aws configure`), or
- Set `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` environment variables.

### Azure CR

werf deletes tags using the _Azure CLI_. Therefore, before performing a cleanup, the user must do the following:

- Install the [Azure CLI](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli?view=azure-cli-latest) (`az`).
- Perform authorization (`az login`).

> The user must be assigned to one of the following roles: `Owner`, `Contributor`, or `AcrDelete` (learn more about [Azure CR roles and permissions](https://docs.microsoft.com/en-us/azure/container-registry/container-registry-roles))

### Docker Hub

werf uses the _Docker Hub API_ to delete tags, so you need to set either the _token_ or the _username/password_ pair to clean up the container registry.

You can use the following script to get a _token_:

```shell
HUB_USERNAME=username
HUB_PASSWORD=password
HUB_TOKEN=$(curl -s -H "Content-Type: application/json" -X POST -d '{"username": "'${HUB_USERNAME}'", "password": "'${HUB_PASSWORD}'"}' https://hub.docker.com/v2/users/login/ | jq -r .token)
```

> You can't use the [personal access token](https://docs.docker.com/docker-hub/access-tokens/) as a _token_ since the deletion of resources is only possible using the user's primary credentials.

You can use the following options (or their respective environment variables) to set the said parameters:

- `-- repo-docker-hub-token` or
- `--repo-docker-hub-username` and `--repo-docker-hub-password`.

### GitHub Packages (docker.pkg.github.com)

werf uses the _GitHub GraphQL API_ to delete tags, so you need to set the _token_ with the appropriate scope (`read:packages`, `write:packages`, `delete:packages`) along with the `repo` scope to clean up the container registry.

> GitHub [does not support](https://help.github.com/en/packages/publishing-and-managing-packages/deleting-a-package) deleting versions of a package in public repositories

You can use the `--repo-github-tack` option or the corresponding environment variable to define the credentials.

### GitLab Registry

werf uses the _GitLab Container Registry API_ or _Docker Registry API_ (depending on the GitLab version) to delete tags.

> Privileges of the temporary CI job token (`$CI_JOB_TOKEN`) are not enough to delete tags. That is why the user have to create a dedicated token in the Access Token section (select the `api` in the Scope section) and perform authorization using it
