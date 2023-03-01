
# ---
# Build Process
# ---

ARG sourcing_image="golang"

FROM ${sourcing_image} as source

ADD . /tmp/indra-source

WORKDIR /tmp/indra-source

RUN set -ex "archiving indra source files" \
    && git archive --format=tar.gz -o /tmp/indra-source.tar.gz HEAD

WORKDIR /tmp

RUN set -ex "cleaning up repository" \
    && rm -rf /tmp/indra-source && mkdir /tmp/indra-source \
    && tar -xzvf /tmp/indra-source.tar.gz --directory indra-source

WORKDIR /tmp/indra-source

RUN set -ex echo "downloading modules" \
    && GOINSECURE=git-indra.lan/* GOPRIVATE=git-indra.lan/* go mod vendor

FROM scratch

COPY --from=source /tmp/indra-source /source
COPY --from=source /tmp/indra-source.tar.gz /source.tar.gz
