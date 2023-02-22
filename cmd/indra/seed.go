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
	"os"
)

var (
	err error
)

var (
	key                string
	listeners          []string
	seeds              []string
	connectors         []string
	rpc_enable         bool
	rpc_listen_port    uint16
	rpc_key            string
	rpc_whitelist_peer []string
	rpc_whitelist_ip   []string
	rpc_unix_path      string
)

func init() {

	seedCmd.PersistentFlags().StringVarP(&key, "key", "k", "", "the base58 encoded private key for the seed node")
	seedCmd.PersistentFlags().StringSliceVarP(&listeners, "listen", "l", []string{"/ip4/127.0.0.1/tcp/8337", "/ip6/::1/tcp/8337"}, "binds to an interface")
	seedCmd.PersistentFlags().StringSliceVarP(&seeds, "seed", "s", []string{}, "adds an additional seed connection  (e.g /dns4/seed0.indra.org/tcp/8337/p2p/<pub_key>)")
	seedCmd.PersistentFlags().StringSliceVarP(&connectors, "connect", "c", []string{}, "connects only to the seed multi-addresses specified")

	seedCmd.PersistentFlags().BoolVarP(&rpc_enable, "rpc-enable", "", false, "enables the rpc server")
	seedCmd.PersistentFlags().Uint16VarP(&rpc_listen_port, "rpc-listen-port", "", 0, "binds the udp server to port (random if not selected)")
	seedCmd.PersistentFlags().StringVarP(&rpc_key, "rpc-key", "", "", "the base58 encoded pre-shared key for accessing the rpc")
	seedCmd.PersistentFlags().StringSliceVarP(&rpc_whitelist_peer, "rpc-whitelist-peer", "", []string{}, "adds a peer id to the whitelist for access")
	seedCmd.PersistentFlags().StringSliceVarP(&rpc_whitelist_ip, "rpc-whitelist-ip", "", []string{}, "adds a cidr ip range to the whitelist for access (e.g /ip4/127.0.0.1/ipcidr/32)")
	seedCmd.PersistentFlags().StringVarP(&rpc_unix_path, "rpc-listen-unix", "", "/tmp/indra.sock", "binds to a unix socket with path (default is /tmp/indra.sock)")

	viper.BindPFlag("key", seedCmd.PersistentFlags().Lookup("key"))
	viper.BindPFlag("listen", seedCmd.PersistentFlags().Lookup("listen"))
	viper.BindPFlag("seed", seedCmd.PersistentFlags().Lookup("seed"))
	viper.BindPFlag("connect", seedCmd.PersistentFlags().Lookup("connect"))
	viper.BindPFlag("rpc-enable", seedCmd.PersistentFlags().Lookup("rpc-enable"))
	viper.BindPFlag("rpc-listen-port", seedCmd.PersistentFlags().Lookup("rpc-listen-port"))
	viper.BindPFlag("rpc-key", seedCmd.PersistentFlags().Lookup("rpc-key"))
	viper.BindPFlag("rpc-whitelist-peer", seedCmd.PersistentFlags().Lookup("rpc-whitelist-peer"))
	viper.BindPFlag("rpc-whitelist-ip", seedCmd.PersistentFlags().Lookup("rpc-whitelist-ip"))
	viper.BindPFlag("rpc-listen-unix", seedCmd.PersistentFlags().Lookup("rpc-listen-unix"))

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

			// We need to enable specific gRPC services here
			//srv := rpc.Server()
			//helloworld.RegisterGreeterServer(srv, helloworld)
			s := chat.Server{}

			chat.RegisterChatServiceServer(rpc.Server(), &s)

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
