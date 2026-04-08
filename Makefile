#
# Copyright 2026 IBM Corporation
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

# Current Operator version
CSV_VERSION ?= 4.2.21
CSV_VERSION_DEVELOPMENT ?= development
OLD_CSV_VERSION ?= 4.2.20

# Tools versions
OPM_VERSION ?= v1.64.0
OPERATOR_SDK_VERSION ?= v1.42.1
YQ_VERSION ?= v4.52.4
KUSTOMIZE_VERSION ?= v5.8.1
CONTROLLER_GEN_VERSION ?= v0.20.1
GOLANGCI_LINT_VERSION ?= v2.11.2
GOIMPORTS_VERSION ?= v0.43.0
SHELLCHECK_VERSION ?= v0.11.0
YAMLLINT_VERSION ?= 1.37.1
MDL_VERSION      ?= 0.15.0

# Local bin directory for all project tools (gitignored)
LOCALBIN := $(PWD)/bin
export PATH := $(LOCALBIN):$(PATH)

# Local temp directory for intermediate build files (gitignored)
LOCAL_TMP := $(PWD)/temp

# Tool binaries (all resolved to LOCALBIN)
CONTROLLER_GEN := $(LOCALBIN)/controller-gen
KUSTOMIZE      := $(LOCALBIN)/kustomize
OPM            := $(LOCALBIN)/opm
OPERATOR_SDK   := $(LOCALBIN)/operator-sdk
YQ             := $(LOCALBIN)/yq
GOLANGCI_LINT  := $(LOCALBIN)/golangci-lint
GOIMPORTS      := $(LOCALBIN)/goimports
DETECT_SECRETS := $(LOCALBIN)/detect-secrets
SHELLCHECK     := $(LOCALBIN)/shellcheck
YAMLLINT       := $(LOCALBIN)/.venv/bin/yamllint
MDL            := $(LOCALBIN)/mdl

# This repo is build locally for dev/test by default;
# Override this variable in CI env.
BUILD_LOCALLY ?= 1

# Image URL to use all building/pushing image targets;
# Use your own docker registry and image name for dev/test by overriding the IMG, REGISTRY and CSV_VERSION environment variable.
IMG ?= ibm-licensing-operator

REGISTRY_URL ?= docker-na-public.artifactory.swg-devops.com

REGISTRY ?= ${REGISTRY_URL}/hyc-cloud-private-integration-docker-local/ibmcom
SCRATCH_REGISTRY ?= ${REGISTRY_URL}/hyc-cloud-private-scratch-docker-local/ibmcom

# Default bundle image tag
IMAGE_BUNDLE_NAME ?= ibm-licensing-operator-bundle
IMAGE_CATALOG_NAME ?= ibm-licensing-operator-catalog

IBM_LICENSING_IMAGE ?= ibm-licensing

CHANNELS=v4.2
DEFAULT_CHANNEL=v4.2
PACKAGE="ibm-licensing-operator"

# Identify default channel based on tag of parent branch
GIT_BRANCH=$(shell git rev-parse --abbrev-ref HEAD)

# Identify tags created on current branch
BRANCH_TAGS=$(shell git tag --merged ${GIT_BRANCH})

