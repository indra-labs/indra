
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
COPY --from=base /tmp/bin/btcd /bin
COPY --from=base /tmp/bin/gencerts /bin

# Enable the btcd user
USER btcd:btcd

# Set the data volumes
#VOLUME ["/etc/btcd"]
#VOLUME ["/var/btcd"]

# :8333  btcd peer-to-peer port
# :8334  btcd RPC port
EXPOSE 8333 8334

ENTRYPOINT ["/bin/btcd", "--configfile=/etc/btcd/btcd.conf"]
