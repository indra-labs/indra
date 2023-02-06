
# ---
# Target Configuration
# ---

ARG scratch_version="latest"

FROM indralabs/scratch-multi-arch:${scratch_version}

ARG platform
ARG version

## We can't use 'COPY --from=...' here. Using ADD will enable multi-architecture releases
ADD ./release/btcd-${version}/bin/${platform}/btcctl /bin
ADD ./release/btcd-${version}/bin/${platform}/gencerts /bin

# Enable the btcd user
USER btcd:btcd

# Set the data volumes
#VOLUME ["/etc/btcd"]
#VOLUME ["/var/btcd"]

ENTRYPOINT ["/bin/btcctl", "--configfile=/dev/null", "--rpccert=/etc/btcd/keys/rpc.cert"]

