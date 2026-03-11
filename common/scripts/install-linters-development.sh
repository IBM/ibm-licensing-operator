#!/bin/bash
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
LOCALBIN="${1:?Usage: $0 <localbin-path>}"
VENV_DIR="${LOCALBIN}/.venv"
ACTIVE_OS=

function detect_os() {
  if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    ACTIVE_OS="Linux"
  elif [[ "$OSTYPE" == "darwin"* ]]; then
    if [[ $(uname -m) == 'arm64' ]]; then
      ACTIVE_OS="MacOS_arm64"
    else
      ACTIVE_OS="MacOS_x86"
    fi
  else
    ACTIVE_OS="Unknown"
  fi

  echo "Active Operating System: ${ACTIVE_OS}"
  echo
}

function check_prerequisites() {
  if ! [ -x "$(command -v python3)" ]; then
    echo ">>> Tool not found: python3. Install suitable version and try again."
    exit 1
  fi
  if ! [ -x "$(command -v go)" ]; then
    echo ">>> Tool not found: go. Install suitable version and try again."
    exit 1
  fi
}

##shellcheck-v0.8.0.darwin.x86_64.tar.xz
##shellcheck-v0.8.0.linux.x86_64.tar.xz
install_shellcheck_from_binary() {
  binaryName=$1
  # Download compressed
  curl -sSfL "https://github.com/koalaman/shellcheck/releases/download/${SHELLCHECK_VERSION}/${binaryName}" \
    -o "${binaryName}"
  # Decompress
  tar -xf "$binaryName"
  # Install binary
  chmod +x shellcheck-"${SHELLCHECK_VERSION}"/shellcheck
  cp shellcheck-"${SHELLCHECK_VERSION}"/shellcheck "${LOCALBIN}/shellcheck"
  rm -rf shellcheck-"${SHELLCHECK_VERSION}" "${binaryName}"
}

# linter versions
SHELLCHECK_VERSION=v0.8.0
YAMLLINT_VERSION=1.28.0
GOLANGCI_LINT_VERSION=v2.11.2
GOIMPORTS_VERSION=v0.3.0
DIFFUTILS_VERSION=v3.8

##### Start script logic #####

detect_os
check_prerequisites

# Ensure the local bin dir and venv exist
mkdir -p "${LOCALBIN}"
if [ ! -d "${VENV_DIR}" ]; then
  echo ">>> Creating Python virtual environment at ${VENV_DIR}"
  python3 -m venv "${VENV_DIR}"
fi

# Shellcheck
# Note: shellcheck v0.8.0 has no arm64 darwin binary; the x86_64 binary runs via Rosetta 2 on Apple Silicon.
if ! [ -x "${LOCALBIN}/shellcheck" ]; then
  echo ">>> Installing shellcheck [${SHELLCHECK_VERSION}]"
  if [ "${ACTIVE_OS}" == 'MacOS_arm64' ] || [ "${ACTIVE_OS}" == 'MacOS_x86' ]; then
    install_shellcheck_from_binary shellcheck-"${SHELLCHECK_VERSION}".darwin.x86_64.tar.xz
  else
    install_shellcheck_from_binary shellcheck-"${SHELLCHECK_VERSION}".linux.x86_64.tar.xz
  fi
else
  echo ">>> Shellcheck already installed."
  "${LOCALBIN}/shellcheck" --version
  echo
fi

# Yamllint (installed in shared venv)
if ! [ -x "${VENV_DIR}/bin/yamllint" ]; then
  echo ">>> Installing yamllint [${YAMLLINT_VERSION}]"
  "${VENV_DIR}/bin/pip" install "yamllint==${YAMLLINT_VERSION}"
  chmod +x "${VENV_DIR}/bin/yamllint"
else
  echo ">>> Yamllint already installed"
  "${VENV_DIR}/bin/yamllint" --version
  echo
fi

# Golangci-lint
if ! [ -x "${LOCALBIN}/golangci-lint" ]; then
  echo ">>> Installing golangci-lint [${GOLANGCI_LINT_VERSION}]"
  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "${LOCALBIN}" "${GOLANGCI_LINT_VERSION}"
else
  echo ">>> Golangci-lint already installed"
  "${LOCALBIN}/golangci-lint" --version
  echo
fi

# goimports
if ! [ -x "${LOCALBIN}/goimports" ]; then
  echo ">>> Installing goimports [${GOIMPORTS_VERSION}]"
  GOBIN="${LOCALBIN}" go install golang.org/x/tools/cmd/goimports@"${GOIMPORTS_VERSION}"
else
  echo ">>> Goimports already installed"
  echo
fi

# Diffutils (system package — cannot be installed to a local directory)
if ! [ -x "$(command -v diff3)" ]; then
  echo ">>> Installing diffutils [${DIFFUTILS_VERSION}]"
  if [ "${ACTIVE_OS}" == 'Linux' ]; then
    sudo apt-get update
    sudo apt-get install diffutils
  else
    brew install diffutils
  fi
else
  echo ">>> Diffutils already installed"
  diff3 --version
  echo
fi

echo ">>> Installation finished"
