
# ---
# Build Process
# ---

ARG builder_image="golang"

FROM indralabs/scratch:latest as config

FROM ${builder_image} AS builder

ARG source_release_url_prefix="https://github.com/btcsuite/btcd"

ARG target_os="linux"
ARG target_platform="amd64"
ARG target_version="v0.23.3"

WORKDIR /tmp

RUN set -ex echo "downloading source and binaries with manifest and signature." \
    && wget ${source_release_url_prefix}/releases/download/${target_version}/manifest-${target_version}.txt  \
    && wget ${source_release_url_prefix}/releases/download/${target_version}/manifest-guggero-${target_version}.sig  \
    && wget ${source_release_url_prefix}/releases/download/${target_version}/btcd-${target_os}-${target_platform}-${target_version}.tar.gz \
    && wget ${source_release_url_prefix}/releases/download/${target_version}/btcd-source-${target_version}.tar.gz

# Importing keys from scratch
COPY --from=config /etc/btcd/keys/guggero.asc /tmp/guggero.asc

RUN set -ex echo "importing keys" \
    && cat guggero.asc | gpg --import

RUN set -ex echo "running signature verification on manifest" \
    && gpg --verify manifest-guggero-${target_version}.sig manifest-${target_version}.txt

RUN set -ex echo "verifying checksum on btcd-${target_os}-${target_platform}-${target_version}.tar.gz" \
    && cat manifest-${target_version}.txt | grep btcd-${target_os}-${target_platform}-${target_version}.tar.gz | shasum -a 256 -c

#RUN set -ex echo "DEBUG: verifying a checksum failure stops the build" \
#    && mv btcd-${target_os}-${target_platform}-${target_version}.tar.gz btcd-source-${target_version}.tar.gz

RUN set -ex echo "verifying checksum on btcd-source-${target_version}.tar.gz" \
    && cat manifest-${target_version}.txt | grep btcd-source-${target_version}.tar.gz | shasum -a 256 -c

RUN set -ex echo "untarring binaries and source code" \
    && mkdir -pv /tmp/btcd-${target_os}-${target_platform}-${target_version} \
    && tar -xzvf btcd-${target_os}-${target_platform}-${target_version}.tar.gz --directory /tmp/btcd-${target_os}-${target_platform}-${target_version} \
    && mkdir -pv /tmp/btcd-source-${target_version} \
    && tar -xzvf btcd-source-${target_version}.tar.gz --directory /tmp/btcd-source-${target_version}

WORKDIR /tmp/btcd-source-${target_version}

RUN set -ex ls -hal /tmp

RUN set -ex echo "building binaries for ${GOOS}/${GOARCH}" \
    && mkdir -pv /tmp/bin \
    && GO111MODULE=on GOOS=${target_os} CGO_ENABLED=0 go build --ldflags '-w -s' -o /tmp/bin/btcd . \
    && GO111MODULE=on GOOS=${target_os} CGO_ENABLED=0 go build --ldflags '-w -s' -o /tmp/bin/ ./cmd/...

#RUN set -ex echo "moving btcd binary to /tmp/bin" \
#    && mkdir -pv /tmp/bin \
#    && cp /tmp/btcd-${target_os}-${target_platform}-${target_version}/btcd /tmp/bin