# Options for 'bundle-build'
ifneq ($(origin CHANNELS), undefined)
BUNDLE_CHANNELS := --channels=$(CHANNELS)
endif
ifneq ($(origin DEFAULT_CHANNEL), undefined)
BUNDLE_DEFAULT_CHANNEL := --default-channel=$(DEFAULT_CHANNEL)
endif
BUNDLE_METADATA_OPTS ?= $(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?=  "crd:crdVersions=v1"

# Set the registry and tag for the operand images
OPERAND_REGISTRY ?= $(REGISTRY)
OPERAND_TAG ?= $(CSV_VERSION)

# When pushing CSV locally you need to have these credentials set as environment variables.
QUAY_USERNAME ?=
QUAY_PASSWORD ?=

# The namespace that operator will be deployed in
NAMESPACE ?= ibm-licensing

# Namespaces for Kind tests
OPREQ_TEST_NAMESPACE ?= opreq-ns

# Github host to use for checking the source tree;
# Override this variable ue with your own value if you're working on forked repo.
GIT_HOST ?= github.com/IBM

PWD := $(shell pwd)
BASE_DIR := $(shell basename $(PWD))

# Keep an existing GOPATH, make a private one if it is undefined
GOPATH_DEFAULT := $(PWD)/.go
export GOPATH ?= $(GOPATH_DEFAULT)
# Go tools are installed into LOCALBIN (not GOPATH/bin)
export GOBIN := $(LOCALBIN)
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

ifdef ARCH
  LOCAL_ARCH := $(ARCH)
else
  ARCH := $(shell uname -m)

  ifeq ($(ARCH),x86_64)
    LOCAL_ARCH := amd64
  else ifeq ($(ARCH),ppc64le)
    LOCAL_ARCH := ppc64le
  else ifeq ($(ARCH),s390x)
    LOCAL_ARCH := s390x
  else ifeq ($(ARCH),arm64)
    LOCAL_ARCH := arm64
  else
    $(error This system's ARCH '$(ARCH)' isn't recognized/supported)
  endif
endif

# Setup DOCKER_BUILD_OPTS after all includes complete
# Variables for redhat ubi certification required labels
IMAGE_NAME=$(IMG)
IMAGE_DISPLAY_NAME=IBM Licensing Operator
IMAGE_MAINTAINER=talk2sam@us.ibm.com
IMAGE_VENDOR=IBM
IMAGE_VERSION=$(VERSION)
IMAGE_DESCRIPTION=Operator used to install a service to measure VPC license use of IBM products deployed in the cluster.
IMAGE_SUMMARY=$(IMAGE_DESCRIPTION)
IMAGE_OPENSHIFT_TAGS=licensing
$(eval WORKING_CHANGES := $(shell git status --porcelain))
$(eval BUILD_DATE := $(shell date +%Y/%m/%d@%H:%M:%S))
$(eval GIT_COMMIT := $(shell git rev-parse --short HEAD))
$(eval VCS_REF := $(GIT_COMMIT))
IMAGE_RELEASE=$(VCS_REF)
IMAGE_BUILDDATE=$(BUILD_DATE)
GIT_REMOTE_URL = $(shell git config --get remote.origin.url)

BUNDLE_IMG ?= $(IMAGE_BUNDLE_NAME)-$(LOCAL_ARCH):$(VERSION)
CATALOG_IMG ?= $(IMAGE_CATALOG_NAME)-$(LOCAL_ARCH):$(VERSION)

DEVOPS_CATALOG_IMG ?= $(IMAGE_CATALOG_NAME)-$(LOCAL_ARCH):$(DEVOPS_STREAM)

$(eval DOCKER_BUILD_OPTS := --build-arg "IMAGE_NAME=$(IMAGE_NAME)" --build-arg "IMAGE_DISPLAY_NAME=$(IMAGE_DISPLAY_NAME)" --build-arg "IMAGE_MAINTAINER=$(IMAGE_MAINTAINER)" --build-arg "IMAGE_VENDOR=$(IMAGE_VENDOR)" --build-arg "IMAGE_VERSION=$(IMAGE_VERSION)" --build-arg "VERSION=$(CSV_VERSION)" --build-arg "IMAGE_RELEASE=$(IMAGE_RELEASE)"  --build-arg "IMAGE_BUILDDATE=$(IMAGE_BUILDDATE)" --build-arg "IMAGE_DESCRIPTION=$(IMAGE_DESCRIPTION)" --build-arg "IMAGE_SUMMARY=$(IMAGE_SUMMARY)" --build-arg "IMAGE_OPENSHIFT_TAGS=$(IMAGE_OPENSHIFT_TAGS)" --build-arg "VCS_REF=$(VCS_REF)" --build-arg "IMAGE_NAME_ARCH=$(IMAGE_NAME)-$(LOCAL_ARCH)")

ifeq ($(BUILD_LOCALLY),0)
    ifneq ("$(realpath $(DEST))", "$(realpath $(PWD))")
        $(error Please run 'make' from $(DEST). Current directory is $(PWD))
    endif
endif

ifeq ($(BUILD_LOCALLY),0)
    export CONFIG_DOCKER_TARGET = config-docker
config-docker:
endif

include common/Makefile.common.mk

all: fmt version.properties check test coverage-kind build images

# generate file containing info about the build
version.properties:
	infofile_path ?= version.properties
	$(shell echo "version="$(CSV_VERSION) > $(infofile_path))
	$(shell echo "build_date="$(BUILD_DATE) >> $(infofile_path))
	$(shell echo "commit="$(VCS_REF) >> $(infofile_path))

############################################################
# work section
############################################################

$(LOCALBIN):
	@mkdir -p $(LOCALBIN)

##@ Development

code-dev: ## Run the default dev commands which are the go tidy, fmt, vet then execute the $ make check
	@echo Running the common required commands for developments purposes
	- make code-tidy
	- make code-fmt
	- make fmt
	- make code-vet
	@echo Running the common required commands for code delivery
	- make check

# All available format: format-go format-protos
# Default value will run all formats, override these make target with your requirements:
#    eg: fmt: format-go format-protos
fmt: format-go

# Run go vet against code
vet:
	@go vet ./...

check: lint ## Check all files lint errors, this is also done before pushing the code to remote branch

# All available linters: lint-dockerfiles lint-scripts lint-yaml lint-copyright-banner lint-go lint-markdown lint-typescript lint-protos
# Default value will run all linters, override these make target with your requirements:
#    eg: lint: lint-go lint-yaml
lint: lint-all vet

# Run `make audit` before committing to ensure no secrets or sensitive credentials are present in the codebase
audit: $(DETECT_SECRETS) ## Run detect-secrets to scan for any secrets or sensitive credentials
	@$(DETECT_SECRETS) scan --update .secrets.baseline --exclude-files ".secrets.baseline|requirements.txt|go.mod|go.sum|\
		pom.xml|build.gradle|package-lock.json|yarn.lock|Cargo.lock|deno.lock|composer.lock|Gemfile.lock|Pipfile.lock"
	@$(DETECT_SECRETS) audit .secrets.baseline

##@ Build

build:
	@echo "Building the $(IMAGE_NAME) binary for $(LOCAL_ARCH)..."
	@GOARCH=$(LOCAL_ARCH) common/scripts/gobuild.sh bin/$(IMAGE_NAME) ./main.go

build-push-image: build-image push-image

build-image: $(CONFIG_DOCKER_TARGET) build
	@echo $(DOCKER_BUILD_OPTS)
	@echo "Building the $(IMAGE_NAME) docker image for $(LOCAL_ARCH)..."
	@docker build --platform linux/$(LOCAL_ARCH) -t $(REGISTRY)/$(IMAGE_NAME)-$(LOCAL_ARCH):$(VERSION) $(DOCKER_BUILD_OPTS) -f Dockerfile .

push-image: $(CONFIG_DOCKER_TARGET) build-image
	@echo "Pushing the $(IMAGE_NAME) docker image for $(LOCAL_ARCH)..."
	@docker push $(REGISTRY)/$(IMAGE_NAME)-$(LOCAL_ARCH):$(VERSION)

build-push-image-development: build-image-development push-image-development ## Build, push image

build-image-development: $(CONFIG_DOCKER_TARGET) build ## Create a docker image locally
	@echo $(DOCKER_BUILD_OPTS)
	@echo "Building the $(IMAGE_NAME) docker image for $(LOCAL_ARCH)..."
	@docker build --platform linux/$(LOCAL_ARCH) -t $(SCRATCH_REGISTRY)/$(IMAGE_NAME)-$(LOCAL_ARCH):$(VERSION) $(DOCKER_BUILD_OPTS) -f Dockerfile .

push-image-development: $(CONFIG_DOCKER_TARGET) build-image-development ## Push previously created image to scratch registry
	@echo "Pushing the $(IMAGE_NAME) docker image for $(LOCAL_ARCH)..."
	@docker push $(SCRATCH_REGISTRY)/$(IMAGE_NAME)-$(LOCAL_ARCH):$(VERSION)

##@ SHA Digest section

.PHONY: get-image-sha
get-image-sha: ## Replaces operand tag for digest in operator.yaml and csv
	@echo Get SHA for ibm-licensing:$(OPERAND_TAG)
	@common/scripts/get-image-sha.sh $(OPERAND_REGISTRY)/ibm-licensing $(OPERAND_TAG)

##@ Release

multiarch-image: $(CONFIG_DOCKER_TARGET)
	@MAX_PULLING_RETRY=20 RETRY_INTERVAL=30 common/scripts/multiarch_image.sh $(REGISTRY) $(IMAGE_NAME) $(VERSION) ${MANIFEST_VERSION}

multiarch-image-latest: $(CONFIG_DOCKER_TARGET)
	@MAX_PULLING_RETRY=20 RETRY_INTERVAL=30 common/scripts/multiarch_image_latest.sh $(REGISTRY) $(IMAGE_NAME) $(VERSION)

multiarch-image-development: $(CONFIG_DOCKER_TARGET)
	@MAX_PULLING_RETRY=20 RETRY_INTERVAL=30 common/scripts/multiarch_image.sh $(SCRATCH_REGISTRY) $(IMAGE_NAME) $(VERSION) ${VERSION} ${GIT_BRANCH}

csv: ## Push CSV package to the catalog
	@RELEASE=${CSV_VERSION} common/scripts/push-csv.sh

##@ Red Hat Certificate Section

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

clean: ## Clean build binary and all installed tools
	rm -rf bin/

##@ Help

help: ## Display this help
	@echo "Usage:  make <target>"
	@awk 'BEGIN {FS = ":.*##"}; \
		/^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } \
		/^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

prepare-unit-test:
	kubectl create namespace ${NAMESPACE} || echo ""
	kubectl create namespace ${OPREQ_TEST_NAMESPACE} || echo ""
	kubectl create secret docker-registry artifactory-token -n ${NAMESPACE} --docker-server=${REGISTRY} --docker-username=${ARTIFACTORY_USERNAME} --docker-password=${ARTIFACTORY_TOKEN} || echo "" ;\
	kubectl apply -f ./config/crd/bases/operator.ibm.com_ibmlicensings.yaml || echo ""
	sed "s/ibm-licensing/${NAMESPACE}/g" < ./config/rbac/role.yaml > ./config/rbac/role_ns.yaml
	kubectl apply -f ./config/rbac/role_ns.yaml || echo ""
	sed "s/ibm-licensing/${NAMESPACE}/g" < ./config/rbac/service_account.yaml > ./config/rbac/service_account_ns.yaml
	kubectl apply -f ./config/rbac/service_account_ns.yaml|| echo ""
	sed "s/ibm-licensing/${NAMESPACE}/g" < ./config/rbac/role_binding.yaml > ./config/rbac/role_binding_ns.yaml
	kubectl apply -f ./config/rbac/role_binding_ns.yaml || echo ""
	curl -O https://raw.githubusercontent.com/redhat-marketplace/redhat-marketplace-operator/674d4e57186b/v2/config/crd/bases/marketplace.redhat.com_meterdefinitions.yaml
	kubectl apply -f marketplace.redhat.com_meterdefinitions.yaml
	curl -O https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/master/example/prometheus-operator-crd/monitoring.coreos.com_servicemonitors.yaml
	kubectl apply -f monitoring.coreos.com_servicemonitors.yaml
	curl -O https://raw.githubusercontent.com/IBM/operand-deployment-lifecycle-manager/v1.23.5/bundle/manifests/operator.ibm.com_operandrequests.yaml
	kubectl apply -f operator.ibm.com_operandrequests.yaml
	kubectl apply -f https://raw.githubusercontent.com/kubernetes-sigs/gateway-api/v1.5.0/config/crd/standard/gateway.networking.k8s.io_gatewayclasses.yaml
	kubectl apply -f https://raw.githubusercontent.com/kubernetes-sigs/gateway-api/v1.5.0/config/crd/standard/gateway.networking.k8s.io_gateways.yaml
	kubectl apply -f https://raw.githubusercontent.com/kubernetes-sigs/gateway-api/v1.5.0/config/crd/standard/gateway.networking.k8s.io_httproutes.yaml
	kubectl apply -f https://raw.githubusercontent.com/kubernetes-sigs/gateway-api/v1.5.0/config/crd/standard/gateway.networking.k8s.io_backendtlspolicies.yaml

unit-test: prepare-unit-test
	export USE_EXISTING_CLUSTER=true; \
	export OPERATOR_NAMESPACE=${NAMESPACE}; \
	export WATCH_NAMESPACE=${NAMESPACE}; \
	export NAMESPACE=${NAMESPACE}; \
	export OPREQ_TEST_NAMESPACE=${OPREQ_TEST_NAMESPACE}; \
	export OCP=${OCP}; \
	export KUBEBUILDER_ATTACH_CONTROL_PLANE_OUTPUT=true; \
	export IBM_LICENSING_IMAGE=${REGISTRY}/${IBM_LICENSING_IMAGE}:${CSV_VERSION}; \
	go test -v ./controllers/... -coverprofile cover.out -timeout 30m

# Build manager binary
manager: generate
	go build -o bin/$(IMAGE_NAME) main.go

# Run against the configured Kubernetes cluster in ~/.kube/config. Adjust namespace variable according to your environment, e.g. NAMESPACE=lsr-ns make run
run: fmt vet
	export IBM_LICENSING_IMAGE=${REGISTRY}/${IBM_LICENSING_IMAGE}:${CSV_VERSION}; \
	WATCH_NAMESPACE=${NAMESPACE} OPERATOR_NAMESPACE=${NAMESPACE} go run ./main.go

# Install CRDs into a cluster
install: manifests kustomize
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

# Uninstall CRDs from a cluster
uninstall: manifests kustomize
	$(KUSTOMIZE) build config/crd | kubectl delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests kustomize
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | kubectl apply -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen yq
	$(YQ) -i '.metadata.annotations."olm.skipRange" = ">=1.0.0 <$(CSV_VERSION)"' ./config/manifests/bases/ibm-licensing-operator.clusterserviceversion.yaml
	$(YQ) -i '.metadata.annotations.containerImage = "icr.io/cpopen/${IMG}:$(CSV_VERSION)"' ./config/manifests/bases/ibm-licensing-operator.clusterserviceversion.yaml
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=ibm-licensing-operator webhook paths="./api/..." paths="./controllers/..." output:crd:artifacts:config=config/crd/bases

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./api/..."

# Build the docker image
docker-build: test
	docker build . -t ${IMG}

# Push the docker image
docker-push:
	docker push ${IMG}

# Take the roles (e.g. permissions) from bundle manifest that are created by kubebuilder and put them in CSV
update-roles-alm-example: alm-example yq
	mkdir -p $(LOCAL_TMP)
	$(YQ) -P '.rules' ./bundle/manifests/ibm-license-service_rbac.authorization.k8s.io_v1_clusterrole.yaml > $(LOCAL_TMP)/clusterrole.yaml
	$(YQ) -P '.rules' ./bundle/manifests/ibm-license-service_rbac.authorization.k8s.io_v1_role.yaml > $(LOCAL_TMP)/role.yaml
	$(YQ) -P '.rules' ./bundle/manifests/ibm-licensing-default-reader_rbac.authorization.k8s.io_v1_clusterrole.yaml > $(LOCAL_TMP)/reader-clusterrole.yaml
	$(YQ) -P '.rules' ./bundle/manifests/ibm-license-service-restricted_rbac.authorization.k8s.io_v1_clusterrole.yaml > $(LOCAL_TMP)/clusterrole2.yaml
	$(YQ) -P '.rules' ./bundle/manifests/ibm-license-service-restricted_rbac.authorization.k8s.io_v1_role.yaml > $(LOCAL_TMP)/role2.yaml

	sed -i -e 's/^/  /' $(LOCAL_TMP)/clusterrole.yaml
	sed -i -e 's/^/  /' $(LOCAL_TMP)/role.yaml
	sed -i -e 's/^/  /' $(LOCAL_TMP)/reader-clusterrole.yaml
	sed -i -e 's/^/  /' $(LOCAL_TMP)/clusterrole2.yaml
	sed -i -e 's/^/  /' $(LOCAL_TMP)/role2.yaml

	$(YQ) -i '.spec.install.spec.clusterPermissions[1].rules |= load("$(LOCAL_TMP)/clusterrole.yaml") | \
		.spec.install.spec.clusterPermissions[1].serviceAccountName = "ibm-license-service" \
	' ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml

	$(YQ) -i '.spec.install.spec.clusterPermissions[2].rules |= load("$(LOCAL_TMP)/clusterrole2.yaml") | \
		.spec.install.spec.clusterPermissions[2].serviceAccountName = "ibm-license-service-restricted" \
	' ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml

	$(YQ) -i '.spec.install.spec.clusterPermissions[3].rules |= load("$(LOCAL_TMP)/reader-clusterrole.yaml") | \
		.spec.install.spec.clusterPermissions[3].serviceAccountName = "ibm-licensing-default-reader" \
	' ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml

	$(YQ) -i '.spec.install.spec.permissions[1].rules |= load("$(LOCAL_TMP)/role.yaml") | \
		.spec.install.spec.permissions[1].serviceAccountName = "ibm-license-service" \
	' ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml

	$(YQ) -i '.spec.install.spec.permissions[2].rules |= load("$(LOCAL_TMP)/role2.yaml") | \
		.spec.install.spec.permissions[2].serviceAccountName = "ibm-license-service-restricted" \
	' ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml

	rm -f ./bundle/manifests/ibm-license-service_rbac.authorization.k8s.io_v1_clusterrole.yaml
	rm -f ./bundle/manifests/ibm-licensing-default-reader_rbac.authorization.k8s.io_v1_clusterrole.yaml
	rm -f ./bundle/manifests/ibm-license-service_rbac.authorization.k8s.io_v1_role.yaml
	rm -f ./bundle/manifests/ibm-license-service_rbac.authorization.k8s.io_v1_clusterrolebinding.yaml
	rm -f ./bundle/manifests/ibm-licensing-default-reader_rbac.authorization.k8s.io_v1_clusterrolebinding.yaml
	rm -f ./bundle/manifests/ibm-license-service_rbac.authorization.k8s.io_v1_rolebinding.yaml
	rm -f ./bundle/manifests/ibm-license-service-restricted_rbac.authorization.k8s.io_v1_clusterrole.yaml
	rm -f ./bundle/manifests/ibm-license-service-restricted_rbac.authorization.k8s.io_v1_role.yaml
	rm -f ./bundle/manifests/ibm-license-service-restricted_rbac.authorization.k8s.io_v1_clusterrolebinding.yaml
	rm -f ./bundle/manifests/ibm-license-service-restricted_rbac.authorization.k8s.io_v1_rolebinding.yaml
	rm -f ./bundle/manifests/ibm-licensing-operator_v1_serviceaccount.yaml
	rm -f ./bundle/manifests/ibm-licensing-default-reader_v1_serviceaccount.yaml
	rm -f ./bundle/manifests/ibm-license-service_v1_serviceaccount.yaml
	rm -f ./bundle/manifests/ibm-license-service-restricted_v1_serviceaccount.yaml

# Takes config samples CRs and update alm-exmaple in CSV
alm-example: yq
	mkdir -p $(LOCAL_TMP)/json
	$(YQ) -P -o=json ./config/samples/operator.ibm.com_v1alpha1_ibmlicensing.yaml > $(LOCAL_TMP)/json/ibmlicensing.json
	$(YQ) -P -o=json ./config/samples/operator_v1_ibmlicensingdefinition.yaml > $(LOCAL_TMP)/json/ibmlicensingdefinition.json
	$(YQ) -P -o=json ./config/samples/operator_v1alpha1_ibmlicensingmetadata.yaml > $(LOCAL_TMP)/json/ibmlicensingmetadata.json
	$(YQ) -P -o=json ./config/samples/operator_v1_ibmlicensingquerysource.yaml > $(LOCAL_TMP)/json/ibmlicensingquerysource.json

	jq -s '.' $(LOCAL_TMP)/json/ibmlicensing.json $(LOCAL_TMP)/json/ibmlicensingdefinition.json $(LOCAL_TMP)/json/ibmlicensingmetadata.json $(LOCAL_TMP)/json/ibmlicensingquerysource.json > $(LOCAL_TMP)/json/merged.json
	$(YQ) -i '.metadata.annotations.alm-examples |= load_str("$(LOCAL_TMP)/json/merged.json")' ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml

	rm -r $(LOCAL_TMP)/json

# Generate bundle manifests and metadata, then validate generated files. Yq is used to change order of owned resources here to ensure Licensing is first.
pre-bundle: generate manifests operator-sdk kustomize yq
	$(OPERATOR_SDK) generate kustomize manifests -q
	$(KUSTOMIZE) build config/manifests | $(OPERATOR_SDK) generate bundle -q --overwrite --version $(CSV_VERSION) $(BUNDLE_METADATA_OPTS)
	$(YQ) -i '.annotations."com.redhat.openshift.versions" = "v4.12"' ./bundle/metadata/annotations.yaml
	$(YQ) '.spec.customresourcedefinitions.owned[0]' ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml > yq_tmp_definitions.yaml
	$(YQ) '.spec.customresourcedefinitions.owned[1]' ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml > yq_tmp_metadata.yaml
	$(YQ) '.spec.customresourcedefinitions.owned[2]' ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml > yq_tmp_querysources.yaml
	$(YQ) '.spec.customresourcedefinitions.owned[3]' ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml > yq_tmp_licensing.yaml
	$(YQ) -i eval-all 'select(fileIndex==0).spec.customresourcedefinitions.owned[0] = select(fileIndex==1) | select(fileIndex==0)' ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml yq_tmp_licensing.yaml
	$(YQ) -i eval-all 'select(fileIndex==0).spec.customresourcedefinitions.owned[1] = select(fileIndex==1) | select(fileIndex==0)' ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml yq_tmp_definitions.yaml
	$(YQ) -i eval-all 'select(fileIndex==0).spec.customresourcedefinitions.owned[2] = select(fileIndex==1) | select(fileIndex==0)' ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml yq_tmp_metadata.yaml
	$(YQ) -i eval-all 'select(fileIndex==0).spec.customresourcedefinitions.owned[3] = select(fileIndex==1) | select(fileIndex==0)' ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml yq_tmp_querysources.yaml
	$(YQ) -i '.spec.relatedImages = load("./common/relatedImages.yaml")' ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml

	rm yq_tmp_licensing.yaml yq_tmp_metadata.yaml yq_tmp_definitions.yaml yq_tmp_querysources.yaml

bundle: pre-bundle update-roles-alm-example
	$(OPERATOR_SDK) bundle validate ./bundle

# Build the bundle image.
bundle-build:
	docker build -f bundle.Dockerfile -t ${REGISTRY}/${BUNDLE_IMG} .

# Build the bundle image.
bundle-build-development:
	docker build -f bundle.Dockerfile -t ${SCRATCH_REGISTRY}/${BUNDLE_IMG} .

scorecard: operator-sdk
	kubectl create serviceaccount scorecard-sa -n ${NAMESPACE} || true
	kubectl create clusterrolebinding scorecard-admin --clusterrole=cluster-admin --serviceaccount=${NAMESPACE}:scorecard-sa || true
	$(OPERATOR_SDK) scorecard ./bundle -n ${NAMESPACE} -w 120s --service-account scorecard-sa --kubeconfig ${HOME}/.kube/config

catalogsource: opm yq
	@echo "Build CatalogSource for $(LOCAL_ARCH)...- ${BUNDLE_IMG} - ${CATALOG_IMG}"
	$(YQ) -i '.spec.install.spec.deployments[0].spec.template.spec.containers[0].image = "${REGISTRY}/${IMG}:${CSV_VERSION}"' ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml
	$(YQ) -i '.spec.install.spec.deployments[0].spec.template.spec.containers[0].env[0].value = "${REGISTRY}/${IBM_LICENSING_IMAGE}:${CSV_VERSION}"' ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml
	$(YQ) -i '.annotations."operators.operatorframework.io.bundle.channels.v1" =  "${CHANNELS}"' ./bundle/metadata/annotations.yaml
	$(YQ) -i '.annotations."operators.operatorframework.io.bundle.channel.default.v1" =  "${DEFAULT_CHANNEL}"' ./bundle/metadata/annotations.yaml
	docker build -f bundle.Dockerfile -t ${REGISTRY}/${BUNDLE_IMG} .
	docker push ${REGISTRY}/${BUNDLE_IMG}
	$(OPM) index add --permissive -c ${PODMAN} --bundles ${REGISTRY}/${BUNDLE_IMG} --tag ${REGISTRY}/${CATALOG_IMG}
	docker push ${REGISTRY}/${CATALOG_IMG}
ifneq (${DEVOPS_STREAM},)
	docker tag ${REGISTRY}/${CATALOG_IMG} ${REGISTRY}/${DEVOPS_CATALOG_IMG}
	docker push ${REGISTRY}/${DEVOPS_CATALOG_IMG}
endif

# pipeline builds the catalog for you and already makes a multi-arch catalog, for amd64 we build it conditionally for dev purposes
catalogsource-development: opm yq
	@echo "Build Development CatalogSource for $(LOCAL_ARCH)...- ${BUNDLE_IMG} - ${CATALOG_IMG}"
	$(YQ) -i '.spec.install.spec.deployments[0].spec.template.spec.containers[0].image = "${SCRATCH_REGISTRY}/${IMG}:${GIT_BRANCH}"' ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml
	$(YQ) -i '.spec.install.spec.deployments[0].spec.template.spec.containers[0].env[0].value = "${REGISTRY}/${IBM_LICENSING_IMAGE}:${CSV_VERSION}"' ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml
	$(YQ) -i '.annotations."operators.operatorframework.io.bundle.channels.v1" =  "${CHANNELS}"' ./bundle/metadata/annotations.yaml
	$(YQ) -i '.annotations."operators.operatorframework.io.bundle.channel.default.v1" =  "${DEFAULT_CHANNEL}"' ./bundle/metadata/annotations.yaml
	$(YQ) -i '.annotations."operators.operatorframework.io.bundle.package.v1" = "${PACKAGE}' ./bundle/metadata/annotations.yaml
	@echo "Verifying bundle annotations..."
	@$(YQ) '.annotations' ./bundle/metadata/annotations.yaml
	docker build -f bundle.Dockerfile -t ${SCRATCH_REGISTRY}/${BUNDLE_IMG} .
	docker push ${SCRATCH_REGISTRY}/${BUNDLE_IMG}
	@echo "Building catalog from scratch (without --from-index)..."
	$(OPM) index add --container-tool docker --bundles ${SCRATCH_REGISTRY}/${BUNDLE_IMG} --tag ${SCRATCH_REGISTRY}/${CATALOG_IMG}
	@echo "Verifying catalog contents..."
	@docker run --rm ${SCRATCH_REGISTRY}/${CATALOG_IMG} ls -la /database/ || echo "Warning: Could not verify catalog database"
	docker push ${SCRATCH_REGISTRY}/${CATALOG_IMG}

############################################################
# Installation section
############################################################

##@ Install

install-linters: $(SHELLCHECK) $(YAMLLINT) $(GOLANGCI_LINT) $(MDL) ## Install/verify required linting tools

verify-installed-tools: ## Verify if tools are installed
	@test -x $(OPERATOR_SDK) || { echo >&2 "Required tool: operator-sdk-$(OPERATOR_SDK_VERSION) is not installed in $(LOCALBIN). Run 'make install-all-tools' to install it."; exit 1; }
	@test -x $(OPM) || { echo >&2 "Required tool: opm-$(OPM_VERSION) is not installed in $(LOCALBIN). Run 'make install-all-tools' to install it."; exit 1; }
	@test -x $(CONTROLLER_GEN) || { echo >&2 "Required tool: controller-gen-$(CONTROLLER_GEN_VERSION) is not installed in $(LOCALBIN). Run 'make install-all-tools' to install it."; exit 1; }
	@test -x $(KUSTOMIZE) || { echo >&2 "Required tool: kustomize-$(KUSTOMIZE_VERSION) is not installed in $(LOCALBIN). Run 'make install-all-tools' to install it."; exit 1; }
	@test -x $(YQ) || { echo >&2 "Required tool: yq-$(YQ_VERSION) is not installed in $(LOCALBIN). Run 'make install-all-tools' to install it."; exit 1; }
	@test -x $(SHELLCHECK) || { echo >&2 "Required tool: shellcheck is not installed in $(LOCALBIN). Run 'make install-linters' to install it."; exit 1; }
	@test -x $(YAMLLINT) || { echo >&2 "Required tool: yamllint is not installed in $(LOCALBIN). Run 'make install-linters' to install it."; exit 1; }
	@test -x $(GOLANGCI_LINT) || { echo >&2 "Required tool: golangci-lint-$(GOLANGCI_LINT_VERSION) is not installed in $(LOCALBIN). Run 'make install-linters' to install it."; exit 1; }
	@test -x $(MDL) || { echo >&2 "Required tool: mdl is not installed in $(LOCALBIN). Run 'make install-linters' to install it."; exit 1; }
	@test -x $(GOIMPORTS) || { echo >&2 "Required tool: goimports-$(GOIMPORTS_VERSION) is not installed in $(LOCALBIN). Run 'make install-all-tools' to install it."; exit 1; }
	@test -x $(DETECT_SECRETS) || { echo >&2 "Required tool: detect-secrets is not installed in $(LOCALBIN). Run 'make install-all-tools' to install it."; exit 1; }
	@echo "Successfully verified all tools present in $(LOCALBIN)."
	@echo "Required | Installed"
	@echo ">>> operator-sdk-$(OPERATOR_SDK_VERSION) | $$($(OPERATOR_SDK) version | awk '{print $$3}')"
	@echo ">>> opm-$(OPM_VERSION) | $$($(OPM) version | awk '{print $$2}' | awk -F ':' '{print $$2}')"
	@echo ">>> controller-gen-$(CONTROLLER_GEN_VERSION) | $$($(CONTROLLER_GEN) --version | awk '{print $$2}')"
	@echo ">>> kustomize-$(KUSTOMIZE_VERSION) | $$($(KUSTOMIZE) version)"
	@echo ">>> yq-$(YQ_VERSION) | $$($(YQ) --version | awk '{print $$4}')"
	@echo ">>> mdl-$(MDL_VERSION) | $$($(MDL) --version)"

install-all-tools: install-operator-sdk install-opm install-controller-gen install-kustomize install-yq install-detect-secrets install-goimports install-linters verify-installed-tools ## Install all tools locally

.PHONY: install-operator-sdk
install-operator-sdk: $(OPERATOR_SDK) ## Install tool locally: operator-sdk

.PHONY: install-opm
install-opm: $(OPM) ## Install tool locally: opm

.PHONY: install-controller-gen
install-controller-gen: $(CONTROLLER_GEN) ## Install tool locally: controller-gen

.PHONY: install-kustomize
install-kustomize: $(KUSTOMIZE) ## Install tool locally: kustomize

.PHONY: install-yq
install-yq: $(YQ) ## Install tool locally: yq

.PHONY: install-detect-secrets
install-detect-secrets: $(DETECT_SECRETS) ## Install tool locally: detect-secrets

.PHONY: install-goimports
install-goimports: $(GOIMPORTS) ## Install tool locally: goimports

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Install controller-gen if not present in LOCALBIN

$(CONTROLLER_GEN): $(LOCALBIN)
	@test -x $(CONTROLLER_GEN) && $(CONTROLLER_GEN) --version 2>/dev/null | grep -q "$(CONTROLLER_GEN_VERSION)" && echo "controller-gen $(CONTROLLER_GEN_VERSION) already installed" || \
		( echo "Installing controller-gen $(CONTROLLER_GEN_VERSION)..." && GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_GEN_VERSION) )

