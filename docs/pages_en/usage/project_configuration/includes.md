---
title: Includes
permalink: usage/project_configuration/includes.html
---

## Overview

`werf` supports importing configuration files from external Git repositories using the **includes** mechanism. This allows you to reuse templates and shared settings across projects and simplify infrastructure maintenance.  
This is useful when:
* You have many similar applications and want to reuse Helm charts, `.werf` templates, `Dockerfile`, or `.dockerignore`.
* You want to centrally update configurations and templates without duplication.

### What can be imported

The following types of files are supported for import:

* **Configuration files** such as `werf.yaml` and `.werf` templates  
* **Template files** from the `.werf` directory (by default) and individual files inside it (e.g., `.tmpl`, `.yaml`)
* **Helm charts** from the `.helm` directory and individual files inside it
* **Dockerfile** and `.dockerignore`

### Excluded files

The following files cannot be imported:

* `werf-includes.yaml`
* `werf-includes.lock`
* `werf-giterminism.yaml`

### How the includes works

1. You define external sources and the paths to the required files in `werf-includes.yaml`.
2. You lock the versions of the external sources using the `werf includes update` command, which creates the `werf-includes.lock` file. The lock file is necessary for reproducible builds and deployments.
3. All `werf` commands (e.g., `werf build`, `werf converge`) will use the lock file and apply the overlay rules accordingly.

### Overlay rules

If the same file exists in multiple places, the following rules apply:

1. Local project files always take precedence over imported files.
2. If a file exists in multiple includes, the version from the source listed lower in the `includes` list in `werf-includes.yaml` (from general to specific) is used.

Below is an example configuration for importing similar configurations:

```yaml
# werf-includes.yaml
includes:
  - git: https://github.com/werf/examples
    branch: main
    add: local-dev
    to: /
    includePaths:
      - /.helm

  - git: https://example.com/helm_examples
    branch: main
    add: /
    to: /
    includePaths:
      - /.helm
```

Directory structure:

* In `examples`: `.helm/Chart.yaml`, `.helm/values.yaml`
* In `helm_examples`: `.helm/Chart.yaml`

Output of `werf includes ls-files`:

```
.helm/Chart.yaml        https://example.com/werf/helm_examples
.helm/values.yaml       https://github.com/werf/examples
```

The `Chart.yaml` file is taken from the latest source (`helm_examples`) according to the overlay rules.

## Using configurations from external sources

Below is an example `werf-includes.yaml` configuration.
For the full list of available directives, see the corresponding [werf-includes.yaml]({{"reference/werf_includes_yaml.html" | true_relative_url }}) reference page.

```yaml
includes:
  - git: https://github.com/werf/examples
    branch: main
    add: local-dev
    to: /
    includePaths:
      - /.helm
```

After locking the includes with the `werf includes update` command, a `werf-includes.lock` file will be created in the project root, which looks like this:

```yaml
includes:
  - git: https://github.com/werf/examples
    branch: main
    commit: 21640b8e619ba4dd480fedf144f7424aa217a2eb
```

> **IMPORTANT.** According to giterminism policies, both `werf-includes.yaml` and `werf-includes.lock` must be committed to the repository. For initial configuration and debugging, you can use the `--dev` flag. Once configuration is finalized, the files must be committed.

## Updating

### Deterministic version updates

There are two ways to update include versions:

* Edit the `werf-includes.lock` file manually or use dependency management tools like Dependabot, Renovate, etc.
* Use the `werf includes update` command. This will update all includes to the `HEAD` of the specified reference (`branch` or `tag`).

### Automatic version updates (not recommended)

If you need to use the latest `HEAD` versions without a lock file—for example, to quickly test recent changes—you can use `--allow-includes-update` option. The usage of this option must be enabled in `werf-giterminism.yaml`:

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
PATH                      SOURCE
.helm/Chart.yaml          local
.helm/values.yaml         https://github.com/werf/examples
```

### Retrieving file content

```bash
werf includes get-file .helm/values.yaml
```

Example output:

```yaml
backend:
  limits:
    cpu: 100m
    memory: 256Mi
```

### Using local repositories

You can also use local repositories as include sources.
The workflow is the same as with remote repositories.

```yaml
# werf-includes.yaml
includes:
  - git: /path/to/repo
    branch: main
    add: local-dev
    to: /
    includePaths:
      - /.helm
```
