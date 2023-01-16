
ARG base_image="golang"
ARG target_image="indralabs/scratch"

# ---
# Build Process
# ---

FROM ${base_image} AS builder

# Get the repo and build
ARG git_repository="github.com/indra-labs/btcd"
ARG git_tag="master"

# Install dependencies and build the binaries.
RUN git clone "https://"${git_repository} /go/src/${git_repository}

WORKDIR $GOPATH/src/${git_repository}

RUN git checkout ${git_tag}

ARG ARCH=amd64
ARG GOARCH=amd64
RUN set -ex \
  && GO111MODULE=on GOOS=linux CGO_ENABLED=0 go build --ldflags '-w -s' -o /tmp/bin/btcd . \
  && GO111MODULE=on GOOS=linux CGO_ENABLED=0 go build --ldflags '-w -s' -o /tmp/bin/ ./cmd/...

# ---
# Target Configuration
# ---

FROM indralabs/scratch:latest

## Migrate the binaries and storage folder
COPY --from=builder /tmp/bin /bin

# Enable the btcd user
USER btcd:btcd

# Set the data volumes
#VOLUME ["/etc/btcd"]
#VOLUME ["/var/btcd"]

# :8333  btcd peer-to-peer port
# :8334  btcd RPC port
EXPOSE 8333 8334

ENTRYPOINT ["/bin/btcd", "--configfile=/etc/btcd/btcd.conf"]
