#!/bin/bash

go mod tidy

IPFS_LOGGING=debug go run ./cmd/indra/. -lcl serve