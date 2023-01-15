
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

# Source/Target release defaults
ARG ARCH=amd64
ARG GOARCH=amd64

ENV GO111MODULE=on GOOS=linux

WORKDIR $GOPATH/src/${git_repository}

RUN cp sample-btcd.conf /tmp/btcd.conf

RUN set -ex \
  && CGO_ENABLED=0 go build --ldflags '-w -s' -o /tmp/bin/btcd . \
  && CGO_ENABLED=0 go build --ldflags '-w -s' -o /tmp/bin/ ./cmd/...

# ---
# Target Configuration
# ---

FROM indralabs/scratch:latest

## Migrate the binaries and storage folder
COPY --from=builder /tmp/btcd.conf /etc/btcd/btcd.conf
COPY --from=builder /tmp/bin /bin

# Enable the btcd user
USER btcd:btcd

# Set the data volumes
#VOLUME ["/etc/btcd"]
#VOLUME ["/var/btcd"]

# :8333  btcd peer-to-peer port
# :8334  btcd RPC port
EXPOSE 8333 8334

ENTRYPOINT ["/bin/btcd", "--configfile=/etc/btcd/btcd.conf", "--datadir=/var/btcd", "--logdir=/var/btcd", "--rpckey=/etc/btcd/keys/rpc.key", "--rpccert=/etc/btcd/keys/rpc.cert", "--listen=0.0.0.0:8333", "--rpclisten=0.0.0.0:8334"]
