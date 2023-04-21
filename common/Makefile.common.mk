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
ZONE    ?= us-east5-c
CLUSTER ?= bedrock-prow

activate-serviceaccount:
ifdef GOOGLE_APPLICATION_CREDENTIALS
	@gcloud auth activate-service-account --key-file="$(GOOGLE_APPLICATION_CREDENTIALS)" || true
endif

get-cluster-credentials: activate-serviceaccount
	mkdir -p ~/.kube; cp -v /etc/kubeconfig/config ~/.kube; kubectl config use-context default; kubectl get nodes; echo going forward retiring google cloud
ifdef GOOGLE_APPLICATION_CREDENTIALS
       gcloud container clusters get-credentials "$(CLUSTER)" --project="$(PROJECT)" --zone="$(ZONE)" || true
endif

config-docker: get-cluster-credentials
	@common/scripts/config_docker.sh

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
	@${FINDFILES} \( -name '*.yml' -o -name '*.yaml' \) \( ! \( -name 'kustomization.yaml' \) \) -print0 | ${XARGS} grep -L -e "{{" | ${CLEANXARGS} yamllint -c ./common/config/.yamllint.yml

lint-copyright-banner:
	@${FINDFILES} \( -name '*.go' -o -name '*.cc' -o -name '*.h' -o -name '*.proto' -o -name '*.py' -o -name '*.sh' \) \( ! \( -name '*.gen.go' -o -name '*.pb.go' -o -name '*_pb2.py' -o -name '*_generated.deepcopy.go' \) \) -print0 |\
		${XARGS} common/scripts/lint_copyright_banner.sh

lint-go:
	@golangci-lint --version
	@${FINDFILES} -name '*.go' \( ! \( -name '*.gen.go' -o -name '*.pb.go' -o -name '*_generated.deepcopy.go' \) \) -print0 | ${XARGS} common/scripts/lint_go.sh

lint-markdown:
	@${FINDFILES} -name '*.md' -print0 | ${XARGS} mdl --ignore-front-matter --style common/config/mdl.rb
ifdef MARKDOWN_LINT_WHITELIST
	@${FINDFILES} -name '*.md' -print0 | ${XARGS} awesome_bot --skip-save-results --allow_ssl --allow-timeout --allow-dupe --allow-redirect --white-list ${MARKDOWN_LINT_WHITELIST}
else
	@${FINDFILES} -name '*.md' -print0 | ${XARGS} awesome_bot --skip-save-results --allow_ssl --allow-timeout --allow-dupe --allow-redirect
endif

lint-all: lint-dockerfiles lint-scripts lint-yaml lint-copyright-banner lint-go lint-markdown

format-go:
	@${FINDFILES} -name '*.go' \( ! \( -name '*.gen.go' -o -name '*.pb.go' -o -name '*_generated.deepcopy.go' \) \) -print0 | ${XARGS} goimports -w -local "github.com/IBM"

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

.PHONY: code-vet code-fmt code-tidy
