#!/usr/bin/env bash
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
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
BUILD_TEMP="${SCRIPT_DIR}/../temp"
mkdir -p "${BUILD_TEMP}"

export PATH="/usr/local/bin:/usr/local/go/bin:$PATH"

# Install Go
echo "Installing Go ${GO_VERSION} ..."
curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -o "${BUILD_TEMP}/go.tgz"
sudo tar -C /usr/local -xzf "${BUILD_TEMP}/go.tgz"

echo "Go installed: $(go version)"


# Install golangci-lint
echo "Installing golangci-lint ${GOLANGCI_LINT_VERSION} ..."
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sudo sh -s -- -b /usr/local/bin "v${GOLANGCI_LINT_VERSION}"

echo "golangci-lint installed: $(golangci-lint version)"


# Install operator-sdk
echo "Installing operator-sdk ${OPERATOR_SDK_VERSION} ..."
curl -fsSL "https://github.com/operator-framework/operator-sdk/releases/download/${OPERATOR_SDK_VERSION}/operator-sdk_linux_amd64" -o "${BUILD_TEMP}/operator-sdk"
sudo install -m 0755 "${BUILD_TEMP}/operator-sdk" /usr/local/bin/operator-sdk

echo "operator-sdk installed: $(operator-sdk version)"


echo "Installation complete"
