---
title: Giterminism
sidebar: documentation
permalink: documentation/advanced/helm/configuration/giterminism.html
---

Werf implements giterminism mode when reading helm configuration. Configuration files or any supplement files should reside in the current commit of the git repository of the project, because werf would read all of these files directly from the git repository. These files include:

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

 2. Any supplement [user defined values files]({{ "/documentation/advanced/helm/configuration/values.html#user-defined-regular-values" | true_relative_url }}) passed with `--values`, `--secret-values`, `--set-file` options.

**NOTE** Werf has a developer mode, which enables usage of the uncommitted files added to the git index (by the `git add` command). To use developer mode specify `--dev` option for any werf command (such such as `werf converge --dev`).

## Subcharts and giterminism

To use subcharts correctly one must add [`.helm/Chart.lock`](https://helm.sh/docs/helm/helm_dependency/) file and commit into the git repository. Werf will automatically download all dependencies specified in the lock-file and load subcharts correctly. It is better to add `.helm/charts` directory into the project `.gitignore` in such case.

One may explicitly add chart files into the `.helm/charts/` directory and commit this directory into the repo. In such case werf will download subcharts specified in the `.helm/Chart.lock` and also load `.helm/charts` directory charts and merge resulting charts. Charts defined in the `.helm/charts` will have a higher priority in such case and should redefine duplicates from `.helm/Chart.lock`.

More info about subcharts available in the [separate article]({{ "/documentation/advanced/helm/configuration/chart_dependencies.html" | true_relative_url }}).
