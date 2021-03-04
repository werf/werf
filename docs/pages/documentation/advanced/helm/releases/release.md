---
title: Release
permalink: documentation/advanced/helm/releases/release.html
---

While the [chart]({{ "/documentation/advanced/helm/configuration/chart.html" | true_relative_url }}) is a collection of configuration files of your application, the **release** is a runtime object that represents a running instance of your application deployed via werf.

Each release has a single name and multiple versions. A new release version is created with each invocation of werf converge. For each release only latest version is the actual version, previous versions of the release are stored for history.

## Storing releases

Each release version is stored in the Kubernetes cluster itself. werf supports storing release data in Secret and ConfigMap objects in the arbitrary namespace.

By default, werf stores releases in the Secrets in the namespace of your deployed application. This is fully compatible with the default [Helm 3](https://helm.sh) configuration.

The [`werf helm list -A`]({{ "documentation/reference/cli/werf_helm_list.html" | true_relative_url }}) command lists releases created by werf. Also, the user can fetch the history of a specific release with the [`werf helm history`]({{ "documentation/reference/cli/werf_helm_history.html" | true_relative_url }}) command.

### Helm compatibility notice

werf is fully compatible with the existing Helm 3 installations since the release data is stored similarly to Helm.

Furthermore, you can use werf and Helm 3 simultaneously in the same cluster and at the same time.

You can inspect and browse releases created by werf using commands such as `helm list` and `helm get`. werf can also upgrade existing releases created by the Helm.

### Compatibility with Helm 2

Existing helm 2 releases could be converted to helm 3 with the [`werf helm migrate2to3` command]({{ "/documentation/reference/cli/werf_helm_migrate2to3.html" | true_relative_url }}).

Also [werf converge command]({{ "/documentation/reference/cli/werf_converge.html" | true_relative_url }}) detects existing helm 2 release for your project and converts it to the helm 3 automatically. Existing helm 2 release could exist in the case when your project has previously been deployed by the werf v1.1.
