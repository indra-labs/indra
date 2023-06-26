#!/usr/bin/env bash
reset
go test -v -gcflags "all=-trimpath=$INDRAROOT" $1 $2 $3 $4 $5