
# ---
# Build Process
# ---

ARG source_version="v0.1.7"
ARG scratch_version="latest"

FROM indralabs/indra-source-local:${source_version} as source

ARG target_os="linux"
ARG target_arch="amd64"
ARG target_arm_version=""

RUN set -ex echo "building binaries for ${target_os}/${target_arch}" \
    && CGO_ENABLED=0 GOOS=${target_os} GOARCH=${target_arch} GOARM=${target_arm_version} go build --ldflags '-w -s' -o /tmp/bin/indra ./cmd/indra/.

# ---
# Target Configuration
# ---

FROM indralabs/scratch:${scratch_version}

## Migrate the binaries and storage folder
COPY --from=source /tmp/bin/indra /bin

# Enable the btcd user
USER indra:indra

# Set the data volumes
#VOLUME ["/etc/indra"]
#VOLUME ["/var/indra"]

# :8333  indra peer-to-peer port
# :8334  indra RPC port
EXPOSE 8337 8338

ENTRYPOINT ["/bin/indra", "--conffile=/etc/indra/indra.conf"]
