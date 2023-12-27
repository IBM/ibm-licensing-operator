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


echo "Building catalog"

IMAGE_REPO=${1}
IMAGE_NAME=${2}
MANIFEST_VERSION=${3}

docker pull "$IMAGE_REPO/$IMAGE_NAME:$MANIFEST_VERSION"

DIGEST="$(docker images --digests "$IMAGE_REPO/$IMAGE_NAME" | grep "$MANIFEST_VERSION" | awk 'FNR==1{print $3}')"
CATALOG_NAME="${IMAGE_REPO}/ibm-licensing-catalog"

echo "Creating new CSV"
cd common/scripts/catalog || exit
cp -r ../../../bundle/manifests manifests
LATEST_VERSION=$(git tag | tail -n1 | tr -d v)
NEW_CSV=manifests/ibm-licensing-operator.clusterserviceversion.yaml
mv manifests/ibm* "$NEW_CSV"

sed -i "/replaces/c\  replaces: ibm-licensing-operator.v$LATEST_VERSION" "$NEW_CSV"
sed -i "/olm.skipRange:/c\    olm.skipRange: \'>=1.0.0 <$LATEST_VERSION\'" "$NEW_CSV"
sed -i "/name: ibm-licensing-operator.v/c\  name: ibm-licensing-operator.v$LATEST_VERSION" "$NEW_CSV"
sed -i "s|icr.io/cpopen/ibm-licensing-operator:.*|${IMAGE_REPO}/${IMAGE_NAME}@${DIGEST}|" "$NEW_CSV"

VCS_URL=https://github.com/IBM/ibm-common-service-catalog
VCS_REF=random

echo "Building and pushing catalog"
docker build -t "$CATALOG_NAME":"$MANIFEST_VERSION" --build-arg \ VCS_REF=${VCS_REF} --build-arg VCS_URL=${VCS_URL} --security-opt=no-new-privileges -f Dockerfile .
docker push "$CATALOG_NAME":"$MANIFEST_VERSION"
docker tag "$CATALOG_NAME":"$MANIFEST_VERSION" "$CATALOG_NAME":latest
docker push "$CATALOG_NAME":latest

rm -rdf manifests


