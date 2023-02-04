
# ---
# Target Configuration
# ---

ARG scratch_version="latest"

FROM indralabs/scratch-multi-arch:${scratch_version}

ARG platform
ARG version

## We can't use 'COPY --from=...' here. Using ADD will enable multi-architecture releases
ADD ./release/btcwallet-${version}/bin/${platform}/btcwallet /bin

# Enable the btcd user
USER btcwallet:btcwallet

# Set the data volumes
#VOLUME ["/etc/btcd"]
#VOLUME ["/var/btcd"]

# :8332  btcwallet RPC port
EXPOSE 8332

ENTRYPOINT ["/bin/btcwallet", "--configfile=/dev/null", "--appdata=/var/btcwallet", "--logdir=/var/btcwallet", "--cafile=/etc/btcd/keys/rpc.cert", "--rpckey=/etc/btcwallet/rpc.key", "--rpccert=/etc/btcwallet/rpc.cert", "--rpclisten=0.0.0.0:8332"]
