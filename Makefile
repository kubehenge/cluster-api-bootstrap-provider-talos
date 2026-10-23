# Everything runs natively with the Go toolchain go.mod names: no container is needed to generate,
# lint, test or build. Tools are installed into $(BIN) at the versions below; `go install` checks
# each one against the Go checksum database, so a version pins its content as a digest would.

REGISTRY ?= ghcr.io/kubehenge
NAME := cluster-api-bootstrap-provider-talos
IMAGE ?= $(REGISTRY)/$(NAME)
TAG ?= $(shell git describe --tags --always --dirty)
ARTIFACTS := _out
BIN := $(abspath $(ARTIFACTS)/bin)
PLATFORMS ?= linux/amd64,linux/arm64

GO ?= go
export CGO_ENABLED ?= 0
export GOFLAGS ?= -trimpath

# renovate: datasource=go depName=sigs.k8s.io/controller-tools
CONTROLLER_GEN_VERSION := v0.23.0
# conversion-gen comes from the Kubernetes minor go.mod builds against, so the generated conversions
# match the apimachinery they call.
# renovate: datasource=go depName=k8s.io/code-generator
CONVERSION_GEN_VERSION := v0.36.5
# renovate: datasource=go depName=github.com/golangci/golangci-lint/v2
GOLANGCI_LINT_VERSION := v2.15.0
# renovate: datasource=go depName=sigs.k8s.io/kustomize/kustomize/v5
KUSTOMIZE_VERSION := v5.8.3
# renovate: datasource=go depName=github.com/google/ko
KO_VERSION := v0.19.1
# renovate: datasource=go depName=github.com/sigstore/cosign/v3
COSIGN_VERSION := v3.1.4
# The Talos release the integration tests' cluster runs in Docker on linux/amd64, as CI does: talosctl
# checked against its SHA-256 and the node image named by digest. TALOS_IMAGE carries TALOS_VERSION.
# renovate: datasource=github-release-attachments depName=siderolabs/talos
TALOS_VERSION := v1.14.4
TALOSCTL_SHA256 := 0a69e3f9b682c5cd844b78dae305f6056ceb86c5fcc6913dcea78c527ed97cee
TALOS_IMAGE := ghcr.io/siderolabs/talos:v1.14.4@sha256:0e85248bd679398e78bd97ce62179bafbbecf0a32df4d93579ee03e89fe1c4b0

# The suffix `go install` gives binaries on the build platform, so the tool targets find what they built.
GOEXE := $(shell $(GO) env GOEXE)
CONTROLLER_GEN := $(BIN)/controller-gen$(GOEXE)
CONVERSION_GEN := $(BIN)/conversion-gen$(GOEXE)
GOLANGCI_LINT := $(BIN)/golangci-lint$(GOEXE)
KUSTOMIZE := $(BIN)/kustomize$(GOEXE)
KO := $(BIN)/ko$(GOEXE)
COSIGN := $(BIN)/cosign$(GOEXE)
TALOSCTL := $(BIN)/talosctl-$(TALOS_VERSION)
# controller-gen is given import paths, never directories: it makes a directory root absolute with
# filepath.Abs, which on some platforms drops a trailing `...` and leaves only the directory itself.
MODULE := $(shell $(GO) list -m)

ENV_NAME := cabpt-env
ENV_DIR := $(abspath $(ARTIFACTS)/env)
# Every image and manifest the integration environment pulls, pinned by digest or SHA-256: the Talos
# configuration patches name the cluster's Kubernetes images, components.txt what clusterctl installs.
ENV_PINS := internal/integration/env
ENV_PATCH := $(ENV_PINS)/talos-patch.yaml
ENV_PATCH_CONTROLPLANES := $(ENV_PINS)/talos-patch-controlplanes.yaml
ENV_COMPONENTS := $(ENV_PINS)/components.txt
# The cluster's Kubernetes version is the one its pinned API server image carries.
KUBERNETES_VERSION = $(shell sed -nE 's|^image: registry\.k8s\.io/kube-apiserver:v([^@]+)@.*|\1|p' $(ENV_PATCH_CONTROLPLANES))
# talosctl gives the control plane node the first address after the gateway and the worker the next.
ENV_NODES := 10.5.0.2,10.5.0.3
INTEGRATION_PKG := $(MODULE)/internal/integration
# The integration tests install the Cluster API release go.mod requires.
CAPI_VERSION = $(shell $(GO) list -m -f '{{.Version}}' sigs.k8s.io/cluster-api)
INTEGRATION_LDFLAGS = -X $(INTEGRATION_PKG).TalosVersion=$(TALOS_VERSION) \
	-X $(INTEGRATION_PKG).CAPIVersion=$(CAPI_VERSION) \
	-X $(INTEGRATION_PKG).Components=$(ENV_DIR)/bootstrap-components.yaml \
	-X $(INTEGRATION_PKG).ClusterctlConfig=$(ENV_DIR)/clusterctl.yaml

