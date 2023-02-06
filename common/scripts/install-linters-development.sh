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

  echo " ✔ Active Operating System: ${ACTIVE_OS}"
  echo
}

function check_prerequisites() {
  if ! [ -x "$(command -v pip)" ] && ! [ -x "$(command -v pip3)" ]; then
    echo " » Tool not found: pip. Install suitable version and try again."
    exit 1
  elif [ -x "$(command -v pip3)" ]; then
    pip="pip3"
  elif [ -x "$(command -v pip)" ]; then
    pip="pip"
  fi
  if ! [ -x "$(command -v go)" ]; then
    echo " » Tool not found: go. Install suitable version and try again."
    exit 1
  fi
  if ! [ -x "$(command -v gem)" ]; then
    echo " » Tool not found: gem. Install suitable version and try again."
    exit 1
  fi
}

##hadolint-Darwin-x86_64
##hadolint-Linux-x86_64
install_hadolint_from_binary() {
  binaryName=$1
  # Download binary
  curl -LO "https://github.com/hadolint/hadolint/releases/download/${HADOLINT_VERSION}/${binaryName}"
  # Install binary
  chmod +x "$binaryName" && mkdir -p /usr/local/bin/ && cp "$binaryName" /usr/local/bin/hadolint && rm "$binaryName"
}

##shellcheck-v0.8.0.darwin.x86_64.tar.xz
##shellcheck-v0.8.0.linux.x86_64.tar.xz
install_shellcheck_from_binary() {
  binaryName=$1
  # Download compressed
  curl -LO "https://github.com/koalaman/shellcheck/releases/download/${SHELLCHECK_VERSION}/${binaryName}"
  # Decompress
  tar -xf "$binaryName"
  # Install binary
  chmod +x shellcheck-"${SHELLCHECK_VERSION}"/shellcheck && mkdir -p /usr/local/bin/ && cp shellcheck-"${SHELLCHECK_VERSION}"/shellcheck /usr/local/bin/shellcheck && rm -r shellcheck-"${SHELLCHECK_VERSION}." && rm "${binaryName}"
}

# linter versiions
HADOLINT_VERSION=v2.12.0
SHELLCHECK_VERSION=v0.8.0
YAMLLINT_VERSION=v1.28.0
GOLANGCI_LINT_VERSION=v1.50.1
MDL_VERSION=0.11.0
AWESOME_BOT_VERSION=1.20.0
GOIMPORTS_VERSION=v0.3.0
DIFFUTILS_VERSION=v3.8

##### Start script logic #####

detect_os
check_prerequisites

# Hadolint
if ! [ -x "$(command -v hadolint)" ]; then
  echo " » Installing hadolint [${HADOLINT_VERSION}]"
  if [ $ACTIVE_OS == 'MacOS_arm64' ]; then
    brew install hadolint
  elif [ $ACTIVE_OS == 'MacOS_x86' ]; then
    install_hadolint_from_binary hadolint-Darwin-x86_64
  else
    install_hadolint_from_binary hadolint-Linux-x86_64
  fi
else
  echo " » Hadolint already installed."
  hadolint --version
  echo
fi

# Shellcheck
if ! [ -x "$(command -v shellcheck)" ]; then
  echo " » Installing shellcheck [${SHELLCHECK_VERSION}]"
  if [ $ACTIVE_OS == 'MacOS_arm64' ]; then
    brew install shellcheck
  elif [ $ACTIVE_OS == 'MacOS_x86' ]; then
    install_shellcheck_from_binary shellcheck-"${SHELLCHECK_VERSION}".darwin.x86_64.tar.xz
  else
    install_shellcheck_from_binary shellcheck-"${SHELLCHECK_VERSION}".linux.x86_64.tar.xz
  fi

else
  echo " » Shellcheck already installed."
  shellcheck --version
  echo
fi

# Yamllint
if ! [ -x "$(command -v yamllint)" ]; then
  echo " » Installing yamllint [${YAMLLINT_VERSION}]"
  $pip install yamllint=="${YAMLLINT_VERSION}"
else
  echo " » Yamllint already installed"
  yamllint --version
  echo
fi

# Golangci-lint
if ! [ -x "$(command -v golangci-lint)" ]; then
  echo " » Installing golangci-lint [${GOLANGCI_LINT_VERSION}]"
  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b /usr/local/bin "${GOLANGCI_LINT_VERSION}"
else
  echo " » Golangci-lint already installed"
  golangci-lint --version
  echo
fi

# Mdl
if ! [ -x "$(command -v mdl)" ]; then
  echo " » Installing mdl [${MDL_VERSION}]"
  sudo gem install mdl -v ${MDL_VERSION}
else
  echo " » Mdl already installed"
  mdl --version
  echo
fi

# Awesome_bot
if ! [ -x "$(command -v awesome_bot)" ]; then
  echo " » Installing awesome_bot [${AWESOME_BOT_VERSION}]"
  sudo gem install awesome_bot -v ${AWESOME_BOT_VERSION}
else
  echo " » Awesome_bot already installed"
  awesome_bot --version
  echo
fi

# goimports
if ! [ -x "$(command -v goimports)" ]; then
  echo " » Installing goimports [${GOIMPORTS_VERSION}]"
  go install golang.org/x/tools/cmd/goimports@"${GOIMPORTS_VERSION}"
else
  echo " » Goimports already installed"
  echo
fi

# Diffutils
if ! [ -x "$(command -v diff3)" ]; then
  echo " » Installing diffutils [${DIFFUTILS_VERSION}]"
  if [ $ACTIVE_OS == 'Linux' ]; then
    sudo apt-get update
    sudo apt-get install diffutils
  else
    brew install diffutils
  fi
else
  echo " » Diffutils already installed"
  diff3 --version
  echo
fi

echo " » Installation finished"
