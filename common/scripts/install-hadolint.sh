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

HADOLINT_VERSION="${1:?Usage: $0 <hadolint-version> <install-path>}"
INSTALL_PATH="${2:?Usage: $0 <hadolint-version> <install-path>}"

mkdir -p "$(dirname "${INSTALL_PATH}")"

if [[ "$OSTYPE" == "darwin"* ]]; then
  OS="macos"
else
  OS="linux"
fi

ARCH="$(uname -m)"
if [[ "${ARCH}" == "aarch64" ]]; then
  ARCH="arm64"
fi

BINARY_NAME="hadolint-${OS}-${ARCH}"
TMPDIR="$(mktemp -d)"
trap 'rm -rf "${TMPDIR}"' EXIT

curl -sSfL \
  "https://github.com/hadolint/hadolint/releases/download/${HADOLINT_VERSION}/${BINARY_NAME}" \
  -o "${TMPDIR}/hadolint"
chmod +x "${TMPDIR}/hadolint"
cp "${TMPDIR}/hadolint" "${INSTALL_PATH}"
echo ">>> hadolint ${HADOLINT_VERSION} installed to ${INSTALL_PATH}"
