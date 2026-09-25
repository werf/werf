---
title: Migration from v2 to v3
permalink: resources/migration_from_v2_to_v3.html
---

First check compatibility with v2, then the changes to building, deployment and registry cleanup. Removed features require changes before upgrading; deprecated keys still work with a warning.

## Default behavior changes

The comparison below uses the default settings in v2.79.1 and v3.6.0. If you enabled experimental features in v2, some changes already apply to your environment. Changes introduced within v3 are marked with their version.

These changes affect existing projects even without configuration edits:

**Before the first `werf plan`, check [sensitive data redaction](#sensitive-data-in-diffs): the previous annotation no longer hides the entire resource.** This is especially important for shared CI logs.

### Building and caching

| Where | Before — v2 | After — v3 | What to check or how to restore the previous behavior |
|---|---|---|---|
| Building and listing images | `build` builds and `config list` lists non-final images too. | Final images are selected by default; builds also include their required dependencies. | `--final-images-only=false`; [details](#which-images-are-built-and-listed). |
| Git dependencies of stages | No direct Git-file dependency without `stageDependencies`. | All mapped files are tracked without explicit configuration. | Check masks; `[]` behaves differently before v3.6.0 — [details](#git-dependencies-of-build-stages). |
| Import cache | Depends on selected files. | Depends on the source image. | Additional rebuilds are possible; [details](#file-imports). |

### Deployment and bundles

| Where | Before — v2 | After — v3 | What to check or how to restore the previous behavior |
|---|---|---|---|
| Resource validation | Schema validation is off without an experimental flag. | Enabled. | Fix manifests or configure exceptions; [details](#resource-and-values-validation). |
| `patches.yaml` | Not applied automatically. | Files from the main chart and dependent charts are applied automatically. | Check their contents; disable with `--no-default-patches`; [details](#automatic-patches-and-null). |
| `null` in manifests | Preserved without an experimental cleanup flag. | Fields and list entries whose value is `null` are removed. | Check CRDs and intentional `null` values; [details](#automatic-patches-and-null). |
| Sensitive data in diffs | `Secret` resources and resources annotated with `werf.io/sensitive: "true"` are hidden except for identifying fields. | Only `data.*` and `stringData.*` are hidden by default. | Set sensitive paths **before running `plan`**; [details](#sensitive-data-in-diffs). |
| Service values | `.Values.global.env` is populated automatically. | `.Values.global.werf.env` is populated automatically; the old key is no longer populated. | Temporary compatibility: `WERF_LEGACY_VALUES_GLOBAL_ENV=1`; [details](#values-and-environment-variables). |
| Release storage | `HELM_DRIVER` is honored if `WERF_RELEASE_STORAGE` is not set. | `WERF_RELEASE_STORAGE` is used; without it, the default storage applies. | Transfer the variable's value; [details](#values-and-environment-variables). |
| Published chart name | `--helm-compatible-chart=false`. | `--helm-compatible-chart=true`. | Pass `false` if you need the previous name; [details](#charts-and-bundles). |
| `.helmignore` | Does not filter files when werf reads the chart. | **Since v3.6.0**, filters files, including Helm's default rules. | Check exclusions before deploying; [details](#charts-and-bundles). |

### CI and automation

| Where | Before — v2 | After — v3 | What to check or how to restore the previous behavior |
|---|---|---|---|
| Credentials in `ci-env` | `DOCKER_AUTH_CONFIG` requires explicit opt-in. | A non-empty variable is selected automatically unless the choice is explicit. | `--use-docker-auth-config=false`; [details](#registry-credentials). |

**Separately — CI contract changes, not new defaults:**

- With `--exit-code` enabled, code `3` means a release-only update — see [Plan exit codes](#plan-exit-codes).
- When reusing an image, the JSON build report may contain `StagesSkipped: true` without `Stages` — see [Build report format](#build-report-format).

Removed features and deprecated keys are covered separately below: replacing them is not the same as restoring previous defaults.

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

Base images and imports use `from` for both internal and external images. An image name from `werf.yaml`, such as `base`, refers to an internal image; a reference with a tag or digest, such as `alpine:3.20`, refers to an external image.

Each block below is a standalone `werf.yaml`. Every referenced project image is declared in the same example.

#### Stapel base image

The `app` image inherits from the `base` image declared above it. Replace `fromImage: base` with `from: base`:

<table>
<thead><tr><th scope="col">Before — v2</th><th scope="col">After — v3</th></tr></thead>
<tbody><tr>
<td markdown="1">

```yaml
configVersion: 1
project: migration-base
---
image: base
from: alpine:3.20
shell:
  install:
    - echo base > /base-marker
---
image: app
fromImage: base
shell:
  setup:
    - cat /base-marker
```

</td>
<td markdown="1">

```yaml
configVersion: 1
project: migration-base
---
image: base
from: alpine:3.20
shell:
  install:
    - echo base > /base-marker
---
image: app
from: base
shell:
  setup:
    - cat /base-marker
```

</td>
</tr></tbody>
</table>

#### Importing files from another image

The `builder` image creates a file, and `app` copies it before the `setup` stage. Replace `import.image` with `import.from`; also remove `stage`, because v3 imports from the completed source image:

<table>
<thead><tr><th scope="col">Before — v2</th><th scope="col">After — v3</th></tr></thead>
<tbody><tr>
<td markdown="1">

```yaml
configVersion: 1
project: migration-import
---
image: builder
from: alpine:3.20
shell:
  setup:
    - mkdir -p /out
    - echo hello > /out/message.txt
---
image: app
from: alpine:3.20
import:
  - image: builder
    stage: setup
    add: /out/message.txt
    to: /message.txt
    before: setup
shell:
  setup:
    - cat /message.txt
```

</td>
<td markdown="1">

```yaml
configVersion: 1
project: migration-import
---
image: builder
from: alpine:3.20
shell:
  setup:
    - mkdir -p /out
    - echo hello > /out/message.txt
---
image: app
from: alpine:3.20
import:
  - from: builder
    add: /out/message.txt
    to: /message.txt
    before: setup
shell:
  setup:
    - cat /message.txt
```

</td>
</tr></tbody>
</table>

#### Image dependency

The `app` image receives the name of the built `backend` image in the `BACKEND_IMAGE` variable. Replace `dependencies.image` with `dependencies.from`; the nested `imports` block passes image information rather than copying files:

<table>
<thead><tr><th scope="col">Before — v2</th><th scope="col">After — v3</th></tr></thead>
<tbody><tr>
<td markdown="1">

```yaml
configVersion: 1
project: migration-dependencies
---
image: backend
from: alpine:3.20
---
image: app
from: alpine:3.20
dependencies:
  - image: backend
    before: setup
    imports:
      - type: ImageName
        targetEnv: BACKEND_IMAGE
shell:
  setup:
    - echo "$BACKEND_IMAGE" > /backend-image.txt
```

</td>
<td markdown="1">

```yaml
configVersion: 1
project: migration-dependencies
---
image: backend
from: alpine:3.20
---
image: app
from: alpine:3.20
dependencies:
  - from: backend
    before: setup
    imports:
      - type: ImageName
        targetEnv: BACKEND_IMAGE
shell:
  setup:
    - echo "$BACKEND_IMAGE" > /backend-image.txt
```

</td>
</tr></tbody>
</table>

The `fromImage`, `import.image` and `dependencies.image` keys **still work** in v3, but emit a deprecation warning. Specifying both the old and new key is an error. Unlike these keys, `import.stage` is removed.

#### External image with an explicit tag

An **external reference** in a base `from` or `import.from` requires an explicit tag or digest. For example, replace the implicit `:latest` with an explicit tag:

<table>
<thead><tr><th scope="col">Before — v2</th><th scope="col">After — v3</th></tr></thead>
<tbody><tr>
<td markdown="1">

```yaml
configVersion: 1
project: migration-external
---
image: app
from: ubuntu
```

</td>
<td markdown="1">

```yaml
configVersion: 1
project: migration-external
---
image: app
from: ubuntu:latest
```

</td>
</tr></tbody>
</table>

Instead of `:latest`, specify the required tag (`:TAG`) or digest (`@sha256:...`). Internal image names from `werf.yaml` do not need a tag.

### Builders and image configuration

**The legacy Docker builder is no longer available.** BuildKit was already the default in v2, but `DOCKER_BUILDKIT=0` or `false` allowed opting back into the old builder. In v3, this variable does not restore that mode: check your Dockerfiles and environment with BuildKit.

`WERF_STAGED_DOCKERFILE_VERSION=v1` no longer selects the old staged Dockerfile implementation. Together with cache calculation changes, this requires a rebuild after moving from v2.

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

### File imports

**Import caching now depends on the source image**, rather than checksums of the selected files as it did by default in v2. Changing the source image can rebuild the importing image even if the copied files are unchanged. `includePaths`/`excludePaths` still select which files to copy, but no longer isolate the cache from other source-image changes. If those rebuilds are expensive, put the imported output in a separate, narrowly scoped image.

**`import.stage` is removed.** Imports use the completed source image, not a selected intermediate stage. If you need an intermediate result, make it a separate image. `before`/`after` still control when the import runs in the **destination** image.

**A trailing slash in export/import `to:` is now an error**, except for the root path `to: /`. Previously werf stripped it with a warning, although users could expect it to mean “copy into this directory”. Check the intended destination before changing `to: /usr/sbin/` to `to: /usr/sbin`:

- If `add` is a directory, its contents are merged into `to`.
- If `add` is a file, it is copied into `to` when `to` is an existing directory; otherwise `to` is the file's destination path. To avoid depending on whether the containing directory exists, specify the full filename, for example `to: /etc/app/config.yaml`, and ensure that path is not a directory.

See [Destination path rules]({{ "/usage/build/stapel/imports.html#destination-path-rules" | true_relative_url }}) for details.

### Git dependencies of build stages

`git.stageDependencies` determines which Git file changes trigger stage rebuilds. **The rules below apply starting with v3.6.0.** In v3.5.0 and earlier v3 versions, both an omitted stage setting and an explicit `[]` are replaced with `**/*`: an empty list does not disable the direct Git-file dependency there. Upgrade to v3.6.0 or later to use `[]` and the partially filled block rules below.

Starting with v3.6.0, **an omitted setting and an explicit empty list can mean different things**:

| Setting | Before — v2 | After — v3 |
|---|---|---|
| The entire block is absent or `{}` | Stages have no direct Git-file dependency. | All three stages get `**/*`: all files in the git mapping are tracked, respecting `includePaths`/`excludePaths`. |
| Explicit stage masks | Matching files are tracked. | The same behavior. |
| An explicit `[]` for a stage | No direct Git-file dependency. | The same behavior; the stage counts as explicitly declared. |

In a partially filled block, omitted stages **before the last explicitly declared stage** get `[]`, and those **after it** get `**/*`. The order is always `install` → `beforeSetup` → `setup`, regardless of YAML key order. The rule applies independently to each git mapping. In v2, every omitted stage had no Git-file dependency regardless of its position: the new `**/*` defaults after the last declared stage can cause additional rebuilds.

For example, only `beforeSetup` is declared:

```yaml
git:
  - add: /
    to: /app
    stageDependencies:
      beforeSetup:
        - "src/**/*"
```

`install` gets `[]`, `beforeSetup` gets `src/**/*`, and `setup` gets `**/*`. Changing a file outside `src` therefore does not by itself trigger `install` or `beforeSetup`, but triggers `setup` if it has instructions.

If `install` and `setup` are declared:

```yaml
git:
  - add: /
    to: /app
    stageDependencies:
      install:
        - package-lock.json
      setup:
        - "src/**/*"
```

The omitted `beforeSetup` gets `[]`: it precedes the last declared stage, `setup`. An empty list also establishes that boundary: with only `beforeSetup: []`, both `install` and `beforeSetup` have empty masks, while `setup` gets `**/*`.

**The default is safer:** without dependency configuration, source changes rerun build commands. Builds may run more often. To keep a stage free of direct Git-file dependencies, specify `[]`; to always track all mapped files, specify `["**/*"]`. If you specify masks, include every file that affects the stage's result.

`[]` disables only the direct Git-file dependency. Changes to preceding stages, base images, commands and other inputs can still trigger a rebuild. Masks do not create stages without instructions; `beforeInstall` is not part of this mechanism.

A non-empty `stageDependencies` entry for a stage with no build instructions now causes an **error instead of a warning**. For example, this fragment fails validation when building:

```yaml
git:
  - add: /
    to: /app
    stageDependencies:
      install:
        - "src/**/*"
shell:
  setup:
    - cd /app && sh src/build.sh
```

Dependencies are declared for `install`, but only `setup` has commands. If the `src` files are inputs to those commands, move the masks from `stageDependencies.install` to `stageDependencies.setup`. If an `install` stage is intended, add `shell.install` instructions; if the dependency is unnecessary, remove it.

See [Dependency on changes in the Git repo]({{ "/usage/build/stapel/instructions.html#dependency-on-changes-in-the-git-repo" | true_relative_url }}).

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

### Synchronization server

werf v3 no longer uses a synchronization server, including the public `synchronization.werf.io`. If you ran your own server, v3 no longer needs it. Shut it down only after upgrading all clients that use it; remaining v2 clients may still need it.

### Buildah

The native Buildah backend switched from CNI/slirp4netns to netavark/pasta. Before upgrading, prepare the environment according to how you run werf.

#### Official werf image

Update the **werf container image to v3**, not just the binary inside it. The image already includes `netavark`; you do not need to install it separately on the runner host to run werf in that container.

The official Ubuntu werf images are now based on Ubuntu 24.04. Rebuild and check any images you derive from them.

For rootless builds that configure a network, also check that `pasta` is available inside the container under the conditions below; having `netavark` alone is not enough. If it is missing, add it to your derived image.

#### Your own installation or image

If you install werf directly on a host or use your own container image, make sure the dependencies are available **where werf runs**:

| Component | Where it is required |
|---|---|
| `netavark` | Wherever the Buildah backend runs, including chroot without network-using instructions. Backend initialization fails without it. |
| `pasta` from the `passt` package | For rootless builds that configure a network. It is not invoked with `network: host`, `network: none`, or chroot mode. |

`netavark` must be in one of `/usr/local/libexec/podman`, `/usr/local/lib/podman`, `/usr/libexec/podman`, `/usr/lib/podman` — `$PATH` is not searched. `pasta` is found in these directories or via `$PATH`.

#### Network setting changes

These changes apply both inside and outside the official image.

**Check network settings.** Native Buildah now honors `network:` and `--backend-network` for non-staged Dockerfile builds; they used to be silently ignored. Only `default`, `host` and `none` are supported: `bridge` and Docker network names now cause an error.

Limitations and compatibility:

- Images with `staged: true` still ignore image-level network settings; only `RUN --network=` on an individual instruction applies.
- In `native-chroot`, Buildah forces host networking, so `network: none` does not provide isolation.
- CNI support is compiled out and cannot be restored. You can restore slirp4netns on an individual host via `CONTAINERS_CONF_OVERRIDE`: `default_rootless_network_cmd="slirp4netns"` in the `[network]` section.
- This is **not a migration of the system Podman/Buildah configuration**: werf neither reads nor rewrites `${graphroot}/defaultNetworkBackend`. If it contains `cni`, that value remains and continues to affect a system CLI sharing the same graphroot.

## Deployment

**Before the first `werf plan`, check [sensitive data redaction](#sensitive-data-in-diffs)**, especially if other users can read your CI logs. Then review the plan before the first `werf converge`: changes to values, patches and file exclusion rules can affect manifests without a warning.

### Sensitive data in diffs

By default in v2, only identifying fields remained visible for `Secret` resources and any resource annotated with `werf.io/sensitive: "true"`: `apiVersion`, `kind`, `metadata.name` and `metadata.namespace`. In v3, only values at `data.*` and `stringData.*` are automatically hidden. Other fields, including metadata and secret key names, may be visible; hidden values are replaced with length and hash information.

**The `werf.io/sensitive: "true"` annotation no longer guarantees redaction of an entire arbitrary resource.** If sensitive data is stored in fields such as `spec`, specify its JSONPath expressions with `werf.io/sensitive-paths`, for example `"$.spec.token"`. This annotation replaces the default path list rather than extending it: include `$.data.*` and `$.stringData.*` as well if you need to keep Secret data redacted. Check this before printing diffs to shared logs. This is independent of the encrypted file format: encrypting `secret-values.yaml` does not define redaction rules for fields in rendered resources.

### Resource and values validation

**Kubernetes resource schema validation is enabled by default.** In v2, it required an experimental flag. Manifests that previously passed may now be rejected before they are applied. Fix the resource or schema; use `--resource-validation-extra-schema` for additional schemas. The old `--local-resource-validation`, `--resource-validation-kube-version` and `--resource-validation-schema` flags are removed.

If you need an exception, use `--resource-validation-skip`; disable all resource validation with `--no-resource-validation`. Do not disable validation for every resource just to accommodate one unsupported type.

**`values.schema.json` validation already existed in v2, but error handling is stricter.** Previously, an error about forbidden service values could turn the entire validation result into a warning, including errors in user values. Now werf retries validation without service values and fails if user values do not match the schema. Fix the values or schema; temporarily disable this check with `--no-values-schema-validation`.

### Automatic patches and null

**`patches.yaml` from the main chart and its dependencies is now read automatically.** An existing file with this name that served another purpose may change manifests or cause an error. Check it before deploying. Disable automatic loading with `--no-default-patches` or `WERF_NO_DEFAULT_PATCHES=true`.

**Fields and list entries whose value is `null` are now removed recursively.** In v2, this required an experimental flag. An absent field and an explicit `null` are not always equivalent for Kubernetes and CRDs, so check resources that intentionally use `null`. There is no switch to restore the previous behavior.

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

### Uninstalling a release

`werf dismiss` now always uses the new uninstall implementation, which v2 enabled with `WERF_EXPERIMENT_NEW_DISMISS`. The old implementation selector and `--with-hooks` flag are removed. Do not carry that flag into v3 scripts: check release and hook removal in a test environment. Deleting the namespace still requires `--with-namespace`.

### Values and environment variables

| Where | Before — v2 | After — v3 |
|---|---|---|
| Main chart and dependency templates | `.Values.global.env` | `.Values.global.werf.env` |
| Release storage selection | `HELM_DRIVER` | `WERF_RELEASE_STORAGE` |

The old `.Values.global.env` key can render as an empty value without an error. Set `WERF_LEGACY_VALUES_GLOBAL_ENV=1` for temporary compatibility.

Without replacing `HELM_DRIVER`, werf uses its default release storage rather than the storage previously selected by that variable. Helm's path variables — `HELM_CACHE_HOME`, `HELM_CONFIG_HOME`, `HELM_DATA_HOME` — continue to work.

### Charts and bundles

**Starting with v3.6.0, `.helmignore` applies when reading the chart.** In v3.5.0 and earlier v3 versions, werf's own chart loader did not apply it. Excluded files disappear from the rendered manifests and the published bundle **without a warning**.

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

## CI and scripts

### Registry credentials

`werf ci-env` now automatically uses a non-empty `DOCKER_AUTH_CONFIG` if neither `--use-docker-auth-config` nor `WERF_USE_DOCKER_AUTH_CONFIG` is set. In v2, this required explicit opt-in.

With this choice, the Docker config is created from `DOCKER_AUTH_CONFIG` **instead of copying the existing config**, not merged with it. Credentials and credential helpers from the previous config may no longer be used. Restore the previous choice with `--use-docker-auth-config=false` or `WERF_USE_DOCKER_AUTH_CONFIG=false`.

### Plan exit codes

With `--exit-code` enabled, `werf plan` and `werf bundle plan` now always distinguish resource changes from release-only updates. In v2, the extended set of codes required an experimental flag.

| Code | Meaning |
|---|---|
| `0` | No changes. |
| `1` | An error. |
| `2` | Resource changes are planned. |
| `3` | Resources do not change, but the release needs to be installed or updated. |

Update CI if it only accepts `0` and `2`. Code `3` is not an error, but is not a reason to automatically skip applying the plan either. Without `--exit-code`, a successful plan does not start returning `2` or `3`.

### Build report format

When reusing a completed image, the JSON build report may now contain `StagesSkipped: true` without a `Stages` field. In v2, `Stages` was present, although it could be `null`. Parsers must tolerate the missing field and not treat it as an error or as an indication that the image is not ready.

### Removed flags and modes

Check scripts that pass old options: a removed flag causes an argument parsing error even if it previously did nothing.

- `--virtual-merge` / `WERF_VIRTUAL_MERGE`, `--skip-image-spec-stage` / `WERF_SKIP_IMAGE_SPEC_STAGE`, `--set-runtime-json` and `--show-verbose-diffs` are removed.
- `--synchronization` / `-S` / `WERF_SYNCHRONIZATION` and the `werf synchronization` command group are removed — see [Synchronization server](#synchronization-server).
- Helm mode via `WERF_HELM3_MODE` and invoking the werf binary under the name `helm` are removed; use supported werf commands or the standalone Helm CLI.
- Positional image names in `converge` and `plan` now take effect without `WERF_CONVERGE_ENABLE_IMAGES_PARAMS`. Check that stray arguments have not become image selectors; selecting images does not by itself limit which Kubernetes resources are deployed.
- `cleanup --kube-scan-namespaces`, available in v2.79.1, was absent in v3.5.0 but is available again starting with v3.6.0. If your script uses it, upgrade to v3.6.0 or later rather than dropping the namespace restriction without checking access permissions and image protection.

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
