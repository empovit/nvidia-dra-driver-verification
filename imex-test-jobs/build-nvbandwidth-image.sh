#!/usr/bin/env bash

set -e

CONTAINER_CMD="${CONTAINER_CMD:-podman}"
IMAGE_NAME="${IMAGE_NAME:-nvbandwidth}"
IMAGE_REGISTRY="${IMAGE_REGISTRY:-localhost}"
IMAGE_TAG="${IMAGE_TAG:-latest}"
FULL_IMAGE="${IMAGE_REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}"
# PLATFORM: Target platform(s) for multi-arch builds (e.g., "linux/amd64" or "linux/amd64,linux/arm64")
# If unset, builds for the current platform only
PLATFORM="${PLATFORM:-}"

if [ -n "${PLATFORM}" ]; then
    if [ "${CONTAINER_CMD}" = "docker" ]; then
        ${CONTAINER_CMD} buildx build --platform ${PLATFORM} -t ${FULL_IMAGE} -f Containerfile.nvbandwidh .
    else
        ${CONTAINER_CMD} build --platform ${PLATFORM} -t ${FULL_IMAGE} -f Containerfile.nvbandwidh .
    fi
else
    ${CONTAINER_CMD} build -t ${FULL_IMAGE} -f Containerfile.nvbandwidh .
fi

