---
title: Debug werf.yaml
permalink: usage/project_configuration/debug_werf_yaml.html
---

## Viewing Template Rendering Results

The `werf config render` command displays the rendered result of `werf.yaml` or a specific image/images.

```bash
$ werf config render
project: demo-app
configVersion: 1
---
image: backend
dockerfile: backend.Dockerfile
```

```bash
$ werf config render backend
image: backend
dockerfile: backend.Dockerfile
```

## Listing All Built Images

The `werf config list` command outputs a list of all images defined in the final `werf.yaml`.

```bash
$ werf config list
backend
frontend
```

The `--final-images-only` flag will display only the final images. You can learn more about final and intermediate images [here]({{ "usage/build/images.html#using-intermediate-and-final-images" | true_relative_url }}).

## Analyzing Image Dependencies

The `werf config graph` command builds a dependency graph between images (or a specific image).

```bash
$ werf config graph
- image: images/argocd
  dependsOn:
    dependencies:
    - images/argocd-source
- image: images/argocd-operator
  dependsOn:
    from: common/distroless
    import:
    - images/argocd-operator-artifact
- image: images/argocd-operator-artifact
- image: images/argocd-artifact
- image: images/argocd-source
  dependsOn:
    import:
    - images/argocd-artifact
- image: common/distroless-artifact
- image: common/distroless
  dependsOn:
    import:
    - common/distroless-artifact
- image: images-digests
  dependsOn:
    dependencies:
    - images/argocd-operator
    - images/argocd
    - common/distroless
- image: python-dependencies
- image: bundle
  dependsOn:
    import:
    - images-digests
    - python-dependencies
- image: release-channel-version-artifact
- image: release-channel-version
  dependsOn:
    import:
    - release-channel-version-artifact
```

```bash
$ werf config graph images-digests
- image: images-digests
  dependsOn:
    dependencies:
    - images/argocd-operator
    - images/argocd
    - common/distroless
```

{% include pages/en/debug_template_flag.md.liquid %}
