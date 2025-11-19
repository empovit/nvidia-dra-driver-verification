#!/usr/bin/env bash

set -e
set -x

helm repo add nvidia https://helm.ngc.nvidia.com/nvidia \
    && helm repo update

# To use the latest version, remove the --version="25.8.0" line
helm install nvidia-dra-driver-gpu nvidia/nvidia-dra-driver-gpu \
    --version="25.8.0" \
    --create-namespace \
    --namespace nvidia-dra-driver-gpu \
    --set nvidiaDriverRoot=/run/nvidia/driver \
    --set resources.gpus.enabled=true \
    -f helm-values-override.yaml \
    --set gpuResourcesEnabledOverride=true