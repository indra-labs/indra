package main

import (
	"git-indra.lan/indra-labs/indra"
	"git-indra.lan/indra-labs/indra/pkg/cfg"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/server"
	"github.com/multiformats/go-multiaddr"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	seedCmd.PersistentFlags().StringSliceVarP(&seeds, "seed", "s", []string{}, "adds an additional seed connection  (e.g /dns4/seed0.indra.org/tcp/8337)")
	seedCmd.PersistentFlags().StringSliceVarP(&connectors, "connect", "c", []string{}, "connects only to the seed multi-address specified")

	viper.BindPFlag("key", seedCmd.PersistentFlags().Lookup("key"))
	viper.BindPFlag("listen", seedCmd.PersistentFlags().Lookup("listen"))
	viper.BindPFlag("seed", seedCmd.PersistentFlags().Lookup("seed"))
	viper.BindPFlag("connect", seedCmd.PersistentFlags().Lookup("connect"))

	rootCmd.AddCommand(seedCmd)
}

var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Serves an instance of the seed node",
	Long:  `Serves an instance of the seed node.`,
	Run: func(cmd *cobra.Command, args []string) {

		log.I.Ln("-- ", log2.App, "("+viper.GetString("network")+") -", indra.SemVer, "- Network Freedom. --")

		var err error
		var params = cfg.SelectNetworkParams(viper.GetString("network"))

		if server.DefaultConfig.PrivKey, err = server.GetOrGeneratePrivKey(viper.GetString("key")); check(err) {
			return
		}

		for _, listener := range viper.GetStringSlice("listen") {
			server.DefaultConfig.ListenAddresses = append(server.DefaultConfig.ListenAddresses, multiaddr.StringCast(listener))
		}

		for _, seed := range viper.GetStringSlice("seed") {
			server.DefaultConfig.SeedAddresses = append(server.DefaultConfig.SeedAddresses, multiaddr.StringCast(seed))
		}

		for _, connector := range viper.GetStringSlice("connect") {
			server.DefaultConfig.ConnectAddresses = append(server.DefaultConfig.ConnectAddresses, multiaddr.StringCast(connector))
		}

		var srv *server.Server

		log.I.Ln("running serve.")

		if srv, err = server.New(params, server.DefaultConfig); check(err) {
			return
		}

		if err = srv.Serve(); check(err) {
			return
		}

		log.I.Ln("-- fin --")

		return
	},
}
