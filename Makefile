NAME ?= mumoshu/okra
VERSION ?= latest
CHART ?= okra
GO ?= go

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
# CRD_OPTIONS ?= "crd:trivialVersions=true"
CRD_OPTIONS ?= "crd"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# default list of platforms for which multiarch image is built
ifeq (${PLATFORMS}, )
	export PLATFORMS="linux/amd64,linux/arm64"
endif

# if IMG_RESULT is unspecified, by default the image will be pushed to registry
ifeq (${IMG_RESULT}, load)
	export PUSH_ARG="--load"
    # if load is specified, image will be built only for the build machine architecture.
    export PLATFORMS="local"
else ifeq (${IMG_RESULT}, cache)
	# if cache is specified, image will only be available in the build cache, it won't be pushed or loaded
	# therefore no PUSH_ARG will be specified
else
	export PUSH_ARG="--push"
endif

# Run tests
test: generate fmt vet manifests
	$(GO) test ./... -coverprofile cover.out
	kustomize build config/default | kubectl apply -f - --dry-run

manifests: controller-gen chart-crds
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

chart-crds:
	cp config/crd/bases/*.yaml charts/$(CHART)/crds/

test-chart:
	 bash -c 'diff --unified <(helm template charts/$(CHART) --include-crds) <(kustomize build config/default)'

# Run go fmt against code
fmt:
	$(GO) fmt ./...

# Run go vet against code
vet:
	$(GO) vet ./...

build: generate
	$(GO) build .

docker-buildx: buildx
	export DOCKER_CLI_EXPERIMENTAL=enabled
	@if ! docker buildx ls | grep -q container-builder; then\
		docker buildx create --platform ${PLATFORMS} --name container-builder --use;\
	fi
	docker buildx build --platform ${PLATFORMS} \
		--build-arg RUNNER_VERSION=${RUNNER_VERSION} \
		--build-arg DOCKER_VERSION=${DOCKER_VERSION} \
		-t "${NAME}:${VERSION}" \
		-f Dockerfile \
		. ${PUSH_ARG}

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths="./..."

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
ifeq (, $(wildcard $(GOBIN)/controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	$(GO) mod init tmp ;\
	$(GO) get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.7.0 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
endif
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

OS_NAME := $(shell uname -s | tr A-Z a-z)

# find or download controller-gen
# download controller-gen if necessary
buildx:
ifeq (, $(shell [ -e ~/.docker/cli-plugins/docker-buildx ]))
	@{ \
	set -e ;\
	BUILDX_TMP_DIR=$$(mktemp -d) ;\
	cd $$BUILDX_TMP_DIR ;\
	wget https://github.com/docker/buildx/releases/download/v0.4.2/buildx-v0.4.2.$(OS_NAME)-amd64 ;\
	chmod a+x buildx-v0.4.2.$(OS_NAME)-amd64 ;\
	mkdir -p ~/.docker/cli-plugins ;\
	mv buildx-v0.4.2.$(OS_NAME)-amd64 ~/.docker/cli-plugins/docker-buildx ;\
	rm -rf $$BUILDX_TMP_DIR ;\
	}
BUILDX_BIN=~/.docker/cli-plugins/docker-buildx
else
BUILDX_BIN=~/.docker/cli-plugins/docker-buildx
endif
