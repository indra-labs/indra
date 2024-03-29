
# ---
# Build Process
# ---

ARG sourcing_image="golang"

FROM indralabs/scratch:latest as scratch

FROM ${sourcing_image} as source

ARG source_url="https://github.com/lightningnetwork/lnd/releases/download"
ARG source_version="v0.15.5-beta"

WORKDIR /tmp

RUN set -ex echo "downloading source and binaries with manifest and signature." \
    && wget ${source_url}/${source_version}/manifest-${source_version}.txt  \
    && wget ${source_url}/${source_version}/manifest-roasbeef-${source_version}.sig  \
    && wget ${source_url}/${source_version}/lnd-source-${source_version}.tar.gz

# Importing keys from scratch
COPY --from=scratch /etc/lnd/keys/roasbeef.asc /tmp/roasbeef.asc

RUN set -ex echo "importing keys" \
    && cat roasbeef.asc | gpg --import

RUN set -ex echo "running signature verification on manifest" \
    && gpg --verify manifest-roasbeef-${source_version}.sig manifest-${source_version}.txt

RUN set -ex echo "verifying checksum on lnd-source-${source_version}.tar.gz" \
    && cat manifest-${source_version}.txt | grep lnd-source-${source_version}.tar.gz | shasum -a 256 -c

RUN set -ex echo "untarring binaries and source code" \
    && mv lnd-source-${source_version}.tar.gz /tmp/lnd-source.tar.gz \
    && mkdir -pv /tmp/lnd-source \
    && tar -xzvf lnd-source.tar.gz --directory /tmp/lnd-source

WORKDIR /tmp/lnd-source/lnd-source

RUN set -ex echo "downloading modules" \
    && go mod vendor

FROM scratch

COPY --from=source /tmp/lnd-source/lnd-source /source
COPY --from=source /tmp/lnd-source.tar.gz /source.tar.gz
