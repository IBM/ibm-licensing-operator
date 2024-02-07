#
# Copyright 2023 IBM Corporation
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
CSV_VERSION ?= 1.20.12
CSV_VERSION_DEVELOPMENT ?= development
OLD_CSV_VERSION ?= 1.20.11

# Tools versions
OPM_VERSION ?= v1.26.2
OPERATOR_SDK_VERSION ?= v1.30
YQ_VERSION ?= v4.30.5
KUSTOMIZE_VERSION ?= v4.5.7
CONTROLLER_GEN_VERSION ?= v0.10.0

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
IBM_LICENSE_SERVICE_REPORTER_UI_IMAGE ?= ibm-license-service-reporter-ui
IBM_POSTGRESQL_IMAGE ?= ibm-postgresql
IBM_LICENSE_SERVICE_REPORTER_IMAGE ?= ibm-license-service-reporter
IBM_LICENSING_USAGE_IMAGE ?= ibm-licensing-usage

CHANNELS=v3,v3.20,v3.21,v3.22,v3.23,beta,dev,stable-v1
DEFAULT_CHANNEL=v3

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

# Linter urls that should be skipped
MARKDOWN_LINT_WHITELIST ?= https://quay.io/cnr,https://www-03preprod.ibm.com/support/knowledgecenter/SSHKN6/installer/3.3.0/install_operator.html,https://github.com/IBM/ibm-licensing-operator/releases/download/,https://github.com/operator-framework/operator-lifecycle-manager/releases/download,http://ibm.biz/,https://ibm.biz/,https://goreportcard.com/,https://docs.vmware.com/en/VMware-vSphere/7.0/vmware-vsphere-with-tanzu/GUID-CD033D1D-BAD2-41C4-A46F-647A560BAEAB.html,https://docs.vmware.com/en/VMware-vSphere/7.0/vmware-vsphere-with-tanzu/GUID-4CCDBB85-2770-4FB8-BF0E-5146B45C9543.html

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
else ifeq ($(ARCH),arm64)
    LOCAL_ARCH="arm64"
else
    $(error "This system's ARCH $(ARCH) isn't recognized/supported")
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

# Identify stream based in current git branch
DEVOPS_STREAM :=
ifeq ($(GIT_BRANCH),master) 
	DEVOPS_STREAM="cd"
	DEFAULT_CHANNEL=v4.0
else ifeq ($(GIT_BRANCH),release-ltsr)
	DEVOPS_STREAM="ltsr"
	DEFAULT_CHANNEL=v3
else ifeq ($(GIT_BRANCH),release-future)
	DEVOPS_STREAM="future"
	DEFAULT_CHANNEL=v4.0
endif

DEVOPS_CATALOG_IMG ?= $(IMAGE_CATALOG_NAME)-$(LOCAL_ARCH):$(DEVOPS_STREAM)

$(eval DOCKER_BUILD_OPTS := --build-arg "IMAGE_NAME=$(IMAGE_NAME)" --build-arg "IMAGE_DISPLAY_NAME=$(IMAGE_DISPLAY_NAME)" --build-arg "IMAGE_MAINTAINER=$(IMAGE_MAINTAINER)" --build-arg "IMAGE_VENDOR=$(IMAGE_VENDOR)" --build-arg "IMAGE_VERSION=$(IMAGE_VERSION)" --build-arg "VERSION=$(CSV_VERSION)" --build-arg "IMAGE_RELEASE=$(IMAGE_RELEASE)"  --build-arg "IMAGE_BUILDDATE=$(IMAGE_BUILDDATE)" --build-arg "IMAGE_DESCRIPTION=$(IMAGE_DESCRIPTION)" --build-arg "IMAGE_SUMMARY=$(IMAGE_SUMMARY)" --build-arg "IMAGE_OPENSHIFT_TAGS=$(IMAGE_OPENSHIFT_TAGS)" --build-arg "VCS_REF=$(VCS_REF)" --build-arg "VCS_URL=$(GIT_REMOTE_URL)" --build-arg "IMAGE_NAME_ARCH=$(IMAGE_NAME)-$(LOCAL_ARCH)")

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

$(GOBIN):
	@echo "create gobin"
	@mkdir -p $(GOBIN)

work: $(GOBIN)

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

