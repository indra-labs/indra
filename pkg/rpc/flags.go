package rpc

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var (
	unixPathFlag  = "rpc-unix-listen"
	tunEnableFlag = "rpc-tun-enable"
	tunKeyFlag    = "rpc-tun-key"
	tunPortFlag   = "rpc-tun-port"
	tunPeersFlag  = "rpc-tun-peer"
)

var (
	tunKeyRaw   string
	tunPeersRaw = []string{}
)

func InitFlags(cmd *cobra.Command) {

	cobra.OnInitialize(initUnixSockPath)

	cmd.PersistentFlags().StringVarP(&unixPath, unixPathFlag, "",
		"",
		"binds to a unix socket with path (default is $HOME/.indra/indra.sock)",
	)

	viper.BindPFlag(unixPathFlag, cmd.PersistentFlags().Lookup(unixPathFlag))

	cmd.PersistentFlags().BoolVarP(&isTunnelEnabled, tunEnableFlag, "",
		false,
		"enables the rpc server tunnel (default false)",
	)

	viper.BindPFlag(tunEnableFlag, cmd.PersistentFlags().Lookup(tunEnableFlag))

	cmd.Flags().StringVarP(&tunKeyRaw, tunKeyFlag, "",
		"",
		"the base58 encoded pre-shared key for accessing the rpc",
	)

	viper.BindPFlag(tunKeyFlag, cmd.Flags().Lookup(tunKeyFlag))

	cmd.PersistentFlags().IntVarP(&tunnelPort, tunPortFlag, "",
		tunnelPort,
		"binds the udp server to port (random if not selected)",
	)

	viper.BindPFlag(tunPortFlag, cmd.PersistentFlags().Lookup(tunPortFlag))

	cmd.PersistentFlags().StringSliceVarP(&tunPeersRaw, tunPeersFlag, "",
		tunPeersRaw,
		"adds a peer id to the whitelist for access",
	)

	viper.BindPFlag(tunPeersFlag, cmd.PersistentFlags().Lookup(tunPeersFlag))
}

func initUnixSockPath() {

	if viper.GetString(unixPathFlag) != "" {
		return
	}

	home, err := os.UserHomeDir()

	cobra.CheckErr(err)

	viper.Set(unixPathFlag, home+"/.indra/indra.sock")
}
