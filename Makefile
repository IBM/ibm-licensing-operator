#
# Copyright 2022 IBM Corporation
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
CSV_VERSION ?= 1.16.1
CSV_VERSION_DEVELOPMENT ?= development
OLD_CSV_VERSION ?= 1.16.0

# This repo is build locally for dev/test by default;
# Override this variable in CI env.
BUILD_LOCALLY ?= 1

# Image URL to use all building/pushing image targets;
# Use your own docker registry and image name for dev/test by overriding the IMG, REGISTRY and CSV_VERSION environment variable.
IMG ?= ibm-licensing-operator
REGISTRY ?= "hyc-cloud-private-integration-docker-local.artifactory.swg-devops.com/ibmcom"

SCRATCH_REGISTRY ?= "hyc-cloud-private-scratch-docker-local.artifactory.swg-devops.com/ibmcom"

# Default bundle image tag
IMAGE_BUNDLE_NAME ?= ibm-licensing-operator-bundle
IMAGE_CATALOG_NAME ?= ibm-licensing-operator-catalog

IBM_LICENSING_IMAGE ?= ibm-licensing
IBM_LICENSE_SERVICE_REPORTER_UI_IMAGE ?= ibm-license-service-reporter-ui
IBM_POSTGRESQL_IMAGE ?= ibm-postgresql
IBM_LICENSE_SERVICE_REPORTER_IMAGE ?= ibm-license-service-reporter
IBM_LICENSING_USAGE_IMAGE ?= ibm-licensing-usage

CHANNELS="v3,beta"
DEFAULT_CHANNEL=v3

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
$(eval BUILD_DATE := $(shell date +%Y/%m/%d@%H:%M:%S))
$(eval GIT_COMMIT := $(shell git rev-parse --short HEAD))
$(eval VCS_REF := $(GIT_COMMIT))
IMAGE_RELEASE=$(VCS_REF)
IMAGE_BUILDDATE=$(BUILD_DATE)
GIT_REMOTE_URL = $(shell git config --get remote.origin.url)

BUNDLE_IMG ?= $(IMAGE_BUNDLE_NAME)-$(LOCAL_ARCH):$(VERSION)
CATALOG_IMG ?= $(IMAGE_CATALOG_NAME)-$(LOCAL_ARCH):$(VERSION)

$(eval DOCKER_BUILD_OPTS := --build-arg "IMAGE_NAME=$(IMAGE_NAME)" --build-arg "IMAGE_DISPLAY_NAME=$(IMAGE_DISPLAY_NAME)" --build-arg "IMAGE_MAINTAINER=$(IMAGE_MAINTAINER)" --build-arg "IMAGE_VENDOR=$(IMAGE_VENDOR)" --build-arg "IMAGE_VERSION=$(IMAGE_VERSION)" --build-arg "VERSION=$(CSV_VERSION)" --build-arg "IMAGE_RELEASE=$(IMAGE_RELEASE)"  --build-arg "IMAGE_BUILDDATE=$(IMAGE_BUILDDATE)" --build-arg "IMAGE_DESCRIPTION=$(IMAGE_DESCRIPTION)" --build-arg "IMAGE_SUMMARY=$(IMAGE_SUMMARY)" --build-arg "IMAGE_OPENSHIFT_TAGS=$(IMAGE_OPENSHIFT_TAGS)" --build-arg "VCS_REF=$(VCS_REF)" --build-arg "VCS_URL=$(GIT_REMOTE_URL)" --build-arg "IMAGE_NAME_ARCH=$(IMAGE_NAME)-$(LOCAL_ARCH)")

all: fmt version.properties check test coverage-kind build images

ifeq ($(BUILD_LOCALLY),0)
    ifneq ("$(realpath $(DEST))", "$(realpath $(PWD))")
        $(error Please run 'make' from $(DEST). Current directory is $(PWD))
    endif
endif

include common/Makefile.common.mk

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

code-dev: ## Run the default dev commands which are the go tidy, fmt, vet then execute the $ make code-gen
	@echo Running the common required commands for developments purposes
	- make code-tidy
	- make code-fmt
	- make code-vet
	- make code-gen
	@echo Running the common required commands for code delivery
	make check

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

