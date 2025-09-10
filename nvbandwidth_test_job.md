# Explain nvbandwidth-test-job-multi-node.yaml
_Exported on 9/10/2025 at 10:10:36 GMT+3 from Cursor (1.5.11)_

---

**User**

Explain me iin detail @nvbandwidth-test-job-multi-node.yaml 

---

**Cursor**

I'll explain the `nvbandwidth-test-job-multi-node.yaml` file in detail. This YAML file defines a multi-node GPU bandwidth testing job using NVIDIA's Dynamic Resource Allocation (DRA) and MPI (Message Passing Interface).

## Overview
This file contains two main Kubernetes resources that work together to run a multi-node nvbandwidth test across GPU nodes in a cluster.

## 1. ComputeDomain Resource (Lines 1-11)

```yaml
apiVersion: resource.nvidia.com/v1beta1
kind: ComputeDomain
metadata:
  name: nvbandwidth-test-compute-domain
spec:
  numNodes: 2
  channel:
    resourceClaimTemplate:
      name: nvbandwidth-test-compute-domain-channel
```

**Purpose**: This defines a custom NVIDIA resource that establishes a compute domain for multi-node GPU communication.

**Key fields**:
- `numNodes: 2`: Specifies that this compute domain spans exactly 2 nodes
- `channel.resourceClaimTemplate.name`: Creates a template for resource claims that will enable high-performance communication channels between GPUs across the nodes
- This ComputeDomain ensures GPUs on different nodes can communicate efficiently using technologies like NVLink or InfiniBand

## 2. MPIJob Resource (Lines 12-93)

```yaml
apiVersion: kubeflow.org/v2beta1
kind: MPIJob
```

**Purpose**: This defines an MPI (Message Passing Interface) job that orchestrates the distributed nvbandwidth test across multiple nodes.

### MPIJob Specification (Lines 16-22)

```yaml
spec:
  slotsPerWorker: 4
  launcherCreationPolicy: WaitForWorkersReady
  runPolicy:
    cleanPodPolicy: Running
  sshAuthMountPath: /home/mpiuser/.ssh
```

- `slotsPerWorker: 4`: Each worker node will use 4 GPU "slots" (4 GPUs per node)
- `launcherCreationPolicy: WaitForWorkersReady`: The launcher pod won't start until all worker pods are ready
- `cleanPodPolicy: Running`: Only clean up pods that are currently running (not failed/completed ones)
- `sshAuthMountPath`: MPI uses SSH for inter-node communication; this specifies where SSH keys are mounted

### Launcher Configuration (Lines 23-56)

The launcher is the "master" node that coordinates the MPI job:

**Node Affinity** (Lines 31-37):
```yaml
affinity:
  nodeAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
      nodeSelectorTerms:
      - matchExpressions:
        - key: node-role.kubernetes.io/control-plane
          operator: Exists
```
- Forces the launcher to run on a control plane node (master node)
- This separates orchestration from compute workloads

**Container Configuration**:
- `image: ghcr.io/nvidia/k8s-samples:nvbandwidth-v0.7-8d103163`: Uses NVIDIA's nvbandwidth testing image
- `runAsUser: 1000`: Runs as non-root user for security

**MPI Command** (Lines 44-56):
```yaml
command: [mpirun]
args:
  - --bind-to core
  - --map-by ppr:4:node  
  - -np "8"
  - --report-bindings
  - -q
  - nvbandwidth
  - -t multinode_device_to_device_memcpy_read_ce
```

**MPI Arguments Explained**:
- `--bind-to core`: Bind MPI processes to CPU cores for consistent performance
- `--map-by ppr:4:node`: Map 4 processes per node (matches 4 GPUs per node)
- `-np "8"`: Total of 8 MPI processes (4 processes × 2 nodes)
- `--report-bindings`: Show which cores processes are bound to
- `-q`: Quiet mode (reduce verbose output)
- `nvbandwidth`: The actual bandwidth testing program
- `-t multinode_device_to_device_memcpy_read_ce`: Test type for multi-node device-to-device memory copy operations using Copy Engine

