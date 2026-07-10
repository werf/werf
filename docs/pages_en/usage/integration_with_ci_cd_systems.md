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

## Docker configuration

By default, `werf ci-env` copies the current Docker config directory (from `~/.docker` or the path specified by `--docker-config`/`WERF_DOCKER_CONFIG`) to a temporary directory and exports the `DOCKER_CONFIG` environment variable pointing to it. This isolates CI job's Docker operations from the host configuration.

If the `DOCKER_AUTH_CONFIG` environment variable is set, `werf ci-env` automatically uses it instead: it creates a fresh Docker config from its contents rather than copying an existing one. This is useful when:

- The CI runner doesn't have a persistent Docker config (e.g., ephemeral runners).
- Registry credentials are injected via `DOCKER_AUTH_CONFIG` (common in GitLab CI/CD).
- You want a clean Docker config with no inherited credential helpers or settings.

The `DOCKER_AUTH_CONFIG` value must be a valid JSON string in Docker config format, e.g.:

```json
{"auths": {"registry.example.com": {"auth": "base64-encoded-user:password"}}}
```

You can also control this behavior explicitly with the `--use-docker-auth-config`/`--use-docker-auth-config=false` flag (or `WERF_USE_DOCKER_AUTH_CONFIG`). An explicit value always takes precedence over the auto-detection: e.g., pass `--use-docker-auth-config=false` to force copying the existing Docker config even when `DOCKER_AUTH_CONFIG` is set.

If `--use-docker-auth-config` is explicitly enabled but `DOCKER_AUTH_CONFIG` is not set, `werf ci-env` will exit with an error.

> After creating the temporary config, if `--login-to-registry` is enabled (the default), `werf ci-env` additionally logs in to the CI container registry, adding credentials to the temporary config.

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

If your CI runner provides registry credentials via the `DOCKER_AUTH_CONFIG` environment variable, you can use `--use-docker-auth-config` to create the Docker config from it:

```yaml
Deploy to Production with Docker Auth Config:
  stage: deploy
  script:
    - . $(werf ci-env gitlab --as-file --use-docker-auth-config)
    - werf converge
  environment:
    name: production
```

> You can find the complete `.gitlab-ci.yml` file for ready-to-use workflows, as well as the details of using werf with Shell, Docker, and Kubernetes executors, 
> in the [Getting started](https://werf.io/getting_started/?usage=ci&ci=gitlabCiCd) configurator by selecting _CI/CD_ as your usage scenario and _GitLab CI/CD_ as your CI/CD system, and choosing the CI runner type you need.

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
      uses: werf/actions/install@v2

    - name: Run script
      run: |
        . $(werf ci-env github --as-file)
        werf converge
      env:
        WERF_KUBECONFIG_BASE64: ${{ secrets.KUBE_CONFIG_BASE64_DATA }}
        WERF_ENV: production
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```
{% endraw %}

> You can find the complete set of configurations (`.github/workflows/*.yml`) for the ready-to-use workflows in the [Getting started](https://werf.io/getting_started/?usage=ci&ci=githubActions) configurator
> by selecting _CI/CD_ as your usage scenario and _GitHub Actions_ as your CI/CD system.
