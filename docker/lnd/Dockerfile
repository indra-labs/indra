
ARG base_image="golang"
ARG target_image="indralabs/scratch"

# ---
# Build Process
# ---

FROM ${base_image} AS builder

# Get the repo and build
ARG git_repository="github.com/indra-labs/lnd"
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

RUN cp sample-lnd.conf /tmp/lnd.conf

RUN set -ex \
  && CGO_ENABLED=0 go build --ldflags '-w -s' -o /tmp/bin/lnd ./cmd/lnd/. \
  && CGO_ENABLED=0 go build --ldflags '-w -s' -o /tmp/bin/lncli ./cmd/lncli/.

# ---
# Configure and Build the target container
# ---

FROM indralabs/scratch:latest

## Migrate the binaries and storage folder
COPY --from=builder /tmp/lnd.conf /etc/lnd/lnd.conf
COPY --from=builder /tmp/bin /bin

# Enable the indra user
USER lnd:lnd

# ENV defaults
# ENV IND_LOGFILEPATH=""

# Set the data volume
# VOLUME ["/etc/lnd"]
# VOLUME ["/var/lnd"]

# :9735  lnd peer-to-peer port
# :10009  lnd RPC port
EXPOSE 9735 10009

ENTRYPOINT ["/bin/lnd", "--lnddir=/var/lnd", "--configfile=/etc/lnd/lnd.conf", "--datadir=/var/lnd", "--tlskeypath=/etc/lnd/keys/", "--tlscertpath=/etc/lnd/keys/", "--btcd.rpccert=/etc/btcd/keys/rpc.cert", "--bitcoin.active"]
CMD ["--bitcoin.mainnet"]

# docker run -it --rm --entrypoint="/bin/lncli" -v ~/rpc.cert:/var/lnd/.lnd/rpc.cert indralabs/lnd --skipverify -s <host>:<port> -u <username> -P <password> <function>