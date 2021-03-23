---
title: Giterminism
permalink: advanced/helm/configuration/giterminism.html
---

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

 2. Any supplement [user defined values files]({{ "/advanced/helm/configuration/values.html#user-defined-regular-values" | true_relative_url }}) passed with `--values`, `--secret-values`, `--set-file` options.

**NOTE** werf has the development mode, allowing working with project files without redundant commits during debugging and development. To enable the mode, specify the `--dev` option (e.g., `werf converge --dev`).

## Subcharts and giterminism

To use subcharts correctly one must add [`.helm/Chart.lock`](https://helm.sh/docs/helm/helm_dependency/) file and commit into the git repository. werf will automatically download all dependencies specified in the lock-file and load subcharts correctly. It is better to add `.helm/charts` directory into the project `.gitignore` in such case.

One may explicitly add chart files into the `.helm/charts/` directory and commit this directory into the repo. In such case werf will download subcharts specified in the `.helm/Chart.lock` and also load `.helm/charts` directory charts and merge resulting charts. Charts defined in the `.helm/charts` will have a higher priority in such case and should redefine duplicates from `.helm/Chart.lock`.

More info about subcharts available in the [separate article]({{ "/advanced/helm/configuration/chart_dependencies.html" | true_relative_url }}).
