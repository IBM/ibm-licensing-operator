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

# Script to build helm development charts
# Usage: build-helm-develop.sh <target-dir> <source-dir> <image-sed-pattern> <yq-namespace-key> <chart-name> <csv-version> <git-branch> <helm-bin> <yq-bin> <chart-destination> <artifactory-token>

set -e

TARGET_DIR=$1                # Target directory name for the generated helm chart
SOURCE_DIR=$2                # Source helm chart directory path
IMAGE_SED_PATTERN=$3         # Sed pattern to replace image tags (empty string "" to skip)
YQ_NAMESPACE_KEY=$4          # YQ key path for setting image registry namespace
CHART_NAME=$5                # Name of the chart for the output .tgz file
CSV_VERSION=$6               # Current CSV version
GIT_BRANCH=$7                # Git branch name to use for development images
HELM_BIN=$8                  # Path to helm binary
YQ_BIN=$9                    # Path to yq binary
CHART_DESTINATION=${10}      # Artifactory destination URL
ARTIFACTORY_TOKEN=${11}      # Artifactory API token

# Safety check: abort if target directory already exists
if [ -d "./${TARGET_DIR}" ]; then
    echo "Error: ${TARGET_DIR} directory already exists. Please remove it before running this script."
    exit 1
fi

echo "Building helm development chart: ${TARGET_DIR}"

# Copy helm directory to target to avoid modifying original files
cp -r "./${SOURCE_DIR}" "./${TARGET_DIR}"

# Set correct images for development (only if sed pattern is provided)
if [ -n "${IMAGE_SED_PATTERN}" ]; then
    tmp_file=$(mktemp)
    sed "${IMAGE_SED_PATTERN}" "./${TARGET_DIR}/templates/deployment.yaml" > "${tmp_file}"
    mv "${tmp_file}" "./${TARGET_DIR}/templates/deployment.yaml"
fi

# Update values.yaml with development registry settings
"${YQ_BIN}" -i '.global.imagePullPrefix = "docker-na-public.artifactory.swg-devops.com"' "./${TARGET_DIR}/values.yaml"
"${YQ_BIN}" -i ".${YQ_NAMESPACE_KEY} = \"hyc-cloud-private-scratch-docker-local/ibmcom\"" "./${TARGET_DIR}/values.yaml"

# Generate helm package
"${HELM_BIN}" package "./${TARGET_DIR}"

echo "Successfully built ${CHART_NAME}-${CSV_VERSION}.tgz"

# Publish helm charts
if [ -n "${CHART_DESTINATION}" ] && [ -n "${ARTIFACTORY_TOKEN}" ]; then
    echo "Publishing helm chart to ${CHART_DESTINATION}..."
    curl -s -w "\n" -H "X-JFrog-Art-Api: ${ARTIFACTORY_TOKEN}" -T "${CHART_NAME}-${CSV_VERSION}.tgz" "${CHART_DESTINATION}/${CHART_NAME}-develop.tgz"
    echo "Chart published successfully"
fi
