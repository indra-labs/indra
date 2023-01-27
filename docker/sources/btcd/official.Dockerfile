
# ---
# Build Process
# ---

ARG sourcing_image="golang"

FROM indralabs/scratch:latest as scratch

FROM ${sourcing_image} as source

ARG source_url="https://github.com/btcsuite/btcd/releases/download"
ARG source_version="v0.23.3"

WORKDIR /tmp

RUN set -ex echo "downloading source and binaries with manifest and signature." \
    && wget ${source_url}/${source_version}/manifest-${source_version}.txt  \
    && wget ${source_url}/${source_version}/manifest-guggero-${source_version}.sig  \
    && wget ${source_url}/${source_version}/btcd-source-${source_version}.tar.gz

# Importing keys from scratch
COPY --from=scratch /etc/btcd/keys/guggero.asc /tmp/guggero.asc

RUN set -ex echo "importing keys" \
    && cat guggero.asc | gpg --import

RUN set -ex echo "running signature verification on manifest" \
    && gpg --verify manifest-guggero-${source_version}.sig manifest-${source_version}.txt

RUN set -ex echo "verifying checksum on btcd-source-${source_version}.tar.gz" \
    && cat manifest-${source_version}.txt | grep btcd-source-${source_version}.tar.gz | shasum -a 256 -c

RUN set -ex echo "untarring binaries and source code" \
    && mv btcd-source-${source_version}.tar.gz /tmp/btcd-source.tar.gz \
    && mkdir -pv /tmp/btcd-source \
    && tar -xzvf btcd-source.tar.gz --directory /tmp/btcd-source

WORKDIR /tmp/btcd-source

RUN set -ex echo "downloading modules" \
    && go mod vendor

FROM scratch

COPY --from=source /tmp/btcd-source /source
COPY --from=source /tmp/btcd-source.tar.gz /source.tar.gz
