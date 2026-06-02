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

mkdir -p "${LOCALBIN}"

if [ ! -d "${VENV_DIR}" ]; then
  echo ">>> Creating Python virtual environment at ${VENV_DIR}"
  if command -v uv >/dev/null 2>&1; then
    uv venv "${VENV_DIR}"
  else
    echo "Creating venv using default python3"
    python3 -m venv "${VENV_DIR}"
  fi 
fi

echo ">>> Installing yamllint [${YAMLLINT_VERSION}]"
if command -v uv >/dev/null 2>&1; then
  uv pip install --python "${VENV_DIR}/bin/python" "yamllint==${YAMLLINT_VERSION}"
else
  "${VENV_DIR}/bin/pip" install "yamllint==${YAMLLINT_VERSION}"
fi

chmod +x "${VENV_DIR}/bin/yamllint"
echo ">>> yamllint ${YAMLLINT_VERSION} installed to ${VENV_DIR}/bin/yamllint"
