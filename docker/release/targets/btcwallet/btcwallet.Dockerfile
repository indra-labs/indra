
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
USER btcd:btcd

# Set the data volumes
#VOLUME ["/etc/btcd"]
#VOLUME ["/var/btcd"]

# :8332  btcwallet RPC port
EXPOSE 8332

ENTRYPOINT ["/bin/btcwallet"]
