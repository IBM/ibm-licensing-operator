#!/bin/bash
#
# Copyright 2020 IBM Corporation
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
TIMESTAMP=$(date +%s)

echo "Creating new CSV"
cd common/scripts/catalog || exit
cp -r ../../../deploy/olm-catalog manifests
LATEST_VERSION=$(ls -F manifests/ibm-licensing-operator | grep /\$ | sort -t. -k 1,1n -k 2,2n -k 3,3n | tail -n 1 | cut -f 1 -d "/")
CATALOG_VERSION=$(echo "$LATEST_VERSION" | cut -d "." -f -2).$TIMESTAMP
NEW_CSV=manifests/ibm-licensing-operator/"$CATALOG_VERSION"/ibm-licensing-operator.v"$CATALOG_VERSION".clusterserviceversion.yaml
PACKAGE=manifests/ibm-licensing-operator/ibm-licensing-operator.package.yaml
cp -r manifests/ibm-licensing-operator/"$LATEST_VERSION" manifests/ibm-licensing-operator/"$CATALOG_VERSION"
mv manifests/ibm-licensing-operator/"$CATALOG_VERSION"/ibm* "$NEW_CSV"


sed -i "/replaces/c\  replaces: ibm-licensing-operator.v$LATEST_VERSION" "$NEW_CSV"
sed -i "/olm.skipRange:/c\    olm.skipRange: \'>=1.0.0 <$CATALOG_VERSION\'" "$NEW_CSV"
sed -i "/name: ibm-licensing-operator.v/c\  name: ibm-licensing-operator.v$CATALOG_VERSION" "$NEW_CSV"
sed -i "s|quay.io/opencloudio/ibm-licensing-operator:.*|${IMAGE_REPO}/${IMAGE_NAME}@${DIGEST}|" "$NEW_CSV"

sed -i "/channels:/a\- currentCSV: ibm-licensing-operator.v$CATALOG_VERSION\n  name: devops" "$PACKAGE"

VCS_URL=https://github.com/IBM/ibm-common-service-catalog
VCS_REF=random

echo "Building and pushing catalog"
docker build -t "$CATALOG_NAME":latest --build-arg \ VCS_REF=${VCS_REF} --build-arg VCS_URL=${VCS_URL} -f Dockerfile .
docker push "$CATALOG_NAME":latest

rm -rdf manifests


