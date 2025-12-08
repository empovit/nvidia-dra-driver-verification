#!/bin/bash

CLUSTER_NAME="${1:?Usage: $0 <hosted-cluster-name>}"

NAMESPACE=$(oc get hostedclusters -A --no-headers | grep "$CLUSTER_NAME" | awk '{print $1}')

oc patch hostedcluster "$CLUSTER_NAME" -n "$NAMESPACE" --type=merge -p '{"spec":{"configuration":{"featureGate":{"featureSet":"TechPreviewNoUpgrade"}}}}'

