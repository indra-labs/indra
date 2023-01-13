FROM golang:1.19.4 AS builder

# User/Group definition
ENV USER=btcd GROUP=btcd UID=8333 GID=8333

## Create a user/group to be migrated to the target container
RUN addgroup ${GROUP} --gid ${GID} \
 && adduser \
    --disabled-password \
    --gecos "" \
    --home "/var/btcd" \
    --shell "/sbin/nologin" \
    #--no-create-home \
    --uid "${UID}" \
    --gid "${GID}" \
    "${USER}"

# Create config/data directories
RUN set -ex \
  && mkdir -pv /var/btcd \
  && mkdir -pv /var/run/btcd \
  && mkdir -pv /etc/btcd \
  && mkdir -pv /etc/btcd/keys

# Pass a tag, branch or a commit using build-arg.  This allows a docker
# image to be built from a specified Git state.  The default image
# will use the Git tip of master by default.
ARG git_repository="github.com/indra-labs/btcd"
ARG git_tag="master"

# Install dependencies and build the binaries.
RUN git clone "https://"${git_repository} /go/src/${git_repository} \
&&  cd /go/src/${git_repository} \
&&  git checkout ${git_tag}

# Source/Target release defaults
ARG ARCH=amd64
ARG GOARCH=amd64

ENV GO111MODULE=on GOOS=linux

WORKDIR $GOPATH/src/${git_repository}

RUN set -ex \
  && CGO_ENABLED=0 go build --ldflags '-w -s' -o /bin/btcd . \
  && CGO_ENABLED=0 go build --ldflags '-w -s' -o /bin/ ./cmd/...

# Copy the sample config file
RUN cp /go/src/${git_repository}/sample-btcd.conf /etc/btcd/btcd.conf

# ---
# Configure and Build the target container
# ---

FROM scratch

# Migrate User/Group to target
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder /etc/btcd /etc/btcd

# Migrate the binaries and storage folder
COPY --from=builder /bin /bin
COPY --from=builder --chown=btcd:btcd /var/btcd /var/btcd
COPY --from=builder --chown=btcd:btcd /var/run/btcd /var/run/btcd

# Enable the indra user
USER btcd:btcd

# ENV defaults
# ENV IND_LOGFILEPATH=""

# Set the data volumes
#VOLUME ["/etc/btcd"]
#VOLUME ["/var/btcd"]

# :8333  btcd peer-to-peer port
# :8334  btcd RPC port
EXPOSE 8333 8334

ENTRYPOINT ["/bin/btcd", "--configfile=/etc/btcd/btcd.conf", "--datadir=/var/btcd", "--rpckey=/etc/btcd/keys/rpc.key", "--rpccert=/etc/btcd/keys/rpc.cert"]
CMD ["--listen=0.0.0.0:8333", "--rpclisten=0.0.0.0:8334"]
