// Package seed is a non-relay node that simply accepts and propagates peer advertisment gossip to clients and relays on the network.
package seed

import (
	"github.com/indra-labs/indra/cmd/indra/seed/client"
	"github.com/indra-labs/indra/pkg/p2p"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/rpc"
	"github.com/indra-labs/indra/pkg/storage"
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
