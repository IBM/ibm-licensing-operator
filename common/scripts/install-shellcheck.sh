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

set -euo pipefail

SHELLCHECK_VERSION="${1:?Usage: $0 <shellcheck-version> <install-path>}"
INSTALL_PATH="${2:?Usage: $0 <shellcheck-version> <install-path>}"

mkdir -p "$(dirname "${INSTALL_PATH}")"

# Note: shellcheck has no arm64 darwin binary; the x86_64 binary runs via Rosetta 2 on Apple Silicon.
if [[ "$OSTYPE" == "darwin"* ]]; then
  BINARY_NAME="shellcheck-${SHELLCHECK_VERSION}.darwin.x86_64.tar.xz"
else
  BINARY_NAME="shellcheck-${SHELLCHECK_VERSION}.linux.x86_64.tar.xz"
fi

TMPDIR="$(mktemp -d)"
trap 'rm -rf "${TMPDIR}"' EXIT

curl -sSfL "https://github.com/koalaman/shellcheck/releases/download/${SHELLCHECK_VERSION}/${BINARY_NAME}" \
  -o "${TMPDIR}/${BINARY_NAME}"
tar -xf "${TMPDIR}/${BINARY_NAME}" -C "${TMPDIR}"
chmod +x "${TMPDIR}/shellcheck-${SHELLCHECK_VERSION}/shellcheck"
cp "${TMPDIR}/shellcheck-${SHELLCHECK_VERSION}/shellcheck" "${INSTALL_PATH}"
echo ">>> shellcheck ${SHELLCHECK_VERSION} installed to ${INSTALL_PATH}"
