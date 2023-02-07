#!/bin/bash

# Remove existing containers
echo "wiping the exising environment..."
docker stop lnd-btcd-1 lnd-lnd-miner-1 lnd-lnd-alice-1 lnd-lnd-bob-1 2>/dev/null 1>/dev/null
docker rm lnd-btcd-1 lnd-lnd-miner-1 lnd-lnd-alice-1 lnd-lnd-bob-1 2>/dev/null 1>/dev/null

# Remove existing volumes
docker volume rm lnd_btcd_config lnd_btcd_data lnd_lnd_miner_config lnd_lnd_miner_data lnd_lnd_alice_config lnd_lnd_alice_data lnd_lnd_bob_config lnd_lnd_bob_data 2>/dev/null 1>/dev/null

echo "running bootstrapping of wallets"
docker-compose -f docker/release/targets/lnd/docker-compose.yml up --quiet-pull --detach 1>/dev/null

rm docker/release/targets/lnd/.env 2>/dev/null

echo "waiting for the environment to start..."
sleep 5

echo "generating an lnd pubkey and address for miner, alice and bob."

docker/release/targets/lnd/bin/lnsim-lncli-miner getinfo | jq -r .identity_pubkey | xargs -I {} echo "MINER_PUBKEY={}" \
  >> docker/release/targets/lnd/.env
docker/release/targets/lnd/bin/lnsim-lncli-miner newaddress np2wkh | jq -r .address | xargs -I {} echo "MINER_ADDRESS={}" \
  >> docker/release/targets/lnd/.env

docker/release/targets/lnd/bin/lnsim-lncli-alice getinfo | jq -r .identity_pubkey | xargs -I {} echo "ALICE_PUBKEY={}" \
  >> docker/release/targets/lnd/.env
docker/release/targets/lnd/bin/lnsim-lncli-alice newaddress np2wkh | jq -r .address | xargs -I {} echo "ALICE_ADDRESS={}" \
  >> docker/release/targets/lnd/.env

docker/release/targets/lnd/bin/lnsim-lncli-bob getinfo | jq -r .identity_pubkey | xargs -I {} echo "BOB_PUBKEY={}" \
  >> docker/release/targets/lnd/.env
docker/release/targets/lnd/bin/lnsim-lncli-bob newaddress np2wkh | jq -r .address | xargs -I {} echo "BOB_ADDRESS={}" \
  >> docker/release/targets/lnd/.env

echo "shutting down environment to enable btcd to mine to the miner address."
docker-compose -f docker/release/targets/lnd/docker-compose.yml down

echo "running coin sending to alice and bob"
docker-compose --env-file=docker/release/targets/lnd/.env -f docker/release/targets/lnd/docker-compose.yml up --quiet-pull --detach

echo "waiting for the environment to start..."
sleep 5

echo "generating 500 blocks to enable segwit."
docker/release/targets/lnd/bin/lnsim-btcctl generate 500 1>/dev/null

echo "getting miners wallet balance"
docker/release/targets/lnd/bin/lnsim-lncli-miner walletbalance

echo "sourcing environment variables"
source docker/release/targets/lnd/.env

echo "sending coins to alice and bob."
docker/release/targets/lnd/bin/lnsim-lncli-miner sendcoins --addr $ALICE_ADDRESS --amt 100000000000
docker/release/targets/lnd/bin/lnsim-lncli-miner sendcoins --addr $BOB_ADDRESS --amt 100000000000

echo "generating 100 blocks to allow them to be spent"
docker/release/targets/lnd/bin/lnsim-btcctl generate 100 1>/dev/null

echo "getting alice's wallet balance:"
docker/release/targets/lnd/bin/lnsim-lncli-alice walletbalance
echo "getting bob's wallet balance:"
docker/release/targets/lnd/bin/lnsim-lncli-bob walletbalance

docker-compose -f docker/release/targets/lnd/docker-compose.yml down

echo "setup is complete!"
