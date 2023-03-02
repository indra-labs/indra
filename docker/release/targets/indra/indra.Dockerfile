
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
#VOLUME ["/var/log/indra"]

# :8333  indra peer-to-peer port
# :8334  indra RPC port
EXPOSE 8337 8338

ENV INDRA_CONF_FILE=/etc/indra/indra.conf
ENV INDRA_DATA_DIR=/var/indra
ENV INDRA_LOGS_DIR=/var/log/indra

ENV INDRA_RPC_UNIX_LISTEN=/var/run/indra/indra.sock

ENTRYPOINT ["/bin/indra"]
