#!/usr/bin/env zsh
reset
go test -v -tags local -gcflags "all=-trimpath=/home/loki/work/loki/indra-labs/indra" $1 $2 $3 $4 $5