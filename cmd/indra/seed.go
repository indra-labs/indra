package main

import (
	"context"
	"git-indra.lan/indra-labs/indra"
	"git-indra.lan/indra-labs/indra/pkg/interrupt"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/rpc"
	"git-indra.lan/indra-labs/indra/pkg/seed"
	"github.com/multiformats/go-multiaddr"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tutorialedge/go-grpc-tutorial/chat"
	"google.golang.org/grpc"
	"os"
)

var (
	err error
)

var (
	key        string
	listeners  []string
	seeds      []string
	connectors []string
)

func init() {

	seedCmd.PersistentFlags().StringVarP(&key, "key", "k", "", "the base58 encoded private key for the seed node")
	seedCmd.PersistentFlags().StringSliceVarP(&listeners, "listen", "l", []string{"/ip4/127.0.0.1/tcp/8337", "/ip6/::1/tcp/8337"}, "binds to an interface")
	seedCmd.PersistentFlags().StringSliceVarP(&seeds, "seed", "s", []string{}, "adds an additional seed connection  (e.g /dns4/seed0.indra.org/tcp/8337/p2p/<pub_key>)")
	seedCmd.PersistentFlags().StringSliceVarP(&connectors, "connect", "c", []string{}, "connects only to the seed multi-addresses specified")

	viper.BindPFlag("key", seedCmd.PersistentFlags().Lookup("key"))
	viper.BindPFlag("listen", seedCmd.PersistentFlags().Lookup("listen"))
	viper.BindPFlag("seed", seedCmd.PersistentFlags().Lookup("seed"))
	viper.BindPFlag("connect", seedCmd.PersistentFlags().Lookup("connect"))

	rpc.Configure(seedCmd)

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

		if viper.GetBool("rpc-enable") {

			log.I.Ln("enabling rpc server")

			if err = rpc.ConfigureWithViper(); check(err) {
				os.Exit(1)
			}

			rpc.Register(func(srv *grpc.Server) {
				chat.RegisterChatServiceServer(srv, &chat.Server{})
			})

			log.I.Ln("starting rpc server")

			go rpc.Start(ctx)

			select {
			case <-rpc.CantStart():

				log.I.Ln("issues starting the rpc server")
				log.I.Ln("attempting a graceful shutdown")

				rpc.Shutdown()

				os.Exit(1)

			case <-rpc.IsReady():

				log.I.Ln("rpc server is ready!")
			}
		}

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
