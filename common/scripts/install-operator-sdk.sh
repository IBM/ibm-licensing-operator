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

echo ">>> Installing Operator SDK"

TARGET_OS=$1
LOCAL_ARCH=$2
OPERATOR_SDK_VERSION=$3

# Download binary
curl -LO https://github.com/operator-framework/operator-sdk/releases/download/"${OPERATOR_SDK_VERSION}"/operator-sdk_"${TARGET_OS}"_"${LOCAL_ARCH}"
# Install binary
chmod +x operator-sdk_"${TARGET_OS}"_"${LOCAL_ARCH}" && mkdir -p /usr/local/bin/ && cp operator-sdk_"${TARGET_OS}"_"${LOCAL_ARCH}" /usr/local/bin/operator-sdk && rm operator-sdk_"${TARGET_OS}"_"${LOCAL_ARCH}"