# Dynamic MIG with the NVIDIA DRA Driver

Dynamic MIG allows MIG (Multi-Instance GPU) slices to be created and destroyed on demand by the DRA driver, without pre-configuring MIG profiles on the node. MIG instances are provisioned when a workload requests one and torn down when the workload is deleted.

Complete the base setup in [README.md](README.md) first, then follow the steps below. A MIG-capable GPU node (NVIDIA Hopper architecture or later) is required.

## 1. Enable DRAPartitionableDevices Feature Gate ([docs](https://docs.redhat.com/en/documentation/openshift_container_platform/4.21/html/config_apis/featuregate-config-openshift-io-v1#spec-11))

```bash
oc apply -f feature-gate-dra-partitionable.yaml
```

**Note:** This sets `featureSet: CustomNoUpgrade` and requires a cluster restart.

## 2. Configure the GPU Operator ClusterPolicy

The following ClusterPolicy settings are required for dynamic MIG:

| Field | Value |
|---|---|
| `spec.devicePlugin.enabled` | `false` |
| `spec.mig.strategy` | `"none"` |
| `spec.cdi.enabled` | `true` |
| `spec.migManager.enabled` | `false` |

Use `./configure-cluster-policy-for-dynamic-mig.sh` to apply these automatically, or configure them manually using any method.

## 3. Install the DRA Driver with Dynamic MIG Enabled

Set `DYNAMIC_MIG=true` when running the driver installation script (see [README.md](README.md)). This enables the `featureGates.DynamicMIG` Helm value and forces GPU resources on (`resources.gpus.enabled=true`, `gpuResourcesEnabledOverride=true`):

```bash
DYNAMIC_MIG=true ./install-dra-driver-from-repo.sh
```

## 4. Verify Dynamic MIG

Verify that ResourceSlices expose MIG devices:

```bash
oc get resourceslices -o json | \
  jq '.items[].spec.devices[] | select(.basic.attributes["gpu.nvidia.com/type"].string == "mig")'
```

Verify the node has no `nvidia.com/gpu` capacity (device plugin is disabled):

```bash
oc get nodes -o json | \
  jq '.items[] | {name: .metadata.name, gpu: .status.capacity["nvidia.com/gpu"]}'
```

Verify MIG mode is not pre-enabled on the GPU (the DRA driver enables it dynamically):

```bash
oc debug node/<gpu-node-name> -- nvidia-smi --query-gpu=mig.mode.current --format=csv,noheader
```

## 5. Run a Dynamic MIG Workload

The [`dynamic-mig-jobs/`](dynamic-mig-jobs/) directory contains sample workloads. Apply the minimal workload:

```bash
oc apply -f dynamic-mig-jobs/mig-workload.yaml
```

The workload requests a `1g.10gb` MIG slice. Once scheduled, the DRA driver automatically enables MIG mode on the GPU and creates the requested slice. Once the pod completes, inspect the logs to confirm the MIG device was visible:

```bash
oc logs mig-test-pod -n dynamic-mig-jobs
```

After deleting the workload, the DRA driver tears down the MIG slice:

```bash
oc delete -f dynamic-mig-jobs/mig-workload.yaml
```

