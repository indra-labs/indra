#!/usr/bin/env zsh
reset
go run -gcflags "all=-trimpath=$INDRAROOT" $1 $2 $3 $4 $5