##@ Build

build:
	@echo "Building the $(IMAGE_NAME) binary for $(LOCAL_ARCH)..."
	@GOARCH=$(LOCAL_ARCH) common/scripts/gobuild.sh bin/$(IMAGE_NAME) ./main.go
	@strip $(STRIP_FLAGS) bin/$(IMAGE_NAME)

build-push-image: build-image push-image catalogsource

build-image: $(CONFIG_DOCKER_TARGET) build
	@echo $(DOCKER_BUILD_OPTS)
	@echo "Building the $(IMAGE_NAME) docker image for $(LOCAL_ARCH)..."
	@docker build -t $(REGISTRY)/$(IMAGE_NAME)-$(LOCAL_ARCH):$(VERSION) $(DOCKER_BUILD_OPTS) -f Dockerfile .

push-image: $(CONFIG_DOCKER_TARGET) build-image
	@echo "Pushing the $(IMAGE_NAME) docker image for $(LOCAL_ARCH)..."
	@docker push $(REGISTRY)/$(IMAGE_NAME)-$(LOCAL_ARCH):$(VERSION)

build-push-image-development: build-image-development push-image-development catalogsource-development ## Build, push image and catalogsource

build-image-development: $(CONFIG_DOCKER_TARGET) build ## Create a docker image locally
	@echo $(DOCKER_BUILD_OPTS)
	@echo "Building the $(IMAGE_NAME) docker image for $(LOCAL_ARCH)..."
	@docker build -t $(SCRATCH_REGISTRY)/$(IMAGE_NAME)-$(LOCAL_ARCH):$(VERSION) $(DOCKER_BUILD_OPTS) -f Dockerfile .

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
	common/scripts/catalog_build.sh $(REGISTRY) $(IMAGE_NAME) ${MANIFEST_VERSION}
	@MAX_PULLING_RETRY=20 RETRY_INTERVAL=30 common/scripts/multiarch_image.sh $(REGISTRY) $(IMAGE_CATALOG_NAME) $(VERSION) ${MANIFEST_VERSION}

multiarch-image-latest: $(CONFIG_DOCKER_TARGET)
	@MAX_PULLING_RETRY=20 RETRY_INTERVAL=30 common/scripts/multiarch_image_latest.sh $(REGISTRY) $(IMAGE_NAME) $(VERSION)
	@MAX_PULLING_RETRY=20 RETRY_INTERVAL=30 common/scripts/multiarch_image_latest.sh $(REGISTRY) $(IMAGE_CATALOG_NAME) $(VERSION)

multiarch-image-development: $(CONFIG_DOCKER_TARGET)
	@MAX_PULLING_RETRY=20 RETRY_INTERVAL=30 common/scripts/multiarch_image.sh $(SCRATCH_REGISTRY) $(IMAGE_NAME) $(VERSION) ${VERSION} ${GIT_BRANCH}
	common/scripts/catalog_build.sh $(SCRATCH_REGISTRY) $(IMAGE_NAME) ${MANIFEST_VERSION}
	@MAX_PULLING_RETRY=20 RETRY_INTERVAL=30 common/scripts/multiarch_image.sh $(SCRATCH_REGISTRY) $(IMAGE_CATALOG_NAME) $(VERSION) ${VERSION} ${GIT_BRANCH}

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

clean: ## Clean build binary
	rm -f bin/$(IMG)

##@ Help

help: ## Display this help
	@echo "Usage:  make <target>"
	@awk 'BEGIN {FS = ":.*##"}; \
		/^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } \
		/^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
	
