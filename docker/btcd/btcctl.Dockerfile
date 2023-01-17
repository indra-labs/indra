
# ---
# Build Process
# ---

ARG btcd_version="latest"
ARG scratch_version="latest"

FROM indralabs/btcd-base:${btcd_version} as base

# ---
# Target Configuration
# ---

FROM indralabs/scratch:${scratch_version}

## Migrate the binaries and storage folder
COPY --from=base /tmp/bin/btcctl /bin

# Enable the btcd user
USER btcd:btcd

# Set the data volumes. Should be read-only.
#VOLUME ["/etc/btcd"]

ENTRYPOINT ["/bin/btcctl", "--configfile=/etc/btcd/btcd.conf"]
