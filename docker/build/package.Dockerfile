
# ---
# Packaging Process
# ---

ARG binaries_image="indralabs/indra-build"
ARG binaries_version="dev"

ARG packaging_image="golang:1.19.4"

FROM ${binaries_image}:${binaries_version} as build

FROM ${packaging_image} as packager

ARG binaries_version="dev"

ARG target_name="indra"
ARG target_packaging_script="docker/indra/build/package.sh"

COPY --from=build /${target_name}-${binaries_version} /tmp/${target_name}-${binaries_version}

WORKDIR /tmp/${target_name}-${binaries_version}

ADD docker/build/scripts/package.sh /tmp/package.sh

RUN set -ex echo "running packaging" \
    && /tmp/package.sh

WORKDIR /tmp/${target_name}-${binaries_version}/release

RUN set -ex echo "generating shasum for release" \
    && shasum -a 256 * > manifest-${binaries_version}.txt

RUN set -ex echo "generating tar archive for the whole release" \
    && tar -czvf /tmp/${target_name}-${binaries_version}.tar.gz -C /tmp/${target_name}-${binaries_version} .

FROM busybox

ARG target_name="indra"
ARG binaries_version="dev"

COPY --from=packager /tmp/${target_name}-${binaries_version} /tmp/${target_name}-${binaries_version}
COPY --from=packager /tmp/${target_name}-${binaries_version}.tar.gz /tmp/${target_name}-${binaries_version}.tar.gz
