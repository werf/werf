---
title: Container registry cleanup
permalink: usage/cleanup/cr_cleanup.html
---

## Overview

The number of images can grow rapidly, taking up more space in the container registry and thus leading to a significant increase in costs. To control the growth and keep it at an acceptable level, werf offers its cleanup approach. It takes into account the images used in Kubernetes as well as their relevance based on the Git history when deciding which images to delete.

The [**werf cleanup**]({{ "reference/cli/werf_cleanup.html" | true_relative_url }}) command is designed to run on a schedule. werf performs the (safe) cleanup according to the cleanup policies.

Most likely, the default cleanup policies will cover all your needs, and no additional configuration will be necessary.

Note that the cleanup does not free up space occupied by images in the container registry. werf only removes tags for irrelevant images (manifests). You will have to run the container registry garbage collector periodically to clean up the associated data.

> The issue of cleaning up images in the container registry and our approach to addressing it are covered in detail in the article [The problem of "smart" cleanup of container images and addressing it in werf](https://www.cncf.io/blog/2020/10/15/overcoming-the-challenges-of-cleaning-up-container-images/)

## Automating the container registry cleanup

Perform the following steps to automate the removal of irrelevant images from the container registry:

- Set [**werf cleanup**]({{ "reference/cli/werf_cleanup.html" | true_relative_url }}) to run periodically to remove the no-longer-relevant tags from the container registry. 
- Set [garbage collector](#container-registrys-garbage-collector) to run on intervals to free up space in the container registry.

## Ignoring images that Kubernetes uses

werf connects to **all Kubernetes clusters** described in **all configuration contexts** of kubectl. It then collects image names for the following object types: `pod`, `deployment`, `replicaset`, `statefulset`, `daemonset`, `job`, `cronjob`, `replicationcontroller`.

The user can configure werf's behavior using the following parameters (and related environment variables):
- `--kube-config`, `--kube-config-base64` set out the kubectl configuration (by default, the user-defined configuration at `~/.kube/config` is used);
- `--kube-context` scans a specific context;
- `--scan-context-namespace-only` scans the namespace linked to a specific context (by default, all namespaces are scanned).

You can disable Kubernetes scanning using the following directive in werf.yaml:

```yaml
cleanup:
  disableKubernetesBasedPolicy: true
```

As long as some object in the Kubernetes cluster uses an image, werf will never delete this image from the container registry. In other words, if you run some object in a Kubernetes cluster, werf will not delete its related images under any circumstances during the cleanup.

## Ignoring freshly built images

When cleaning up, werf ignores images built during a specified time period (the default is 2 hours). If necessary, the period can be adjusted or the policy can be disabled altogether using the following directives in `werf.yaml`:

```yaml
cleanup:
  disableBuiltWithinLastNHoursPolicy: false
  keepImagesBuiltWithinLastNHours: 2
```

## Configuring Git history-based cleanup policies

The cleanup configuration consists of a set of policies called `keepPolicies`. They are used to select relevant images using the git history. Thus, during a cleanup, __images not meeting the criteria of any policy are deleted__.

Each policy consists of two parts:
- `references` defines a set of references, git tags, or git branches to perform scanning on.
- `imagesPerReference` defines the limit on the number of images for each reference contained in the set.

Each policy must be associated with a set of git tags (`tag`) or git branches (`branch`). You can specify a specific reference name or a specific group using [golang regular expression syntax](https://golang.org/pkg/regexp/syntax/#hdr-Syntax).

```yaml
tag: v1.1.1  # or /^v.*$/
branch: main # or /^(main|production)$/
```

> When scanning, werf searches for the provided set of git branches in the origin remote references, but in the configuration, the  `origin/` prefix is omitted in branch names.

You can limit the set of references on the basis of the date when the git tag was created or the activity in the git branch. The `limit` group of parameters allows the user to define flexible and efficient policies for various workflows.

```yaml
- references:
    branch: /^features\/.*/
    limit:
      last: 10
      in: 168h
      operator: And
``` 

In the example above, werf selects no more than 10 latest branches that have the `features/` prefix in the name and have shown any activity during the last week.

- The `last` parameter allows you to select the last `n` references from the set defined in `branch`/`tag`.
- The `in` parameter (see the [documentation](https://golang.org/pkg/time/#ParseDuration) to learn more) allows you to select git tags that were created during the specified period, or git branches with activity within the period. It can also be used for a specific set of `branch` / `tag`.
- The `operator` parameter defines the references resulting from the policy. They may satisfy both conditions or either of them (`And` is set by default).

When scanning references, the number of images is not limited by default. However, you can configure this behavior using the `imagesPerReference` set of parameters:

```yaml
imagesPerReference:
  last: int
  in: string
  operator: string
```

- The `last` parameter specifies the number of images for each reference. By default, there is no limit (`-1`).
- The `in` parameter (refer to the [documentation](https://golang.org/pkg/time/#ParseDuration) to learn more) defines the period for which to search for images.
- The `operator` parameter determines which images will be saved after applying the policy. The images may satisfy both conditions or either of them (`And` is set by default).

> In the case of git tags, werf checks the HEAD commit only; the value of `last`>1 does not make any sense and is invalid

### Policy priority for a specific reference

When describing a group of policies, you have to move from the general to the particular. In other words, `imagesPerReference` for a specific reference will match the latest policy it falls under:

```yaml
- references:
    branch: /.*/
  imagesPerReference:
    last: 1
- references:
    branch: master
  imagesPerReference:
    last: 5
```

In the above example, the _master_ reference matches both policies. Thus, when scanning the branch, the `last` parameter will equal to 5.

### Default policies

If there are no custom cleanup policies defined in `werf.yaml`, werf uses default policies configured as follows:

```yaml
cleanup:
  keepPolicies:
  - references:
      tag: /.*/
      limit:
        last: 10
  - references:
      branch: /.*/
      limit:
        last: 10
        in: 168h
        operator: And
    imagesPerReference:
      last: 2
      in: 168h
      operator: And
  - references:  
      branch: /^(main|master|staging|production)$/
    imagesPerReference:
      last: 10
``` 

Let us examine each policy individually:

1. Keep an image for the last 10 tags (by date of creation).
2. Keep no more than two images published over the past week, for no more than 10 branches active over the past week.
3. Keep the 10 latest images for main, staging, and production branches.

### Disabling policies

If Git history-based cleanup is not needed, you can disable it in `werf.yaml` as follows:

```yaml
cleanup:
  disableGitHistoryBasedPolicy: true
```

## Features of working with different container registries

By default, werf uses the [_Docker Registry API_](https://docs.docker.com/registry/spec/api/) for deleting tags. The user must be authenticated and have a sufficient set of permissions. If the _Docker Registry API_ isn't supported and tags are deleted using the native API, then some additional container registry-specific actions are required on the user's part.

|                             |                             |
|-----------------------------|:---------------------------:|
| _AWS ECR_                   |     [***ok**](#aws-ecr)     |
| _Azure CR_                  |    [***ok**](#azure-cr)     |
| _Default_                   |           **ok**            |
| _Docker Hub_                |   [***ok**](#docker-hub)    |
| _GCR_                       |           **ok**            |
| _GitHub Packages_           | [***ok**](#github-packages) |
| _GitLab Registry_           | [***ok**](#gitlab-registry) |
| _Harbor_                    |           **ok**            |
| _JFrog Artifactory_         |           **ok**            |
| _Nexus_                     |           **ok**            |
| _Quay_                      |           **ok**            |
| _Yandex container registry_ |           **ok**            |
| _Selectel CRaaS_            | [***ok**](#selectel-craas)  |

werf tries to automatically detect the type of container registry using the repository address provided (via the `--repo` option). The user can explicitly specify the container registry using the `--repo-container-registry` option or via the `WERF_REPO_CONTAINER_REGISTRY` environment variable.

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

- `--repo-docker-hub-token` or
- `--repo-docker-hub-username` and `--repo-docker-hub-password`.

### GitHub Packages

When organizing CI/CD pipelines in GitHub Actions, we recommend using [our set of actions](https://github.com/werf/actions) to solve most of the challenges for you.

werf uses the _GitHub API_ to delete tags, so you need to set the _token_ with the appropriate scopes (`read:packages`, `delete:packages`) to clean up the container registry.

You can use the `--repo-github-token` option or the corresponding environment variable to define the token.

### GitLab Registry

werf uses the _GitLab container registry API_ or _Docker Registry API_ (depending on the GitLab version) to delete tags.

> Privileges of the temporary CI job token (`$CI_JOB_TOKEN`) are not enough to delete tags. That is why the user have to create a dedicated token in the Access Token section (select the `api` in the Scope section) and perform authorization using it

### Selectel CRaaS

werf uses the [_Selectel CR API_](https://developers.selectel.ru/docs/selectel-cloud-platform/craas_api/) to delete tags, so you need to set either the _username/password_, _account_ and _vpc_ or _vpcID_ to clean up the container registry.

You can use the following options (or their respective environment variables) to set the said parameters:
- `--repo-selectel-username`
- `--repo-selectel-password`
- `--repo-selectel-account`
- `--repo-selectel-vpc` or
- `--repo-selectel-vpc-id`

#### Known limitations

* Sometimes, Selectel drop connection with VPC ID. Try to use the VPC name instead.
* CR API has limitations, so you can not remove layers from the root of the container registry.
* Low API rate limit. It can cause cleanup problems with rapid development.

## Container registry’s garbage collector

Note that during the cleanup, werf only removes tags from the images (manifests) to be deleted. The container registry garbage collector (GC) is responsible for the actual deletion.

While the garbage collector is running, the container registry must be set to read-only mode or turned off completely. Otherwise, there is a good chance that the garbage collector will not respect the images published during the procedure and may corrupt them.

You can read more about the garbage collector and how to use it in the documentation for the garbage collector you are using. Here are links for some of the most popular ones:

- [Docker Registry GC](https://docs.docker.com/registry/garbage-collection/#more-details-about-garbage-collection ).
- [GitLab CR GC](https://docs.gitlab.com/ee/administration/packages/container_registry.html#container-registry-garbage-collection).
- [Harbor GC](https://goharbor.io/docs/2.6.0/administration/garbage-collection/).
