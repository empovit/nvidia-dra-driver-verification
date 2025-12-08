#!/bin/bash

VERSION=$(oc get clusterversion version -o jsonpath='{.status.desired.version}' | cut -d. -f2)

if [[ "$VERSION" -lt 20 ]]; then
    oc patch scheduler cluster --type=merge -p '{"spec": {"profile": "HighNodeUtilization", "profileCustomizations": {"dynamicResourceAllocation": "Enabled"}}}'
else
    oc patch scheduler cluster --type=merge -p '{"spec": {"profile": "HighNodeUtilization"}}'
fi