prepare-unit-test:
	kubectl create namespace ${NAMESPACE} || echo ""
	kubectl create secret docker-registry artifactory-token -n ${NAMESPACE} --docker-server=${REGISTRY} --docker-username=${ARTIFACTORY_USERNAME} --docker-password=${ARTIFACTORY_TOKEN} || echo "" ;\
	kubectl apply -f ./config/crd/bases/operator.ibm.com_ibmlicenseservicereporters.yaml || echo ""
	kubectl apply -f ./config/crd/bases/operator.ibm.com_ibmlicensings.yaml || echo ""
	sed "s/ibm-common-services/${NAMESPACE}/g" < ./config/rbac/role.yaml > ./config/rbac/role_ns.yaml
	kubectl apply -f ./config/rbac/role_ns.yaml || echo ""
	sed "s/ibm-common-services/${NAMESPACE}/g" < ./config/rbac/service_account.yaml > ./config/rbac/service_account_ns.yaml
	kubectl apply -f ./config/rbac/service_account_ns.yaml|| echo ""
	sed "s/ibm-common-services/${NAMESPACE}/g" < ./config/rbac/role_binding.yaml > ./config/rbac/role_binding_ns.yaml
	kubectl apply -f ./config/rbac/role_binding_ns.yaml || echo ""
	curl -O https://raw.githubusercontent.com/redhat-marketplace/redhat-marketplace-operator/master/v2/config/crd/bases/marketplace.redhat.com_meterdefinitions.yaml
	kubectl apply -f marketplace.redhat.com_meterdefinitions.yaml
	curl -O https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/master/example/prometheus-operator-crd/monitoring.coreos.com_servicemonitors.yaml
	kubectl apply -f monitoring.coreos.com_servicemonitors.yaml

unit-test: prepare-unit-test
	export USE_EXISTING_CLUSTER=true; \
	export WATCH_NAMESPACE=${NAMESPACE}; \
	export NAMESPACE=${NAMESPACE}; \
	export OCP=${OCP}; \
	export KUBEBUILDER_ATTACH_CONTROL_PLANE_OUTPUT=true; \
	export IBM_LICENSING_IMAGE=${REGISTRY}/${IBM_LICENSING_IMAGE}:${CSV_VERSION}; \
	export IBM_LICENSE_SERVICE_REPORTER_IMAGE=${REGISTRY}/${IBM_LICENSE_SERVICE_REPORTER_IMAGE}:${CSV_VERSION}; \
	export IBM_LICENSE_SERVICE_REPORTER_UI_IMAGE=${REGISTRY}/${IBM_LICENSE_SERVICE_REPORTER_UI_IMAGE}:${CSV_VERSION}; \
	export IBM_POSTGRESQL_IMAGE=${REGISTRY}/${IBM_POSTGRESQL_IMAGE}:${CSV_VERSION}; \
	export IBM_LICENSING_USAGE_IMAGE=${REGISTRY}/${IBM_LICENSING_USAGE_IMAGE}:${CSV_VERSION}; \
	go test -v ./controllers/... -coverprofile cover.out

# Build manager binary
manager: generate
	go build -o bin/$(IMAGE_NAME) main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: fmt vet
	export IBM_LICENSING_IMAGE=${REGISTRY}/${IBM_LICENSING_IMAGE}:${CSV_VERSION}; \
	export IBM_LICENSE_SERVICE_REPORTER_IMAGE=${REGISTRY}/${IBM_LICENSE_SERVICE_REPORTER_IMAGE}:${CSV_VERSION}; \
	export IBM_LICENSE_SERVICE_REPORTER_UI_IMAGE=${REGISTRY}/${IBM_LICENSE_SERVICE_REPORTER_UI_IMAGE}:${CSV_VERSION}; \
	export IBM_POSTGRESQL_IMAGE=${REGISTRY}/${IBM_POSTGRESQL_IMAGE}:${CSV_VERSION}; \
	export IBM_LICENSING_USAGE_IMAGE=${REGISTRY}/${IBM_LICENSING_USAGE_IMAGE}:${CSV_VERSION}; \
	WATCH_NAMESPACE= go run ./main.go

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
manifests: controller-gen
	yq -i '.metadata.annotations."olm.skipRange" = ">=1.0.0 <$(CSV_VERSION)"' ./config/manifests/bases/ibm-licensing-operator.clusterserviceversion.yaml
	yq -i '.metadata.annotations.containerImage = "icr.io/cpopen/${IMG}:$(CSV_VERSION)"' ./config/manifests/bases/ibm-licensing-operator.clusterserviceversion.yaml
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=ibm-licensing-operator webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

# Build the docker image
docker-build: test
	docker build . -t ${IMG}

# Push the docker image
docker-push:
	docker push ${IMG}

