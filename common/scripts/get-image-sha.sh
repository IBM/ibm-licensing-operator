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

# Get the SHA from an operand image and put it in operator.yaml and the CSV file.
# Do "docker login" before running this script.
# Run this script from the parent dir by typing "scripts/get-image-sha.sh"


# check the input parms
NAME=$1
VERSION=$2
if [[ $VERSION == "" ]]
then
   echo "Missing parm. Need image type, image name and image tag"
   echo "Example:"
   echo "   icr.io/cpopen/cpfs/ibm-licensing 1.15.0"
   exit 1
fi

# pull the image
IMAGE="$NAME:$VERSION"
echo "Pulling image $IMAGE"
docker pull "$IMAGE"

# get the SHA for the image
DIGEST="$(docker images --digests "$NAME" | grep "$VERSION" | awk 'FNR==1{print $3}')"

# DIGEST should look like this: sha256:10a844ffaf7733176e927e6c4faa04c2bc4410cf4d4ef61b9ae5240aa62d1456
if [[ $DIGEST != sha256* ]]
then
    echo "Cannot find SHA (sha256:nnnnnnnnnnnn) in digest: $DIGEST"
    exit 1
fi

SHA=$DIGEST
echo "SHA=$SHA"

#---------------------------------------------------------
# update operator.yaml
#---------------------------------------------------------
OPER_FILE=config/manager/manager.yaml

# delete the "name" and "value" lines for the old SHA
# for example:
#     - name: IBM_LICENSING_IMAGE
#       value: icr.io/cpopen/cpfs/ibm-licensing@sha256:10a844ffaf7733176e927e6c4faa04c2bc4410cf4d4ef61b9ae5240aa62d1456

sed -i "/name: IBM_LICENSING_IMAGE/{N;d;}" $OPER_FILE

# insert the new SHA lines
LINE1="\            - name: IBM_LICENSING_IMAGE"
LINE2="\              value: $NAME@$SHA"
sed -i "/env:/a $LINE1\n$LINE2" $OPER_FILE

#---------------------------------------------------------
# update the CSV
#---------------------------------------------------------
CSV_FILE=deploy/olm-catalog/ibm-licensing-operator/"${VERSION}"/ibm-licensing-operator.v"${VERSION}".clusterserviceversion.yaml

# delete the "name" and "value" lines for the old SHA
# for example:
#     - name: IBM_LICENSING_IMAGE
#       value: icr.io/cpopen/cpfs/ibm-licensing@sha256:10a844ffaf7733176e927e6c4faa04c2bc4410cf4d4ef61b9ae5240aa62d1456

sed -i "/name: IBM_LICENSING_IMAGE/{N;d;}" "${CSV_FILE}"

# insert the new SHA lines. need 4 more leading spaces compared to operator.yaml
LINE1="\                - name: IBM_LICENSING_IMAGE"
LINE2="\                  value: $NAME@$SHA"
sed -i "/env:/a $LINE1\n$LINE2" "${CSV_FILE}"
