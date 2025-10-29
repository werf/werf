#!/bin/bash
set -euo pipefail

NEW="trdl_channels.yaml"
OLD="trdl_channels_old.yaml"

if [[ ! -f "$NEW" ]]; then
  echo "ERROR: $NEW not found!"
  exit 1
fi

changed_channels=()
new_versions=()

group_count=$(yq e '.groups | length' "$NEW")

for i in $(seq 0 $((group_count - 1))); do
  group=$(yq e ".groups[$i].name" "$NEW")
  channel_count=$(yq e ".groups[$i].channels | length" "$NEW")

  for j in $(seq 0 $((channel_count - 1))); do
    channel=$(yq e ".groups[$i].channels[$j].name" "$NEW")
    new_version=$(yq e ".groups[$i].channels[$j].version" "$NEW")
    old_version=$(yq e ".groups[] | select(.name == \"$group\") | .channels[] | select(.name == \"$channel\") | .version" "$OLD" || true)

    if [ "$new_version" != "$old_version" ]; then
      echo "Channel changed: $group:$channel: $old_version → $new_version"
      changed_channels+=("$group:$channel:$new_version")
      new_versions+=("$new_version")
    fi
  done
done

if [ ${#new_versions[@]} -eq 0 ]; then
  echo "No version changes found — nothing to publish"
  echo "versions=" >> "$GITHUB_OUTPUT"
  exit 0
fi

# Соберём бинарники и загрузим в релизы
for v in "${new_versions[@]}"; do
  echo "=== Building binary for $v ==="
  mkdir -p dist/$v/linux-amd64/bin
  export TASK_X_REMOTE_TASKFILES=1
  task --yes build:dev:linux:amd64 \
    outputDir="dist/$v/linux-amd64/bin" \
    extraGoBuildArgs="-ldflags='-s -w'" \
    pkg=./cmd/delivery-kit

  BIN="dist/$v/linux-amd64/bin/delivery-kit"
  cp "$BIN" "delivery-kit-linux-amd64"

  TAG="v$v"

  if gh release view "$TAG" >/dev/null 2>&1; then
    echo "Uploading binary to existing release $TAG"
    gh release upload "$TAG" delivery-kit-linux-amd64 --clobber --name "delivery-kit-linux-amd64"
  else
    echo "Creating new prerelease $TAG"
    gh release create "$TAG" \
      --title "$TAG [pre]" \
      --prerelease
    gh release upload "$TAG" delivery-kit-linux-amd64 --clobber --name "delivery-kit-linux-amd64"
  fi
done

# Displaying modified versions for GitHub Actions
CHANGED_INLINE=$(echo "${new_versions[@]}" | tr ' ' ',')
echo "versions=$CHANGED_INLINE" >> "$GITHUB_OUTPUT"
