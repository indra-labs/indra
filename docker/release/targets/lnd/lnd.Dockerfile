
# ---
# Target Configuration
# ---

ARG scratch_version="latest"

FROM indralabs/scratch-multi-arch:${scratch_version}

ARG platform
ARG version

## We can't use 'COPY --from=...' here. Using ADD will enable multi-architecture releases
ADD ./release/lnd-${version}/bin/${platform}/lnd /bin

# Enable the btcd user
USER lnd:lnd

# Set the data volumes
#VOLUME ["/etc/lnd"]
#VOLUME ["/var/lnd"]

# :9735  lnd peer-to-peer port
# :10009  lnd RPC port
EXPOSE 9735 10009

ENTRYPOINT ["/bin/lnd", "--configfile=/etc/lnd/lnd.conf", "--lnddir=/var/lnd", "--datadir=/var/lnd", "--feeurl=https://nodes.lightning.computer/fees/v1/btc-fee-estimates.json"]
CMD ["--bitcoin.active", "--bitcoin.mainnet", "--bitcoin.node=neutrino"]

