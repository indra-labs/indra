package main

import (
	"git-indra.lan/indra-labs/indra"
	"git-indra.lan/indra-labs/indra/pkg/cfg"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/rpc"
	"git-indra.lan/indra-labs/indra/pkg/seed"
	"github.com/multiformats/go-multiaddr"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"math/rand"
	"time"
)

var (
	key                string
	listeners          []string
	seeds              []string
	connectors         []string
	rpc_listen_port    uint16
	rpc_key            string
	rpc_whitelist_peer []string
	rpc_whitelist_ip   []string
)

func init() {

	seedCmd.PersistentFlags().StringVarP(&key, "key", "k", "", "the base58 encoded private key for the seed node")
	seedCmd.PersistentFlags().StringSliceVarP(&listeners, "listen", "l", []string{"/ip4/127.0.0.1/tcp/8337", "/ip6/::1/tcp/8337"}, "binds to an interface")
	seedCmd.PersistentFlags().StringSliceVarP(&seeds, "seed", "s", []string{}, "adds an additional seed connection  (e.g /dns4/seed0.indra.org/tcp/8337/p2p/<pub_key>)")
	seedCmd.PersistentFlags().StringSliceVarP(&connectors, "connect", "c", []string{}, "connects only to the seed multi-addresses specified")
	seedCmd.PersistentFlags().Uint16VarP(&rpc_listen_port, "rpc-listen-port", "", 0, "binds the udp server to port (random if not selected)")
	seedCmd.PersistentFlags().StringVarP(&rpc_key, "rpc-key", "", "", "the base58 encoded pre-shared key for accessing the rpc")
	seedCmd.PersistentFlags().StringSliceVarP(&rpc_whitelist_peer, "rpc-whitelist-peer", "", []string{}, "adds a peer id to the whitelist for access")
	seedCmd.PersistentFlags().StringSliceVarP(&rpc_whitelist_ip, "rpc-whitelist-ip", "", []string{}, "adds a cidr ip range to the whitelist for access (e.g /ip4/127.0.0.1/ipcidr/32)")

	viper.BindPFlag("key", seedCmd.PersistentFlags().Lookup("key"))
	viper.BindPFlag("listen", seedCmd.PersistentFlags().Lookup("listen"))
	viper.BindPFlag("seed", seedCmd.PersistentFlags().Lookup("seed"))
	viper.BindPFlag("connect", seedCmd.PersistentFlags().Lookup("connect"))
	viper.BindPFlag("rpc-listen-port", seedCmd.PersistentFlags().Lookup("rpc-listen-port"))
	viper.BindPFlag("rpc-key", seedCmd.PersistentFlags().Lookup("rpc-key"))
	viper.BindPFlag("rpc-whitelist-peer", seedCmd.PersistentFlags().Lookup("rpc-whitelist-peer"))
	viper.BindPFlag("rpc-whitelist-ip", seedCmd.PersistentFlags().Lookup("rpc-whitelist-ip"))

	rootCmd.AddCommand(seedCmd)
}

var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Serves an instance of the seed node",
	Long:  `Serves an instance of the seed node.`,
	Run: func(cmd *cobra.Command, args []string) {

		log.I.Ln("-- ", log2.App, "("+viper.GetString("network")+") -", indra.SemVer, "- Network Freedom. --")

		var err error
		var config = seed.DefaultConfig

		config.Params = cfg.SelectNetworkParams(viper.GetString("network"))

		config.RPCConfig.Key.Decode(viper.GetString("rpc-key"))

		if config.RPCConfig.IsEnabled() {

			config.RPCConfig.ListenPort = viper.GetUint16("rpc-listen-port")

			if config.RPCConfig.ListenPort == 0 {
				rand.Seed(time.Now().Unix())
				config.RPCConfig.ListenPort = uint16(rand.Intn(45534) + 10000)

				viper.Set("rpc-listen-port", config.RPCConfig.ListenPort)
			}

			for _, ip := range viper.GetStringSlice("rpc-whitelist-ip") {
				config.RPCConfig.IP_Whitelist = append(config.RPCConfig.IP_Whitelist, multiaddr.StringCast(ip))
			}

			for _, peer := range viper.GetStringSlice("rpc-whitelist-peer") {

				var pubKey rpc.RPCPublicKey

				pubKey.Decode(peer)

				config.RPCConfig.Peer_Whitelist = append(config.RPCConfig.Peer_Whitelist, pubKey)
			}
		}

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

		log.I.Ln("running serve.")

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
