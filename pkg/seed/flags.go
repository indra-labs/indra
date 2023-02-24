package seed

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	key        string
	listeners  []string
	seeds      []string
	connectors []string
)

func InitFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&key, "key", "k", "", "the base58 encoded private key for the seed node")
	cmd.PersistentFlags().StringSliceVarP(&listeners, "listen", "l", []string{"/ip4/127.0.0.1/tcp/8337", "/ip6/::1/tcp/8337"}, "binds to an interface")
	cmd.PersistentFlags().StringSliceVarP(&seeds, "seed", "s", []string{}, "adds an additional seed connection  (e.g /dns4/seed0.indra.org/tcp/8337/p2p/<pub_key>)")
	cmd.PersistentFlags().StringSliceVarP(&connectors, "connect", "c", []string{}, "connects only to the seed multi-addresses specified")

	viper.BindPFlag("key", cmd.PersistentFlags().Lookup("key"))
	viper.BindPFlag("listen", cmd.PersistentFlags().Lookup("listen"))
	viper.BindPFlag("seed", cmd.PersistentFlags().Lookup("seed"))
	viper.BindPFlag("connect", cmd.PersistentFlags().Lookup("connect"))
}
