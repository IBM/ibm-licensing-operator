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

echo ">>> Installing Yq"

TARGET_OS=$1
LOCAL_ARCH=$2
YQ_VERSION=$3
INSTALL_PATH=$4

mkdir -p "$(dirname "${INSTALL_PATH}")"
# Download binary directly to the target path
curl -sSfL \
  "https://github.com/mikefarah/yq/releases/download/${YQ_VERSION}/yq_${TARGET_OS}_${LOCAL_ARCH}" \
  -o "${INSTALL_PATH}"
chmod +x "${INSTALL_PATH}"
echo ">>> yq ${YQ_VERSION} installed to ${INSTALL_PATH}"
