
# ---
# Build Process
# ---

ARG sourcing_image="golang"

FROM indralabs/scratch:latest as scratch

FROM ${sourcing_image}

ADD . /tmp/indra-source

WORKDIR /tmp/indra-source

RUN set -ex echo "downloading modules" \
    && go mod vendor