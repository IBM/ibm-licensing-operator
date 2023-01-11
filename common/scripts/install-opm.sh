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

echo ">>> Installing Opm"

TARGET_OS=$1
LOCAL_ARCH=$2
OPM_VERSION=$3
# Download binary
curl -LO https://github.com/operator-framework/operator-registry/releases/download/"${OPM_VERSION}"/"${TARGET_OS}"-"${LOCAL_ARCH}"-opm
# Install binary
chmod +x "${TARGET_OS}"-"${LOCAL_ARCH}"-opm && mkdir -p /usr/local/bin/ && cp "${TARGET_OS}"-"${LOCAL_ARCH}"-opm /usr/local/bin/opm && rm "${TARGET_OS}"-"${LOCAL_ARCH}"-opm