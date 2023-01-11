FROM golang:1.19.4 AS builder

# User/Group definition
ENV USER=lnd GROUP=lnd UID=9735 GID=9735

## Create a user/group for indra, to be migrated to the target container
RUN addgroup ${GROUP} --gid ${GID} \
 && adduser \
    --disabled-password \
    --gecos "" \
    --home "/var/lnd" \
    --shell "/sbin/nologin" \
    #--no-create-home \
    --uid "${UID}" \
    --gid "${GID}" \
    "${USER}" \
 && mkdir -pv /var/lnd/.lnd && chown -R lnd:lnd /var/lnd

# Pass a tag, branch or a commit using build-arg.  This allows a docker
# image to be built from a specified Git state.  The default image
# will use the Git tip of master by default.
ARG checkout="v0.15.5-beta"
ARG git_url="https://github.com/lightningnetwork/lnd"

# Install dependencies and build the binaries.
RUN git clone $git_url /go/src/github.com/lightningnetwork/lnd \
&&  cd /go/src/github.com/lightningnetwork/lnd \
&&  git checkout $checkout

# Source/Target release defaults
ARG ARCH=amd64
ARG GOARCH=amd64

ENV GO111MODULE=on GOOS=linux

WORKDIR $GOPATH/src/github.com/lightningnetwork/lnd

RUN set -ex \
  #&& go install -v ./cmd/indra/. \
  && CGO_ENABLED=0 go build --ldflags '-w -s' -o /bin/lnd ./cmd/lnd/. \
  && CGO_ENABLED=0 go build --ldflags '-w -s' -o /bin/lncli ./cmd/lncli/.

# ---
# Configure and Build the target container
# ---

FROM scratch

# Migrate User/Group to target
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# Migrate the binaries and storage folder
COPY --from=builder --chown=lnd:lnd /bin /bin
COPY --from=builder --chown=lnd:lnd /var/lnd /var/lnd

# Enable the indra user
USER lnd:lnd

# ENV defaults
# ENV IND_LOGFILEPATH=""

# Set the data volume
#VOLUME ["/var/lnd"]

# :9735  lnd peer-to-peer port
# :10009  lnd RPC port
EXPOSE 9735 10009

ENTRYPOINT ["/bin/lnd"]
