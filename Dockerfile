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

FROM docker-na-public.artifactory.swg-devops.com/hyc-cloud-private-edge-docker-local/build-images/ubi8-minimal:latest

ARG IMAGE_NAME
ARG IMAGE_DISPLAY_NAME
ARG IMAGE_NAME_ARCH
ARG IMAGE_MAINTAINER
ARG IMAGE_VENDOR
ARG IMAGE_VERSION
ARG VERSION
ARG IMAGE_RELEASE
ARG IMAGE_BUILDDATE
ARG IMAGE_DESCRIPTION
ARG IMAGE_SUMMARY
ARG IMAGE_OPENSHIFT_TAGS
ARG VCS_REF
ARG VCS_URL

LABEL org.label-schema.vendor="$IMAGE_VENDOR" \
      org.label-schema.name="$IMAGE_NAME_ARCH" \
      org.label-schema.description="$IMAGE_DESCRIPTION" \
      org.label-schema.vcs-ref=$VCS_REF \
      org.label-schema.vcs-url=$VCS_URL \
      org.label-schema.license="Licensed Materials - Property of IBM" \
      org.label-schema.schema-version="1.0" \
      name="$IMAGE_NAME" \
      maintainer="$IMAGE_MAINTAINER" \
      vendor="$IMAGE_VENDOR" \
      image-version="$IMAGE_VERSION" \
      version="$VERSION" \
      release="$IMAGE_RELEASE" \
      build-date="$IMAGE_BUILDDATE" \
      description="$IMAGE_DESCRIPTION" \
      summary="$IMAGE_SUMMARY" \
      io.k8s.display-name="$IMAGE_DISPLAY_NAME" \
      io.k8s.description="$IMAGE_DESCRIPTION" \
      io.openshift.tags="$IMAGE_OPENSHIFT_TAGS"

ENV OPERATOR=/usr/local/bin/ibm-licensing-operator \
  DEPLOY_DIR=/deploy \
  USER_UID=1001 \
  USER_NAME=ibm-licensing-operator \
  IMAGE_RELEASE="$IMAGE_RELEASE"

# install operator binary
COPY bin/ibm-licensing-operator ${OPERATOR}
COPY bundle ${DEPLOY_DIR}

# copy licenses
RUN mkdir /licenses
COPY LICENSE /licenses

COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup

# add commit image release
RUN  echo "$IMAGE_RELEASE" > /IMAGE_RELEASE \ 
  && echo "$IMAGE_BUILDDATE" > /IMAGE_BUILDDATE

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
