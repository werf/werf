---
title: Migration from v1.2 to v2.0
permalink: resources/migration_from_v1_2_to_v2_0.html
toc: false
---

## Breaking changes in v2.0

Key changes:
1. The new deployment engine Nelm is enabled by default. The old deployment engine cannot be used anymore.
1. Commands `werf converge`, `werf plan` and `werf bundle apply` have better resource validation, which may require fixing your charts.
1. Commands `werf render` and `werf bundle render` now format the resulting YAML, stripping comments, sorting fields and formatting values.
1. Commands `werf render` and `werf bundle render` sort manifests in the resulting YAML differently.
1. Removed commands `werf bundle download` and `werf bundle export`. Use `werf bundle copy --from REPO:TAG --to archive:mybundle.tar.gz`.
1. Renamed flag `--skip-build` to `--require-built-images`.
1. Replaced Helm templating function `werf_image` with {% raw %}`{{ $.Values.werf.image.<MY_IMAGE_NAME> }}`{% endraw %}.
1. Replaced flags `--report-path`, `--report-format` with `--save-build-report`, `--build-report-path`.
1. In command `werf bundle copy` replaced flags `--repo`, `--tag`, `--to-tag` with `--from=REPO`, `--from=REPO:TAG`, `--to=REPO:TAG`.
1. Removed automatic migrations from Helm 2 releases to Helm 3 releases.
    
Other changes:
1. Renamed flags `--repo-implementation`, `--final-repo-implementation` to `--repo-container-registry`, `--final-repo-container-registry`.
1. Removed Selectel Container Registry flags `--repo-selectel-account`, `--repo-selectel-password`, `--repo-selectel-username`, `--repo-selectel-vpc`, `--repo-selectel-vpc-id`, `--final-repo-selectel-account`, `--final-repo-selectel-password`, `--final-repo-selectel-username`, `--final-repo-selectel-vpc`, `--final-repo-selectel-vpc-id`. Use the regular container registry authentication.
1. Special werf annotations such as `werf.io/version` or `project.werf.io/name` are no longer saved in the Helm release (i.e. Secret resource by default), but still applied to the cluster.
