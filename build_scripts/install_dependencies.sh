#!/usr/bin/env bash
set -euo pipefail

export PATH="/usr/local/bin:/usr/local/go/bin:$PATH"

# Install Go
echo "Installing Go ${GO_VERSION} ..."
curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -o /tmp/go.tgz
sudo tar -C /usr/local -xzf /tmp/go.tgz

echo "Go installed: $(go version)"


# Install golangci-lint
echo "Installing golangci-lint ${GOLANGCI_LINT_VERSION} ..."
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sudo sh -s -- -b /usr/local/bin "v${GOLANGCI_LINT_VERSION}"

echo "golangci-lint installed: $(golangci-lint version)"


# Install operator-sdk
echo "Installing operator-sdk ${OPERATOR_SDK_VERSION} ..."
curl -fsSL "https://github.com/operator-framework/operator-sdk/releases/download/${OPERATOR_SDK_VERSION}/operator-sdk_linux_amd64" -o /tmp/operator-sdk
sudo install -m 0755 /tmp/operator-sdk /usr/local/bin/operator-sdk

echo "operator-sdk installed: $(operator-sdk version)"


# Install kind
echo "Installing kind ${KIND_VERSION} ..."
curl -fsSL "https://kind.sigs.k8s.io/dl/${KIND_VERSION}/kind-$(uname)-amd64" -o /tmp/kind
sudo install -m 0755 /tmp/kind /usr/local/bin/kind

echo "kind installed: $(kind --version)"


echo "Installation complete"
