---
title: Migration from v2 to v3
permalink: resources/migration_from_v2_to_v3.html
---

The v3 changes are grouped by where you configure them: `werf.yaml`, build and CI, and deployment. Removed features require changes before upgrading; deprecated keys still work with a warning.

## Build configuration — werf.yaml

### Images instead of artifacts

The `artifact` directive is removed. Replace it with `image` and `final: false`:

**Before — v2:**

```yaml
artifact: builder
from: ubuntu:22.04
```

**After — v3:**

```yaml
image: builder
from: ubuntu:22.04
final: false
```

Nameless stapel images (`image: ~`) are no longer supported — give every image a name.

Names are now checked when loading the configuration. Latin letters, digits, `_`, `.`, `-` and `+` are allowed, with optional `/`-separated segments. Each segment must start with a letter or digit and end with a letter, digit or `+`. `modules/controller` is valid; an empty name, whitespace, `/api`, `api-` and `modules//controller` cause a configuration error. Check names produced by Go templates as well.

### Image references

Use the unified `from` key for internal references:

| Where | Before — v2 | After — v3 |
|---|---|---|
| Stapel base image | `fromImage: base` | `from: base` |
| An `import` entry | `image: builder` | `from: builder` |
| A `dependencies` entry | `image: backend` | `from: backend` |

The old keys are **not removed**, but deprecated. `fromImage` and `import.image` emit a warning; specifying both the old and new key is an error. `dependencies.image` also continues to work, but prefer `dependencies.from` in new configurations.

An **external reference** in a base `from:` or `import.from` now requires an explicit tag or digest. For example, the previously implicit `:latest` must be written explicitly:

| Before — v2 | After — v3 |
|---|---|
| `from: ubuntu` | `from: ubuntu:latest` |

Instead of `:latest`, you can specify the required tag (`:TAG`) or digest (`@sha256:...`). Internal image names from `werf.yaml` do not need a tag.

### Builders and image configuration

**The Ansible builder is removed.** Rewrite `ansible:` steps using the Shell builder; renaming the key is not enough.

**The `docker:` directive is removed.** Move its settings to `imageSpec.config`, translating field names and formats. For example, for a stapel image fragment:

**Before — v2:**

```yaml
docker:
  WORKDIR: /app
  ENV:
    APP_ENV: production
```

**After — v3:**

```yaml
imageSpec:
  config:
    workingDir: /app
    env:
      APP_ENV: production
```

See [Image specification configuration]({{ "/usage/build/images.html" | true_relative_url }}) for the full set of fields.

`WERF_COMMIT_HASH`, `WERF_COMMIT_TIME_HUMAN` and `WERF_COMMIT_TIME_UNIX` are no longer forcibly removed when modifying env through imageSpec. If an old base image still carries them, add them explicitly to `imageSpec.config.removeEnv`.

### Paths and giterminism

- Remove the trailing slash from export/import `to:` paths: `to: /usr/sbin/` → `to: /usr/sbin`. The root path `to: /` is exempt. What used to produce a warning now causes the build to fail.
- Fix unknown or misspelled keys in `werf-giterminism.yaml`: strict schema validation no longer lets them be silently ignored.

## Build and CI

### Which images are built and listed

`werf build` and other build-triggering commands now default to `--final-images-only=true`, as `converge`, `render`, `export`, `plan`, `lint` and `bundle publish` already did. Orphan non-final images that no final image references are no longer built by default.

`werf config list` also lists only final images by default; its deprecated `--images-only` alias is removed.

**To keep the previous behavior:**

```shell
werf build --final-images-only=false
werf config list --final-images-only=false
```

### Cache and git patches

- **Expect a one-time full rebuild after upgrading.** `WERF_STAGED_DOCKERFILE_VERSION=v1` is no longer supported: staged Dockerfile always uses the v2 code path. Together with the digest-calculation fixes in this release, this invalidates pre-v3 stage caches.
- **Git stages are reused without checking commit ancestry.** The cached commit no longer has to be an ancestor of the current one. `WERF_DISABLE_GIT_COMMIT_ANCESTRY_CHECK` is removed.
- **Git patches can overwrite changes made by build commands.** The legacy stapel builder no longer uses `git apply`: a file modified by an earlier `install`/`beforeSetup`/`setup` command is silently overwritten instead of producing a conflict error.
- If external tools parse werf's raw git patch output, account for the removed `index <sha>..<sha>` line.
- If integrations use the `werf.io/base-image-id` label or `Info.ParentID` field, update them: both are removed. Cleanup now tracks ancestry using only `werf.io/parent-stage-id`.

### Removed CI settings

| Removed | What to change |
|---|---|
| `--synchronization` / `-S`, `WERF_SYNCHRONIZATION` | Remove them from commands and the environment. There is no replacement: stage tags are content-addressable, so a distributed lock manager is no longer needed. |
| `--virtual-merge`, `WERF_VIRTUAL_MERGE` | Remove them from commands and the environment. Virtual merge functionality is removed. |

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
- Do not use a leading `!` as in `.gitignore`: it does not undo earlier exclusions, but excludes everything that does not match the pattern, usually leaving the chart empty.

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
| The original YAML type of an encrypted value was not recorded | Newly encrypted YAML values retain their type and style; damaged ciphertext and incorrect keys are detected. |

Additional considerations:

- Whole secret-file ciphertext can still be used as a value in `secret-values.yaml`.
- Whole secret files use format 2; values in `secret-values.yaml` use format 3. Automatic format detection prevents a whole secret from being interpreted as YAML metadata.
- Old encrypted scalars remain strings. To restore a number, boolean, timestamp or another type, re-enter the value with `werf helm secret values edit`.
- **Comments on encrypted values are kept as cleartext. Do not put secrets in them.**
