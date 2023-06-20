#!/usr/bin/env zsh
reset
go test -v -tags local -gcflags "all=-trimpath=$INDRAROOT" $1 $2 $3 $4 $5