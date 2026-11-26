---
title: Migration from v2 to v3
permalink: resources/migration_from_v2_to_v3.html
---

Migrate in three stages:

1. [**Before upgrading**](#before-upgrading-find-problems-with-v3-keep-deploying-with-v2): run werf v3 in a branch to find what needs changing, and fix it in a form werf v2 also accepts. Production keeps deploying with v2. Prepare the CI changes that work only in v3.
2. [**Upgrading**](#upgrading): switch to v3 in a fixed order. The new secret and bundle formats can't be undone by switching back.
3. [**After upgrading**](#after-upgrading): check the defaults that changed and remove compatibility settings.

Go through the [checklist](#checklist) and skip what doesn't apply to your project. Every removed command, flag and variable is listed in the [reference](#reference-removed-commands-flags-and-variables).

## Checklist

**Before upgrading (v2 still deploys):**

- **Everyone:** [upgrade to the latest v2](#upgrade-to-the-latest-v2), [run v3 to find problems](#find-problems-with-v3) and [fix what it reports](#fix-what-v3-reports).

Then check the [changes v3 doesn't report](#changes-v3-doesnt-report) that apply to you:

- **Stapel:** [set `stageDependencies` explicitly](#stapel-git-dependencies-of-stages), [check imports](#stapel-imports).
- **Dockerfile:** [check the builder, ignore files and Git LFS](#dockerfile).
- **Native Buildah:** [install netavark and pasta, check network settings](#buildah).
- **Charts:** [update templates and annotations](#charts-templates-and-annotations), [check values and secrets](#charts-values-and-secrets), [check `.helmignore` and `patches.yaml`](#charts-files-werf-now-reads-differently).

And [prepare CI](#prepare-ci):

- **CI:** [replace `werf helm` commands](#ci-werf-helm-commands), [remove dropped flags and variables](#ci-flags-variables-and-environment), [accept the new exit codes and report formats](#ci-exit-codes-and-reports).

**Upgrading:**

- **Everyone:** [switch in this order](#order-of-the-switch): protect rollback history, switch runners, migrate SQL release storage, upgrade everything that reads secrets or deploys bundles, then publish bundles and edit secrets.
- **SQL release storage:** [run the storage migration with a privileged role](#release-storage) before other v3 jobs use it.
- **TypeScript charts:** [move to Node.js](#typescript-charts).
- **Bundles:** [upgrade consumers before publishing, check the chart name](#bundles).
- **Secrets:** [upgrade all readers before writing secrets](#secrets).
- **Everyone:** [expect a full rebuild and protect rollback images](#first-build-and-cleanup), [check the first deploy](#first-deploy).

**After upgrading:**

- **Everyone:** [check the changed defaults](#check-the-changed-defaults), [remove compatibility settings](#clean-up).
- **Go code importing werf:** [change the module path](#go-module-path).

## Before upgrading: find problems with v3, keep deploying with v2

In this stage, production CI keeps using werf v2. You run v3 only to find problems, and fix each one in a form that v2 also accepts, so the fixes can be merged one by one. The sections below say how for each change.

### Upgrade to the latest v2

Upgrade to the latest werf v2. Older v2 releases deploy bundles published by v3 without their values instead of refusing them.

### Find problems with v3

1. Install werf v3 next to v2: on your machine or in a separate CI job.
2. In a branch, run these v3 commands for each environment. Pass `werf render` and `werf lint` the same `--values`, `--secret-values` and `--set*` as your deploy jobs:

   ```shell
   werf config render --env ENV
   werf render --env ENV --stub-tags --validate=false
   werf lint --env ENV --stub-tags --validate=false
   ```

   They don't build, push or deploy, and don't need the cluster. Skip `render` and `lint` if the project has no chart. Pass the secret key as usual. Without the key, add `--ignore-secret-key`; secret values then aren't checked. `werf render` prints Secret data: run it where the output isn't shared, or write it to a private file with `--output`.
3. v3 may stop at the first error. Fix it using the [table below](#fix-what-v3-reports), and run again until the commands pass. Also fix the deprecation warnings that `werf lint` or `werf build` print. If v3 rejects a flag (`unknown flag`), use its replacement from [CI: flags](#ci-flags-variables-and-environment).
4. Check each fix with v2 too (`werf render`, `werf build`, your tests) before merging it: production still deploys with v2.
5. Compare `werf render` output of v2 and v3 for each environment: [values priority](#charts-values-and-secrets) and other silent changes show up only there.

Then go through the [changes v3 doesn't report](#changes-v3-doesnt-report). To try a v3 build, run `werf build` on a separate runner, with a separate `--repo` and without the production `--final-repo`, `--cache-repo`, `--meta-repo` and custom tags.

**Don't run v3 `build`, `converge`, `plan` or `cleanup` against production repositories, runners or release storage before [the switch](#order-of-the-switch):** v3 rebuilds all images, can't share a runner with v2, migrates SQL release storage if you use it, and its cleanup can delete v2 images.

### Fix what v3 reports

All these fixes also work in v2.

| v3 reports | Fix |
|---|---|
| `... is a forbidden property` in `werf-giterminism.yaml` | v2 ignored the key. Remove it to keep the v2 behavior, or fix its name to turn the setting on. See the [giterminism reference]({{ "/reference/werf_giterminism_yaml.html" | true_relative_url }}). |
| `artifact` ("cannot recognize type of config section") | Make it an image with `final: false`, and replace `fromArtifact` with `from` and `import[].artifact` with `import[].from`: see [the example below](#replacing-an-artifact). |
| `ansible` | Rewrite the steps as `shell:` commands, or move the image to a Dockerfile. |
| `docker` | Move the settings to `imageSpec.config`, see [Changing image configuration spec]({{ "/usage/build/images.html#changing-image-configuration-spec" | true_relative_url }}). In v2, `docker` had no effect in an image with `disableGitAfterPatch: true`: moving it applies these settings for the first time, so delete `docker` instead to keep the v2 result. If v2 CI passes `--skip-image-spec-stage`, v2 ignores `imageSpec`, so the settings are lost until you switch: make this change together with the binary. |
| `import[].stage` | Remove it: imports now copy from the finished source image. For an intermediate result, import from a separate image with `final: false`. |
| An image without a name | Name it, and use {% raw %}`{{ index .Values.werf.image "NAME" }}`{% endraw %} instead of `.Values.werf.nameless_image`. |
| `disableGitAfterPatch` | Remove it: Git files are always updated. |
| Deprecated `fromImage`, `import[].image`, `dependencies[].image` | Rename them to `from`, `import[].from`, `dependencies[].from`. Don't set an old key and its new key together. |
| An external image without a tag or digest | Add one: `from: ubuntu` becomes `from: ubuntu:latest`, which is the image v2 pulled. Don't add a tag to images from `werf.yaml`. Rename an image called `scratch`: the name is reserved. |
| An invalid image name | Use Latin letters, digits, `_`, `.`, `-` and `+` in segments separated by `/`. Each segment starts with a letter or digit and ends with a letter, digit or `+`. |
| `to` of `git` or `import` ending with `/` | Remove the slash (keep `to: /` as is), and check where the files go: [Destination path rules]({{ "/usage/build/stapel/imports.html#destination-path-rules" | true_relative_url }}). |
| A number or boolean in annotations or labels | Quote it: `example.com/replicas: "3"` or {% raw %}`{{ .Values.replicas | quote }}`{% endraw %}. v2 dropped such values with a warning, so the first deploy adds them to live resources: remove the ones you don't need. |
| Values or a schema rejected by `values.schema.json` validation | Fix the values or the schema: v2 often hid these errors. Schemas without `$schema` are now checked as JSON Schema 2020-12; add `"$schema": "http://json-schema.org/draft-07/schema#"` to keep draft-07. The error text changed: update scripts that parse it. `--no-values-schema-validation` skips the check temporarily in v3; v2 doesn't have the flag. |

#### Replacing an artifact

<table>
<thead><tr><th scope="col">Before</th><th scope="col">After</th></tr></thead>
<tbody><tr>
<td markdown="1">

```yaml
artifact: builder
from: alpine:3.20
# ...
---
image: app
fromArtifact: builder
import:
- artifact: builder
  add: /out
  to: /app
  before: setup
```

</td>
<td markdown="1">

```yaml
image: builder
final: false
from: alpine:3.20
# ...
---
image: app
from: builder
import:
- from: builder
  add: /out
  to: /app
  before: setup
```

</td>
</tr></tbody>
</table>

An artifact didn't get Git updates after its first build; an image does. If the artifact's instructions process files from Git, set `stageDependencies` of that stage to those files: with v2 defaults the files update but the result built from them doesn't.

### Changes v3 doesn't report

The commands above don't reliably catch these changes, so check the ones that apply to your project.

#### Stapel: Git dependencies of stages

**Set `install`, `beforeSetup` and `setup` explicitly** in each git mapping with Shell instructions, so that v2 and v3 behave the same: masks of the files that should rerun the stage, `["**/*"]` for all files, or `[]` for none. For example, if only `beforeSetup` uses `src/`:

```yaml
git:
- add: /
  to: /app
  stageDependencies:
    install: []
    beforeSetup: ["src/**/*"]
    setup: []
```

| Setting | v2 | v3 |
|---|---|---|
| No `stageDependencies`, or `{}` | Git changes rerun no stages. | All three stages get `**/*`: every file of the mapping, respecting `includePaths`/`excludePaths`. |
| Some stages set | Omitted stages don't depend on Git. | In the order `install` → `beforeSetup` → `setup`, omitted stages before the last set one get `[]`, those after it get `**/*`. See [Masks of the stages that are not set]({{ "/usage/build/stapel/instructions.html#masks-of-the-stages-that-are-not-set" | true_relative_url }}). |

**A non-empty entry for a stage without instructions** is an error at `werf build` (`[]` is fine); the commands above don't report it. Move the masks to the stage with instructions, or set the entry to `[]`.

**Commands that change files from Git.** On the Docker backend, v2 stopped the build with a `git apply` conflict when a later Git change touched a file that a command had modified. v3 silently overwrites the touched files with the files from Git. Find these commands, then either make the stage that changes the files depend on their sources in `stageDependencies`, or write the generated files outside the mapped paths.

See [Dependency on changes in the Git repo]({{ "/usage/build/stapel/instructions.html#dependency-on-changes-in-the-git-repo" | true_relative_url }}).

#### Stapel: imports

**Several imports into one directory.** v2 automatically excluded one import's paths from another when their `to` paths were equal or nested; v3 copies them as written, in order. To keep the v2 result, add the excluded paths to `excludePaths` (works in v2 too). v2 prints them in the build log as `import[... excludePaths=[...]]`: for imports into `/app/bin` and `/app/bin/tools`, the first one excludes `tools`.

**Import caching.** An importing image is now rebuilt whenever the source image changes, even if the copied files are the same. If that's expensive, import from a small separate image.

#### Dockerfile

- **BuildKit only.** `DOCKER_BUILDKIT=0` no longer selects the legacy Docker builder. Remove the variable and check that your Dockerfiles build with BuildKit; this works in v2 too.
- **Docker 19.03 or newer** is required for the Docker backend.
- **Staged Dockerfiles** (`staged: true`) use a new implementation: check that these images build with v3.
- **Ignore files.** werf now filters the build context itself on both backends, using the first file that exists: `<Dockerfile>.dockerignore`, `<Dockerfile>.containerignore`, `.dockerignore`, `.containerignore`. There's no warning about the change, so check the files in the image after switching.

  | Backend | What changes | To keep the v2 result |
  |---|---|---|
  | Docker | v2 ignored `.containerignore` and `<Dockerfile>.containerignore`. If a context has only those, their rules now apply. | Add an empty file that comes earlier in the list above: `.dockerignore` for `.containerignore`, `<Dockerfile>.dockerignore` for `<Dockerfile>.containerignore`. To get the v3 result already in v2, copy the rules to `.dockerignore` instead. |
  | Buildah | Buildah read `.containerignore` before `.dockerignore`. If a context has both, v3 uses `.dockerignore`. | Put the rules that v2 applied into the file v3 picks first, from the list above, Dockerfile-specific files included. Also check `contextAddFiles`: files that an ignore file removed can now reach the image through a broad `COPY`. |

- **Git LFS.** Contents of Git LFS files are now included in Stapel git mappings and Dockerfile contexts. If your repositories use LFS, install `git-lfs` where werf runs and make sure LFS objects can be downloaded.

#### Buildah

**netavark and pasta replace CNI and slirp4netns.**

- With the official werf image, switch to the v3 image, not only the binary: it includes netavark. It's based on newer Alpine, Fedora and Ubuntu 24.04, so rebuild derived images and check that `pasta` is there for rootless builds with a network.
- With your own installation, install `netavark` (always needed, chroot mode too) and `pasta` from `passt` (rootless builds with a network). v2 doesn't use them, so you can install them in advance.
- To use slirp4netns instead of pasta, set `default_rootless_network_cmd="slirp4netns"` in the `[network]` section via `CONTAINERS_CONF_OVERRIDE`. CNI can't be used. See [Network tools]({{ "/usage/build/backends.html#network-tools" | true_relative_url }}).

**Network settings.** Native Buildah now applies `network:` of Dockerfile images and `--backend-network`, which v2 ignored. Only `default`, `host` and `none` are supported: replace `bridge` and Docker network names, and check builds that relied on host networking. Images with `staged: true` still ignore image-level settings: use `RUN --network=`. In `native-chroot` mode the network is always the host's, so `network: none` doesn't isolate the build.

#### Charts: templates and annotations

**`.Values.global.env` is no longer set.** Use `.Values.global.werf.env` instead (or `.Values.werf.env` in the main chart) in templates and `values.schema.json`. v2 sets both unless `WERF_EXPERIMENT_NO_GLOBAL_SERVICE_VALUES` is enabled, so unset that variable first and check `werf render` on v2. If you can't migrate in time, `WERF_LEGACY_VALUES_GLOBAL_ENV=1` makes v3 set the old key temporarily.

**External dependencies.** v3 ignores `<name>.external-dependency.werf.io/resource` and `/namespace` annotations, with a warning. Add `werf.io/deploy-dependency-<name>` **without** `external=` next to them, and remove the old ones after switching:

```yaml
annotations:
  # Used by v2. Remove after switching.
  secret.external-dependency.werf.io/resource: secret/my-vault-secret
  secret.external-dependency.werf.io/namespace: vault
  # Used by v3.
  werf.io/deploy-dependency-secret: state=present,kind=Secret,version=v1,name=my-vault-secret,namespace=vault
```

- v2 ignores the new annotation if no release resource matches it; v3 waits for the resource in the cluster. v3 warns about the old annotations on every command until you remove them.
- Set `kind`, `version` and `name`; `group` for resources outside the core API group; and `namespace` for namespaced resources ({% raw %}`{{ .Release.Namespace }}`{% endraw %} for the release namespace): otherwise a release resource with the same name in another namespace also matches.
- `state=present` waits for the resource to exist, as before; `state=ready` also waits for readiness.
- After switching, you can add `external=true` so that only the cluster is searched. After you remove the old annotations, the next deploy may still warn about them once, because the previous release has them.
- The one-key form `<name>.external-dependency.werf.io: <apiVersion>:<kind>[:<namespace>]:<name>` is replaced the same way. See [Resource dependencies]({{ "/reference/deploy_annotations.html#resource-dependencies" | true_relative_url }}).

**`<name>.dependency.werf.io`** is now ignored without a warning. Replace it with `werf.io/deploy-dependency-<name>`, which v2 understands too, and add `external=false` together with switching (v2 rejects it):

```yaml
# Before
db.dependency.werf.io: apps/v1:StatefulSet:postgres
# After
werf.io/deploy-dependency-db: state=present,group=apps,version=v1,kind=StatefulSet,name=postgres,namespace={% raw %}{{ .Release.Namespace }}{% endraw %}
```

**Annotation keys like `werf-io/weight` or `helm-sh/hook`** worked in v2 as `werf.io/weight` and `helm.sh/hook`. They are now ordinary annotations: rename them to the real keys.

**Explicit `null` values** are now removed from manifests. Set a real value where a missing field means something else to Kubernetes or a CRD.

**Sensitive fields in diffs.** For `Secret` resources and resources annotated `werf.io/sensitive: "true"`, v2 diffs hid all values; v3 hides only `data.*` and `stringData.*`. **Before the first v3 plan**, list every field to hide in `werf.io/sensitive-paths`, including `$.data.*` and `$.stringData.*` if needed: the annotation replaces the default list. It works in v2 too, where it immediately replaces hiding the whole resource. Check the first plans in private output, and don't use `--show-sensitive-diffs` where logs are shared.

```yaml
annotations:
  werf.io/sensitive-paths: "$.spec.token,$.data.*,$.stringData.*"
```

**Deletion propagation** is now `Background`: werf doesn't wait for dependents of a deleted resource, such as its Pods. If a later step needs them gone, set `werf.io/delete-propagation: Foreground` on the resource (works in v2 too), or `--delete-propagation=Foreground` for all resources.

**`.Capabilities`.** `.Capabilities.HelmVersion` reports Helm 4: update templates that check it. Without cluster access, `werf render`, `werf lint` and `werf bundle render` assume Kubernetes `1.36.0` instead of `1.20.0`; pass `--kube-version` if templates depend on it.

#### Charts: values and secrets

**Keys set in both secret and non-secret values.** In v2, the chart's `secret-values.yaml` and `--secret-values` overrode `--values` and all `--set*`. v3 orders values as documented, from lowest to highest: `values.yaml`, the chart's `secret-values.yaml`, `--values`, `--secret-values`, `--set-json`, `--set`, `--set-string`, `--set-file`, `--set-literal`, werf's service values. To keep the secret value on both versions, remove the same key from `--values` and `--set*`. See [Parameter sources and their priority]({{ "/usage/deploy/values.html#parameter-sources-and-their-priority" | true_relative_url }}).

**Decrypted secrets in the release storage.** v3 stores the decrypted `--secret-values` in the release like other values, and `werf bundle apply` stores the publish-time `--secret-values` from `.werf-data/secret-values.yaml` of a bundle the same way: `werf release get --print-values` shows them, and changing only `--secret-values` creates a new revision. All parameters from the files are stored, including the ones that the templates do not use: if one file keeps the secrets of several environments, a release in staging stores the production secrets too. If `Chart.yaml` of the main chart lists `dependencies`, the release also stores all decrypted parameters from `secret-values.yaml` of the main chart and of the dependent charts. Before switching, split such files by environment and pass only the file of the current environment (if `Chart.yaml` of the main chart lists `dependencies`, move the secrets of the environments out of `secret-values.yaml` as well), and restrict access to the release storage (Secrets, ConfigMaps or SQL) and to `werf release get` output as you do for the secrets themselves. See [Secret parameters in the release]({{ "/usage/deploy/values.html#secret-parameters-in-the-release" | true_relative_url }}).

**`export-values` and `import-values`** now also apply to `--values`, `--secret-values`, `--set*` and werf's service values. A subchart key you set directly now loses to an exported parent value from the same or a higher source: `--set backend.config.greeting=x` loses to an exported `--set backendConfig.greeting=y`. Set the parent key instead, or remove the export.

**`global.werf` service values** are always set now (`WERF_EXPERIMENT_NO_GLOBAL_SERVICE_VALUES` is removed). If you used that variable, check that schemas accept them.

**New secret values keep their type.** Values entered as unquoted numbers or booleans reach templates as numbers or booleans: quote them if templates expect strings. Existing encrypted values stay strings. Comments on encrypted values are stored in cleartext: don't put secrets in them.

#### Charts: files werf now reads differently

**`.helmignore`** is now applied, and excluded files disappear from rendered manifests and bundles without a warning. Review it in charts and subcharts. Unlike `.gitignore`, a leading `!` keeps only what matches; `**` is an error; Helm's default rules exclude dot-prefixed files and directories directly under `templates/` even without `.helmignore`. See [Excluding files or directories from the chart]({{ "/usage/deploy/charts.html#excluding-files-or-directories-from-the-chart" | true_relative_url }}).

**`patches.yaml`** in the root of a chart or subchart is now read as [resource patches]({{ "/usage/deploy/resource_patching.html" | true_relative_url }}). Rename a file with this name if it's for something else, or pass `--no-default-patches` in v3.

**`.werf-data/`** in the chart root is now read as bundle data and is never excluded by `.helmignore`. Rename an unrelated directory with this name.

### Prepare CI

Removed commands and flags fail only when CI runs v3, so prepare these changes in advance.

#### CI: `werf helm` commands

Deploy and release commands under `werf helm` are removed.

| v2 | v3 | Works in v2 |
|---|---|---|
| `werf helm install`, `werf helm upgrade` | `werf converge`, which installs or upgrades as needed | Yes |
| `werf helm template` | `werf render` | Yes |
| `werf helm lint` | `werf lint` | Yes |
| `werf helm uninstall NAME` | `werf dismiss --release NAME --namespace NAMESPACE` | Yes |
| `werf helm rollback NAME REVISION` | `werf rollback --release NAME --namespace NAMESPACE --revision REVISION` | Yes |
| `werf helm list` | `werf release list` | No |
| `werf helm get …`, `werf helm status` | `werf release get --release NAME --namespace NAMESPACE` | No |
| `werf helm history NAME` | `werf release history --release NAME --namespace NAMESPACE` | No |
| `werf helm test`, `werf helm plugin` | No replacement in werf or Nelm: use the Helm CLI | — |

Replace the commands that work in v2 now. Switch the others together with the binary. For a chart without a werf project, use Nelm.

`werf converge`, `render` and `lint` take the werf project instead of Helm's `RELEASE CHART` arguments: pass `--release NAME --namespace NAMESPACE` to keep the old target. Their output, and that of `werf release`, differs from `werf helm`:

- `werf release get` prints the whole release, including Secret data: don't save its output where logs are shared. With `--print-values` it adds `.values`: the chart defaults merged with the stored values, including decrypted `--secret-values`, not only the values you passed as with `werf helm get values`. Get them with `--output-format json | jq .values`.
- `werf release list` lists all namespaces if `--namespace` isn't set. Use `--output-format json` in scripts.
- Don't pass werf's service values to `werf render` yourself (`--set werf.image.app=...`): a strict schema rejects them. Use `--stub-tags`, a real `--repo`, or `--without-images` if the chart renders without image references.

**Other `werf helm` commands** (`secret`, `dependency`, `repo` and others) remain, but:

- they no longer accept Kubernetes connection flags (see the [reference](#flags)), but `werf helm get-release` and `werf helm get-autogenerated-values` keep their own `--namespace`; Helm-based subcommands ignore `--log-*`, use `HELM_DEBUG=true`;
- `werf helm registry login` no longer gives builds and deploys registry access: use `werf cr login` or `docker login`, which work in v2 too;
- `werf helm secret rotate-secret-key` is removed: re-encrypt each file, see [Rotating the secret key]({{ "/usage/deploy/values.html#rotating-the-secret-key" | true_relative_url }});
- `werf helm repo update` fails if any repository fails.

**werf no longer works as `helm`** through a `helm` symlink or `WERF_HELM3_MODE`. Use Nelm instead, or the Helm CLI where a tool needs the `helm` binary.

#### CI: flags, variables and environment

Removed flags are argument errors in v3; the [reference](#reference-removed-commands-flags-and-variables) lists them all. Flags that did nothing in v2 (marked "had no effect" there, plus `--force-adoption` on render commands and registry API credential flags on commands other than `cleanup`, `purge`, `managed-images rm` and `meta-repo detach`) can be removed now. Remove these together with switching the binary, because they change what v2 does:

- `--virtual-merge`, `--skip-image-spec-stage`, `--show-verbose-diffs`.
- `--with-hooks` on `werf dismiss`: v2 deletes hooks by default; v3 never deletes them on dismiss. Where hooks must go with the release, set `werf.io/ownership: release` on them, delete them separately, or use `--with-namespace`.
- `--helm-compatible-chart` and `--rename-chart` on `werf bundle publish`: see [Bundles](#bundles).
- `--set-runtime-json` is removed, and v3 no longer sets `.Runtime`. **Also unset `WERF_SET_RUNTIME_JSON` and `WERF_SET_RUNTIME_JSON_*`**: v3 would read them as `--set` values. Replace it with one of:

  {% raw %}
  ```shell
  # v2. Template: {{ $.Runtime.ci.job }}
  --set-runtime-json 'ci={"job":"42"}'

  # v3, templates unchanged (works in subcharts too). Template: {{ $.Runtime.ci.job }}
  --set-root-json 'Runtime.ci={"job":"42"}'

  # v3, data in values. Template: {{ $.Values.ci.job }}
  --set-json 'ci={"job":"42"}'
  ```
  {% endraw %}

  v2 accepts `--set-root-json` but replaces `.Runtime` with its own, so switch it together with the binary. With `--set-json`, the data is stored in the release, may need to be allowed by `values.schema.json`, and reaches a subchart only under `global` or the subchart's key, such as `backend.ci`.

**Variables to change:**

| v2 | v3 |
|---|---|
| `HELM_DRIVER` | `WERF_RELEASE_STORAGE` (or `--release-storage`); works in v2 too. Without it, werf doesn't find releases stored in ConfigMaps or SQL and creates new ones in Secrets. |
| `HELM_CONFIG_HOME`, `HELM_CACHE_HOME`, `HELM_DATA_HOME`, `HELM_REPOSITORY_*` | Not read when rendering and deploying. Put `repositories.yaml` in `$XDG_CONFIG_HOME/helm/` (by default `~/.config/helm/` on Linux, `~/Library/Preferences/helm/` on macOS), and unset `HELM_REPOSITORY_CONFIG`, `HELM_REPOSITORY_CACHE`, `HELM_CONTENT_CACHE`: `werf helm repo` would still write there. |
| `DOCKER_AUTH_CONFIG` with `werf ci-env` | Now used automatically for the CI session's Docker config instead of copying the existing one. To keep copying the existing config, set `WERF_USE_DOCKER_AUTH_CONFIG=false` (works in v2 too). |
| `WERF_LOG_COLOR_MODE` from `werf ci-env` | No longer exported: set it yourself if CI relied on it. Don't force it to `on` where `werf render` output is parsed: render commands now follow it. |
| `WERF_CONVERGE_ENABLE_IMAGES_PARAMS` | Remove: positional image names of `werf converge` and `werf plan` now always select images. |

**Other changes:**

- **`--runtime-annotations` and `--runtime-labels`** no longer update a resource when only their values change, as werf's service annotations already didn't in v2. Use `--add-annotation`/`--add-label` with your own key for metadata that must reach live resources.
- `werf config list` lists only final images by default: pass `--final-images-only=false` (works in v2 too) where scripts need all images.
- `werf ci-env github` no longer logs in to `docker.pkg.github.com`: use `ghcr.io`.
- `werf kubectl` is kubectl 1.37: `autoscale --cpu-percent` is deprecated in favor of `--cpu=N%` (v2 doesn't accept `--cpu`), and `debug --profile` defaults to `general`, with `legacy` deprecated. `KUBECTL_KUBERC=false` stops a host kuberc file from affecting it.
- macOS 12 or newer and Linux kernel 3.2 or newer are required.

#### CI: exit codes and reports

Update parsers to accept both formats before switching:

- **`werf plan --exit-code` and `werf bundle plan --exit-code`** can return `3`: no resource changes, but the release still has to be installed or updated. Don't treat it as an error or skip the deploy.

  | Code | Meaning |
  |---|---|
  | `0` | No changes. |
  | `1` | Error. |
  | `2` | Resource changes are planned. |
  | `3` | Only the release needs to be installed or updated. |

- **Build report.** When a whole image is reused, the JSON report can have `StagesSkipped: true` and no `Stages`. That's not an error.
- **Deploy, rollback and uninstall reports** (`--save-deploy-report` and others) have `"apiVersion": "v3"` instead of `"version": 3`.
- **Saved plans** from v2 are rejected by v3: run `werf plan --save-plan PATH` again with v3. Plans contain the release values, including decrypted secret values, and are encrypted only if the secret key is passed: store them like secrets.
- **`werf render` output** has a new resource order, no longer adds empty `annotations: {}`/`labels: {}` to resources without them, and includes werf's service annotations. Regenerate saved copies you compare it with.
- **Images** no longer have the `werf.io/base-image-id` label (`werf.io/parent-stage-id` records the parent stage instead), the `werf-import-checksum-*` labels and the `werf-git-*-virtual-merge` labels, and no longer get `COLUMNS`, `WERF_COMMIT_*` and `SSH_AUTH_SOCK` in `ENV`. Images based on v2-built images still inherit them: drop them with `imageSpec.config.removeEnv` and `removeLabels`.

## Upgrading

### Order of the switch

Some effects of v3 stay if you switch back to v2. Switch in this order:

1. **Protect rollback history.** If you deployed with `werf converge --use-plan`, v2 ignored `--releases-history-max` and may have kept all revisions. The first v3 command that creates a revision of such a release removes old revisions down to the limit (5 by default). To keep more for a rollback, raise `--releases-history-max` in every deploy and rollback job before switching them.
2. **Switch runners.** Don't run werf v2 and v3 at the same time on the same runner or with the same container backend storage: v3 host cleanup can remove a lock file while a v2 job is acquiring it, so the lock no longer protects the v2 build, and cleanup can delete its stage images. A separate `WERF_HOME` doesn't help. Wait for v2 jobs on a runner to finish before switching it, or migrate on separate runners.
3. **SQL release storage**, if used: [run the storage migration](#release-storage) before any v3 job uses it.
4. **Upgrade everything that reads secrets or deploys bundles:** developer machines, CI jobs, jobs that apply saved plans, every environment that runs `werf bundle apply`, `plan` or `render`. Switch the prepared CI changes at the same time. With the official werf image, switch to the v3 image.
5. **Only then publish bundles with v3 and edit secrets:** see [Bundles](#bundles) and [Secrets](#secrets).

### Release storage

If releases are stored in SQL (`--release-storage=sql`), v3 adds a `nelm_release_annotations_v1` table and records migrations in `nelm_migrations`. Every v3 command that opens the storage runs the migration, including `plan` and `release list`.

1. Before any v3 job uses the storage, run `werf release list` against it once with a role that can create tables and reference `releases_v1`. Don't create the tables by hand: werf would try to create them again.
2. Row-level security is enabled on `nelm_release_annotations_v1` without policies, so other roles can't read its rows. Create RLS policies for every role that needs these rows, including read-only roles, matching those for `releases_v1`.

### TypeScript charts

TypeScript charts are now rendered with Node.js instead of Deno and are always enabled. v3 can't render a chart built for Deno, so prepare these steps in the v3 branch and switch them together with the binary.

1. **Install Node.js 26 or newer** wherever a chart with `ts/src` is rendered or deployed: CI runners and deploy images. The `ubuntu-node`, `centos-node` and `fedora-node` werf images include it. If `node` isn't in `PATH`, pass `--node-binary-path`.
2. **Move the chart to `@nelm/chart-ts-sdk` 0.2.0:** replace `ts/deno.json` with `ts/package.json` and `ts/tsconfig.json` as `werf chart ts init` creates them, and move the dependencies.
3. **Generate and commit a lock file** with npm, pnpm, yarn or bun (`deno.lock` doesn't count). A clean build of the chart (`werf chart ts build`, `werf bundle publish`) needs that package manager in `PATH` to install dependencies. Deploying a built chart needs neither the package manager nor the lock file, only Node.js.
4. **Rebuild `ts/dist/bundle.js`** and publish bundles again: bundles built for Deno don't render.
5. **Publish a test bundle** and deploy it with v3 before switching production.

Also:

- `NELM_FEAT_TYPESCRIPT` is no longer needed: a chart with `ts/src/index.ts` or `ts/src/index.js` is rendered even if the variable wasn't set in v2. Remove `ts/` if you don't use it.
- `--deno-binary-path` becomes `--node-binary-path`. `werf helm package` no longer builds TypeScript bundles: use `werf bundle publish`.

### Bundles

`werf bundle publish` now publishes the chart unchanged and stores what werf adds in `.werf-data/`: image references, `--values`, encrypted `--secret-values`, `--set*`, patches, annotations and labels, the release name and namespace. `werf bundle apply` gives the same result as `werf converge` with the same options, except for `.Chart.Name` and `.Chart.Version`. See [Bundles and charts]({{ "/usage/distribute/bundles.html" | true_relative_url }}).

**Upgrade consumers before publishing.** werf 2.85.0 and later v2 releases, and Nelm v1 releases after 1.31.3, refuse a bundle published by v3. werf 2.84 and earlier and Nelm 1.31.3 and earlier ignore what `.werf-data/` stores: they deploy the bundle **without images and published values**, or fail later, without saying why.

1. Upgrade every environment that runs `werf bundle apply`, `plan` or `render`, every `werf lint` or `werf converge` of a chart that depends on a bundle, and every Nelm that installs bundles.
2. Publish a test bundle with a value used in a template, and check in each environment that the value reaches the rendered manifests.

v3 still deploys bundles published by v2, except TypeScript charts built for Deno: publish those again.

**Check the chart name before the first apply.** `werf bundle publish` and `werf bundle copy` to a registry now always name the chart after the last part of the repository path, as Helm requires for OCI charts. `.Chart.Name` and template paths change accordingly. **If resource names use `.Chart.Name` and the name changes, the first v3 apply creates resources with new names and deletes the old ones**, unless a resource policy keeps them. If only labels in an immutable selector use it, the apply fails, or recreates the resource if its delete policy allows that.

1. On v2, publish a test bundle with a separate tag, to the production repository or to a test repository whose path ends the same way, with `--helm-compatible-chart` and **without** `--rename-chart` or `WERF_RENAME_CHART`: it gets the same name as in v3. Bundles published this way already have the v3 name.
2. Compare `werf bundle render` of the current bundle and the test one, and review `werf bundle plan` of the test bundle against the existing release before applying.
3. To keep resource names, publish to a repository whose path ends with the old chart name, or don't use `.Chart.Name` in resource names.
4. When switching, remove `--helm-compatible-chart` and `--rename-chart`.

**Helm, Argo CD and Flux** now get only the chart's own values, so image references render empty. Pass `.werf-data/values.yaml` and `.werf-data/extra-values.yaml` as values files: see [Using the chart with Helm, Argo CD and Flux]({{ "/usage/distribute/bundles.html#using-the-chart-with-helm-argo-cd-and-flux" | true_relative_url }}). `.werf-data/values.yaml` may be missing, so allow it. These tools still don't get published `--set*`, secrets, annotations and labels.

**Other changes:**

- **Chart version.** Only a semver `--tag` (`1.2.3`, also `v1.2.3`) becomes the chart version; `1.2` becomes a `0.0.0-<unix time>-1-2` prerelease, which semver ranges in Helm, Argo CD and Flux skip unless prereleases are allowed. The chart is also pushed under its version tag.
- **Tags.** `werf bundle publish` refuses a tag that would overwrite one of the bundle's images in the same repository. Other images and charts in the repository aren't checked: don't reuse their tags.
- **Secret key.** Publishing no longer needs it (`--secret-key` and `--ignore-secret-key` are removed); deploying does.
- **Release name and namespace** come from the bundle: pass `--release` and `--namespace` when applying with a different `--env`.
- **`werf.io/release-channel`** isn't stored any more: pass `--add-annotation werf.io/release-channel=<channel>` to `werf bundle apply`.
- **Archives.** `werf bundle copy --to archive:` writes an [OCI image layout](https://github.com/opencontainers/image-spec/blob/main/image-layout.md) that v2 can't read: use v3 on both sides (v3 still reads v2 archives). Image signatures, SBOMs and other referrers aren't copied: copy them separately. `--platform` and `--container-registry-mirror` are removed.
- **`--disable-default-values`/`--disable-default-secret-values`** of `werf bundle apply` don't disable values stored at publish. If a committed `.werf-data/metadata.yaml` (for example in an unpacked chart in `.helm`) records `defaultValuesDisable` or `defaultSecretValuesDisable` (under `values`) or `defaultPatchesDisable`, publishing keeps it: remove the key to turn the default back on.
- **v2 bundles.** `--runtime-annotations` and `--release-info-annotations` of `werf bundle apply` and `plan` now win over `project.werf.io/*`, `ci.werf.io/*` and `werf.io/release-channel` from the bundle's `extra_annotations.json`.

### Secrets

v3 encrypts secrets with authenticated AES-GCM. **werf v2 and Nelm v1 can't read secrets written by v3, and v3 can't write the old format.** v3 reads old secrets.

1. Upgrade every environment that reads the repository's secrets: developer machines, CI jobs and jobs that apply saved plans.
2. Only then create, edit or rotate secrets. After that, don't use v2 or Nelm v1 with the repository: switching the binary back doesn't restore the old format.

To give an existing value a number or boolean type, enter it again with `werf helm secret values edit`.

### First build and cleanup

**The first v3 build rebuilds all images:** stages built by v2 aren't reused. Plan time and registry space for it.

**Non-final images.** `werf build` and `werf config list` now default to `--final-images-only=true`. `werf build` builds non-final images only if a final image needs them; `werf config list` lists only final images. Pass `--final-images-only=false` to build or list all images. The `--images-only` flag of `werf config list` is removed.

**v3 cleanup can delete v2 images.** Cleanup deletes images that aren't used in Kubernetes. If v2 and v3 use the same `--repo`, the images you need to roll back to v2 are no longer used after the switch. Before the first cleanup:

1. List their tags in a file, one per line. If CI adds mutable custom tags (`--add-custom-tag`), the first v3 build moves them to new images: record the digests or content tags of the v2 images instead.
2. Run cleanup with `--dry-run` and the same repositories and Kubernetes access as the scheduled job, and check that these tags aren't deleted:

   ```shell
   werf cleanup --repo registry.example.com/app --dry-run --keep-list keep-list.txt
   ```

3. Pass `--keep-list keep-list.txt` to the scheduled cleanup.

Pass the same `--final-repo` and `--meta-repo` as for building, if any. `--without-kube` turns off the protection of images used in Kubernetes. Don't remove v2 images with `werf purge`: it ignores retention policies. See [Container registry cleanup]({{ "/usage/cleanup/cr_cleanup.html" | true_relative_url }}).

### First deploy

Before the first `werf converge` with v3:

1. If CI logs are shared, make sure [sensitive fields are hidden](#charts-templates-and-annotations).
2. Compare `werf render` of v2 and v3 for each environment. Values priority, `export-values`, `.helmignore`, `patches.yaml`, removed `null` values and Helm 4 can change manifests without an error.
3. Review `werf plan`. Created and deleted CRDs show a full diff only with `--show-verbose-crd-diffs`.

Releases whose stored values change, for example releases deployed with `--secret-values`, get one new revision.

## After upgrading

### Check the changed defaults

These changes apply without configuration changes. Check them on the first deploys:

| What | v2 | v3 | What to check |
|---|---|---|---|
| Resource validation | Off unless `NELM_FEAT_RESOURCE_VALIDATION=true`. | Resources are checked against schemas of Kubernetes and popular CRDs embedded in the binary. | Fix rejected resources. Skip some with `--resource-validation-skip` (for example `kind=MyKind`), add schemas with `--resource-validation-extra-schema`, or turn validation off with `--no-resource-validation`. |
| Readiness of custom resources | `Ready`/`Available` conditions weren't recognized, so the resource was ready at once. | werf waits for the condition to become `True`. | If a resource never sets it, fix its controller, or set `werf.io/track-termination-mode: NonBlocking` on it. This doesn't make it ready: a `state=ready` dependency on it still waits. |
| Dependencies on resources outside the release | A `werf.io/deploy-dependency-*` or `delete-dependency-*` annotation that matched no release resource was ignored. | If it has `kind`, `version` and `name`, werf waits for the resource in the cluster. | Add `external=false` to keep ignoring it, or remove the annotation. |
| `werf dismiss` | The old implementation unless `WERF_EXPERIMENT_NEW_DISMISS` was set. | Always the new one. | Check release and hook removal in a test environment. `--with-namespace` still deletes the namespace. |
| Images passed to `converge`/`plan` | Ignored unless `WERF_CONVERGE_ENABLE_IMAGES_PARAMS` was set. | Only these images are built and passed to the chart. | All resources are still deployed. |
| OCI charts | Pulled by tag. | Pulled by one manifest digest; provenance, if verified, is checked against it. | To verify provenance, pass `--provenance-strategy=verify` and `--provenance-keyring`. For registries with a self-signed certificate, add the CA to the system trust store. |
| Helm engine | Helm 3. | Helm 4.2. | Compare `werf render` output: template functions, merging of values and chart loading follow Helm 4. |

### Clean up

- Remove the `NELM_FEAT_*` and `WERF_EXPERIMENT*` variables and the other variables listed in the [reference](#environment-variables). Keep `NELM_FEAT_PERIODIC_STACK_TRACES` and `NELM_FEAT_ADOPT_DECKHOUSE_CONTROLLER_FIELDS` if you use them.
- Remove the old external dependency annotations you kept for v2.
- Rename `fromImage`, `import[].image` and `dependencies[].image` to `from` if you haven't yet.
- If you set `WERF_LEGACY_VALUES_GLOBAL_ENV=1`, replace `.Values.global.env` and remove the variable.
- When you no longer need a rollback, remove the v2 images from the keep-list.

### Go module path

The werf Go module path is now `github.com/werf/werf/v3`. Build or install werf from `github.com/werf/werf/v3/cmd/werf`, and replace `github.com/werf/werf/v2/` with `github.com/werf/werf/v3/` in imports.

## Reference: removed commands, flags and variables

### Commands

| Removed | Replacement |
|---|---|
| `werf helm install`, `upgrade`, `template`, `lint`, `uninstall`, `rollback` | See [CI: `werf helm` commands](#ci-werf-helm-commands) |
| `werf helm get`, `history`, `list`, `status` | `werf release get`, `history`, `list` |
| `werf helm test`, `werf helm plugin` | The Helm CLI (werf and Nelm have no equivalent) |
| `werf helm secret rotate-secret-key` | Re-encrypt files one by one |
| `helm` symlink to werf, `WERF_HELM3_MODE` | Nelm, or the Helm CLI where the `helm` binary is required |

### Flags

| Flag | Where | What to do |
|---|---|---|
| `--virtual-merge` | Build and deploy commands | Remove |
| `--skip-image-spec-stage` | Build and deploy commands | Remove; `imageSpec` is always applied |
| `--set-runtime-json` | Deploy commands | `--set-root-json 'Runtime.<key>=...'`, see [CI: flags](#ci-flags-variables-and-environment) |
| `--show-verbose-diffs` | `werf plan`, `werf bundle plan` | Remove; diffs of created and deleted resources are always shown, except CRDs (`--show-verbose-crd-diffs`) |
| `--force-adoption` | `werf render`, `werf bundle render` | Remove; other commands keep it |
| `--helm-compatible-chart`, `-C`, `--rename-chart` | `werf bundle publish`, `werf bundle copy` | Remove; see [Bundles](#bundles) |
| `--secret-key`, `--ignore-secret-key` | `werf bundle publish` | Remove |
| `--platform`, `--container-registry-mirror` | `werf bundle copy` | Remove |
| `--with-hooks` | `werf dismiss` | Remove |
| `--images-only` | `werf config list` | `--final-images-only` |
| `--local-resource-validation`, `--resource-validation-kube-version`, `--resource-validation-schema` | Deploy commands | Remove; use `--resource-validation-extra-schema` for extra schemas |
| `--repo-docker-hub-*`, `--repo-github-token`, `--repo-harbor-*`, `--repo-quay-token` and the `--final-repo-*` ones | Commands other than `cleanup`, `purge`, `managed-images rm`, `meta-repo detach` | Remove; the environment variables can stay |
| `-n`, `--namespace`, `--kube-config`, `--kube-config-base64`, `--kube-context`, `--status-progress-period`, `--hooks-status-progress-period` | `werf helm` subcommands | Remove; `get-release` and `get-autogenerated-values` keep `--namespace` |
| `--no-update`, `--fail-on-repo-update-fail` | `werf helm repo` | Remove |
| `--kube-config`, `--kube-config-base64`, `--kube-context` | `werf meta-repo migrate`, `detach` | Use `--synchronization=kubernetes://NAMESPACE[:CONTEXT][@CONFIG_PATH]` |
| `--docker-compose-bin-path` | `werf compose` | Remove |
| `--deno-binary-path` | Rendering and TypeScript commands | `--node-binary-path` on rendering commands; remove elsewhere |
| `--docker-config`, `--releases-history-max`, `--insecure-helm-dependencies` | `werf helm` (before the subcommand) | Remove; set them on the subcommand where it has them |
| `--insecure-helm-dependencies` | `werf build`, `bundle copy`, `cleanup`, `compose`, `dismiss`, `export`, `managed-images`, `purge`, `run`, `stage image` | Remove; had no effect there |
| `--hooks-status-progress-period` | `werf converge`, `bundle apply`, `dismiss`, `rollback`, `plan`, `render` | Remove; had no effect there |
| `--status-progress-period`, `--releases-history-max` | `werf plan`, `werf render` | Remove; had no effect there |
| `--timeout`, `-t` | `werf plan` | Remove; had no effect there |

### Environment variables

| Variable | What to do |
|---|---|
| `HELM_DRIVER` | `WERF_RELEASE_STORAGE` |
| `HELM_CONFIG_HOME`, `HELM_CACHE_HOME`, `HELM_DATA_HOME`, `HELM_REPOSITORY_*` | `XDG_CONFIG_HOME`, `XDG_CACHE_HOME`, `XDG_DATA_HOME` |
| `DOCKER_BUILDKIT=0` | Remove; BuildKit is always used |
| `WERF_STAGED_DOCKERFILE_VERSION` | Remove |
| `WERF_HELM3_MODE` | Remove |
| `WERF_OLD_SECRET_KEY` | Remove together with `werf helm secret rotate-secret-key` |
| `WERF_DISABLE_GIT_COMMIT_ANCESTRY_CHECK`, `WERF_EXPERIMENTAL_STAPEL_IMPORT_PERMISSIONS`, `WERF_EXPERIMENTAL_STAPEL_ARM` | Remove |
| `WERF_DOCKER_IMAGE_NAME` in the env-format build report | Name the image and read `WERF_<NAME>_DOCKER_IMAGE_NAME` |
| `WERF_EXPERIMENT_NEW_DISMISS`, `WERF_EXPERIMENTAL_IMPORT_BY_SOURCE_IMAGE_TAG`, `WERF_CONVERGE_ENABLE_IMAGES_PARAMS` | Remove; the behavior is always on |
| `WERF_EXPERIMENT_NO_GLOBAL_SERVICE_VALUES` | Remove; `global.werf` service values are always set |
| `NELM_FEAT_PREVIEW_V2`, `NELM_FEAT_RESOURCE_VALIDATION`, `NELM_FEAT_CLEAN_NULL_FIELDS`, `NELM_FEAT_FIELD_SENSITIVE`, `NELM_FEAT_MORE_DETAILED_EXIT_CODE_FOR_PLAN`, `NELM_FEAT_CASE_INSENSITIVE_CONDITION_TRACKING`, `NELM_FEAT_TYPESCRIPT` | Remove; the behavior is always on |
| `WERF_SET_RUNTIME_JSON`, `WERF_SET_RUNTIME_JSON_*` | Unset: v3 would read them as `--set` values. Use `WERF_SET_ROOT_JSON_*` or `WERF_SET_JSON_*`, see [CI: flags](#ci-flags-variables-and-environment) |
| `WERF_DENO_BINARY_PATH` | `WERF_NODE_BINARY_PATH` |
| `WERF_VIRTUAL_MERGE`, `WERF_SKIP_IMAGE_SPEC_STAGE`, `WERF_SHOW_VERBOSE_DIFFS`, `WERF_HELM_COMPATIBLE_CHART`, `WERF_RENAME_CHART`, `WERF_WITH_HOOKS`, `WERF_DOCKER_COMPOSE_BIN_PATH`, `WERF_LOCAL_RESOURCE_VALIDATION`, `WERF_RESOURCE_VALIDATION_KUBE_VERSION`, `WERF_RESOURCE_VALIDATION_SCHEMA` | Remove together with their flags |
