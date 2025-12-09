#!/usr/bin/env bash

set -e
set -x
set -o nounset

# Validate required variables
if [ -z "${DRA_DRIVER_DIR:-}" ]; then
    echo "Error: DRA_DRIVER_DIR is required (path to k8s-dra-driver repository)"
    exit 1
fi

# Set defaults from Helm chart configuration
REGISTRY=${REGISTRY:-"nvcr.io/nvidia"}
TAG=${TAG:-"v25.8.0"}
IMAGE=${IMAGE:-"k8s-dra-driver-gpu"}
FORCE_GPU_SUPPORT=${FORCE_GPU_SUPPORT:-false}
COMPUTE_DOMAIN_TEST_MODE=${COMPUTE_DOMAIN_TEST_MODE:-false}

echo "DRA driver directory: ${DRA_DRIVER_DIR}"
echo "Image tag: ${TAG}"
echo "Image registry: ${REGISTRY}"
echo "Image name: ${IMAGE}"
echo "Force GPU support: ${FORCE_GPU_SUPPORT}"
echo "Compute domain test mode: ${COMPUTE_DOMAIN_TEST_MODE}"
FORCE_GPU_SUPPORT_OPTIONS=""
if [ "$FORCE_GPU_SUPPORT" = true ]; then
    FORCE_GPU_SUPPORT_OPTIONS="--set resources.gpus.enabled=true --set gpuResourcesEnabledOverride=true"
else
    FORCE_GPU_SUPPORT_OPTIONS="--set resources.gpus.enabled=false"
fi

COMPUTE_DOMAIN_TEST_MODE_OPTIONS=""
if [ "$COMPUTE_DOMAIN_TEST_MODE" = true ]; then
    COMPUTE_DOMAIN_TEST_MODE_OPTIONS="--set resources.computeDomains.testingMode=true"
fi

# Override controller.affinity to allow scheduling on worker nodes
# By default, the controller requires control-plane nodes, but this cluster only has worker nodes
helm install nvidia-dra-driver-gpu ${DRA_DRIVER_DIR}/deployments/helm/nvidia-dra-driver-gpu\
    --create-namespace \
    --namespace nvidia-dra-driver-gpu \
    --set nvidiaDriverRoot=/run/nvidia/driver \
    ${FORCE_GPU_SUPPORT_OPTIONS} \
    ${COMPUTE_DOMAIN_TEST_MODE_OPTIONS} \
    --set image.tag=${TAG} \
    --set image.repository=${REGISTRY}/${IMAGE} \
    --set controller.affinity=null