### Worker Configuration (Lines 57-93)

The workers are the compute nodes that actually run the bandwidth tests:

**Replica Configuration**:
- `replicas: 2`: Creates 2 worker pods (one per node)

**Pod Affinity** (Lines 64-73):
```yaml
affinity:
  podAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
    - labelSelector:
        matchExpressions:
        - key: nvbandwidth-test-replica
          operator: In
          values: [mpi-worker]
      topologyKey: nvidia.com/gpu.clique
```
- Ensures worker pods are scheduled on nodes with GPUs in the same "clique"
- `nvidia.com/gpu.clique`: A topology key that groups nodes with high-bandwidth GPU interconnects

**Container Configuration**:
- Same base image as launcher
- `command: [/usr/sbin/sshd]` with args `[-De, -f, /home/mpiuser/.sshd_config]`: Runs SSH daemon for MPI communication

**Resource Requirements** (Lines 86-90):
```yaml
resources:
  limits:
    nvidia.com/gpu: 4
  claims:
  - name: compute-domain-channel
```
- Requests exactly 4 GPUs per worker node
- Claims the compute domain channel for inter-node GPU communication

**Resource Claims** (Lines 91-93):
```yaml
resourceClaims:
- name: compute-domain-channel
  resourceClaimTemplateName: nvbandwidth-test-compute-domain-channel
```
- Links back to the ComputeDomain resource defined at the beginning
- This enables the DRA system to set up proper GPU-to-GPU communication channels

## How It All Works Together

1. **ComputeDomain** sets up a 2-node GPU communication domain with optimized channels
2. **MPIJob launcher** coordinates the distributed test from a control plane node
3. **MPI workers** (2 replicas) each claim 4 GPUs and join the compute domain
4. **MPI processes** (8 total: 4 per node) perform bandwidth tests between GPUs across nodes
5. **Test type** `multinode_device_to_device_memcpy_read_ce` specifically measures GPU-to-GPU memory transfer performance across nodes using Copy Engine

This setup is designed to validate that NVIDIA's DRA driver properly enables high-performance, multi-node GPU communication for demanding HPC and AI workloads.

```yaml
apiVersion: resource.nvidia.com/v1beta1
kind: ComputeDomain
metadata:
  name: nvbandwidth-test-compute-domain
spec:
  numNodes: 2
  channel:
    resourceClaimTemplate:
      name: nvbandwidth-test-compute-domain-channel
```

```yaml
apiVersion: kubeflow.org/v2beta1
kind: MPIJob
```

```yaml
spec:
  slotsPerWorker: 4
  launcherCreationPolicy: WaitForWorkersReady
  runPolicy:
    cleanPodPolicy: Running
  sshAuthMountPath: /home/mpiuser/.ssh
```

```yaml
affinity:
  nodeAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
      nodeSelectorTerms:
      - matchExpressions:
        - key: node-role.kubernetes.io/control-plane
          operator: Exists
```

```yaml
command: [mpirun]
args:
  - --bind-to core
  - --map-by ppr:4:node  
  - -np "8"
  - --report-bindings
  - -q
  - nvbandwidth
  - -t multinode_device_to_device_memcpy_read_ce
```

```yaml
affinity:
  podAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
    - labelSelector:
        matchExpressions:
        - key: nvbandwidth-test-replica
          operator: In
          values: [mpi-worker]
      topologyKey: nvidia.com/gpu.clique
```

```yaml
resources:
  limits:
    nvidia.com/gpu: 4
  claims:
  - name: compute-domain-channel
```

```yaml
resourceClaims:
- name: compute-domain-channel
  resourceClaimTemplateName: nvbandwidth-test-compute-domain-channel
```

