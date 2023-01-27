
# ---
# Target Configuration
# ---

ARG scratch_version="latest"

FROM indralabs/scratch-multi-arch:${scratch_version}

ARG platform
ARG version

## We can't use 'COPY --from=...' here. Using ADD will enable multi-architecture releases
ADD ./release/indra-${version}/bin/${platform}/indra /bin

# Enable the btcd user
USER indra:indra

# Set the data volumes
#VOLUME ["/etc/indra"]
#VOLUME ["/var/indra"]

# :8333  indra peer-to-peer port
# :8334  indra RPC port
EXPOSE 8337 8338

ENTRYPOINT ["/bin/indra", "--conffile=/etc/indra/indra.conf"]
