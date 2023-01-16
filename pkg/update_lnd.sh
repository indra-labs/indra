#!/bin/bash

# Run this script from the root of the repository to update the static lnd

go run ./cmd/gimport/. https://github.com/lightningnetwork/lnd.git \
      v0.15.5-beta pkg/lnd