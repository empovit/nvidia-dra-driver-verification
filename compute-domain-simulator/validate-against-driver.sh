#!/bin/bash
# Validate simulator against real driver

set -e

CONTAINER_CMD="${CONTAINER_CMD:-docker}"
BASE_IMAGE="${BASE_IMAGE:-nvcr.io/nvidia/k8s-dra-driver-gpu:v25.8.0}"
SIMULATOR_IMAGE="${SIMULATOR_IMAGE:-localhost/compute-domain-simulator:latest}"

echo "Validating simulator against driver..."
echo "Using: $CONTAINER_CMD"
echo "Base: $BASE_IMAGE"
echo "Simulator: $SIMULATOR_IMAGE"
echo ""

# Check driver version
DRIVER_VERSION=$($CONTAINER_CMD run --rm --entrypoint="" "$BASE_IMAGE" /busybox/sh -c '/usr/bin/compute-domain-daemon --version' 2>&1 | head -1 || echo "unknown")
echo "✓ Driver version: $DRIVER_VERSION"

# Check simulator binaries - use direct file checks instead of 'which' (not available in BusyBox)
$CONTAINER_CMD run --rm --entrypoint="" "$SIMULATOR_IMAGE" /busybox/sh -c 'test -f /usr/local/bin/nvidia-imex && echo "nvidia-imex found"' > /dev/null || { echo "✗ nvidia-imex not found"; exit 1; }
$CONTAINER_CMD run --rm --entrypoint="" "$SIMULATOR_IMAGE" /busybox/sh -c 'test -f /usr/local/bin/nvidia-imex-ctl && echo "nvidia-imex-ctl found"' > /dev/null || { echo "✗ nvidia-imex-ctl not found"; exit 1; }
echo "✓ Simulator binaries present"

# Check config template
$CONTAINER_CMD run --rm --entrypoint="" "$SIMULATOR_IMAGE" /busybox/sh -c 'test -f /templates/compute-domain-daemon-config.tmpl.cfg' || { echo "✗ Config template not found"; exit 1; }
echo "✓ Config template present"

# Validate critical parameters
CRITICAL_PARAMS="SERVER_PORT IMEX_CMD_PORT IMEX_NODE_CONFIG_FILE BIND_INTERFACE_IP"
for param in $CRITICAL_PARAMS; do
    if ! $CONTAINER_CMD run --rm --entrypoint="" "$BASE_IMAGE" /busybox/sh -c "/busybox/grep -q '^${param}=' /templates/compute-domain-daemon-config.tmpl.cfg" 2>/dev/null; then
        echo "✗ $param missing from config template"
        exit 1
    fi
done
echo "✓ Critical config parameters present"

echo ""
echo "Validation complete. Deploy to test environment to verify."

