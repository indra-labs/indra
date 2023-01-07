#!/bin/bash

go mod tidy

IPFS_LOGGING=info go run ./cmd/indra/. -lcl serve