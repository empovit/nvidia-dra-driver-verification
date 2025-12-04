#!/usr/bin/env bash

set -e
set -x

DRA_DRIVER_VERSION=${DRA_DRIVER_VERSION:-"25.8.0"}
FORCE_GPU_SUPPORT=${FORCE_GPU_SUPPORT:-false}
FORCE_GPU_SUPPORT_OPTIONS=""
if [ "$FORCE_GPU_SUPPORT" = true ]; then
    FORCE_GPU_SUPPORT_OPTIONS="--set resources.gpus.enabled=true --set gpuResourcesEnabledOverride=true"
else
    FORCE_GPU_SUPPORT_OPTIONS="--set resources.gpus.enabled=false"
fi


helm repo add nvidia https://helm.ngc.nvidia.com/nvidia \
    && helm repo update

# To use the latest version, remove the --version line
helm install nvidia-dra-driver-gpu nvidia/nvidia-dra-driver-gpu \
    --version="${DRA_DRIVER_VERSION}" \
    --create-namespace \
    --namespace nvidia-dra-driver-gpu \
    --set nvidiaDriverRoot=/run/nvidia/driver \
    ${FORCE_GPU_SUPPORT_OPTIONS}