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

# Check for prerequisites:
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

# Install detect-secrets:
if ! [ -x "$(command -v detect-secrets)" ]; then
  echo " » Installing detect-secrets"
  $pip install --upgrade "git+https://github.com/ibm/detect-secrets.git@master#egg=detect-secrets"
else
  echo " » detect-secrets already installed"
  detect-secrets --version
  echo
fi
