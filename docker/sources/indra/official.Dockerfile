
# ---
# Build Process
# ---

ARG sourcing_image="golang"

FROM indralabs/scratch:latest as scratch

FROM ${sourcing_image} as source

ARG source_url="https://github.com/indra-labs/indra/releases/download"
ARG source_version="v0.1.10"

WORKDIR /tmp

RUN set -ex echo "downloading source and binaries with manifest and signature." \
    && wget ${source_url}/${source_version}/manifest-${source_version}.txt  \
    && wget ${source_url}/${source_version}/manifest-greg.stone-${source_version}.sig  \
    && wget ${source_url}/${source_version}/manifest-херетик-${source_version}.sig  \
    && wget ${source_url}/${source_version}/indra-source-${source_version}.tar.gz

# Importing keys from scratch
COPY --from=scratch /etc/indra/keys/greg.stone.asc /tmp/greg.stone.asc
COPY --from=scratch /etc/indra/keys/херетик.asc /tmp/херетик.asc

RUN set -ex echo "importing keys" \
    && cat a.asc | gpg --import \
    && cat b.asc | gpg --import

RUN set -ex echo "running signature verification on manifest" \
    && gpg --verify manifest-greg.stone-${source_version}.sig manifest-${source_version}.txt \
    && gpg --verify manifest-херетик-${source_version}.sig manifest-${source_version}.txt

RUN set -ex echo "verifying checksum on indra-source-${source_version}.tar.gz" \
    && cat manifest-${source_version}.txt | grep indra-source-${source_version}.tar.gz | shasum -a 256 -c

RUN set -ex echo "untarring binaries and source code" \
    && mv indra-source-${source_version}.tar.gz /tmp/indra-source.tar.gz \
    && mkdir -pv /tmp/indra-source \
    && tar -xzvf indra-source.tar.gz --directory /tmp/indra-source

WORKDIR /tmp/indra-source

RUN set -ex echo "downloading modules" \
    && go mod vendor

FROM scratch

COPY --from=source /tmp/indra-source /source
COPY --from=source /tmp/indra-source.tar.gz /source.tar.gz
