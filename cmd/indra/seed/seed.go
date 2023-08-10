// Package seed is a non-relay node that simply accepts and propagates peer advertisment gossip to clients and relays on the network.
package seed

import (
	"git.indra-labs.org/dev/ind/cmd/indra/seed/client"
	"git.indra-labs.org/dev/ind/pkg/p2p"
	log2 "git.indra-labs.org/dev/ind/pkg/proc/log"
	"git.indra-labs.org/dev/ind/pkg/rpc"
	"git.indra-labs.org/dev/ind/pkg/storage"
	"github.com/spf13/cobra"
)

var (
	log   = log2.GetLogger()
	check = log.E.Chk
)

func init() {
	storage.InitFlags(serveCmd)
	p2p.InitFlags(serveCmd)
	rpc.InitFlags(serveCmd)
}

func Init(c *cobra.Command) {
	client.Init(rpcCmd)

	seedCommand.AddCommand(rpcCmd)
	seedCommand.AddCommand(serveCmd)

	c.AddCommand(seedCommand)
}

var seedCommand = &cobra.Command{
	Use:   "seed",
	Short: "run and manage your seed node",
	Long:  `run and manage your seed node`,
}
