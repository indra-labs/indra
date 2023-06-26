#!/usr/bin/env bash
reset
go run -tags local -gcflags "all=-trimpath=$INDRAROOT" $1 --logs-level=trace $2 $3 $4 $5