.DEFAULT_GOAL := help

.PHONY: help
help: ## Show the targets.
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "%-20s %s\n", $$1, $$2}'

$(CONTROLLER_GEN):
	GOBIN=$(BIN) $(GO) install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_GEN_VERSION)

$(CONVERSION_GEN):
	GOBIN=$(BIN) $(GO) install k8s.io/code-generator/cmd/conversion-gen@$(CONVERSION_GEN_VERSION)

$(GOLANGCI_LINT):
	GOBIN=$(BIN) $(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

$(KUSTOMIZE):
	GOBIN=$(BIN) $(GO) install sigs.k8s.io/kustomize/kustomize/v5@$(KUSTOMIZE_VERSION)

$(KO):
	GOBIN=$(BIN) $(GO) install github.com/google/ko@$(KO_VERSION)

$(COSIGN):
	GOBIN=$(BIN) $(GO) install github.com/sigstore/cosign/v3/cmd/cosign@$(COSIGN_VERSION)

$(TALOSCTL):
	mkdir -p $(BIN)
	curl -fsSL --proto '=https' --retry 5 -o $(TALOSCTL).download \
		https://github.com/siderolabs/talos/releases/download/$(TALOS_VERSION)/talosctl-linux-amd64
	echo "$(TALOSCTL_SHA256)  $(TALOSCTL).download" | sha256sum -c -
	chmod +x $(TALOSCTL).download
	mv $(TALOSCTL).download $(TALOSCTL)

.PHONY: tools
tools: $(CONTROLLER_GEN) $(CONVERSION_GEN) $(GOLANGCI_LINT) $(KUSTOMIZE) $(KO) $(COSIGN) ## Install the pinned tools into _out/bin.

.PHONY: generate
generate: $(CONTROLLER_GEN) $(CONVERSION_GEN) ## Generate deepcopy and conversion code.
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths=$(MODULE)/api/...
	$(CONVERSION_GEN) --output-file=zz_generated.conversion.go --go-header-file=./hack/boilerplate.go.txt $(MODULE)/api/v1alpha3

.PHONY: manifests
manifests: $(CONTROLLER_GEN) ## Generate the CRDs, webhooks and RBAC rules.
	$(CONTROLLER_GEN) crd:crdVersions=v1 paths=$(MODULE)/api/... output:crd:dir=config/crd/bases output:webhook:dir=config/webhook webhook
	$(CONTROLLER_GEN) rbac:roleName=manager-role paths=$(MODULE)/controllers/... output:rbac:dir=config/rbac

.PHONY: check-dirty
check-dirty: ## Fail when the tree differs from what is committed, as after generate or manifests.
	@if test -n "`git status --porcelain`"; then echo "Source tree is dirty"; git status; git diff; exit 1; fi

.PHONY: lint
lint: $(GOLANGCI_LINT) ## Run golangci-lint.
	$(GOLANGCI_LINT) run --allow-serial-runners ./...

.PHONY: unit
unit: ## Run the unit tests under the race detector.
	mkdir -p $(ARTIFACTS)
	CGO_ENABLED=1 $(GO) test -race -count=1 -covermode=atomic -coverprofile=$(ARTIFACTS)/coverage.txt \
		$$($(GO) list ./... | grep -v /internal/integration)

.PHONY: env-up
env-up: $(TALOSCTL) ## Create the Talos cluster in Docker the integration tests run against (linux/amd64).
	@case "$(TALOS_IMAGE)" in *:$(TALOS_VERSION)@sha256:*) ;; *) echo "TALOS_IMAGE is no image of Talos $(TALOS_VERSION)"; exit 1 ;; esac
	@test -n "$(KUBERNETES_VERSION)" || { echo "$(ENV_PATCH_CONTROLPLANES) names no kube-apiserver image"; exit 1; }
	@if grep -hoE '/(kubelet|kube-apiserver|kube-controller-manager|kube-scheduler|kube-proxy):[^@]+@' $(ENV_PATCH) $(ENV_PATCH_CONTROLPLANES) | \
		grep -vxE '/[a-z-]+:v$(subst .,\.,$(KUBERNETES_VERSION))@'; then \
		echo "the Kubernetes images above are not of Kubernetes $(KUBERNETES_VERSION), as the API server is"; exit 1; fi
	mkdir -p $(ENV_DIR)
	$(TALOSCTL) cluster create docker --name=$(ENV_NAME) --image=$(TALOS_IMAGE) --workers=1 \
		--kubernetes-version=$(KUBERNETES_VERSION) \
		--config-patch=@$(ENV_PATCH) --config-patch-controlplanes=@$(ENV_PATCH_CONTROLPLANES) \
		--talosconfig-destination=$(ENV_DIR)/talosconfig
	$(TALOSCTL) kubeconfig $(ENV_DIR)/kubeconfig --talosconfig=$(ENV_DIR)/talosconfig --nodes=10.5.0.2 --force