.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Install kustomize if not present in LOCALBIN

$(KUSTOMIZE): $(LOCALBIN)
	@test -x $(KUSTOMIZE) && $(KUSTOMIZE) version 2>/dev/null | grep -q "$(KUSTOMIZE_VERSION)" && echo "kustomize $(KUSTOMIZE_VERSION) already installed" || \
		( echo "Installing kustomize $(KUSTOMIZE_VERSION)..." && GOBIN=$(LOCALBIN) go install sigs.k8s.io/kustomize/kustomize/v5@$(KUSTOMIZE_VERSION) )

.PHONY: opm
opm: $(OPM) ## Install opm if not present in LOCALBIN

$(OPM): $(LOCALBIN)
	@test -x $(OPM) && $(OPM) version 2>/dev/null | grep -q "$(OPM_VERSION)" && echo "opm $(OPM_VERSION) already installed" || \
		( echo "Installing opm $(OPM_VERSION)..." && \
		  curl -sSfL https://github.com/operator-framework/operator-registry/releases/download/$(OPM_VERSION)/$(TARGET_OS)-$(LOCAL_ARCH)-opm -o $(OPM) && \
		  chmod +x $(OPM) )

.PHONY: operator-sdk
operator-sdk: $(OPERATOR_SDK) ## Install operator-sdk if not present in LOCALBIN