# Take the roles (e.g. permissions) from bundle manifest that are created by kubebuilder and put them in CSV
update-roles-alm-example: alm-example
	yq -P '.rules' ./bundle/manifests/ibm-license-service_rbac.authorization.k8s.io_v1_clusterrole.yaml > /tmp/clusterrole.yaml
	yq -P '.rules' ./bundle/manifests/ibm-license-service_rbac.authorization.k8s.io_v1_role.yaml > /tmp/role.yaml
	yq -P '.rules' ./bundle/manifests/ibm-licensing-default-reader_rbac.authorization.k8s.io_v1_clusterrole.yaml > /tmp/reader-clusterrole.yaml
	yq -P '.rules' ./bundle/manifests/ibm-license-service-restricted_rbac.authorization.k8s.io_v1_clusterrole.yaml > /tmp/clusterrole2.yaml
	yq -P '.rules' ./bundle/manifests/ibm-license-service-restricted_rbac.authorization.k8s.io_v1_role.yaml > /tmp/role2.yaml

	sed -i -e 's/^/  /' /tmp/clusterrole.yaml
	sed -i -e 's/^/  /' /tmp/role.yaml
	sed -i -e 's/^/  /' /tmp/reader-clusterrole.yaml
	sed -i -e 's/^/  /' /tmp/clusterrole2.yaml
	sed -i -e 's/^/  /' /tmp/role2.yaml

	yq -i '.spec.install.spec.clusterPermissions[1].rules |= load("/tmp/clusterrole.yaml") | \
		.spec.install.spec.clusterPermissions[1].serviceAccountName = "ibm-license-service" \
	' ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml

	yq -i '.spec.install.spec.clusterPermissions[2].rules |= load("/tmp/clusterrole2.yaml") | \
		.spec.install.spec.clusterPermissions[2].serviceAccountName = "ibm-license-service-restricted" \
	' ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml

	yq -i '.spec.install.spec.clusterPermissions[3].rules |= load("/tmp/reader-clusterrole.yaml") | \
		.spec.install.spec.clusterPermissions[3].serviceAccountName = "ibm-licensing-default-reader" \
	' ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml	

	yq -i '.spec.install.spec.permissions[1].rules |= load("/tmp/role.yaml") | \
		.spec.install.spec.permissions[1].serviceAccountName = "ibm-license-service" \
	' ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml

	yq -i '.spec.install.spec.permissions[2].rules |= load("/tmp/role2.yaml") | \
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
alm-example:
	mkdir -p /tmp/json
	yq -P -o=json ./config/samples/operator.ibm.com_v1alpha1_ibmlicensing.yaml > /tmp/json/ibmlicensing.json
	yq -P -o=json ./config/samples/operator.ibm.com_v1alpha1_ibmlicenseservicereporter.yaml > /tmp/json/ibmlicenseservicereporter.json
	yq -P -o=json ./config/samples/operator.ibm.com_v1alpha1_ibmlicensingbindinfo.yaml > /tmp/json/ibmlicensingbindinfo.json
	yq -P -o=json ./config/samples/operator.ibm.com_v1alpha1_ibmlicensingrequest.yaml > /tmp/json/ibmlicensingrequest.json

	jq -s '.' /tmp/json/ibmlicensing.json /tmp/json/ibmlicenseservicereporter.json /tmp/json/ibmlicensingbindinfo.json /tmp/json/ibmlicensingrequest.json > /tmp/json/merged.json
	yq -i '.metadata.annotations.alm-examples |= load_str("/tmp/json/merged.json")' ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml

	rm -r /tmp/json

