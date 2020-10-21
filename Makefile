#
# Copyright 2020 IBM Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# This repo is build locally for dev/test by default;
# Override this variable in CI env.

# Current Operator version
VERSION ?= 0.0.1
# Default bundle image tag
BUNDLE_IMG ?= controller-bundle:$(VERSION)
# Options for 'bundle-build'
ifneq ($(origin CHANNELS), undefined)
BUNDLE_CHANNELS := --channels=$(CHANNELS)
endif
ifneq ($(origin DEFAULT_CHANNEL), undefined)
BUNDLE_DEFAULT_CHANNEL := --default-channel=$(DEFAULT_CHANNEL)
endif
BUNDLE_METADATA_OPTS ?= $(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# This repo is build locally for dev/test by default;
# Override this variable in CI env.
BUILD_LOCALLY ?= 1

# Image URL to use all building/pushing image targets;
# Use your own docker registry and image name for dev/test by overriding the IMG, REGISTRY and CSV_VERSION environment variable.
IMG ?= ibm-licensing-operator
## # Image URL to use all building/pushing image targets
## IMG ?= controller:latest


REGISTRY ?= "hyc-cloud-private-integration-docker-local.artifactory.swg-devops.com/ibmcom"
CSV_VERSION ?= 1.3.0

SCRATCH_REGISTRY ?= "hyc-cloud-private-scratch-docker-local.artifactory.swg-devops.com/ibmcom"

# Set the registry and tag for the operand images
OPERAND_REGISTRY ?= $(REGISTRY)
OPERAND_TAG ?= 1.3.0

# When pushing CSV locally you need to have these credentials set as environment variables.
QUAY_USERNAME ?=
QUAY_PASSWORD ?=

# Linter urls that should be skipped
MARKDOWN_LINT_WHITELIST ?= https://quay.io/cnr,https://www-03preprod.ibm.com/support/knowledgecenter/SSHKN6/installer/3.3.0/install_operator.html,https://github.com/IBM/ibm-licensing-operator/releases/download/,https://github.com/operator-framework/operator-lifecycle-manager/releases/download,http://ibm.biz/

# The namespace that operator will be deployed in
NAMESPACE ?= ibm-common-services

# Github host to use for checking the source tree;
# Override this variable ue with your own value if you're working on forked repo.
GIT_HOST ?= github.com/IBM

PWD := $(shell pwd)
BASE_DIR := $(shell basename $(PWD))

# Keep an existing GOPATH, make a private one if it is undefined
GOPATH_DEFAULT := $(PWD)/.go
export GOPATH ?= $(GOPATH_DEFAULT)
GOBIN_DEFAULT := $(GOPATH)/bin
export GOBIN ?= $(GOBIN_DEFAULT)
TESTARGS_DEFAULT := "-v"
export TESTARGS ?= $(TESTARGS_DEFAULT)
DEST := $(GOPATH)/src/$(GIT_HOST)/$(BASE_DIR)
VERSION ?= $(shell git describe --exact-match 2> /dev/null || \
                 git describe --match=$(git rev-parse --short=8 HEAD) --always --dirty --abbrev=8)
MANIFEST_VERSION ?= $(shell cat ./version/version.go | grep "Version =" | awk '{ print $$3}' | tr -d '"')

LOCAL_OS := $(shell uname)
ifeq ($(LOCAL_OS),Linux)
    TARGET_OS ?= linux
    XARGS_FLAGS="-r"
	STRIP_FLAGS=
else ifeq ($(LOCAL_OS),Darwin)
    TARGET_OS ?= darwin
    XARGS_FLAGS=
	STRIP_FLAGS="-x"
else
    $(error "This system's OS $(LOCAL_OS) isn't recognized/supported")
endif

ARCH := $(shell uname -m)
LOCAL_ARCH := "amd64"
ifeq ($(ARCH),x86_64)
    LOCAL_ARCH="amd64"
else ifeq ($(ARCH),ppc64le)
    LOCAL_ARCH="ppc64le"
else ifeq ($(ARCH),s390x)
    LOCAL_ARCH="s390x"
else
    $(error "This system's ARCH $(ARCH) isn't recognized/supported")
endif

# Setup DOCKER_BUILD_OPTS after all includes complete
#Variables for redhat ubi certification required labels
IMAGE_NAME=$(IMG)
IMAGE_DISPLAY_NAME=IBM Licensing Operator
IMAGE_MAINTAINER=talk2sam@us.ibm.com
IMAGE_VENDOR=IBM
IMAGE_VERSION=$(VERSION)
IMAGE_DESCRIPTION=Operator used to install a service to measure VPC license use of IBM products deployed in the cluster.
IMAGE_SUMMARY=$(IMAGE_DESCRIPTION)
IMAGE_OPENSHIFT_TAGS=licensing
$(eval WORKING_CHANGES := $(shell git status --porcelain))
$(eval BUILD_DATE := $(shell date +%m/%d@%H:%M:%S))
$(eval GIT_COMMIT := $(shell git rev-parse --short HEAD))
$(eval VCS_REF := $(GIT_COMMIT))
IMAGE_RELEASE=$(VCS_REF)
GIT_REMOTE_URL = $(shell git config --get remote.origin.url)
$(eval DOCKER_BUILD_OPTS := --build-arg "IMAGE_NAME=$(IMAGE_NAME)" --build-arg "IMAGE_DISPLAY_NAME=$(IMAGE_DISPLAY_NAME)" --build-arg "IMAGE_MAINTAINER=$(IMAGE_MAINTAINER)" --build-arg "IMAGE_VENDOR=$(IMAGE_VENDOR)" --build-arg "IMAGE_VERSION=$(IMAGE_VERSION)" --build-arg "IMAGE_RELEASE=$(IMAGE_RELEASE)" --build-arg "IMAGE_DESCRIPTION=$(IMAGE_DESCRIPTION)" --build-arg "IMAGE_SUMMARY=$(IMAGE_SUMMARY)" --build-arg "IMAGE_OPENSHIFT_TAGS=$(IMAGE_OPENSHIFT_TAGS)" --build-arg "VCS_REF=$(VCS_REF)" --build-arg "VCS_URL=$(GIT_REMOTE_URL)" --build-arg "IMAGE_NAME_ARCH=$(IMAGE_NAME)-$(LOCAL_ARCH)")

all: fmt check test coverage build images

ifeq ($(BUILD_LOCALLY),0)
    ifneq ("$(realpath $(DEST))", "$(realpath $(PWD))")
        $(error Please run 'make' from $(DEST). Current directory is $(PWD))
    endif
endif

include common/Makefile.common.mk


##@ Application

install: ## Install all resources (CR/CRD's, RBAC and operator)
	@echo ....... Set environment variables ......
	- export WATCH_NAMESPACE=
	@echo ....... Creating namespace .......
	- kubectl create namespace ${NAMESPACE}
	@echo ....... Applying CRDs .......
	- kubectl apply -f deploy/crds/operator.ibm.com_ibmlicensings_crd.yaml
	@echo ....... Applying RBAC .......
	- kubectl apply -f deploy/service_account.yaml -n ${NAMESPACE}
	- kubectl apply -f deploy/role.yaml
	- kubectl apply -f deploy/role_binding.yaml
	@echo ....... Applying Operator .......
	- kubectl apply -f deploy/operator.yaml -n ${NAMESPACE}
	@echo ....... Creating the Instances .......
	- kubectl apply -f deploy/crds/operator.ibm.com_v1alpha1_ibmlicensing_cr.yaml
uninstall: ## Uninstall all that all performed in the $ make install, without namespace as there might be other things there
	@echo ....... Uninstalling .......
	@echo ....... Deleting the Instances .......
	- kubectl delete -f deploy/crds/operator.ibm.com_v1alpha1_ibmlicensing_cr.yaml --ignore-not-found
	@echo ....... Deleting Operator .......
	- kubectl delete -f deploy/operator.yaml -n ${NAMESPACE}
	@echo ....... Deleting CRDs .......
	- kubectl delete -f deploy/crds/operator.ibm.com_ibmlicensings_crd.yaml --ignore-not-found
	@echo ....... Deleting RBAC .......
	- kubectl delete -f deploy/role_binding.yaml --ignore-not-found
	- kubectl delete -f deploy/service_account.yaml -n ${NAMESPACE} --ignore-not-found
	- kubectl delete -f deploy/role.yaml --ignore-not-found

############################################################
# work section
############################################################
$(GOBIN):
	@echo "create gobin"
	@mkdir -p $(GOBIN)

work: $(GOBIN)

##@ Development

code-dev: ## Run the default dev commands which are the go tidy, fmt, vet then execute the $ make code-gen
	@echo Running the common required commands for developments purposes
	- make code-tidy
	- make code-fmt
	- make code-vet
	- make code-gen
	@echo Running the common required commands for code delivery
	make check

# All available format: format-go format-protos format-python
# Default value will run all formats, override these make target with your requirements:
#    eg: fmt: format-go format-protos
fmt: format-go format-protos format-python
	go fmt ./...

check: lint ## Check all files lint errors, this is also done before pushing the code to remote branch

# All available linters: lint-dockerfiles lint-scripts lint-yaml lint-copyright-banner lint-go lint-python lint-helm lint-markdown lint-sass lint-typescript lint-protos
# Default value will run all linters, override these make target with your requirements:
#    eg: lint: lint-go lint-yaml
lint: lint-all

test: ## Run all tests if available
	@go test ${TESTARGS} ./...

coverage: ## Run coverage if possible
	@common/scripts/codecov.sh ${BUILD_LOCALLY}

run: ## Run against the configured Kubernetes cluster in ~/.kube/config
	@echo ....... Start Operator locally with go run ......
	WATCH_NAMESPACE= go run ./cmd/manager/main.go

############################################################
# install operator sdk section
############################################################

install-operator-sdk:
	@operator-sdk version 2> /dev/null ; if [ $$? -ne 0 ]; then ./common/scripts/install-operator-sdk.sh; fi

ifeq ($(BUILD_LOCALLY),0)
    export CONFIG_DOCKER_TARGET = config-docker
config-docker:
endif

ifeq ($(BUILD_LOCALLY),0)
    export CONFIG_DOCKER_TARGET_SCRATCH = config-docker-scratch
config-docker-scratch:
endif

##@ Build

build:
	@echo "Building the $(IMAGE_NAME) binary for $(LOCAL_ARCH)..."
	@GOARCH=$(LOCAL_ARCH) common/scripts/gobuild.sh build/_output/bin/$(IMAGE_NAME) ./cmd/manager
	@strip $(STRIP_FLAGS) build/_output/bin/$(IMAGE_NAME)

build-push-image: build-image push-image

build-image: build
	@echo "Building the $(IMAGE_NAME) docker image for $(LOCAL_ARCH)..."
	@docker build -t $(REGISTRY)/$(IMAGE_NAME)-$(LOCAL_ARCH):$(VERSION) $(DOCKER_BUILD_OPTS) -f build/Dockerfile .

push-image: $(CONFIG_DOCKER_TARGET) build-image
	@echo "Pushing the $(IMAGE_NAME) docker image for $(LOCAL_ARCH)..."
	@docker push $(REGISTRY)/$(IMAGE_NAME)-$(LOCAL_ARCH):$(VERSION)

build-push-image-development: build-image-development push-image-development

build-image-development: build
	@echo "Building the $(IMAGE_NAME) docker image for $(LOCAL_ARCH)..."
	@docker build -t $(SCRATCH_REGISTRY)/$(IMAGE_NAME)-$(LOCAL_ARCH):$(VERSION) $(DOCKER_BUILD_OPTS) -f build/Dockerfile .

push-image-development: $(CONFIG_DOCKER_TARGET_SCRATCH) build-image-development
	@echo "Pushing the $(IMAGE_NAME) docker image for $(LOCAL_ARCH)..."
	@docker push $(SCRATCH_REGISTRY)/$(IMAGE_NAME)-$(LOCAL_ARCH):$(VERSION)


##@ SHA Digest section

.PHONY: get-image-sha
get-image-sha: ## replaces operand tag for digest in operator.yaml and csv
	@echo Get SHA for ibm-licensing:$(OPERAND_TAG)
	@common/scripts/get-image-sha.sh $(OPERAND_REGISTRY)/ibm-licensing $(OPERAND_TAG)

##@ Release

multiarch-image: $(CONFIG_DOCKER_TARGET)
	@MAX_PULLING_RETRY=20 RETRY_INTERVAL=30 common/scripts/multiarch_image.sh $(REGISTRY) $(IMAGE_NAME) $(VERSION) ${MANIFEST_VERSION}
	common/scripts/catalog_build.sh $(REGISTRY) $(IMAGE_NAME) ${MANIFEST_VERSION}

multiarch-image-development: $(CONFIG_DOCKER_TARGET_SCRATCH)
	@MAX_PULLING_RETRY=20 RETRY_INTERVAL=30 common/scripts/multiarch_image.sh $(SCRATCH_REGISTRY) $(IMAGE_NAME) $(VERSION) ${MANIFEST_VERSION}
	common/scripts/catalog_build.sh $(SCRATCH_REGISTRY) $(IMAGE_NAME) ${MANIFEST_VERSION}

csv: ## Push CSV package to the catalog
	@RELEASE=${CSV_VERSION} common/scripts/push-csv.sh

##@ Red Hat Certificate Section

.PHONY: bundle
bundle: ## bundle zip file for certification
	@echo --- Updating the bundle directory with latest yamls from olm-catalog ---
	-mkdir bundle
	rm -rf bundle/*
	cp -r deploy/olm-catalog/ibm-licensing-operator/${CSV_VERSION}/* bundle/
	cp deploy/olm-catalog/ibm-licensing-operator/ibm-licensing-operator.package.yaml bundle/
	cd bundle && zip ibm-licensing-metadata ./*.yaml && cd ..

.PHONY: install-operator-courier
install-operator-courier: ## installs courier for certification check
	@echo --- Installing Operator Courier ---
	pip3 install operator-courier

.PHONY: verify-bundle
verify-bundle: ## verify bundle
	@echo --- Verify bundle is ready for Red Hat certification ---
	operator-courier --verbose verify --ui_validate_io bundle/

.PHONY: redhat-certify-ready
redhat-certify-ready: bundle verify-bundle ## makes bundle and verify it using operator courier

##@ Cleanup
clean: ## Clean build binary
	rm -f build/_output/bin/$(IMG)

##@ Help
help: ## Display this help
	@echo "Usage:  make <target>"
	@awk 'BEGIN {FS = ":.*##"}; \
		/^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } \
		/^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

kustomize:
ifeq (, $(shell which kustomize))
	@{ \
	set -e ;\
	KUSTOMIZE_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$KUSTOMIZE_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/kustomize/kustomize/v3@v3.5.4 ;\
	rm -rf $$KUSTOMIZE_GEN_TMP_DIR ;\
	}
KUSTOMIZE=$(GOBIN)/kustomize
else
KUSTOMIZE=$(shell which kustomize)
endif

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

# Run go vet against code
vet:
	go vet ./...

# Build the bundle image.
bundle-build:
	docker build -f bundle.Dockerfile -t $(BUNDLE_IMG) .

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.3.0 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager main.go

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Build the docker image
docker-build: test
	docker build . -t ${IMG}

# Push the docker image
docker-push:
	docker push ${IMG}

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests kustomize
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | kubectl apply -f -

.PHONY: all build run install uninstall code-dev check lint test coverage build multiarch-image csv clean help








# # Run tests
# test: generate fmt vet manifests
# 	go test ./... -coverprofile cover.out

# # Run against the configured Kubernetes cluster in ~/.kube/config
# run: generate fmt vet manifests
# 	go run ./main.go

# # Install CRDs into a cluster
# install: manifests kustomize
# 	$(KUSTOMIZE) build config/crd | kubectl apply -f -

# # Uninstall CRDs from a cluster
# uninstall: manifests kustomize
# 	$(KUSTOMIZE) build config/crd | kubectl delete -f -

# # Generate bundle manifests and metadata, then validate generated files.
# bundle: manifests
# 	operator-sdk generate kustomize manifests -q
# 	kustomize build config/manifests | operator-sdk generate bundle -q --overwrite --version $(VERSION) $(BUNDLE_METADATA_OPTS)
# 	operator-sdk bundle validate ./bundle

