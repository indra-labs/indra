
# ---
# Build Process
# ---

ARG source_image="indralabs/indra-source"
ARG source_version="dev"

ARG builder_image="golang:1.19.4"

FROM ${source_image}:${source_version} as source

FROM ${builder_image} as build

ARG source_version="dev"

ARG target_name="indra"
ARG target_build_script="docker/indra/build/build.sh"

COPY --from=source /source /tmp/source
COPY --from=source /source.tar.gz /tmp/source.tar.gz

WORKDIR /tmp/source

RUN set -ex echo "making directories for release and binaries" \
    && mkdir -pv /tmp/${target_name}-${source_version}/release \
    && mkdir -pv /tmp/${target_name}-${source_version}/bin

RUN set -ex echo "migrating source files to release" \
    && cp /tmp/source.tar.gz /tmp/${target_name}-${source_version}/release/${target_name}-source-${source_version}.tar.gz

ADD ${target_build_script} ./build.sh

RUN set -ex echo "running build" && ./build.sh

FROM scratch

ARG target_name="indra"
ARG source_version="dev"

COPY --from=build /tmp/${target_name}-${source_version} /${target_name}-${source_version}
