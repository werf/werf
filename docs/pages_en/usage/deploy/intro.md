---
title: Introduction
permalink: usage/deploy/intro.html
---

You can easily start using werf for deploying your projects via the existing [Helm](https://helm.sh) charts since they are fully compatible with werf. The configuration has a format similar to that of [Helm charts]({{ "usage/deploy/charts.html" | true_relative_url }}).

werf includes all the existing Helm functionality (the latter is built into werf) as well as some customizations:

 - werf has several configurable modes for tracking the resources being deployed, including processing logs and events;
 - you can integrate images built by werf into Helm chart's [templates]({{ "usage/deploy/templates.html" | true_relative_url }});
 - you can assign arbitrary annotations and labels to all resources being deployed to Kubernetes globally with cli options;
 - werf reads all helm configuration from the git due to [giterminism]({{ "usage/project_configuration/giterminism.html" | true_relative_url }}) which enables really reproducible deploys;
 - also, werf has some more features which will be described further.

With all of these features werf can be considered as an alternative or better helm-client to deploy standard helm-compatible charts.

Briefly, the following commands are used to deal with an application in the Kubernetes cluster:
- [converge]({{ "reference/cli/werf_converge.html" | true_relative_url }}) — to release an application;
- [dismiss]({{ "reference/cli/werf_dismiss.html" | true_relative_url }}) — to delete an application from the cluster.
- [bundle apply]({{ "reference/cli/werf_bundle_apply.html" | true_relative_url }}) — to release an application [bundle]({{ "usage/deploy/bundles.html" | true_relative_url }}).

## Giterminism

werf implements giterminism mode when reading helm configuration. Configuration files or any supplement files should reside in the current commit of the git repository of the project, because werf would read all of these files directly from the git repository. These files include:

 1. All chart configuration files:

    ```
    .helm/
      Chart.yaml
      Chart.lock
      templates/
        <name>.yaml
        <name>.tpl
        <some_dir>/
          <name>.yaml
          <name>.tpl
      charts/
      secret/
      values.yaml
      secret-values.yaml
    ```

 2. Any supplement [user defined values files]({{ "/usage/deploy/values.html#user-defined-regular-values" | true_relative_url }}) passed with `--values`, `--secret-values`, `--set-file` options.

**NOTE** werf has the development mode, allowing working with project files without redundant commits during debugging and development. To enable the mode, specify the `--dev` option (e.g., `werf converge --dev`).

## Subcharts and giterminism

To use subcharts correctly one must add [`.helm/Chart.lock`](https://helm.sh/docs/helm/helm_dependency/) file and commit into the git repository. werf will automatically download all dependencies specified in the lock-file and load subcharts correctly. It is better to add `.helm/charts` directory into the project `.gitignore` in such case.

One may explicitly add chart files into the `.helm/charts/` directory and commit this directory into the repo. In such case werf will download subcharts specified in the `.helm/Chart.lock` and also load `.helm/charts` directory charts and merge resulting charts. Charts defined in the `.helm/charts` will have a higher priority in such case and should redefine duplicates from `.helm/Chart.lock`.

More info about subcharts available in the [separate article]({{ "/usage/deploy/charts.html#dependent-charts" | true_relative_url }}).
