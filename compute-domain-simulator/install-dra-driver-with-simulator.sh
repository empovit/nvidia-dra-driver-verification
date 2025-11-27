#!/usr/bin/env bash
#
# Simple DRA Driver Installation with Compute Domain Support
#

set -e
set -x

# Configuration
# Read version from file if not set
if [ -z "${DRA_DRIVER_VERSION:-}" ]; then
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    DRA_DRIVER_VERSION=$(cat "${SCRIPT_DIR}/VERSION" 2>/dev/null || echo "25.8.0")
fi
SIMULATOR_REGISTRY=${SIMULATOR_REGISTRY:-"ghcr.io/empovit"}
SIMULATOR_IMAGE=${SIMULATOR_IMAGE:-"compute-domain-simulator"}
SIMULATOR_TAG=${SIMULATOR_TAG:-"v${DRA_DRIVER_VERSION}"}

echo "Installing NVIDIA DRA Driver v${DRA_DRIVER_VERSION} with simulator image ${SIMULATOR_REGISTRY}/${SIMULATOR_IMAGE}:${SIMULATOR_TAG}..."

# Check for required node labels
echo ""
echo "⚠️  IMPORTANT: Label worker nodes for the DRA driver to schedule:"
echo "   oc label nodes <node-name> nvidia.com/gpu.present=true"
echo ""

# Add NVIDIA Helm repository
helm repo add nvidia https://helm.ngc.nvidia.com/nvidia || true
helm repo update

# Install DRA driver with simulator image
echo "Installing DRA driver with simulator..."
helm install nvidia-dra-driver-gpu nvidia/nvidia-dra-driver-gpu \
    --version="${DRA_DRIVER_VERSION}" \
    --create-namespace \
    --namespace nvidia-dra-driver-gpu \
    --set image.repository="${SIMULATOR_REGISTRY}/${SIMULATOR_IMAGE}" \
    --set image.tag="${SIMULATOR_TAG}"

