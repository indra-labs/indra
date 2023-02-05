#!/bin/bash

docker rm btcd-btcd-1 btcd-btcctl-1 btcd-btcwallet-1 2>/dev/null

docker volume rm btcd_config btcd_data btcd_btcwallet_config btcd_btcwallet_data 2>/dev/null

docker run --rm -it \
  --volume=btcd_btcwallet_config:/etc/btcwallet \
  --entrypoint="/bin/gencerts" \
  --user=8332:8332 \
  indralabs/btcd-multi-arch:linux-amd64-dev \
    --directory=/etc/btcwallet -H 172.16.42.3 -f

docker run --rm -it \
  --volume=btcd_btcwallet_config:/etc/btcwallet \
  --volume=btcd_btcwallet_data:/var/btcwallet \
  indralabs/btcwallet-multi-arch:linux-amd64-dev \
    --simnet --create

