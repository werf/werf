---
title: Includes
permalink: usage/project_configuration/includes.html
---

## Overview

`werf` supports importing configuration files from external sources using the **includes** mechanism.

This is especially useful when you have many similar applications and want to manage and reuse configuration in a centralized way.

### What can be imported

The following types of files are supported for import:

* The main configuration file (`werf.yaml`), templates (`.werf/**/*.tmpl`), as well as arbitrary files used during templating.
* Helm charts and individual files.
* `Dockerfile` and `.dockerignore`.

### What cannot be imported

The mechanism does not allow importing files for the build context, as well as the following configuration files:

* `werf-includes.yaml`
* `werf-includes.lock`
* `werf-giterminism.yaml`

### How the includes works

1. You define external sources and the files to import in `werf-includes.yaml`.
2. You lock the versions of the external sources using the `werf includes update` command, which creates the `werf-includes.lock` file. The lock file is required for reproducible builds and deployments.
3. All `werf` commands (e.g., `werf build`, `werf converge`) will respect the lock file and apply the configuration according to the overlay rules.

### Overlay rules

If the same file exists in multiple places, the following rules apply:

1. Local project files always take precedence over imported files.
2. If a file exists in multiple includes, the version from the source listed lower in the `includes` list in `werf-includes.yaml` (from general to specific) is used.

#### Example of applying overlay rules

For instance, suppose the local repository contains file `c`, and the following `includes` configuration is used:

```yaml
# werf-includes.yaml
includes:
  - git: repo1 # imports a, b, c
    branch: main
    add: /
    to: /
  - git: repo2 # imports b, c
    branch: main
    add: /
    to: /
```

Then the resulting merged file structure will look like this:

```bash
.
├── a      # from repo1
├── b      # from repo2
├── c      # from local repo
```

## Connecting configuration from external sources

Below is an example configuration of external sources (a full description of directives can be found on the corresponding page [werf-includes.yaml]({{"reference/werf_includes_yaml.html" | true_relative_url }})):

```yaml
# werf-includes.yaml
includes:
  - git: https://github.com/werf/werf
    branch: main
    add: /docs/examples/includes/common
    to: /
    includePaths:
      - .werf
```

The procedure for working with private repositories is the same as described [here]({{ "/usage/build/stapel/git.html#working-with-remote-repositories" | true_relative_url }}).

After the configuration, it is necessary to lock the versions of external dependencies. You can do this using the command `werf includes update`, which will create the file `werf-includes.lock` in the project root:

```yaml
includes:
  - git: https://github.com/werf/werf
    branch: main
    commit: 21640b8e619ba4dd480fedf144f7424aa217a2eb
```

> **IMPORTANT.** According to giterminism policies, the files `werf-includes.yaml` and `werf-includes.lock` must be committed. During configuration and debugging, for convenience, it is recommended to use the `--dev` flag.

### Example of using external sources for configuring similar applications

Suppose you decided to centrally manage the configuration of applications in your organization, keeping common configuration in one source and per-project-type configuration in others (in one or several Git repositories — at your discretion).

The configuration of a typical project in this case might look like this:

* Project:
{% tree_file_viewer 'examples/includes/app' default_file='werf-includes.yaml' %}

* Source with common configuration:
{% tree_file_viewer 'examples/includes/werf-common' default_file='.werf/cleanup.tpl' %}

* Source with configuration specific to your project type:
{% tree_file_viewer 'examples/includes/common_js' default_file='backend.Dockerfile' %}


## Updating includes

### Deterministic

There are two ways to update include versions:

* Use the `werf includes update` command. This will update all includes to the `HEAD` of the specified reference (`branch` or `tag`).
* Edit the `werf-includes.lock` file manually or use dependency management tools like Dependabot, Renovate, etc.

### Automatic (not recommended)

If you need to use the latest `HEAD` versions without a lock file you can use `--allow-includes-update` option. The usage of this option must be enabled in `werf-giterminism.yaml`:

```yaml
includes:
  allowIncludesUpdate: true
```

> **IMPORTANT.** We do not recommend using this approach, as it may break reproducibility of builds and deployments.

## Debugging

The includes mechanism has built-in commands for debugging:

### Listing project files

```bash
werf includes ls-files .helm
```

Example output:

```
PATH                                             SOURCE
.helm/Chart.yaml                                 https://github.com/werf/werf
.helm/charts/backend/Chart.yaml                  https://github.com/werf/werf
.helm/charts/backend/templates/deployment.yaml   https://github.com/werf/werf
.helm/charts/backend/templates/service.yaml      https://github.com/werf/werf
.helm/charts/frontend/Chart.yaml                 https://github.com/werf/werf
.helm/charts/frontend/templates/deployment.yaml  https://github.com/werf/werf
.helm/charts/frontend/templates/service.yaml     https://github.com/werf/werf
.helm/requirements.lock                          https://github.com/werf/werf
.helm/values.yaml                                local
```

### Retrieving file content

```bash
werf includes get-file .helm/charts/backend/templates/deployment.yaml
```

Example output:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ $.Chart.Name }}
  annotations:
    werf.io/weight: "30"
spec:
  selector:
    matchLabels:
      app: {{ $.Chart.Name }}
  template:
    metadata:
      labels:
        app: {{ $.Chart.Name }}
    spec:
      containers:
        - name: backend
          image: {{ $.Values.werf.image.backend }}
          ports:
            - containerPort: 3000
          resources:
            requests:
              cpu: {{ $.Values.limits.cpu }}
              memory: {{ $.Values.limits.memory}}
            limits:
              cpu: {{ $.Values.limits.cpu }}
              memory: {{ $.Values.limits.memory }}
```

### Using local repositories

You can also use local repositories as include sources.
The workflow is the same as with remote repositories.

```yaml
# werf-includes.yaml
includes:
  - git: /local/repo
    branch: main
    add: local-dev
    to: /
    includePaths:
      - /.helm
```