build-push-image-development: build-image-development push-image-development catalogsource-development

build-image-development: $(CONFIG_DOCKER_TARGET) build
	@echo $(DOCKER_BUILD_OPTS)
	@echo "Building the $(IMAGE_NAME) docker image for $(LOCAL_ARCH)..."
	@docker build -t $(SCRATCH_REGISTRY)/$(IMAGE_NAME)-$(LOCAL_ARCH):$(VERSION) $(DOCKER_BUILD_OPTS) -f Dockerfile .

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
	@MAX_PULLING_RETRY=20 RETRY_INTERVAL=30 common/scripts/multiarch_image.sh $(REGISTRY) $(IMAGE_CATALOG_NAME) $(VERSION) ${MANIFEST_VERSION}

multiarch-image-latest: $(CONFIG_DOCKER_TARGET)
	@MAX_PULLING_RETRY=20 RETRY_INTERVAL=30 common/scripts/multiarch_image_latest.sh $(REGISTRY) $(IMAGE_NAME) $(VERSION)
	@MAX_PULLING_RETRY=20 RETRY_INTERVAL=30 common/scripts/multiarch_image_latest.sh $(REGISTRY) $(IMAGE_CATALOG_NAME) $(VERSION)

multiarch-image-development: $(CONFIG_DOCKER_TARGET_SCRATCH)
	@MAX_PULLING_RETRY=20 RETRY_INTERVAL=30 common/scripts/multiarch_image.sh $(SCRATCH_REGISTRY) $(IMAGE_NAME) $(VERSION) ${MANIFEST_VERSION}
	common/scripts/catalog_build.sh $(SCRATCH_REGISTRY) $(IMAGE_NAME) ${MANIFEST_VERSION}
	@MAX_PULLING_RETRY=20 RETRY_INTERVAL=30 common/scripts/multiarch_image.sh $(SCRATCH_REGISTRY) $(IMAGE_CATALOG_NAME) $(VERSION) ${MANIFEST_VERSION}

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

## FROM NEW OPERATOR

# Run tests
#ENVTEST_ASSETS_DIR=$(shell pwd)/testbin
test:
	@echo "Running tests for the controllers."
	#@mkdir -p ${ENVTEST_ASSETS_DIR}
	#@test -f ${ENVTEST_ASSETS_DIR}/setup-envtest.sh || curl -sSLo ${ENVTEST_ASSETS_DIR}/setup-envtest.sh https://raw.githubusercontent.com/kubernetes-sigs/controller-runtime/master/hack/setup-envtest.sh
	#@source ${ENVTEST_ASSETS_DIR}/setup-envtest.sh; fetch_envtest_tools $(ENVTEST_ASSETS_DIR); setup_envtest_env $(ENVTEST_ASSETS_DIR); go test ./... -coverprofile cover.out

prepare-unit-test:
	kubectl create namespace ${NAMESPACE} || echo ""
	kubectl create secret generic artifactory-token -n ${NAMESPACE} --from-file=.dockerconfigjson=./artifactory.yaml --type=kubernetes.io/dockerconfigjson || echo ""
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

