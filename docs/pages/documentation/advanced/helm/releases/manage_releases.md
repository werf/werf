---
title: Manage releases
permalink: documentation/advanced/helm/releases/manage_releases.html
---

There are 2 main commands that will create and delete releases:

 - [`werf converge` command]({{ "documentation/reference/cli/werf_converge.html" | true_relative_url }}) creates a new release version for your project;
 - [`werf dismiss` commnad]({{ "documentation/reference/cli/werf_dismiss.html" | true_relative_url }}) removes all release versions existing for your project.

Werf also provides low-level helm commands to manage releases:

 - [`werf helm list -A` command]({{ "documentation/reference/cli/werf_helm_list.html" | true_relative_url }}) lists all releases of all namespaces available in the cluster;
 - [`werf helm get all RELEASE` command]({{ "documentation/reference/cli/werf_helm_get_all.html" | true_relative_url }}) to get release manifests, hooks or values recorded into the release;
 - [`werf helm status RELEASE` command]({{ "documentation/reference/cli/werf_helm_status.html" | true_relative_url }}) to get status of the latest version of the specified release;
 - [`werf helm history RELEASE` command]({{ "documentation/reference/cli/werf_helm_history.html" | true_relative_url }}) to get list of recorded versions for the specified release.
