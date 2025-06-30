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

if [[ "$FORCE_PUBLISH" == "true" ]]; then
  echo "FORCE_PUBLISH is true — exporting all images"
  export --config "$2.yaml" --tag "$DEST_SUBREPO/$2:%image%"
  exit 0
fi

NEW="../../trdl_channels.yaml"
OLD="../../trdl_channels_old.yaml"

changed_channels=()
new_versions=()
glob_patterns=()

group_count=$(yq e '.groups | length' "$NEW")

for i in $(seq 0 $((group_count - 1))); do
  group=$(yq e ".groups[$i].name" "$NEW")
  channel_count=$(yq e ".groups[$i].channels | length" "$NEW")

  for j in $(seq 0 $((channel_count - 1))); do
    channel=$(yq e ".groups[$i].channels[$j].name" "$NEW")
    new_version=$(yq e ".groups[$i].channels[$j].version" "$NEW")
    old_version=$(yq e ".groups[] | select(.name == \"$group\") | .channels[] | select(.name == \"$channel\") | .version" "$OLD")

    if [ "$new_version" != "$old_version" ]; then
      echo "Channel changed: $group:$channel: $old_version → $new_version"
      changed_channels+=("$group:$channel:$new_version")
      new_versions+=("$new_version")
    fi
  done
done

for v in "${new_versions[@]}"; do
  glob_patterns+=("$v*")
done

for ch in "${changed_channels[@]}"; do
  group=$(echo "$ch" | cut -d: -f1)
  channel=$(echo "$ch" | cut -d: -f2)

  glob_patterns+=("$group-$channel*")

  if [[ "$channel" == "stable" ]]; then
    glob_patterns+=("$group")
    glob_patterns+=("latest")
  fi
done

glob_patterns=($(printf "%s\n" "${glob_patterns[@]}" | sort -u))
export_globs=$(IFS=" "; echo "${glob_patterns[*]}")
werf export $export_globs --config "$2.yaml" --tag "$DEST_SUBREPO/$2:%image%" 