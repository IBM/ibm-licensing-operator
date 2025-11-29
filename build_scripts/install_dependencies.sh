# Go installation:
curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -o /tmp/go.tgz
sudo tar -C /usr/local -xzf /tmp/go.tgz
export PATH="$PATH:/usr/local/go/bin"
echo "Go installed: $(go version)"

# golangci-lint installation:
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b /usr/local/bin v${{ GOLANGCI_LINT_VERSION }}
echo "golangci-lint installed: $(golangci-lint version)"
