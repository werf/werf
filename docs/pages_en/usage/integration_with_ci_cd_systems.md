---
title: Integration with CI/CD systems
permalink: usage/integration_with_ci_cd_systems.html
---

## Introduction

When working with a CI/CD system, the user must keep its particularities in mind and integrate with the existing primitives and components.

werf provides a ready-made integrations for GitLab CI/CD and GitHub Actions. By leveraging CI jobs' service environment variables, the integration does the following:

- Creating a temporary Docker configuration based on the current user configuration and authorizing in the CI container registry.
- Setting default values for werf commands:
  - Using the CI container registry (`WERF_REPO`).
  - Detecting the current environment (`WERF_ENV`).
  - Annotating the chart resources being deployed, binding to the CI system (`WERF_ADD_ANNOTATION_*`). Annotations are added to all resources being deployed, allowing the user to go to the bound pipeline, job, and commit if necessary.
  - Setting up werf logging (`WERF_LOG_*`).
  - Enabling automatic cleaning of werf processes for cancelled CI jobs (`WERF_ENABLE_PROCESS_EXTERMINATOR=1`). This procedure is only required in CI systems that cannot send termination signals to spawned processes (e.g., GitLab CI/CD).

## GitLab CI/CD

The entire integration boils down to invoking the [ci-env]({{"reference/cli/werf_ci_env.html" | true_relative_url }}) command and then following the instructions the command prints to stdout.

```shell
. $(werf ci-env gitlab --as-file)
```

Then, within a shell session, all werf commands will use the preset values by default and work with the CI container registry.

For example, the pipeline stage to deploy to production might look as follows:

```yaml
Deploy to Production:
  stage: deploy
  script:
    - . $(werf ci-env gitlab --as-file)
    - werf converge
  environment:
    name: production
```

> The complete `.gitlab-ci.yml` file for ready-to-use workflows, as well as the details of using werf with Shell, Docker, and Kubernetes executors, can be found in the corresponding guides:
>
> - [Ready-to-use workflows](/guides/nodejs/400_ci_cd_workflow/030_gitlab_ci_cd/010_workflows.html);
> - [Docker executor](/guides/nodejs/400_ci_cd_workflow/030_gitlab_ci_cd/020_docker_executor.html);
> - [Kubernetes executor](/guides/nodejs/400_ci_cd_workflow/030_gitlab_ci_cd/030_kubernetes_executor.html).

## GitHub Actions

Just like with GitLab CI/CD, the integration boils down to invoking the [ci-env]({{"reference/cli/werf_ci_env.html" | true_relative_url }}) command and then following the instructions the command prints to stdout.


```shell
. $(werf ci-env github --as-file)
```

Then, within a particular step, all werf commands will use the preset values by default and work with the CI container registry.

For example, the job to deploy to production might look as follows:

{% raw %}
```yaml
converge:
  name: Converge
  runs-on: ubuntu-latest
  steps:

    - name: Checkout code
      uses: actions/checkout@v3
      with:
        fetch-depth: 0

    - name: Install werf
      uses: werf/actions/install@v1.2

    - name: Run script
      run: |
        . $(werf ci-env github --as-file)
        werf converge
      env:
        WERF_KUBECONFIG_BASE64: ${{ secrets.KUBE_CONFIG_BASE64_DATA }}
        WERF_ENV: production
```
{% endraw %}

> The complete set of configurations (`.github/workflows/*.yml`) for the ready-to-use workflows can be found [in the corresponding section](/guides/nodejs/400_ci_cd_workflow/040_github_actions.html) of the manual.
