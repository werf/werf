---
title: Container registry cleanup
permalink: usage/cleanup/cr_cleanup.html
---

The [**werf cleanup**]({{ "reference/cli/werf_cleanup.html" | true_relative_url }}) command is designed to run on a schedule. werf performs the (safe) cleanup according to the specified cleanup policies.

The algorithm automatically selects the images to delete. It consists of the following steps:

- Pulling the necessary data from the container registry.
- Preparing a list of images to keep. werf leaves intact:
  - [Images that Kubernetes uses](#images-in-kubernetes);
  - Images that meet the criteria of the [user-defined policies](#user-defined-policies) when [scanning the Git history](#scanning-the-git-history);
  - New images that were built within the predefined time frame (you can set it via the `--keep-stages-built-within-last-n-hours` option; it is set to 2 hours by default);
  - Images related to the images selected in previous steps.
- Deleting all the remaining images.

#### Images in Kubernetes

werf connects to **all Kubernetes clusters** described in **all configuration contexts** of kubectl. It then collects image names for the following object types: `pod`, `deployment`, `replicaset`, `statefulset`, `daemonset`, `job`, `cronjob`, `replicationcontroller`.

The user can configure werf's behavior using the following parameters (and related environment variables):
- `--kube-config`, `--kube-config-base64` set out the kubectl configuration (by default, the user-defined configuration at `~/.kube/config` is used);
- `--kube-context` scans a specific context;
- `--scan-context-namespace-only` scans the namespace linked to a specific context (by default, all namespaces are scanned).
- `--without-kube` disables Kubernetes scanning.

As long as some object in the Kubernetes cluster uses an image, werf will never delete this image from the container registry. In other words, if you run some object in a Kubernetes cluster, werf will not delete its related images under any circumstances during the cleanup.

#### Scanning the git history

werf's cleanup algorithm uses the fact that the container registry keeps the information about the commits on which the build is based (it does not matter if an image was added to the container registry or some changes were made to it). For each build, werf saves the information about the commit, [stage digest]({{ "usage/build/stages_and_storage.html#stage-digest" | true_relative_url }}), and the image name to the registry (for each `image` defined in `werf.yaml`).

Once the new commit triggers the build process, werf adds the information that the [stage digest]({{ "usage/build/stages_and_storage.html#stage-digest" | true_relative_url }}) corresponds to a specific commit to the registry (even if the _resulting image_ does not change).

This approach ensures the unbreakable bond between the [stage digest]({{ "usage/build/stages_and_storage.html#stage-digest" | true_relative_url}}) and the git history. Also, it makes it possible to effectively clean up outdated images based on the git state and [selected policies](#user-defined-policies). The algorithm scans the git history and selects relevant images.

Information about commits is the only source of truth for the algorithm, so images lacking such information will be deleted.

#### User-defined policies

The user can specify images that will not be deleted during a cleanup using the so-called `keepPolicies` [cleanup policies]({{ "usage/cleanup/cleanup.html" | true_relative_url }}). If there is no configuration provided in the `werf.yaml`, werf will use the [default policy set]({{ "reference/werf_yaml.html#default-policies" | true_relative_url }}).

It is worth noting that the algorithm scans the local state of the git repository. Therefore, it is essential to keep all git branches and git tags up-to-date. By default, werf performs synchronization automatically (you can change its behavior using the [gitWorktree.allowFetchOriginBranchesAndTags]({{ "reference/werf_yaml.html#git-worktree" | true_relative_url }}) directive in `werf.yaml`).

#### Aspects of cleaning up the images that are being built

During the cleanup, werf applies user-defined policies to the set of images for each `image` defined in `werf.yaml`. The cleanup must respect all the `images` in use. On the other hand, the set of images based on the Git repository's main branch may not cover all the suitable images (for example, `images` may be added to/deleted from some feature branch).

werf adds the name of the image being built to the container registry to avoid deleting images and stages in use. Such an approach frees werf from the strict set of image names defined in `werf.yaml` and forces it to take into account all the images ever used.

The `werf managed-images ls|add|rm` family of commands allows the user to edit the so-called _managed images_ set and explicitly delete images that are no longer needed and can be removed entirely.
