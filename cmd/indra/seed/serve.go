package seed

import (
	"context"
	
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	
	"github.com/indra-labs/indra"
	"github.com/indra-labs/indra/pkg/cfg"
	"github.com/indra-labs/indra/pkg/interrupt"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/seed"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serves an instance of the seed node",
	Long:  `Serves an instance of the seed node.`,
	Run: func(cmd *cobra.Command, args []string) {
		
		log.I.Ln("-- ", log2.App.Load(), "("+viper.GetString("network")+") -", indra.SemVer, "- Network Freedom. --")
		
		cfg.SelectNetworkParams(viper.GetString("network"))
		
		ctx, cancel := context.WithCancel(context.Background())
		interrupt.AddHandler(cancel)
		
		// Seed //
		
		go seed.Run(ctx)
		
		select {
		case <-seed.WhenStartFailed():
			log.I.Ln("stopped")
		case <-seed.WhenShutdown():
			log.I.Ln("shutdown complete")
		}
		
		log.I.Ln("-- fin --")
	},
}
