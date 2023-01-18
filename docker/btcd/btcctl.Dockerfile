
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
    && CGO_ENABLED=0 GOOS=${target_os} GOARCH=${target_arch} GOARM=${target_arm_version} go build --ldflags '-w -s' -o /tmp/bin/btcctl .

# ---
# Target Configuration
# ---

FROM indralabs/scratch:${scratch_version}

## Migrate the binaries and storage folder
COPY --from=source /tmp/bin/btcctl /bin

# Enable the btcd user
USER btcd:btcd

# Set the data volumes. Should be read-only.
#VOLUME ["/etc/btcd"]

ENTRYPOINT ["/bin/btcctl", "--configfile=/etc/btcd/btcd.conf"]
