package main

import (
	"context"
	"git-indra.lan/indra-labs/indra"
	"git-indra.lan/indra-labs/indra/pkg/interrupt"
	"git-indra.lan/indra-labs/indra/pkg/p2p"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/rpc"
	"git-indra.lan/indra-labs/indra/pkg/seed"
	"git-indra.lan/indra-labs/indra/pkg/storage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {

	storage.InitFlags(seedServeCommand)
	p2p.InitFlags(seedServeCommand)
	rpc.InitFlags(seedServeCommand)

	seedCommand.AddCommand(seedServeCommand)
}

var seedServeCommand = &cobra.Command{
	Use:   "serve",
	Short: "Serves an instance of the seed node",
	Long:  `Serves an instance of the seed node.`,
	Run: func(cmd *cobra.Command, args []string) {

		log.I.Ln("-- ", log2.App, "("+viper.GetString("network")+") -", indra.SemVer, "- Network Freedom. --")

		ctx, cancel := context.WithCancel(context.Background())

		interrupt.AddHandler(cancel)

		// Seed //

		go seed.Run(ctx)

		select {
		case err := <-seed.WhenStartFailed():
			log.E.Ln("startup error:", err)
			return
		case <-seed.IsShutdown():
			log.I.Ln("shutdown complete")
		}

		log.I.Ln("-- fin --")
	},
}