unit-test-development: prepare-unit-test
	export USE_EXISTING_CLUSTER=true; \
	export WATCH_NAMESPACE=${NAMESPACE}; \
	export NAMESPACE=${NAMESPACE}; \
	export OCP=${OCP}; \
	export KUBEBUILDER_ATTACH_CONTROL_PLANE_OUTPUT=true; \
	export IBM_LICENSING_IMAGE=${SCRATCH_REGISTRY}/${IBM_LICENSING_IMAGE}:${CSV_VERSION_DEVELOPMENT}; \
	export IBM_LICENSE_SERVICE_REPORTER_IMAGE=${SCRATCH_REGISTRY}/${IBM_LICENSE_SERVICE_REPORTER_IMAGE}:${CSV_VERSION_DEVELOPMENT}; \
	export IBM_LICENSE_SERVICE_REPORTER_UI_IMAGE=${SCRATCH_REGISTRY}/${IBM_LICENSE_SERVICE_REPORTER_UI_IMAGE}:${CSV_VERSION_DEVELOPMENT}; \
	export IBM_POSTGRESQL_IMAGE=${REGISTRY}/${IBM_POSTGRESQL_IMAGE}:${CSV_VERSION}; \
	export IBM_LICENSING_USAGE_IMAGE=${SCRATCH_REGISTRY}/${IBM_LICENSING_USAGE_IMAGE}:${CSV_VERSION_DEVELOPMENT}; \
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
	yq w -i ./config/manifests/bases/ibm-licensing-operator.clusterserviceversion.yaml 'metadata.annotations."olm.skipRange"' '>=1.0.0 <$(CSV_VERSION)'
	yq w -i ./config/manifests/bases/ibm-licensing-operator.clusterserviceversion.yaml 'metadata.annotations.containerImage' 'icr.io/cpopen/${IMG}:$(CSV_VERSION)'
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

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.7.0 ;\
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
	go get sigs.k8s.io/kustomize/kustomize/v3@v3.5.4 ;\
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
	git clone  --branch v1.16.1  https://github.com/operator-framework/operator-registry.git ;\
	cd ./operator-registry ; \
	git checkout v1.16.1;\
	GOARCH=$(LOCAL_ARCH) GOFLAGS="-mod=vendor" go build -ldflags "-X 'github.com/operator-framework/operator-registry/cmd/opm/version.gitCommit=eb9fff53' -X 'github.com/operator-framework/operator-registry/cmd/opm/version.opmVersion=v1.16.1' -X 'github.com/operator-framework/operator-registry/cmd/opm/version.buildDate=2021-03-30T13:32:56Z'"  -tags "json1" -o bin/opm ./cmd/opm ;\
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


alm-example:
	yq w -i ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml "metadata.annotations.alm-examples" \
	"[\
	`yq r -P  -j ./config/samples/operator.ibm.com_v1alpha1_ibmlicensing.yaml`,\
	`yq r -P  -j ./config/samples/operator.ibm.com_v1alpha1_ibmlicenseservicereporter.yaml`,\
	`yq r -P  -j ./config/samples/operator.ibm.com_v1alpha1_ibmlicensingbindinfo.yaml`,\
	`yq r -P  -j ./config/samples/operator.ibm.com_v1alpha1_ibmlicensingrequest.yaml`\
	]"
	yq r -P ./bundle/manifests/ibm-license-service_rbac.authorization.k8s.io_v1_clusterrole.yaml rules > /tmp/clusterrole.yaml
	yq r -P ./bundle/manifests/ibm-license-service_rbac.authorization.k8s.io_v1_role.yaml rules > /tmp/role.yaml
	yq r -P bundle/manifests/ibm-licensing-default-reader_rbac.authorization.k8s.io_v1_clusterrole.yaml rules > /tmp/reader-clusterrole.yaml
	yq r -P ./bundle/manifests/ibm-license-service-restricted_rbac.authorization.k8s.io_v1_clusterrole.yaml rules > /tmp/clusterrole2.yaml
	yq r -P ./bundle/manifests/ibm-license-service-restricted_rbac.authorization.k8s.io_v1_role.yaml rules > /tmp/role2.yaml

	sed -i -e 's/^/  /' /tmp/clusterrole.yaml
	sed -i -e 's/^/  /' /tmp/role.yaml
	sed -i -e 's/^/  /' /tmp/reader-clusterrole.yaml
	sed -i -e 's/^/  /' /tmp/clusterrole2.yaml
	sed -i -e 's/^/  /' /tmp/role2.yaml

	cp ./common/scripts/updateCSV/updateCP.yaml /tmp/updateCP.yaml
	cat /tmp/clusterrole.yaml >> /tmp/updateCP.yaml
	cat ./common/scripts/updateCSV/saCP.yaml >> /tmp/updateCP.yaml
	yq w -i -s /tmp/updateCP.yaml ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml

	cp ./common/scripts/updateCSV/updateCP2.yaml /tmp/updateCP2.yaml
	cat /tmp/clusterrole2.yaml >> /tmp/updateCP2.yaml
	cat ./common/scripts/updateCSV/saCP2.yaml >> /tmp/updateCP2.yaml
	yq w -i -s /tmp/updateCP2.yaml ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml
	cp ./common/scripts/updateCSV/updateP.yaml /tmp/updateP.yaml
	cat /tmp/role.yaml >> /tmp/updateP.yaml
	cat ./common/scripts/updateCSV/saP.yaml >> /tmp/updateP.yaml
	yq w -i -s /tmp/updateP.yaml ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml

	cp ./common/scripts/updateCSV/updateP2.yaml /tmp/updateP2.yaml
	cat /tmp/role2.yaml >> /tmp/updateP2.yaml
	cat ./common/scripts/updateCSV/saP2.yaml >> /tmp/updateP2.yaml
	yq w -i -s /tmp/updateP2.yaml ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml

	cp ./common/scripts/updateCSV/updateAccessCP.yaml /tmp/updateAccessCP.yaml
	cat /tmp/reader-clusterrole.yaml >> /tmp/updateAccessCP.yaml
	cat ./common/scripts/updateCSV/saAccessCP.yaml >> /tmp/updateAccessCP.yaml
	yq w -i -s /tmp/updateAccessCP.yaml ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml

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