# Generate bundle manifests and metadata, then validate generated files. Yq is used to change order of owned resources here to ensure Licensing is first and Reporter second.
pre-bundle: manifests
	operator-sdk generate kustomize manifests -q
	$(KUSTOMIZE) build config/manifests | operator-sdk generate bundle -q --overwrite --version $(CSV_VERSION) $(BUNDLE_METADATA_OPTS)
	yq '.spec.customresourcedefinitions.owned[0]' ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml > yq_tmp_reporter.yaml
	yq '.spec.customresourcedefinitions.owned[1]' ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml > yq_tmp_definitions.yaml
	yq '.spec.customresourcedefinitions.owned[2]' ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml > yq_tmp_metadata.yaml
	yq '.spec.customresourcedefinitions.owned[3]' ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml > yq_tmp_querysources.yaml
	yq '.spec.customresourcedefinitions.owned[4]' ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml > yq_tmp_licensing.yaml
	yq -i eval-all 'select(fileIndex==0).spec.customresourcedefinitions.owned[0] = select(fileIndex==1) | select(fileIndex==0)' ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml yq_tmp_licensing.yaml
	yq -i eval-all 'select(fileIndex==0).spec.customresourcedefinitions.owned[1] = select(fileIndex==1) | select(fileIndex==0)' ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml yq_tmp_reporter.yaml
	yq -i eval-all 'select(fileIndex==0).spec.customresourcedefinitions.owned[2] = select(fileIndex==1) | select(fileIndex==0)' ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml yq_tmp_definitions.yaml
	yq -i eval-all 'select(fileIndex==0).spec.customresourcedefinitions.owned[3] = select(fileIndex==1) | select(fileIndex==0)' ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml yq_tmp_metadata.yaml
	yq -i eval-all 'select(fileIndex==0).spec.customresourcedefinitions.owned[4] = select(fileIndex==1) | select(fileIndex==0)' ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml yq_tmp_querysources.yaml
	yq -i '.spec.relatedImages = load("./common/relatedImages.yaml")' ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml

	rm yq_tmp_reporter.yaml yq_tmp_licensing.yaml yq_tmp_metadata.yaml yq_tmp_definitions.yaml yq_tmp_querysources.yaml
	operator-sdk bundle validate ./bundle

bundle: pre-bundle update-roles-alm-example

# Build the bundle image.
bundle-build:
	docker build -f bundle.Dockerfile -t ${REGISTRY}/${BUNDLE_IMG} .

# Build the bundle image.
bundle-build-development:
	docker build -f bundle.Dockerfile -t ${SCRATCH_REGISTRY}/${BUNDLE_IMG} .

scorecard:
	operator-sdk scorecard ./bundle -n ${NAMESPACE} -w 120s

catalogsource: opm
	@echo "Build CatalogSource for $(LOCAL_ARCH)...- ${BUNDLE_IMG} - ${CATALOG_IMG}"
	curl -Lo ./yq "https://github.com/mikefarah/yq/releases/download/${YQ_VERSION}/yq_$(TARGET_OS)_$(LOCAL_ARCH)"
	chmod +x ./yq
	./yq -i '.spec.install.spec.deployments[0].spec.template.spec.containers[0].image = "${REGISTRY}/${IMG}:${CSV_VERSION}"' ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml
	./yq -i '.spec.install.spec.deployments[0].spec.template.spec.containers[0].env[0].value = "${REGISTRY}/${IBM_LICENSING_IMAGE}:${CSV_VERSION}"' ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml
	./yq -i '.spec.install.spec.deployments[0].spec.template.spec.containers[0].env[1].value = "${REGISTRY}/${IBM_LICENSE_SERVICE_REPORTER_UI_IMAGE}:${CSV_VERSION}"' ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml
	./yq -i '.spec.install.spec.deployments[0].spec.template.spec.containers[0].env[2].value = "${REGISTRY}/${IBM_POSTGRESQL_IMAGE}:${CSV_VERSION}"' ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml
	./yq -i '.spec.install.spec.deployments[0].spec.template.spec.containers[0].env[3].value = "${REGISTRY}/${IBM_LICENSE_SERVICE_REPORTER_IMAGE}:${CSV_VERSION}"' ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml
	./yq -i '.spec.install.spec.deployments[0].spec.template.spec.containers[0].env[4].value = "${REGISTRY}/${IBM_LICENSING_USAGE_IMAGE}:${CSV_VERSION}"' ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml
	./yq -i '.annotations."operators.operatorframework.io.bundle.channels.v1" =  "${CHANNELS}"' ./bundle/metadata/annotations.yaml
	./yq -i '.annotations."operators.operatorframework.io.bundle.channel.default.v1" =  "${DEFAULT_CHANNEL}"' ./bundle/metadata/annotations.yaml	
	docker build -f bundle.Dockerfile -t ${REGISTRY}/${BUNDLE_IMG} .
	docker push ${REGISTRY}/${BUNDLE_IMG}
	$(OPM) index add --permissive -c ${PODMAN} --bundles ${REGISTRY}/${BUNDLE_IMG} --tag ${REGISTRY}/${CATALOG_IMG}
	docker push ${REGISTRY}/${CATALOG_IMG}
