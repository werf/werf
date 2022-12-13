---
title: Container registry cleanup
permalink: usage/cleanup/cr_cleanup.html
---

The [**werf cleanup**]({{ "reference/cli/werf_cleanup.html" | true_relative_url }}) command is designed to run on a schedule. werf performs the (safe) cleanup according to the specified cleanup policies.

## Principle of operation

The algorithm automatically selects the images to delete. It consists of the following steps:

- Pulling the necessary data from the container registry.
- Preparing a list of images to keep. werf leaves intact:
  - [Images that Kubernetes uses](#images-in-kubernetes);
  - Images that meet the criteria of the user-defined policies when [scanning the Git history](#scanning-the-git-history);
  - New images that were built within the predefined time frame;
  - Images related to the images selected in previous steps.
- Deleting all the remaining images.

### Images in Kubernetes

werf connects to **all Kubernetes clusters** described in **all configuration contexts** of kubectl. It then collects image names for the following object types: `pod`, `deployment`, `replicaset`, `statefulset`, `daemonset`, `job`, `cronjob`, `replicationcontroller`.

The user can configure werf's behavior using the following parameters (and related environment variables):
- `--kube-config`, `--kube-config-base64` set out the kubectl configuration (by default, the user-defined configuration at `~/.kube/config` is used);
- `--kube-context` scans a specific context;
- `--scan-context-namespace-only` scans the namespace linked to a specific context (by default, all namespaces are scanned).
- `--without-kube` disables Kubernetes scanning.

As long as some object in the Kubernetes cluster uses an image, werf will never delete this image from the container registry. In other words, if you run some object in a Kubernetes cluster, werf will not delete its related images under any circumstances during the cleanup.

### Scanning the git history

werf's cleanup algorithm uses the fact that the container registry keeps the information about the commits on which the build is based (it does not matter if an image was added to the container registry or some changes were made to it). For each build, werf saves the information about the commit, [stage digest]({{ "usage/build/stages_and_storage.html#stage-digest" | true_relative_url }}), and the image name to the registry (for each `image` defined in `werf.yaml`).

Once the new commit triggers the build process, werf adds the information that the [stage digest]({{ "usage/build/stages_and_storage.html#stage-digest" | true_relative_url }}) corresponds to a specific commit to the registry (even if the _resulting image_ does not change).

This approach ensures the unbreakable bond between the [stage digest]({{ "usage/build/stages_and_storage.html#stage-digest" | true_relative_url}}) and the git history. Also, it makes it possible to effectively clean up outdated images based on the git state and selected policies. The algorithm scans the git history and selects relevant images.

Information about commits is the only source of truth for the algorithm, so images lacking such information will be deleted.

It is worth noting that the algorithm scans the local state of the git repository. Therefore, it is essential to keep all git branches and git tags up-to-date. By default, werf performs synchronization automatically (you can change its behavior using the [gitWorktree.allowFetchOriginBranchesAndTags]({{ "reference/werf_yaml.html#git-worktree" | true_relative_url }}) directive in `werf.yaml`).

### Aspects of cleaning up the images that are being built

During the cleanup, werf applies user-defined policies to the set of images for each `image` defined in `werf.yaml`. The cleanup must respect all the `images` in use. On the other hand, the set of images based on the Git repository's main branch may not cover all the suitable images (for example, `images` may be added to/deleted from some feature branch).

werf adds the name of the image being built to the container registry to avoid deleting images and stages in use. Such an approach frees werf from the strict set of image names defined in `werf.yaml` and forces it to take into account all the images ever used.

The `werf managed-images ls|add|rm` family of commands allows the user to edit the so-called _managed images_ set and explicitly delete images that are no longer needed and can be removed entirely.

## Configuring cleanup policies

The cleanup configuration consists of a set of policies called `keepPolicies`. They are used to select relevant images using the git history. Thus, during a cleanup, __images not meeting the criteria of any policy are deleted__.

Each policy consists of two parts:

- `references` defines a set of references, git tags, or git branches to perform scanning on.
- `imagesPerReference` defines the limit on the number of images for each reference contained in the set.

Each policy should be linked to some set of git tags (`tag: string || /REGEXP/`) or git branches (`branch: string || /REGEXP/`). You can specify the name/group of a reference using the [Golang's regular expression syntax](https://golang.org/pkg/regexp/syntax/#hdr-Syntax).

```yaml
tag: v1.1.1
tag: /^v.*$/
branch: main
branch: /^(main|production)$/
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

- The `last: int` parameter allows you to select n last references from those defined in the `branch` / `tag`.
- The `in: duration string` parameter (you can learn more about the syntax in the [docs](https://golang.org/pkg/time/#ParseDuration)) allows you to select git tags that were created during the specified period or git branches that were active during the period. You can also do that for the specific set of `branches` / `tags`.
- The `operator: And || Or` parameter defines if references should satisfy both conditions or either of them (`And` is set by default).

When scanning references, the number of images is not limited by default. However, you can configure this behavior using the `imagesPerReference` set of parameters:

```yaml
imagesPerReference:
  last: int
  in: duration string
  operator: And || Or
```

- The `last: int` parameter defines the number of images to search for each reference. Their amount is unlimited by default (`-1`).
- The `in: duration string` parameter (you can learn more about the syntax in the [docs](https://golang.org/pkg/time/#ParseDuration)) defines the time frame in which werf searches for images.
- The `operator: And || Or` parameter defines what images will stay after applying the policy: those that satisfy both conditions or either of them (`And` is set by default).

> In the case of git tags, werf checks the HEAD commit only; the value of `last`>1 does not make any sense and is invalid

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
| _Yandex Container Registry_ |           **ok**            |
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

werf uses the _GitLab Container Registry API_ or _Docker Registry API_ (depending on the GitLab version) to delete tags.

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
