#!/bin/bash
set -euo pipefail

script_dir="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
cd $script_dir

export WERF_REPO=ghcr.io/werf/werf-storage

# Extra labels for artifacthub
export WERF_EXPORT_ADD_LABEL_1=io.artifacthub.package.readme-url=https://raw.githubusercontent.com/werf/werf/main/README.md \
       WERF_EXPORT_ADD_LABEL_2=org.opencontainers.image.created=$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
       WERF_EXPORT_ADD_LABEL_3=org.opencontainers.image.description="Official image to run werf in containers"

werf build

for dest_repo in "ghcr.io/werf" "registry-write.werf.io/werf"; do
  DEST_REPO=$dest_repo

  werf export --tag "$DEST_REPO/werf:latest" "1.2-stable-alpine"
  werf export --tag "$DEST_REPO/werf-argocd-cmp-sidecar:latest" "argocd-cmp-sidecar-1.2-stable-ubuntu"

  for group in "1.2"; do
    werf export --tag "$DEST_REPO/werf:$group" "$group-stable-alpine"
    werf export --tag "$DEST_REPO/werf-argocd-cmp-sidecar:$group" "argocd-cmp-sidecar-$group-stable-ubuntu"

    for distro in "alpine" "ubuntu" "centos" "fedora"; do
      werf export --tag "$DEST_REPO/werf:$group-$distro" "$group-stable-$distro"
    done

    for channel in "alpha" "beta" "ea" "stable" "rock-solid"; do
      werf export --tag "$DEST_REPO/werf:$group-$channel" "$group-$channel-alpine"
      werf export --tag "$DEST_REPO/werf-argocd-cmp-sidecar:$group-$channel" "argocd-cmp-sidecar-$group-$channel-ubuntu"

      for distro in "alpine" "ubuntu" "centos" "fedora"; do
        werf export --tag "$DEST_REPO/werf:$group-$channel-$distro" "$group-$channel-$distro"
      done

      for distro in "ubuntu"; do
        werf export --tag "$DEST_REPO/werf-argocd-cmp-sidecar:$group-$channel-$distro" "argocd-cmp-sidecar-$group-$channel-$distro"
      done
    done
  done
done
