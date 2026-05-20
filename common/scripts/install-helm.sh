#!/usr/bin/env bash

# Copyright 2026 IBM Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

LOCAL_BIN_DIR=$1
HELM_VERSION=$2
LOCAL_OS=$3
LOCAL_ARCH=$4

# Download and install helm https://helm.sh/docs/intro/install/#from-script

HELM_URL="https://get.helm.sh/helm-${HELM_VERSION}-${LOCAL_OS}-${LOCAL_ARCH}.tar.gz"
echo "Downloading helm from: $HELM_URL"

curl -sSL "$HELM_URL" -o "$LOCAL_BIN_DIR"/helm.tar.gz
tar -C "$LOCAL_BIN_DIR" -xzf "$LOCAL_BIN_DIR"/helm.tar.gz
mv "$LOCAL_BIN_DIR/${LOCAL_OS}-${LOCAL_ARCH}/helm" "$LOCAL_BIN_DIR"/helm
chmod +x "$LOCAL_BIN_DIR"/helm
rm -rf "$LOCAL_BIN_DIR"/helm.tar.gz "$LOCAL_BIN_DIR/${LOCAL_OS}-${LOCAL_ARCH}"