
# ---
# Build Process
# ---

ARG source_version="v0.15.5-beta"
ARG scratch_version="latest"

FROM indralabs/lnd-source:${source_version} as source

ARG target_os="linux"
ARG target_arch="amd64"
ARG target_arm_version=""

RUN set -ex echo "building binaries for ${target_os}/${target_arch}" \
    && CGO_ENABLED=0 GOOS=${target_os} GOARCH=${target_arch} GOARM=${target_arm_version} go build --ldflags '-w -s' -o /tmp/bin/lncli ./cmd/lncli/.

# ---
# Target Configuration
# ---

FROM indralabs/scratch:${scratch_version}

## Migrate the binaries and storage folder
COPY --from=source /tmp/bin/lncli /bin

# Enable the btcd user
USER lnd:lnd

# Set the data volumes
#VOLUME ["/etc/btcd"]
#VOLUME ["/var/btcd"]

# :8333  btcd peer-to-peer port
# :8334  btcd RPC port
EXPOSE 8333 8334

ENTRYPOINT ["/bin/lncli"]
