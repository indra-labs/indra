package client

import (
	"github.com/indra-labs/indra"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/spf13/cobra"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

func init() {
	initUnlock(unlockRPCCmd)
}

func Init(c *cobra.Command) {
	c.AddCommand(unlockRPCCmd)
}
