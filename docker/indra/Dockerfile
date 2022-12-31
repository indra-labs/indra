
FROM golang:1.19.4 AS builder

# User/Group definition
ENV USER=indra GROUP=indra UID=8337 GID=8337

# Create a user/group for indra, to be migrated to the target container
RUN addgroup ${GROUP} --gid ${GID} \
 && adduser \
    --disabled-password \
    --gecos "" \
    --home "/var/indra" \
    --shell "/sbin/nologin" \
    #--no-create-home \
    --uid "${UID}" \
    --gid "${GID}" \
    "${USER}" \
 && mkdir -pv /var/indra/.indra && chown -R indra:indra /var/indra

# Source/Target release defaults
ARG ARCH=amd64
ARG GOARCH=amd64

ENV GO111MODULE=on GOOS=linux

WORKDIR $GOPATH/src/github.com/Indra-Labs/indra
ADD . .

RUN set -ex \
#  && if [ "${ARCH}" = "amd64" ]; then export GOARCH=amd64; fi \
#  && if [ "${ARCH}" = "arm32v7" ]; then export GOARCH=arm; fi \
#  && if [ "${ARCH}" = "arm64v8" ]; then export GOARCH=arm64; fi \
  && go install -v ./cmd/indra/. \
  && CGO_ENABLED=0 go build --ldflags '-w -s' -o /bin/indra ./cmd/indra/.

# ---
# Configure and Build the target container
# ---

FROM scratch

# Migrate User/Group to target
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# Migrate the binaries and storage folder
COPY --from=builder --chown=indra:indra /bin/indra /bin/indra
COPY --from=builder --chown=indra:indra /var/indra /var/indra

# Enable the indra user
USER indra:indra

# ENV defaults
# ENV IND_LOGFILEPATH=""

# Set the data volume
#VOLUME ["/var/indra"]

# :8337  indra peer-to-peer port
# :8338  indra RPC port
EXPOSE 8337 8338

ENTRYPOINT ["/bin/indra"]
