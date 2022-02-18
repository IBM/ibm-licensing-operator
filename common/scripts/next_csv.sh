#!/usr/bin/env bash

#
# Copyright 2022 IBM Corporation
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

# This script needs to inputs
# The CSV version that is currently in dev

# cs operator
CURRENT_DEV_CSV=$1
NEW_DEV_CSV=$2
PREVIOUS_DEV_CSV=$3

# Update bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml
gsed -i "s/$CURRENT_DEV_CSV/$NEW_DEV_CSV/g" bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml
gsed -i "s/$PREVIOUS_DEV_CSV/$CURRENT_DEV_CSV/g" bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml
echo "Updated the bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml"

# Update config/manifests/bases/ibm-licensing-operator.clusterserviceversion.yaml
gsed -i "s/$CURRENT_DEV_CSV/$NEW_DEV_CSV/g" config/manifests/bases/ibm-licensing-operator.clusterserviceversion.yaml
echo "Updated the config/manifests/bases/ibm-licensing-operator.clusterserviceversion.yaml"

# Update cs operator version only
gsed -i "s/$CURRENT_DEV_CSV/$NEW_DEV_CSV/g" version/version.go
echo "Updated the multiarch_image.sh"
gsed -i "s/$CURRENT_DEV_CSV/$NEW_DEV_CSV/" README.md
echo "Updated the README.md"

# Update bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml
gsed -i "s/$CURRENT_DEV_CSV/$NEW_DEV_CSV/g" Makefile
gsed -i "s/$PREVIOUS_DEV_CSV/$CURRENT_DEV_CSV/g" Makefile
echo "Updated the Makefile"

gsed -i "s/$CURRENT_DEV_CSV/$NEW_DEV_CSV/" config/manager/manager.yaml
echo "Updated the config/manager/manager.yaml"

gsed -i "s/$CURRENT_DEV_CSV/$NEW_DEV_CSV/" bundle/manifests/operator.ibm.com_ibmlicensings.yaml
gsed -i "s/$CURRENT_DEV_CSV/$NEW_DEV_CSV/" bundle/manifests/operator.ibm.com_ibmlicenseservicereporters.yaml
echo "Updated the CR examples"