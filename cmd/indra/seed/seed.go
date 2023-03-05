package seed

import (
	"git-indra.lan/indra-labs/indra"
	"git-indra.lan/indra-labs/indra/cmd/indra/seed/client"
	"git-indra.lan/indra-labs/indra/pkg/p2p"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/rpc"
	"git-indra.lan/indra-labs/indra/pkg/storage"
	"github.com/spf13/cobra"
)

var (
	log   = log2.GetLogger(indra.PathBase)
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
