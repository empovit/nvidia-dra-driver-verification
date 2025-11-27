# Compute Domain Simulator Makefile
# Copyright (c) 2025 NVIDIA CORPORATION.  All rights reserved.

# Version configuration - simulator version matches driver version exactly
DRIVER_VERSION ?= $(shell cat VERSION)

# Image configuration
VERSION ?= v$(DRIVER_VERSION)
GIT_COMMIT ?= $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
BASE_IMAGE ?= nvcr.io/nvidia/k8s-dra-driver-gpu:v$(DRIVER_VERSION)
IMAGE_NAME ?= compute-domain-simulator
IMAGE_REGISTRY ?= ghcr.io/empovit
IMAGE_TAG ?= $(VERSION)
FULL_IMAGE ?= $(IMAGE_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)

# Build platform configuration
BUILDX_PLATFORM ?= linux/amd64,linux/arm64
DOCKER_BUILDKIT ?= 1
CONTAINER_CMD ?= docker

.PHONY: help
help: ## Show this help message
	@echo "Compute Domain Simulator Build Targets:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'
	@echo ""
	@echo "Configuration:"
	@echo "  DRIVER_VERSION:  $(DRIVER_VERSION)"
	@echo "  VERSION:         $(VERSION)"
	@echo "  GIT_COMMIT:      $(GIT_COMMIT)"
	@echo "  BASE_IMAGE:      $(BASE_IMAGE)"
	@echo "  FULL_IMAGE:      $(FULL_IMAGE)"
	@echo "  BUILDX_PLATFORM: $(BUILDX_PLATFORM)"
	@echo "  CONTAINER_CMD:   $(CONTAINER_CMD)"

.PHONY: build
build: ## Build the simulator container image for current platform
	$(CONTAINER_CMD) build \
		--build-arg VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--build-arg BASE_IMAGE=$(BASE_IMAGE) \
		--build-arg DRIVER_VERSION=$(DRIVER_VERSION) \
		-t $(FULL_IMAGE) \
		-f Dockerfile \
		.

.PHONY: build-multiarch
build-multiarch: ## Build multi-architecture image (requires buildx)
	$(CONTAINER_CMD) buildx build \
		--platform $(BUILDX_PLATFORM) \
		--build-arg VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--build-arg BASE_IMAGE=$(BASE_IMAGE) \
		--build-arg DRIVER_VERSION=$(DRIVER_VERSION) \
		-t $(FULL_IMAGE) \
		-f Dockerfile \
		.

.PHONY: push
push: build ## Build and push the image to registry
	$(CONTAINER_CMD) push $(FULL_IMAGE)

.PHONY: push-multiarch
push-multiarch: ## Build and push multi-architecture image
	$(CONTAINER_CMD) buildx build \
		--platform $(BUILDX_PLATFORM) \
		--build-arg VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--build-arg BASE_IMAGE=$(BASE_IMAGE) \
		--build-arg DRIVER_VERSION=$(DRIVER_VERSION) \
		-t $(FULL_IMAGE) \
		-f Dockerfile \
		--push \
		.

.PHONY: test-local
test-local: ## Test the simulator binaries locally (requires Go)
	@echo "Testing nvidia-imex simulator..."
	go run nvidia-imex-simulator.go -h || true
	@echo ""
	@echo "Testing nvidia-imex-ctl simulator..."
	go run nvidia-imex-ctl-simulator.go -h || true

.PHONY: test-container
test-container: build ## Test the built container image
	@echo "Testing simulator container..."
	$(CONTAINER_CMD) run --rm $(FULL_IMAGE) nvidia-imex --help 2>&1 || true
	@echo ""
	$(CONTAINER_CMD) run --rm $(FULL_IMAGE) nvidia-imex-ctl --help 2>&1 || true
	@echo ""
	$(CONTAINER_CMD) run --rm $(FULL_IMAGE) compute-domain-daemon --version

.PHONY: clean
clean: ## Clean up built artifacts
	$(CONTAINER_CMD) rmi $(FULL_IMAGE) 2>/dev/null || true

.PHONY: shell
shell: build ## Run a shell in the container for debugging
	$(CONTAINER_CMD) run --rm -it $(FULL_IMAGE) /bin/bash

.PHONY: inspect
inspect: ## Inspect the container image
	$(CONTAINER_CMD) inspect $(FULL_IMAGE) | jq '.[0].Config.Env'
	@echo ""
	@echo "Checking for nvidia-imex binaries:"
	$(CONTAINER_CMD) run --rm $(FULL_IMAGE) which nvidia-imex
	$(CONTAINER_CMD) run --rm $(FULL_IMAGE) which nvidia-imex-ctl
	$(CONTAINER_CMD) run --rm $(FULL_IMAGE) which compute-domain-daemon

.PHONY: validate
validate: build ## Validate simulator against real driver
	@echo "Running validation against base image..."
	CONTAINER_CMD=$(CONTAINER_CMD) BASE_IMAGE=$(BASE_IMAGE) SIMULATOR_IMAGE=$(FULL_IMAGE) ./validate-against-driver.sh

.PHONY: install
install: ## Install DRA driver with simulator image
	./install-dra-driver-with-simulator.sh

