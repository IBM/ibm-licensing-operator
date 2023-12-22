FROM golang:latest AS operator-registry

RUN apk --no-cache add \
        bash \
        binutils \
        build-base \
        curl \
        git \
        jq \
        mercurial \
        sqlite

RUN VERSION=$(curl -s https://api.github.com/repos/operator-framework/operator-registry/releases/latest | jq -r .tag_name) \
    && git clone -b $VERSION https://github.com/operator-framework/operator-registry /build

WORKDIR /build

RUN make static \
    && strip /build/bin/*

RUN GRPC_HEALTH_PROBE_VERSION=$(curl -s https://api.github.com/repos/grpc-ecosystem/grpc-health-probe/releases/latest | jq -r .tag_name) \
    && wget -qO/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-$(go env GOARCH) \
    && chmod +x /bin/grpc_health_probe \
    && strip /bin/grpc_health_probe

RUN cp /build/bin/opm /bin/opm \
    && cp /build/bin/initializer /bin/initializer \
    && cp /build/bin/appregistry-server /bin/appregistry-server \
    && cp /build/bin/configmap-server /bin/configmap-server \
    && cp /build/bin/registry-server /bin/registry-server

USER 1001

FROM alpine AS builder

COPY manifests manifests
COPY --from=operator-registry /bin/initializer /bin/initializer

RUN /bin/initializer -o ./bundles.db

USER 1001

FROM scratch

ARG VCS_REF
ARG VCS_URL

LABEL org.label-schema.vendor="IBM" \
  org.label-schema.name="IBM Common Service Catalog" \
  org.label-schema.description="The IBM Common Service Catalog image" \
  org.label-schema.vcs-ref=$VCS_REF \
  org.label-schema.vcs-url=$VCS_URL \
  org.label-schema.license="Licensed Materials - Property of IBM" \
  org.label-schema.schema-version="1.0" \
  name="ibm-common-service-catalog" \
  vendor="IBM" \
  description="The Operator CatalogSource image to host all IBM Common Services Operators" \
  summary="The Operator CatalogSource image to host all IBM Common Services Operators"

COPY --from=builder /tmp /tmp
COPY --from=builder bundles.db /bundles.db
COPY --from=operator-registry /bin/registry-server /bin/grpc_health_probe /bin/

EXPOSE 50051
ENTRYPOINT ["/bin/registry-server"]
CMD ["--database", "/bundles.db"]

USER 1001