$(OPERATOR_SDK): $(LOCALBIN)
	@test -x $(OPERATOR_SDK) && $(OPERATOR_SDK) version 2>/dev/null | grep -q "$(OPERATOR_SDK_VERSION)" && echo "operator-sdk $(OPERATOR_SDK_VERSION) already installed" || \
		( echo "Installing operator-sdk $(OPERATOR_SDK_VERSION)..." && \
		  bash common/scripts/install-operator-sdk.sh $(TARGET_OS) $(LOCAL_ARCH) $(OPERATOR_SDK_VERSION) $(OPERATOR_SDK) )

.PHONY: yq
yq: $(YQ) ## Install yq if not present in LOCALBIN

$(YQ): $(LOCALBIN)
	@test -x $(YQ) && $(YQ) --version 2>/dev/null | grep -q "$(YQ_VERSION)" && echo "yq $(YQ_VERSION) already installed" || \
		( echo "Installing yq $(YQ_VERSION)..." && \
		  bash common/scripts/install-yq.sh $(TARGET_OS) $(LOCAL_ARCH) $(YQ_VERSION) $(YQ) )

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT) ## Install golangci-lint if not present in LOCALBIN

$(GOLANGCI_LINT): $(LOCALBIN)
	@test -x $(GOLANGCI_LINT) && $(GOLANGCI_LINT) --version 2>/dev/null | grep -q "$(GOLANGCI_LINT_VERSION:v%=%)" && echo "golangci-lint $(GOLANGCI_LINT_VERSION) already installed" || \
		( echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION)..." && \
		  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(LOCALBIN) $(GOLANGCI_LINT_VERSION) )

