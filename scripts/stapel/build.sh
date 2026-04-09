#!/usr/bin/env bash
set -euo pipefail

REGISTRY="${STAPEL_REGISTRY:-registry-write.werf.io/werf}"
TAG="${STAPEL_TAG:-dev}"
DOCKERFILE="${STAPEL_DOCKERFILE:-stapel/Dockerfile}"
IMAGE="${REGISTRY}/stapel:${TAG}"

PLATFORMS="${STAPEL_PLATFORMS:-linux/amd64,linux/arm64}"

BUILDER="${STAPEL_BUILDER:-werf-stapel-builder}"

if ! docker buildx inspect "${BUILDER}" >/dev/null 2>&1; then
  docker buildx create --name "${BUILDER}" --driver docker-container --use
else
  docker buildx use "${BUILDER}"
fi

docker buildx inspect --bootstrap >/dev/null

docker buildx build \
  --file "${DOCKERFILE}" \
  --target final \
  --platform "${PLATFORMS}" \
  --tag "${IMAGE}" \
  --push \
  .

echo "Published: ${IMAGE}"
docker buildx imagetools inspect "${IMAGE}" | sed -n '/Platform:/p'