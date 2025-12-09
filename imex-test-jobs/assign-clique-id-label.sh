#!/bin/bash

# WARNING: This is only for simulation/testing purposes. In real deployments, this label
# is automatically set by the kubelet plugin based on actual GPU fabric topology.

set -euo pipefail

LABEL_KEY="nvidia.com/gpu.clique"
LABEL_VALUE="550e8400-e29b-41d4-a716-446655440000.0"

WORKER_NODES=$(oc get nodes -l '!node-role.kubernetes.io/control-plane' -o jsonpath='{.items[*].metadata.name}')

for node in $WORKER_NODES; do
    oc label node "$node" "${LABEL_KEY}=${LABEL_VALUE}"
done

