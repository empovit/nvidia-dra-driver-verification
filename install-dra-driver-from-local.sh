#!/usr/bin/env bash

set -e
set -x
set -o nounset

echo "${DRA_DRIVER_DIR}"
echo "${TAG}"
echo "${REGISTRY}"

export IMAGE="${IMAGE:-k8s-dra-driver-gpu}"

helm install nvidia-dra-driver-gpu ${DRA_DRIVER_DIR}/deployments/helm/nvidia-dra-driver-gpu\
    --create-namespace \
    --namespace nvidia-dra-driver-gpu \
    --set nvidiaDriverRoot=/run/nvidia/driver \
    --set resources.gpus.enabled=true \
    -f helm-values-override.yaml \
    --set gpuResourcesEnabledOverride=true \
    --set image.tag=${TAG} \
    --set image.repository=${REGISTRY}/${IMAGE}
