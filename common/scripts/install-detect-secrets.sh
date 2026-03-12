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

LOCALBIN="${1:?Usage: $0 <localbin-path>}"
VENV_DIR="${LOCALBIN}/.venv"
WRAPPER="${LOCALBIN}/detect-secrets"

mkdir -p "${LOCALBIN}"

# Create venv if it doesn't exist
if [ ! -d "${VENV_DIR}" ]; then
  echo ">>> Creating Python virtual environment at ${VENV_DIR}"
  python3 -m venv "${VENV_DIR}"
fi

# Install or verify detect-secrets inside the venv
if "${VENV_DIR}/bin/detect-secrets" --version >/dev/null 2>&1; then
  echo ">>> detect-secrets already installed in venv"
  "${VENV_DIR}/bin/detect-secrets" --version
else
  echo ">>> Installing detect-secrets into ${VENV_DIR}"
  "${VENV_DIR}/bin/pip" install --upgrade \
    "git+https://github.com/ibm/detect-secrets.git@master#egg=detect-secrets"
fi

# Create/update wrapper script in LOCALBIN so detect-secrets is on PATH
cat > "${WRAPPER}" <<WRAPPER_SCRIPT
#!/bin/bash
exec "${VENV_DIR}/bin/detect-secrets" "\$@"
WRAPPER_SCRIPT
chmod +x "${WRAPPER}"
echo ">>> detect-secrets wrapper installed at ${WRAPPER}"
