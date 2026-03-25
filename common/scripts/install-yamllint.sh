#!/bin/bash
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

LOCALBIN="${1:?Usage: $0 <localbin-path> <yamllint-version>}"
YAMLLINT_VERSION="${2:?Usage: $0 <localbin-path> <yamllint-version>}"
VENV_DIR="${LOCALBIN}/.venv"

if ! [ -x "$(command -v python3)" ]; then
  echo ">>> Tool not found: python3. Install suitable version and try again."
  exit 1
fi

mkdir -p "${LOCALBIN}"

if [ ! -d "${VENV_DIR}" ]; then
  echo ">>> Creating Python virtual environment at ${VENV_DIR}"
  python3 -m venv "${VENV_DIR}"
fi

echo ">>> Installing yamllint [${YAMLLINT_VERSION}]"
"${VENV_DIR}/bin/pip" install "yamllint==${YAMLLINT_VERSION}"
chmod +x "${VENV_DIR}/bin/yamllint"
echo ">>> yamllint ${YAMLLINT_VERSION} installed to ${VENV_DIR}/bin/yamllint"
