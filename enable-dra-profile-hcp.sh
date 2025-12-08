#!/bin/bash

CLUSTER_NAME="${1:?Usage: $0 <hosted-cluster-name>}"
VERSION=$(oc get clusterversion version -o jsonpath='{.status.desired.version}' | cut -d. -f2)
NAMESPACE=$(oc get hostedclusters -A --no-headers | grep "$CLUSTER_NAME" | awk '{print $1}')

if [[ "$VERSION" -lt 20 ]]; then
    oc patch hostedcluster "$CLUSTER_NAME" -n "$NAMESPACE" --type=merge -p '{"spec":{"configuration":{"scheduler":{"profile":"HighNodeUtilization","profileCustomizations":{"dynamicResourceAllocation":"Enabled"}}}}}'
else
    oc patch hostedcluster "$CLUSTER_NAME" -n "$NAMESPACE" --type=merge -p '{"spec":{"configuration":{"scheduler":{"profile":"HighNodeUtilization"}}}}'
fi

