
# ---
# Target Configuration
# ---

ARG scratch_version="latest"

FROM indralabs/scratch-multi-arch:${scratch_version}

ARG platform
ARG version

## We can't use 'COPY --from=...' here. Using ADD will enable multi-architecture releases
ADD ./release/lnd-${version}/bin/${platform}/lncli /bin

# Enable the btcd user
USER lnd:lnd

# Set the data volumes
#VOLUME ["/etc/lnd"]
#VOLUME ["/var/lnd"]

ENTRYPOINT ["/bin/lncli", "--lnddir=/var/lnd", "--tlscertpath=/etc/lnd/keys/rpc.cert"]
