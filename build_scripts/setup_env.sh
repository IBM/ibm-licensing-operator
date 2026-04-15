#!/usr/bin/env bash
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
set -euo pipefail


# Tool versions:
export GO_VERSION="1.26.2"
export GOLANGCI_LINT_VERSION="2.11.2"
export OPERATOR_SDK_VERSION="v1.42.1"
export KIND_VERSION="v0.31.0"

# Build flags:
export GOFLAGS="-buildvcs=false"

# Test variables:
export LICENSING_NAMESPACE="ibm-licensing"

# Re-exporting build pipeline variables:
ARTIFACTORY_USERNAME="$(get_env ARTIFACTORY_USERNAME)"
export ARTIFACTORY_USERNAME
ARTIFACTORY_TOKEN="$(get_env ARTIFACTORY_TOKEN)"
export ARTIFACTORY_TOKEN
GIT_BRANCH="$(get_env git-branch)"
export GIT_BRANCH
GIT_COMMIT="$(get_env git-commit)"
export GIT_COMMIT
DOCKER_REGISTRY="$(get_env DOCKER_REGISTRY)"
export DOCKER_REGISTRY
