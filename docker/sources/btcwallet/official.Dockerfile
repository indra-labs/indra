
# ---
# Build Process
# ---

ARG sourcing_image="golang"

FROM indralabs/scratch:latest as scratch

FROM ${sourcing_image} as source

ARG source_url="https://github.com/btcsuite/btcwallet"
ARG source_version="v0.16.5"

WORKDIR /tmp

RUN set -ex echo "downloading source" \
    && git clone ${source_url}

# Importing keys from scratch
COPY --from=scratch /etc/btcd/keys/guggero.asc /tmp/guggero.asc

RUN set -ex echo "importing keys" \
    && cat guggero.asc | gpg --import

WORKDIR /tmp/btcwallet

RUN set -ex echo "checking out tag" \
    && git checkout -b ${source_version}

RUN set -ex echo "running signature verification on tag" \
    && git verify-tag ${source_version}

RUN set -ex "archiving indra source files" \
    && git archive --format=tar.gz -o /tmp/btcwallet-source.tar.gz ${source_version}

WORKDIR /tmp

RUN set -ex echo "untarring archived source code" \
    && rm -rf btcwallet \
    && mkdir -pv btcwallet-source \
    && tar -xzvf btcwallet-source.tar.gz --directory btcwallet-source

WORKDIR /tmp/btcwallet-source

RUN set -ex echo "downloading modules" \
    && go mod vendor

FROM scratch

COPY --from=source /tmp/btcwallet-source /source
COPY --from=source /tmp/btcwallet-source.tar.gz /source.tar.gz
