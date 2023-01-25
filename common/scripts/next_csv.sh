#!/usr/bin/env bash

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


if [[ $OSTYPE == 'darwin'* ]]; then
    inline_sed() {
        sed -i "" "$@"
    }
else
    inline_sed() {
        sed -i "$@"
    }
fi

# This script needs to inputs
# The CSV version that is currently in dev

# cs operator
CURRENT_DEV_CSV=$1
NEW_DEV_CSV=$2
PREVIOUS_DEV_CSV=$3

# Update bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml
inline_sed "s/$CURRENT_DEV_CSV/$NEW_DEV_CSV/g" bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml
inline_sed "s/$PREVIOUS_DEV_CSV/$CURRENT_DEV_CSV/g" bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml
echo "Updated the bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml"

# Update config/manifests/bases/ibm-licensing-operator.clusterserviceversion.yaml
inline_sed "s/$CURRENT_DEV_CSV/$NEW_DEV_CSV/g" config/manifests/bases/ibm-licensing-operator.clusterserviceversion.yaml
echo "Updated the config/manifests/bases/ibm-licensing-operator.clusterserviceversion.yaml"

# Update cs operator version only
inline_sed "s/$CURRENT_DEV_CSV/$NEW_DEV_CSV/g" version/version.go
echo "Updated the multiarch_image.sh"

inline_sed "s/$CURRENT_DEV_CSV/&, $NEW_DEV_CSV/" README.md
echo "Updated the README.md"

# Update bundle/manifests/ibm-licensing-operator.clusterserviceversion.yaml
inline_sed "s/$CURRENT_DEV_CSV/$NEW_DEV_CSV/g" Makefile
inline_sed "s/$PREVIOUS_DEV_CSV/$CURRENT_DEV_CSV/g" Makefile
echo "Updated the Makefile"

inline_sed "s/$CURRENT_DEV_CSV/$NEW_DEV_CSV/" config/manager/manager.yaml
echo "Updated the config/manager/manager.yaml"

inline_sed "s/$CURRENT_DEV_CSV/$NEW_DEV_CSV/" config/samples/operator.ibm.com_v1alpha1_ibmlicensing.yaml
echo "Updated the config/samples/operator.ibm.com_v1alpha1_ibmlicensing.yaml"

inline_sed "s/$CURRENT_DEV_CSV/$NEW_DEV_CSV/" config/samples/operator.ibm.com_v1alpha1_ibmlicenseservicereporter.yaml
echo "Updated the config/samples/operator.ibm.com_v1alpha1_ibmlicenseservicereporter.yaml"

# Update relatedImages (make bundle target)
inline_sed "s/$CURRENT_DEV_CSV/$NEW_DEV_CSV/" common/relatedImages.yaml
echo "Updated the common/relatedImages.yaml"