.PHONY: goimports
goimports: $(GOIMPORTS) ## Install goimports if not present in LOCALBIN

$(GOIMPORTS): $(LOCALBIN)
	@test -x $(GOIMPORTS) && echo "goimports already installed" || \
		( echo "Installing goimports $(GOIMPORTS_VERSION)..." && \
		  GOBIN=$(LOCALBIN) go install golang.org/x/tools/cmd/goimports@$(GOIMPORTS_VERSION) )

$(DETECT_SECRETS): $(LOCALBIN)
	@test -x $(DETECT_SECRETS) && echo "detect-secrets already installed" || \
		bash common/scripts/install-detect-secrets.sh $(LOCALBIN)

.PHONY: shellcheck
shellcheck: $(SHELLCHECK) ## Install shellcheck if not present in LOCALBIN

$(SHELLCHECK): $(LOCALBIN)
	@test -x $(SHELLCHECK) && echo "shellcheck already installed" || \
		( echo "Installing shellcheck $(SHELLCHECK_VERSION)..." && bash common/scripts/install-shellcheck.sh $(SHELLCHECK_VERSION) $(SHELLCHECK) )

.PHONY: yamllint
yamllint: $(YAMLLINT) ## Install yamllint if not present in LOCALBIN

$(YAMLLINT): $(LOCALBIN)
	@test -x $(YAMLLINT) && echo "yamllint already installed" || \
		( echo "Installing yamllint $(YAMLLINT_VERSION)..." && bash common/scripts/install-yamllint.sh $(LOCALBIN) $(YAMLLINT_VERSION) )

