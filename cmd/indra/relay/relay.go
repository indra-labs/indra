package relay

import (
	"github.com/indra-labs/indra"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/spf13/cobra"
)

var (
	log   = log2.GetLogger()
	check = log.E.Chk
)

var (
	wireguardEnable bool
	wireguardCIDR   string
	socksEnable     bool
	socksListener   string
)

func Init(c *cobra.Command) {
	relayCommand.PersistentFlags().BoolVarP(&wireguardEnable, "wireguard",
		"w", false, "enable wiregfuard tunnel")
	relayCommand.PersistentFlags().BoolVarP(&socksEnable, "socks",
		"s", false, "enable socks proxy")
	relayCommand.PersistentFlags().StringVar(&socksListener, "socks-listener",
		"localhost:8080", "set address for socks 5 proxy listener")

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
