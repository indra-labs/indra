// Package client is a client for the seed RPC service for remote unlock and management.
package client

import (
	log2 "git.indra-labs.org/dev/ind/pkg/proc/log"
	"github.com/spf13/cobra"
)

var (
	log   = log2.GetLogger()
	check = log.E.Chk
)

func init() {
	initUnlock(unlockRPCCmd)
}

func Init(c *cobra.Command) {
	c.AddCommand(unlockRPCCmd)
}
