# Importing Configurations from Remote Repositories

**`werf`** allows you to import configuration files and templates from remote Git repositories (includes). This helps avoid code duplication when managing similar applications. Imported files are handled via a virtual file system that follows specific overlay rules and restrictions.

## Configuration

The includes configuration is defined in the `werf-includes.yaml` file located at the root of the project. This file cannot be imported and is processed according to the rules of Giterminism.

Example `werf-includes.yaml`:

```yaml
includes:
  - git: https://github.com/werf/examples
    branch: main
    add: local-dev
    to: /
    includePaths:
      - /.helm

  - git: https://github.com/werf/helm_examples
    tag: v0.0.1
    add: /
    to: examples/helm

  - git: https://github.com/werf/werf_examples
    commit: 831fb9ff936ed6f1db5175c3e65891cfd25580dd
    add: /
    to: /
    includePaths:
      - /.werf
    excludePaths:
      - /.helm
```

## Version Locking

To ensure predictable and reproducible builds and deployments, you need to lock the specific commits of branches or tags used in your includes. This is done using the `werf-includes.lock` file, which can be generated or updated with the following command:

```bash
werf includes update
```

This command records the current `HEAD` commit of each declared source.

Example `werf-includes.lock`:

```yaml
includes:
  - git: https://github.com/werf/examples
    branch: main
    commit: 21640b8e619ba4dd480fedf144f7424aa217a2eb

  - git: https://github.com/werf/helm_examples
    tag: v0.0.1
    commit: e9eff747f82d6959d1c0b4da284e23fd650f4be3

  - git: https://github.com/werf/werf_examples
    commit: 831fb9ff936ed6f1db5175c3e65891cfd25580dd
```

> **Important.** The lock file cannot be imported into other projects and is subject to giterminism policies.

### Using Latest Versions of Sources (Not Recommended)

If you need to work with the latest version of a remote include source (without locking commits), you can enable this behavior in the `werf-giterminism.yaml` file:

```yaml
includes:
  allowIncludesUpdate: true
```

However, this approach doesn't provide reproducibility of builds and deployments and is not recommended for production use.


## Limitations

### What could be imported:

* `.werf` template files
* Helm chart files in the `.helm` directory
* The main configuration file `werf.yaml`
* `Dockerfile` and `.dockerignore`

### What **could not be** imported:

* `werf-includes.yaml` and `werf-includes.lock`
* `werf-giterminism.yaml` configuration file

> **Note:** Includes only work at the configuration level — files from remote repositories **are not** added to the Docker build context.

## Overlay Rules

If multiple sources provide the same file paths, the following rules are applied:

1. If a file exists in the local project directory — **it takes precedence**, even if a file with the same name exists in includes.
2. If multiple includes define the same file — the file from the **first defined repository** in the list will be used.

Example:

```yaml
includes:
  - git: https://github.com/werf/examples
    branch: main
    add: local-dev
    to: /
    includePaths:
      - /.helm

  - git: https://github.com/werf/helm_examples
    branch: main
    to: /
    includePaths:
      - /.helm
```

## Debugging

Let’s look at an example involving overlay behavior between two includes.

```yaml
includes:
  - git: https://github.com/werf/examples
    branch: main
    add: local-dev
    to: /
    includePaths:
      - /.helm

  - git: https://github.com/werf/helm_examples
    branch: main
    to: /
    includePaths:
      - /.helm
```

Directory structures in the source repositories:

**`https://github.com/werf/examples`**

```bash
.helm/
└── Chart.yaml
```

**`https://github.com/werf/helm_examples`**

```bash
.helm/
├── Chart.yaml
└── values.yaml
```

To see a full list of imported files, use the command:

```bash
werf includes ls-files .helm
```

Example output:

```
PATH                                                SOURCE
.helm/Chart.yaml                                    https://github.com/werf/examples
.helm/values.yaml                                   https://github.com/werf/helm_examples
```

As you can see, `.helm/Chart.yaml` was imported from `examples` because it was listed earlier in the includes list.

You can also view the content of any imported file:

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