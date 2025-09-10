#!/usr/bin/env bash

set -ex
set -o pipefail


export OPERATOR_NAMESPACE="openshift-nmstate"

oc apply -f nmstate-operator-resources.yaml

sleep 15

timeout 120s bash -c "until oc get csv -n ${OPERATOR_NAMESPACE} -o name; do sleep 10; done"
CSV_NAME=$(oc get csv -n ${OPERATOR_NAMESPACE} -o name)
oc wait --for=jsonpath='{.status.reason}'=InstallSucceeded --timeout=360s -n ${OPERATOR_NAMESPACE} ${CSV_NAME}

cat << EOF | oc apply -f -
apiVersion: nmstate.io/v1
kind: NMState
metadata:
  name: nmstate
EOF
