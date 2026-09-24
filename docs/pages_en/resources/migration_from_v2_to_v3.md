---
title: Migration from v2 to v3
permalink: resources/migration_from_v2_to_v3.html
---

First check compatibility with v2, then the changes to building, deployment and registry cleanup. Removed features require changes before upgrading; deprecated keys still work with a warning.

## Compatibility with v2

Running both versions or rolling back to v2 depends on shared configuration, secrets and registry state, not just the binaries:

- **Configuration.** After adopting v3 settings, do not assume the same `werf.yaml` will work in v2. Keep a v2-compatible configuration revision for rollback — see [building changes](#building).
- **Secrets.** v3 reads old secrets, but an old client cannot read new v3 writes. Upgrade all readers before the first rewrite; switching the binary back to v2 does not restore the old format — see [encrypted secrets](#encrypted-secret-values).
- **Registry.** A shared `--repo` means shared images, even with a separate `--meta-repo`. Stop old jobs before moving metadata, and protect rollback images from cleanup. Do not resume v2 cleanup against that repository after the move — see [registry cleanup](#registry-cleanup).

## Building

**Expect a full rebuild of images after moving to v3. Account for it when planning the upgrade.**

### Images instead of artifacts

The `artifact` directive is removed. Replace it with `image` and `final: false`:

<table>
<thead><tr><th scope="col">Before — v2</th><th scope="col">After — v3</th></tr></thead>
<tbody><tr>
<td markdown="1">

```yaml
artifact: builder
from: ubuntu:22.04
```

</td>
<td markdown="1">

```yaml
image: builder
from: ubuntu:22.04
final: false
```

</td>
</tr></tbody>
</table>

### Image names

Nameless stapel images (`image: ~`) are no longer supported — give every image a name.

Names are now checked when loading the configuration. Latin letters, digits, `_`, `.`, `-` and `+` are allowed, with optional `/`-separated segments. Each segment must start with a letter or digit and end with a letter, digit or `+`. `modules/controller` is valid; an empty name, whitespace, `/api`, `api-` and `modules//controller` cause a configuration error. Check names produced by Go templates as well.

### One `from` key for image references

Base images and imports use `from` for both internal and external images. The reference itself determines which kind it is — a separate directive for internal images is no longer needed:

- `from: base` refers to the image named `base` in `werf.yaml`.
- `from: ubuntu:24.04` or a reference with `@sha256:...` refers to an external image.

Replace the old keys, including references in `dependencies`:

| Where | Before — v2 | After — v3 |
|---|---|---|
| Stapel base image | `fromImage: base` | `from: base` |
| An `import` entry | `image: builder` | `from: builder` |
| A `dependencies` entry | `image: backend` | `from: backend` |

All three old keys **still work**, but emit a deprecation warning. Specifying both the old and new key is an error.

An **external reference** in a base `from:` or `import.from` now requires an explicit tag or digest. For example, the previously implicit `:latest` must be written explicitly:

| Before — v2 | After — v3 |
|---|---|
| `from: ubuntu` | `from: ubuntu:latest` |

Instead of `:latest`, you can specify the required tag (`:TAG`) or digest (`@sha256:...`). Internal image names from `werf.yaml` do not need a tag.

### Builders and image configuration

**The Ansible builder is removed.** Rewrite `ansible:` steps using the Shell builder; renaming the key is not enough.

**The `docker:` directive is removed.** Move its settings to `imageSpec.config`, translating field names and formats. For example, for a stapel image fragment:

<table>
<thead><tr><th scope="col">Before — v2</th><th scope="col">After — v3</th></tr></thead>
<tbody><tr>
<td markdown="1">

```yaml
docker:
  WORKDIR: /app
  ENV:
    APP_ENV: production
```

</td>
<td markdown="1">

```yaml
imageSpec:
  config:
    workingDir: /app
    env:
      APP_ENV: production
```

</td>
</tr></tbody>
</table>

See [Changing image configuration spec]({{ "/usage/build/images.html#changing-image-configuration-spec" | true_relative_url }}) for the full set of fields.

**Old base images containing `WERF_COMMIT_*`.** Older werf builds could persist `WERF_COMMIT_HASH`, `WERF_COMMIT_TIME_HUMAN` and `WERF_COMMIT_TIME_UNIX` from the build container in the image configuration. That leak is fixed: werf now passes those variables only while running build commands. However, they can still be inherited from an old base image. imageSpec no longer removes them automatically when modifying env; if your base image contains them, add them to `imageSpec.config.removeEnv`.

### File imports

**Import caching now depends on the source image**, rather than checksums of the selected files as it did by default in v2. Changing the source image can rebuild the importing image even if the copied files are unchanged. `includePaths`/`excludePaths` still select which files to copy, but no longer isolate the cache from other source-image changes. If those rebuilds are expensive, put the imported output in a separate, narrowly scoped image.

**`import.stage` is removed.** Imports use the completed source image, not a selected intermediate stage. If you need an intermediate result, make it a separate image. `before`/`after` still control when the import runs in the **destination** image.

**A trailing slash in export/import `to:` is now an error**, except for the root path `to: /`. Previously werf stripped it with a warning, although users could expect it to mean “copy into this directory”. Check the intended destination before changing `to: /usr/sbin/` to `to: /usr/sbin`:

- If `add` is a directory, its contents are merged into `to`.
- If `add` is a file, it is copied into `to` when `to` is an existing directory; otherwise `to` is the file's destination path. To avoid depending on whether the containing directory exists, specify the full filename, for example `to: /etc/app/config.yaml`, and ensure that path is not a directory.

See [Destination path rules]({{ "/usage/build/stapel/imports.html#destination-path-rules" | true_relative_url }}) for details.

### Git dependencies of build stages

The default for `git.stageDependencies` has changed **for each of `install`, `beforeSetup` and `setup`**:

| Setting | Before — v2 | After — v3 |
|---|---|---|
| No masks, or an empty list | Git file changes did not invalidate the stage through `stageDependencies`. | Equivalent to `**/*`: all files in the git mapping are tracked, respecting `includePaths`/`excludePaths`. |
| Explicit masks | Only matching files affected the stage checksum. | The same: use masks to narrow the dependencies of expensive stages. |

**The default is safer:** changes to source files rerun build commands instead of requiring manual dependency configuration. Builds may run more often. If you specify masks, include every file that affects the stage's result; an empty list no longer disables file tracking.

A non-empty `stageDependencies` entry for a stage with no build instructions now causes an **error instead of a warning**. Correct the stage name, add the missing instructions, or remove the unused entry. See [Dependency on changes in the Git repo]({{ "/usage/build/stapel/instructions.html#dependency-on-changes-in-the-git-repo" | true_relative_url }}).

### Validation of werf-giterminism.yaml

Fix unknown or misspelled keys in `werf-giterminism.yaml`: strict schema validation no longer lets them be silently ignored.

### Which images are built and listed

`werf build` and other build-triggering commands now default to `--final-images-only=true`, as `converge`, `render`, `export`, `plan`, `lint` and `bundle publish` already did. Orphan non-final images that no final image references are no longer built by default.

`werf config list` also lists only final images by default; its deprecated `--images-only` alias is removed.

**To keep the previous behavior:**

```shell
werf build --final-images-only=false
werf config list --final-images-only=false
```

### Cache and git patches

**Git stages are reused without checking commit ancestry.** The cached commit no longer has to be an ancestor of the current one. `WERF_DISABLE_GIT_COMMIT_ANCESTRY_CHECK` is removed.

**Updating files from Git in legacy Stapel.** With the default `git.stageDependencies`, source changes rerun the build commands: no additional action is needed for this change. Updates affect files inside the image, not your Git working tree.

Check configurations that **narrow `git.stageDependencies` and modify source files in build commands**. If Git changes such a file but the processing stage is reused from cache, the file is replaced with its Git version, losing the command's changes. Previously a conflict could stop the build; now it can succeed with unexpected image contents.

- Include the files that trigger processing in the corresponding stage's `git.stageDependencies`, or write generated output separately from the source files.
- Test a rebuild after changing such a file, not just a clean build.

Text and binary files now use the same update mechanism: copying the required versions from Git. This avoids text-patch application errors, but no longer detects conflicts with changes made by build commands.

### Removed build settings

| Removed | What to change |
|---|---|
| `--synchronization` / `-S`, `WERF_SYNCHRONIZATION` | Remove them from commands and the environment. There is no replacement: stage tags are content-addressable, so a distributed lock manager is no longer needed. |
| `--virtual-merge`, `WERF_VIRTUAL_MERGE` | Remove them from commands and the environment. Virtual merge functionality is removed. |
| `WERF_STAGED_DOCKERFILE_VERSION=v1` | Remove the setting: staged Dockerfile always uses the v2 code path. |

The `werf synchronization` subsystem and the public `synchronization.werf.io` dependency are also removed. There is nothing to migrate; if you ran a private synchronization server, werf v3 no longer needs it.

### Buildah: host requirements and networking

The native Buildah backend switched from CNI/slirp4netns to netavark/pasta. Prepare the hosts before upgrading:

| Component | Where it is required |
|---|---|
| `netavark` | On every host using the Buildah backend, including chroot without network-using instructions. Backend initialization fails without it. |
| `pasta` from the `passt` package | For rootless builds that configure a network. It is not invoked with `network: host`, `network: none`, or chroot mode. |

`netavark` must be in one of `/usr/local/libexec/podman`, `/usr/local/lib/podman`, `/usr/libexec/podman`, `/usr/lib/podman` — `$PATH` is not searched. `pasta` is found in these directories or via `$PATH`. The official Ubuntu werf images are now based on Ubuntu 24.04.

**Check network settings.** Native Buildah now honors `network:` and `--backend-network` for non-staged Dockerfile builds; they used to be silently ignored. Only `default`, `host` and `none` are supported: `bridge` and Docker network names now cause an error.

Limitations and compatibility:

- Images with `staged: true` still ignore image-level network settings; only `RUN --network=` on an individual instruction applies.
- In `native-chroot`, Buildah forces host networking, so `network: none` does not provide isolation.
- CNI support is compiled out and cannot be restored. You can restore slirp4netns on an individual host via `CONTAINERS_CONF_OVERRIDE`: `default_rootless_network_cmd="slirp4netns"` in the `[network]` section.
- This is **not a migration of the system Podman/Buildah configuration**: werf neither reads nor rewrites `${graphroot}/defaultNetworkBackend`. If it contains `cni`, that value remains and continues to affect a system CLI sharing the same graphroot.

## Deployment

**Check `werf plan` before the first `werf converge`:** changes to values and file exclusion rules can affect manifests without a warning.

### Replacing removed werf helm commands

Deployment and release-inspection commands under `werf helm` are removed. For werf projects, use the dedicated commands:

| Before — v2 | After — v3 |
|---|---|
| `werf helm install` / `werf helm upgrade` | `werf converge` to install or update the project |
| `werf helm template` | `werf render` |
| `werf helm lint` | `werf lint` |
| `werf helm list` | `werf release list` |
| `werf helm get …` / `werf helm status` | `werf release get --release NAME --namespace NAMESPACE` |
| `werf helm history NAME` | `werf release history --release NAME --namespace NAMESPACE` |
| `werf helm rollback NAME REVISION` | `werf rollback --release NAME --namespace NAMESPACE --revision REVISION` |
| `werf helm uninstall NAME` | `werf dismiss --release NAME --namespace NAMESPACE` |
| `werf helm test` | Use the standalone `helm test` command; werf has no replacement command. |
| `werf helm plugin` | Manage and run plugins with Helm CLI directly; werf no longer loads Helm plugins. |

These are **workflow replacements, not drop-in aliases**. `converge`, `render` and `lint` use the werf project rather than Helm's positional `RELEASE CHART` arguments. Check the new command's `--help`, pass the intended release and namespace explicitly, and adapt scripts that parse output. For example, `release get --print-values` includes computed values in its structured output; it does not reproduce `helm get values` output. For standalone Helm chart workflows, use the Helm CLI directly.

`werf helm secret` and chart-management commands such as `werf helm dependency` remain available; the whole `werf helm` group has not been removed.

### Values and environment variables

| Where | Before — v2 | After — v3 |
|---|---|---|
| Main chart and dependency templates | `.Values.global.env` | `.Values.global.werf.env` |
| Release storage selection | `HELM_DRIVER` | `WERF_RELEASE_STORAGE` |

The old `.Values.global.env` key can render as an empty value without an error. Set `WERF_LEGACY_VALUES_GLOBAL_ENV=1` for temporary compatibility.

Without replacing `HELM_DRIVER`, werf uses its default release storage rather than the storage previously selected by that variable. Helm's path variables — `HELM_CACHE_HOME`, `HELM_CONFIG_HOME`, `HELM_DATA_HOME` — continue to work.

### Charts and bundles

**`.helmignore` now applies when reading the chart.** Excluded files disappear from the rendered manifests and the published bundle **without a warning**.

- Even without `.helmignore`, Helm's default rules exclude dot-prefixed files and directories directly under `templates/`. Directories are excluded with their contents.
- `**` now causes an error, although it previously had no effect.
- A leading `!` previously had no effect. It now excludes everything that does not match the pattern, usually leaving the chart empty. Do not use it to undo earlier exclusions as in `.gitignore`.

Rules for dependent charts vary by how the chart is included — see [Charts and dependencies]({{ "/usage/deploy/charts.html#excluding-files-or-directories-from-the-chart" | true_relative_url }}).

**`werf bundle publish` now defaults to `--helm-compatible-chart=true`.** The chart name in the published bundle's `Chart.yaml` becomes the last path component of the repository address, changing `.Chart.Name` and template paths. To keep the name from your `Chart.yaml`, pass:

```shell
werf bundle publish --helm-compatible-chart=false
```

The v1.2 `AllowMissedSecretKeyMode` compatibility mode is removed. However, `bundle publish` still does not require a secret key by default: secret values are handled without decryption by a different mechanism.

### Encrypted secret values

**An old client cannot read a secret written by v3, and there is no option to write the previous format.** Follow this upgrade order:

1. Upgrade **every** environment that reads the repository's secrets to werf v3 or nelm v2: developer machines, CI jobs, and jobs applying saved deploy plans.
1. Only then create, edit or rotate secrets. An old werf or nelm must not handle the repository after secrets are re-encrypted.

| Before — v2 | After — v3 |
|---|---|
| Secrets in AES-CBC format | Existing values remain readable; new writes use the new format. |
| The original YAML type of an encrypted value was not recorded | Newly encrypted YAML values retain their type and style. |

Additional considerations:

- The new format detects damaged ciphertext and incorrect keys.
- Whole secret-file ciphertext can still be used as a value in `secret-values.yaml`.
- Whole secret files use format 2; values in `secret-values.yaml` use format 3. Automatic format detection prevents a whole secret from being interpreted as YAML metadata.
- Old encrypted scalars remain strings. To restore a number, boolean, timestamp or another type, re-enter the value with `werf helm secret values edit`.
- **Comments on encrypted values are kept as cleartext. Do not put secrets in them.**

## Registry cleanup

**Old images are not deleted merely by upgrading werf**, but `cleanup` v3 can remove v2 images under the retention policies. Their version does not give them separate protection. Keep the tags needed for rollback in a `--keep-list` file, one stage tag per line. Do not rely only on Kubernetes protection: a rollback image may no longer be referenced by any scanned resource.

Before the first real cleanup, check that rollback images are not listed for deletion. Use the same repositories, keep-list and Kubernetes access that the scheduled job will use:

```shell
werf cleanup --repo registry.example.com/app --dry-run --keep-list keep-list.txt
```

If configured, also pass the same `--final-repo` and `--meta-repo` as in the build. Ensure the scan covers the clusters and namespaces using these images; `--without-kube` disables Kubernetes protection. Do not use `werf purge` to remove only v2 images: it deletes the project's images without cleanup's retention policies.

**A separate `--meta-repo` is optional.** Without it, metadata stays in `--repo`; upgrading alone does not require moving it. If you choose a separate metadata repository for an existing project:

1. Pause scheduled cleanup and old v2 jobs that write to the same repository during the transition.
1. From the project directory, migrate its existing metadata **before the first cleanup with `--meta-repo`**:

   ```shell
   werf meta-repo migrate --from registry.example.com/app --to registry.example.com/app-meta
   ```

1. Use the same `--meta-repo registry.example.com/app-meta` in all subsequent v3 commands, including build, cleanup and purge. The images stay in `--repo`; only metadata moves. By default, migration removes the originals after verifying the copies; `--remove-source=false` keeps them.
1. Recheck cleanup with `--dry-run` and the new `--meta-repo` before resuming the v3 cleanup job. Do not resume v2 cleanup against the migrated repository: it cannot see the current metadata in the new location.

**Adding `--meta-repo` alone does not migrate old metadata.** New writes go to the separate repository, while cleanup no longer sees the metadata left in `--repo`. It can therefore delete images that should have been retained. The safeguard checks that subsequent commands use the same metadata address, **not that migration is complete**. `meta-repo detach` only removes the safeguard; it does not move metadata back.

See [Container registry cleanup]({{ "/usage/cleanup/cr_cleanup.html" | true_relative_url }}) and [a separate metadata repository]({{ "/usage/build/process.html" | true_relative_url }}) for the full workflow.
