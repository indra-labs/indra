#!/bin/bash

GOINSECURE="git-indra.lan/*" GOPRIVATE="git-indra.lan/*" go mod tidy

IPFS_LOGGING=info go run ./cmd/indra/. $@