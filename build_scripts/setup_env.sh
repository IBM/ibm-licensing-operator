#!/usr/bin/env bash
set -euo pipefail

# Tool versions:
export GO_VERSION="1.24.10"
export GOLANGCI_LINT_VERSION="1.64.2"
export OPERATOR_SDK_VERSION="v1.25.2"
export KIND_VERSION="v0.17.0"

# Build flags:
export GOFLAGS="-buildvcs=false"

# Test veriables:
export LICENSING_NAMESPACE="ibm-license-service"

# Re-exporting build pipeline variables:
export ARTIFACTORY_USERNAME="$(get_env ARTIFACTORY_USERNAME)"
export ARTIFACTORY_TOKEN="$(get_env ARTIFACTORY_TOKEN)"
export GIT_BRANCH="$(get_env git-branch)"
export GIT_COMMIT="$(get_env git-commit)"
export DOCKER_REGISTRY="$(get_env DOCKER_REGISTRY)"