.PHONY: env-down
env-down: $(TALOSCTL) ## Remove the cluster env-up created.
	$(TALOSCTL) cluster destroy --name=$(ENV_NAME)
	rm -rf $(ENV_DIR)

.PHONY: integration
integration: $(KUSTOMIZE) $(TALOSCTL) ## Run the integration tests against the cluster env-up created; they install Cluster API in it.
	mkdir -p $(ENV_DIR)
	$(KUSTOMIZE) build config/default > $(ENV_DIR)/bootstrap-components.yaml
	hack/integration-components.sh $(ENV_COMPONENTS) $(ENV_DIR) $(CAPI_VERSION)
	KUBECONFIG=$(ENV_DIR)/kubeconfig CGO_ENABLED=1 $(GO) test -race -count=1 -timeout=30m -v \
		-ldflags "$(INTEGRATION_LDFLAGS)" ./internal/integration/...
	hack/check-images.sh $(TALOSCTL) $(ENV_DIR)/talosconfig $(ENV_NODES) \
		$(ENV_COMPONENTS) $(ENV_PATCH) $(ENV_PATCH_CONTROLPLANES)

.PHONY: image
image: $(KO) ## Build the controller image with ko and push it to $(IMAGE); prints IMAGE@digest.
	KO_DOCKER_REPO=$(IMAGE) $(KO) build --bare --sbom=none --platform=$(PLATFORMS) \
		--image-label org.opencontainers.image.source=https://github.com/kubehenge/$(NAME) \
		--image-label org.opencontainers.image.licenses=MPL-2.0 \
		--tags $(TAG) $(KO_FLAGS) .

.PHONY: release-manifests
release-manifests: $(KUSTOMIZE) ## Render bootstrap-components.yaml and metadata.yaml for CONTROLLER_IMAGE into _out/release.
	@test -n "$(CONTROLLER_IMAGE)" || { echo "set CONTROLLER_IMAGE to the image reference, by digest"; exit 1; }
	rm -rf $(ARTIFACTS)/release $(ARTIFACTS)/config
	mkdir -p $(ARTIFACTS)/release
	cp -r config $(ARTIFACTS)/config
	cd $(ARTIFACTS)/config/manager && $(KUSTOMIZE) edit set image controller=$(CONTROLLER_IMAGE)
	$(KUSTOMIZE) build $(ARTIFACTS)/config/default > $(ARTIFACTS)/release/bootstrap-components.yaml
	cp config/metadata/metadata.yaml $(ARTIFACTS)/release/metadata.yaml
	rm -rf $(ARTIFACTS)/config

.PHONY: clean
clean: ## Remove _out.
	rm -rf $(ARTIFACTS)
