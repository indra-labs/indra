
FROM golang:1.19.4 AS builder

# User/Group definition
ENV USER=ind UID=8337 GID=8337

# Create a user/group for ind, to be migrated to the target container
RUN addgroup "ind" --gid ${GID} \
 && adduser \
    --disabled-password \
    --gecos "" \
    --home "/var/ind" \
    --shell "/sbin/nologin" \
    #--no-create-home \
    --uid "${UID}" \
    --gid "${GID}" \
    "${USER}" \
 && mkdir -pv /var/ind/.ind && chown -R ind:ind /var/ind

# Source/Target build defaults
ARG ARCH=amd64
ARG GOARCH=amd64

ENV GO111MODULE=on GOOS=linux

WORKDIR $GOPATH/src/github.com/Indra-Labs/indra
ADD . .

RUN set -ex \
#  && if [ "${ARCH}" = "amd64" ]; then export GOARCH=amd64; fi \
#  && if [ "${ARCH}" = "arm32v7" ]; then export GOARCH=arm; fi \
#  && if [ "${ARCH}" = "arm64v8" ]; then export GOARCH=arm64; fi \
  && go install -v ./cmd/ind/main.go \
  && mkdir -pv /build/bin \
  && CGO_ENABLED=0 go build --ldflags '-w -s' -o /build/bin/ind ./cmd/ind/.

# ---
# Configure and Build the target container
# ---

FROM scratch

# Migrate User/Group to target
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# Migrate the binaries
COPY --from=builder /build/bin /bin
COPY --from=builder /var/ind /var/ind

# Enable the btcd user
USER ind:ind

# ENV defaults
# ENV IND_LOGFILEPATH=""

# Set the data volume
#VOLUME ["/ind"]

# :8337  ind peer-to-peer port
# :8338  ind RPC port
EXPOSE 8337 8338

ENTRYPOINT ["/bin/ind"]