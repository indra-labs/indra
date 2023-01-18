
# ---
# Build Process
# ---

ARG source_version="v0.23.3"
ARG scratch_version="latest"

FROM indralabs/btcd-source:${source_version} as source

ARG target_os="linux"
ARG target_arch="amd64"
ARG target_arm_version=""

RUN set -ex echo "building binaries for ${target_os}/${target_arch}" \
    && CGO_ENABLED=0 GOOS=${target_os} GOARCH=${target_arch} GOARM=${target_arm_version} go build --ldflags '-w -s' -o /tmp/bin/btcd . \
    && CGO_ENABLED=0 GOOS=${target_os} GOARCH=${target_arch} GOARM=${target_arm_version} go build --ldflags '-w -s' -o /tmp/bin/gencerts ./cmd/gencerts/.

# ---
# Target Configuration
# ---

FROM indralabs/scratch:${scratch_version}

## Migrate the binaries and storage folder
COPY --from=source /tmp/bin/btcd /bin
COPY --from=source /tmp/bin/gencerts /bin

# Enable the btcd user
USER btcd:btcd

# Set the data volumes
#VOLUME ["/etc/btcd"]
#VOLUME ["/var/btcd"]

# :8333  btcd peer-to-peer port
# :8334  btcd RPC port
EXPOSE 8333 8334

ENTRYPOINT ["/bin/btcd", "--configfile=/etc/btcd/btcd.conf"]
