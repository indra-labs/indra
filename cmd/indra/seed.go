package main

import (
	"context"
	"git-indra.lan/indra-labs/indra"
	"git-indra.lan/indra-labs/indra/pkg/interrupt"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/rpc"
	"git-indra.lan/indra-labs/indra/pkg/rpc/client"
	"git-indra.lan/indra-labs/indra/pkg/seed"
	"github.com/multiformats/go-multiaddr"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tutorialedge/go-grpc-tutorial/chat"
	"google.golang.org/grpc"
	"os"
	"time"
)

var (
	err error
)

func init() {

	// Init flags belonging to the seed package
	seed.InitFlags(seedCmd)

	// Init flags belonging to the rpc package
	rpc.InitFlags(seedCmd)

	rootCmd.AddCommand(seedCmd)
}

var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Serves an instance of the seed node",
	Long:  `Serves an instance of the seed node.`,
	Run: func(cmd *cobra.Command, args []string) {

		log.I.Ln("-- ", log2.App, "("+viper.GetString("network")+") -", indra.SemVer, "- Network Freedom. --")

		log.I.Ln("running seed")

		var ctx context.Context
		var cancel context.CancelFunc

		ctx, cancel = context.WithCancel(context.Background())

		interrupt.AddHandler(cancel)

		//
		// RPC
		//

		rpc.RunWith(ctx, func(srv *grpc.Server) {
			chat.RegisterChatServiceServer(srv, &chat.Server{})
		})

		select {
		case <-rpc.CantStart():

			log.I.Ln("issues starting the rpc server")
			log.I.Ln("attempting a graceful shutdown")

			ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)

			rpc.Shutdown(ctx)

			select {
			case <-ctx.Done():

				log.I.Ln("can't shutdown gracefully, exiting.")

				os.Exit(1)

			default:

				log.I.Ln("graceful shutdown complete")

				os.Exit(0)
			}

		case <-rpc.IsReady():

			log.I.Ln("rpc server is ready")
		}

		client.Run(ctx)

		//
		// P2P
		//

		var config = seed.DefaultConfig

		config.SetNetwork(viper.GetString("network"))

		if config.PrivKey, err = seed.GetOrGeneratePrivKey(viper.GetString("key")); check(err) {
			return
		}

		for _, listener := range viper.GetStringSlice("listen") {
			config.ListenAddresses = append(config.ListenAddresses, multiaddr.StringCast(listener))
		}

		for _, seed := range viper.GetStringSlice("seed") {
			config.SeedAddresses = append(config.SeedAddresses, multiaddr.StringCast(seed))
		}

		for _, connector := range viper.GetStringSlice("connect") {
			config.ConnectAddresses = append(config.ConnectAddresses, multiaddr.StringCast(connector))
		}

		var srv *seed.Server

		if srv, err = seed.New(config); check(err) {
			return
		}

		if err = srv.Serve(); check(err) {
			return
		}

		log.I.Ln("-- fin --")

		return
	},
}
