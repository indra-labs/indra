package client

import (
	"github.com/spf13/cobra"
)

func Init(c *cobra.Command) {
	c.AddCommand(clientCommand)
}

var clientCommand = &cobra.Command{
	Use:   "client",
	Short: "run a client",
	Long: "Runs indra as a client, providing a wireguard tunnel and socks5 " +
		"proxy as connectivity options",
	Run: func(cmd *cobra.Command, args []string) {
	},
}
