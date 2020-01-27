FROM registry.access.redhat.com/ubi8/ubi-minimal:8.1-328

ENV OPERATOR=/usr/local/bin/ibm-licensing-operator \
    USER_UID=1001 \
    USER_NAME=ibm-licensing-operator

# install operator binary
COPY build/_output/bin/ibm-licensing-operator ${OPERATOR}

COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup

LABEL org.label-schema.vendor="IBM" \
      org.label-schema.name="IBM License Service" \
      org.label-schema.description="Measure VPC license use of IBM products deployed in the cluster"

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
