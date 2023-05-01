package holo

import (
	"git-indra.lan/indra-labs/indra"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	fails = log.E.Chk
)

type Set struct {
}

func New() *Set {
	return &Set{}
}
