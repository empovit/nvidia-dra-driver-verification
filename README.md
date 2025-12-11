# NVIDIA DRA Driver Verification

This repository provides scripts and tools for verifying the [NVIDIA DRA (Dynamic Resource Allocation) Driver for GPUs](https://github.com/NVIDIA/k8s-dra-driver-gpu/) on OpenShift.

## Prerequisites

- OpenShift 4.20 cluster with GPU nodes
- Cluster administrator privileges (`cluster-admin` role)
- `oc` CLI tool
- Helm 3.x (for DRA driver installation)

## Installation Overview

1. Enable DRA as it is disabled by default in OpenShift 4.20:
   * Enable `TechPreviewNoUpgrade` feature set
   * Enable `HighNodeUtilization` scheduler profile
2. Install the NVIDIA GPU operator for NVIDIA drivers and GPU allocation:
   * Install the Node Feature Discovery (NFD) operator and create a `NodeFeatureDiscovery` CR.
   * Install the NVIDIA GPU Operator and create a `ClusterPolicy` CR.
3. Install the NVIDIA DRA driver for DRA features.
4. Apply additional Security Context Constraints (SCC) to work around security-related errors on OpenShift. This will be fixed later in the DRA driver's Helm chart.

## Setup Instructions

Follow these steps in order to get the NVIDIA DRA driver up and running in your cluster:

### 1. Enable DRA

Enable the Dynamic Resource Allocation feature gate as part of the `TechPreviewNoUpgrade` feature set on your cluster:

```bash
oc apply -f feature-gate-dra.yaml
```
**Warning**: Enabling the `TechPreviewNoUpgrade` feature set is not reversible, and will prevent future upgrades on the cluster. Not to be done on production clusters!

**Note:** This requires cluster restart and may take several minutes to propagate.

After the cluster restart, verify that the DRA API is available:

```bash
oc api-resources --api-group='resource.k8s.io'
```

After enabling the DRA feature gate, you must configure the correct scheduler profile:

```bash
./enable-dra-profile.sh
```

### 2. Install the NVIDIA GPU Operator


First, install the Node Feature Discovery (NFD) Operator:

```bash
./install-nfd-operator.sh
```

Then install the NVIDIA GPU Operator. Choose **one** of the following installation methods:

#### Option A: Using OLM (Operator Lifecycle Manager)

```bash
./install-gpu-operator-olm.sh
```

#### Option B: Using Bundle Installation

```bash
# Set the bundle image (required)
export BUNDLE_IMAGE="your-bundle-image:tag"
./install-gpu-operator-bundle.sh

# Example with pre-release version:
export BUNDLE_IMAGE="ghcr.io/nvidia/gpu-operator/gpu-operator-bundle:main-latest"
./install-gpu-operator-bundle.sh
```

**Note:** The `BUNDLE_IMAGE` environment variable is required for bundle installation.

### 3. Install the NVIDIA DRA Driver

Choose **one** of the following installation methods:

#### Option A: Install from NVIDIA Helm Repository (Recommended)

```bash
# Use default version
./install-dra-driver-from-repo.sh

# Or use a specific version
export DRA_DRIVER_VERSION="25.8.0"
./install-dra-driver-from-repo.sh
```

#### Option B: Install from Local Build

```bash
# First build the driver image
export DRA_DRIVER_DIR="/path/to/k8s-dra-driver-gpu"
export REGISTRY="your-registry.com/username"
export TAG="your-tag"
./build-dra-driver-image.sh

# Then install from local image
./install-dra-driver-from-local.sh
```

### 4. Apply OpenShift Role Bindings

After installing the NVIDIA DRA driver, apply the required OpenShift Security Context Constraints (SCC) role bindings:

```bash
oc apply -f kubelet-plugin-privileged-role-binging.yaml
oc apply -f compute-domain-daemon-anyuid-role-binding.yaml
```

**Note:** These role bindings are currently required for the DRA driver to function properly on OpenShift. This requirement in
[PR #569](https://github.com/NVIDIA/k8s-dra-driver-gpu/pull/569) of the NVIDIA DRA Driver.

## Verification

After installation, verify that the DRA driver is working:

```bash
# Check DRA driver pods
oc get pods -n nvidia-dra-driver-gpu

# Check for available GPU resource classes
oc get resourceclass
```

## IMEX Multi-Node Tests

The [`imex-test-jobs`](imex-test-jobs/) directory contains validation tests for multi-node NVLink communication using NVIDIA IMEX channels.

## Documentation Links

- [NVIDIA DRA Driver for GPUs](https://github.com/NVIDIA/k8s-dra-driver-gpu/)
- [NVIDIA GPU Operator](https://docs.nvidia.com/datacenter/cloud-native/openshift/latest/index.html)
- [Kubernetes Dynamic Resource Allocation](https://kubernetes.io/docs/concepts/scheduling-eviction/dynamic-resource-allocation/)
- [Feature Gates](https://docs.redhat.com/en/documentation/openshift_container_platform/4.20/html/postinstallation_configuration/post-install-cluster-tasks#nodes-cluster-enabling-features-about_post-install-cluster-tasks)

