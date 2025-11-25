#!/usr/bin/env bash

set -e
set -x
set -o nounset

FORCE_GPU_SUPPORT=${FORCE_GPU_SUPPORT:-false}

echo "DRA driver directory: ${DRA_DRIVER_DIR}"
echo "Image tag: ${TAG}"
echo "Image registry: ${REGISTRY}"
echo "Force GPU support: ${FORCE_GPU_SUPPORT}"

IMAGE="${IMAGE:-k8s-dra-driver-gpu}"
FORCE_GPU_SUPPORT_OPTIONS=""
if [ "$FORCE_GPU_SUPPORT" = true ]; then
    FORCE_GPU_SUPPORT_OPTIONS="--set resources.gpus.enabled=false --set gpuResourcesEnabledOverride=true"
fi

helm install nvidia-dra-driver-gpu ${DRA_DRIVER_DIR}/deployments/helm/nvidia-dra-driver-gpu\
    --create-namespace \
    --namespace nvidia-dra-driver-gpu \
    --set nvidiaDriverRoot=/run/nvidia/driver \
    ${FORCE_GPU_SUPPORT_OPTIONS} \
    --set image.tag=${TAG} \
    --set image.repository=${REGISTRY}/${IMAGE}
