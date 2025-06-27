#!/bin/bash
set -euo pipefail

script_dir="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
cd $script_dir

if [[ -z "$1" ]]; then
  echo "script requires argument <destination registry>" >&2
  exit 1
fi

if [[ -z "$2" ]]; then
  echo "script requires arguments <werf app>" >&2
  exit 1
fi

if ! command -v crane &>/dev/null; then
  echo "crane not found!" >&2
  exit 1
fi

DEST_SUBREPO=$1/werf
FORCE_PUBLISH=${3:-false}

unset WERF_PLATFORM

export WERF_REPO=ghcr.io/werf/werf-storage

# Extra labels for artifacthub
export WERF_EXPORT_ADD_LABEL_SEPARATOR='\n'
export WERF_EXPORT_ADD_LABEL_AH1=io.artifacthub.package.readme-url=https://raw.githubusercontent.com/werf/werf/main/README.md \
       WERF_EXPORT_ADD_LABEL_AH2=io.artifacthub.package.logo-url=https://raw.githubusercontent.com/werf/website/main/assets/images/werf-logo.svg \
       WERF_EXPORT_ADD_LABEL_AH3=io.artifacthub.package.category=integration-delivery \
       WERF_EXPORT_ADD_LABEL_AH4=io.artifacthub.package.keywords="cli,ci,cd,build,test,deploy,distribute,cleanup" \
       WERF_EXPORT_ADD_LABEL_OC1=org.opencontainers.image.url=https://github.com/werf/werf/tree/main/scripts/werf-in-image \
       WERF_EXPORT_ADD_LABEL_OC2=org.opencontainers.image.source=https://github.com/werf/werf/tree/main/scripts/werf-in-image \
       WERF_EXPORT_ADD_LABEL_OC3=org.opencontainers.image.created=$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
       WERF_EXPORT_ADD_LABEL_OC4=org.opencontainers.image.description="Official image to run werf in containers"

configs=$(werf config list --final-images-only=true)
export_tags=()

if [[ "$FORCE_PUBLISH" == "true" ]]; then
  echo "FORCE_PUBLISH enabled — publishing all images"
  werf export --config "$2.yaml" --tag "$DEST_SUBREPO/$2:%image%"
else
  for config in $configs; do
    tag="$config"

    if crane manifest "$DEST_SUBREPO/$2:$tag" &>/dev/null; then
      echo "Image $tag already exists, skipping..."
      continue
    else
      echo "crane failed to access $tag"
    fi

    echo "Will publish $tag"
    export_tags+=("$tag")
  done

  if [[ ${#export_tags[@]} -eq 0 ]]; then
    echo "Nothing to publish"
  else
    echo "Publishing images: ${export_tags[*]}"
    werf export --config "$2.yaml" "${export_tags[@]}" --tag "$DEST_SUBREPO/$2:%image%"
  fi
fi

