package relay

import (
	"git.indra-labs.org/dev/ind"
	log2 "git.indra-labs.org/dev/ind/pkg/proc/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	log   = log2.GetLogger()
	check = log.E.Chk
)

var (
	wireguardEnable bool
	wireguardCIDR   string // todo: there must be something like this. default route to 1
	socksEnable     bool
	socksListener   string
)

func Init(c *cobra.Command) {
	relayCommand.PersistentFlags().BoolVarP(&wireguardEnable, "wireguard",
		"w", false, "enable wireguard tunnel")
	relayCommand.PersistentFlags().BoolVarP(&socksEnable, "socks",
		"s", false, "enable socks proxy")
	relayCommand.PersistentFlags().StringVarP(&socksListener, "socks-listener",
		"l", "localhost:8080", "set address for socks 5 proxy listener")

	viper.BindPFlag("wireguard", relayCommand.PersistentFlags().Lookup("wireguard"))
	viper.BindPFlag("socks", relayCommand.PersistentFlags().Lookup("socks"))
	viper.BindPFlag("socks-listener", relayCommand.PersistentFlags().Lookup("socks-listener"))

	c.AddCommand(relayCommand)
}

var relayCommand = &cobra.Command{
	Use:   "relay",
	Short: "run a relay",
	Long:  "Runs indra as a full relay, with optional client.",

	Run: func(cmd *cobra.Command, args []string) {
		log.I.Ln(log2.App.Load(), indra.SemVer)
		nw, _ := cmd.Parent().PersistentFlags().GetString("network")
		var dd string
		dd, _ = cmd.Parent().PersistentFlags().GetString("data-dir")
		log.T.S("cmd", dd, nw, args)
	},
}
