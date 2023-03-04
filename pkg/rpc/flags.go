package rpc

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	UnixPathFlag  = "rpc-unix-listen"
	TunEnableFlag = "rpc-tun-enable"
	tunKeyFlag    = "rpc-tun-key"
	TunPortFlag   = "rpc-tun-port"
	TunPeersFlag  = "rpc-tun-peer"
)
var (
	unixPath    string
	tunEnabled  bool = false
	tunPort     int  = 0
	tunPeersRaw      = []string{}
)

func InitFlags(cmd *cobra.Command) {

	cmd.PersistentFlags().StringVarP(&unixPath, UnixPathFlag, "",
		"",
		"binds to a unix socket with path (default is $HOME/.indra/indra.sock)",
	)

	viper.BindPFlag(UnixPathFlag, cmd.PersistentFlags().Lookup(UnixPathFlag))

	cmd.PersistentFlags().BoolVarP(&tunEnabled, TunEnableFlag, "",
		false,
		"enables the rpc server tunnel (default false)",
	)

	viper.BindPFlag(TunEnableFlag, cmd.PersistentFlags().Lookup(TunEnableFlag))

	cmd.PersistentFlags().IntVarP(&tunPort, TunPortFlag, "",
		tunPort,
		"binds the udp server to port (random if not selected)",
	)

	viper.BindPFlag(TunPortFlag, cmd.PersistentFlags().Lookup(TunPortFlag))

	cmd.PersistentFlags().StringSliceVarP(&tunPeersRaw, TunPeersFlag, "",
		tunPeersRaw,
		"adds a peer id to the whitelist for access",
	)

	viper.BindPFlag(TunPeersFlag, cmd.PersistentFlags().Lookup(TunPeersFlag))
}
