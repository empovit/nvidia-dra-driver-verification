# NVIDIA Compute Domain Simulator

Drop-in replacement for the NVIDIA Compute Domain Driver that simulates IMEX daemon behavior without requiring GPU hardware. Use it to test Kubernetes DRA deployments, IMEX channel orchestration, and security contexts without physical infrastructure.

## What It Does

Simulates `nvidia-imex` and `nvidia-imex-ctl` binaries:
- Reads configs from `/imexd/imexd.cfg` and `/imexd/nodes.cfg`
- Opens TCP ports (50000 for peers, 50005 for command)
- Responds to health checks with "READY" status
- Handles SIGUSR1 for config reload
- Writes PID, log, and stats files (for security context testing)

What's NOT simulated: actual GPU operations.

## Building

```bash
# Build (uses docker by default, set CONTAINER_CMD=podman for podman)
make build

# Validate against real driver
make validate

# Push to registry
IMAGE_REGISTRY=your-registry.io make push

# Install DRA driver with simulator
make install

# Using podman
CONTAINER_CMD=podman make build
```

## Usage

Replace image in your DaemonSet:

```yaml
# Before
image: nvcr.io/nvidia/k8s-dra-driver-gpu:v25.8.0

# After (version matches driver exactly)
image: your-registry.io/compute-domain-simulator:v25.8.0
```

Everything else stays the same - same environment variables, volumes, and configuration.

## Files Created

- `/var/run/nvidia-imex.pid` - PID file
- `/var/log/nvidia-imex-stats.log` - Stats (updated every 30s)
- `/imexd/imexd.cfg` and `/imexd/nodes.cfg` - Configs (by compute-domain-daemon)

## OpenShift

Grant `anyuid` SCC (usually sufficient):

```bash
oc adm policy add-scc-to-user anyuid -z nvidia-dra-driver -n nvidia-dra-driver
```

## Troubleshooting

Check logs:
```bash
kubectl logs <pod> -c compute-domain-daemon | grep -E "Warning|cannot|permission"
```

Common issues:
- Permission denied: Mount `/var/run` and `/var/log` as emptyDir, or use `anyuid` SCC
- Missing env vars: Check CLIQUE_ID, POD_IP, NODE_NAME
- Status not READY: Wait 2-3 seconds for initialization

## Version Management

Update driver version by editing the VERSION file directly:
```bash
echo "25.9.0" > VERSION
```

Simulator version matches driver version exactly (e.g., driver v25.9.0 → simulator v25.9.0).

Then:
1. Run `make validate` to check compatibility
2. Test in your environment

## Installation

Install DRA driver with simulator image:

```bash
./install-dra-driver-with-simulator.sh
# or
make install
```