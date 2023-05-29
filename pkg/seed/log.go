package seed

import (
	"github.com/indra-labs/indra"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)
