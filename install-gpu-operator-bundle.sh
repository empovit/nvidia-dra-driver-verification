#!/usr/bin/env bash

set -e
set -o pipefail
set -o nounset

echo "${BUNDLE_IMAGE}"
export GPU_OPERATOR_NAMESPACE="nvidia-gpu-operator"

# Create namespace with pod-security label using oc apply
echo "Creating/updating namespace $GPU_OPERATOR_NAMESPACE with pod-security label"
cat <<EOF | oc apply -f -
apiVersion: v1
kind: Namespace
metadata:
  name: $GPU_OPERATOR_NAMESPACE
  labels:
    pod-security.kubernetes.io/enforce: privileged
EOF

"$(dirname "$0")/operator-sdk" run bundle --timeout=10m -n $GPU_OPERATOR_NAMESPACE --install-mode OwnNamespace ${BUNDLE_IMAGE}