# Generate bundle manifests and metadata, then validate generated files. Yq is used to change order of owned resources here to ensure Licensing is first and Reporter second.
pre-bundle: manifests
	operator-sdk generate kustomize manifests -q
	$(KUSTOMIZE) build config/manifests | operator-sdk generate bundle -q --overwrite --version $(CSV_VERSION) $(BUNDLE_METADATA_OPTS)
	yq r ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml "spec.customresourcedefinitions.owned[0]" > yq_tmp_reporter.yaml
	yq r ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml "spec.customresourcedefinitions.owned[1]" > yq_tmp_definitions.yaml
	yq r ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml "spec.customresourcedefinitions.owned[2]" > yq_tmp_metadata.yaml
	yq r ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml "spec.customresourcedefinitions.owned[3]" > yq_tmp_querysources.yaml
	yq r ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml "spec.customresourcedefinitions.owned[4]" > yq_tmp_licensing.yaml
	yq w -i ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml "spec.customresourcedefinitions.owned[0]" -f yq_tmp_licensing.yaml
	yq w -i ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml "spec.customresourcedefinitions.owned[1]" -f yq_tmp_reporter.yaml
	yq w -i ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml "spec.customresourcedefinitions.owned[2]" -f yq_tmp_definitions.yaml
	yq w -i ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml "spec.customresourcedefinitions.owned[3]" -f yq_tmp_metadata.yaml
	yq w -i ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml "spec.customresourcedefinitions.owned[4]" -f yq_tmp_querysources.yaml
	yq w -i ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml 'spec.replaces' ibm-licensing-operator.v${OLD_CSV_VERSION}
	rm yq_tmp_reporter.yaml yq_tmp_licensing.yaml yq_tmp_metadata.yaml yq_tmp_definitions.yaml yq_tmp_querysources.yaml
	operator-sdk bundle validate ./bundle

