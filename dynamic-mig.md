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
  jq -r '
    [.items[] | select(.spec.driver=="gpu.nvidia.com") | .spec.devices[]?
     | select(.attributes.type.string=="mig")]
    | .[]
    | [.name,
       (.attributes.profile.string // "?"),
       (.attributes.parentUUID.string // "?")]
    | @tsv' | column -t -s $'\t' \
    | { echo -e "DEVICE\tPROFILE\tPARENT_GPU"; cat; }
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

The [`dynamic-mig-samples/`](dynamic-mig-samples/) directory contains sample workloads for different use cases:

| File | Description |
|---|---|
| `case1-any-mig.yaml` | Any MIG slice — scheduler picks freely |
| `case2-capacity-constraints.yaml` | Slice with memory/compute lower bounds (recommended) |
| `case3-exact-profile.yaml` | Exact profile name (e.g. `1g.10gb`) |
| `case4-multiple-slices.yaml` | Multiple slices in one pod, each assigned to a different container |
| `case5-same-gpu.yaml` | Multiple slices pinned to the same parent GPU |

Apply a workload, for example:

```bash
oc apply -f dynamic-mig-samples/case1-any-mig.yaml
```

Once scheduled, the DRA driver automatically enables MIG mode on the GPU and creates the requested slice. Once the pod completes, inspect the logs to confirm the MIG device was visible:

```bash
oc logs -n dynamic-mig-samples mig-any
```

After deleting the workload, the DRA driver tears down the MIG slice:

```bash
oc delete -f dynamic-mig-samples/case1-any-mig.yaml
```

