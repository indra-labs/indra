
# ---
# Target Configuration
# ---

ARG scratch_version="latest"

FROM indralabs/scratch-multi-arch:${scratch_version}

ARG platform
ARG version

## We can't use 'COPY --from=...' here. Using ADD will enable multi-architecture releases
ADD ./release/btcd-${version}/bin/${platform}/btcd /bin
ADD ./release/btcd-${version}/bin/${platform}/gencerts /bin

# Enable the btcd user
USER btcd:btcd

# Set the data volumes
#VOLUME ["/etc/btcd"]
#VOLUME ["/var/btcd"]

# :8333  btcd peer-to-peer port
# :8334  btcd RPC port
EXPOSE 8333 8334

ENTRYPOINT ["/bin/btcd", "--configfile=/etc/btcd/btcd.conf", "--datadir=/var/btcd", "--logdir=/var/btcd", "--listen=0.0.0.0:8333", "--rpckey=/etc/btcd/keys/rpc.key", "--rpccert=/etc/btcd/keys/rpc.cert", "--rpclisten=0.0.0.0:8334"]