bundle: pre-bundle alm-example

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
	curl -Lo ./yq "https://github.com/mikefarah/yq/releases/download/3.4.0/yq_linux_$(LOCAL_ARCH)"
	chmod +x ./yq
	./yq d -i ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml 'spec.replaces'
	./yq w -i ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml 'spec.install.spec.deployments[0].spec.template.spec.containers[0].image' "${REGISTRY}/${IMG}:${CSV_VERSION}"
	./yq w -i ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml 'spec.install.spec.deployments[0].spec.template.spec.containers[0].env[0].value'  "${REGISTRY}/${IBM_LICENSING_IMAGE}:${CSV_VERSION}"
	./yq w -i ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml 'spec.install.spec.deployments[0].spec.template.spec.containers[0].env[1].value'  "${REGISTRY}/${IBM_LICENSE_SERVICE_REPORTER_UI_IMAGE}:${CSV_VERSION}"
	./yq w -i ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml 'spec.install.spec.deployments[0].spec.template.spec.containers[0].env[2].value'  "${REGISTRY}/${IBM_POSTGRESQL_IMAGE}:${CSV_VERSION}"
	./yq w -i ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml 'spec.install.spec.deployments[0].spec.template.spec.containers[0].env[3].value'  "${REGISTRY}/${IBM_LICENSE_SERVICE_REPORTER_IMAGE}:${CSV_VERSION}"
	./yq w -i ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml 'spec.install.spec.deployments[0].spec.template.spec.containers[0].env[4].value'  "${REGISTRY}/${IBM_LICENSING_USAGE_IMAGE}:${CSV_VERSION}"
	./yq w -i ./bundle/metadata/annotations.yaml 'annotations."operators.operatorframework.io.bundle.channels.v1"' "v3,beta,dev,stable-v1"
	docker build -f bundle.Dockerfile -t ${REGISTRY}/${BUNDLE_IMG} .
	docker push ${REGISTRY}/${BUNDLE_IMG}
	$(OPM) index add --permissive -c ${PODMAN} --bundles ${REGISTRY}/${BUNDLE_IMG} --tag ${REGISTRY}/${CATALOG_IMG}
	docker push  ${REGISTRY}/${CATALOG_IMG}

catalogsource-development: opm
	@echo "Build Development CatalogSource for $(LOCAL_ARCH)...- ${BUNDLE_IMG} - ${CATALOG_IMG}"
	curl -Lo ./yq "https://github.com/mikefarah/yq/releases/download/3.4.0/yq_linux_$(LOCAL_ARCH)"
	chmod +x ./yq
	./yq d -i ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml 'spec.replaces'
	./yq w -i ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml 'spec.install.spec.deployments[0].spec.template.spec.containers[0].image' "${SCRATCH_REGISTRY}/${IMG}:${CSV_VERSION}"
	./yq w -i ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml 'spec.install.spec.deployments[0].spec.template.spec.containers[0].env[0].value'  "${SCRATCH_REGISTRY}/${IBM_LICENSING_IMAGE}:${CSV_VERSION}"
	./yq w -i ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml 'spec.install.spec.deployments[0].spec.template.spec.containers[0].env[1].value'  "${SCRATCH_REGISTRY}/${IBM_LICENSE_SERVICE_REPORTER_UI_IMAGE}:${CSV_VERSION}"
	./yq w -i ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml 'spec.install.spec.deployments[0].spec.template.spec.containers[0].env[2].value'  "${SCRATCH_REGISTRY}/${IBM_POSTGRESQL_IMAGE}:${CSV_VERSION}"
	./yq w -i ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml 'spec.install.spec.deployments[0].spec.template.spec.containers[0].env[3].value'  "${SCRATCH_REGISTRY}/${IBM_LICENSE_SERVICE_REPORTER_IMAGE}:${CSV_VERSION}"
	./yq w -i ./bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml 'spec.install.spec.deployments[0].spec.template.spec.containers[0].env[4].value'  "${SCRATCH_REGISTRY}/${IBM_LICENSING_USAGE_IMAGE}:${CSV_VERSION}"
	./yq w -i ./bundle/metadata/annotations.yaml 'annotations."operators.operatorframework.io.bundle.channels.v1"' "v3,beta,dev,stable-v1"
	docker build -f bundle.Dockerfile -t ${SCRATCH_REGISTRY}/${BUNDLE_IMG} .
	docker push ${SCRATCH_REGISTRY}/${BUNDLE_IMG}
	$(OPM) index add --permissive  -c ${PODMAN}  --bundles ${SCRATCH_REGISTRY}/${BUNDLE_IMG} --tag ${SCRATCH_REGISTRY}/${CATALOG_IMG}
	docker push  ${SCRATCH_REGISTRY}/${CATALOG_IMG}

.PHONY: all opm build bundle-build bundle pre-bundle kustomize catalogsource controller-gen generate docker-build docker-push deploy manifests run install uninstall code-dev check lint test coverage-kind coverage build multiarch-image csv clean help