ifneq (${DEVOPS_STREAM},)
	docker tag ${REGISTRY}/${CATALOG_IMG} ${REGISTRY}/${DEVOPS_CATALOG_IMG}
	docker push ${REGISTRY}/${DEVOPS_CATALOG_IMG}
endif

catalogsource-development: opm
	@echo "Build Development CatalogSource for $(LOCAL_ARCH)...- ${BUNDLE_IMG} - ${CATALOG_IMG}"
	curl -Lo ./yq "https://github.com/mikefarah/yq/releases/download/${YQ_VERSION}/yq_$(TARGET_OS)_$(LOCAL_ARCH)"
	chmod +x ./yq
	./yq -i '.spec.install.spec.deployments[0].spec.template.spec.containers[0].image = "${SCRATCH_REGISTRY}/${IMG}:${GIT_BRANCH}"' ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml
	./yq -i '.spec.install.spec.deployments[0].spec.template.spec.containers[0].env[0].value = "${REGISTRY}/${IBM_LICENSING_IMAGE}:${CSV_VERSION}"' ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml
	./yq -i '.spec.install.spec.deployments[0].spec.template.spec.containers[0].env[1].value = "${REGISTRY}/${IBM_LICENSE_SERVICE_REPORTER_UI_IMAGE}:${CSV_VERSION}"' ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml
	./yq -i '.spec.install.spec.deployments[0].spec.template.spec.containers[0].env[2].value = "${REGISTRY}/${IBM_POSTGRESQL_IMAGE}:${CSV_VERSION}"' ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml
	./yq -i '.spec.install.spec.deployments[0].spec.template.spec.containers[0].env[3].value = "${REGISTRY}/${IBM_LICENSE_SERVICE_REPORTER_IMAGE}:${CSV_VERSION}"' ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml
	./yq -i '.spec.install.spec.deployments[0].spec.template.spec.containers[0].env[4].value = "${REGISTRY}/${IBM_LICENSING_USAGE_IMAGE}:${CSV_VERSION}"' ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml
	./yq -i '.annotations."operators.operatorframework.io.bundle.channels.v1" =  "${CHANNELS}"' ./bundle/metadata/annotations.yaml
	./yq -i '.annotations."operators.operatorframework.io.bundle.channel.default.v1" =  "${DEFAULT_CHANNEL}"' ./bundle/metadata/annotations.yaml	
	docker build -f bundle.Dockerfile -t ${SCRATCH_REGISTRY}/${BUNDLE_IMG} .
	docker push ${SCRATCH_REGISTRY}/${BUNDLE_IMG}
	$(OPM) index add --permissive  -c ${PODMAN}  --bundles ${SCRATCH_REGISTRY}/${BUNDLE_IMG} --tag ${SCRATCH_REGISTRY}/${CATALOG_IMG}
	docker push  ${SCRATCH_REGISTRY}/${CATALOG_IMG}

############################################################
# Installation section
############################################################

##@ Install

install-linters:  ## Install/verify required linting tools
	common/scripts/install-linters-development.sh