.PHONY: mdl
mdl: $(MDL) ## Install mdl if not present in LOCALBIN

$(MDL): $(LOCALBIN)
	@test -x $(MDL) && echo "mdl already installed" || \
		( echo "Installing mdl $(MDL_VERSION)..." && bash common/scripts/install-mdl.sh $(LOCALBIN) $(MDL_VERSION) )

ifeq (, $(shell which podman))
PODMAN=docker
else
PODMAN=podman
endif

.PHONY: all opm build bundle-build bundle pre-bundle kustomize catalogsource controller-gen generate docker-build docker-push deploy manifests run install uninstall code-dev check lint test coverage-kind coverage build multiarch-image csv clean help operator-sdk yq golangci-lint goimports shellcheck yamllint hadolint mdl install-all-tools install-operator-sdk install-opm install-controller-gen install-kustomize install-yq install-detect-secrets install-goimports install-linters verify-installed-tools audit scorecard

.PHONY: generate-yaml-argo-cd
generate-yaml-argo-cd: kustomize yq
	@mkdir -p argo-cd && $(KUSTOMIZE) build config/manifests > argo-cd/tmp.yaml

	# Split the resources into separate YAML files
	@(echo "---" && $(YQ) 'select(.kind == "ClusterRole" or .kind == "ClusterRoleBinding")' argo-cd/tmp.yaml) > argo-cd/cluster-rbac.yaml
	@(echo "---" && $(YQ) 'select(.kind == "IBMLicensing")' argo-cd/tmp.yaml) > argo-cd/cr.yaml
	@(echo "---" && $(YQ) 'select(.kind == "CustomResourceDefinition")' argo-cd/tmp.yaml) > argo-cd/crd.yaml
	@(echo "---" && $(YQ) 'select(.kind == "Deployment")' argo-cd/tmp.yaml) > argo-cd/deployment.yaml
	@(echo "---" && $(YQ) 'select(.kind == "Role" or .kind == "RoleBinding")' argo-cd/tmp.yaml) > argo-cd/rbac.yaml
	@(echo "---" && $(YQ) 'select(.kind == "ServiceAccount")' argo-cd/tmp.yaml) > argo-cd/serviceaccounts.yaml

	# Add missing namespaces
	@$(YQ) -i 'select(.kind == "ClusterRoleBinding").subjects[0].namespace = "sed-me"' argo-cd/cluster-rbac.yaml
	@$(YQ) -i 'select(.kind == "RoleBinding").subjects[0].namespace = "sed-me"' argo-cd/rbac.yaml

	# Remove redundant data
	@$(YQ) -i 'del(.metadata.namespace)' argo-cd/cluster-rbac.yaml

	# Prepare resources for templating with helm
	@$(YQ) -i 'del(.spec)' argo-cd/cr.yaml
	@$(YQ) -i '.metadata.annotations.sed-deployment-annotations-top = "sed-me" \
	| .metadata.labels.sed-deployment-labels-top = "sed-me" \
	| .spec.template.metadata.annotations.sed-deployment-annotations-bottom = "sed-me" \
	| .spec.template.metadata.labels.sed-deployment-labels-bottom = "sed-me" \
	| .spec.template.spec.containers[0].env[1].valueFrom = "sed-me"' argo-cd/deployment.yaml
	@$(YQ) -i '.metadata.labels.component-id = "sed-me"' argo-cd/cluster-rbac.yaml
	@$(YQ) -i '.metadata.labels.component-id = "sed-me"' argo-cd/cr.yaml
	@$(YQ) -i '.metadata.labels.component-id = "sed-me"' argo-cd/crd.yaml
	@$(YQ) -i '.metadata.labels.component-id = "sed-me"' argo-cd/deployment.yaml
	@$(YQ) -i '.metadata.labels.component-id = "sed-me"' argo-cd/rbac.yaml
	@$(YQ) -i '.metadata.labels.component-id = "sed-me"' argo-cd/serviceaccounts.yaml

	# Add extra fields, for example argo-cd sync waves
	@$(YQ) -i '.metadata.annotations."argocd.argoproj.io/sync-options" = "ServerSideApply=true"' argo-cd/cr.yaml
	@$(YQ) -i '.metadata.annotations."argocd.argoproj.io/sync-wave" = "-1"' argo-cd/crd.yaml
	# This sync wave is crucial because the deployment must be created after the CR, to avoid a situation when ArgoCD
	# starts creating the CR at the same time as the operator does it (patch isn't applied and a name conflict happens)
	@$(YQ) -i '.metadata.annotations."argocd.argoproj.io/sync-wave" = "1"' argo-cd/deployment.yaml

	# Replace all component-id labels to template them with helm
	@sed -i '' "s/component-id: sed-me/component-id: {{ .Chart.Name }}/g" argo-cd/cluster-rbac.yaml
	@sed -i '' "s/component-id: sed-me/component-id: {{ .Chart.Name }}/g" argo-cd/cr.yaml
	@sed -i '' "s/component-id: sed-me/component-id: {{ .Chart.Name }}/g" argo-cd/crd.yaml
	@sed -i '' "s/component-id: sed-me/component-id: {{ .Chart.Name }}/g" argo-cd/deployment.yaml
	@sed -i '' "s/component-id: sed-me/component-id: {{ .Chart.Name }}/g" argo-cd/rbac.yaml
	@sed -i '' "s/component-id: sed-me/component-id: {{ .Chart.Name }}/g" argo-cd/serviceaccounts.yaml

	# Replace all namespaces to template them with helm
	@sed -i '' "s/namespace: [^ ]*/namespace: {{ .Values.ibmLicensing.namespace }}/g" argo-cd/cluster-rbac.yaml
	@sed -i '' "s/namespace: [^ ]*/namespace: {{ .Values.ibmLicensing.namespace }}/g" argo-cd/deployment.yaml
	@sed -i '' "s/namespace: [^ ]*/namespace: {{ .Values.ibmLicensing.namespace }}/g" argo-cd/rbac.yaml
	@sed -i '' "s/namespace: [^ ]*/namespace: {{ .Values.ibmLicensing.namespace }}/g" argo-cd/serviceaccounts.yaml

	# Replace all registry occurrences to template them with helm
	@sed -i '' "s/icr.io/{{ .Values.global.imagePullPrefix }}/g" argo-cd/deployment.yaml
	@sed -i '' "s/cpopen\/cpfs/{{ .Values.ibmLicensing.imageRegistryNamespaceOperand }}/g" argo-cd/deployment.yaml
	@sed -i '' "s/cpopen/{{ .Values.ibmLicensing.imageRegistryNamespaceOperator }}/g" argo-cd/deployment.yaml

	# Replace extra fields (in addition to the namespaces) to template them with helm
	@cat ./common/makefile-generate/yaml-cr-spec-part >> argo-cd/cr.yaml
	@sed -i '' "s/sed-deployment-annotations-top: sed-me/{{- if ((.Values.ibmLicensing.operator).annotations) }}\n      {{- toYaml .Values.ibmLicensing.operator.annotations | nindent 4 -}}\n    {{ end }}/g" argo-cd/deployment.yaml
	@sed -i '' "s/sed-deployment-labels-top: sed-me/{{- if ((.Values.ibmLicensing.operator).labels) }}\n      {{- toYaml .Values.ibmLicensing.operator.labels | nindent 4 -}}\n    {{ end }}/g" argo-cd/deployment.yaml
	@sed -i '' "s/sed-deployment-annotations-bottom: sed-me/{{- if ((.Values.ibmLicensing.operator).annotations) }}\n          {{- toYaml .Values.ibmLicensing.operator.annotations | nindent 8 -}}\n        {{ end }}/g" argo-cd/deployment.yaml
	@sed -i '' "s/sed-deployment-labels-bottom: sed-me/{{- if ((.Values.ibmLicensing.operator).labels) }}\n          {{- toYaml .Values.ibmLicensing.operator.labels | nindent 8 -}}\n        {{ end }}/g" argo-cd/deployment.yaml
	@sed -i '' "s/valueFrom: sed-me/value: {{ .Values.ibmLicensing.watchNamespace }}/g" argo-cd/deployment.yaml
	@cat ./common/makefile-generate/yaml-deployment-pull-secrets-part >> argo-cd/deployment.yaml

	@rm argo-cd/tmp.yaml
