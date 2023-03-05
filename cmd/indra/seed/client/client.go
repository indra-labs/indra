package client

import (
	"git-indra.lan/indra-labs/indra"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
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