verify-installed-tools: ## Verify if tools are installed
	@command -v operator-sdk >/dev/null 2>&1 || { echo >&2 "Required tool: operator-sdk-${OPERATOR_SDK_VERSION} is not installed.  Run 'make install-all-tools' to install it."; exit 1; }
	@command -v opm >/dev/null 2>&1 || { echo >&2 "Required tool: opm-${OPM_VERSION} is not installed.  Run 'make install-all-tools' to install it."; exit 1; }
	@command -v controller-gen >/dev/null 2>&1 || { echo >&2 "Required tool: controller-gen-${CONTROLLER_GEN_VERSION} is not installed.  Run 'make install-all-tools' to install it."; exit 1; }
	@command -v kustomize >/dev/null 2>&1 || { echo >&2 "Required tool: kustomize-${KUSTOMIZE_VERSION} is not installed.  Run 'make install-all-tools' to install it."; exit 1; }
	@command -v yq >/dev/null 2>&1 || { echo >&2 "Required tool: yq-${YQ_VERSION} is not installed.  Run 'make install-all-tools' to install it."; exit 1; }
	@echo "Successfully verified installed tools. Make sure the version matches required to avoid further issues.$'\n"

	@echo "Printing installed tools summary $'\n\
	Required | Installed $'\n\
	» operator-sdk-${OPERATOR_SDK_VERSION} | operator-sdk-"$(shell operator-sdk version | awk '{print $$3}')" $'\n\
	» opm-${OPM_VERSION} | opm-"$(shell opm version | awk '{print $$2}' | awk -F ':' '{print $$2}')" $'\n\
	» controller-gen-${CONTROLLER_GEN_VERSION} | controller-gen-"$(shell controller-gen --version | awk '{print $$2}')", $'\n\
	» kustomize-${KUSTOMIZE_VERSION} | kustomize-"$(shell kustomize version | awk '{print $$1}' | awk -F ':' '{print $$2}')", $'\n\
	» yq-${YQ_VERSION} | yq-"$(shell yq --version | awk '{print $$4}')" $'\n\
	"

install-all-tools: install-operator-sdk install-opm install-controller-gen install-kustomize install-yq verify-installed-tools ## Install all tools locally

install-operator-sdk: ## Install tool locally: operator-sdk
	@operator-sdk version 2> /dev/null ; if [ $$? -ne 0 ]; then bash common/scripts/install-operator-sdk.sh ${TARGET_OS} ${LOCAL_ARCH} ${OPERATOR_SDK_VERSION}; fi

install-opm: ## Install tool locally: opm
	@opm version 2> /dev/null ; if [ $$? -ne 0 ]; then bash common/scripts/install-opm.sh ${TARGET_OS} ${LOCAL_ARCH} ${OPM_VERSION}; fi	

install-controller-gen: ## Install tool locally: controller-gen
	@controller-gen --version 2> /dev/null ; if [ $$? -ne 0 ]; then go install sigs.k8s.io/controller-tools/cmd/controller-gen@${CONTROLLER_GEN_VERSION}; fi	

install-kustomize: ## Install tool locally: kustomize
	@kustomize version 2> /dev/null ; if [ $$? -ne 0 ]; then go install sigs.k8s.io/kustomize/kustomize/v4@${KUSTOMIZE_VERSION}; fi	

install-yq: ## Install tool locally: yq
	@yq --version 2> /dev/null ; if [ $$? -ne 0 ]; then bash common/scripts/install-yq.sh ${TARGET_OS} ${LOCAL_ARCH} ${YQ_VERSION}; fi	

controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@${CONTROLLER_GEN_VERSION} ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

kustomize:
ifeq (, $(shell which kustomize))
	@{ \
	set -e ;\
	KUSTOMIZE_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$KUSTOMIZE_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/kustomize/kustomize/v4@${KUSTOMIZE_VERSION} ;\
	rm -rf $$KUSTOMIZE_GEN_TMP_DIR ;\
	}
KUSTOMIZE=$(GOBIN)/kustomize
else
KUSTOMIZE=$(shell which kustomize)
endif

opm:
ifeq (, $(shell which opm))
	@{ \
	set -e ;\
	OPM_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$OPM_GEN_TMP_DIR ;\
	git clone  --branch ${OPM_VERSION}  https://github.com/operator-framework/operator-registry.git ;\
	cd ./operator-registry ; \
	git checkout ${OPM_VERSION};\
	GOARCH=$(LOCAL_ARCH) GOFLAGS="-mod=vendor" go build -ldflags "-X 'github.com/operator-framework/operator-registry/cmd/opm/version.opmVersion=${OPM_VERSION}'"  -tags "json1" -o bin/opm ./cmd/opm ;\
	cp ./bin/opm ~/ ; \
	rm -rf $$OPM_GEN_TMP_DIR ;\
	}
OPM=~/opm
else
OPM=$(shell which opm)
endif

ifeq (, $(shell which podman))
PODMAN=docker
else
PODMAN=podman
endif

.PHONY: all opm build bundle-build bundle pre-bundle kustomize catalogsource controller-gen generate docker-build docker-push deploy manifests run install uninstall code-dev check lint test coverage-kind coverage build multiarch-image csv clean help
