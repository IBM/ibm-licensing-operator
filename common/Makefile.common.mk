# Copyright 2019 The Kubernetes Authors.
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

############################################################
# GKE section
############################################################
PROJECT ?= oceanic-guard-191815
ZONE    ?= us-west1-a
CLUSTER ?= prow

activate-serviceaccount:
ifdef GOOGLE_APPLICATION_CREDENTIALS
	@gcloud auth activate-service-account --key-file="$(GOOGLE_APPLICATION_CREDENTIALS)"
endif

get-cluster-credentials: activate-serviceaccount
	@gcloud container clusters get-credentials "$(CLUSTER)" --project="$(PROJECT)" --zone="$(ZONE)"

config-docker: get-cluster-credentials
	@common/scripts/config_docker.sh

config-docker-scratch: get-cluster-credentials
	@common/scripts/config_docker_scratch.sh

############################################################
# install git hooks
############################################################
INSTALL_HOOKS := $(shell find .git/hooks -type l -exec rm {} \; && \
                         find common/scripts/.githooks -type f -exec ln -sf ../../{} .git/hooks/ \; )

############################################################
# lint section
############################################################

FINDFILES=find . \( -path ./.git -o -path ./.github -o -path ./common/scripts/catalog -o -path ./common/scripts/tests -o -path ./common/scripts/catalog_build.sh -o -path ./.go \) -prune -o -type f
XARGS = xargs -0 ${XARGS_FLAGS}
CLEANXARGS = xargs ${XARGS_FLAGS}

lint-dockerfiles:
	@${FINDFILES} -name 'Dockerfile*' -print0 | ${XARGS} hadolint -c ./common/config/.hadolint.yml

lint-scripts:
	@${FINDFILES} -name '*.sh' -print0 | ${XARGS} shellcheck

lint-yaml:
	@${FINDFILES} \( -name '*.yml' -o -name '*.yaml' \) -print0 | ${XARGS} grep -L -e "{{" | ${CLEANXARGS} yamllint -c ./common/config/.yamllint.yml

lint-copyright-banner:
	@${FINDFILES} \( -name '*.go' -o -name '*.cc' -o -name '*.h' -o -name '*.proto' -o -name '*.py' -o -name '*.sh' \) \( ! \( -name '*.gen.go' -o -name '*.pb.go' -o -name '*_pb2.py' \) \) -print0 |\
		${XARGS} common/scripts/lint_copyright_banner.sh

lint-go:
	@${FINDFILES} -name '*.go' \( ! \( -name '*.gen.go' -o -name '*.pb.go' \) \) -print0 | ${XARGS} common/scripts/lint_go.sh

lint-markdown:
	@${FINDFILES} -name '*.md' -print0 | ${XARGS} mdl --ignore-front-matter --style common/config/mdl.rb
ifdef MARKDOWN_LINT_WHITELIST
	@${FINDFILES} -name '*.md' -print0 | ${XARGS} awesome_bot --skip-save-results --allow_ssl --allow-timeout --allow-dupe --allow-redirect --white-list ${MARKDOWN_LINT_WHITELIST}
else
	@${FINDFILES} -name '*.md' -print0 | ${XARGS} awesome_bot --skip-save-results --allow_ssl --allow-timeout --allow-dupe --allow-redirect
endif
lint-all: lint-dockerfiles lint-scripts lint-yaml lint-copyright-banner lint-go lint-markdown

format-go:
	@${FINDFILES} -name '*.go' \( ! \( -name '*.gen.go' -o -name '*.pb.go' \) \) -print0 | ${XARGS} goimports -w -local "github.com/IBM"

.PHONY: lint-dockerfiles lint-scripts lint-yaml lint-copyright-banner lint-go lint-markdown lint-all format-go

# Run go vet for this project. More info: https://golang.org/cmd/vet/
code-vet:
	@echo go vet
	go vet $$(go list ./... )

# Run go fmt for this project
code-fmt:
	@echo go fmt
	go fmt $$(go list ./... )

# Run go mod tidy to update dependencies
code-tidy:
	@echo go mod tidy
	go mod tidy -v

define CRD_LABELS_BODY
  labels:\n    app.kubernetes.io\/instance: \"ibm-licensing-operator\"\n    app.kubernetes.io\/managed-by: \"ibm-licensing-operator\"\n    app.kubernetes.io\/name: \"ibm-licensing\"
endef

export CRD_LABELS_BODY
# Run the operator-sdk commands to generated code (k8s and openapi and csv)
code-gen:
	@echo Updating the deep copy files with the changes in the API
	operator-sdk generate k8s
	@echo Updating the CRD files with the OpenAPI validations
	operator-sdk generate crds
	@echo Adding labels for CRD
	sed -i 's/  name: ibmlicensings.operator.ibm.com/  name: ibmlicensings.operator.ibm.com\n'"$$CRD_LABELS_BODY"'/g' deploy/crds/operator.ibm.com_ibmlicensings_crd.yaml
	@echo Generate openapi
	openapi-gen --logtostderr=true -o "" -i ./pkg/apis/operator/v1alpha1 -O zz_generated.openapi -p ./pkg/apis/operator/v1alpha1 -h hack/boilerplate.go.txt -r "-"
#	Not generating csv as it may break existing one, csv needs human changes
#	@echo Updating/Generating a ClusterServiceVersion YAML manifest for the operator
#	operator-sdk generate csv --csv-version ${CSV_VERSION} --update-crds

csv-gen:
	@echo Remember to fix things after csv generation
	operator-sdk generate csv --csv-version ${CSV_VERSION} --update-crds

.PHONY: code-vet code-fmt code-tidy code-gen csv-gen
