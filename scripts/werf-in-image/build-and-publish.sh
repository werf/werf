#!/bin/bash
set -euo pipefail

BASE_REPO="ghcr.io/werf"

script_dir="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
cd $script_dir

export WERF_REPO="$BASE_REPO/werf"

werf build

werf export --tag "$BASE_REPO/werf:latest" "1.2-stable-alpine"
werf export --tag "$BASE_REPO/werf-argocd-cmp-sidecar:latest" "argocd-cmp-sidecar-1.2-stable-ubuntu"

for group in "1.2"; do
  werf export --tag "$BASE_REPO/werf:$group" "$group-stable-alpine"
  werf export --tag "$BASE_REPO/werf-argocd-cmp-sidecar:$group" "argocd-cmp-sidecar-$group-stable-ubuntu"

  for channel in "alpha" "beta" "ea" "stable"; do
    werf export --tag "$BASE_REPO/werf:$group-$channel" "$group-$channel-alpine"
    werf export --tag "$BASE_REPO/werf-argocd-cmp-sidecar:$group-$channel" "argocd-cmp-sidecar-$group-$channel-ubuntu"

    for distro in "alpine" "ubuntu" "centos"; do
      werf export --tag "$BASE_REPO/werf:$group-$channel-$distro" "$group-$channel-$distro"
    done

    for distro in "ubuntu"; do
      werf export --tag "$BASE_REPO/werf-argocd-cmp-sidecar:$group-$channel-$distro" "argocd-cmp-sidecar-$group-$channel-$distro"
    done
  done
done
