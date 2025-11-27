# syntax=docker.io/docker/dockerfile:1.7-labs

# Copyright (c) 2025 NVIDIA CORPORATION.  All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Version args - passed from Makefile
ARG DRIVER_VERSION=25.8.0
ARG BASE_IMAGE=nvcr.io/nvidia/k8s-dra-driver-gpu:v${DRIVER_VERSION}

# Build stage
FROM --platform=$BUILDPLATFORM golang:1.23-alpine AS build

ARG TARGETARCH
ARG VERSION=v${DRIVER_VERSION}
ARG GIT_COMMIT=unknown
ARG DRIVER_VERSION

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git make

# Copy source files
COPY nvidia-imex-simulator.go .
COPY nvidia-imex-ctl-simulator.go .

# Build the simulator binaries with version info
RUN GOARCH=${TARGETARCH} CGO_ENABLED=0 go build \
    -ldflags="-s -w -X main.SimulatorVersion=v${DRIVER_VERSION} -X main.TargetDriverVersion=v${DRIVER_VERSION}" \
    -o /artifacts/nvidia-imex \
    nvidia-imex-simulator.go

RUN GOARCH=${TARGETARCH} CGO_ENABLED=0 go build \
    -ldflags="-s -w" \
    -o /artifacts/nvidia-imex-ctl \
    nvidia-imex-ctl-simulator.go

# Create wrapper script for compute-domain-daemon
RUN echo '#!/bin/sh' > /artifacts/compute-domain-daemon-wrapper.sh && \
    echo '# Wrapper script that ensures nvidia-imex binaries are in PATH' >> /artifacts/compute-domain-daemon-wrapper.sh && \
    echo 'export PATH=/usr/local/bin:$PATH' >> /artifacts/compute-domain-daemon-wrapper.sh && \
    echo 'exec /usr/bin/compute-domain-daemon.real "$@"' >> /artifacts/compute-domain-daemon-wrapper.sh

RUN chmod +x /artifacts/compute-domain-daemon-wrapper.sh

# Production stage - based on the original k8s-dra-driver image
FROM ${BASE_IMAGE}

LABEL io.k8s.display-name="NVIDIA Compute Domain Simulator"
LABEL name="NVIDIA Compute Domain Simulator"
LABEL vendor="NVIDIA"
LABEL version="simulator-1.0.0"
LABEL description="Simulator for NVIDIA Compute Domain Driver for testing IMEX channels without hardware"
LABEL org.opencontainers.image.description="NVIDIA Compute Domain Simulator"
LABEL org.opencontainers.image.source="https://github.com/NVIDIA/compute-domain-simulator"

# Copy simulator binaries to /usr/local/bin (higher priority in PATH)
COPY --from=build --chmod=755 /artifacts/nvidia-imex /usr/local/bin/nvidia-imex
COPY --from=build --chmod=755 /artifacts/nvidia-imex-ctl /usr/local/bin/nvidia-imex-ctl

# Note: We rely on PATH priority (/usr/local/bin before /usr/bin)
# so our simulator binaries are found before any real nvidia-imex in the base image

# Set PATH to prioritize /usr/local/bin
ENV PATH=/usr/local/bin:/usr/bin:/bin

