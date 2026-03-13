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

# Diffutils is a system package — it cannot be installed to a local directory.
# Version v3.8 is the minimum required.

function detect_os() {
  if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    echo "Linux"
  elif [[ "$OSTYPE" == "darwin"* ]]; then
    echo "MacOS"
  else
    echo "Unknown"
  fi
}

ACTIVE_OS="$(detect_os)"

if [ -x "$(command -v diff3)" ]; then
  echo ">>> Diffutils already installed"
  diff3 --version
  exit 0
fi

echo ">>> Installing diffutils"
if [ "${ACTIVE_OS}" == 'Linux' ]; then
  sudo apt-get update
  sudo apt-get install diffutils
else
  brew install diffutils
fi
