package p2p

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	keyFlag     = "p2p-key"
	listenFlag  = "p2p-listen"
	seedFlag    = "p2p-seed"
	connectFlag = "p2p-connect"
)

var (
	key        string
	listeners  []string
	seeds      []string
	connectors []string
)

func InitFlags(cmd *cobra.Command) {

	cmd.Flags().StringVarP(&key, keyFlag, "",
		"",
		"the base58 encoded private key for the seed node")

	viper.BindPFlag(keyFlag, cmd.Flags().Lookup(keyFlag))

	cmd.PersistentFlags().StringSliceVarP(&listeners, listenFlag, "",
		[]string{
			"/ip4/127.0.0.1/tcp/8337",
			"/ip6/::1/tcp/8337",
		},
		"binds to an interface",
	)

	viper.BindPFlag(listenFlag, cmd.PersistentFlags().Lookup(listenFlag))

	cmd.PersistentFlags().StringSliceVarP(&seeds, seedFlag, "",
		[]string{},
		"adds an additional seed connection  (e.g /dns4/seed0.indra.org/tcp/8337/p2p/<pub_key>)",
	)

	viper.BindPFlag(seedFlag, cmd.PersistentFlags().Lookup(seedFlag))

	cmd.PersistentFlags().StringSliceVarP(&connectors, connectFlag, "",
		[]string{},
		"connects only to the seed multi-addresses specified",
	)

	viper.BindPFlag(connectFlag, cmd.PersistentFlags().Lookup(connectFlag))
}
