#!/bin/bash

# Remove existing containers
docker rm lnd-btcd-1 lnd-lnd-alice-1 lnd-lnd-bob-1 2>/dev/null

# Remove existing volumes
docker volume rm lnd_btcd_config lnd_btcd_data lnd_lnd_alice_config lnd_lnd_alice_data lnd_lnd_bob_config lnd_lnd_bob_data 2>/dev/null

# Setup an rpc key/cert for the btcwallet daemon
#docker run --rm -it \
#  --volume=lnd_:/etc/btcwallet \
#  --entrypoint="/bin/gencerts" \
#  --user=8332:8332 \
#  indralabs/btcctl-multi-arch:linux-amd64-dev \
#    --directory=/etc/btcwallet -H * -f

# Create a new wallet
#docker run --rm -it \
#  --volume=btcd_btcwallet_config:/etc/btcwallet \
#  --volume=btcd_btcwallet_data:/var/btcwallet \
#  indralabs/btcwallet-multi-arch:linux-amd64-dev \
#    --simnet --createtemp

#docker run --rm -it \
#  --volume=btcd_btcwallet_config:/etc/btcwallet \
#  --volume=btcd_btcwallet_data:/var/btcwallet \
#  indralabs/btcwallet-multi-arch:linux-amd64-dev \
#    --simnet importprivkey FuarsNCxniX277tBYt1BDGPB6cRTUfeEhUBXNAjrg3cdsWZTNcPj
